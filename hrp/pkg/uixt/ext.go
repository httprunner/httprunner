package uixt

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"math/rand"
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
	"github.com/httprunner/httprunner/v4/hrp/internal/code"
	"github.com/httprunner/httprunner/v4/hrp/internal/env"
)

// TemplateMatchMode is the type of the template matching operation.
type TemplateMatchMode int

type CVArgs struct {
	matchMode TemplateMatchMode
	threshold float64
}

type CVOption func(*CVArgs)

func WithTemplateMatchMode(mode TemplateMatchMode) CVOption {
	return func(args *CVArgs) {
		args.matchMode = mode
	}
}

func WithThreshold(threshold float64) CVOption {
	return func(args *CVArgs) {
		args.threshold = threshold
	}
}

type ScreenResult struct {
	bufSource   *bytes.Buffer // raw image buffer bytes
	imagePath   string        // image file path
	imageResult *ImageResult  // image result

	Resolution  Size        `json:"resolution"`
	UploadedURL string      `json:"uploaded_url"` // uploaded image url
	Texts       OCRTexts    `json:"texts"`        // dumped raw OCRTexts
	Icons       UIResultMap `json:"icons"`        // CV 识别的图标
	Tags        []string    `json:"tags"`         // tags for image, e.g. ["feed", "ad", "live"]
	Video       *Video      `json:"video,omitempty"`
	Popup       *PopupInfo  `json:"popup,omitempty"`

	SwipeStartTime  int64 `json:"swipe_start_time"`  // 滑动开始时间戳
	SwipeFinishTime int64 `json:"swipe_finish_time"` // 滑动结束时间戳

	ScreenshotTakeElapsed int64 `json:"screenshot_take_elapsed"` // 设备截图耗时(ms)
	ScreenshotCVElapsed   int64 `json:"screenshot_cv_elapsed"`   // CV 识别耗时(ms)

	// 当前 Feed/Live 整体耗时
	TotalElapsed int64 `json:"total_elapsed"` // current_swipe_finish -> next_swipe_start 整体耗时(ms)
}

type ScreenResultMap map[string]*ScreenResult // key is date time

// getScreenShotUrls returns screenShotsUrls using imagePath as key and uploaded URL as value
func (screenResults ScreenResultMap) getScreenShotUrls() map[string]string {
	screenShotsUrls := make(map[string]string)
	for _, screenResult := range screenResults {
		if screenResult.UploadedURL == "" {
			continue
		}
		screenShotsUrls[screenResult.imagePath] = screenResult.UploadedURL
	}
	return screenShotsUrls
}

type cacheStepData struct {
	// cache step screenshot paths
	screenShots []string
	// cache step screenshot ocr results, key is image path, value is ScreenResult
	screenResults ScreenResultMap
	// cache feed/live video stat
	videoCrawler *VideoCrawler
}

func (d *cacheStepData) reset() {
	d.screenShots = make([]string, 0)
	d.screenResults = make(map[string]*ScreenResult)
	d.videoCrawler = nil
}

type DriverExt struct {
	CVArgs
	Device          Device
	Driver          WebDriver
	windowSize      Size
	frame           *bytes.Buffer
	doneMjpegStream chan bool
	ImageService    IImageService // used to extract image data
	interruptSignal chan os.Signal

	// cache step data
	cacheStepData cacheStepData

	// funplugin
	plugin funplugin.IPlugin

	// cache last popup to check if popup handle result
	lastPopup *PopupInfo
}

func newDriverExt(device Device, driver WebDriver, plugin funplugin.IPlugin) (dExt *DriverExt, err error) {
	dExt = &DriverExt{
		Device:          device,
		Driver:          driver,
		plugin:          plugin,
		cacheStepData:   cacheStepData{},
		interruptSignal: make(chan os.Signal, 1),
	}

	err = dExt.extendCV()
	if err != nil {
		return nil, errors.Wrap(code.MobileUIDriverError,
			fmt.Sprintf("extend OpenCV failed: %v", err))
	}

	dExt.cacheStepData.reset()
	signal.Notify(dExt.interruptSignal, syscall.SIGTERM, syscall.SIGINT)
	dExt.doneMjpegStream = make(chan bool, 1)

	// get device window size
	dExt.windowSize, err = dExt.Driver.WindowSize()
	if err != nil {
		return nil, errors.Wrap(err, "get screen resolution failed")
	}

	if dExt.ImageService, err = newVEDEMImageService(); err != nil {
		return nil, err
	}

	// create results directory
	if err = builtin.EnsureFolderExists(env.ResultsPath); err != nil {
		return nil, errors.Wrap(err, "create results directory failed")
	}
	if err = builtin.EnsureFolderExists(env.ScreenShotsPath); err != nil {
		return nil, errors.Wrap(err, "create screenshots directory failed")
	}
	return dExt, nil
}

// takeScreenShot takes screenshot and saves image file to $CWD/screenshots/ folder
func (dExt *DriverExt) takeScreenShot(fileName string) (raw *bytes.Buffer, path string, err error) {
	// iOS 优先使用 MJPEG 流进行截图，性能最优
	// 如果 MJPEG 流未开启，则使用 WebDriver 的截图接口
	if dExt.frame != nil {
		return dExt.frame, "", nil
	}
	if raw, err = dExt.Driver.Screenshot(); err != nil {
		log.Error().Err(err).Msg("capture screenshot data failed")
		return nil, "", err
	}

	// compress image data
	compressed, err := compressImageBuffer(raw)
	if err != nil {
		log.Error().Err(err).Msg("compress screenshot data failed")
		return nil, "", err
	}

	// save screenshot to file
	path = filepath.Join(env.ScreenShotsPath, fileName)
	path, err = dExt.saveScreenShot(compressed, path)
	if err != nil {
		log.Error().Err(err).Msg("save screenshot file failed")
		return nil, "", err
	}
	return compressed, path, nil
}

func compressImageBuffer(raw *bytes.Buffer) (compressed *bytes.Buffer, err error) {
	// TODO: compress image data
	return raw, nil
}

// saveScreenShot saves image file with file name
func (dExt *DriverExt) saveScreenShot(raw *bytes.Buffer, fileName string) (string, error) {
	// notice: screenshot data is a stream, so we need to copy it to a new buffer
	copiedBuffer := &bytes.Buffer{}
	if _, err := copiedBuffer.Write(raw.Bytes()); err != nil {
		log.Error().Err(err).Msg("copy screenshot buffer failed")
	}

	img, format, err := image.Decode(copiedBuffer)
	if err != nil {
		return "", errors.Wrap(err, "decode screenshot image failed")
	}

	screenshotPath := filepath.Join(fmt.Sprintf("%s.%s", fileName, format))
	file, err := os.Create(screenshotPath)
	if err != nil {
		return "", errors.Wrap(err, "create screenshot image file failed")
	}
	defer func() {
		_ = file.Close()
	}()

	switch format {
	case "png":
		err = png.Encode(file, img)
	case "jpeg":
		err = jpeg.Encode(file, img, nil)
	case "gif":
		err = gif.Encode(file, img, nil)
	default:
		return "", fmt.Errorf("unsupported image format: %s", format)
	}
	if err != nil {
		return "", errors.Wrap(err, "encode screenshot image failed")
	}

	dExt.cacheStepData.screenShots = append(dExt.cacheStepData.screenShots, screenshotPath)
	log.Info().Str("path", screenshotPath).Msg("save screenshot file success")
	return screenshotPath, nil
}

func (dExt *DriverExt) GetStepCacheData() map[string]interface{} {
	cacheData := make(map[string]interface{})
	cacheData["video_stat"] = dExt.cacheStepData.videoCrawler
	cacheData["screenshots"] = dExt.cacheStepData.screenShots

	cacheData["screenshots_urls"] = dExt.cacheStepData.screenResults.getScreenShotUrls()
	cacheData["screen_results"] = dExt.cacheStepData.screenResults

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

func init() {
	rand.Seed(time.Now().UnixNano())
}

func (dExt *DriverExt) FindUIRectInUIKit(search string, options ...ActionOption) (point PointF, err error) {
	// click on text, using OCR
	if !isPathExists(search) {
		return dExt.FindScreenText(search, options...)
	}
	// click on image, using opencv
	return dExt.FindImageRectInUIKit(search, options...)
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

func (dExt *DriverExt) AssertImage(imagePath, assert string) bool {
	var err error
	switch assert {
	case AssertionExists:
		_, err = dExt.FindImageRectInUIKit(imagePath)
		return err == nil
	case AssertionNotExists:
		_, err = dExt.FindImageRectInUIKit(imagePath)
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
	case SelectorImage:
		result = dExt.AssertImage(expected, assert)
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
