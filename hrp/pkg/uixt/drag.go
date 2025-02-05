package uixt

import (
	"fmt"

	"github.com/httprunner/httprunner/v5/hrp/code"
	"github.com/pkg/errors"
)

func (dExt *DriverExt) Drag(fromX, fromY, toX, toY float64, options ...ActionOption) (err error) {
	windowSize, err := dExt.Driver.WindowSize()
	if err != nil {
		return errors.Wrap(code.DeviceGetInfoError, err.Error())
	}
	width := windowSize.Width
	height := windowSize.Height

	if !assertRelative(fromX) || !assertRelative(fromY) ||
		!assertRelative(toX) || !assertRelative(toY) {
		return fmt.Errorf("fromX(%f), fromY(%f), toX(%f), toY(%f) must be less than 1",
			fromX, fromY, toX, toY)
	}
	fromX = float64(width) * fromX
	fromY = float64(height) * fromY
	toX = float64(width) * toX
	toY = float64(height) * toY

	return dExt.Driver.Drag(fromX, fromY, toX, toY, options...)
}
