package uixt

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/internal/config"
	"github.com/httprunner/httprunner/v5/uixt/ai"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/httprunner/httprunner/v5/uixt/types"
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

	// take screenshot
	bufSource, err := dExt.ScreenShot()
	if err != nil {
		return nil, errors.Wrapf(code.DeviceScreenShotError,
			"take screenshot failed %v", err)
	}

	// compress screenshot
	compressBufSource, err := compressImageBuffer(bufSource)
	if err != nil {
		return nil, errors.Wrapf(code.DeviceScreenShotError,
			"compress screenshot failed %v", err)
	}

	// save compressed screenshot to file
	var fileName string
	optionsList := screenshotOptions.List()
	if screenshotOptions.ScreenShotFileName != "" {
		fileName = builtin.GenNameWithTimestamp("%d_" + screenshotOptions.ScreenShotFileName)
	} else if len(optionsList) != 0 {
		fileName = builtin.GenNameWithTimestamp("%d_" + strings.Join(optionsList, "_"))
	} else {
		fileName = builtin.GenNameWithTimestamp("%d_screenshot")
	}
	imagePath := filepath.Join(
		config.GetConfig().ScreenShotsPath,
		fmt.Sprintf("%s.%s", fileName, "jpeg"),
	)
	go func() {
		err := saveScreenShot(compressBufSource, imagePath)
		if err != nil {
			log.Error().Err(err).Msg("save screenshot file failed")
		} else {
			log.Info().Str("path", imagePath).Msg("screenshot saved")
		}
	}()

	windowSize, err := dExt.WindowSize()
	if err != nil {
		return nil, errors.Wrap(code.DeviceGetInfoError, err.Error())
	}

	// read image from buffer with CV
	screenResult = &ScreenResult{
		bufSource:  compressBufSource,
		ImagePath:  imagePath,
		Tags:       nil,
		Resolution: windowSize,
	}
	imageResult, err := dExt.CVService.ReadFromBuffer(compressBufSource, opts...)
	if err != nil {
		log.Error().Err(err).Msg("ReadFromBuffer from ImageService failed")
		return nil, err
	}
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

	// cache screen result
	dExt.screenResults = append(dExt.screenResults, screenResult)

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

func (dExt *XTDriver) FindScreenText(text string, opts ...option.ActionOption) (textRect ai.OCRText, err error) {
	options := option.NewActionOptions(opts...)
	if options.ScreenShotFileName == "" {
		opts = append(opts, option.WithScreenShotFileName(fmt.Sprintf("find_screen_text_%s", text)))
	}

	// convert relative scope to absolute scope
	if options.AbsScope == nil && len(options.Scope) == 4 {
		windowSize, err := dExt.WindowSize()
		if err != nil {
			return ai.OCRText{}, err
		}
		absScope := option.AbsScope{
			int(options.Scope[0] * float64(windowSize.Width)),
			int(options.Scope[1] * float64(windowSize.Height)),
			int(options.Scope[2] * float64(windowSize.Width)),
			int(options.Scope[3] * float64(windowSize.Height)),
		}
		opts = append(opts, option.WithAbsScope(
			absScope[0], absScope[1], absScope[2], absScope[3]))
		log.Info().Interface("scope", options.Scope).
			Interface("absScope", absScope).Msg("convert to abs scope")
	}

	ocrTexts, err := dExt.GetScreenTexts(opts...)
	if err != nil {
		return
	}

	textRect, err = ocrTexts.FindText(text, opts...)
	if err != nil {
		log.Warn().Msgf("FindText failed: %s", err.Error())
		return
	}

	log.Info().Str("text", text).
		Interface("textRect", textRect).Msgf("FindScreenText success")
	return textRect, nil
}

func (dExt *XTDriver) FindUIResult(opts ...option.ActionOption) (uiResult ai.UIResult, err error) {
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
	uiResult, err = uiResults.GetUIResult(opts...)

	log.Info().Interface("text", options.ScreenShotWithUITypes).
		Interface("uiResult", uiResult).Msg("FindUIResult success")
	return
}

// saveScreenShot saves compressed image file with file name
func saveScreenShot(raw *bytes.Buffer, screenshotPath string) error {
	// notice: screenshot data is a stream, so we need to copy it to a new buffer
	copiedBuffer := &bytes.Buffer{}
	if _, err := copiedBuffer.Write(raw.Bytes()); err != nil {
		log.Error().Err(err).Msg("copy screenshot buffer failed")
	}

	img, format, err := image.Decode(copiedBuffer)
	if err != nil {
		return errors.Wrap(err, "decode screenshot image failed")
	}

	file, err := os.Create(screenshotPath)
	if err != nil {
		return errors.Wrap(err, "create screenshot image file failed")
	}
	defer func() {
		_ = file.Close()
	}()

	// compress image and save to file
	switch format {
	case "jpeg":
		jpegOptions := &jpeg.Options{Quality: 95}
		err = jpeg.Encode(file, img, jpegOptions)
	case "png":
		encoder := png.Encoder{
			CompressionLevel: png.BestCompression,
		}
		err = encoder.Encode(file, img)
	case "gif":
		gifOptions := &gif.Options{
			NumColors: 256,
		}
		err = gif.Encode(file, img, gifOptions)
	default:
		return fmt.Errorf("unsupported image format %s", format)
	}
	if err != nil {
		return errors.Wrap(err, "save image file failed")
	}

	var fileSize int64
	fileInfo, err := file.Stat()
	if err == nil {
		fileSize = fileInfo.Size()
	}
	log.Info().Str("path", screenshotPath).
		Int("rawBytes", raw.Len()).Int64("saveBytes", fileSize).
		Msg("save screenshot file success")

	return nil
}

func compressImageBuffer(raw *bytes.Buffer) (compressed *bytes.Buffer, err error) {
	// decode image from buffer
	img, format, err := image.Decode(raw)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer

	switch format {
	// compress image
	case "jpeg", "png":
		jpegOptions := &jpeg.Options{Quality: 60}
		err = jpeg.Encode(&buf, img, jpegOptions)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported image format: %s", format)
	}

	// return compressed image buffer
	return &buf, nil
}
