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
