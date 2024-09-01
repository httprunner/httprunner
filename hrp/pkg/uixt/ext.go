package uixt

import (
	"bytes"
	"fmt"
	_ "image/gif"
	_ "image/png"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/httprunner/funplugin"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
	"github.com/httprunner/httprunner/v4/hrp/internal/env"
)

type cacheStepData struct {
	// cache step screenshot paths
	screenShots []string
	// cache step screenshot ocr results, key is image path, value is ScreenResult
	screenResults ScreenResultMap
	// cache e2e delay
	e2eDelay []timeLog
}

func (d *cacheStepData) reset() {
	d.screenShots = make([]string, 0)
	d.screenResults = make(map[string]*ScreenResult)
	d.e2eDelay = nil
}

type DriverExt struct {
	Device          Device
	Driver          WebDriver
	WindowSize      Size
	frame           *bytes.Buffer
	doneMjpegStream chan bool
	ImageService    IImageService // used to extract image data
	interruptSignal chan os.Signal

	// cache step data
	cacheStepData cacheStepData

	// funplugin
	plugin funplugin.IPlugin
}

func newDriverExt(device Device, driver WebDriver, options ...DriverOption) (dExt *DriverExt, err error) {
	driverOptions := NewDriverOptions()
	for _, option := range options {
		option(driverOptions)
	}

	dExt = &DriverExt{
		Device:          device,
		Driver:          driver,
		plugin:          driverOptions.plugin,
		cacheStepData:   cacheStepData{},
		interruptSignal: make(chan os.Signal, 1),
	}

	dExt.cacheStepData.reset()
	signal.Notify(dExt.interruptSignal, syscall.SIGTERM, syscall.SIGINT)
	dExt.doneMjpegStream = make(chan bool, 1)

	// get device window size
	dExt.WindowSize, err = dExt.Driver.WindowSize()
	if err != nil {
		return nil, errors.Wrap(err, "get screen resolution failed")
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

func (dExt *DriverExt) InstallByUrl(url string, opts *InstallOptions) error {
	// 获取当前目录
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// 将文件保存到当前目录
	appPath := filepath.Join(cwd, fmt.Sprint(time.Now().UnixNano())) // 替换为你想保存的文件名
	err = builtin.DownloadFile(appPath, url)
	if err != nil {
		return err
	}

	err = dExt.Install(appPath, opts)
	if err != nil {
		return err
	}
	return nil
}

func (dExt *DriverExt) Uninstall(packageName string, options ...ActionOption) error {
	actionOptions := NewActionOptions(options...)
	err := dExt.Device.Uninstall(packageName)
	if err != nil {
		log.Warn().Err(err).Msg("failed to uninstall")
	}
	if actionOptions.IgnoreNotFoundError {
		return nil
	}
	return err
}

func (dExt *DriverExt) Install(filePath string, opts *InstallOptions) error {
	if _, ok := dExt.Device.(*AndroidDevice); ok {
		stopChan := make(chan struct{})
		go func() {
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					actions := []TapTextAction{
						{Text: "^.*无视风险安装$", Options: []ActionOption{WithTapOffset(100, 0), WithRegex(true), WithIgnoreNotFoundError(true)}},
						{Text: "^已了解此应用未经检测.*", Options: []ActionOption{WithTapOffset(-450, 0), WithRegex(true), WithIgnoreNotFoundError(true)}},
					}
					_ = dExt.Driver.TapByTexts(actions...)
					_ = dExt.TapByOCR("^(.*无视风险安装|确定|继续|完成|点击继续安装|继续安装旧版本|替换|安装|授权本次安装|继续安装|重新安装)$", WithRegex(true), WithIgnoreNotFoundError(true))
				case <-stopChan:
					fmt.Println("Ticker stopped")
					return
				}
			}
		}()
		defer func() {
			close(stopChan)
		}()
	}

	return dExt.Device.Install(filePath, opts)
}

func (dExt *DriverExt) GetStepCacheData() map[string]interface{} {
	cacheData := make(map[string]interface{})
	cacheData["screenshots"] = dExt.cacheStepData.screenShots

	cacheData["screenshots_urls"] = dExt.cacheStepData.screenResults.getScreenShotUrls()
	cacheData["screen_results"] = dExt.cacheStepData.screenResults
	cacheData["e2e_results"] = dExt.cacheStepData.e2eDelay
	cacheData["driver_request_results"] = dExt.Driver.GetDriverResults()
	// clear cache
	dExt.cacheStepData.reset()
	return cacheData
}

// isPathExists returns true if path exists, whether path is file or dir
func isPathExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func (dExt *DriverExt) FindUIRectInUIKit(search string, options ...ActionOption) (point PointF, err error) {
	// click on text, using OCR
	if !isPathExists(search) {
		return dExt.FindScreenText(search, options...)
	}
	err = errors.New("ocr text not found")
	return
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
