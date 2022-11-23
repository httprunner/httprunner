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

// findCondition indicates the condition to find a UI element
// foundAction indicates the action to do after a UI element is found
func (dExt *DriverExt) SwipeUntil(direction interface{}, findCondition Action, foundAction Action, options ...DataOption) error {
	dataOptions := NewDataOptions(options...)
	maxRetryTimes := dataOptions.MaxRetryTimes
	interval := dataOptions.Interval

	for i := 0; i < maxRetryTimes; i++ {
		if err := findCondition(dExt); err == nil {
			// do action after found
			return foundAction(dExt)
		}
		if d, ok := direction.(string); ok {
			if err := dExt.SwipeTo(d); err != nil {
				log.Error().Err(err).Msgf("swipe %s failed", d)
			}
		} else if d, ok := direction.([]float64); ok {
			if err := dExt.SwipeRelative(d[0], d[1], d[2], d[3]); err != nil {
				log.Error().Err(err).Msgf("swipe %v failed", d)
			}
		} else if d, ok := direction.([]interface{}); ok {
			sx, _ := builtin.Interface2Float64(d[0])
			sy, _ := builtin.Interface2Float64(d[1])
			ex, _ := builtin.Interface2Float64(d[2])
			ey, _ := builtin.Interface2Float64(d[3])
			if err := dExt.SwipeRelative(sx, sy, ex, ey); err != nil {
				log.Error().Err(err).Msgf("swipe (%v, %v) to (%v, %v) failed", sx, sy, ex, ey)
			}
		}
		// wait for swipe action to completed and content to load completely
		time.Sleep(time.Duration(1000*interval) * time.Millisecond)
	}
	return errors.Wrap(code.OCRTextNotFoundError,
		fmt.Sprintf("swipe %s %d times, match condition failed", direction, maxRetryTimes))
}

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

func (dExt *DriverExt) swipeToTapApp(appName string, action MobileAction) error {
	if len(action.Scope) != 4 {
		action.Scope = []float64{0, 0, 1, 1}
	}
	if len(action.Offset) != 2 {
		action.Offset = []int{0, -25}
	}

	identifierOption := WithDataIdentifier(action.Identifier)
	indexOption := WithDataIndex(action.Index)
	offsetOption := WithDataOffset(action.Offset[0], action.Offset[1])
	scopeOption := WithDataScope(dExt.getAbsScope(action.Scope[0], action.Scope[1], action.Scope[2], action.Scope[3]))

	// default to retry 5 times
	if action.MaxRetryTimes == 0 {
		action.MaxRetryTimes = 5
	}
	maxRetryOption := WithDataMaxRetryTimes(action.MaxRetryTimes)
	waitTimeOption := WithDataWaitTime(action.WaitTime)

	var point PointF
	findAppAction := func(d *DriverExt) error {
		return dExt.SwipeLeft()
	}
	findAppCondition := func(d *DriverExt) error {
		var err error
		point, err = d.GetTextXY(appName, scopeOption, indexOption)
		return err
	}
	foundAppAction := func(d *DriverExt) error {
		// click app to launch
		return d.TapAbsXY(point.X, point.Y, identifierOption, offsetOption)
	}

	// go to home screen
	if err := dExt.Driver.Homescreen(); err != nil {
		return errors.Wrap(err, "go to home screen failed")
	}

	// swipe to first screen
	for i := 0; i < 5; i++ {
		dExt.SwipeRight()
	}

	// swipe next screen until app found
	return dExt.LoopUntil(findAppAction, findAppCondition, foundAppAction, maxRetryOption, waitTimeOption)
}
