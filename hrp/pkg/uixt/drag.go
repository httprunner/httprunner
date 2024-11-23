package uixt

import (
	"fmt"

	"github.com/httprunner/httprunner/v4/hrp/code"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func (dExt *DriverExt) Drag(fromX, fromY, toX, toY float64, options ...ActionOption) (err error) {
	windowSize, err := dExt.Driver.WindowSize()
	if err != nil {
		return errors.Wrap(code.DeviceGetInfoError, err.Error())
	}
	width := windowSize.Width
	height := windowSize.Height

	orientation, err := dExt.Driver.Orientation()
	if err != nil {
		log.Warn().Err(err).Msgf("drag from (%v, %v) to (%v, %v) get orientation failed, use default orientation",
			fromX, fromY, toX, toY)
		orientation = OrientationPortrait
	}

	if !assertRelative(fromX) || !assertRelative(fromY) ||
		!assertRelative(toX) || !assertRelative(toY) {
		return fmt.Errorf("fromX(%f), fromY(%f), toX(%f), toY(%f) must be less than 1",
			fromX, fromY, toX, toY)
	}
	// 左转和右转都是"LANDSCAPE"
	if orientation == OrientationPortrait {
		fromX = float64(width) * fromX
		fromY = float64(height) * fromY
		toX = float64(width) * toX
		toY = float64(height) * toY
	} else {
		fromX = float64(height) * fromX
		fromY = float64(width) * fromY
		toX = float64(height) * toX
		toY = float64(width) * toY
	}

	return dExt.Driver.Drag(fromX, fromY, toX, toY, options...)
}
