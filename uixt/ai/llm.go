package ai

import (
	"github.com/cloudwego/eino/schema"
	"github.com/httprunner/httprunner/v5/uixt/types"
)

type ILLMService interface {
	Call(opts *PlanningOptions) (*PlanningResult, error)
}

func NewGPT4oLLMService() (*openaiLLMService, error) {
	return &openaiLLMService{}, nil
}

type openaiLLMService struct{}

func (s openaiLLMService) Call(opts *PlanningOptions) (*PlanningResult, error) {
	return nil, nil
}

// PlanningOptions represents the input options for planning
type PlanningOptions struct {
	UserInstruction     string            `json:"user_instruction"`
	ConversationHistory []*schema.Message `json:"conversation_history"`
	Size                types.Size        `json:"size"`
}

// PlanningResult represents the result of planning
type PlanningResult struct {
	NextActions   []ParsedAction `json:"actions"`
	ActionSummary string         `json:"summary"`
}

// VLMResponse represents the response from the Vision Language Model
type VLMResponse struct {
	Actions []ParsedAction `json:"actions"`
	Error   string         `json:"error,omitempty"`
}

// ParsedAction represents a parsed action from the VLM response
type ParsedAction struct {
	ActionType   ActionType             `json:"actionType"`
	ActionInputs map[string]interface{} `json:"actionInputs"`
	Thought      string                 `json:"thought"`
}

type ActionType string

const (
	ActionTypeClick    ActionType = "click"
	ActionTypeTap      ActionType = "tap"
	ActionTypeDrag     ActionType = "drag"
	ActionTypeSwipe    ActionType = "swipe"
	ActionTypeWait     ActionType = "wait"
	ActionTypeFinished ActionType = "finished"
	ActionTypeCallUser ActionType = "call_user"
	ActionTypeType     ActionType = "type"
	ActionTypeScroll   ActionType = "scroll"
)
