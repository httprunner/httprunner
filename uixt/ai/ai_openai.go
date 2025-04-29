package ai

import (
	"os"

	"github.com/cloudwego/eino-ext/components/model/openai"
	openai2 "github.com/cloudwego/eino-ext/libs/acl/openai"
	"github.com/getkin/kin-openapi/openapi3gen"
	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/config"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
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
