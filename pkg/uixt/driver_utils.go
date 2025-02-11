package uixt

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/httprunner/httprunner/v5/code"
)

func convertToAbsolutePoint(driver IDriver, x, y float64) (absX, absY float64, err error) {
	if !assertRelative(x) || !assertRelative(y) {
		err = errors.Wrap(code.InvalidCaseError,
			fmt.Sprintf("x(%f), y(%f) must be less than 1", x, y))
		return
	}

	windowSize, err := driver.WindowSize()
	if err != nil {
		err = errors.Wrap(code.DeviceGetInfoError, err.Error())
		return
	}

	absX = float64(windowSize.Width) * x
	absY = float64(windowSize.Height) * y
	return
}

func convertToAbsoluteCoordinates(driver IDriver, fromX, fromY, toX, toY float64) (
	absFromX, absFromY, absToX, absToY float64, err error) {

	if !assertRelative(fromX) || !assertRelative(fromY) ||
		!assertRelative(toX) || !assertRelative(toY) {
		err = errors.Wrap(code.InvalidCaseError,
			fmt.Sprintf("fromX(%f), fromY(%f), toX(%f), toY(%f) must be less than 1",
				fromX, fromY, toX, toY))
		return
	}

	windowSize, err := driver.WindowSize()
	if err != nil {
		err = errors.Wrap(code.DeviceGetInfoError, err.Error())
		return
	}
	width := windowSize.Width
	height := windowSize.Height

	absFromX = float64(width) * fromX
	absFromY = float64(height) * fromY
	absToX = float64(width) * toX
	absToY = float64(height) * toY

	return absFromX, absFromY, absToX, absToY, nil
}

func assertRelative(p float64) bool {
	return p >= 0 && p <= 1
}
