package uixt

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/electricbubble/gwda"
	cvHelper "github.com/electricbubble/opencv-helper"
	"github.com/pkg/errors"
)

// TemplateMatchMode is the type of the template matching operation.
type TemplateMatchMode int

const (
	// TmSqdiff maps to TM_SQDIFF
	TmSqdiff TemplateMatchMode = iota
	// TmSqdiffNormed maps to TM_SQDIFF_NORMED
	TmSqdiffNormed
	// TmCcorr maps to TM_CCORR
	TmCcorr
	// TmCcorrNormed maps to TM_CCORR_NORMED
	TmCcorrNormed
	// TmCcoeff maps to TM_CCOEFF
	TmCcoeff
	// TmCcoeffNormed maps to TM_CCOEFF_NORMED
	TmCcoeffNormed
)

type DebugMode int

const (
	// DmOff no output
	DmOff DebugMode = iota
	// DmEachMatch output matched and mismatched values
	DmEachMatch
	// DmNotMatch output only values that do not match
	DmNotMatch
)

type DriverExt struct {
	gwda.WebDriver
	windowSize      gwda.Size
	scale           float64
	MatchMode       TemplateMatchMode
	Threshold       float64
	frame           *bytes.Buffer
	doneMjpegStream chan bool
}

// Extend 获得扩展后的 Driver，
// 并指定匹配阀值，
// 获取当前设备的 Scale，
// 默认匹配模式为 TmCcoeffNormed，
// 默认关闭 OpenCV 匹配值计算后的输出
func Extend(driver gwda.WebDriver, threshold float64, matchMode ...TemplateMatchMode) (dExt *DriverExt, err error) {
	dExt = &DriverExt{WebDriver: driver}
	dExt.doneMjpegStream = make(chan bool, 1)

	if dExt.scale, err = dExt.Scale(); err != nil {
		return &DriverExt{}, err
	}

	// get device window size
	dExt.windowSize, err = dExt.WebDriver.WindowSize()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get windows size")
	}

	if len(matchMode) == 0 {
		matchMode = []TemplateMatchMode{TmCcoeffNormed}
	}
	dExt.MatchMode = matchMode[0]
	cvHelper.Debug(cvHelper.DebugMode(DmOff))
	dExt.Threshold = threshold
	return dExt, nil
}

func (dExt *DriverExt) OnlyOnceThreshold(threshold float64) (newExt *DriverExt) {
	newExt = new(DriverExt)
	newExt.WebDriver = dExt.WebDriver
	newExt.scale = dExt.scale
	newExt.MatchMode = dExt.MatchMode
	newExt.Threshold = threshold
	return
}

func (dExt *DriverExt) OnlyOnceMatchMode(matchMode TemplateMatchMode) (newExt *DriverExt) {
	newExt = new(DriverExt)
	newExt.WebDriver = dExt.WebDriver
	newExt.scale = dExt.scale
	newExt.MatchMode = matchMode
	newExt.Threshold = dExt.Threshold
	return
}

func (dExt *DriverExt) Debug(dm DebugMode) {
	cvHelper.Debug(cvHelper.DebugMode(dm))
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

// func (sExt *DriverExt) findImgRect(search string) (rect image.Rectangle, err error) {
// 	pathSource := filepath.Join(sExt.pathname, cvHelper.GenFilename())
// 	if err = sExt.driver.ScreenshotToDisk(pathSource); err != nil {
// 		return image.Rectangle{}, err
// 	}
//
// 	if rect, err = cvHelper.FindImageRectFromDisk(pathSource, search, float32(sExt.Threshold), cvHelper.TemplateMatchMode(sExt.MatchMode)); err != nil {
// 		return image.Rectangle{}, err
// 	}
// 	return
// }

func (dExt *DriverExt) FindAllImageRect(search string) (rects []image.Rectangle, err error) {
	var bufSource, bufSearch *bytes.Buffer
	if bufSearch, err = getBufFromDisk(search); err != nil {
		return nil, err
	}
	if bufSource, err = dExt.takeScreenShot(); err != nil {
		return nil, err
	}

	if rects, err = cvHelper.FindAllImageRectsFromRaw(bufSource, bufSearch, float32(dExt.Threshold), cvHelper.TemplateMatchMode(dExt.MatchMode)); err != nil {
		return nil, err
	}
	return
}

func getBufFromDisk(name string) (*bytes.Buffer, error) {
	var f *os.File
	var err error
	if f, err = os.Open(name); err != nil {
		return nil, err
	}
	var all []byte
	if all, err = ioutil.ReadAll(f); err != nil {
		return nil, err
	}
	return bytes.NewBuffer(all), nil
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

func (dExt *DriverExt) FindImageRectInUIKit(search string) (x, y, width, height float64, err error) {
	var bufSource, bufSearch *bytes.Buffer
	if bufSearch, err = getBufFromDisk(search); err != nil {
		return 0, 0, 0, 0, err
	}
	if bufSource, err = dExt.takeScreenShot(); err != nil {
		return 0, 0, 0, 0, err
	}

	var rect image.Rectangle
	if rect, err = cvHelper.FindImageRectFromRaw(bufSource, bufSearch, float32(dExt.Threshold), cvHelper.TemplateMatchMode(dExt.MatchMode)); err != nil {
		return 0, 0, 0, 0, err
	}

	// if rect, err = dExt.findImgRect(search); err != nil {
	// 	return 0, 0, 0, 0, err
	// }
	x, y, width, height = dExt.MappingToRectInUIKit(rect)
	return
}

func (dExt *DriverExt) MappingToRectInUIKit(rect image.Rectangle) (x, y, width, height float64) {
	x, y = float64(rect.Min.X)/dExt.scale, float64(rect.Min.Y)/dExt.scale
	width, height = float64(rect.Dx())/dExt.scale, float64(rect.Dy())/dExt.scale
	return
}

func (dExt *DriverExt) PerformTouchActions(touchActions *gwda.TouchActions) error {
	return dExt.PerformAppiumTouchActions(touchActions)
}

func (dExt *DriverExt) PerformActions(actions *gwda.W3CActions) error {
	return dExt.PerformW3CActions(actions)
}

// IsExist
