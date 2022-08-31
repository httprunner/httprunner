package uixt

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

func (dExt *DriverExt) SwipeTo(direction string) (err error) {
	width := dExt.windowSize.Width
	height := dExt.windowSize.Height

	var fromX, fromY, toX, toY int
	switch direction {
	case "up":
		fromX, fromY, toX, toY = width/2, height*3/4, width/2, height*1/4
	case "down":
		fromX, fromY, toX, toY = width/2, height*1/4, width/2, height*3/4
	case "left":
		fromX, fromY, toX, toY = width*3/4, height/2, width*1/4, height/2
	case "right":
		fromX, fromY, toX, toY = width*1/4, height/2, width*3/4, height/2
	}
	return dExt.WebDriver.Swipe(fromX, fromY, toX, toY)
}

type Condition func(driver *DriverExt) error

func (dExt *DriverExt) SwipeUntil(direction string, condition Condition, maxTimes int) error {
	for i := 0; i < maxTimes; i++ {
		err := condition(dExt)
		if err == nil {
			return nil
		}
		err = dExt.SwipeTo(direction)
		if err != nil {
			log.Error().Err(err).Msgf("swipe %s failed", direction)
		}
	}
	return fmt.Errorf("swipe %s %d times, run condition failed", direction, maxTimes)
}

func (dExt *DriverExt) Swipe(pathname string, toX, toY int) (err error) {
	return dExt.SwipeFloat(pathname, float64(toX), float64(toY))
}

func (dExt *DriverExt) SwipeFloat(pathname string, toX, toY float64) (err error) {
	return dExt.SwipeOffsetFloat(pathname, toX, toY, 0.5, 0.5)
}

func (dExt *DriverExt) SwipeOffset(pathname string, toX, toY int, xOffset, yOffset float64) (err error) {
	return dExt.SwipeOffsetFloat(pathname, float64(toX), float64(toY), xOffset, yOffset)
}

func (dExt *DriverExt) SwipeOffsetFloat(pathname string, toX, toY, xOffset, yOffset float64) (err error) {
	var x, y, width, height float64
	if x, y, width, height, err = dExt.FindUIRectInUIKit(pathname); err != nil {
		return err
	}

	fromX := x + width*xOffset
	fromY := y + height*yOffset

	return dExt.WebDriver.SwipeFloat(fromX, fromY, toX, toY)
}

func (dExt *DriverExt) SwipeUp(pathname string, distance ...float64) (err error) {
	return dExt.SwipeUpOffset(pathname, 0.5, 0.9, distance...)
}

func (dExt *DriverExt) SwipeUpOffset(pathname string, xOffset, yOffset float64, distance ...float64) (err error) {
	if len(distance) == 0 {
		distance = []float64{1.0}
	}

	var x, y, width, height float64
	if x, y, width, height, err = dExt.FindUIRectInUIKit(pathname); err != nil {
		return err
	}

	fromX := x + width*xOffset
	fromY := (y + height) - height*(1.0-yOffset)

	toX := fromX
	toY := fromY - height*distance[0]

	return dExt.WebDriver.SwipeFloat(fromX, fromY, toX, toY)
}

func (dExt *DriverExt) SwipeDown(pathname string, distance ...float64) (err error) {
	return dExt.SwipeDownOffset(pathname, 0.5, 0.1, distance...)
}

func (dExt *DriverExt) SwipeDownOffset(pathname string, xOffset, yOffset float64, distance ...float64) (err error) {
	if len(distance) == 0 {
		distance = []float64{1.0}
	}

	var x, y, width, height float64
	if x, y, width, height, err = dExt.FindUIRectInUIKit(pathname); err != nil {
		return err
	}

	fromX := x + width*xOffset
	fromY := y + height*yOffset

	toX := fromX
	toY := fromY + height*distance[0]

	return dExt.WebDriver.SwipeFloat(fromX, fromY, toX, toY)
}

func (dExt *DriverExt) SwipeLeft(pathname string, distance ...float64) (err error) {
	return dExt.SwipeLeftOffset(pathname, 0.9, 0.5, distance...)
}

func (dExt *DriverExt) SwipeLeftOffset(pathname string, xOffset, yOffset float64, distance ...float64) (err error) {
	if len(distance) == 0 {
		distance = []float64{1.0}
	}

	var x, y, width, height float64
	if x, y, width, height, err = dExt.FindUIRectInUIKit(pathname); err != nil {
		return err
	}

	fromX := x + width*xOffset
	fromY := y + height*yOffset

	toX := fromX - width*distance[0]
	toY := fromY

	return dExt.WebDriver.SwipeFloat(fromX, fromY, toX, toY)
}

func (dExt *DriverExt) SwipeRight(pathname string, distance ...float64) (err error) {
	return dExt.SwipeRightOffset(pathname, 0.1, 0.5, distance...)
}

func (dExt *DriverExt) SwipeRightOffset(pathname string, xOffset, yOffset float64, distance ...float64) (err error) {
	if len(distance) == 0 {
		distance = []float64{1.0}
	}

	var x, y, width, height float64
	if x, y, width, height, err = dExt.FindUIRectInUIKit(pathname); err != nil {
		return err
	}

	fromX := x + width*xOffset
	fromY := y + height*yOffset

	toX := fromX + width*distance[0]
	toY := fromY

	return dExt.WebDriver.SwipeFloat(fromX, fromY, toX, toY)
}
