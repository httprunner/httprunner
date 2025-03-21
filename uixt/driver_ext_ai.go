package uixt

import (
	"github.com/cloudwego/eino/schema"
	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/uixt/ai"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func (dExt *XTDriver) StartToGoal(text string, opts ...option.ActionOption) error {
	options := option.NewActionOptions(opts...)
	var attempt int
	for {
		attempt++
		log.Info().Int("attempt", attempt).Msg("planning attempt")

		// plan next action
		result, err := dExt.PlanNextAction(text, opts...)
		if err != nil {
			return errors.Wrap(err, "failed to get next action from planner")
		}

		// do actions
		for _, action := range result.NextActions {
			switch action.ActionType {
			case ai.ActionTypeClick:
				point := action.ActionInputs["startBox"].([]float64)
				if err := dExt.TapAbsXY(point[0], point[1], opts...); err != nil {
					return err
				}
			case ai.ActionTypeFinished:
				return nil
			}
		}

		if options.MaxRetryTimes > 1 && attempt >= options.MaxRetryTimes {
			return errors.New("reached max retry times")
		}
	}
}

func (dExt *XTDriver) PlanNextAction(text string, opts ...option.ActionOption) (*ai.PlanningResult, error) {
	if dExt.LLMService == nil {
		return nil, errors.New("LLM service is not initialized")
	}

	screenShotBase64, err := dExt.GetScreenShotBase64()
	if err != nil {
		return nil, err
	}

	size, err := dExt.IDriver.WindowSize()
	if err != nil {
		return nil, errors.Wrap(code.DeviceGetInfoError, err.Error())
	}

	planningOpts := &ai.PlanningOptions{
		UserInstruction: text,
		Message: &schema.Message{
			Role: schema.User,
			MultiContent: []schema.ChatMessagePart{
				{
					Type: schema.ChatMessagePartTypeImageURL,
					ImageURL: &schema.ChatMessageImageURL{
						URL: screenShotBase64,
					},
				},
			},
		},
		Size: size,
	}

	result, err := dExt.LLMService.Call(planningOpts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get next action from planner")
	}
	return result, nil
}
