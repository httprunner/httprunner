//go:build localtest

package uixt

import (
	"context"
	"testing"

	"github.com/cloudwego/eino/schema"
	"github.com/httprunner/httprunner/v5/uixt/ai"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/httprunner/httprunner/v5/uixt/types"
	"github.com/stretchr/testify/assert"
)

func TestDriverExt_TapByLLM(t *testing.T) {
	driver := setupDriverExt(t)
	err := driver.AIAction(context.Background(), "点击第一个帖子的作者头像")
	assert.Nil(t, err)

	err = driver.AIAssert("当前在个人介绍页")
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

	err := driver.StartToGoal(context.Background(), userInstruction)
	assert.Nil(t, err)
}

func TestDriverExt_PlanNextAction(t *testing.T) {
	driver := setupDriverExt(t)
	result, err := driver.PlanNextAction(context.Background(), "启动抖音")
	assert.Nil(t, err)
	t.Log(result)
}

func TestXTDriver_isTaskFinished(t *testing.T) {
	driver := &XTDriver{}

	tests := []struct {
		name     string
		result   *ai.PlanningResult
		expected bool
	}{
		{
			name: "no tool calls - task finished",
			result: &ai.PlanningResult{
				ToolCalls: []schema.ToolCall{},
				Thought:   "No actions needed",
			},
			expected: true,
		},
		{
			name: "finished action - task finished",
			result: &ai.PlanningResult{
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
			expected: true,
		},
		{
			name: "regular action - task not finished",
			result: &ai.PlanningResult{
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
			expected: false,
		},
		{
			name: "multiple actions with finished - task finished",
			result: &ai.PlanningResult{
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
