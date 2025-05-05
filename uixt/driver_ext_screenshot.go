package uixt

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

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

func (dExt *XTDriver) GetScreenShotBuffer() (compressedBufSource *bytes.Buffer, err error) {
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

	return compressBufSource, nil
}

// GetScreenResult takes a screenshot, returns the image recognition result
func (dExt *XTDriver) GetScreenResult(opts ...option.ActionOption) (screenResult *ScreenResult, err error) {
	// get compressed screenshot buffer
	compressBufSource, err := dExt.GetScreenShotBuffer()
	if err != nil {
		return nil, err
	}

	screenshotOptions := option.NewActionOptions(opts...)

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

	// save screen result to session
	session := dExt.GetSession()
	session.screenResults = append(session.screenResults, screenResult)

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

// MarkUIOperation add operation mark for UI operation
func MarkUIOperation(driver IDriver, actionType ActionMethod, actionCoordinates []float64) error {
	if actionType == "" || len(actionCoordinates) == 0 {
		return nil
	}
	start := time.Now()

	// get screenshot
	compressedBufSource, err := driver.ScreenShot()
	if err != nil {
		return err
	}

	// create screenshot save path
	timestamp := builtin.GenNameWithTimestamp("action_%d")
	var imagePath string

	if actionType == ACTION_TapAbsXY {
		if len(actionCoordinates) != 2 {
			return fmt.Errorf("invalid tap action coordinates: %v", actionCoordinates)
		}
		imagePath = filepath.Join(
			config.GetConfig().ScreenShotsPath,
			fmt.Sprintf("%s_tap.png", timestamp),
		)
		x, y := actionCoordinates[0], actionCoordinates[1]
		point := image.Point{X: int(x), Y: int(y)}
		err = SaveImageWithCircleMarker(compressedBufSource, point, imagePath)
	} else if actionType == ACTION_Swipe || actionType == ACTION_Drag {
		if len(actionCoordinates) != 4 {
			return fmt.Errorf("invalid swipe action coordinates: %v", actionCoordinates)
		}
		imagePath = filepath.Join(
			config.GetConfig().ScreenShotsPath,
			fmt.Sprintf("%s_%s.png", timestamp, actionType),
		)
		fromX, fromY := actionCoordinates[0], actionCoordinates[1]
		toX, toY := actionCoordinates[2], actionCoordinates[3]
		from := image.Point{X: int(fromX), Y: int(fromY)}
		to := image.Point{X: int(toX), Y: int(toY)}
		err = SaveImageWithArrowMarker(compressedBufSource, from, to, imagePath)
	}
	if err != nil {
		log.Error().Err(err).
			Int64("duration(ms)", time.Since(start).Milliseconds()).
			Msg("mark UI operation failed")
		return err
	}

	if imagePath != "" {
		log.Info().Str("operation", string(actionType)).
			Str("imagePath", imagePath).
			Int64("duration(ms)", time.Since(start).Milliseconds()).
			Msg("mark UI operation success")

		// save screenshot to session
		session := driver.GetSession()
		session.screenResults = append(session.screenResults, &ScreenResult{
			bufSource: compressedBufSource,
			ImagePath: imagePath,
		})
	}

	return nil
}

// SaveImageWithCircleMarker saves an image with circle marker
func SaveImageWithCircleMarker(imgBuf *bytes.Buffer, point image.Point, outputPath string) error {
	img, _, err := image.Decode(imgBuf)
	if err != nil {
		return fmt.Errorf("failed to decode image data: %w", err)
	}
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	// draw a red circle at the tap point
	centerX := point.X
	centerY := point.Y
	radius := 20
	lineWidth := 5
	red := color.RGBA{255, 0, 0, 255}

	for angle := 0.0; angle < 2*math.Pi; angle += 0.01 {
		for w := 0; w < lineWidth; w++ {
			r := float64(radius - w)
			x := int(float64(centerX) + r*math.Cos(angle))
			y := int(float64(centerY) + r*math.Sin(angle))
			if x >= 0 && x < bounds.Max.X && y >= 0 && y < bounds.Max.Y {
				rgba.Set(x, y, red)
			}
		}
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()
	if err := png.Encode(outFile, rgba); err != nil {
		return fmt.Errorf("failed to encode and save image: %w", err)
	}
	return nil
}

// SaveImageWithArrowMarker saves an image with an arrow marker
func SaveImageWithArrowMarker(imgBuf *bytes.Buffer, from, to image.Point, outputPath string) error {
	img, _, err := image.Decode(imgBuf)
	if err != nil {
		return fmt.Errorf("failed to decode image data: %w", err)
	}
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)
	drawArrow(rgba, from, to, color.RGBA{255, 0, 0, 255}, 5)
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()
	if err := png.Encode(outFile, rgba); err != nil {
		return fmt.Errorf("failed to encode and save image: %w", err)
	}
	return nil
}

// drawArrow draws an arrow from 'from' to 'to' on the image
func drawArrow(rgba *image.RGBA, from, to image.Point, color color.RGBA, lineWidth int) {
	bounds := rgba.Bounds()
	dx, dy := to.X-from.X, to.Y-from.Y
	steps := int(math.Sqrt(float64(dx*dx + dy*dy)))
	if steps == 0 {
		steps = 1
	}
	stepX, stepY := float64(dx)/float64(steps), float64(dy)/float64(steps)
	// main line
	for i := 0; i < steps; i++ {
		x := int(float64(from.X) + stepX*float64(i))
		y := int(float64(from.Y) + stepY*float64(i))
		for w := 0; w < lineWidth; w++ {
			offsetX, offsetY := 0, 0
			if math.Abs(stepX) > math.Abs(stepY) {
				offsetY = w - lineWidth/2
			} else {
				offsetX = w - lineWidth/2
			}
			drawX, drawY := x+offsetX, y+offsetY
			if drawX >= 0 && drawX < bounds.Max.X && drawY >= 0 && drawY < bounds.Max.Y {
				rgba.Set(drawX, drawY, color)
			}
		}
	}
	// arrow head
	arrowLength := float64(steps) * 0.15
	if arrowLength < 10 {
		arrowLength = 10
	} else if arrowLength > 30 {
		arrowLength = 30
	}
	head := calculateArrowHead(float64(from.X), float64(from.Y), float64(to.X), float64(to.Y), arrowLength)
	if head != nil {
		for _, point := range head[:2] {
			drawLineInImage(rgba, to.X, to.Y, int(point.X), int(point.Y), color, lineWidth, bounds)
		}
		for _, point := range head[1:] {
			drawLineInImage(rgba, to.X, to.Y, int(point.X), int(point.Y), color, lineWidth, bounds)
		}
	}
}

// calculateArrowHead calculates the endpoint and arrowhead coordinates
func calculateArrowHead(fromX, fromY, toX, toY float64, arrowLength float64) []struct{ X, Y float64 } {
	// calculate direction vector
	dx, dy := toX-fromX, toY-fromY
	// calculate distance
	length := math.Sqrt(dx*dx + dy*dy)
	if length < 1e-6 {
		return nil
	}

	// unit vector
	dx, dy = dx/length, dy/length

	// calculate orthogonal vector of arrow direction (counterclockwise 90 degrees)
	orthX, orthY := -dy, dx

	// calculate two wing points of arrow
	headWidth := arrowLength * 0.5
	backX, backY := toX-dx*arrowLength, toY-dy*arrowLength

	// two wing points of arrow
	leftWingX, leftWingY := backX+orthX*headWidth, backY+orthY*headWidth
	rightWingX, rightWingY := backX-orthX*headWidth, backY-orthY*headWidth

	return []struct{ X, Y float64 }{
		{leftWingX, leftWingY},
		{toX, toY},
		{rightWingX, rightWingY},
	}
}

// drawLineInImage draws a line on the image
func drawLineInImage(img *image.RGBA, x0, y0, x1, y1 int, lineColor color.RGBA, lineWidth int, bounds image.Rectangle) {
	// use Bresenham algorithm to draw line
	dx, dy := math.Abs(float64(x1-x0)), math.Abs(float64(y1-y0))
	sx, sy := 1, 1
	if x0 >= x1 {
		sx = -1
	}
	if y0 >= y1 {
		sy = -1
	}
	err := dx - dy

	for {
		// draw point (consider line width)
		for w := 0; w < lineWidth; w++ {
			offsetX, offsetY := 0, 0

			// decide offset direction based on line angle
			if dx > dy {
				// more horizontal line
				offsetY = w - lineWidth/2
			} else {
				// more vertical line
				offsetX = w - lineWidth/2
			}

			drawX, drawY := x0+offsetX, y0+offsetY
			if drawX >= 0 && drawX < bounds.Max.X && drawY >= 0 && drawY < bounds.Max.Y {
				img.Set(drawX, drawY, lineColor)
			}
		}

		// end of line
		if x0 == x1 && y0 == y1 {
			break
		}

		// calculate next point
		e2 := 2 * err
		if e2 > -dy {
			err = err - dy
			x0 = x0 + sx
		}
		if e2 < dx {
			err = err + dx
			y0 = y0 + sy
		}
	}
}
