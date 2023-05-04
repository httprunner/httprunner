package uixt

import (
	"bytes"
	"encoding/json"
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

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

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

type Popularity struct {
	Stars     string `json:"stars"`     // 点赞数
	Comments  string `json:"comments"`  // 评论数
	Favorites string `json:"favorites"` // 收藏数
	Shares    string `json:"shares"`    // 分享数
}

type OcrResult struct {
	Texts      OCRTexts   `json:"texts"`      // dumped OCRTexts
	Tags       []string   `json:"tags"`       // tags for image, e.g. ["feed", "ad", "live"]
	Popularity Popularity `json:"popularity"` // video popularity data
}

type cacheStepData struct {
	// cache step screenshot paths
	ScreenShots []string
	// cache step screenshot ocr results, key is image path, value is OcrResult
	OcrResults map[string]*OcrResult
	// cache feed/live video stat
	VideoStat *VideoStat
}

func (d *cacheStepData) reset() {
	d.ScreenShots = make([]string, 0)
	d.OcrResults = make(map[string]*OcrResult)
	d.VideoStat = nil
}

type DriverExt struct {
	CVArgs
	Device          Device
	Driver          WebDriver
	windowSize      Size
	frame           *bytes.Buffer
	doneMjpegStream chan bool
	OCRService      IOCRService // used to get texts from image
	interruptSignal chan os.Signal

	// cache step data
	cacheStepData cacheStepData
}

func NewDriverExt(device Device, driver WebDriver) (dExt *DriverExt, err error) {
	dExt = &DriverExt{
		Device: device,
		Driver: driver,
		cacheStepData: cacheStepData{
			ScreenShots: make([]string, 0),
			OcrResults:  make(map[string]*OcrResult),
		},
		interruptSignal: make(chan os.Signal, 1),
	}
	signal.Notify(dExt.interruptSignal, syscall.SIGTERM, syscall.SIGINT)
	dExt.doneMjpegStream = make(chan bool, 1)

	// get device window size
	dExt.windowSize, err = dExt.Driver.WindowSize()
	if err != nil {
		return nil, err
	}

	if dExt.OCRService, err = newVEDEMOCRService(); err != nil {
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

// TakeScreenShot takes screenshot and saves image file to $CWD/screenshots/ folder
// if fileName is empty, it will not save image file and only return raw image data
func (dExt *DriverExt) TakeScreenShot(fileName ...string) (raw *bytes.Buffer, path string, err error) {
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
	if len(fileName) > 0 && fileName[0] != "" {
		path := filepath.Join(env.ScreenShotsPath, fileName[0])
		path, err := dExt.saveScreenShot(compressed, path)
		if err != nil {
			log.Error().Err(err).Msg("save screenshot file failed")
			return nil, "", err
		}
		return compressed, path, nil
	}

	return compressed, "", nil
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

	dExt.cacheStepData.ScreenShots = append(dExt.cacheStepData.ScreenShots, screenshotPath)
	log.Info().Str("path", screenshotPath).Msg("save screenshot file success")
	return screenshotPath, nil
}

func (dExt *DriverExt) GetStepCacheData() map[string]interface{} {
	cacheData := make(map[string]interface{})
	cacheData["video_stat"] = dExt.cacheStepData.VideoStat
	cacheData["screenshots"] = dExt.cacheStepData.ScreenShots

	ocrResults := make(map[string]interface{})
	for imagePath, ocrResult := range dExt.cacheStepData.OcrResults {
		o, _ := json.Marshal(ocrResult.Texts)
		data := map[string]interface{}{
			"tags":  ocrResult.Tags,
			"texts": string(o),
		}
		ocrResults[imagePath] = data
	}
	cacheData["ocr_results"] = ocrResults

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
		return dExt.FindScreenTextByOCR(search, options...)
	}
	// click on image, using opencv
	return dExt.FindImageRectInUIKit(search, options...)
}

func (dExt *DriverExt) IsOCRExist(text string) bool {
	_, err := dExt.FindScreenTextByOCR(text)
	return err == nil
}

func (dExt *DriverExt) IsImageExist(text string) bool {
	_, err := dExt.FindImageRectInUIKit(text)
	return err == nil
}

func (dExt *DriverExt) DoValidation(check, assert, expected string, message ...string) bool {
	var exp bool
	if assert == AssertionExists || assert == AssertionEqual {
		exp = true
	} else {
		exp = false
	}
	var result bool
	switch check {
	case SelectorOCR:
		result = (dExt.IsOCRExist(expected) == exp)
	case SelectorImage:
		result = (dExt.IsImageExist(expected) == exp)
	case SelectorForegroundApp:
		result = ((dExt.Driver.AssertAppForeground(expected) == nil) == exp)
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
