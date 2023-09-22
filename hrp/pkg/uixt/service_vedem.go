package uixt

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
	"github.com/httprunner/httprunner/v4/hrp/internal/code"
	"github.com/httprunner/httprunner/v4/hrp/internal/env"
	"github.com/httprunner/httprunner/v4/hrp/internal/json"
)

var client = &http.Client{
	Timeout: time.Second * 10,
}

type OCRResult struct {
	Text   string   `json:"text"`
	Points []PointF `json:"points"`
}

type OCRResults []OCRResult

func (o OCRResults) ToOCRTexts() (ocrTexts OCRTexts) {
	for _, ocrResult := range o {
		rect := image.Rectangle{
			// ocrResult.Points 顺序：左上 -> 右上 -> 右下 -> 左下
			Min: image.Point{
				X: int(ocrResult.Points[0].X),
				Y: int(ocrResult.Points[0].Y),
			},
			Max: image.Point{
				X: int(ocrResult.Points[2].X),
				Y: int(ocrResult.Points[2].Y),
			},
		}
		ocrText := OCRText{
			Text: ocrResult.Text,
			Rect: rect,
		}
		ocrTexts = append(ocrTexts, ocrText)
	}
	return
}

type ImageResult struct {
	URL       string     `json:"url,omitempty"`       // image uploaded url
	OCRResult OCRResults `json:"ocrResult,omitempty"` // OCR texts
	// NoLive（非直播间）
	// Shop（电商）
	// LifeService（生活服务）
	// Show（秀场）
	// Game（游戏）
	// People（多人）
	// PK（PK）
	// Media（媒体）
	// Chat（语音）
	// Event（赛事）
	LiveType          string             `json:"liveType,omitempty"`    // 直播间类型
	UIResult          UIResultMap        `json:"uiResult,omitempty"`    // 图标检测
	ClosePopupsResult *ClosePopupsResult `json:"closeResult,omitempty"` // 弹窗按钮检测
}

type APIResponseImage struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Result  ImageResult `json:"result"`
}

type OCRText struct {
	Text string
	Rect image.Rectangle
}

func (t OCRText) Center() PointF {
	return getRectangleCenterPoint(t.Rect)
}

type OCRTexts []OCRText

func (t OCRTexts) texts() (texts []string) {
	for _, text := range t {
		texts = append(texts, text.Text)
	}
	return texts
}

func (t OCRTexts) FilterScope(scope AbsScope) (results OCRTexts) {
	for _, ocrText := range t {
		rect := ocrText.Rect

		// check if text in scope
		if len(scope) == 4 {
			if rect.Min.X < scope[0] ||
				rect.Min.Y < scope[1] ||
				rect.Max.X > scope[2] ||
				rect.Max.Y > scope[3] {
				// not in scope
				continue
			}
		}

		results = append(results, ocrText)
	}
	return
}

func (t OCRTexts) FindText(text string, options ...ActionOption) (result OCRText, err error) {
	actionOptions := NewActionOptions(options...)

	var results []OCRText
	for _, ocrText := range t.FilterScope(actionOptions.AbsScope) {
		if actionOptions.Regex {
			// regex on, check if match regex
			if !regexp.MustCompile(text).MatchString(ocrText.Text) {
				continue
			}
		} else {
			// regex off, check if match exactly
			if ocrText.Text != text {
				continue
			}
		}

		results = append(results, ocrText)

		// return the first one matched exactly when index not specified
		if ocrText.Text == text && actionOptions.Index == 0 {
			return ocrText, nil
		}
	}

	if len(results) == 0 {
		return OCRText{}, errors.Wrap(code.CVResultNotFoundError,
			fmt.Sprintf("text %s not found in %v", text, t.texts()))
	}

	// get index
	idx := actionOptions.Index
	if idx < 0 {
		idx = len(results) + idx
	}

	// index out of range
	if idx >= len(results) || idx < 0 {
		return OCRText{}, errors.Wrap(code.CVResultNotFoundError,
			fmt.Sprintf("text %s found %d, index %d out of range", text, len(results), idx))
	}

	return results[idx], nil
}

func (t OCRTexts) FindTexts(texts []string, options ...ActionOption) (results OCRTexts, err error) {
	actionOptions := NewActionOptions(options...)
	for _, text := range texts {
		ocrText, err := t.FindText(text, options...)
		if err != nil {
			continue
		}
		results = append(results, ocrText)
	}

	if len(results) == len(texts) {
		return results, nil
	}

	if actionOptions.MatchOne && len(results) > 0 {
		return results, nil
	}

	return nil, errors.Wrap(code.CVResultNotFoundError,
		fmt.Sprintf("texts %s not found in %v", texts, t.texts()))
}

func newVEDEMImageService() (*veDEMImageService, error) {
	if err := checkEnv(); err != nil {
		return nil, err
	}
	return &veDEMImageService{}, nil
}

// veDEMImageService implements IImageService interface
// actions:
//
//	ocr - get ocr texts
//	upload - get image uploaded url
//	liveType - get live type
//	popup - get popup windows
//	close - get close popup
//	ui - get ui position by type(s)
type veDEMImageService struct{}

func (s *veDEMImageService) GetImage(imageBuf *bytes.Buffer, options ...ActionOption) (imageResult *ImageResult, err error) {
	actionOptions := NewActionOptions(options...)
	screenshotActions := actionOptions.screenshotActions()
	if len(screenshotActions) == 0 {
		// skip
		return nil, nil
	}

	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	for _, action := range screenshotActions {
		bodyWriter.WriteField("actions", action)
	}
	for _, uiType := range actionOptions.ScreenShotWithUITypes {
		bodyWriter.WriteField("uiTypes", uiType)
	}

	bodyWriter.WriteField("ocrCluster", "highPrecision")

	formWriter, err := bodyWriter.CreateFormFile("image", "screenshot.png")
	if err != nil {
		err = errors.Wrap(code.CVRequestError,
			fmt.Sprintf("create form file error: %v", err))
		return
	}

	size, err := formWriter.Write(imageBuf.Bytes())
	if err != nil {
		err = errors.Wrap(code.CVRequestError,
			fmt.Sprintf("write form error: %v", err))
		return
	}

	err = bodyWriter.Close()
	if err != nil {
		err = errors.Wrap(code.CVRequestError,
			fmt.Sprintf("close body writer error: %v", err))
		return
	}
	var req *http.Request
	var resp *http.Response
	// retry 3 times
	for i := 1; i <= 3; i++ {
		copiedBodyBuf := &bytes.Buffer{}
		if _, err := copiedBodyBuf.Write(bodyBuf.Bytes()); err != nil {
			log.Error().Err(err).Msg("copy screenshot buffer failed")
			continue
		}

		req, err = http.NewRequest("POST", env.VEDEM_IMAGE_URL, copiedBodyBuf)
		if err != nil {
			err = errors.Wrap(code.CVRequestError,
				fmt.Sprintf("construct request error: %v", err))
			return
		}

		// ppe env
		// req.Header.Add("x-tt-env", "ppe_vedem_algorithm")
		// req.Header.Add("x-use-ppe", "1")

		signToken := "UNSIGNED-PAYLOAD"
		token := builtin.Sign("auth-v2", env.VEDEM_IMAGE_AK, env.VEDEM_IMAGE_SK, []byte(signToken))

		req.Header.Add("Agw-Auth", token)
		req.Header.Add("Agw-Auth-Content", signToken)
		req.Header.Add("Content-Type", bodyWriter.FormDataContentType())

		start := time.Now()
		resp, err = client.Do(req)
		elapsed := time.Since(start)
		if err != nil {
			log.Error().Err(err).
				Int("imageBufSize", size).
				Msgf("request veDEM OCR service error, retry %d", i)
			continue
		}

		logID := getLogID(resp.Header)
		statusCode := resp.StatusCode
		if statusCode != http.StatusOK {
			log.Error().
				Str("X-TT-LOGID", logID).
				Int("imageBufSize", size).
				Int("statusCode", statusCode).
				Msgf("request veDEM OCR service failed, retry %d", i)
			time.Sleep(1 * time.Second)
			continue
		}

		log.Debug().
			Str("X-TT-LOGID", logID).
			Int("image_bytes", size).
			Int64("elapsed(ms)", elapsed.Milliseconds()).
			Msg("request OCR service success")
		break
	}
	if resp == nil {
		err = code.CVServiceConnectionError
		return
	}

	defer resp.Body.Close()

	results, err := io.ReadAll(resp.Body)
	if err != nil {
		err = errors.Wrap(code.CVResponseError,
			fmt.Sprintf("read response body error: %v", err))
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = errors.Wrap(code.CVResponseError,
			fmt.Sprintf("unexpected response status code: %d, results: %v",
				resp.StatusCode, string(results)))
		return
	}

	var imageResponse APIResponseImage
	err = json.Unmarshal(results, &imageResponse)
	if err != nil {
		log.Error().Err(err).
			Str("response", string(results)).
			Msg("json unmarshal veDEM image response body failed")
		err = errors.Wrap(code.CVResponseError,
			"json unmarshal veDEM image response body error")
		return
	}

	if imageResponse.Code != 0 {
		log.Error().
			Int("code", imageResponse.Code).
			Str("message", imageResponse.Message).
			Msg("request veDEM OCR service failed")
	}

	imageResult = &imageResponse.Result
	log.Debug().Interface("imageResult", imageResult).Msg("get image data by veDEM")
	return imageResult, nil
}

func checkEnv() error {
	if env.VEDEM_IMAGE_URL == "" {
		return errors.Wrap(code.CVEnvMissedError, "VEDEM_IMAGE_URL missed")
	}
	log.Info().Str("VEDEM_IMAGE_URL", env.VEDEM_IMAGE_URL).Msg("get env")
	if env.VEDEM_IMAGE_AK == "" {
		return errors.Wrap(code.CVEnvMissedError, "VEDEM_IMAGE_AK missed")
	}
	if env.VEDEM_IMAGE_SK == "" {
		return errors.Wrap(code.CVEnvMissedError, "VEDEM_IMAGE_SK missed")
	}
	return nil
}

func getLogID(header http.Header) string {
	if len(header) == 0 {
		return ""
	}

	logID, ok := header["X-Tt-Logid"]
	if !ok || len(logID) == 0 {
		return ""
	}
	return logID[0]
}

type IImageService interface {
	// GetImage returns image result including ocr texts, uploaded image url, etc
	GetImage(imageBuf *bytes.Buffer, options ...ActionOption) (imageResult *ImageResult, err error)
}

// GetScreenResult takes a screenshot, returns the image recognization result
func (dExt *DriverExt) GetScreenResult(options ...ActionOption) (screenResult *ScreenResult, err error) {
	startTime := time.Now()
	fileName := builtin.GenNameWithTimestamp("%d_screenshot")
	actionOptions := NewActionOptions(options...)
	screenshotActions := actionOptions.screenshotActions()
	if len(screenshotActions) != 0 {
		fileName = builtin.GenNameWithTimestamp("%d_" + strings.Join(screenshotActions, "_"))
	}
	bufSource, imagePath, err := dExt.takeScreenShot(fileName)
	if err != nil {
		return
	}
	screenshotTakeElapsed := time.Since(startTime).Milliseconds()

	screenResult = &ScreenResult{
		bufSource:             bufSource,
		imagePath:             imagePath,
		Tags:                  nil,
		Resolution:            dExt.windowSize,
		ScreenshotTakeElapsed: screenshotTakeElapsed,
	}

	imageResult, err := dExt.ImageService.GetImage(bufSource, options...)
	if err != nil {
		log.Error().Err(err).Msg("GetImage from ImageService failed")
		return nil, err
	}
	if imageResult != nil {
		screenResult.imageResult = imageResult
		screenResult.ScreenshotCVElapsed = time.Since(startTime).Milliseconds() - screenshotTakeElapsed
		screenResult.Texts = imageResult.OCRResult.ToOCRTexts()
		screenResult.UploadedURL = imageResult.URL
		screenResult.Icons = imageResult.UIResult

		if actionOptions.ScreenShotWithClosePopups {
			popup := &PopupInfo{
				ClosePopupsResult: imageResult.ClosePopupsResult,
			}

			closeAreas, _ := imageResult.UIResult.FilterUIResults([]string{"close"})
			for _, closeArea := range closeAreas {
				popup.ClosePoints = append(popup.ClosePoints, closeArea.Center())
			}

			screenResult.Popup = popup
		}
	}

	dExt.cacheStepData.screenResults[time.Now().String()] = screenResult

	log.Debug().
		Str("imagePath", imagePath).
		Str("imageUrl", screenResult.UploadedURL).
		Int64("screenshot_take_elapsed(ms)", screenResult.ScreenshotTakeElapsed).
		Int64("screenshot_cv_elapsed(ms)", screenResult.ScreenshotCVElapsed).
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

func getRectangleCenterPoint(rect image.Rectangle) (point PointF) {
	x, y := float64(rect.Min.X), float64(rect.Min.Y)
	width, height := float64(rect.Dx()), float64(rect.Dy())
	point = PointF{
		X: x + width*0.5,
		Y: y + height*0.5,
	}
	return point
}

type Box struct {
	Point  PointF  `json:"point"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

func (box Box) IsEmpty() bool {
	return builtin.IsZeroFloat64(box.Width) && builtin.IsZeroFloat64(box.Height)
}

func (box Box) IsIdentical(box2 Box) bool {
	// set the coordinate precision to 1 pixel
	return box.Point.IsIdentical(box2.Point) &&
		math.Abs(box.Width-box2.Width) < 1 &&
		math.Abs(box.Height-box2.Height) < 1
}

func (box Box) Center() PointF {
	return PointF{
		X: box.Point.X + box.Width*0.5,
		Y: box.Point.Y + box.Height*0.5,
	}
}

type UIResult struct {
	Box
}

type UIResults []UIResult

func (u UIResults) FilterScope(scope AbsScope) (results UIResults) {
	for _, uiResult := range u {
		rect := image.Rectangle{
			Min: image.Point{
				X: int(uiResult.Point.X),
				Y: int(uiResult.Point.Y),
			},
			Max: image.Point{
				X: int(uiResult.Point.X + uiResult.Width),
				Y: int(uiResult.Point.Y + uiResult.Height),
			},
		}

		// check if ui result in scope
		if len(scope) == 4 {
			if rect.Min.X < scope[0] ||
				rect.Min.Y < scope[1] ||
				rect.Max.X > scope[2] ||
				rect.Max.Y > scope[3] {
				// not in scope
				continue
			}
		}
		results = append(results, uiResult)
	}
	return
}

type UIResultMap map[string]UIResults

// FilterUIResults filters ui icons, the former the uiTypes, the higher the priority
func (u UIResultMap) FilterUIResults(uiTypes []string) (uiResults UIResults, err error) {
	var ok bool
	for _, uiType := range uiTypes {
		uiResults, ok = u[uiType]
		if ok && len(uiResults) != 0 {
			return
		}
	}
	err = errors.Errorf("UI types %v not detected", uiTypes)
	return
}

func (u UIResults) GetUIResult(options ...ActionOption) (UIResult, error) {
	actionOptions := NewActionOptions(options...)

	uiResults := u.FilterScope(actionOptions.AbsScope)
	if len(uiResults) == 0 {
		return UIResult{}, errors.Wrap(code.CVResultNotFoundError,
			"ui types not found in scope")
	}
	// get index
	idx := actionOptions.Index
	if idx < 0 {
		idx = len(uiResults) + idx
	}

	// index out of range
	if idx >= len(uiResults) || idx < 0 {
		return UIResult{}, errors.Wrap(code.CVResultNotFoundError,
			fmt.Sprintf("ui types index %d out of range", idx))
	}
	return uiResults[idx], nil
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
