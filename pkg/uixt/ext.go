package uixt

import (
	"context"
	"fmt"
	_ "image/gif"
	_ "image/png"

	"github.com/httprunner/funplugin"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/internal/config"
	"github.com/httprunner/httprunner/v5/pkg/uixt/options"
)

type DriverExt struct {
	Ctx          context.Context
	Device       IDevice
	Driver       IWebDriver
	ImageService IImageService // used to extract image data

	// funplugin
	plugin funplugin.IPlugin
}

func newDriverExt(device IDevice, driver IWebDriver, opts ...options.DriverOption) (dExt *DriverExt, err error) {
	driverOptions := options.NewDriverOptions(opts...)

	dExt = &DriverExt{
		Device: device,
		Driver: driver,
		plugin: driverOptions.Plugin,
	}

	if driverOptions.WithImageService {
		if dExt.ImageService, err = newVEDEMImageService(); err != nil {
			return nil, err
		}
	}
	if driverOptions.WithResultFolder {
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

func (dExt *DriverExt) Init() error {
	// unlock device screen
	err := dExt.Driver.Unlock()
	if err != nil {
		log.Error().Err(err).Msg("unlock device screen failed")
		return err
	}

	return nil
}

func (dExt *DriverExt) assertOCR(text, assert string) error {
	var options []ActionOption
	options = append(options, WithScreenShotFileName(fmt.Sprintf("assert_ocr_%s", text)))

	switch assert {
	case AssertionEqual:
		_, err := dExt.FindScreenText(text, options...)
		if err != nil {
			return errors.Wrap(err, "assert ocr equal failed")
		}
	case AssertionNotEqual:
		_, err := dExt.FindScreenText(text, options...)
		if err == nil {
			return errors.New("assert ocr not equal failed")
		}
	case AssertionExists:
		options = append(options, WithRegex(true))
		_, err := dExt.FindScreenText(text, options...)
		if err != nil {
			return errors.Wrap(err, "assert ocr exists failed")
		}
	case AssertionNotExists:
		options = append(options, WithRegex(true))
		_, err := dExt.FindScreenText(text, options...)
		if err == nil {
			return errors.New("assert ocr not exists failed")
		}
	default:
		return fmt.Errorf("unexpected assert method %s", assert)
	}
	return nil
}

func (dExt *DriverExt) assertForegroundApp(appName, assert string) (err error) {
	err = dExt.Driver.AssertForegroundApp(appName)
	switch assert {
	case AssertionEqual:
		if err != nil {
			return errors.Wrap(err, "assert foreground app equal failed")
		}
	case AssertionNotEqual:
		if err == nil {
			return errors.New("assert foreground app not equal failed")
		}
	default:
		return fmt.Errorf("unexpected assert method %s", assert)
	}
	return nil
}

func (dExt *DriverExt) DoValidation(check, assert, expected string, message ...string) (err error) {
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
