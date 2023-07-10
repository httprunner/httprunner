package uixt

import (
	"bytes"
	"fmt"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
	"github.com/httprunner/httprunner/v4/hrp/internal/code"
)

type CPResult struct {
	Point  PointF  `json:"point"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

func (c CPResult) Center() PointF {
	return getCenterPoint(c.Point, c.Width, c.Height)
}

type CPResults []CPResult

// newVEDEMCPService return image service for closing popup
func newVEDEMCPService() (*veDEMImageService, error) {
	return newVEDEMImageService("close", "upload")
}

// FindPopupCloseButton takes a screenshot, returns the image recognization result
func (dExt *DriverExt) FindPopupCloseButton(options ...ActionOption) (point PointF, err error) {
	var bufSource *bytes.Buffer
	var imagePath string
	if bufSource, imagePath, err = dExt.takeScreenShot(
		builtin.GenNameWithTimestamp("%d_ocr")); err != nil {
		return
	}

	vedemCPService, err := newVEDEMCPService()
	if err != nil {
		return
	}
	imageResult, err := vedemCPService.GetImage(bufSource)
	if err != nil {
		log.Error().Err(err).Msg("GetImage from vedemCPService failed")
		return
	}

	// TODO: save popup closing data to evaluate performance of vedemCPService
	imageUrl := imageResult.URL
	if imageUrl != "" {
		dExt.cacheStepData.screenShotsUrls[imagePath] = imageUrl
		log.Debug().Str("imagePath", imagePath).Str("imageUrl", imageUrl).Msg("log screenshot")
	}

	cpResult := imageResult.CPResult

	actionOptions := NewActionOptions(options...)
	// get index
	idx := actionOptions.Index
	if idx < 0 {
		idx = len(cpResult) + idx
	}

	// index out of range
	if idx >= len(cpResult) || idx < 0 {
		return PointF{}, errors.Wrap(code.OCRTextNotFoundError,
			fmt.Sprintf("index %d out of range", idx))
	}
	return cpResult[idx].Center(), nil
}

func (dExt *DriverExt) ClosePopupHandler() {
	retryCount := 3
	for retryCount > 0 {
		rect, err := dExt.FindPopupCloseButton()
		if err != nil {
			break
		}
		err = dExt.Driver.TapFloat(rect.X, rect.Y)
		if err != nil {
			break
		}
		retryCount--
	}
}
