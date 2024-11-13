package uixt

import (
	"context"
	"fmt"
	_ "image/gif"
	_ "image/png"

	"github.com/httprunner/funplugin"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
	"github.com/httprunner/httprunner/v4/hrp/internal/config"
)

type DriverExt struct {
	Ctx          context.Context
	Device       IDevice
	Driver       IWebDriver
	ImageService IImageService // used to extract image data

	// funplugin
	plugin funplugin.IPlugin
}

func newDriverExt(device IDevice, driver IWebDriver, options ...DriverOption) (dExt *DriverExt, err error) {
	driverOptions := NewDriverOptions()
	for _, option := range options {
		option(driverOptions)
	}

	dExt = &DriverExt{
		Device: device,
		Driver: driver,
		plugin: driverOptions.plugin,
	}

	if driverOptions.withImageService {
		if dExt.ImageService, err = newVEDEMImageService(); err != nil {
			return nil, err
		}
	}
	if driverOptions.withResultFolder {
		// create results directory
		if err = builtin.EnsureFolderExists(config.ResultsPath); err != nil {
			return nil, errors.Wrap(err, "create results directory failed")
		}
		if err = builtin.EnsureFolderExists(config.ScreenShotsPath); err != nil {
			return nil, errors.Wrap(err, "create screenshots directory failed")
		}
	}
	return dExt, nil
}

func (dExt *DriverExt) AssertOCR(text, assert string) bool {
	var options []ActionOption
	options = append(options, WithScreenShotFileName(fmt.Sprintf("assert_ocr_%s", text)))

	var err error
	switch assert {
	case AssertionEqual:
		_, err = dExt.FindScreenText(text, options...)
		return err == nil
	case AssertionNotEqual:
		_, err = dExt.FindScreenText(text, options...)
		return err != nil
	case AssertionExists:
		options = append(options, WithRegex(true))
		_, err = dExt.FindScreenText(text, options...)
		return err == nil
	case AssertionNotExists:
		options = append(options, WithRegex(true))
		_, err = dExt.FindScreenText(text, options...)
		return err != nil
	default:
		log.Warn().Str("assert method", assert).Msg("unexpected assert method")
	}
	return false
}

func (dExt *DriverExt) AssertForegroundApp(appName, assert string) bool {
	app, err := dExt.Driver.GetForegroundApp()
	if err != nil {
		log.Warn().Err(err).Msg("get foreground app failed, skip app/activity assertion")
		return true // Notice: ignore error when get foreground app failed
	}
	log.Debug().Interface("app", app).Msg("get foreground app")

	// assert package name
	switch assert {
	case AssertionEqual:
		return app.PackageName == appName
	case AssertionNotEqual:
		return app.PackageName != appName
	default:
		log.Warn().Str("assert method", assert).Msg("unexpected assert method")
	}
	return false
}

func (dExt *DriverExt) DoValidation(check, assert, expected string, message ...string) bool {
	var result bool
	switch check {
	case SelectorOCR:
		result = dExt.AssertOCR(expected, assert)
	case SelectorForegroundApp:
		result = dExt.AssertForegroundApp(expected, assert)
	}

	if !result {
		if message == nil {
			message = []string{""}
		}
		log.Error().
			Str("assert", assert).
			Str("expect", expected).
			Str("msg", message[0]).
			Msg("validate UI failed")
		return false
	}

	log.Info().
		Str("assert", assert).
		Str("expect", expected).
		Msg("validate UI success")
	return true
}
