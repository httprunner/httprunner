package ai

import (
	"context"

	"github.com/httprunner/httprunner/v5/uixt/option"
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
