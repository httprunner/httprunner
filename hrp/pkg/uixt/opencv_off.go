//go:build !opencv

package uixt

import (
	"image"

	"github.com/rs/zerolog/log"
)

func (dExt *DriverExt) extendCV(options ...CVOption) (err error) {
	return nil
}

func (dExt *DriverExt) FindAllImageRect(search string) (rects []image.Rectangle, err error) {
	log.Fatal().Msg("opencv is not supported")
	return
}

func (dExt *DriverExt) FindImageRectInUIKit(imagePath string, options ...ActionOption) (point PointF, err error) {
	log.Fatal().Msg("opencv is not supported")
	return
}
