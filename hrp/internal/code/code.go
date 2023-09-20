package code

import (
	"fmt"

	"github.com/pkg/errors"
)

// general: [0, 2)
const (
	Success     = 0
	GeneralFail = 1
)

// environment: [2, 10)
var (
	InvalidPython3Venv = errors.New("prepare python3 venv failed") // 9
)

// loader: [10, 20)
var (
	LoadFileError            = errors.New("load file error")            // 10
	LoadJSONError            = errors.New("load json error")            // 11
	LoadYAMLError            = errors.New("load yaml error")            // 12
	LoadEnvError             = errors.New("load .env error")            // 13
	LoadCSVError             = errors.New("load csv error")             // 14
	InvalidCaseFormat        = errors.New("invalid case format")        // 15
	UnsupportedFileExtension = errors.New("unsupported file extension") // 16
	ReferencedFileNotFound   = errors.New("referenced file not found")  // 17
	InvalidPluginFile        = errors.New("invalid plugin file")        // 18
)

// parser: [20, 30)
var (
	ParseError          = errors.New("parse error")            // 20
	VariableNotFound    = errors.New("variable not found")     // 21
	ParseFunctionError  = errors.New("parse function failed")  // 22
	CallFunctionError   = errors.New("call function failed")   // 23
	ParseVariablesError = errors.New("parse variables failed") // 24
)

// runner: [30, 40)

var (
	InitPluginFailed    = errors.New("init plugin failed")     // 31
	BuildGoPluginFailed = errors.New("build go plugin failed") // 32
	BuildPyPluginFailed = errors.New("build py plugin failed") // 33
	InterruptError      = errors.New("interrupt error")        // 38
	TimeoutError        = errors.New("timeout error")          // 39
)

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
	AndroidDeviceConnectionError        = errors.New("android device general connection error") // 60
	AndroidDeviceConnectionRefusedError = errors.New("android device connection refused")       // 61
	AndroidShellExecError               = errors.New("android adb shell exec error")            // 62
	AndroidDeviceOfflineError           = errors.New("android device offline")                  // 63
	AndroidScreenShotError              = errors.New("android screenshot error")                // 65
	AndroidCaptureLogError              = errors.New("android capture log error")               // 66
)

// UI automation related: [70, 80)
var (
	MobileUIDriverError                   = errors.New("mobile UI driver error")                     // 70
	MobileUILaunchAppError                = errors.New("mobile UI launch app error")                 // 71
	MobileUIValidationError               = errors.New("mobile UI validation error")                 // 75
	MobileUIAssertForegroundAppError      = errors.New("mobile UI assert foreground app error")      // 76
	MobileUIAssertForegroundActivityError = errors.New("mobile UI assert foreground activity error") // 77
	MobileUIPopupError                    = errors.New("mobile UI popup error")                      // 78
	LoopActionNotFoundError               = errors.New("loop action not found error")                // 79
)

// CV related: [80, 90)
var (
	CVEnvMissedError         = errors.New("CV env missed error")      // 80
	CVRequestError           = errors.New("CV prepare request error") // 81
	CVServiceConnectionError = errors.New("CV service connect error") // 82
	CVResponseError          = errors.New("CV parse response error")  // 83
	CVResultNotFoundError    = errors.New("CV result not found")      // 84
)

// trackings related: [90, 100)
var (
	TrackingGetError = errors.New("get trackings failed") // 90
)

var errorsMap = map[error]int{
	// environment
	InvalidPython3Venv: 9,

	// loader
	LoadFileError:            10,
	LoadJSONError:            11,
	LoadYAMLError:            12,
	LoadEnvError:             13,
	LoadCSVError:             14,
	InvalidCaseFormat:        15,
	UnsupportedFileExtension: 16,
	ReferencedFileNotFound:   17,
	InvalidPluginFile:        18,

	// parser
	ParseError:          20,
	VariableNotFound:    21,
	ParseFunctionError:  22,
	CallFunctionError:   23,
	ParseVariablesError: 24,

	// runner
	InitPluginFailed:    31,
	BuildGoPluginFailed: 32,
	BuildPyPluginFailed: 33,
	InterruptError:      38,
	TimeoutError:        39,

	// ios related
	IOSDeviceConnectionError: 50,
	IOSDeviceHTTPDriverError: 51,
	IOSDeviceUSBDriverError:  52,
	IOSScreenShotError:       55,
	IOSCaptureLogError:       56,

	// android related
	AndroidDeviceConnectionError:        60,
	AndroidDeviceConnectionRefusedError: 61,
	AndroidShellExecError:               62,
	AndroidDeviceOfflineError:           63,
	AndroidScreenShotError:              65,
	AndroidCaptureLogError:              66,

	// UI automation related
	MobileUIDriverError:                   70,
	MobileUILaunchAppError:                71,
	MobileUIValidationError:               75,
	MobileUIAssertForegroundAppError:      76,
	MobileUIAssertForegroundActivityError: 77,
	MobileUIPopupError:                    78,
	LoopActionNotFoundError:               79,

	// OCR related
	CVEnvMissedError:         80,
	CVRequestError:           81,
	CVServiceConnectionError: 82,
	CVResponseError:          83,
	CVResultNotFoundError:    84,

	// trackings related
	TrackingGetError: 90,
}

func IsErrorPredefined(err error) bool {
	_, ok := errorsMap[errors.Cause(err)]
	return ok
}

func GetErrorCode(err error) (errCode int) {
	if err == nil {
		return Success
	}

	e := errors.Cause(err)
	if code, ok := errorsMap[e]; ok {
		errCode = code
	} else {
		errCode = GeneralFail
	}

	fmt.Printf("hrp exit %d\n", errCode)
	return
}
