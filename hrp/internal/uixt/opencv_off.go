//go:build !opencv

package uixt

import (
	"image"

	"github.com/electricbubble/gwda"
	"github.com/rs/zerolog/log"
)

func Extend(driver gwda.WebDriver, options ...CVOption) (dExt *DriverExt, err error) {
	return extend(driver)
}

func (dExt *DriverExt) FindAllImageRect(search string) (rects []image.Rectangle, err error) {
	log.Fatal().Msg("opencv is not supported")
	return
}

func (dExt *DriverExt) FindImageRectInUIKit(imagePath string) (x, y, width, height float64, err error) {
	log.Fatal().Msg("opencv is not supported")
	return
}
