package uixt

import (
	"fmt"

	"github.com/httprunner/httprunner/v5/uixt/ai"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/rs/zerolog/log"
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
	log.Info().Str("text", text).Interface("rawTextRect", textRect).
		Interface("tapPoint", point).Msg("TapByOCR success")

	return dExt.TapAbsXY(point.X, point.Y, opts...)
}

func (dExt *XTDriver) TapByCV(opts ...option.ActionOption) error {
	actionOptions := option.NewActionOptions(opts...)

	uiResult, err := dExt.FindUIResult(opts...)
	if err != nil {
		if actionOptions.IgnoreNotFoundError {
			return nil
		}
		return err
	}

	var point ai.PointF
	if actionOptions.TapRandomRect {
		point = uiResult.RandomPoint()
	} else {
		point = uiResult.Center()
	}
	log.Info().Interface("rawUIResult", uiResult).
		Interface("tapPoint", point).Msg("TapByCV success")

	return dExt.TapAbsXY(point.X, point.Y, opts...)
}

func (dExt *XTDriver) SecondaryClickByOCR(ocrText string, opts ...option.ActionOption) error {
	actionOptions := option.NewActionOptions(opts...)
	point, err := dExt.FindScreenText(ocrText, opts...)
	if err != nil {
		if actionOptions.IgnoreNotFoundError {
			return nil
		}
		return err
	}
	return dExt.SecondaryClick(point.Center().X, point.Center().Y)
}
