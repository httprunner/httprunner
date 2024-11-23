package uixt

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

func (dExt *DriverExt) Drag(fromX, fromY, toX, toY float64, options ...ActionOption) (err error) {
	return dExt.Driver.Drag(fromX, fromY, toX, toY, options...)
}

func (dExt *DriverExt) DragRelative(fromX, fromY, toX, toY float64, options ...ActionOption) (err error) {
	width := dExt.windowSize.Width
	height := dExt.windowSize.Height
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
