package uixt

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/httprunner/httprunner/v5/code"
)

func convertToAbsoluteCoordinates(driver IDriver, fromX, fromY, toX, toY float64) (
	absFromX, absFromY, absToX, absToY float64, err error) {

	err = assertCoordinatesRelative(fromX, fromY, toX, toY)
	if err != nil {
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

func assertCoordinatesRelative(fromX, fromY, toX, toY float64) error {
	if !assertRelative(fromX) || !assertRelative(fromY) ||
		!assertRelative(toX) || !assertRelative(toY) {
		return errors.Wrap(code.InvalidCaseError,
			fmt.Sprintf("fromX(%f), fromY(%f), toX(%f), toY(%f) must be less than 1",
				fromX, fromY, toX, toY))
	}
	return nil
}

func assertRelative(p float64) bool {
	return p >= 0 && p <= 1
}
