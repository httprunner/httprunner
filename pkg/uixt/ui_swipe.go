package uixt

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/pkg/uixt/options"
)

func assertRelative(p float64) bool {
	return p >= 0 && p <= 1
}

// SwipeRelative swipe from relative position [fromX, fromY] to relative position [toX, toY]
func (dExt *DriverExt) SwipeRelative(fromX, fromY, toX, toY float64, opts ...options.ActionOption) error {
	if !assertRelative(fromX) || !assertRelative(fromY) ||
		!assertRelative(toX) || !assertRelative(toY) {
		return errors.Wrap(code.InvalidCaseError,
			fmt.Sprintf("fromX(%f), fromY(%f), toX(%f), toY(%f) must be less than 1",
				fromX, fromY, toX, toY))
	}

	windowSize, err := dExt.Driver.WindowSize()
	if err != nil {
		return errors.Wrap(code.DeviceGetInfoError, err.Error())
	}
	width := windowSize.Width
	height := windowSize.Height

	fromX = float64(width) * fromX
	fromY = float64(height) * fromY
	toX = float64(width) * toX
	toY = float64(height) * toY
	err = dExt.Driver.Swipe(fromX, fromY, toX, toY, opts...)
	if err != nil {
		return errors.Wrap(code.MobileUISwipeError, err.Error())
	}
	return nil
}

func (dExt *DriverExt) SwipeUp(opts ...options.ActionOption) (err error) {
	return dExt.SwipeRelative(0.5, 0.5, 0.5, 0.1, opts...)
}

func (dExt *DriverExt) SwipeDown(opts ...options.ActionOption) (err error) {
	return dExt.SwipeRelative(0.5, 0.5, 0.5, 0.9, opts...)
}

func (dExt *DriverExt) SwipeLeft(opts ...options.ActionOption) (err error) {
	return dExt.SwipeRelative(0.5, 0.5, 0.1, 0.5, opts...)
}

func (dExt *DriverExt) SwipeRight(opts ...options.ActionOption) (err error) {
	return dExt.SwipeRelative(0.5, 0.5, 0.9, 0.5, opts...)
}

type Action func(driver *DriverExt) error

func (dExt *DriverExt) LoopUntil(findAction, findCondition, foundAction Action, opts ...options.ActionOption) error {
	actionOptions := options.NewActionOptions(opts...)
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
			log.Error().Err(err).Msgf("find action failed")
		}
	}

	return errors.Wrap(code.LoopActionNotFoundError,
		fmt.Sprintf("loop %d times, match find condition failed", maxRetryTimes))
}

func (dExt *DriverExt) prepareSwipeAction(params interface{}, opts ...options.ActionOption) func(d *DriverExt) error {
	actionOptions := options.NewActionOptions(opts...)

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

	return func(d *DriverExt) error {
		defer func() {
			// wait for swipe action to completed and content to load completely
			time.Sleep(time.Duration(1000*actionOptions.Interval) * time.Millisecond)
		}()

		if d, ok := swipeDirection.(string); ok {
			// enum direction: up, down, left, right
			switch d {
			case "up":
				return dExt.SwipeUp(opts...)
			case "down":
				return dExt.SwipeDown(opts...)
			case "left":
				return dExt.SwipeLeft(opts...)
			case "right":
				return dExt.SwipeRight(opts...)
			default:
				return errors.Wrap(code.InvalidParamError,
					fmt.Sprintf("get unexpected swipe direction: %s", d))
			}
		} else if params, err := builtin.ConvertToFloat64Slice(swipeDirection); err == nil && len(params) == 4 {
			// custom direction: [fromX, fromY, toX, toY]
			if err := dExt.SwipeRelative(params[0], params[1], params[2], params[3], opts...); err != nil {
				log.Error().Err(err).Msgf("swipe from (%v, %v) to (%v, %v) failed",
					params[0], params[1], params[2], params[3])
				return err
			}
		} else {
			return fmt.Errorf("invalid swipe params %v", swipeDirection)
		}
		return nil
	}
}

func (dExt *DriverExt) swipeToTapTexts(texts []string, opts ...options.ActionOption) error {
	if len(texts) == 0 {
		return errors.New("no text to tap")
	}

	opts = append(opts, options.WithMatchOne(true), options.WithRegex(true))
	actionOptions := options.NewActionOptions(opts...)
	actionOptions.Identifier = ""
	optionsWithoutIdentifier := actionOptions.Options()
	var point PointF
	findTexts := func(d *DriverExt) error {
		var err error
		screenResult, err := d.GetScreenResult(
			options.WithScreenShotOCR(true),
			options.WithScreenShotUpload(true),
			options.WithScreenShotFileName(
				fmt.Sprintf("swipe_to_tap_texts_%s", strings.Join(texts, "_")),
			),
		)
		if err != nil {
			return err
		}
		points, err := screenResult.Texts.FindTexts(texts,
			dExt.ParseActionOptions(optionsWithoutIdentifier...)...)
		if err != nil {
			log.Error().Err(err).Strs("texts", texts).Msg("find texts failed")
			return err
		}
		log.Info().Strs("texts", texts).Interface("results", points).Msg("swipeToTapTexts successful")

		// target texts found, pick the first one
		point = points[0].Center() // FIXME
		return nil
	}
	foundTextAction := func(d *DriverExt) error {
		// tap text
		return d.TapAbsXY(point.X, point.Y, opts...)
	}

	findAction := dExt.prepareSwipeAction(nil, optionsWithoutIdentifier...)
	return dExt.LoopUntil(findAction, findTexts, foundTextAction, optionsWithoutIdentifier...)
}

func (dExt *DriverExt) swipeToTapApp(appName string, opts ...options.ActionOption) error {
	// go to home screen
	if err := dExt.Driver.Homescreen(); err != nil {
		return errors.Wrap(err, "go to home screen failed")
	}

	// automatic handling popups before swipe
	if err := dExt.ClosePopupsHandler(); err != nil {
		log.Error().Err(err).Msg("auto handle popup failed")
	}

	// swipe to first screen
	for i := 0; i < 5; i++ {
		dExt.SwipeRight()
	}

	opts = append(opts, options.WithDirection("left"))

	actionOptions := options.NewActionOptions(opts...)
	// default to retry 5 times
	if actionOptions.MaxRetryTimes == 0 {
		opts = append(opts, options.WithMaxRetryTimes(5))
	}
	// tap app icon above the text
	if len(actionOptions.Offset) == 0 {
		opts = append(opts, options.WithTapOffset(0, -25))
	}
	// set default swipe interval to 1 second
	if builtin.IsZeroFloat64(actionOptions.Interval) {
		opts = append(opts, options.WithInterval(1))
	}

	return dExt.swipeToTapTexts([]string{appName}, opts...)
}
