package uixt

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/pkg/uixt/options"
)

func (dExt *DriverExt) TapAbsXY(x, y float64, opts ...options.ActionOption) error {
	// tap on absolute coordinate [x, y]
	err := dExt.Driver.Tap(x, y, opts...)
	if err != nil {
		return errors.Wrap(code.MobileUITapError, err.Error())
	}
	return nil
}

func (dExt *DriverExt) TapXY(x, y float64, opts ...options.ActionOption) error {
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
	return dExt.TapAbsXY(x, y, opts...)
}

func (dExt *DriverExt) TapByOCR(ocrText string, opts ...options.ActionOption) error {
	actionOptions := options.NewActionOptions(opts...)
	if actionOptions.ScreenShotFileName == "" {
		opts = append(opts, options.WithScreenShotFileName(fmt.Sprintf("tap_by_ocr_%s", ocrText)))
	}

	point, err := dExt.FindScreenText(ocrText, opts...)
	if err != nil {
		if actionOptions.IgnoreNotFoundError {
			return nil
		}
		return err
	}

	return dExt.TapAbsXY(point.X, point.Y, opts...)
}

func (dExt *DriverExt) TapByUIDetection(opts ...options.ActionOption) error {
	actionOptions := options.NewActionOptions(opts...)

	point, err := dExt.FindUIResult(opts...)
	if err != nil {
		if actionOptions.IgnoreNotFoundError {
			return nil
		}
		return err
	}

	return dExt.TapAbsXY(point.X, point.Y, opts...)
}

func (dExt *DriverExt) Tap(param string, opts ...options.ActionOption) error {
	return dExt.TapOffset(param, 0, 0, opts...)
}

func (dExt *DriverExt) TapOffset(param string, xOffset, yOffset float64, opts ...options.ActionOption) (err error) {
	actionOptions := options.NewActionOptions(opts...)

	point, err := dExt.FindUIRectInUIKit(param, opts...)
	if err != nil {
		if actionOptions.IgnoreNotFoundError {
			return nil
		}
		return err
	}

	return dExt.TapAbsXY(point.X+xOffset, point.Y+yOffset, opts...)
}

func (dExt *DriverExt) DoubleTapXY(x, y float64, opts ...options.ActionOption) error {
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
	err = dExt.Driver.DoubleTap(x, y, opts...)
	if err != nil {
		return errors.Wrap(code.MobileUITapError, err.Error())
	}
	return nil
}

func (dExt *DriverExt) DoubleTap(param string, opts ...options.ActionOption) (err error) {
	return dExt.DoubleTapOffset(param, 0, 0, opts...)
}

func (dExt *DriverExt) DoubleTapOffset(param string, xOffset, yOffset float64, opts ...options.ActionOption) (err error) {
	point, err := dExt.FindUIRectInUIKit(param)
	if err != nil {
		return err
	}

	err = dExt.Driver.DoubleTap(point.X+xOffset, point.Y+yOffset, opts...)
	if err != nil {
		return errors.Wrap(code.MobileUITapError, err.Error())
	}
	return nil
}
