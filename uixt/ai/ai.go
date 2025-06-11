package ai

import (
	"context"

	"github.com/cloudwego/eino/schema"
	"github.com/httprunner/httprunner/v5/uixt/option"
)

// ILLMService 定义了 LLM 服务接口，包括规划、断言和查询功能
type ILLMService interface {
	Plan(ctx context.Context, opts *PlanningOptions) (*PlanningResult, error)
	Assert(ctx context.Context, opts *AssertOptions) (*AssertionResult, error)
	Query(ctx context.Context, opts *QueryOptions) (*QueryResult, error)
	// RegisterTools registers tools for function calling
	RegisterTools(tools []*schema.ToolInfo) error
}

// NewLLMService creates a new LLM service with the same model for all components (backward compatibility)
func NewLLMService(modelType option.LLMServiceType) (ILLMService, error) {
	config := option.NewLLMServiceConfig(modelType)
	return NewLLMServiceWithOptionConfig(config)
}

// NewLLMServiceWithOptionConfig creates a new LLM service with different models for each component
func NewLLMServiceWithOptionConfig(config *option.LLMServiceConfig) (ILLMService, error) {
	// Get model configs for each component
	plannerModelConfig, err := GetModelConfig(config.PlannerModel)
	if err != nil {
		return nil, err
	}

	asserterModelConfig, err := GetModelConfig(config.AsserterModel)
	if err != nil {
		return nil, err
	}

	querierModelConfig, err := GetModelConfig(config.QuerierModel)
	if err != nil {
		return nil, err
	}

	// Create components with their respective model configs
	planner, err := NewPlanner(context.Background(), plannerModelConfig)
	if err != nil {
		return nil, err
	}

	asserter, err := NewAsserter(context.Background(), asserterModelConfig)
	if err != nil {
		return nil, err
	}

	querier, err := NewQuerier(context.Background(), querierModelConfig)
	if err != nil {
		return nil, err
	}

	return &combinedLLMService{
		planner:  planner,
		asserter: asserter,
		querier:  querier,
	}, nil
}

// combinedLLMService 实现了 ILLMService 接口，组合了规划、断言和查询功能
// ⭐️支持采用不同的模型服务进行规划、断言和查询
type combinedLLMService struct {
	planner  IPlanner  // 提供规划功能
	asserter IAsserter // 提供断言功能
	querier  IQuerier  // 提供查询功能
}

// Plan 执行规划功能
func (c *combinedLLMService) Plan(ctx context.Context, opts *PlanningOptions) (*PlanningResult, error) {
	return c.planner.Plan(ctx, opts)
}

// Assert 执行断言功能
func (c *combinedLLMService) Assert(ctx context.Context, opts *AssertOptions) (*AssertionResult, error) {
	return c.asserter.Assert(ctx, opts)
}

// Query 执行查询功能
func (c *combinedLLMService) Query(ctx context.Context, opts *QueryOptions) (*QueryResult, error) {
	return c.querier.Query(ctx, opts)
}

// RegisterTools registers tools for function calling
func (c *combinedLLMService) RegisterTools(tools []*schema.ToolInfo) error {
	// Only register tools to planner since asserter and querier don't need tools
	if planner, ok := c.planner.(*Planner); ok {
		return planner.RegisterTools(tools)
	}
	return nil
}
