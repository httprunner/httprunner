package uixt

import (
	"encoding/base64"
	"fmt"
	"path/filepath"

	"github.com/cloudwego/eino/schema"
	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/internal/config"
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
		if err := dExt.AIAction(text, opts...); err != nil {
			return err
		}

		if options.MaxRetryTimes > 1 && attempt >= options.MaxRetryTimes {
			return errors.New("reached max retry times")
		}
	}
}

func (dExt *XTDriver) AIAction(text string, opts ...option.ActionOption) error {
	// plan next action
	result, err := dExt.PlanNextAction(text, opts...)
	if err != nil {
		return err
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
			log.Info().Msg("ai action done")
			return nil
		}
	}

	return nil
}

func (dExt *XTDriver) PlanNextAction(text string, opts ...option.ActionOption) (*ai.PlanningResult, error) {
	if dExt.LLMService == nil {
		return nil, errors.New("LLM service is not initialized")
	}

	compressedBufSource, err := getScreenShotBuffer(dExt.IDriver)
	if err != nil {
		return nil, err
	}

	// convert buffer to base64 string
	screenShotBase64 := "data:image/jpeg;base64," +
		base64.StdEncoding.EncodeToString(compressedBufSource.Bytes())

	// save screenshot to file
	imagePath := filepath.Join(
		config.GetConfig().ScreenShotsPath,
		fmt.Sprintf("%s.jpeg", builtin.GenNameWithTimestamp("%d_screenshot")),
	)
	go func() {
		err := saveScreenShot(compressedBufSource, imagePath)
		if err != nil {
			log.Error().Err(err).Msg("save screenshot file failed")
		}
	}()

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

func (dExt *XTDriver) AIQuery(text string, opts ...option.ActionOption) (string, error) {
	return "", nil
}

func (dExt *XTDriver) AIAssert(assertion string, opts ...option.ActionOption) error {
	if dExt.LLMService == nil {
		return errors.New("LLM service is not initialized")
	}

	compressedBufSource, err := getScreenShotBuffer(dExt.IDriver)
	if err != nil {
		return err
	}

	// convert buffer to base64 string
	screenShotBase64 := "data:image/jpeg;base64," +
		base64.StdEncoding.EncodeToString(compressedBufSource.Bytes())

	// get window size
	size, err := dExt.IDriver.WindowSize()
	if err != nil {
		return errors.Wrap(err, "get window size for AI assertion failed")
	}

	// execute assertion
	assertOpts := &ai.AssertOptions{
		Assertion:  assertion,
		Screenshot: screenShotBase64,
		Size:       size,
	}
	result, err := dExt.LLMService.Assert(assertOpts)
	if err != nil {
		return errors.Wrap(err, "AI assertion failed")
	}

	if !result.Pass {
		return errors.New(result.Thought)
	}

	return nil
}
