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
	Call(opts *PlanningOptions) (*PlanningResult, error)
}

// PlanningOptions represents the input options for planning
type PlanningOptions struct {
	UserInstruction string          `json:"user_instruction"` // append to system prompt
	Message         *schema.Message `json:"message"`
	Size            types.Size      `json:"size"`
}

// PlanningResult represents the result of planning
type PlanningResult struct {
	NextActions   []ParsedAction `json:"actions"`
	ActionSummary string         `json:"summary"`
	Error         string         `json:"error,omitempty"`
}

func NewPlanner(ctx context.Context, modelType option.LLMServiceType) (*Planner, error) {
	planner := &Planner{
		ctx:       ctx,
		modelType: modelType,
	}

	config, err := GetOpenAIModelConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenAI config: %w", err)
	}

	if modelType == option.LLMServiceTypeUITARS {
		planner.systemPrompt = uiTarsPlanningPrompt
	} else {
		planner.systemPrompt = defaultPlanningResponseJsonFormat
	}

	planner.model, err = openai.NewChatModel(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize OpenAI model: %w", err)
	}

	return planner, nil
}

type Planner struct {
	ctx          context.Context
	model        model.ToolCallingChatModel
	systemPrompt string
	modelType    option.LLMServiceType
	history      ConversationHistory
}

// Call performs UI planning using Vision Language Model
func (p *Planner) Call(opts *PlanningOptions) (*PlanningResult, error) {
	// validate input parameters
	if err := validatePlanningInput(opts); err != nil {
		return nil, errors.Wrap(err, "validate planning parameters failed")
	}

	// prepare prompt
	if len(p.history) == 0 {
		// add system message
		systemPrompt := uiTarsPlanningPrompt + opts.UserInstruction
		p.history = ConversationHistory{
			{
				Role:    schema.System,
				Content: systemPrompt,
			},
		}
	}
	// append user image message
	p.history.Append(opts.Message)

	// call model service, generate response
	logRequest(p.history)
	startTime := time.Now()
	resp, err := p.model.Generate(p.ctx, p.history)
	log.Info().Float64("elapsed(s)", time.Since(startTime).Seconds()).
		Str("model", string(p.modelType)).Msg("call model service")
	if err != nil {
		return nil, errors.Wrap(code.LLMRequestServiceError, err.Error())
	}
	logResponse(resp)

	// parse result
	result, err := p.parseResult(resp, opts.Size)
	if err != nil {
		return nil, errors.Wrap(code.LLMParsePlanningResponseError, err.Error())
	}

	// append assistant message
	p.history.Append(&schema.Message{
		Role:    schema.Assistant,
		Content: result.ActionSummary,
	})

	return result, nil
}

func (p *Planner) parseResult(msg *schema.Message, size types.Size) (*PlanningResult, error) {
	var parseActions []ParsedAction
	var err error
	if p.modelType == option.LLMServiceTypeUITARS {
		// parse Thought/Action format from UI-TARS
		parseActions, err = parseThoughtAction(msg.Content)
		if err != nil {
			return nil, err
		}
	} else {
		// parse JSON format, from VLM like openai/gpt-4o
		parseActions, err = parseJSON(msg.Content)
		if err != nil {
			return nil, err
		}
	}

	// process response
	result, err := processVLMResponse(parseActions, size)
	if err != nil {
		return nil, errors.Wrap(err, "process VLM response failed")
	}

	log.Info().
		Interface("summary", result.ActionSummary).
		Interface("actions", result.NextActions).
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
