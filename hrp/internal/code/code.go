package code

import (
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// general: [0, 2)
const (
	Success     = 0
	GeneralFail = 1
)

// environment: [2, 10)

// loader: [10, 20)
var (
	LoadError     = errors.New("load error")      // 10
	LoadJSONError = errors.New("load json error") // 11
	LoadYAMLError = errors.New("load yaml error") // 12
)

// parser: [20, 30)
var (
	ParseError          = errors.New("parse error")            // 20
	ParseStringError    = errors.New("parse string failed")    // 21
	ParseVariablesError = errors.New("parse variables failed") // 22
	ParseConfigError    = errors.New("parse config error")     // 25
)

// runner: [30, 40)

// summary: [40, 50)

// ios device related: [50, 60)
var (
	IOSDeviceConnectionError = errors.New("ios device connection error")  // 50
	IOSDeviceHTTPDriverError = errors.New("ios device HTTP driver error") // 51
	IOSDeviceUSBDriverError  = errors.New("ios device USB driver error")  // 52
	IOSScreenShotError       = errors.New("ios screenshot error")         // 55
	IOSCaptureLogError       = errors.New("ios capture log error")        // 56
)

// android device related: [60, 70)
var (
	AndroidDeviceConnectionError = errors.New("android device connection error") // 60
	AndroidDeviceDriverError     = errors.New("android device driver error")     // 61
	AndroidScreenShotError       = errors.New("android screenshot error")        // 65
	AndroidCaptureLogError       = errors.New("android capture log error")       // 66
)

// UI automation related: [70, 80)
var (
	MobileUIDriverError     = errors.New("mobile UI driver error")     // 70
	MobileUIValidationError = errors.New("mobile UI validation error") // 75
)

// OCR related: [80, 90)
var (
	OCREnvMissedError         = errors.New("OCR env missed error")      // 80
	OCRRequestError           = errors.New("OCR prepare request error") // 81
	OCRServiceConnectionError = errors.New("OCR service connect error") // 82
	OCRResponseError          = errors.New("OCR parse response error")  // 83
	OCRTextNotFoundError      = errors.New("OCR text not found")        // 84
)

// CV related: [90, 100)

var errorsMap = map[error]int{
	// loader
	LoadError:     10,
	LoadJSONError: 11,
	LoadYAMLError: 12,

	// parser
	ParseError:          20,
	ParseStringError:    21,
	ParseVariablesError: 22,
	ParseConfigError:    25,

	// ios related
	IOSDeviceConnectionError: 50,
	IOSDeviceHTTPDriverError: 51,
	IOSDeviceUSBDriverError:  52,
	IOSScreenShotError:       55,
	IOSCaptureLogError:       56,

	// android related
	AndroidDeviceConnectionError: 60,
	AndroidDeviceDriverError:     61,
	AndroidScreenShotError:       65,
	AndroidCaptureLogError:       66,

	// UI automation related
	MobileUIDriverError:     70,
	MobileUIValidationError: 75,

	// OCR related
	OCREnvMissedError:         80,
	OCRRequestError:           81,
	OCRServiceConnectionError: 82,
	OCRResponseError:          83,
	OCRTextNotFoundError:      84,
}

func GetErrorCode(err error) (exitCode int) {
	if err == nil {
		log.Info().Int("code", Success).Msg("hrp exit")
		return Success
	}

	e := errors.Cause(err)
	if code, ok := errorsMap[e]; ok {
		exitCode = code
	} else {
		exitCode = GeneralFail
	}

	log.Warn().Int("code", exitCode).Msg("hrp exit")
	return
}
