package env

import "os"

var (
	WDA_USB_DRIVER = os.Getenv("WDA_USB_DRIVER")
	VEDEM_OCR_URL  = os.Getenv("VEDEM_OCR_URL")
	DISABLE_GA     = os.Getenv("DISABLE_GA")
	DISABLE_SENTRY = os.Getenv("DISABLE_SENTRY")
	PYPI_INDEX_URL = os.Getenv("PYPI_INDEX_URL")
	PATH           = os.Getenv("PATH")
)
