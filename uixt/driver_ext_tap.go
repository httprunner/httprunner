package uixt

import (
	"fmt"

	"github.com/httprunner/httprunner/v5/uixt/ai"
	"github.com/httprunner/httprunner/v5/uixt/option"
)

func (dExt *XTDriver) TapByOCR(text string, opts ...option.ActionOption) error {
	actionOptions := option.NewActionOptions(opts...)
	if actionOptions.ScreenShotFileName == "" {
		opts = append(opts, option.WithScreenShotFileName(fmt.Sprintf("tap_by_ocr_%s", text)))
	}

	textRect, err := dExt.FindScreenText(text, opts...)
	if err != nil {
		if actionOptions.IgnoreNotFoundError {
			return nil
		}
		return err
	}

	var point ai.PointF
	if actionOptions.TapRandomRect {
		point = textRect.RandomPoint()
	} else {
		point = textRect.Center()
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
