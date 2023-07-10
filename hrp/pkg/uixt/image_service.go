package uixt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
	"github.com/httprunner/httprunner/v4/hrp/internal/code"
	"github.com/httprunner/httprunner/v4/hrp/internal/env"
)

var client = &http.Client{
	Timeout: time.Second * 10,
}

type ImageResult struct {
	imagePath string
	URL       string     `json:"url"`       // image uploaded url
	OCRResult OCRResults `json:"ocrResult"` // OCR texts
	LiveType  string     `json:"liveType"`  // 直播间类型
}

type APIResponseImage struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Result  ImageResult `json:"result"`
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

// veDEMImageService implements IImageService interface
// actions:
// 	ocr - get ocr texts
// 	upload - get image uploaded url
// 	liveType - get live type
// 	popup - get popup windows
// 	close - get close popup
type veDEMImageService struct {
	actions []string
}

func (s *veDEMImageService) GetImage(imageBuf *bytes.Buffer) (
	imageResult ImageResult, err error,
) {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	for _, action := range s.actions {
		bodyWriter.WriteField("actions", action)
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

func getRectangleCenterPoint(rect image.Rectangle) (point PointF) {
	x, y := float64(rect.Min.X), float64(rect.Min.Y)
	width, height := float64(rect.Dx()), float64(rect.Dy())
	point = PointF{
		X: x + width*0.5,
		Y: y + height*0.5,
	}
	return point
}
