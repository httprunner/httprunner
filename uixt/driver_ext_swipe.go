package uixt

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/uixt/ai"
	"github.com/httprunner/httprunner/v5/uixt/option"
)

type Action func(driver *XTDriver) error

func (dExt *XTDriver) LoopUntil(findAction, findCondition, foundAction Action, opts ...option.ActionOption) error {
	actionOptions := option.NewActionOptions(opts...)
	maxRetryTimes := actionOptions.MaxRetryTimes
	interval := actionOptions.Interval

	for i := 0; i < maxRetryTimes; i++ {
		// wait interval between each findAction
		time.Sleep(time.Duration(interval) * time.Second)

		if err := findCondition(dExt); err == nil {
			// do action after found
			return foundAction(dExt)
		}

		if err := findAction(dExt); err != nil {
			// find action failed, abort loop
			log.Error().Err(err).Msgf("find action failed")
			return err
		}
	}

	return errors.Wrap(code.LoopActionNotFoundError,
		fmt.Sprintf("loop %d times, match find condition failed", maxRetryTimes))
}

func prepareSwipeAction(dExt *XTDriver, params interface{}, opts ...option.ActionOption) func(d *XTDriver) error {
	actionOptions := option.NewActionOptions(opts...)

	var swipeDirection interface{}
	// priority: params > actionOptions.Direction, default swipe up
	if params != nil {
		swipeDirection = params
	} else if actionOptions.Direction != nil {
		swipeDirection = actionOptions.Direction
	} else {
		swipeDirection = "up" // default swipe up
	}

	if actionOptions.Steps == 0 {
		actionOptions.Steps = 10
	}

	return func(d *XTDriver) error {
		defer func() {
			// wait for swipe action to completed and content to load completely
			time.Sleep(time.Duration(1000*actionOptions.Interval) * time.Millisecond)
		}()

		if d, ok := swipeDirection.(string); ok {
			// enum direction: up, down, left, right
			switch d {
			case "up":
				return dExt.Swipe(0.5, 0.5, 0.5, 0.1, opts...)
			case "down":
				return dExt.Swipe(0.5, 0.5, 0.5, 0.9, opts...)
			case "left":
				return dExt.Swipe(0.5, 0.5, 0.1, 0.5, opts...)
			case "right":
				return dExt.Swipe(0.5, 0.5, 0.9, 0.5, opts...)
			default:
				return errors.Wrap(code.InvalidParamError,
					fmt.Sprintf("get unexpected swipe direction: %s", d))
			}
		} else if params, err := builtin.ConvertToFloat64Slice(swipeDirection); err == nil && len(params) == 4 {
			// custom direction: [fromX, fromY, toX, toY]
			if err := dExt.Swipe(params[0], params[1], params[2], params[3], opts...); err != nil {
				log.Error().Err(err).Msgf("swipe from (%v, %v) to (%v, %v) failed",
					params[0], params[1], params[2], params[3])
				return err
			}
		} else {
			return errors.Wrap(code.InvalidParamError,
				fmt.Sprintf("invalid swipe params %v", swipeDirection))
		}
		return nil
	}
}

func (dExt *XTDriver) SwipeToTapTexts(texts []string, opts ...option.ActionOption) error {
	if len(texts) == 0 {
		return errors.New("no text to tap")
	}

	log.Info().Strs("texts", texts).Msg("swipe to tap texts")
	opts = append([]option.ActionOption{option.WithMatchOne(true), option.WithRegex(true), option.WithInterval(1)}, opts...)
	// Remove identifier for swipe operations to avoid WDA/UIA2 logging
	actionOptions := option.NewActionOptions(opts...)
	actionOptions.Identifier = ""
	optionsWithoutIdentifier := actionOptions.Options()

	var point ai.PointF
	findTexts := func(d *XTDriver) error {
		var err error
		screenResult, err := d.GetScreenResult(
			option.WithScreenShotOCR(true),
			option.WithScreenShotUpload(true),
			option.WithScreenShotFileName(
				fmt.Sprintf("swipe_to_tap_texts_%s", strings.Join(texts, "_")),
			),
		)
		if err != nil {
			return err
		}

		points, err := screenResult.Texts.FindTexts(texts,
			convertToAbsoluteScope(dExt.IDriver, opts...)...)
		if err != nil {
			log.Error().Err(err).Strs("texts", texts).Msg("find texts failed")
			return err
		}
		log.Info().Strs("texts", texts).Interface("results", points).Msg("swipeToTapTexts successful")

		// target texts found, pick the first one
		point = points[0].Center() // FIXME
		return nil
	}
	foundTextAction := func(d *XTDriver) error {
		// tap text
		return d.TapAbsXY(point.X, point.Y, opts...)
	}

	findAction := prepareSwipeAction(dExt, nil, optionsWithoutIdentifier...)
	return dExt.LoopUntil(findAction, findTexts, foundTextAction, opts...)
}

func (dExt *XTDriver) SwipeToTapApp(appName string, opts ...option.ActionOption) error {
	log.Info().Str("appName", appName).Msg("swipe to tap app")

	// go to home screen
	if err := dExt.Home(); err != nil {
		return errors.Wrap(err, "go to home screen failed")
	}

	// automatic handling popups before swipe
	if err := dExt.ClosePopupsHandler(); err != nil {
		log.Error().Err(err).Msg("auto handle popup failed")
	}

	// Remove identifier for swipe operations to avoid WDA/UIA2 logging
	actionOptions := option.NewActionOptions(opts...)
	actionOptions.Identifier = ""
	optionsWithoutIdentifier := actionOptions.Options()

	// swipe to first screen
	for i := 0; i < 5; i++ {
		err := dExt.Swipe(0.5, 0.5, 0.9, 0.5, optionsWithoutIdentifier...)
		if err != nil {
			log.Error().Err(err).Msgf("swipe to tap app %s failed", appName)
			return errors.Wrapf(err, "swipe to tap app %s failed", appName)
		}
		time.Sleep(1 * time.Second)
	}

	opts = append(opts, option.WithDirection("left"))
	opts = append(opts, option.WithMaxRetryTimes(5))

	// tap app icon above the text
	if len(actionOptions.TapOffset) == 0 {
		opts = append(opts, option.WithTapOffset(0, -100))
	}
	// set default swipe interval to 1 second
	if builtin.IsZeroFloat64(actionOptions.Interval) {
		opts = append(opts, option.WithInterval(1))
	}

	return dExt.SwipeToTapTexts([]string{appName}, opts...)
}
