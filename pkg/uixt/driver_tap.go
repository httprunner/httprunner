package uixt

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
)

func (dExt *XTDriver) TapByOCR(text string, opts ...option.ActionOption) error {
	actionOptions := option.NewActionOptions(opts...)
	if actionOptions.ScreenShotFileName == "" {
		opts = append(opts, option.WithScreenShotFileName(fmt.Sprintf("tap_by_ocr_%s", text)))
	}

	point, err := dExt.FindScreenText(text, opts...)
	if err != nil {
		if actionOptions.IgnoreNotFoundError {
			return nil
		}
		return err
	}

	return dExt.TapAbsXY(point.X, point.Y, opts...)
}

func (dExt *XTDriver) TapByCV(opts ...option.ActionOption) error {
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
