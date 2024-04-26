package uixt

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
	"github.com/httprunner/httprunner/v4/hrp/internal/code"
)

func assertRelative(p float64) bool {
	return p >= 0 && p <= 1
}

// SwipeRelative swipe from relative position [fromX, fromY] to relative position [toX, toY]
func (dExt *DriverExt) SwipeRelative(fromX, fromY, toX, toY float64, options ...ActionOption) error {
	width := dExt.windowSize.Width
	height := dExt.windowSize.Height

	if !assertRelative(fromX) || !assertRelative(fromY) ||
		!assertRelative(toX) || !assertRelative(toY) {
		return fmt.Errorf("fromX(%f), fromY(%f), toX(%f), toY(%f) must be less than 1",
			fromX, fromY, toX, toY)
	}

	fromX = float64(width) * fromX
	fromY = float64(height) * fromY
	toX = float64(width) * toX
	toY = float64(height) * toY

	return dExt.Driver.SwipeFloat(fromX, fromY, toX, toY, options...)
}

func (dExt *DriverExt) SwipeTo(direction string, options ...ActionOption) (err error) {
	switch direction {
	case "up":
		return dExt.SwipeUp(options...)
	case "down":
		return dExt.SwipeDown(options...)
	case "left":
		return dExt.SwipeLeft(options...)
	case "right":
		return dExt.SwipeRight(options...)
	}
	return fmt.Errorf("unexpected direction: %s", direction)
}

func (dExt *DriverExt) SwipeUp(options ...ActionOption) (err error) {
	return dExt.SwipeRelative(0.5, 0.5, 0.5, 0.1, options...)
}

func (dExt *DriverExt) SwipeDown(options ...ActionOption) (err error) {
	return dExt.SwipeRelative(0.5, 0.5, 0.5, 0.9, options...)
}

func (dExt *DriverExt) SwipeLeft(options ...ActionOption) (err error) {
	return dExt.SwipeRelative(0.5, 0.5, 0.1, 0.5, options...)
}

func (dExt *DriverExt) SwipeRight(options ...ActionOption) (err error) {
	return dExt.SwipeRelative(0.5, 0.5, 0.9, 0.5, options...)
}

type Action func(driver *DriverExt) error

func (dExt *DriverExt) LoopUntil(findAction, findCondition, foundAction Action, options ...ActionOption) error {
	actionOptions := NewActionOptions(options...)
	maxRetryTimes := actionOptions.MaxRetryTimes
	interval := actionOptions.Interval

	for i := 0; i < maxRetryTimes; i++ {
		// wait interval between each findAction
		time.Sleep(time.Duration(1000*interval) * time.Millisecond)

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

func (dExt *DriverExt) prepareSwipeAction(options ...ActionOption) func(d *DriverExt) error {
	actionOptions := NewActionOptions(options...)
	var swipeDirection interface{}
	if actionOptions.Direction != nil {
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
			if err := dExt.SwipeTo(d, options...); err != nil {
				log.Error().Err(err).Msgf("swipe %s failed", d)
				return err
			}
		} else if d, ok := swipeDirection.([]float64); ok {
			// custom direction: [fromX, fromY, toX, toY]
			if err := dExt.SwipeRelative(d[0], d[1], d[2], d[3], options...); err != nil {
				log.Error().Err(err).Msgf("swipe from (%v, %v) to (%v, %v) failed",
					d[0], d[1], d[2], d[3])
				return err
			}
		} else {
			return fmt.Errorf("invalid swipe params %v", swipeDirection)
		}
		return nil
	}
}

func (dExt *DriverExt) swipeToTapTexts(texts []string, options ...ActionOption) error {
	if len(texts) == 0 {
		return errors.New("no text to tap")
	}

	options = append(options, WithMatchOne(true), WithRegex(true))
	actionOptions := NewActionOptions(options...)
	actionOptions.Identifier = ""
	optionsWithoutIdentifier := actionOptions.Options()
	var point PointF
	findTexts := func(d *DriverExt) error {
		var err error
		screenResult, err := d.GetScreenResult(
			WithScreenShotOCR(true),
			WithScreenShotUpload(true),
		)
		if err != nil {
			return err
		}
		points, err := screenResult.Texts.FindTexts(texts,
			dExt.ParseActionOptions(optionsWithoutIdentifier...)...)
		if err != nil {
			log.Error().Err(err).Strs("texts", texts).Msg("find texts failed")
			// target texts not found, try to auto handle popup
			if e := dExt.ClosePopupsHandler(); e != nil {
				log.Error().Err(e).Msg("run popup handler failed")
			}
			return err
		}
		log.Info().Strs("texts", texts).Interface("results", points).Msg("swipeToTapTexts successful")

		// target texts found, pick the first one
		point = points[0].Center() // FIXME
		return nil
	}
	foundTextAction := func(d *DriverExt) error {
		// tap text
		return d.TapAbsXY(point.X, point.Y, options...)
	}

	findAction := dExt.prepareSwipeAction(optionsWithoutIdentifier...)
	return dExt.LoopUntil(findAction, findTexts, foundTextAction, optionsWithoutIdentifier...)
}

func (dExt *DriverExt) swipeToTapApp(appName string, options ...ActionOption) error {
	// go to home screen
	if err := dExt.Driver.Homescreen(); err != nil {
		return errors.Wrap(err, "go to home screen failed")
	}

	// swipe to first screen
	for i := 0; i < 5; i++ {
		dExt.SwipeRight()
	}

	options = append(options, WithDirection("left"))

	actionOptions := NewActionOptions(options...)
	// default to retry 5 times
	if actionOptions.MaxRetryTimes == 0 {
		options = append(options, WithMaxRetryTimes(5))
	}
	// tap app icon above the text
	if len(actionOptions.Offset) == 0 {
		options = append(options, WithTapOffset(0, -25))
	}
	// set default swipe interval to 1 second
	if builtin.IsZeroFloat64(actionOptions.Interval) {
		options = append(options, WithInterval(1))
	}

	return dExt.swipeToTapTexts([]string{appName}, options...)
}
