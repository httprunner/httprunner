package uixt

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
)

func (dExt *XTDriver) TapAbsXY(x, y float64, opts ...option.ActionOption) error {
	// tap on absolute coordinate [x, y]
	err := dExt.TapXY(x, y, opts...)
	if err != nil {
		return errors.Wrap(code.MobileUITapError, err.Error())
	}
	return nil
}

func (dExt *XTDriver) TapXY(x, y float64, opts ...option.ActionOption) error {
	// tap on [x, y] percent of window size
	if x > 1 || y > 1 {
		return fmt.Errorf("x, y percentage should be <= 1, got x=%v, y=%v", x, y)
	}

	windowSize, err := dExt.WindowSize()
	if err != nil {
		return err
	}
	x = x * float64(windowSize.Width)
	y = y * float64(windowSize.Height)
	return dExt.TapAbsXY(x, y, opts...)
}

func (dExt *XTDriver) TapByOCR(ocrText string, opts ...option.ActionOption) error {
	actionOptions := option.NewActionOptions(opts...)
	if actionOptions.ScreenShotFileName == "" {
		opts = append(opts, option.WithScreenShotFileName(fmt.Sprintf("tap_by_ocr_%s", ocrText)))
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

func (dExt *XTDriver) TapByUIDetection(opts ...option.ActionOption) error {
	options := option.NewActionOptions(opts...)

	point, err := dExt.FindUIResult(opts...)
	if err != nil {
		if options.IgnoreNotFoundError {
			return nil
		}
		return err
	}

	return dExt.TapAbsXY(point.X, point.Y, opts...)
}

func (dExt *XTDriver) Tap(param string, opts ...option.ActionOption) error {
	return dExt.TapOffset(param, 0, 0, opts...)
}

func (dExt *XTDriver) TapOffset(param string, xOffset, yOffset float64, opts ...option.ActionOption) (err error) {
	options := option.NewActionOptions(opts...)

	point, err := dExt.FindUIRectInUIKit(param, opts...)
	if err != nil {
		if options.IgnoreNotFoundError {
			return nil
		}
		return err
	}

	return dExt.TapAbsXY(point.X+xOffset, point.Y+yOffset, opts...)
}

func (dExt *XTDriver) DoubleTapXY(x, y float64, opts ...option.ActionOption) error {
	// double tap on coordinate: [x, y] should be relative
	if x > 1 || y > 1 {
		return fmt.Errorf("x, y percentage should be < 1, got x=%v, y=%v", x, y)
	}

	windowSize, err := dExt.WindowSize()
	if err != nil {
		return err
	}
	x = x * float64(windowSize.Width)
	y = y * float64(windowSize.Height)
	err = dExt.DoubleTapXY(x, y, opts...)
	if err != nil {
		return errors.Wrap(code.MobileUITapError, err.Error())
	}
	return nil
}

func (dExt *XTDriver) DoubleTap(param string, opts ...option.ActionOption) (err error) {
	return dExt.DoubleTapOffset(param, 0, 0, opts...)
}

func (dExt *XTDriver) DoubleTapOffset(param string, xOffset, yOffset float64, opts ...option.ActionOption) (err error) {
	point, err := dExt.FindUIRectInUIKit(param)
	if err != nil {
		return err
	}

	err = dExt.DoubleTapXY(point.X+xOffset, point.Y+yOffset, opts...)
	if err != nil {
		return errors.Wrap(code.MobileUITapError, err.Error())
	}
	return nil
}
