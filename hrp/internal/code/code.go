package code

import (
	"github.com/pkg/errors"
)

// general: [0, 20)
const (
	Success     = 0
	GeneralFail = 1
)

// loader: [20, 40)
var (
	LoadError     = errors.New("load error")      // 20
	LoadJSONError = errors.New("load json error") // 21
	LoadYAMLError = errors.New("load yaml error") // 22
)

// parser: [40, 60)

// runner: [60, 100)

// ios related: [100, 130)
var (
	IOSScreenShotError = errors.New("ios screenshot error") // 110
)

// android related: [130, 160)
var (
	AndroidScreenShotError = errors.New("android screenshot error") // 150
)

// OCR related: [160, 180)
var (
	OCREnvMissedError         = errors.New("veDEM OCR env missed error")      // 160
	OCRRequestError           = errors.New("vedem ocr prepare request error") // 161
	OCRServiceConnectionError = errors.New("vedem ocr service connect error") // 162
	OCRResponseError          = errors.New("vedem ocr parse response error")  // 163
	OCRTextNotFoundError      = errors.New("vedem ocr text not found")        // 164
)

// CV related: [180, 200)

// report related: [200, 220)

var errorsMap = map[error]int{
	LoadJSONError: 10,
	LoadYAMLError: 11,

	// ios related
	IOSScreenShotError: 110,

	// android related
	AndroidScreenShotError: 130,

	// OCR related
	OCREnvMissedError:         160,
	OCRRequestError:           161,
	OCRServiceConnectionError: 162,
	OCRResponseError:          163,
	OCRTextNotFoundError:      164,
}

func GetErrorCode(err error) int {
	if err == nil {
		return Success
	}

	e := errors.Cause(err)
	if code, ok := errorsMap[e]; ok {
		return code
	}

	return GeneralFail
}
