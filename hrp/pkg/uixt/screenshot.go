package uixt

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
	"github.com/httprunner/httprunner/v4/hrp/internal/env"
)

type ScreenResult struct {
	bufSource   *bytes.Buffer // raw image buffer bytes
	imagePath   string        // image file path
	ImageResult *ImageResult  // image result

	Resolution  Size        `json:"resolution"`
	UploadedURL string      `json:"uploaded_url"` // uploaded image url
	Texts       OCRTexts    `json:"texts"`        // dumped raw OCRTexts
	Icons       UIResultMap `json:"icons"`        // CV 识别的图标
	Tags        []string    `json:"tags"`         // tags for image, e.g. ["feed", "ad", "live"]
	Popup       *PopupInfo  `json:"popup,omitempty"`
}

type ScreenResultMap map[string]*ScreenResult // key is date time

// GetScreenResult takes a screenshot, returns the image recognition result
func (dExt *DriverExt) GetScreenResult(options ...ActionOption) (screenResult *ScreenResult, err error) {
	fileName := builtin.GenNameWithTimestamp("%d_screenshot")
	actionOptions := NewActionOptions(options...)
	screenshotActions := actionOptions.screenshotActions()
	if len(screenshotActions) != 0 {
		fileName = builtin.GenNameWithTimestamp("%d_" + strings.Join(screenshotActions, "_"))
	}
	bufSource, imagePath, err := dExt.GetScreenShot(fileName)
	if err != nil {
		return
	}

	screenResult = &ScreenResult{
		bufSource:  bufSource,
		imagePath:  imagePath,
		Tags:       nil,
		Resolution: dExt.WindowSize,
	}
	// cache screen result
	dExt.DataCache.addScreenResult(screenResult)

	imageResult, err := dExt.ImageService.GetImage(bufSource, options...)
	if err != nil {
		log.Error().Err(err).Msg("GetImage from ImageService failed")
		return screenResult, err
	}
	if imageResult != nil {
		screenResult.ImageResult = imageResult
		screenResult.Texts = imageResult.OCRResult.ToOCRTexts()
		screenResult.UploadedURL = imageResult.URL
		screenResult.Icons = imageResult.UIResult

		if actionOptions.ScreenShotWithClosePopups && imageResult.ClosePopupsResult != nil {
			screenResult.Popup = &PopupInfo{
				ClosePopupsResult: imageResult.ClosePopupsResult,
				PicName:           imagePath,
				PicURL:            imageResult.URL,
			}

			closeAreas, _ := imageResult.UIResult.FilterUIResults([]string{"close"})
			for _, closeArea := range closeAreas {
				screenResult.Popup.ClosePoints = append(screenResult.Popup.ClosePoints, closeArea.Center())
			}
		}
	}

	log.Debug().
		Str("imagePath", imagePath).
		Str("imageUrl", screenResult.UploadedURL).
		Msg("log screenshot")
	return screenResult, nil
}

func (dExt *DriverExt) GetScreenTexts() (ocrTexts OCRTexts, err error) {
	screenResult, err := dExt.GetScreenResult(
		WithScreenShotOCR(true), WithScreenShotUpload(true))
	if err != nil {
		return
	}
	return screenResult.Texts, nil
}

func (dExt *DriverExt) FindUIRectInUIKit(search string, options ...ActionOption) (point PointF, err error) {
	// find text using OCR
	if !builtin.IsPathExists(search) {
		return dExt.FindScreenText(search, options...)
	}
	// TODO: find image using CV
	err = errors.New("ocr text not found")
	return
}

func (dExt *DriverExt) FindScreenText(text string, options ...ActionOption) (point PointF, err error) {
	ocrTexts, err := dExt.GetScreenTexts()
	if err != nil {
		return
	}

	result, err := ocrTexts.FindText(text, dExt.ParseActionOptions(options...)...)
	if err != nil {
		log.Warn().Msgf("FindText failed: %s", err.Error())
		return
	}
	point = result.Center()

	log.Info().Str("text", text).
		Interface("point", point).Msgf("FindScreenText success")
	return
}

func (dExt *DriverExt) FindUIResult(options ...ActionOption) (point PointF, err error) {
	actionOptions := NewActionOptions(options...)

	screenResult, err := dExt.GetScreenResult(options...)
	if err != nil {
		return
	}

	uiResults, err := screenResult.Icons.FilterUIResults(actionOptions.ScreenShotWithUITypes)
	if err != nil {
		return
	}
	uiResult, err := uiResults.GetUIResult(dExt.ParseActionOptions(options...)...)
	point = uiResult.Center()

	log.Info().Interface("text", actionOptions.ScreenShotWithUITypes).
		Interface("point", point).Msg("FindUIResult success")
	return
}

// GetScreenShot takes screenshot and saves image file to $CWD/screenshots/ folder
func (dExt *DriverExt) GetScreenShot(fileName string) (raw *bytes.Buffer, path string, err error) {
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
	path, err = saveScreenShot(compressed, path)
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
func saveScreenShot(raw *bytes.Buffer, fileName string) (string, error) {
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

	log.Info().Str("path", screenshotPath).Msg("save screenshot file success")
	return screenshotPath, nil
}
