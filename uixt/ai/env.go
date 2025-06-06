package ai

import (
	"os"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/config"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// LLM model config env variables
const (
	EnvOpenAIBaseURL = "OPENAI_BASE_URL"
	EnvOpenAIAPIKey  = "OPENAI_API_KEY"
	EnvModelName     = "LLM_MODEL_NAME"
)

const (
	defaultTimeout = 30 * time.Second
)

// GetModelConfig get OpenAI config
func GetModelConfig(modelType option.LLMServiceType) (*ModelConfig, error) {
	if err := config.LoadEnv(); err != nil {
		return nil, errors.Wrap(code.LoadEnvError, err.Error())
	}

	baseURL, apiKey, modelName, err := getModelConfigFromEnv(modelType)
	if err != nil {
		return nil, errors.Wrap(code.LLMEnvMissedError, err.Error())
	}

	// https://www.volcengine.com/docs/82379/1536429
	temperature := float32(0)
	topP := float32(0.7)
	modelConfig := &openai.ChatModelConfig{
		BaseURL:     baseURL,
		APIKey:      apiKey,
		Model:       modelName,
		Timeout:     defaultTimeout,
		Temperature: &temperature,
		TopP:        &topP,
	}

	// log config info
	log.Info().Str("model", modelConfig.Model).
		Str("baseURL", modelConfig.BaseURL).
		Str("apiKey", maskAPIKey(modelConfig.APIKey)).
		Str("timeout", defaultTimeout.String()).
		Str("serviceType", string(modelType)).
		Msg("get model config")

	return &ModelConfig{
		ChatModelConfig: modelConfig,
		ModelType:       modelType,
	}, nil
}

type ModelConfig struct {
	*openai.ChatModelConfig
	ModelType option.LLMServiceType
}

// getServiceEnvPrefix converts LLMServiceType to environment variable prefix
// e.g., "doubao-1.5-thinking-vision-pro-250428" -> "DOUBAO_1_5_THINKING_VISION_PRO_250428"
func getServiceEnvPrefix(modelType option.LLMServiceType) string {
	// Convert service name to uppercase and replace hyphens and dots with underscores
	prefix := strings.ToUpper(string(modelType))
	prefix = strings.ReplaceAll(prefix, "-", "_")
	prefix = strings.ReplaceAll(prefix, ".", "_")
	return prefix
}

// getModelConfigFromEnv retrieves model configuration from environment variables
// It first tries to get service-specific config, then falls back to default config
// Model name is derived from the service type, no need for separate MODEL_NAME env var
func getModelConfigFromEnv(modelType option.LLMServiceType) (baseURL, apiKey, modelName string, err error) {
	servicePrefix := getServiceEnvPrefix(modelType)

	// Try to get service-specific configuration first
	baseURL = os.Getenv(servicePrefix + "_BASE_URL")
	apiKey = os.Getenv(servicePrefix + "_API_KEY")

	// Model name is derived from the service type itself
	modelName = string(modelType)

	envBaseURL := os.Getenv(EnvOpenAIBaseURL)
	envAPIKey := os.Getenv(EnvOpenAIAPIKey)

	// If service-specific config is not found, fall back to default config
	if baseURL == "" {
		baseURL = envBaseURL
	}
	if apiKey == "" {
		apiKey = envAPIKey
	}

	// If we're using default config completely (both base URL and API key from default),
	// then use default model name if available
	if baseURL == envBaseURL && apiKey == envAPIKey {
		defaultModelName := os.Getenv(EnvModelName)
		if defaultModelName != "" {
			modelName = defaultModelName
		}
	}

	// Check if all required configs are available
	if baseURL == "" {
		return "", "", "", errors.Errorf("env %s or %s missed", servicePrefix+"_BASE_URL", EnvOpenAIBaseURL)
	}
	if apiKey == "" {
		return "", "", "", errors.Errorf("env %s or %s missed", servicePrefix+"_API_KEY", EnvOpenAIAPIKey)
	}

	return baseURL, apiKey, modelName, nil
}

// maskAPIKey masks the API key
func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "******"
	}

	return key[:4] + "******" + key[len(key)-4:]
}
