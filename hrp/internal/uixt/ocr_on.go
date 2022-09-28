//go:build ocr

package uixt

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/json"
)

var client = &http.Client{
	Timeout: time.Second * 10,
}

type OCRResult struct {
	Text   string   `json:"text"`
	Points []PointF `json:"points"`
}

type ResponseOCR struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	OCRResult []OCRResult `json:"ocrResult"`
}

type veDEMOCRService struct{}

func (s *veDEMOCRService) getOCRResult(imageBuf []byte) ([]OCRResult, error) {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	bodyWriter.WriteField("withDet", "true")
	// bodyWriter.WriteField("timestampOnly", "true")

	formWriter, err := bodyWriter.CreateFormFile("image", "screenshot.png")
	if err != nil {
		return nil, fmt.Errorf("create form file error: %v", err)
	}
	_, err = formWriter.Write(imageBuf)
	if err != nil {
		return nil, fmt.Errorf("write form error: %v", err)
	}

	err = bodyWriter.Close()
	if err != nil {
		return nil, fmt.Errorf("close body writer error: %v", err)
	}

	url, _ := base64.StdEncoding.DecodeString("aHR0cHM6Ly9odWJibGUuYnl0ZWRhbmNlLm5ldC92aWRlby9hcGkvdjEvYWxnb3JpdGhtL29jcg==")
	req, err := http.NewRequest("POST", string(url), bodyBuf)
	if err != nil {
		return nil, fmt.Errorf("construct request error: %v", err)
	}

	req.Header.Add("Content-Type", bodyWriter.FormDataContentType())
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http reqeust OCR server error: %v", err)
	}
	defer resp.Body.Close()

	results, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status code: %d, results: %v", resp.StatusCode, string(results))
	}

	var ocrResult ResponseOCR
	err = json.Unmarshal(results, &ocrResult)
	if err != nil {
		return nil, fmt.Errorf("json unmarshal response body error: %v", err)
	}

	return ocrResult.OCRResult, nil
}

func (s *veDEMOCRService) FindText(text string, imageBuf []byte, index ...int) (rect image.Rectangle, err error) {
	if len(index) == 0 {
		index = []int{0} // index not specified
	}

	ocrResults, err := s.getOCRResult(imageBuf)
	if err != nil {
		log.Error().Err(err).Msg("getOCRResult failed")
		return
	}

	var rects []image.Rectangle
	var ocrTexts []string
	for _, ocrResult := range ocrResults {
		ocrTexts = append(ocrTexts, ocrResult.Text)

		// not contains text
		if !strings.Contains(ocrResult.Text, text) {
			continue
		}

		rect = image.Rectangle{
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
		rects = append(rects, rect)

		// contains text while not match exactly
		if ocrResult.Text != text {
			continue
		}

		// match exactly, and not specify index, return the first one
		if index[0] == 0 {
			return rect, nil
		}
	}

	if len(rects) == 0 {
		return image.Rectangle{},
			fmt.Errorf("text %s not found in %v", text, ocrTexts)
	}

	// get index
	idx := index[0]
	if idx > 0 {
		// NOTICE: index start from 1
		idx = idx - 1
	} else if idx < 0 {
		idx = len(rects) + idx
	}

	// index out of range
	if idx >= len(rects) {
		return image.Rectangle{}, fmt.Errorf("text %s found %d, index %d out of range",
			text, len(rects), idx)
	}

	return rects[idx], nil
}

type OCRService interface {
	FindText(text string, imageBuf []byte, index ...int) (rect image.Rectangle, err error)
}

func (dExt *DriverExt) FindTextByOCR(ocrText string, index ...int) (x, y, width, height float64, err error) {
	var bufSource *bytes.Buffer
	if bufSource, err = dExt.takeScreenShot(); err != nil {
		err = fmt.Errorf("takeScreenShot error: %v", err)
		return
	}

	service := &veDEMOCRService{}
	rect, err := service.FindText(ocrText, bufSource.Bytes(), index...)
	if err != nil {
		log.Warn().Msgf("FindText failed: %s", err.Error())
		err = fmt.Errorf("FindText failed: %v", err)
		return
	}

	log.Info().Str("ocrText", ocrText).Msgf("FindText success")
	x, y, width, height = dExt.MappingToRectInUIKit(rect)
	return
}
