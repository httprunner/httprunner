package ai

import (
	"context"
	"fmt"
	"time"

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
}

// PlanningResult represents the result of planning
type PlanningResult struct {
	ToolCalls     []schema.ToolCall `json:"tool_calls"`
	ActionSummary string            `json:"summary"`
	Thought       string            `json:"thought"`
	Content       string            `json:"content"` // original content from model
	Error         string            `json:"error,omitempty"`
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
	if p.modelConfig.ModelType == option.LLMServiceTypeUITARS {
		// tools have been registered in ui-tars system prompt
		return nil
	}

	// register tools for models with function calling
	toolCallingModel, err := p.model.WithTools(tools)
	if err != nil {
		return errors.Wrap(err, "failed to register tools")
	}
	p.model = toolCallingModel
	return nil
}

// Call performs UI planning using Vision Language Model
func (p *Planner) Call(ctx context.Context, opts *PlanningOptions) (*PlanningResult, error) {
	// validate input parameters
	if err := validatePlanningInput(opts); err != nil {
		return nil, errors.Wrap(err, "validate planning parameters failed")
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
	logRequest(p.history)
	startTime := time.Now()
	message, err := p.model.Generate(ctx, p.history)
	log.Info().Float64("elapsed(s)", time.Since(startTime).Seconds()).
		Str("model", string(p.modelConfig.ModelType)).Msg("call model service")
	if err != nil {
		return nil, errors.Wrap(code.LLMRequestServiceError, err.Error())
	}
	logResponse(message)

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
		result := &PlanningResult{
			ToolCalls:     message.ToolCalls,
			ActionSummary: message.Content,
		}
		return result, nil
	}

	// parse message content to actions (tool calls)
	result, err := p.parser.Parse(message.Content, opts.Size)
	if err != nil {
		result = &PlanningResult{
			ActionSummary: message.Content,
			Error:         err.Error(),
		}
		log.Debug().Str("reason", err.Error()).Msg("parse content to actions failed")
		// append assistant message
		p.history.Append(&schema.Message{
			Role:    schema.Assistant,
			Content: message.Content,
		})
	} else {
		// append assistant message with tool calls
		p.history.Append(&schema.Message{
			Role:       schema.Tool,
			Content:    result.Content,
			ToolCalls:  result.ToolCalls,
			ToolCallID: fmt.Sprintf("%d", time.Now().Unix()),
		})
	}

	log.Info().
		Interface("summary", result.ActionSummary).
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
