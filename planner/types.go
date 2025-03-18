package planner

import (
	"github.com/cloudwego/eino/schema"
)

// PlanningOptions represents the input options for planning
type PlanningOptions struct {
	UserInstruction     string            `json:"user_instruction"`
	ConversationHistory []*schema.Message `json:"conversation_history"`
	Size                Size              `json:"size"`
}

// Size represents the dimensions of a screen
type Size struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// PlanningResult represents the result of planning
type PlanningResult struct {
	Actions       []interface{}  `json:"actions"`
	RealActions   []ParsedAction `json:"real_actions"`
	ActionSummary string         `json:"action_summary"`
}

// VLMResponse represents the response from the Vision Language Model
type VLMResponse struct {
	Actions []ParsedAction `json:"actions"`
	Error   string         `json:"error,omitempty"`
}

// ParsedAction represents a parsed action from the VLM response
type ParsedAction struct {
	ActionType   string                 `json:"actionType"`
	ActionInputs map[string]interface{} `json:"actionInputs"`
	Thought      string                 `json:"thought"`
}
