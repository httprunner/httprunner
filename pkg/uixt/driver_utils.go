package uixt

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
)

func convertToAbsolutePoint(driver IDriver, x, y float64) (absX, absY float64, err error) {
	if !assertRelative(x) || !assertRelative(y) {
		err = errors.Wrap(code.InvalidCaseError,
			fmt.Sprintf("x(%f), y(%f) must be less than 1", x, y))
		return
	}

	windowSize, err := driver.WindowSize()
	if err != nil {
		err = errors.Wrap(code.DeviceGetInfoError, err.Error())
		return
	}

	absX = float64(windowSize.Width) * x
	absY = float64(windowSize.Height) * y
	return
}

func convertToAbsoluteCoordinates(driver IDriver, fromX, fromY, toX, toY float64) (
	absFromX, absFromY, absToX, absToY float64, err error) {

	if !assertRelative(fromX) || !assertRelative(fromY) ||
		!assertRelative(toX) || !assertRelative(toY) {
		err = errors.Wrap(code.InvalidCaseError,
			fmt.Sprintf("fromX(%f), fromY(%f), toX(%f), toY(%f) must be less than 1",
				fromX, fromY, toX, toY))
		return
	}

	windowSize, err := driver.WindowSize()
	if err != nil {
		err = errors.Wrap(code.DeviceGetInfoError, err.Error())
		return
	}
	width := windowSize.Width
	height := windowSize.Height

	absFromX = float64(width) * fromX
	absFromY = float64(height) * fromY
	absToX = float64(width) * toX
	absToY = float64(height) * toY

	return absFromX, absFromY, absToX, absToY, nil
}

func assertRelative(p float64) bool {
	return p >= 0 && p <= 1
}

func (dExt *XTDriver) Setup() error {
	// unlock device screen
	err := dExt.Unlock()
	if err != nil {
		log.Error().Err(err).Msg("unlock device screen failed")
		return err
	}

	return nil
}

func (dExt *XTDriver) assertOCR(text, assert string) error {
	var opts []option.ActionOption
	opts = append(opts, option.WithScreenShotFileName(fmt.Sprintf("assert_ocr_%s", text)))

	switch assert {
	case AssertionEqual:
		_, err := dExt.FindScreenText(text, opts...)
		if err != nil {
			return errors.Wrap(err, "assert ocr equal failed")
		}
	case AssertionNotEqual:
		_, err := dExt.FindScreenText(text, opts...)
		if err == nil {
			return errors.New("assert ocr not equal failed")
		}
	case AssertionExists:
		opts = append(opts, option.WithRegex(true))
		_, err := dExt.FindScreenText(text, opts...)
		if err != nil {
			return errors.Wrap(err, "assert ocr exists failed")
		}
	case AssertionNotExists:
		opts = append(opts, option.WithRegex(true))
		_, err := dExt.FindScreenText(text, opts...)
		if err == nil {
			return errors.New("assert ocr not exists failed")
		}
	default:
		return fmt.Errorf("unexpected assert method %s", assert)
	}
	return nil
}

func (dExt *XTDriver) assertForegroundApp(appName, assert string) error {
	app, err := dExt.ForegroundInfo()
	if err != nil {
		log.Warn().Err(err).Msg("get foreground app failed, skip app assertion")
		return nil // Notice: ignore error when get foreground app failed
	}

	switch assert {
	case AssertionEqual:
		if app.PackageName != appName {
			return errors.Wrap(err, "assert foreground app equal failed")
		}
	case AssertionNotEqual:
		if app.PackageName == appName {
			return errors.New("assert foreground app not equal failed")
		}
	default:
		return fmt.Errorf("unexpected assert method %s", assert)
	}
	return nil
}

func (dExt *XTDriver) DoValidation(check, assert, expected string, message ...string) (err error) {
	switch check {
	case SelectorOCR:
		err = dExt.assertOCR(expected, assert)
	case SelectorForegroundApp:
		err = dExt.assertForegroundApp(expected, assert)
	}

	if err != nil {
		if message == nil {
			message = []string{""}
		}
		log.Error().Err(err).Str("assert", assert).Str("expect", expected).
			Str("msg", message[0]).Msg("validate failed")
		return err
	}

	log.Info().Str("assert", assert).Str("expect", expected).Msg("validate success")
	return nil
}
