package uixt

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/code"
	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
	"github.com/httprunner/httprunner/v4/hrp/internal/config"
)

type ScreenResult struct {
	bufSource   *bytes.Buffer // raw image buffer bytes
	ImagePath   string        `json:"image_path"` // image file path
	Resolution  Size          `json:"resolution"`
	UploadedURL string        `json:"uploaded_url"` // uploaded image url
	Texts       OCRTexts      `json:"texts"`        // dumped raw OCRTexts
	Icons       UIResultMap   `json:"icons"`        // CV 识别的图标
	Tags        []string      `json:"tags"`         // tags for image, e.g. ["feed", "ad", "live"]
	Popup       *PopupInfo    `json:"popup,omitempty"`
}

// GetScreenResult takes a screenshot, returns the image recognition result
func (dExt *DriverExt) GetScreenResult(options ...ActionOption) (screenResult *ScreenResult, err error) {
	actionOptions := NewActionOptions(options...)
	if actionOptions.MaxRetryTimes == 0 {
		actionOptions.MaxRetryTimes = 1
	}

	var fileName string
	screenshotActions := actionOptions.screenshotActions()
	if actionOptions.ScreenShotFileName != "" {
		fileName = builtin.GenNameWithTimestamp("%d_" + actionOptions.ScreenShotFileName)
	} else if len(screenshotActions) != 0 {
		fileName = builtin.GenNameWithTimestamp("%d_" + strings.Join(screenshotActions, "_"))
	} else {
		fileName = builtin.GenNameWithTimestamp("%d_screenshot")
	}

	var bufSource *bytes.Buffer
	var imageResult *ImageResult
	var imagePath string
	var windowSize Size
	var lastErr error

	// get screenshot info with retry
	for i := 0; i < actionOptions.MaxRetryTimes; i++ {
		bufSource, imagePath, err = dExt.GetScreenShot(fileName)
		if err != nil {
			lastErr = err
			continue
		}

		windowSize, err = dExt.Driver.WindowSize()
		if err != nil {
			lastErr = errors.Wrap(code.DeviceGetInfoError, err.Error())
			continue
		}

		screenResult = &ScreenResult{
			bufSource:  bufSource,
			ImagePath:  imagePath,
			Tags:       nil,
			Resolution: windowSize,
		}
		imageResult, err = dExt.ImageService.GetImage(bufSource, options...)
		if err != nil {
			log.Error().Err(err).Msg("GetImage from ImageService failed")
			lastErr = err
			continue
		}
		// success, break the loop
		lastErr = nil
		break
	}
	if lastErr != nil {
		return nil, lastErr
	}

	// cache screen result
	dExt.Driver.GetSession().addScreenResult(screenResult)

	if imageResult != nil {
		screenResult.Texts = imageResult.OCRResult.ToOCRTexts()
		screenResult.UploadedURL = imageResult.URL
		screenResult.Icons = imageResult.UIResult

		if actionOptions.ScreenShotWithClosePopups && imageResult.ClosePopupsResult != nil {
			if imageResult.ClosePopupsResult.IsEmpty() {
				// set nil to reduce unnecessary summary info
				imageResult.ClosePopupsResult = nil
			} else {
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
	}

	log.Debug().
		Str("imagePath", imagePath).
		Str("imageUrl", screenResult.UploadedURL).
		Msg("log screenshot")
	return screenResult, nil
}

func (dExt *DriverExt) GetScreenTexts(options ...ActionOption) (ocrTexts OCRTexts, err error) {
	actionOptions := NewActionOptions(options...)
	if actionOptions.ScreenShotFileName == "" {
		options = append(options, WithScreenShotFileName("get_screen_texts"))
	}
	options = append(options, WithScreenShotOCR(true), WithScreenShotUpload(true))
	screenResult, err := dExt.GetScreenResult(options...)
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
	actionOptions := NewActionOptions(options...)
	if actionOptions.ScreenShotFileName == "" {
		options = append(options, WithScreenShotFileName(fmt.Sprintf("find_screen_text_%s", text)))
	}
	ocrTexts, err := dExt.GetScreenTexts(options...)
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
	if actionOptions.ScreenShotFileName == "" {
		options = append(options, WithScreenShotFileName(
			fmt.Sprintf("find_ui_result_%s", strings.Join(actionOptions.ScreenShotWithUITypes, "_"))))
	}

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
	if raw, err = dExt.Driver.Screenshot(); err != nil {
		log.Error().Err(err).Msg("capture screenshot data failed")
		return nil, "", errors.Wrap(code.DeviceScreenShotError, err.Error())
	}

	// save screenshot to file
	path = filepath.Join(config.ScreenShotsPath, fileName)
	path, err = saveScreenShot(raw, path)
	if err != nil {
		log.Error().Err(err).Msg("save screenshot file failed")
		return nil, "", errors.Wrap(code.DeviceScreenShotError,
			fmt.Sprintf("save screenshot file failed: %s", err.Error()))
	}
	return raw, path, nil
}

// saveScreenShot saves compressed image file with file name
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

	// compress image and save to file
	switch format {
	case "jpeg", "png":
		jpegOptions := &jpeg.Options{Quality: 95}
		err = jpeg.Encode(file, img, jpegOptions)
	// case "png":
	// 	encoder := png.Encoder{
	// 		CompressionLevel: png.BestCompression,
	// 	}
	// 	err = encoder.Encode(file, img)
	case "gif":
		gifOptions := &gif.Options{
			NumColors: 256,
		}
		err = gif.Encode(file, img, gifOptions)
	default:
		return "", fmt.Errorf("unsupported image format %s", format)
	}
	if err != nil {
		return "", errors.Wrap(err, "save image file failed")
	}

	var fileSize int64
	fileInfo, err := file.Stat()
	if err == nil {
		fileSize = fileInfo.Size()
	}
	log.Info().Str("path", screenshotPath).
		Int("rawBytes", raw.Len()).Int64("saveBytes", fileSize).
		Msg("save screenshot file success")

	return screenshotPath, nil
}
