package code

import (
	"github.com/pkg/errors"
)

// general: [0, 20)
const (
	SUCCESS = 0
	FAIL    = 1
)

// loader: [20, 40)
var (
	LoadError     = errors.New("load error")      // 20
	LoadJSONError = errors.New("load json error") // 21
	LoadYAMLError = errors.New("load yaml error") // 22
)

// parser: [40, 60)

// runner: [60, 100)

// ios related: [100, 120)
var (
	IOSScreenShotError = errors.New("ios screenshot error") // 100
)

// android related: [120, 140)

// OCR related: [140, 160)
var (
	OCREnvMissedError           = errors.New("veDEM OCR env missed error")                // 140
	OCRRequestError             = errors.New("vedem ocr prepare request error")           // 141
	OCRServiceConnectionError   = errors.New("vedem ocr service connect error")           // 142
	OCRResponseStatusCodeNot200 = errors.New("vedem ocr response status code is not 200") // 143
	OCRResponseError            = errors.New("vedem ocr parse response error")            // 143
	OCRTextNotFoundError        = errors.New("vedem ocr text not found")                  // 144
)

// CV related: [160, 180)

// report related: [200, 220)

var errorsMap = map[error]int{
	LoadJSONError: 10,
	LoadYAMLError: 11,
}

func GetErrorCode(err error) int {
	if err == nil {
		return SUCCESS
	}

	e := errors.Cause(err)
	if code, ok := errorsMap[e]; ok {
		return code
	}

	return FAIL
}
