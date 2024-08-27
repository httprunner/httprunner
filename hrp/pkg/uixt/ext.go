package uixt

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
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

	"github.com/httprunner/httprunner/v4/hrp/code"
	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
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
	ImageResult *ImageResult  // image result

	Resolution  Size        `json:"resolution"`
	UploadedURL string      `json:"uploaded_url"` // uploaded image url
	Texts       OCRTexts    `json:"texts"`        // dumped raw OCRTexts
	Icons       UIResultMap `json:"icons"`        // CV 识别的图标
	Tags        []string    `json:"tags"`         // tags for image, e.g. ["feed", "ad", "live"]
	Video       *Video      `json:"video,omitempty"`
	Popup       *PopupInfo  `json:"popup,omitempty"`

	SwipeStartTime       int64 `json:"swipe_start_time"`        // 滑动开始时间戳
	SwipeFinishTime      int64 `json:"swipe_finish_time"`       // 滑动结束时间戳
	FetchVideoStartTime  int64 `json:"fetch_video_start_time"`  // 抓取视频开始时间戳
	FetchVideoFinishTime int64 `json:"fetch_video_finish_time"` // 抓取视频结束时间戳

	FetchVideoElapsed     int64 `json:"fetch_video_elapsed"`     // 抓取视频耗时(ms)
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
	e2eDelay     []timeLog
}

func (d *cacheStepData) reset() {
	d.screenShots = make([]string, 0)
	d.screenResults = make(map[string]*ScreenResult)
	d.videoCrawler = nil
	d.e2eDelay = nil
}

type DriverExt struct {
	CVArgs
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

	// cache last popup to check if popup handle result
	lastPopup *PopupInfo
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

	err = dExt.extendCV()
	if err != nil {
		return nil, errors.Wrap(code.MobileUIDriverError,
			fmt.Sprintf("extend OpenCV failed: %v", err))
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
	// 解码原始图像数据
	img, format, err := image.Decode(raw)
	if err != nil {
		return nil, err
	}

	// 创建一个用来保存压缩后数据的buffer
	var buf bytes.Buffer

	switch format {
	// Convert to jpeg uniformly and compress with a compression rate of 95
	case "jpeg", "png":
		jpegOptions := &jpeg.Options{Quality: 95}
		err = jpeg.Encode(&buf, img, jpegOptions)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported image format: %s", format)
	}

	// 返回压缩后的图像数据
	return &buf, nil
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

	// The default format uses jpeg for compression
	screenshotPath := filepath.Join(fmt.Sprintf("%s.%s", fileName, "jpeg"))
	file, err := os.Create(screenshotPath)
	if err != nil {
		return "", errors.Wrap(err, "create screenshot image file failed")
	}
	defer func() {
		_ = file.Close()
	}()

	switch format {
	case "jpeg", "png":
		jpegOptions := &jpeg.Options{}
		err = jpeg.Encode(file, img, jpegOptions)
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
