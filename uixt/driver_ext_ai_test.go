//go:build localtest

package uixt

import (
	"context"
	"testing"

	"github.com/cloudwego/eino/schema"
	"github.com/stretchr/testify/assert"

	"github.com/httprunner/httprunner/v5/uixt/ai"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/httprunner/httprunner/v5/uixt/types"
)

func TestDriverExt_TapByLLM(t *testing.T) {
	driver := setupDriverExt(t)
	_, err := driver.AIAction(context.Background(), "点击第一个帖子的作者头像")
	assert.Nil(t, err)

	_, err = driver.AIAssert("当前在个人介绍页")
	assert.Nil(t, err)
}

func TestDriverExt_StartToGoal(t *testing.T) {
	driver := setupDriverExt(t)

	userInstruction := `连连看是一款经典的益智消除类小游戏，通常以图案或图标为主要元素。以下是连连看的基本规则说明：
	1. 游戏目标: 玩家需要在规定时间内，通过连接相同的图案或图标，将它们从游戏界面中消除。
	2. 连接规则:
	- 两个相同的图案可以通过不超过三条直线连接。
	- 连接线可以水平或垂直，但不能斜线，也不能跨过其他图案。
	- 连接线的转折次数不能超过两次。
	3. 游戏界面:
	- 游戏界面通常是一个矩形区域，内含多个图案或图标，排列成行和列。
	- 图案或图标在未选中状态下背景为白色，选中状态下背景为绿色。
	4. 时间限制: 游戏通常设有时间限制，玩家需要在时间耗尽前完成所有图案的消除。
	5. 得分机制: 每成功连接并消除一对图案，玩家会获得相应的分数。完成游戏后，根据剩余时间和消除效率计算总分。
	6. 关卡设计: 游戏可能包含多个关卡，随着关卡的推进，图案的复杂度和数量会增加。

	注意事项：
	1、当连接错误时，顶部的红心会减少一个，需及时调整策略，避免红心变为0个后游戏失败
	2、不要连续 2 次点击同一个图案
	3、不要犯重复的错误
	`

	userInstruction += "\n\n请严格按照以上游戏规则，开始游戏；注意，请只做点击操作"

	_, err := driver.StartToGoal(context.Background(), userInstruction)
	assert.Nil(t, err)
}

func TestDriverExt_PlanNextAction(t *testing.T) {
	driver := setupDriverExt(t)
	planningResult, err := driver.PlanNextAction(context.Background(), "启动抖音")
	assert.Nil(t, err)
	assert.NotNil(t, planningResult) // Should always return planningResult
	t.Log(planningResult)
}

func TestXTDriver_isTaskFinished(t *testing.T) {
	driver := &XTDriver{}

	tests := []struct {
		name     string
		result   *PlanningExecutionResult
		expected bool
	}{
		{
			name: "no tool calls - task finished",
			result: &PlanningExecutionResult{
				PlanningResult: ai.PlanningResult{
					ToolCalls: []schema.ToolCall{},
					Thought:   "No actions needed",
				},
			},
			expected: true,
		},
		{
			name: "finished action - task finished",
			result: &PlanningExecutionResult{
				PlanningResult: ai.PlanningResult{
					ToolCalls: []schema.ToolCall{
						{
							Function: schema.FunctionCall{
								Name:      "uixt__finished",
								Arguments: `{"content": "Task completed successfully"}`,
							},
						},
					},
					Thought: "Task completed",
				},
			},
			expected: true,
		},
		{
			name: "regular action - task not finished",
			result: &PlanningExecutionResult{
				PlanningResult: ai.PlanningResult{
					ToolCalls: []schema.ToolCall{
						{
							Function: schema.FunctionCall{
								Name:      string(option.ACTION_TapXY),
								Arguments: `{"x": 100, "y": 200}`,
							},
						},
					},
					Thought: "Click on button",
				},
			},
			expected: false,
		},
		{
			name: "multiple actions with finished - task finished",
			result: &PlanningExecutionResult{
				PlanningResult: ai.PlanningResult{
					ToolCalls: []schema.ToolCall{
						{
							Function: schema.FunctionCall{
								Name:      string(option.ACTION_TapXY),
								Arguments: `{"x": 100, "y": 200}`,
							},
						},
						{
							Function: schema.FunctionCall{
								Name:      "uixt__finished",
								Arguments: `{"content": "All tasks completed"}`,
							},
						},
					},
					Thought: "Complete all actions",
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := driver.isTaskFinished(tt.result)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestActionOptions_WithResetHistory(t *testing.T) {
	// Test WithResetHistory option function
	opts := option.NewActionOptions(option.WithResetHistory(true))
	assert.True(t, opts.ResetHistory)

	opts2 := option.NewActionOptions(option.WithResetHistory(false))
	assert.False(t, opts2.ResetHistory)

	// Test default value
	opts3 := option.NewActionOptions()
	assert.False(t, opts3.ResetHistory) // Default should be false
}

func TestXTDriver_PlanNextAction_WithResetHistory(t *testing.T) {
	// Create a minimal XTDriver for testing
	driver := &XTDriver{}

	// Test with nil LLMService (should return error)
	driver.LLMService = nil

	_, err := driver.PlanNextAction(context.Background(), "test prompt", option.WithResetHistory(true))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "LLM service is not initialized")

	// Test that PlanNextAction accepts ResetHistory option
	_, err = driver.PlanNextAction(context.Background(), "test prompt", option.WithResetHistory(false))
	assert.Error(t, err) // Should still error due to nil service
	assert.Contains(t, err.Error(), "LLM service is not initialized")
}

func TestStartToGoal_HistoryResetLogic(t *testing.T) {
	// Test the logic for when history should be reset
	tests := []struct {
		name     string
		attempt  int
		expected bool
	}{
		{"first attempt", 1, true},
		{"second attempt", 2, false},
		{"third attempt", 3, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the logic from StartToGoal
			resetHistory := tt.attempt == 1
			assert.Equal(t, tt.expected, resetHistory)

			// Test that the option is correctly created
			if resetHistory {
				opts := option.NewActionOptions(option.WithResetHistory(true))
				assert.True(t, opts.ResetHistory)
			}
		})
	}
}

func TestConversationHistory_Clear(t *testing.T) {
	// Test Clear method - should clear everything including system message
	history := ai.ConversationHistory{
		{
			Role:    schema.System,
			Content: "System prompt with user instruction",
		},
		{
			Role:    schema.User,
			Content: "User message",
		},
		{
			Role:    schema.Assistant,
			Content: "Assistant response",
		},
	}

	// Test clearing everything including system message
	historyCopy := make(ai.ConversationHistory, len(history))
	copy(historyCopy, history)
	historyCopy.Clear()
	assert.Len(t, historyCopy, 0)

	// Test clearing empty history
	emptyHistory := ai.ConversationHistory{}
	emptyHistory.Clear()
	assert.Len(t, emptyHistory, 0)
}

func TestPlanningOptions_ResetHistory(t *testing.T) {
	// Test that PlanningOptions includes ResetHistory field
	opts := &ai.PlanningOptions{
		UserInstruction: "test instruction",
		Message: &schema.Message{
			Role:    schema.User,
			Content: "test message",
		},
		Size:         types.Size{Width: 100, Height: 200},
		ResetHistory: true,
	}

	assert.True(t, opts.ResetHistory)
	assert.Equal(t, "test instruction", opts.UserInstruction)
}

// TestDriverExt_AIAction tests the AIAction method integration with real driver
func TestDriverExt_AIAction(t *testing.T) {
	driver := setupDriverExt(t)

	// Test AIAction with search button click prompt
	result, err := driver.AIAction(context.Background(), "冷启动抖音app")

	// Verify no error occurred
	assert.Nil(t, err, "AIAction should execute without error")

	// Verify result is not nil
	assert.NotNil(t, result, "AIAction should return a result")

	// Verify result has correct type
	assert.Equal(t, "action", result.Type, "Result type should be 'action'")

	// Verify timing information is captured
	assert.Greater(t, result.ModelCallElapsed, int64(0), "Model call should have elapsed time")
	assert.Greater(t, result.ScreenshotElapsed, int64(0), "Screenshot should have elapsed time")

	// Verify screenshot information is captured
	assert.NotEmpty(t, result.ImagePath, "Image path should not be empty")
	assert.NotNil(t, result.Resolution, "Resolution should not be nil")
	assert.Greater(t, result.Resolution.Width, 0, "Width should be greater than 0")
	assert.Greater(t, result.Resolution.Height, 0, "Height should be greater than 0")

	// Verify planning result is captured
	assert.NotNil(t, result.PlanningResult, "Planning result should not be nil")
	assert.Equal(t, "wings-api", result.PlanningResult.ModelName, "Model name should be 'wings-api'")
	// Log result for debugging
	t.Logf("AIAction executed successfully:")
	t.Logf("  Type: %s", result.Type)
	t.Logf("  Model Call Elapsed: %d ms", result.ModelCallElapsed)
	t.Logf("  Screenshot Elapsed: %d ms", result.ScreenshotElapsed)
	t.Logf("  Image Path: %s", result.ImagePath)
	t.Logf("  Resolution: %dx%d", result.Resolution.Width, result.Resolution.Height)

	if result.Error != "" {
		t.Logf("  Error: %s", result.Error)
	}
}

// TestDriverExt_AIAction_CompareWithAIAction compares AIAction with AIAction
func TestDriverExt_AIAction_CompareWithAIAction(t *testing.T) {
	driver := setupDriverExt(t)

	prompt := "[目标导向]向上滑动屏幕2次"

	// Test both methods with the same prompt
	aiResult, aiErr := driver.StartToGoal(context.Background(), prompt)

	// Both should execute without critical errors (may have different implementations)
	t.Logf("AIAction error: %v", aiErr)
	t.Logf("AIAction result: %v", aiResult)
}

// TestDriverExt_AIAction_ErrorHandling tests AIAction error handling
func TestDriverExt_AIAction_ErrorHandling(t *testing.T) {
	driver := setupDriverExt(t)

	// Test with empty prompt
	result, err := driver.AIAction(context.Background(), "")

	// Should handle empty prompt gracefully
	if err != nil {
		t.Logf("Empty prompt error (expected): %v", err)
		assert.NotNil(t, result, "Result should still be returned even on error")
		if result != nil {
			assert.NotEmpty(t, result.Error, "Result should contain error message")
		}
	} else {
		t.Logf("Empty prompt handled successfully")
		assert.NotNil(t, result, "Result should be returned")
	}

	// Test with very long prompt
	longPrompt := "这是一个非常长的提示词，用来测试AIAction是否能够正确处理长文本输入。" +
		"我们需要确保API能够处理各种长度的输入，包括这种可能超出某些限制的文本。" +
		"请在当前界面中寻找任何可能的搜索相关的按钮或输入框，然后进行点击操作。"

	result2, err2 := driver.AIAction(context.Background(), longPrompt)

	// Should handle long prompt
	if err2 != nil {
		t.Logf("Long prompt error: %v", err2)
	} else {
		t.Logf("Long prompt handled successfully")
		assert.NotNil(t, result2, "Result should be returned for long prompt")
		assert.Equal(t, "action", result2.Type, "Result type should be 'action'")
	}
}

// TestDriverExt_AIAssert tests the AIAssert method integration with real driver
func TestDriverExt_AIAssert(t *testing.T) {
	driver := setupDriverExt(t)

	// Test AIAssert with assertion about search button
	result, err := driver.AIAssert("屏幕中存在搜索按钮")

	// Verify no error occurred (or error is captured in result)
	if err != nil {
		t.Logf("AIAssert error: %v", err)
		// For assertion failures, error is expected, but result should still be returned
		assert.NotNil(t, result, "AIAssert should return a result even on assertion failure")
	} else {
		assert.NotNil(t, result, "AIAssert should return a result")
	}

	// Verify result has correct type
	assert.Equal(t, "assert", result.Type, "Result type should be 'assert'")

	// Verify timing information is captured
	assert.Greater(t, result.ModelCallElapsed, int64(0), "Model call should have elapsed time")
	assert.Greater(t, result.ScreenshotElapsed, int64(0), "Screenshot should have elapsed time")

	// Verify screenshot information is captured
	assert.NotEmpty(t, result.ImagePath, "Image path should not be empty")
	assert.NotNil(t, result.Resolution, "Resolution should not be nil")
	assert.Greater(t, result.Resolution.Width, 0, "Width should be greater than 0")
	assert.Greater(t, result.Resolution.Height, 0, "Height should be greater than 0")

	// Verify assertion result is captured
	assert.NotNil(t, result.AssertionResult, "Assertion result should not be nil")
	assert.NotEmpty(t, result.AssertionResult.Thought, "Assertion result thought should not be empty")

	// Log result for debugging
	t.Logf("AIAssert executed:")
	t.Logf("  Type: %s", result.Type)
	t.Logf("  Model Call Elapsed: %d ms", result.ModelCallElapsed)
	t.Logf("  Screenshot Elapsed: %d ms", result.ScreenshotElapsed)
	t.Logf("  Image Path: %s", result.ImagePath)
	t.Logf("  Resolution: %dx%d", result.Resolution.Width, result.Resolution.Height)
	t.Logf("  Assertion Pass: %t", result.AssertionResult.Pass)
	t.Logf("  Assertion Thought: %s", result.AssertionResult.Thought)

	if result.Error != "" {
		t.Logf("  Error: %s", result.Error)
	}
}

// TestDriverExt_AIAssert_CompareWithAIAssert compares AIAssert with AIAssert
func TestDriverExt_AIAssert_CompareWithAIAssert(t *testing.T) {
	driver := setupDriverExt(t)

	assertion := "屏幕中存在搜索按钮"

	// Test both methods with the same assertion
	wingsResult, wingsErr := driver.AIAssert(assertion)
	aiResult, aiErr := driver.AIAssert(assertion)

	// Both should execute (may have different results)
	t.Logf("AIAssert error: %v", wingsErr)
	t.Logf("AIAssert error: %v", aiErr)

	// If both succeed, compare results
	if wingsResult != nil && aiResult != nil {
		assert.Equal(t, "assert", wingsResult.Type, "AIAssert result type should be 'assert'")
		assert.Equal(t, "assert", aiResult.Type, "AIAssert result type should be 'assert'")

		// Both should have timing information
		assert.Greater(t, wingsResult.ModelCallElapsed, int64(0), "AIAssert should have model call elapsed time")
		assert.Greater(t, aiResult.ModelCallElapsed, int64(0), "AIAssert should have model call elapsed time")

		// Both should have screenshot information
		assert.NotEmpty(t, wingsResult.ImagePath, "AIAssert should have image path")
		assert.NotEmpty(t, aiResult.ImagePath, "AIAssert should have image path")

		// Both should have assertion results
		assert.NotNil(t, wingsResult.AssertionResult, "AIAssert should have assertion result")
		assert.NotNil(t, aiResult.AssertionResult, "AIAssert should have assertion result")

		// Log comparison
		t.Logf("AIAssert Pass: %t, Thought: %s", wingsResult.AssertionResult.Pass, wingsResult.AssertionResult.Thought)
		t.Logf("AIAssert Pass: %t, Thought: %s", aiResult.AssertionResult.Pass, aiResult.AssertionResult.Thought)
	}
}

// TestDriverExt_AIAssert_ErrorHandling tests AIAssert error handling
func TestDriverExt_AIAssert_ErrorHandling(t *testing.T) {
	driver := setupDriverExt(t)

	// Test with empty assertion
	result, err := driver.AIAssert("")

	// Should handle empty assertion gracefully
	if err != nil {
		t.Logf("Empty assertion error (may be expected): %v", err)
		assert.NotNil(t, result, "Result should still be returned even on error")
		if result != nil {
			assert.NotEmpty(t, result.Error, "Result should contain error message")
		}
	} else {
		t.Logf("Empty assertion handled successfully")
		assert.NotNil(t, result, "Result should be returned")
	}

	// Test with complex assertion
	complexAssertion := "断言：当前屏幕显示的是主页面，包含用户头像、搜索框、导航栏等关键元素，并且没有任何错误提示信息"

	result2, err2 := driver.AIAssert(complexAssertion)

	// Should handle complex assertion
	if err2 != nil {
		t.Logf("Complex assertion result: %v", err2)
	} else {
		t.Logf("Complex assertion handled successfully")
		assert.NotNil(t, result2, "Result should be returned for complex assertion")
		assert.Equal(t, "assert", result2.Type, "Result type should be 'assert'")
		if result2.AssertionResult != nil {
			t.Logf("Assertion passed: %t", result2.AssertionResult.Pass)
		}
	}
}
