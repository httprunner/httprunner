package uixt

import (
	"bytes"
	_ "image/gif"
	_ "image/png"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

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

	frame           *bytes.Buffer
	doneMjpegStream chan bool
	interruptSignal chan os.Signal
}

func newDriverExt(device Device, driver WebDriver, options ...DriverOption) (dExt *DriverExt, err error) {
	driverOptions := NewDriverOptions()
	for _, option := range options {
		option(driverOptions)
	}

	driver.GetSession().Clear()
	dExt = &DriverExt{
		Device:          device,
		Driver:          driver,
		plugin:          driverOptions.plugin,
		interruptSignal: make(chan os.Signal, 1),
	}

	signal.Notify(dExt.interruptSignal, syscall.SIGTERM, syscall.SIGINT)
	dExt.doneMjpegStream = make(chan bool, 1)

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

func (dExt *DriverExt) ConnectMjpegStream(httpClient *http.Client) (err error) {
	if httpClient == nil {
		return errors.New(`'httpClient' can't be nil`)
	}

	var req *http.Request
	if req, err = http.NewRequest(http.MethodGet, "http://*", nil); err != nil {
		return err
	}

	var resp *http.Response
	if resp, err = httpClient.Do(req); err != nil {
		return err
	}
	// defer func() { _ = resp.Body.Close() }()

	var boundary string
	if _, param, err := mime.ParseMediaType(resp.Header.Get("Content-Type")); err != nil {
		return err
	} else {
		boundary = strings.Trim(param["boundary"], "-")
	}

	mjpegReader := multipart.NewReader(resp.Body, boundary)

	go func() {
		for {
			select {
			case <-dExt.doneMjpegStream:
				_ = resp.Body.Close()
				return
			default:
				var part *multipart.Part
				if part, err = mjpegReader.NextPart(); err != nil {
					dExt.frame = nil
					continue
				}

				raw := new(bytes.Buffer)
				if _, err = raw.ReadFrom(part); err != nil {
					dExt.frame = nil
					continue
				}
				dExt.frame = raw
			}
		}
	}()

	return
}

func (dExt *DriverExt) CloseMjpegStream() {
	dExt.doneMjpegStream <- true
}
