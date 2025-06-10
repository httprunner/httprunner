package ai

import (
	"context"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/httprunner/httprunner/v5/uixt/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type IPlanner interface {
	Call(ctx context.Context, opts *PlanningOptions) (*PlanningResult, error)
}

// PlanningOptions represents the input options for planning
type PlanningOptions struct {
	UserInstruction string          `json:"user_instruction"` // append to system prompt
	Message         *schema.Message `json:"message"`
	Size            types.Size      `json:"size"`
	ResetHistory    bool            `json:"reset_history"` // whether to reset conversation history before planning
}

// PlanningResult represents the result of planning
type PlanningResult struct {
	ToolCalls []schema.ToolCall  `json:"tool_calls"`
	Thought   string             `json:"thought"`
	Content   string             `json:"content"` // original content from model
	Error     string             `json:"error,omitempty"`
	ModelName string             `json:"model_name"`      // model name used for planning
	Usage     *schema.TokenUsage `json:"usage,omitempty"` // token usage statistics
}

func NewPlanner(ctx context.Context, modelConfig *ModelConfig) (*Planner, error) {
	planner := &Planner{
		modelConfig: modelConfig,
		parser:      NewLLMContentParser(modelConfig.ModelType),
	}

	var err error
	planner.model, err = openai.NewChatModel(ctx, modelConfig.ChatModelConfig)
	if err != nil {
		return nil, errors.Wrap(code.LLMPrepareRequestError, err.Error())
	}

	return planner, nil
}

type Planner struct {
	modelConfig *ModelConfig
	model       model.ToolCallingChatModel
	parser      LLMContentParser
	history     ConversationHistory
}

func (p *Planner) SystemPrompt() string {
	return p.parser.SystemPrompt()
}

func (p *Planner) History() *ConversationHistory {
	return &p.history
}

func (p *Planner) RegisterTools(tools []*schema.ToolInfo) error {
	if option.IS_UI_TARS(p.modelConfig.ModelType) {
		// tools have been registered in ui-tars system prompt
		return nil
	}

	// register tools for models with function calling
	toolCallingModel, err := p.model.WithTools(tools)
	if err != nil {
		return errors.Wrap(err, "failed to register tools")
	}

	var toolNames []string
	for _, tool := range tools {
		toolNames = append(toolNames, tool.Name)
	}
	log.Debug().Strs("tools", toolNames).
		Str("model", string(p.modelConfig.ModelType)).
		Msg("registered tools to model")

	p.model = toolCallingModel
	return nil
}

// Call performs UI planning using Vision Language Model
func (p *Planner) Call(ctx context.Context, opts *PlanningOptions) (result *PlanningResult, err error) {
	// validate input parameters
	if err := validatePlanningInput(opts); err != nil {
		return nil, errors.Wrap(err, "validate planning parameters failed")
	}

	// reset conversation history if requested
	if opts.ResetHistory {
		p.history.Clear() // Clear everything including system message for complete isolation
	}

	// prepare prompt
	if len(p.history) == 0 && opts.UserInstruction != "" {
		// add system message
		p.history = ConversationHistory{
			{
				Role:    schema.System,
				Content: p.parser.SystemPrompt() + opts.UserInstruction,
			},
		}
	}
	// append user image message
	p.history.Append(opts.Message)

	// call model service, generate response
	message, err := callModelWithLogging(ctx, p.model, p.history,
		p.modelConfig.ModelType, "planning")
	if err != nil {
		return nil, errors.Wrap(code.LLMRequestServiceError, err.Error())
	}

	defer func() {
		// Extract usage information if available
		if message.ResponseMeta != nil && message.ResponseMeta.Usage != nil {
			result.Usage = message.ResponseMeta.Usage
		}
	}()

	// handle tool calls
	if len(message.ToolCalls) > 0 {
		// append tool call message
		toolCallID := ""
		for _, toolCall := range message.ToolCalls {
			toolCallID += toolCall.ID
		}
		p.history.Append(&schema.Message{
			Role:       schema.Tool,
			Content:    message.Content,
			ToolCalls:  message.ToolCalls,
			ToolCallID: toolCallID,
		})
		// history will be appended with tool calls execution result
		result = &PlanningResult{
			ToolCalls: message.ToolCalls,
			Thought:   message.Content,
			ModelName: string(p.modelConfig.ModelType),
		}
		return result, nil
	}

	// parse message content to actions (tool calls)
	result, err = p.parser.Parse(message.Content, opts.Size)
	if err != nil {
		result = &PlanningResult{
			Thought:   message.Content,
			Error:     err.Error(),
			ModelName: string(p.modelConfig.ModelType),
		}
		log.Debug().Str("reason", err.Error()).Msg("parse content to actions failed")
	}
	// append assistant message (since we're parsing content, not using native function calling)
	p.history.Append(&schema.Message{
		Role:    schema.Assistant,
		Content: message.Content,
	})

	log.Info().
		Interface("thought", result.Thought).
		Interface("tool_calls", result.ToolCalls).
		Msg("get VLM planning result")
	return result, nil
}

func validatePlanningInput(opts *PlanningOptions) error {
	if opts.UserInstruction == "" {
		return errors.Wrap(code.LLMPrepareRequestError, "user instruction is empty")
	}

	if opts.Message == nil || opts.Message.Role == "" {
		return errors.Wrap(code.LLMPrepareRequestError, "user message is empty")
	}

	if opts.Message.Role == schema.User {
		// check MultiContent
		if len(opts.Message.MultiContent) > 0 {
			for _, content := range opts.Message.MultiContent {
				if content.Type == schema.ChatMessagePartTypeImageURL && content.ImageURL == nil {
					return errors.Wrap(code.LLMPrepareRequestError, "invalid image data")
				}
			}
		}
	}

	return nil
}
