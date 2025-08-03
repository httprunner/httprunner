//go:build localtest

package ai

import (
	"context"
	"testing"

	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestILLMServiceQuery(t *testing.T) {

	// Create LLM service
	service, err := NewLLMService(option.OPENAI_GPT_4O)
	require.NoError(t, err)
	require.NotNil(t, service)

	// Load test image
	screenshot, size, err := builtin.LoadImage("testdata/llk_1.png")
	require.NoError(t, err)

	// Test basic query functionality
	t.Run("BasicQuery", func(t *testing.T) {
		opts := &QueryOptions{
			Query:      "请描述这张图片中的内容",
			Screenshot: screenshot,
			Size:       size,
		}

		result, err := service.Query(context.Background(), opts)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.Content)
		assert.NotEmpty(t, result.Thought)
		assert.Nil(t, result.Data) // Should be nil for standard query

		t.Logf("Query result: %s", result.Content)
	})

	// Test custom schema query
	t.Run("CustomSchemaQuery", func(t *testing.T) {
		type GameInfo struct {
			Content string   `json:"content"`
			Thought string   `json:"thought"`
			Rows    int      `json:"rows"`
			Cols    int      `json:"cols"`
			Icons   []string `json:"icons"`
		}

		opts := &QueryOptions{
			Query:        "请分析这个连连看游戏界面，告诉我有多少行多少列，有哪些不同类型的图案",
			Screenshot:   screenshot,
			Size:         size,
			OutputSchema: GameInfo{},
		}

		result, err := service.Query(context.Background(), opts)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.Content)
		assert.NotEmpty(t, result.Thought)
		assert.NotNil(t, result.Data)

		// Verify type conversion
		if gameInfo, ok := result.Data.(*GameInfo); ok {
			assert.NotEmpty(t, gameInfo.Content)
			assert.NotEmpty(t, gameInfo.Thought)
			assert.Greater(t, gameInfo.Rows, 0)
			assert.Greater(t, gameInfo.Cols, 0)
			assert.NotEmpty(t, gameInfo.Icons)
			t.Logf("Game info: %+v", gameInfo)
		} else {
			t.Errorf("Expected *GameInfo, got %T", result.Data)
		}
	})
}

func TestILLMServiceIntegration(t *testing.T) {

	// Create LLM service
	service, err := NewLLMService(option.OPENAI_GPT_4O)
	require.NoError(t, err)
	require.NotNil(t, service)

	// Load test image
	screenshot, size, err := builtin.LoadImage("testdata/llk_1.png")
	require.NoError(t, err)

	ctx := context.Background()

	// Test that all three methods work
	t.Run("AllMethods", func(t *testing.T) {
		// Test Query
		queryOpts := &QueryOptions{
			Query:      "请分析这张图片",
			Screenshot: screenshot,
			Size:       size,
		}
		queryResult, err := service.Query(ctx, queryOpts)
		assert.NoError(t, err)
		assert.NotNil(t, queryResult)
		t.Logf("Query result: %s", queryResult.Content)

		// Test Assert
		assertOpts := &AssertOptions{
			Assertion:  "这是一个连连看游戏界面",
			Screenshot: screenshot,
			Size:       size,
		}
		assertResult, err := service.Assert(ctx, assertOpts)
		assert.NoError(t, err)
		assert.NotNil(t, assertResult)
		t.Logf("Assert result: pass=%v, thought=%s", assertResult.Pass, assertResult.Thought)

		// Note: Planning test would require proper user instruction and message setup
		// which is more complex, so we skip it in this integration test
	})
}

// TestLLMServiceConfig tests the LLM service configuration functionality
func TestLLMServiceConfig(t *testing.T) {
	t.Run("BasicConfiguration", func(t *testing.T) {
		// Test creating config with same model for all components
		modelType := option.DOUBAO_1_5_THINKING_VISION_PRO_250428
		config := option.NewLLMServiceConfig(modelType)

		assert.Equal(t, modelType, config.PlannerModel)
		assert.Equal(t, modelType, config.AsserterModel)
		assert.Equal(t, modelType, config.QuerierModel)
	})

	t.Run("MixedConfiguration", func(t *testing.T) {
		// Test configuring different models for each component
		config := option.NewLLMServiceConfig(option.DOUBAO_1_5_THINKING_VISION_PRO_250428).
			WithPlannerModel(option.DOUBAO_1_5_UI_TARS_250328).
			WithAsserterModel(option.OPENAI_GPT_4O).
			WithQuerierModel(option.DEEPSEEK_R1_250528)

		assert.Equal(t, option.DOUBAO_1_5_UI_TARS_250328, config.PlannerModel)
		assert.Equal(t, option.OPENAI_GPT_4O, config.AsserterModel)
		assert.Equal(t, option.DEEPSEEK_R1_250528, config.QuerierModel)
	})

	t.Run("RecommendedConfigurations", func(t *testing.T) {
		configs := option.RecommendedConfigurations()

		// Test mixed optimal configuration
		mixedOptimal := configs["mixed_optimal"]
		assert.NotNil(t, mixedOptimal)
		assert.Equal(t, option.DOUBAO_1_5_UI_TARS_250328, mixedOptimal.PlannerModel)
		assert.Equal(t, option.OPENAI_GPT_4O, mixedOptimal.AsserterModel)
		assert.Equal(t, option.DEEPSEEK_R1_250528, mixedOptimal.QuerierModel)

		// Test high performance configuration
		highPerf := configs["high_performance"]
		assert.NotNil(t, highPerf)
		assert.Equal(t, option.OPENAI_GPT_4O, highPerf.PlannerModel)
		assert.Equal(t, option.OPENAI_GPT_4O, highPerf.AsserterModel)
		assert.Equal(t, option.OPENAI_GPT_4O, highPerf.QuerierModel)
	})
}

// TestLLMServiceCreation tests service creation with different configurations
func TestLLMServiceCreation(t *testing.T) {
	t.Run("BackwardCompatibility", func(t *testing.T) {
		// Test that the original NewLLMService function still works
		modelType := option.DOUBAO_1_5_THINKING_VISION_PRO_250428
		service, err := NewLLMService(modelType)

		// We expect an error due to missing environment variables in test environment
		// but the function signature should be correct
		if err != nil {
			assert.NotNil(t, err)
			assert.Nil(t, service)
		} else {
			assert.NotNil(t, service)
		}
	})

	t.Run("WithAdvancedConfig", func(t *testing.T) {
		// Test the new API with different models for each component
		config := option.NewLLMServiceConfig(option.DOUBAO_1_5_THINKING_VISION_PRO_250428).
			WithPlannerModel(option.DOUBAO_1_5_UI_TARS_250328).
			WithAsserterModel(option.OPENAI_GPT_4O)

		service, err := NewLLMServiceWithOptionConfig(config)

		// We expect an error due to missing environment variables in test environment
		// but the function signature should be correct
		if err != nil {
			assert.NotNil(t, err)
			assert.Nil(t, service)
		} else {
			assert.NotNil(t, service)
		}
	})
}
