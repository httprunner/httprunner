package ai

import (
	"os"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/config"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

const (
	EnvArkBaseURL = "ARK_BASE_URL"
	EnvArkAPIKey  = "ARK_API_KEY"
	EnvArkModelID = "ARK_MODEL_ID"
)

func GetArkModelConfig() (*ark.ChatModelConfig, error) {
	if err := config.LoadEnv(); err != nil {
		return nil, errors.Wrap(code.LoadEnvError, err.Error())
	}

	arkBaseURL := os.Getenv(EnvArkBaseURL)
	arkAPIKey := os.Getenv(EnvArkAPIKey)
	if arkAPIKey == "" {
		return nil, errors.Wrapf(code.LLMEnvMissedError,
			"env %s missed", EnvArkAPIKey)
	}
	modelName := os.Getenv(EnvArkModelID)
	if modelName == "" {
		return nil, errors.Wrapf(code.LLMEnvMissedError,
			"env %s missed", EnvArkModelID)
	}
	timeout := defaultTimeout

	// https://www.volcengine.com/docs/82379/1494384?redirect=1
	temperature := float32(0.01)
	modelConfig := &ark.ChatModelConfig{
		BaseURL:     arkBaseURL,
		APIKey:      arkAPIKey,
		Model:       modelName,
		Timeout:     &timeout,
		Temperature: &temperature,
	}

	// log config info
	log.Info().Str("model", modelConfig.Model).
		Str("baseURL", modelConfig.BaseURL).
		Str("apiKey", maskAPIKey(modelConfig.APIKey)).
		Str("timeout", defaultTimeout.String()).
		Msg("get model config")

	return modelConfig, nil
}
