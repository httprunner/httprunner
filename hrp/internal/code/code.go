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
var (
	ParseError          = errors.New("parse error")            // 40
	ParseConfigError    = errors.New("parse config error")     // 41
	ParseStringError    = errors.New("parse string failed")    // 42
	ParseVariablesError = errors.New("parse variables failed") // 43
)

// runner: [60, 100)

// ios related: [100, 130)
var (
	IOSDeviceConnectionError = errors.New("ios device connection error")  // 100
	IOSDeviceHTTPDriverError = errors.New("ios device HTTP driver error") // 101
	IOSDeviceUSBDriverError  = errors.New("ios device USB driver error")  // 102
	IOSScreenShotError       = errors.New("ios screenshot error")         // 110
	IOSCaptureLogError       = errors.New("ios capture log error")        // 111

	MobileUIDriverError     = errors.New("mobile UI driver error")     // 120
	MobileUIValidationError = errors.New("mobile UI validation error") // 121
)

// android related: [130, 160)
var (
	AndroidDeviceConnectionError = errors.New("android device connection error") // 130
	AndroidDeviceDriverError     = errors.New("android device driver error")     // 131
	AndroidScreenShotError       = errors.New("android screenshot error")        // 150
	AndroidCaptureLogError       = errors.New("android capture log error")       // 151
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
	// loader
	LoadJSONError: 10,
	LoadYAMLError: 11,

	// parser
	ParseError:          40,
	ParseConfigError:    41,
	ParseStringError:    42,
	ParseVariablesError: 43,

	// ios related
	IOSDeviceConnectionError: 100,
	IOSDeviceHTTPDriverError: 101,
	IOSDeviceUSBDriverError:  102,
	IOSScreenShotError:       110,
	IOSCaptureLogError:       111,

	// android related
	AndroidDeviceConnectionError: 130,
	AndroidDeviceDriverError:     131,
	AndroidScreenShotError:       150,
	AndroidCaptureLogError:       151,

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
