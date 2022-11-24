package uixt

import (
	"bytes"
	"fmt"
	"image"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
	"github.com/httprunner/httprunner/v4/hrp/internal/code"
	"github.com/httprunner/httprunner/v4/hrp/internal/env"
	"github.com/httprunner/httprunner/v4/hrp/internal/json"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type SDResult struct {
	Image  string   `json:"image"`
	Points []PointF `json:"points"`
}

type SDResponse struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	Result  []SDResult `json:"result"`
}

type veDEMSDService struct{}

func newVEDEMSDService() (*veDEMSDService, error) {
	if err := checkSDEnv(); err != nil {
		return nil, err
	}
	return &veDEMSDService{}, nil
}

func checkSDEnv() error {
	if env.VEDEM_SD_URL == "" {
		return errors.Wrap(code.CVEnvMissedError, "VEDEM_SD_URL missed")
	}
	if env.VEDEM_SD_AK == "" {
		return errors.Wrap(code.CVEnvMissedError, "VEDEM_SD_AK missed")
	}
	if env.VEDEM_SD_SK == "" {
		return errors.Wrap(code.CVEnvMissedError, "VEDEM_SD_SK missed")
	}
	return nil
}

func (s *veDEMSDService) getSDResult(searchImage []byte, sourceImage []byte) ([]SDResult, error) {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	bodyWriter.WriteField("withDet", "true")
	// bodyWriter.WriteField("timestampOnly", "true")

	formWriter, err := bodyWriter.CreateFormFile("searchImage", "searchImage.png")
	if err != nil {
		return nil, errors.Wrap(code.CVRequestError,
			fmt.Sprintf("create form file error: %v", err))
	}
	size, err := formWriter.Write(searchImage)
	if err != nil {
		return nil, errors.Wrap(code.CVRequestError,
			fmt.Sprintf("write form error: %v", err))
	}

	formWriter, err = bodyWriter.CreateFormFile("sourceImage", "sourceImage.png")
	if err != nil {
		return nil, errors.Wrap(code.CVRequestError,
			fmt.Sprintf("create form file error: %v", err))
	}
	_, err = formWriter.Write(sourceImage)
	if err != nil {
		return nil, errors.Wrap(code.CVRequestError,
			fmt.Sprintf("write form error: %v", err))
	}

	err = bodyWriter.Close()
	if err != nil {
		return nil, errors.Wrap(code.CVRequestError,
			fmt.Sprintf("close body writer error: %v", err))
	}

	req, err := http.NewRequest("POST", env.VEDEM_SD_URL, bodyBuf)
	if err != nil {
		return nil, errors.Wrap(code.CVRequestError,
			fmt.Sprintf("construct request error: %v", err))
	}

	token := builtin.Sign("auth-v2", env.VEDEM_SD_AK, env.VEDEM_SD_SK, bodyBuf.Bytes())
	req.Header.Add("Agw-Auth", token)
	req.Header.Add("Content-Type", bodyWriter.FormDataContentType())

	var resp *http.Response
	// retry 3 times
	for i := 1; i <= 3; i++ {
		resp, err = client.Do(req)
		if err == nil {
			break
		}

		var logID string
		if resp != nil {
			logID = getLogID(resp.Header)
		}
		log.Error().Err(err).
			Str("logID", logID).
			Int("imageBufSize", size).
			Msgf("request CV service failed, retry %d", i)
		time.Sleep(1 * time.Second)
	}
	if resp == nil {
		return nil, code.CVServiceConnectionError
	}

	defer resp.Body.Close()

	results, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(code.CVResponseError,
			fmt.Sprintf("read response body error: %v", err))
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Wrap(code.CVResponseError,
			fmt.Sprintf("unexpected response status code: %d, results: %v",
				resp.StatusCode, string(results)))
	}

	var cvResult SDResponse
	err = json.Unmarshal(results, &cvResult)
	if err != nil {
		return nil, errors.Wrap(code.CVResponseError,
			fmt.Sprintf("json unmarshal response body error: %v", err))
	}

	return cvResult.Result, nil
}

func (s *veDEMSDService) FindImage(byteSearch []byte, byteSource []byte, options ...DataOption) (rect image.Rectangle, err error) {
	data := NewData(map[string]interface{}{}, options...)

	cvResults, err := s.getSDResult(byteSearch, byteSource)
	if err != nil {
		log.Error().Err(err).Msg("getSDResult failed")
		return
	}

	var rects []image.Rectangle
	var cvImages []string
	for _, cvResult := range cvResults {
		rect = image.Rectangle{
			// cvResult.Points 顺序：左上 -> 右上 -> 右下 -> 左下
			Min: image.Point{
				X: int(cvResult.Points[0].X),
				Y: int(cvResult.Points[0].Y),
			},
			Max: image.Point{
				X: int(cvResult.Points[2].X),
				Y: int(cvResult.Points[2].Y),
			},
		}
		if rect.Min.X >= data.Scope[0] && rect.Max.X <= data.Scope[2] && rect.Min.Y >= data.Scope[1] && rect.Max.Y <= data.Scope[3] {
			cvImages = append(cvImages, cvResult.Image)

			rects = append(rects, rect)

			// match exactly, and not specify index, return the first one
			if data.Index == 0 {
				return rect, nil
			}
		}
	}

	if len(rects) == 0 {
		return image.Rectangle{}, errors.Wrap(code.CVImageNotFoundError,
			fmt.Sprintf("image not found"))
	}

	// get index
	idx := data.Index
	if idx > 0 {
		// NOTICE: index start from 1
		idx = idx - 1
	} else if idx < 0 {
		idx = len(rects) + idx
	}

	// index out of range
	if idx >= len(rects) {
		return image.Rectangle{}, errors.Wrap(code.CVImageNotFoundError,
			fmt.Sprintf("image found, index %d out of range", idx))
	}

	return rects[idx], nil
}
