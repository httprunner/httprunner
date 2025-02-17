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
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/internal/config"
	"github.com/httprunner/httprunner/v5/pkg/uixt/ai"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
	"github.com/httprunner/httprunner/v5/pkg/uixt/types"
)

type ScreenResult struct {
	bufSource   *bytes.Buffer  // raw image buffer bytes
	ImagePath   string         `json:"image_path"` // image file path
	Resolution  types.Size     `json:"resolution"`
	UploadedURL string         `json:"uploaded_url"` // uploaded image url
	Texts       ai.OCRTexts    `json:"texts"`        // dumped raw OCRTexts
	Icons       ai.UIResultMap `json:"icons"`        // CV 识别的图标
	Tags        []string       `json:"tags"`         // tags for image, e.g. ["feed", "ad", "live"]
	Popup       *PopupInfo     `json:"popup,omitempty"`
}

func (s *ScreenResult) FilterTextsByScope(x1, y1, x2, y2 float64) ai.OCRTexts {
	if x1 > 1 || y1 > 1 || x2 > 1 || y2 > 1 {
		log.Warn().Msg("x1, y1, x2, y2 should be in percentage, skip filter scope")
		return s.Texts
	}
	return s.Texts.FilterScope(option.AbsScope{
		int(float64(s.Resolution.Width) * x1), int(float64(s.Resolution.Height) * y1),
		int(float64(s.Resolution.Width) * x2), int(float64(s.Resolution.Height) * y2),
	})
}

// GetScreenResult takes a screenshot, returns the image recognition result
func (dExt *XTDriver) GetScreenResult(opts ...option.ActionOption) (screenResult *ScreenResult, err error) {
	screenshotOptions := option.NewActionOptions(opts...)

	var fileName string
	optionsList := screenshotOptions.List()
	if screenshotOptions.ScreenShotFileName != "" {
		fileName = builtin.GenNameWithTimestamp("%d_" + screenshotOptions.ScreenShotFileName)
	} else if len(optionsList) != 0 {
		fileName = builtin.GenNameWithTimestamp("%d_" + strings.Join(optionsList, "_"))
	} else {
		fileName = builtin.GenNameWithTimestamp("%d_screenshot")
	}

	var bufSource *bytes.Buffer
	var imageResult *ai.CVResult
	var imagePath string
	var windowSize types.Size
	var lastErr error

	// get screenshot info with retry
	for i := 0; i < 3; i++ {
		imagePath = filepath.Join(config.ScreenShotsPath, fileName)
		bufSource, err = dExt.ScreenShot(option.WithScreenShotFileName(imagePath))
		if err != nil {
			lastErr = err
			time.Sleep(time.Second * 1)
			continue
		}

		windowSize, err = dExt.WindowSize()
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
		imageResult, err = dExt.CVService.ReadFromBuffer(bufSource, opts...)
		if err != nil {
			log.Error().Err(err).Msg("ReadFromBuffer from ImageService failed")
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
	dExt.screenResults = append(dExt.screenResults, screenResult)

	if imageResult != nil {
		screenResult.Texts = imageResult.OCRResult.ToOCRTexts()
		screenResult.UploadedURL = imageResult.URL
		screenResult.Icons = imageResult.UIResult

		if screenshotOptions.ScreenShotWithClosePopups && imageResult.ClosePopupsResult != nil {
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

func (dExt *XTDriver) GetScreenTexts(opts ...option.ActionOption) (ocrTexts ai.OCRTexts, err error) {
	options := option.NewActionOptions(opts...)
	if options.ScreenShotFileName == "" {
		opts = append(opts, option.WithScreenShotFileName("get_screen_texts"))
	}
	opts = append(opts, option.WithScreenShotOCR(true), option.WithScreenShotUpload(true))
	screenResult, err := dExt.GetScreenResult(opts...)
	if err != nil {
		return
	}
	return screenResult.Texts, nil
}

func (dExt *XTDriver) FindScreenText(text string, opts ...option.ActionOption) (point ai.PointF, err error) {
	options := option.NewActionOptions(opts...)
	if options.ScreenShotFileName == "" {
		opts = append(opts, option.WithScreenShotFileName(fmt.Sprintf("find_screen_text_%s", text)))
	}
	ocrTexts, err := dExt.GetScreenTexts(opts...)
	if err != nil {
		return
	}

	result, err := ocrTexts.FindText(text, opts...)
	if err != nil {
		log.Warn().Msgf("FindText failed: %s", err.Error())
		return
	}
	point = result.Center()

	log.Info().Str("text", text).
		Interface("point", point).Msgf("FindScreenText success")
	return
}

func (dExt *XTDriver) FindUIResult(opts ...option.ActionOption) (point ai.PointF, err error) {
	options := option.NewActionOptions(opts...)
	if options.ScreenShotFileName == "" {
		opts = append(opts, option.WithScreenShotFileName(
			fmt.Sprintf("find_ui_result_%s", strings.Join(options.ScreenShotWithUITypes, "_"))))
	}

	screenResult, err := dExt.GetScreenResult(opts...)
	if err != nil {
		return
	}

	uiResults, err := screenResult.Icons.FilterUIResults(options.ScreenShotWithUITypes)
	if err != nil {
		return
	}
	uiResult, err := uiResults.GetUIResult(opts...)
	point = uiResult.Center()

	log.Info().Interface("text", options.ScreenShotWithUITypes).
		Interface("point", point).Msg("FindUIResult success")
	return
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
