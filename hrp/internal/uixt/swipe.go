package uixt

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

func assertRelative(p float64) bool {
	return p >= 0 && p <= 1
}

// SwipeRelative swipe from relative position [fromX, fromY] to relative position [toX, toY]
func (dExt *DriverExt) SwipeRelative(fromX, fromY, toX, toY float64, identifier ...string) error {
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

	if len(identifier) > 0 && identifier[0] != "" {
		option := WithCustomOption("log", map[string]interface{}{
			"enable": true,
			"data":   identifier[0],
		})
		dExt.Driver.SwipeFloat(fromX, fromY, toX, toY, option)
	}
	return dExt.Driver.SwipeFloat(fromX, fromY, toX, toY)
}

func (dExt *DriverExt) SwipeTo(direction string, identifier ...string) (err error) {
	switch direction {
	case "up":
		return dExt.SwipeUp(identifier...)
	case "down":
		return dExt.SwipeDown(identifier...)
	case "left":
		return dExt.SwipeLeft(identifier...)
	case "right":
		return dExt.SwipeRight(identifier...)
	}
	return fmt.Errorf("unexpected direction: %s", direction)
}

func (dExt *DriverExt) SwipeUp(identifier ...string) (err error) {
	return dExt.SwipeRelative(0.5, 0.5, 0.5, 0.1, identifier...)
}

func (dExt *DriverExt) SwipeDown(identifier ...string) (err error) {
	return dExt.SwipeRelative(0.5, 0.5, 0.5, 0.9, identifier...)
}

func (dExt *DriverExt) SwipeLeft(identifier ...string) (err error) {
	return dExt.SwipeRelative(0.5, 0.5, 0.1, 0.5, identifier...)
}

func (dExt *DriverExt) SwipeRight(identifier ...string) (err error) {
	return dExt.SwipeRelative(0.5, 0.5, 0.9, 0.5, identifier...)
}

// FindCondition indicates the condition to find a UI element
type FindCondition func(driver *DriverExt) error

// FoundAction indicates the action to do after a UI element is found
type FoundAction func(driver *DriverExt) error

func (dExt *DriverExt) SwipeUntil(direction string, condition FindCondition, action FoundAction, maxTimes int) error {
	for i := 0; i < maxTimes; i++ {
		if err := condition(dExt); err == nil {
			// do action after found
			return action(dExt)
		}
		if err := dExt.SwipeTo(direction); err != nil {
			log.Error().Err(err).Msgf("swipe %s failed", direction)
		}
		// wait for swipe done
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("swipe %s %d times, match condition failed", direction, maxTimes)
}
