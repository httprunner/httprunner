package ai

import (
	"context"
	"os"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/config"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// ILLMService 定义了 LLM 服务接口，包括规划和断言功能
type ILLMService interface {
	Call(ctx context.Context, opts *PlanningOptions) (*PlanningResult, error)
	Assert(ctx context.Context, opts *AssertOptions) (*AssertionResult, error)
}

func NewLLMService(modelType option.LLMServiceType) (ILLMService, error) {
	modelConfig, err := GetModelConfig(modelType)
	if err != nil {
		return nil, err
	}

	planner, err := NewPlanner(context.Background(), modelConfig)
	if err != nil {
		return nil, err
	}
	asserter, err := NewAsserter(context.Background(), modelConfig)
	if err != nil {
		return nil, err
	}

	return &combinedLLMService{
		planner:  planner,
		asserter: asserter,
	}, nil
}

// combinedLLMService 实现了 ILLMService 接口，组合了规划和断言功能
// ⭐️支持采用不同的模型服务进行规划和断言
type combinedLLMService struct {
	planner  IPlanner  // 提供规划功能
	asserter IAsserter // 提供断言功能
}

// Call 执行规划功能
func (c *combinedLLMService) Call(ctx context.Context, opts *PlanningOptions) (*PlanningResult, error) {
	return c.planner.Call(ctx, opts)
}

// Assert 执行断言功能
func (c *combinedLLMService) Assert(ctx context.Context, opts *AssertOptions) (*AssertionResult, error) {
	return c.asserter.Assert(ctx, opts)
}

// LLM model config env variables
const (
	EnvOpenAIBaseURL = "OPENAI_BASE_URL"
	EnvOpenAIAPIKey  = "OPENAI_API_KEY"
	EnvModelName     = "LLM_MODEL_NAME"
)

const (
	defaultTimeout = 30 * time.Second
)

type ModelConfig struct {
	*openai.ChatModelConfig
	ModelType option.LLMServiceType
}

// GetModelConfig get OpenAI config
func GetModelConfig(modelType option.LLMServiceType) (*ModelConfig, error) {
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

	maxTokens := 4096
	temperature := float32(0.7)
	modelConfig := &openai.ChatModelConfig{
		BaseURL:     openaiBaseURL,
		APIKey:      openaiAPIKey,
		Model:       modelName,
		Timeout:     defaultTimeout,
		MaxTokens:   &maxTokens,
		Temperature: &temperature,
	}

	// log config info
	log.Info().Str("model", modelConfig.Model).
		Str("baseURL", modelConfig.BaseURL).
		Str("apiKey", maskAPIKey(modelConfig.APIKey)).
		Str("timeout", defaultTimeout.String()).
		Msg("get model config")

	return &ModelConfig{
		ChatModelConfig: modelConfig,
		ModelType:       modelType,
	}, nil
}

// maskAPIKey masks the API key
func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "******"
	}

	return key[:4] + "******" + key[len(key)-4:]
}
