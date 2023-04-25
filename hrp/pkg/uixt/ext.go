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
	"path/filepath"
	"strings"
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

type DriverExt struct {
	Device          Device
	Driver          WebDriver
	windowSize      Size
	frame           *bytes.Buffer
	doneMjpegStream chan bool
	scale           float64
	OCRService      IOCRService // used to get text from image
	screenShots     []string    // cache screenshot paths

	CVArgs
}

func NewDriverExt(device Device, driver WebDriver) (dExt *DriverExt, err error) {
	dExt = &DriverExt{
		Device: device,
		Driver: driver,
	}
	dExt.doneMjpegStream = make(chan bool, 1)

	// get device window size
	dExt.windowSize, err = dExt.Driver.WindowSize()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get windows size")
	}

	if dExt.scale, err = dExt.Driver.Scale(); err != nil {
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
func (dExt *DriverExt) TakeScreenShot(fileName ...string) (raw *bytes.Buffer, err error) {
	// wait for action done
	time.Sleep(500 * time.Millisecond)

	// iOS 优先使用 MJPEG 流进行截图，性能最优
	// 如果 MJPEG 流未开启，则使用 WebDriver 的截图接口
	if dExt.frame != nil {
		return dExt.frame, nil
	}
	if raw, err = dExt.Driver.Screenshot(); err != nil {
		log.Error().Err(err).Msg("capture screenshot data failed")
		return nil, err
	}

	// save screenshot to file
	if len(fileName) > 0 && fileName[0] != "" {
		path := filepath.Join(env.ScreenShotsPath, fileName[0])
		path, err := dExt.saveScreenShot(raw, path)
		if err != nil {
			log.Error().Err(err).Msg("save screenshot file failed")
			return nil, err
		}
		dExt.screenShots = append(dExt.screenShots, path)
		log.Info().Str("path", path).Msg("save screenshot file success")
	}

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

	return screenshotPath, nil
}

func (dExt *DriverExt) GetScreenShots() []string {
	defer func() {
		dExt.screenShots = nil
	}()
	return dExt.screenShots
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

func (dExt *DriverExt) FindUIRectInUIKit(search string, options ...DataOption) (point PointF, err error) {
	// click on text, using OCR
	if !isPathExists(search) {
		return dExt.FindScreenTextByOCR(search, options...)
	}
	// click on image, using opencv
	return dExt.FindImageRectInUIKit(search, options...)
}

func (dExt *DriverExt) MappingToRectInUIKit(rect image.Rectangle) (x, y, width, height float64) {
	x, y = float64(rect.Min.X)/dExt.scale, float64(rect.Min.Y)/dExt.scale
	width, height = float64(rect.Dx())/dExt.scale, float64(rect.Dy())/dExt.scale
	return
}

func (dExt *DriverExt) IsOCRExist(text string) bool {
	_, err := dExt.FindScreenTextByOCR(text)
	return err == nil
}

func (dExt *DriverExt) IsImageExist(text string) bool {
	_, err := dExt.FindImageRectInUIKit(text)
	return err == nil
}

func (dExt *DriverExt) IsAppInForeground(packageName string) bool {
	// check if app is in foreground
	yes, err := dExt.Driver.IsAppInForeground(packageName)
	if !yes || err != nil {
		log.Info().Str("packageName", packageName).Msg("app is not in foreground")
		return false
	}
	return true
}

func (dExt *DriverExt) getAbsScope(x1, y1, x2, y2 float64) (int, int, int, int) {
	return int(x1 * float64(dExt.windowSize.Width) * dExt.scale),
		int(y1 * float64(dExt.windowSize.Height) * dExt.scale),
		int(x2 * float64(dExt.windowSize.Width) * dExt.scale),
		int(y2 * float64(dExt.windowSize.Height) * dExt.scale)
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
		result = (dExt.IsAppInForeground(expected) == exp)
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
