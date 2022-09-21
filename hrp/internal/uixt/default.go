//go:build !opencv

package uixt

import (
	"image"

	"github.com/rs/zerolog/log"
)

func (dExt *DriverExt) extendOpenCV(threshold float64, matchMode ...TemplateMatchMode) (err error) {
	log.Fatal().Msg("opencv is not supported")
	return
}

func (dExt *DriverExt) FindAllImageRect(search string) (rects []image.Rectangle, err error) {
	log.Fatal().Msg("opencv is not supported")
	return
}

func (dExt *DriverExt) FindImageRectInUIKit(imagePath string) (x, y, width, height float64, err error) {
	log.Fatal().Msg("opencv is not supported")
	return
}

func (dExt *DriverExt) MappingToRectInUIKit(rect image.Rectangle) (x, y, width, height float64) {
	log.Fatal().Msg("opencv is not supported")
	return
}
