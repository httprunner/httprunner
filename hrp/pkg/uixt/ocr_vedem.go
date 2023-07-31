package uixt

import (
	"bytes"
	"fmt"
	"image"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"regexp"
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
	imagePath string
	URL       string      `json:"url"`       // image uploaded url
	OCRResult OCRResults  `json:"ocrResult"` // OCR texts
	LiveType  string      `json:"liveType"`  // 直播间类型
	UIResult  UIResultMap `json:"uiResult"`  // 图标检测
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
	}

	if len(results) == 0 {
		return OCRText{}, errors.Wrap(code.OCRTextNotFoundError,
			fmt.Sprintf("text %s not found in %v", text, t.texts()))
	}

	// get index
	idx := actionOptions.Index
	if idx < 0 {
		idx = len(results) + idx
	}

	// index out of range
	if idx >= len(results) || idx < 0 {
		return OCRText{}, errors.Wrap(code.OCRTextNotFoundError,
			fmt.Sprintf("text %s found %d, index %d out of range", text, len(results), idx))
	}

	return results[idx], nil
}

func (t OCRTexts) FindTexts(texts []string, options ...ActionOption) (results OCRTexts, err error) {
	for _, text := range texts {
		ocrText, err := t.FindText(text, options...)
		if err != nil {
			continue
		}
		results = append(results, ocrText)
	}

	if len(results) != len(texts) {
		return nil, errors.Wrap(code.OCRTextNotFoundError,
			fmt.Sprintf("texts %s not found in %v", texts, t.texts()))
	}
	return results, nil
}

func newVEDEMImageService(actions ...string) (*veDEMImageService, error) {
	if err := checkEnv(); err != nil {
		return nil, err
	}
	if len(actions) == 0 {
		actions = []string{"ocr"}
	}
	return &veDEMImageService{
		actions: actions,
	}, nil
}

func (v *veDEMImageService) WithUITypes(uiTypes ...string) *veDEMImageService {
	v.uiTypes = uiTypes
	return v
}

// veDEMImageService implements IImageService interface
// actions:
//
//	ocr - get ocr texts
//	upload - get image uploaded url
//	liveType - get live type
//	popup - get popup windows
//	close - get close popup
type veDEMImageService struct {
	actions []string
	uiTypes []string
}

func (s *veDEMImageService) GetImage(imageBuf *bytes.Buffer) (imageResult ImageResult, err error) {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	for _, action := range s.actions {
		bodyWriter.WriteField("actions", action)
	}
	for _, uiType := range s.uiTypes {
		bodyWriter.WriteField("uiTypes", uiType)
	}
	bodyWriter.WriteField("ocrCluster", "highPrecision")

	formWriter, err := bodyWriter.CreateFormFile("image", "screenshot.png")
	if err != nil {
		err = errors.Wrap(code.OCRRequestError,
			fmt.Sprintf("create form file error: %v", err))
		return
	}

	size, err := formWriter.Write(imageBuf.Bytes())
	if err != nil {
		err = errors.Wrap(code.OCRRequestError,
			fmt.Sprintf("write form error: %v", err))
		return
	}

	err = bodyWriter.Close()
	if err != nil {
		err = errors.Wrap(code.OCRRequestError,
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
			err = errors.Wrap(code.OCRRequestError,
				fmt.Sprintf("construct request error: %v", err))
			return
		}

		signToken := "UNSIGNED-PAYLOAD"
		token := builtin.Sign("auth-v2", env.VEDEM_IMAGE_AK, env.VEDEM_IMAGE_SK, []byte(signToken))

		req.Header.Add("Agw-Auth", token)
		req.Header.Add("Agw-Auth-Content", signToken)
		req.Header.Add("Content-Type", bodyWriter.FormDataContentType())

		start := time.Now()
		resp, err = client.Do(req)
		elapsed := time.Since(start)
		var logID string
		if resp != nil {
			logID = getLogID(resp.Header)
		}
		if err == nil && resp.StatusCode == http.StatusOK {
			log.Debug().
				Str("X-TT-LOGID", logID).
				Int("image_bytes", size).
				Float64("elapsed(s)", elapsed.Seconds()).
				Msg("request OCR service success")
			break
		}
		log.Error().Err(err).
			Str("X-TT-LOGID", logID).
			Int("imageBufSize", size).
			Msgf("request veDEM OCR service failed, retry %d", i)
		time.Sleep(1 * time.Second)
	}
	if resp == nil {
		err = code.OCRServiceConnectionError
		return
	}

	defer resp.Body.Close()

	results, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = errors.Wrap(code.OCRResponseError,
			fmt.Sprintf("read response body error: %v", err))
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = errors.Wrap(code.OCRResponseError,
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
		err = errors.Wrap(code.OCRResponseError,
			"json unmarshal veDEM image response body error")
		return
	}

	if imageResponse.Code != 0 {
		log.Error().
			Int("code", imageResponse.Code).
			Str("message", imageResponse.Message).
			Msg("request veDEM OCR service failed")
	}

	imageResult = imageResponse.Result
	log.Debug().Interface("imageResult", imageResult).Msg("get image data by veDEM")
	return imageResult, nil
}

func checkEnv() error {
	if env.VEDEM_IMAGE_URL == "" {
		return errors.Wrap(code.OCREnvMissedError, "VEDEM_IMAGE_URL missed")
	}
	log.Info().Str("VEDEM_IMAGE_URL", env.VEDEM_IMAGE_URL).Msg("get env")
	if env.VEDEM_IMAGE_AK == "" {
		return errors.Wrap(code.OCREnvMissedError, "VEDEM_IMAGE_AK missed")
	}
	if env.VEDEM_IMAGE_SK == "" {
		return errors.Wrap(code.OCREnvMissedError, "VEDEM_IMAGE_SK missed")
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
	GetImage(imageBuf *bytes.Buffer) (imageResult ImageResult, err error)
}

// GetScreenResult takes a screenshot, returns the image recognization result
func (dExt *DriverExt) GetScreenResult() (screenResult *ScreenResult, err error) {
	var bufSource *bytes.Buffer
	var imagePath string
	if bufSource, imagePath, err = dExt.takeScreenShot(
		builtin.GenNameWithTimestamp("%d_ocr")); err != nil {
		return
	}

	imageResult, err := dExt.ImageService.GetImage(bufSource)
	if err != nil {
		log.Error().Err(err).Msg("GetImage from ImageService failed")
		return
	}
	imageResult.imagePath = imagePath

	imageUrl := imageResult.URL
	if imageUrl != "" {
		dExt.cacheStepData.screenShotsUrls[imagePath] = imageUrl
		log.Debug().Str("imagePath", imagePath).Str("imageUrl", imageUrl).Msg("log screenshot")
	}

	screenResult = &ScreenResult{
		Texts:      imageResult.OCRResult.ToOCRTexts(),
		Tags:       nil,
		Popularity: Popularity{},
	}
	if imageResult.LiveType != "" {
		screenResult.Tags = []string{imageResult.LiveType}
	}
	dExt.cacheStepData.screenResults[imagePath] = screenResult

	return screenResult, nil
}

func (dExt *DriverExt) GetScreenTexts() (ocrTexts OCRTexts, err error) {
	screenResult, err := dExt.GetScreenResult()
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

func getCenterPoint(point PointF, width, height float64) PointF {
	return PointF{
		X: point.X + width*0.5,
		Y: point.Y + height*0.5,
	}
}

type UIResult struct {
	Point  PointF  `json:"point"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

func (u UIResult) Center() PointF {
	return getCenterPoint(u.Point, u.Width, u.Height)
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
		return UIResult{}, errors.Wrap(code.OCRTextNotFoundError,
			"ui types not found in scope")
	}
	// get index
	idx := actionOptions.Index
	if idx < 0 {
		idx = len(uiResults) + idx
	}

	// index out of range
	if idx >= len(uiResults) || idx < 0 {
		return UIResult{}, errors.Wrap(code.OCRTextNotFoundError,
			fmt.Sprintf("ui types index %d out of range", idx))
	}
	return uiResults[idx], nil
}

// newVEDEMUIService return image service for
func newVEDEMUIService(uiTypes []string) (*veDEMImageService, error) {
	vedemUIService, err := newVEDEMImageService("ui")
	if err != nil {
		return nil, err
	}
	return vedemUIService.WithUITypes(uiTypes...), nil
}

func (dExt *DriverExt) GetUIResultMap(uiTypes []string) (uiResultMap UIResultMap, err error) {
	var bufSource *bytes.Buffer
	var imagePath string
	if bufSource, imagePath, err = dExt.takeScreenShot(
		builtin.GenNameWithTimestamp("%d_ocr")); err != nil {
		return
	}

	vedemUIService, err := newVEDEMUIService(uiTypes)
	if err != nil {
		return
	}
	imageResult, err := vedemUIService.GetImage(bufSource)
	if err != nil {
		log.Error().Err(err).Msg("GetImage from ImageService failed")
		return
	}

	imageUrl := imageResult.URL
	if imageUrl != "" {
		dExt.cacheStepData.screenShotsUrls[imagePath] = imageUrl
		log.Debug().Str("imagePath", imagePath).Str("imageUrl", imageUrl).Msg("log screenshot")
	}
	uiResultMap = imageResult.UIResult
	return
}

func (dExt *DriverExt) FindUIResult(uiTypes []string, options ...ActionOption) (point PointF, err error) {
	uiResultMap, err := dExt.GetUIResultMap(uiTypes)
	if err != nil {
		return
	}
	uiResults, err := uiResultMap.FilterUIResults(uiTypes)
	if err != nil {
		return
	}
	uiResult, err := uiResults.GetUIResult(dExt.ParseActionOptions(options...)...)
	point = uiResult.Center()

	log.Info().Interface("text", uiTypes).
		Interface("point", point).Msg("FindUIResult success")
	return
}
