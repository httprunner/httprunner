package env

import "os"

var (
	WDA_USB_DRIVER       = os.Getenv("WDA_USB_DRIVER")
	WDA_LOCAL_PORT       = os.Getenv("WDA_LOCAL_PORT")
	WDA_LOCAL_MJPEG_PORT = os.Getenv("WDA_LOCAL_MJPEG_PORT")
	VEDEM_OCR_URL        = os.Getenv("VEDEM_OCR_URL")
	VEDEM_OCR_AK         = os.Getenv("VEDEM_OCR_AK")
	VEDEM_OCR_SK         = os.Getenv("VEDEM_OCR_SK")
	DISABLE_GA           = os.Getenv("DISABLE_GA")
	DISABLE_SENTRY       = os.Getenv("DISABLE_SENTRY")
	PYPI_INDEX_URL       = os.Getenv("PYPI_INDEX_URL")
	PATH                 = os.Getenv("PATH")
)
