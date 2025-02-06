package uixt

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/internal/json"
	"github.com/httprunner/httprunner/v5/pkg/uixt/options"
)

var client = &http.Client{
	Timeout: time.Second * 10,
}

type APIResponseImage struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Result  ImageResult `json:"result"`
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

func (s *veDEMImageService) GetImage(imageBuf *bytes.Buffer, opts ...options.ActionOption) (imageResult *ImageResult, err error) {
	actionOptions := options.NewActionOptions(opts...)
	screenshotActions := actionOptions.ScreenshotActions()
	if len(screenshotActions) == 0 {
		// skip
		return nil, nil
	}

	start := time.Now()
	defer func() {
		elapsed := time.Since(start).Milliseconds()
		var logger *zerolog.Event
		if err != nil {
			logger = log.Error().Err(err)
		} else {
			logger = log.Debug()
			if imageResult.URL != "" {
				logger = logger.Str("url", imageResult.URL)
			}
			if imageResult.UIResult != nil {
				logger = logger.Interface("uiResult", imageResult.UIResult)
			}
			if imageResult.ClosePopupsResult != nil {
				if imageResult.ClosePopupsResult.IsEmpty() {
					// set nil to reduce unnecessary summary info
					imageResult.ClosePopupsResult = nil
				} else {
					logger = logger.Interface("closePopupsResult", imageResult.ClosePopupsResult)
				}
			}
		}
		logger = logger.Int64("elapsed(ms)", elapsed)
		logger.Msg("get image data by veDEM")
	}()

	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	for _, action := range screenshotActions {
		bodyWriter.WriteField("actions", action)
	}
	for _, uiType := range actionOptions.ScreenShotWithUITypes {
		bodyWriter.WriteField("uiTypes", uiType)
	}

	// 使用高精度集群
	bodyWriter.WriteField("ocrCluster", "highPrecision")

	if actionOptions.ScreenShotWithOCRCluster != "" {
		bodyWriter.WriteField("ocrCluster", actionOptions.ScreenShotWithOCRCluster)
	}

	if actionOptions.Timeout > 0 {
		bodyWriter.WriteField("timeout", fmt.Sprintf("%v", actionOptions.Timeout))
	} else {
		bodyWriter.WriteField("timeout", fmt.Sprintf("%v", 10))
	}

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

		req, err = http.NewRequest("POST", os.Getenv("VEDEM_IMAGE_URL"), copiedBodyBuf)
		if err != nil {
			err = errors.Wrap(code.CVRequestError,
				fmt.Sprintf("construct request error: %v", err))
			return
		}

		// ppe env
		// req.Header.Add("x-tt-env", "ppe_vedem_algorithm")
		// req.Header.Add("x-use-ppe", "1")

		signToken := "UNSIGNED-PAYLOAD"
		token := builtin.Sign("auth-v2", os.Getenv("VEDEM_IMAGE_AK"), os.Getenv("VEDEM_IMAGE_SK"), []byte(signToken))

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
		err = errors.Wrap(code.CVResponseError,
			fmt.Sprintf("json unmarshal veDEM image response body error, response=%s", string(results)))
		return
	}

	if imageResponse.Code != 0 {
		err = errors.Wrap(code.CVResponseError,
			fmt.Sprintf("unexpected response data code: %d, message: %s",
				imageResponse.Code, imageResponse.Message))
		return
	}

	imageResult = &imageResponse.Result
	return imageResult, nil
}

func checkEnv() error {
	vedemImageURL := os.Getenv("VEDEM_IMAGE_URL")
	if vedemImageURL == "" {
		return errors.Wrap(code.CVEnvMissedError, "VEDEM_IMAGE_URL missed")
	}
	log.Info().Str("VEDEM_IMAGE_URL", vedemImageURL).Msg("get env")
	if os.Getenv("VEDEM_IMAGE_AK") == "" {
		return errors.Wrap(code.CVEnvMissedError, "VEDEM_IMAGE_AK missed")
	}
	if os.Getenv("VEDEM_IMAGE_SK") == "" {
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
