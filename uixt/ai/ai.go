package ai

import (
	"os"

	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/code"
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
	LLMServiceTypeGPT4o      LLMServiceType = "gpt-4o"
	LLMServiceTypeDeepSeekV3 LLMServiceType = "deepseek-v3"
)

func WithLLMService(service LLMServiceType) AIServiceOption {
	return func(opts *AIServices) {
		if service == LLMServiceTypeGPT4o {
			var err error
			opts.ILLMService, err = NewGPT4oLLMService()
			if err != nil {
				log.Error().Err(err).Msg("init gpt-4o llm service failed")
				os.Exit(code.GetErrorCode(err))
			}
		}
	}
}
