package ai

import (
	"context"
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

func NewPlanner(ctx context.Context, modelConfig *ModelConfig) (*Planner, error) {
	planner := &Planner{
		ctx:         ctx,
		modelConfig: modelConfig,
	}

	if modelConfig.ModelType == option.LLMServiceTypeUITARS {
		planner.systemPrompt = uiTarsPlanningPrompt
	} else {
		planner.systemPrompt = defaultPlanningResponseJsonFormat
	}

	var err error
	planner.model, err = openai.NewChatModel(ctx, modelConfig.ChatModelConfig)
	if err != nil {
		return nil, errors.Wrap(code.LLMPrepareRequestError, err.Error())
	}

	return planner, nil
}

type Planner struct {
	ctx          context.Context
	modelConfig  *ModelConfig
	model        model.ToolCallingChatModel
	systemPrompt string
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
		p.history = ConversationHistory{
			{
				Role:    schema.System,
				Content: p.systemPrompt + opts.UserInstruction,
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
		Str("model", string(p.modelConfig.ModelType)).Msg("call model service")
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
	if p.modelConfig.ModelType == option.LLMServiceTypeUITARS {
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
