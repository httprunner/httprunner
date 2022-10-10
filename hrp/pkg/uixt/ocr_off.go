//go:build !ocr

package uixt

import "github.com/rs/zerolog/log"

func (dExt *DriverExt) FindTextByOCR(ocrText string, index ...int) (x, y, width, height float64, err error) {
	log.Fatal().Msg("OCR is not supported")
	return
}
