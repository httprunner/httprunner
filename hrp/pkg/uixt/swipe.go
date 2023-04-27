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
func (dExt *DriverExt) SwipeRelative(fromX, fromY, toX, toY float64, options ...DataOption) error {
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

func (dExt *DriverExt) SwipeTo(direction string, options ...DataOption) (err error) {
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

func (dExt *DriverExt) SwipeUp(options ...DataOption) (err error) {
	return dExt.SwipeRelative(0.5, 0.5, 0.5, 0.1, options...)
}

func (dExt *DriverExt) SwipeDown(options ...DataOption) (err error) {
	return dExt.SwipeRelative(0.5, 0.5, 0.5, 0.9, options...)
}

func (dExt *DriverExt) SwipeLeft(options ...DataOption) (err error) {
	return dExt.SwipeRelative(0.5, 0.5, 0.1, 0.5, options...)
}

func (dExt *DriverExt) SwipeRight(options ...DataOption) (err error) {
	return dExt.SwipeRelative(0.5, 0.5, 0.9, 0.5, options...)
}

type Action func(driver *DriverExt) error

func (dExt *DriverExt) LoopUntil(findAction, findCondition, foundAction Action, options ...DataOption) error {
	dataOptions := NewDataOptions(options...)
	maxRetryTimes := dataOptions.MaxRetryTimes
	interval := dataOptions.Interval

	for i := 0; i < maxRetryTimes; i++ {
		if err := findCondition(dExt); err == nil {
			// do action after found
			return foundAction(dExt)
		}

		if err := findAction(dExt); err != nil {
			log.Error().Err(err).Msgf("find action failed")
		}

		// wait interval between each findAction
		time.Sleep(time.Duration(1000*interval) * time.Millisecond)
	}

	return errors.Wrap(code.OCRTextNotFoundError,
		fmt.Sprintf("loop %d times, match find condition failed", maxRetryTimes))
}

func (dExt *DriverExt) prepareSwipeAction(action MobileAction) func(d *DriverExt) error {
	identifierOption := WithDataIdentifier(action.Identifier)
	durationOption := WithDataPressDuration(action.Duration)

	if action.Steps == 0 {
		action.Steps = 10
	}
	stepsOption := WithDataSteps(action.Steps)

	dataOptions := make([]DataOption, 3)
	dataOptions = append(dataOptions, identifierOption, durationOption, stepsOption)

	return func(d *DriverExt) error {
		defer func() {
			// wait for swipe action to completed and content to load completely
			time.Sleep(time.Duration(1000*action.WaitTime) * time.Millisecond)
		}()

		if d, ok := action.Params.(string); ok {
			// enum direction: up, down, left, right
			if err := dExt.SwipeTo(d, dataOptions...); err != nil {
				log.Error().Err(err).Msgf("swipe %s failed", d)
				return err
			}
		} else if d, ok := action.Params.([]float64); ok {
			// custom direction: [fromX, fromY, toX, toY]
			if err := dExt.SwipeRelative(d[0], d[1], d[2], d[3], dataOptions...); err != nil {
				log.Error().Err(err).Msgf("swipe from (%v, %v) to (%v, %v) failed",
					d[0], d[1], d[2], d[3])
				return err
			}
		} else if d, ok := action.Params.([]interface{}); ok {
			// loaded from json case
			// custom direction: [fromX, fromY, toX, toY]
			sx, _ := builtin.Interface2Float64(d[0])
			sy, _ := builtin.Interface2Float64(d[1])
			ex, _ := builtin.Interface2Float64(d[2])
			ey, _ := builtin.Interface2Float64(d[3])
			if err := dExt.SwipeRelative(sx, sy, ex, ey, dataOptions...); err != nil {
				log.Error().Err(err).Msgf("swipe from (%v, %v) to (%v, %v) failed",
					sx, sy, ex, ey)
				return err
			}
		} else {
			return fmt.Errorf("invalid swipe params %v", action.Params)
		}
		return nil
	}
}

func (dExt *DriverExt) swipeToTapTexts(texts []string, action MobileAction) error {
	if len(action.Scope) != 4 {
		action.Scope = []float64{0, 0, 1, 1}
	}
	if len(action.Offset) != 2 {
		action.Offset = []int{0, 0}
	}

	identifierOption := WithDataIdentifier(action.Identifier)
	offsetOption := WithDataOffset(action.Offset[0], action.Offset[1])
	indexOption := WithDataIndex(action.Index)
	scopeOption := WithDataScope(dExt.getAbsScope(action.Scope[0], action.Scope[1], action.Scope[2], action.Scope[3]))
	// default to retry 10 times
	if action.MaxRetryTimes == 0 {
		action.MaxRetryTimes = 10
	}
	maxRetryOption := WithDataMaxRetryTimes(action.MaxRetryTimes)
	waitTimeOption := WithDataWaitTime(action.WaitTime)

	var point PointF
	findTexts := func(d *DriverExt) error {
		var err error
		ocrTexts, err := d.GetScreenTextsByOCR()
		if err != nil {
			return err
		}
		points, err := ocrTexts.FindTexts(texts, indexOption, scopeOption)
		if err != nil {
			return err
		}
		// FIXME: handle index
		for _, point = range points {
			if point != (PointF{X: 0, Y: 0}) {
				return nil
			}
		}
		return errors.New("failed to find text position")
	}
	foundTextAction := func(d *DriverExt) error {
		// tap text
		return d.TapAbsXY(point.X, point.Y, identifierOption, offsetOption)
	}

	findAction := dExt.prepareSwipeAction(action)
	return dExt.LoopUntil(findAction, findTexts, foundTextAction, maxRetryOption, waitTimeOption)
}

func (dExt *DriverExt) swipeToTapApp(appName string, action MobileAction) error {
	// go to home screen
	if err := dExt.Driver.Homescreen(); err != nil {
		return errors.Wrap(err, "go to home screen failed")
	}

	// swipe to first screen
	for i := 0; i < 5; i++ {
		dExt.SwipeRight()
	}

	action.Offset = []int{0, -25}
	action.Params = "left"

	return dExt.swipeToTapTexts([]string{appName}, action)
}
