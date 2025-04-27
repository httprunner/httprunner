package ai

import (
	"context"
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
	LLMServiceTypeUITARS     LLMServiceType = "ui-tars"
	LLMServiceTypeGPT4o      LLMServiceType = "gpt-4o"
	LLMServiceTypeDeepSeekV3 LLMServiceType = "deepseek-v3"
)

// ILLMService 定义了 LLM 服务接口，包括规划和断言功能
type ILLMService interface {
	Call(opts *PlanningOptions) (*PlanningResult, error)
	Assert(opts *AssertOptions) (*AssertionResponse, error)
}

func WithLLMService(service LLMServiceType) AIServiceOption {
	return func(opts *AIServices) {
		switch service {
		case LLMServiceTypeGPT4o:
			planner, err := NewPlanner(context.Background())
			if err != nil {
				log.Error().Err(err).Msg("init gpt-4o planner failed")
				os.Exit(code.GetErrorCode(err))
			}

			asserter, err := NewUITarsAsserter(context.Background())
			if err != nil {
				log.Error().Err(err).Msg("init ui-tars asserter failed")
				os.Exit(code.GetErrorCode(err))
			}

			opts.ILLMService = &combinedLLMService{
				planner:  planner,
				asserter: asserter,
			}

		case LLMServiceTypeUITARS:
			planner, err := NewUITarsPlanner(context.Background())
			if err != nil {
				log.Error().Err(err).Msg("init ui-tars planner failed")
				os.Exit(code.GetErrorCode(err))
			}

			asserter, err := NewUITarsAsserter(context.Background())
			if err != nil {
				log.Error().Err(err).Msg("init ui-tars asserter failed")
				os.Exit(code.GetErrorCode(err))
			}

			opts.ILLMService = &combinedLLMService{
				planner:  planner,
				asserter: asserter,
			}
		}
	}
}

// combinedLLMService 实现了 ILLMService 接口，组合了规划和断言功能
type combinedLLMService struct {
	planner  IPlanner  // 提供规划功能
	asserter IAsserter // 提供断言功能
}

// IPlanner 定义了规划功能接口
type IPlanner interface {
	Call(opts *PlanningOptions) (*PlanningResult, error)
}

// IAsserter 定义了断言功能接口
type IAsserter interface {
	Assert(opts *AssertOptions) (*AssertionResponse, error)
}

// Call 执行规划功能
func (c *combinedLLMService) Call(opts *PlanningOptions) (*PlanningResult, error) {
	return c.planner.Call(opts)
}

// Assert 执行断言功能
func (c *combinedLLMService) Assert(opts *AssertOptions) (*AssertionResponse, error) {
	return c.asserter.Assert(opts)
}
