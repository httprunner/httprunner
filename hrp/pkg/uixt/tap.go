package uixt

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/httprunner/httprunner/v4/hrp/code"
)

func (dExt *DriverExt) TapAbsXY(x, y float64, options ...ActionOption) error {
	// tap on absolute coordinate [x, y]
	err := dExt.Driver.TapFloat(x, y, options...)
	return errors.Wrap(code.MobileUITapError, err.Error())
}

func (dExt *DriverExt) TapXY(x, y float64, options ...ActionOption) error {
	// tap on [x, y] percent of window size
	if x > 1 || y > 1 {
		return fmt.Errorf("x, y percentage should be <= 1, got x=%v, y=%v", x, y)
	}

	windowSize, err := dExt.Driver.WindowSize()
	if err != nil {
		return err
	}
	x = x * float64(windowSize.Width)
	y = y * float64(windowSize.Height)
	return dExt.TapAbsXY(x, y, options...)
}

func (dExt *DriverExt) TapByOCR(ocrText string, options ...ActionOption) error {
	actionOptions := NewActionOptions(options...)

	point, err := dExt.FindScreenText(ocrText, options...)
	if err != nil {
		if actionOptions.IgnoreNotFoundError {
			return nil
		}
		return err
	}

	return dExt.TapAbsXY(point.X, point.Y, options...)
}

func (dExt *DriverExt) TapByUIDetection(options ...ActionOption) error {
	actionOptions := NewActionOptions(options...)

	point, err := dExt.FindUIResult(options...)
	if err != nil {
		if actionOptions.IgnoreNotFoundError {
			return nil
		}
		return err
	}

	return dExt.TapAbsXY(point.X, point.Y, options...)
}

func (dExt *DriverExt) Tap(param string, options ...ActionOption) error {
	return dExt.TapOffset(param, 0, 0, options...)
}

func (dExt *DriverExt) TapOffset(param string, xOffset, yOffset float64, options ...ActionOption) (err error) {
	actionOptions := NewActionOptions(options...)

	point, err := dExt.FindUIRectInUIKit(param, options...)
	if err != nil {
		if actionOptions.IgnoreNotFoundError {
			return nil
		}
		return err
	}

	return dExt.TapAbsXY(point.X+xOffset, point.Y+yOffset, options...)
}

func (dExt *DriverExt) DoubleTapXY(x, y float64, options ...ActionOption) error {
	// double tap on coordinate: [x, y] should be relative
	if x > 1 || y > 1 {
		return fmt.Errorf("x, y percentage should be < 1, got x=%v, y=%v", x, y)
	}

	windowSize, err := dExt.Driver.WindowSize()
	if err != nil {
		return err
	}
	x = x * float64(windowSize.Width)
	y = y * float64(windowSize.Height)
	err = dExt.Driver.DoubleTapFloat(x, y, options...)
	return errors.Wrap(code.MobileUITapError, err.Error())
}

func (dExt *DriverExt) DoubleTap(param string, options ...ActionOption) (err error) {
	return dExt.DoubleTapOffset(param, 0, 0, options...)
}

func (dExt *DriverExt) DoubleTapOffset(param string, xOffset, yOffset float64, options ...ActionOption) (err error) {
	point, err := dExt.FindUIRectInUIKit(param)
	if err != nil {
		return err
	}

	err = dExt.Driver.DoubleTapFloat(point.X+xOffset, point.Y+yOffset, options...)
	return errors.Wrap(code.MobileUITapError, err.Error())
}
