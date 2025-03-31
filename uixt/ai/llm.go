package ai

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/config"
	"github.com/httprunner/httprunner/v5/internal/json"
	"github.com/httprunner/httprunner/v5/uixt/types"
)

type ILLMService interface {
	Call(opts *PlanningOptions) (*PlanningResult, error)
}

func NewGPT4oLLMService() (*openaiLLMService, error) {
	return &openaiLLMService{}, nil
}

type openaiLLMService struct{}

func (s openaiLLMService) Call(opts *PlanningOptions) (*PlanningResult, error) {
	return nil, nil
}

// PlanningOptions represents the input options for planning
type PlanningOptions struct {
	UserInstruction string          `json:"user_instruction"` // append to system prompt
	Message         *schema.Message `json:"message"`
	Size            types.Size      `json:"size"`
}

// PlanningResult represents the result of planning
type PlanningResult struct {
	NextActions   []ParsedAction `json:"actions"`
	ActionSummary string         `json:"summary"`
}

// VLMResponse represents the response from the Vision Language Model
type VLMResponse struct {
	Actions []ParsedAction `json:"actions"`
	Error   string         `json:"error,omitempty"`
}

// ParsedAction represents a parsed action from the VLM response
type ParsedAction struct {
	ActionType   ActionType             `json:"actionType"`
	ActionInputs map[string]interface{} `json:"actionInputs"`
	Thought      string                 `json:"thought"`
}

type ActionType string

const (
	ActionTypeClick    ActionType = "click"
	ActionTypeTap      ActionType = "tap"
	ActionTypeDrag     ActionType = "drag"
	ActionTypeSwipe    ActionType = "swipe"
	ActionTypeWait     ActionType = "wait"
	ActionTypeFinished ActionType = "finished"
	ActionTypeCallUser ActionType = "call_user"
	ActionTypeType     ActionType = "type"
	ActionTypeScroll   ActionType = "scroll"
)

const (
	defaultTimeout = 60 * time.Second
)

type OpenAIInitConfig struct {
	ReportURL string            `json:"REPORT_SERVER_URL"`
	Headers   map[string]string `json:"defaultHeaders"`
}

const (
	EnvOpenAIBaseURL        = "OPENAI_BASE_URL"
	EnvOpenAIAPIKey         = "OPENAI_API_KEY"
	EnvModelName            = "LLM_MODEL_NAME"
	EnvOpenAIInitConfigJSON = "OPENAI_INIT_CONFIG_JSON"
)

func checkEnvLLM() error {
	if err := config.LoadEnv(); err != nil {
		return errors.Wrap(code.LoadEnvError, err.Error())
	}
	openaiBaseURL := os.Getenv("OPENAI_BASE_URL")
	if openaiBaseURL == "" {
		return errors.Wrap(code.LLMEnvMissedError, "OPENAI_BASE_URL missed")
	}
	log.Info().Str("OPENAI_BASE_URL", openaiBaseURL).Msg("get env")
	openaiAPIKey := os.Getenv("OPENAI_API_KEY")
	if openaiAPIKey == "" {
		return errors.Wrap(code.LLMEnvMissedError, "OPENAI_API_KEY missed")
	}
	log.Info().Str("OPENAI_API_KEY", maskAPIKey(openaiAPIKey)).Msg("get env")
	modelName := os.Getenv("LLM_MODEL_NAME")
	if modelName == "" {
		return errors.Wrap(code.LLMEnvMissedError, "LLM_MODEL_NAME missed")
	}
	log.Info().Str("LLM_MODEL_NAME", modelName).Msg("get env")
	return nil
}

// CustomTransport is a custom RoundTripper that adds headers to every request
type CustomTransport struct {
	Transport http.RoundTripper
	Headers   map[string]string
}

// RoundTrip executes a single HTTP transaction and adds custom headers
func (c *CustomTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for key, value := range c.Headers {
		req.Header.Set(key, value)
	}
	return c.Transport.RoundTrip(req)
}

type OutputFormat struct {
	Thought string `json:"thought"`
	Action  string `json:"action"`
	Error   string `json:"error,omitempty"`
}

// GetModelConfig get OpenAI config
func GetModelConfig() (*openai.ChatModelConfig, error) {
	if err := checkEnvLLM(); err != nil {
		log.Error().Err(err).Msg("check LLM env failed")
		return nil, err
	}
	envConfig := &OpenAIInitConfig{
		Headers: make(map[string]string),
	}

	// read from JSON config first
	jsonStr := config.GetEnvConfig(EnvOpenAIInitConfigJSON)
	if jsonStr != "" {
		if err := json.Unmarshal([]byte(jsonStr), envConfig); err != nil {
			return nil, err
		}
	}

	// outputFormatSchema, err := openapi3gen.NewSchemaRefForValue(&OutputFormat{}, nil)
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("NewSchemaRefForValue failed")
	// }

	modelConfig := &openai.ChatModelConfig{
		HTTPClient: &http.Client{
			Timeout: defaultTimeout,
			Transport: &CustomTransport{
				Transport: http.DefaultTransport,
				Headers:   envConfig.Headers,
			},
		},
		// TODO: set structured response format
		// https://github.com/cloudwego/eino-ext/blob/main/components/model/openai/examples/structured/structured.go
		// ResponseFormat: &openai2.ChatCompletionResponseFormat{
		// 	Type: openai2.ChatCompletionResponseFormatTypeJSONSchema,
		// 	JSONSchema: &openai2.ChatCompletionResponseFormatJSONSchema{
		// 		Name:        "thought_and_action",
		// 		Description: "data that describes planning thought and action",
		// 		Schema:      outputFormatSchema.Value,
		// 		Strict:      false,
		// 	},
		// },
	}

	if baseURL := config.GetEnvConfig(EnvOpenAIBaseURL); baseURL != "" {
		modelConfig.BaseURL = baseURL
	} else {
		return nil, fmt.Errorf("miss env %s", EnvOpenAIBaseURL)
	}

	if apiKey := config.GetEnvConfig(EnvOpenAIAPIKey); apiKey != "" {
		modelConfig.APIKey = apiKey
	} else {
		return nil, fmt.Errorf("miss env %s", EnvOpenAIAPIKey)
	}

	if modelName := config.GetEnvConfig(EnvModelName); modelName != "" {
		modelConfig.Model = modelName
	} else {
		return nil, fmt.Errorf("miss env %s", EnvModelName)
	}

	// log config info
	log.Info().Str("model", modelConfig.Model).
		Str("baseURL", modelConfig.BaseURL).
		Str("apiKey", maskAPIKey(modelConfig.APIKey)).
		Str("timeout", defaultTimeout.String()).
		Msg("get model config")

	return modelConfig, nil
}

// maskAPIKey masks the API key
func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "******"
	}

	return key[:4] + "******" + key[len(key)-4:]
}
