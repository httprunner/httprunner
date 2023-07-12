package uixt

import (
	"bytes"
	"fmt"
	"image"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
	"github.com/httprunner/httprunner/v4/hrp/internal/code"
)

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
