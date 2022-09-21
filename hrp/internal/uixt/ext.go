package uixt

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/electricbubble/gwda"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// TemplateMatchMode is the type of the template matching operation.
type TemplateMatchMode int

type CVArgs struct {
	scale     float64
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
	gwda.WebDriver
	windowSize      gwda.Size
	frame           *bytes.Buffer
	doneMjpegStream chan bool

	CVArgs
}

func extend(driver gwda.WebDriver) (dExt *DriverExt, err error) {
	dExt = &DriverExt{WebDriver: driver}
	dExt.doneMjpegStream = make(chan bool, 1)

	// get device window size
	dExt.windowSize, err = dExt.WebDriver.WindowSize()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get windows size")
	}

	return dExt, nil
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

func (dExt *DriverExt) takeScreenShot() (raw *bytes.Buffer, err error) {
	// 优先使用 MJPEG 流进行截图，性能最优
	// 如果 MJPEG 流未开启，则使用 WebDriver 的截图接口
	if dExt.frame != nil {
		return dExt.frame, nil
	}
	if raw, err = dExt.WebDriver.Screenshot(); err != nil {
		log.Error().Err(err).Msgf("screenshot failed: %v", err)
		return nil, err
	}
	return
}

// saveScreenShot saves image file to $CWD/screenshots/ folder
func (dExt *DriverExt) saveScreenShot(raw *bytes.Buffer, fileName string) (string, error) {
	img, format, err := image.Decode(raw)
	if err != nil {
		return "", errors.Wrap(err, "decode screenshot image failed")
	}

	dir, _ := os.Getwd()
	screenshotsDir := filepath.Join(dir, "screenshots")
	if err = os.MkdirAll(screenshotsDir, os.ModePerm); err != nil {
		return "", errors.Wrap(err, "create screenshots directory failed")
	}
	screenshotPath := filepath.Join(screenshotsDir,
		fmt.Sprintf("%s.%s", fileName, format))

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
	default:
		return "", fmt.Errorf("unsupported image format: %s", format)
	}
	if err != nil {
		return "", errors.Wrap(err, "encode screenshot image failed")
	}

	return screenshotPath, nil
}

// ScreenShot takes screenshot and saves image file to $CWD/screenshots/ folder
func (dExt *DriverExt) ScreenShot(fileName string) (string, error) {
	raw, err := dExt.takeScreenShot()
	if err != nil {
		return "", errors.Wrap(err, "screenshot by WDA failed")
	}

	return dExt.saveScreenShot(raw, fileName)
}

// isPathExists returns true if path exists, whether path is file or dir
func isPathExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func (dExt *DriverExt) FindUIElement(param string) (ele gwda.WebElement, err error) {
	var selector gwda.BySelector
	if strings.HasPrefix(param, "/") {
		// xpath
		selector = gwda.BySelector{
			XPath: param,
		}
	} else {
		// name
		selector = gwda.BySelector{
			LinkText: gwda.NewElementAttribute().WithName(param),
		}
	}

	return dExt.WebDriver.FindElement(selector)
}

func (dExt *DriverExt) FindUIRectInUIKit(search string) (x, y, width, height float64, err error) {
	// click on text, using OCR
	if !isPathExists(search) {
		return dExt.FindTextByOCR(search)
	}
	// click on image, using opencv
	return dExt.FindImageRectInUIKit(search)
}

func (dExt *DriverExt) PerformTouchActions(touchActions *gwda.TouchActions) error {
	return dExt.PerformAppiumTouchActions(touchActions)
}

func (dExt *DriverExt) PerformActions(actions *gwda.W3CActions) error {
	return dExt.PerformW3CActions(actions)
}

func (dExt *DriverExt) IsNameExist(name string) bool {
	selector := gwda.BySelector{
		LinkText: gwda.NewElementAttribute().WithName(name),
	}
	_, err := dExt.FindElement(selector)
	return err == nil
}

func (dExt *DriverExt) IsLabelExist(label string) bool {
	selector := gwda.BySelector{
		LinkText: gwda.NewElementAttribute().WithLabel(label),
	}
	_, err := dExt.FindElement(selector)
	return err == nil
}

func (dExt *DriverExt) IsOCRExist(text string) bool {
	_, _, _, _, err := dExt.FindTextByOCR(text)
	return err == nil
}

func (dExt *DriverExt) IsImageExist(text string) bool {
	_, _, _, _, err := dExt.FindImageRectInUIKit(text)
	return err == nil
}
