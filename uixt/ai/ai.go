package ai

import (
	"context"
	"os"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/config"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func NewAIService(opts ...AIServiceOption) *AIServices {
	services := &AIServices{}
	for _, option := range opts {
		option(services)
	}
	return services
}

type AIServices struct {
	ICVService
	ILLMService
}

type AIServiceOption func(*AIServices)

type CVServiceType string

const (
	CVServiceTypeVEDEM  CVServiceType = "vedem"
	CVServiceTypeOpenCV CVServiceType = "opencv"
)

func WithCVService(service CVServiceType) AIServiceOption {
	return func(opts *AIServices) {
		if service == CVServiceTypeVEDEM {
			var err error
			opts.ICVService, err = NewVEDEMImageService()
			if err != nil {
				log.Error().Err(err).Msg("init vedem image service failed")
				os.Exit(code.GetErrorCode(err))
			}
		}
	}
}

type LLMServiceType string

const (
	LLMServiceTypeUITARS LLMServiceType = "ui-tars"
	LLMServiceTypeGPT    LLMServiceType = "gpt"
	LLMServiceTypeQwenVL LLMServiceType = "qwen-vl"
)

// ILLMService 定义了 LLM 服务接口，包括规划和断言功能
type ILLMService interface {
	Call(opts *PlanningOptions) (*PlanningResult, error)
	Assert(opts *AssertOptions) (*AssertionResponse, error)
}

func WithLLMService(modelType LLMServiceType) AIServiceOption {
	return func(opts *AIServices) {
		// init planner
		var planner IPlanner
		var err error
		switch modelType {
		case LLMServiceTypeGPT:
			// TODO: implement gpt-4o planner and asserter
			planner, err = NewPlanner(context.Background())
		case LLMServiceTypeUITARS:
			planner, err = NewUITarsPlanner(context.Background())
		}
		if err != nil {
			log.Error().Err(err).Msgf("init %s planner failed", modelType)
			os.Exit(code.GetErrorCode(err))
		}

		// init asserter
		asserter, err := NewAsserter(context.Background())
		if err != nil {
			log.Error().Err(err).Msg("init asserter failed")
			os.Exit(code.GetErrorCode(err))
		}

		opts.ILLMService = &combinedLLMService{
			planner:  planner,
			asserter: asserter,
		}
	}
}

// combinedLLMService 实现了 ILLMService 接口，组合了规划和断言功能
type combinedLLMService struct {
	planner  IPlanner  // 提供规划功能
	asserter IAsserter // 提供断言功能
}

// Call 执行规划功能
func (c *combinedLLMService) Call(opts *PlanningOptions) (*PlanningResult, error) {
	return c.planner.Call(opts)
}

// Assert 执行断言功能
func (c *combinedLLMService) Assert(opts *AssertOptions) (*AssertionResponse, error) {
	return c.asserter.Assert(opts)
}

// LLM model config env variables
const (
	EnvOpenAIBaseURL = "OPENAI_BASE_URL"
	EnvOpenAIAPIKey  = "OPENAI_API_KEY"
	EnvModelName     = "LLM_MODEL_NAME"
)

var EnvModelUse string

// GetOpenAIModelConfig get OpenAI config
func GetOpenAIModelConfig() (*openai.ChatModelConfig, error) {
	if err := config.LoadEnv(); err != nil {
		return nil, errors.Wrap(code.LoadEnvError, err.Error())
	}
	EnvModelUse = os.Getenv("LLM_MODEL_USE")

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

	temperature := float32(0.01)
	modelConfig := &openai.ChatModelConfig{
		BaseURL:     openaiBaseURL,
		APIKey:      openaiAPIKey,
		Model:       modelName,
		Timeout:     defaultTimeout,
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
