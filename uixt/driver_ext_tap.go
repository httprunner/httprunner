package uixt

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/uixt/ai"
	"github.com/httprunner/httprunner/v5/uixt/option"
)

func (dExt *XTDriver) TapByOCR(text string, opts ...option.ActionOption) error {
	actionOptions := option.NewActionOptions(opts...)
	log.Info().Str("text", text).Interface("options", actionOptions).Msg("TapByOCR")

	if actionOptions.ScreenShotFileName == "" {
		opts = append(opts, option.WithScreenShotFileName(fmt.Sprintf("%s_tap_by_ocr_%s", dExt.GetDevice().UUID(), time.Now().Format("20060102150405"))))
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
		Interface("tapPoint", point).Msg("TapByOCR")

	return dExt.TapAbsXY(point.X, point.Y, opts...)
}

func (dExt *XTDriver) TapByCV(opts ...option.ActionOption) error {
	actionOptions := option.NewActionOptions(opts...)
	log.Info().Interface("options", actionOptions).Msg("TapByCV")
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
		Interface("tapPoint", point).Msg("TapByCV")

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
