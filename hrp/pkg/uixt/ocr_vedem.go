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

func (o OCRResult) ToOCRText() OCRText {
	rect := image.Rectangle{
		// ocrResult.Points 顺序：左上 -> 右上 -> 右下 -> 左下
		Min: image.Point{
			X: int(o.Points[0].X),
			Y: int(o.Points[0].Y),
		},
		Max: image.Point{
			X: int(o.Points[2].X),
			Y: int(o.Points[2].Y),
		},
	}

	return OCRText{
		Text: o.Text,
		Rect: rect,
	}
}

type ImageResult struct {
	URL       string      `json:"url"`       // image uploaded url
	OCRResult []OCRResult `json:"ocrResult"` // OCR texts
	LiveType  string      `json:"liveType"`  // 直播间类型
}

type ImageResponse struct {
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

func (t OCRTexts) FindText(text string, options ...ActionOption) (
	result OCRText, err error) {

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

func (t OCRTexts) FindTexts(texts []string, options ...ActionOption) (
	results OCRTexts, err error) {
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

func newVEDEMImageService() (*veDEMImageService, error) {
	if err := checkEnv(); err != nil {
		return nil, err
	}
	return &veDEMImageService{}, nil
}

// veDEMImageService implements IImageService interface
type veDEMImageService struct{}

var actions = []string{
	"ocr",      // get ocr texts
	"upload",   // get image uploaded url
	"liveType", // get live type
	// "popup",
	// "close",
}

func (s *veDEMImageService) GetImage(imageBuf *bytes.Buffer) (
	imageResult ImageResult, err error) {

	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	for _, action := range actions {
		bodyWriter.WriteField("actions", action)
	}

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

	req, err := http.NewRequest("POST", env.VEDEM_IMAGE_URL, bodyBuf)
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

	// ppe
	// req.Header.Add("x-use-ppe", "1")
	// req.Header.Add("x-tt-env", "ppe_vedem_algorithm")

	var resp *http.Response
	// retry 3 times
	for i := 1; i <= 3; i++ {
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
				Float64("elapsed_seconds", elapsed.Seconds()).
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

	var imageResponse ImageResponse
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

// GetScreenTextsByOCR takes a screenshot, returns the image path and OCR texts.
func (dExt *DriverExt) GetScreenTextsByOCR() (imagePath string, ocrTexts OCRTexts, err error) {
	var bufSource *bytes.Buffer
	if bufSource, imagePath, err = dExt.TakeScreenShot(
		builtin.GenNameWithTimestamp("%d_ocr")); err != nil {
		return
	}

	imageResult, err := dExt.ImageService.GetImage(bufSource)
	if err != nil {
		log.Error().Err(err).Msg("GetScreenTextsByOCR failed")
		return
	}

	for _, ocrResult := range imageResult.OCRResult {
		ocrTexts = append(ocrTexts, ocrResult.ToOCRText())
	}

	imageUrl := imageResult.URL
	if imageUrl != "" {
		dExt.cacheStepData.screenShotsUrls[imagePath] = imageUrl
		log.Debug().Str("imagePath", imagePath).Str("imageUrl", imageUrl).Msg("log screenshot")
	}

	dExt.cacheStepData.OcrResults[imagePath] = &OcrResult{
		Texts: ocrTexts,
	}

	return imagePath, ocrTexts, nil
}

func (dExt *DriverExt) FindScreenTextByOCR(text string, options ...ActionOption) (point PointF, err error) {
	_, ocrTexts, err := dExt.GetScreenTextsByOCR()
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
		Interface("point", point).Msgf("FindScreenTextByOCR success")
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
