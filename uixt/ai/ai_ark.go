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
	temperature := float32(0.01) // [0, 2] 采样温度。控制了生成文本时对每个候选词的概率分布进行平滑的程度。
	// topP := float32(0.7)           // [0, 1] 核采样概率阈值。模型会考虑概率质量在 top_p 内的 token 结果。
	// maxTokens := int(4096)         // 模型可以生成的最大 token 数量。输入 token 和输出 token 的总长度还受模型的上下文长度限制。
	// frequencyPenalty := float32(0) // [-2, 2] 频率惩罚系数。如果值为正，会根据新 token 在文本中的出现频率对其进行惩罚，从而降低模型逐字重复的可能性。

	modelConfig := &ark.ChatModelConfig{
		BaseURL:     arkBaseURL,
		APIKey:      arkAPIKey,
		Model:       modelName,
		Timeout:     &timeout,
		Temperature: &temperature,
		// TopP:             &topP,
		// MaxTokens:        &maxTokens,
		// FrequencyPenalty: &frequencyPenalty,
	}

	// log config info
	log.Info().Str("model", modelConfig.Model).
		Str("baseURL", modelConfig.BaseURL).
		Str("apiKey", maskAPIKey(modelConfig.APIKey)).
		Str("timeout", defaultTimeout.String()).
		Msg("get model config")

	return modelConfig, nil
}
