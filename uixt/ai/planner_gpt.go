package ai

import (
	"context"
	"fmt"
	_ "image/jpeg"
	"os"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	openai2 "github.com/cloudwego/eino-ext/libs/acl/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/getkin/kin-openapi/openapi3gen"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/config"
	"github.com/httprunner/httprunner/v5/internal/json"
	"github.com/httprunner/httprunner/v5/uixt/types"
)

const (
	EnvOpenAIBaseURL = "OPENAI_BASE_URL"
	EnvOpenAIAPIKey  = "OPENAI_API_KEY"
	EnvModelName     = "LLM_MODEL_NAME"
)

// GetOpenAIModelConfig get OpenAI config
func GetOpenAIModelConfig() (*openai.ChatModelConfig, error) {
	if err := config.LoadEnv(); err != nil {
		return nil, errors.Wrap(code.LoadEnvError, err.Error())
	}

	openaiBaseURL := os.Getenv(EnvOpenAIBaseURL)
	if openaiBaseURL == "" {
		return nil, errors.Wrapf(code.LLMEnvMissedError,
			"env %s missed", EnvOpenAIBaseURL)
	}
	openaiAPIKey := os.Getenv(EnvOpenAIAPIKey)
	if openaiAPIKey == "" {
		return nil, errors.Wrapf(code.LLMEnvMissedError,
			"env %s missed", EnvOpenAIAPIKey)
	}
	modelName := os.Getenv(EnvModelName)
	if modelName == "" {
		return nil, errors.Wrapf(code.LLMEnvMissedError,
			"env %s missed", EnvModelName)
	}

	type OutputFormat struct {
		Thought string `json:"thought"`
		Action  string `json:"action"`
		Error   string `json:"error,omitempty"`
	}
	outputFormatSchema, err := openapi3gen.NewSchemaRefForValue(&OutputFormat{}, nil)
	if err != nil {
		return nil, err
	}

	modelConfig := &openai.ChatModelConfig{
		BaseURL: openaiBaseURL,
		APIKey:  openaiAPIKey,
		Model:   modelName,
		Timeout: defaultTimeout,
		// set structured response format
		// https://github.com/cloudwego/eino-ext/blob/main/components/model/openai/examples/structured/structured.go
		ResponseFormat: &openai2.ChatCompletionResponseFormat{
			Type: openai2.ChatCompletionResponseFormatTypeJSONSchema,
			JSONSchema: &openai2.ChatCompletionResponseFormatJSONSchema{
				Name:        "thought_and_action",
				Description: "data that describes planning thought and action",
				Schema:      outputFormatSchema.Value,
				Strict:      false,
			},
		},
	}

	// log config info
	log.Info().Str("model", modelConfig.Model).
		Str("baseURL", modelConfig.BaseURL).
		Str("apiKey", maskAPIKey(modelConfig.APIKey)).
		Str("timeout", defaultTimeout.String()).
		Msg("get model config")

	return modelConfig, nil
}

func NewPlanner(ctx context.Context) (*Planner, error) {
	config, err := GetOpenAIModelConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenAI config: %w", err)
	}
	model, err := openai.NewChatModel(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize OpenAI model: %w", err)
	}
	return &Planner{
		ctx:          ctx,
		config:       config,
		model:        model,
		systemPrompt: uiTarsPlanningPrompt, // TODO: change prompt with function calling
	}, nil
}

type Planner struct {
	ctx          context.Context
	model        model.ToolCallingChatModel
	config       *openai.ChatModelConfig
	systemPrompt string
	history      ConversationHistory
}

// Call performs UI planning using Vision Language Model
func (p *Planner) Call(opts *PlanningOptions) (*PlanningResult, error) {
	// validate input parameters
	if err := validateInput(opts); err != nil {
		return nil, errors.Wrap(err, "validate input parameters failed")
	}

	// prepare prompt
	if len(p.history) == 0 {
		// add system message
		systemPrompt := uiTarsPlanningPrompt + opts.UserInstruction
		p.history = ConversationHistory{
			{
				Role:    schema.System,
				Content: systemPrompt,
			},
		}
	}
	// append user image message
	p.history.Append(opts.Message)

	// call model service, generate response
	logRequest(p.history)
	startTime := time.Now()
	resp, err := p.model.Generate(p.ctx, p.history)
	log.Info().Float64("elapsed(s)", time.Since(startTime).Seconds()).
		Str("model", p.config.Model).Msg("call model service")
	if err != nil {
		return nil, fmt.Errorf("request model service failed: %w", err)
	}
	logResponse(resp)

	// parse result
	result, err := p.parseResult(resp, opts.Size)
	if err != nil {
		return nil, errors.Wrap(err, "parse result failed")
	}

	// append assistant message
	p.history.Append(&schema.Message{
		Role:    schema.Assistant,
		Content: result.ActionSummary,
	})

	return result, nil
}

func (p *Planner) parseResult(msg *schema.Message, size types.Size) (*PlanningResult, error) {
	// parse JSON format, from VLM like openai/gpt-4o
	parseActions, jsonErr := parseJSON(msg.Content)
	if jsonErr != nil {
		return nil, jsonErr
	}

	// process response
	result, err := processVLMResponse(parseActions, size)
	if err != nil {
		return nil, errors.Wrap(err, "process VLM response failed")
	}

	log.Info().
		Interface("summary", result.ActionSummary).
		Interface("actions", result.NextActions).
		Msg("get VLM planning result")
	return result, nil
}

// parseJSON tries to parse the response as JSON format
func parseJSON(predictionText string) ([]ParsedAction, error) {
	predictionText = strings.TrimSpace(predictionText)
	if strings.HasPrefix(predictionText, "```json") && strings.HasSuffix(predictionText, "```") {
		predictionText = strings.TrimPrefix(predictionText, "```json")
		predictionText = strings.TrimSuffix(predictionText, "```")
	}
	predictionText = strings.TrimSpace(predictionText)

	var response PlanningResult
	if err := json.Unmarshal([]byte(predictionText), &response); err != nil {
		return nil, fmt.Errorf("failed to parse VLM response: %v", err)
	}

	if response.Error != "" {
		return nil, errors.New(response.Error)
	}

	if len(response.NextActions) == 0 {
		return nil, errors.New("no actions returned from VLM")
	}

	// normalize actions
	var normalizedActions []ParsedAction
	for i := range response.NextActions {
		// create a new variable, avoid implicit memory aliasing in for loop.
		action := response.NextActions[i]
		if err := normalizeAction(&action); err != nil {
			return nil, errors.Wrap(err, "failed to normalize action")
		}
		normalizedActions = append(normalizedActions, action)
	}

	return normalizedActions, nil
}

// normalizeAction normalizes the coordinates in the action
func normalizeAction(action *ParsedAction) error {
	switch action.ActionType {
	case "click", "drag":
		// handle click and drag action coordinates
		if startBox, ok := action.ActionInputs["startBox"].(string); ok {
			normalized, err := normalizeCoordinates(startBox)
			if err != nil {
				return fmt.Errorf("failed to normalize startBox: %w", err)
			}
			action.ActionInputs["startBox"] = normalized
		}

		if endBox, ok := action.ActionInputs["endBox"].(string); ok {
			normalized, err := normalizeCoordinates(endBox)
			if err != nil {
				return fmt.Errorf("failed to normalize endBox: %w", err)
			}
			action.ActionInputs["endBox"] = normalized
		}
	}

	return nil
}
