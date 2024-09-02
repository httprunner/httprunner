package uixt

import (
	_ "image/gif"
	_ "image/png"

	"github.com/httprunner/funplugin"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
	"github.com/httprunner/httprunner/v4/hrp/internal/env"
)

type DriverExt struct {
	Device       Device
	Driver       WebDriver
	ImageService IImageService // used to extract image data

	// funplugin
	plugin funplugin.IPlugin
}

func newDriverExt(device Device, driver WebDriver, options ...DriverOption) (dExt *DriverExt, err error) {
	driverOptions := NewDriverOptions()
	for _, option := range options {
		option(driverOptions)
	}

	driver.GetSession().Clear()
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
		if err = builtin.EnsureFolderExists(env.ResultsPath); err != nil {
			return nil, errors.Wrap(err, "create results directory failed")
		}
		if err = builtin.EnsureFolderExists(env.ScreenShotsPath); err != nil {
			return nil, errors.Wrap(err, "create screenshots directory failed")
		}
	}
	return dExt, nil
}

func (dExt *DriverExt) AssertOCR(text, assert string) bool {
	var err error
	switch assert {
	case AssertionEqual:
		_, err = dExt.FindScreenText(text)
		return err == nil
	case AssertionNotEqual:
		_, err = dExt.FindScreenText(text)
		return err != nil
	case AssertionExists:
		_, err = dExt.FindScreenText(text, WithRegex(true))
		return err == nil
	case AssertionNotExists:
		_, err = dExt.FindScreenText(text, WithRegex(true))
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
