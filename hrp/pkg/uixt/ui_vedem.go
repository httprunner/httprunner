package uixt

import (
	"bytes"
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
	"github.com/httprunner/httprunner/v4/hrp/internal/json"
)

const (
	ShoppingBag = "shoppingbag"
	DyHouse     = "dyhouse"
)

type UIResult map[string][]Box

type UIResponse struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Result  UIResult `json:"result"`
}

type veDEMUIService struct{}

func newVEDEMUIService() (*veDEMUIService, error) {
	if err := checkUIEnv(); err != nil {
		return nil, err
	}
	return &veDEMUIService{}, nil
}

func checkUIEnv() error {
	if env.VEDEM_UI_URL == "" {
		return errors.Wrap(code.CVEnvMissedError, "VEDEM_UI_URL missed")
	}
	if env.VEDEM_UI_AK == "" {
		return errors.Wrap(code.CVEnvMissedError, "VEDEM_UI_AK missed")
	}
	if env.VEDEM_UI_SK == "" {
		return errors.Wrap(code.CVEnvMissedError, "VEDEM_UI_SK missed")
	}
	return nil
}

func (s *veDEMUIService) getUIResult(uiType string, sourceImage []byte) (UIResult, error) {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	bodyWriter.WriteField("types", uiType)

	formWriter, err := bodyWriter.CreateFormFile("image", "screenshot.png")
	if err != nil {
		return nil, errors.Wrap(code.CVRequestError,
			fmt.Sprintf("create form file error: %v", err))
	}
	size, err := formWriter.Write(sourceImage)
	if err != nil {
		return nil, errors.Wrap(code.CVRequestError,
			fmt.Sprintf("write form error: %v", err))
	}

	err = bodyWriter.Close()
	if err != nil {
		return nil, errors.Wrap(code.CVRequestError,
			fmt.Sprintf("close body writer error: %v", err))
	}

	req, err := http.NewRequest("POST", env.VEDEM_UI_URL, bodyBuf)
	if err != nil {
		return nil, errors.Wrap(code.CVRequestError,
			fmt.Sprintf("construct request error: %v", err))
	}

	signToken := "UNSIGNED-PAYLOAD"
	token := builtin.Sign("auth-v2", env.VEDEM_OCR_AK, env.VEDEM_OCR_SK, []byte(signToken))
	req.Header.Add("Agw-Auth", token)
	req.Header.Add("Agw-Auth-Content", signToken)
	req.Header.Add("Content-Type", bodyWriter.FormDataContentType())
	req.Header.Add("x-tt-env", "boe_0529162338")

	var resp *http.Response
	// retry 3 times
	for i := 1; i <= 3; i++ {
		resp, err = client.Do(req)
		var logID string
		if resp != nil {
			logID = getLogID(resp.Header)
		}
		if err == nil && resp.StatusCode == http.StatusOK {
			log.Debug().
				Str("X-TT-LOGID", logID).
				Int("imageBufSize", size).
				Msg("request UI service success")
			break
		}
		log.Error().Err(err).
			Str("X-TT-LOGID", logID).
			Int("imageBufSize", size).
			Msgf("request UI service failed, retry %d", i)
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

	var uiResult UIResponse
	err = json.Unmarshal(results, &uiResult)
	if err != nil {
		return nil, errors.Wrap(code.CVResponseError,
			fmt.Sprintf("json unmarshal response body error: %v", err))
	}

	return uiResult.Result, nil
}

func (s *veDEMUIService) FindUI(uiType string, byteSource []byte, options ...DataOption) (rect image.Rectangle, err error) {
	data := NewDataOptions(options...)

	uiResultMap, err := s.getUIResult(uiType, byteSource)
	if err != nil {
		log.Error().Err(err).Msg("getUIResult failed")
		return
	}

	uiResult, ok := uiResultMap[uiType]
	if !ok {
		err = fmt.Errorf("UI type %v not detected", uiResult)
		log.Error().Err(err).Msg("getUIResult failed")
		return
	}

	var rects []image.Rectangle
	for _, box := range uiResult {
		rect = image.Rectangle{
			// cvResult.Points 顺序：左上 -> 右上 -> 右下 -> 左下
			Min: image.Point{
				X: int(box.Point.X),
				Y: int(box.Point.Y),
			},
			Max: image.Point{
				X: int(box.Point.X + box.Width),
				Y: int(box.Point.Y + box.Height),
			},
		}
		if rect.Min.X >= data.Scope[0] && rect.Max.X <= data.Scope[2] && rect.Min.Y >= data.Scope[1] && rect.Max.Y <= data.Scope[3] {
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
