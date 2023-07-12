package uixt

import (
	"bytes"

	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
)

// newVEDEMSDService return image service for scenario (live type) detection
func newVEDEMSDService() (*veDEMImageService, error) {
	return newVEDEMImageService("liveType", "upload")
}

func (dExt *DriverExt) DetectLiveType(options ...ActionOptions) (liveType string, err error) {
	var bufSource *bytes.Buffer
	var imagePath string
	if bufSource, imagePath, err = dExt.takeScreenShot(
		builtin.GenNameWithTimestamp("%d_ocr")); err != nil {
		return
	}

	vedemSDService, err := newVEDEMSDService()
	if err != nil {
		return
	}
	imageResult, err := vedemSDService.GetImage(bufSource)
	if err != nil {
		log.Error().Err(err).Msg("GetImage from vedemCPService failed")
		return
	}

	imageUrl := imageResult.URL
	if imageUrl != "" {
		dExt.cacheStepData.screenShotsUrls[imagePath] = imageUrl
		log.Debug().Str("imagePath", imagePath).Str("imageUrl", imageUrl).Msg("log screenshot")
	}

	return imageResult.LiveType, nil
}

func (dExt *DriverExt) AssertLiveType(expect string) bool {
	st, _ := dExt.DetectLiveType()
	return expect == st
}
