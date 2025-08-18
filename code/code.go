package code

import (
	"github.com/pkg/errors"
)

// general: [0, 2)
const (
	Success     = 0
	GeneralFail = 1
)

// environment: [2, 10)
var (
	ConfigureError        = errors.New("configure error")             // 3
	UnauthorizedError     = errors.New("unauthorized error")          // 4
	NetworkError          = errors.New("network error")               // 5
	NetworkConfigureError = errors.New("network configure error")     // 6
	PanicError            = errors.New("panic error")                 // 7
	InvalidPython3Venv    = errors.New("prepare python3 venv failed") // 9
)

// loader: [10, 20)
var (
	LoadFileError            = errors.New("load file error")            // 10
	LoadJSONError            = errors.New("load json error")            // 11
	LoadYAMLError            = errors.New("load yaml error")            // 12
	LoadEnvError             = errors.New("load .env error")            // 13
	LoadCSVError             = errors.New("load csv error")             // 14
	InvalidCaseError         = errors.New("invalid case error")         // 15
	UnsupportedFileExtension = errors.New("unsupported file extension") // 16
	ReferencedFileNotFound   = errors.New("referenced file not found")  // 17
	InvalidPluginFile        = errors.New("invalid plugin file")        // 18
	InvalidParamError        = errors.New("invalid param error")        // 19
)

// parser: [20, 30)
var (
	ParseError          = errors.New("parse error")            // 20
	VariableNotFound    = errors.New("variable not found")     // 21
	ParseFunctionError  = errors.New("parse function failed")  // 22
	CallFunctionError   = errors.New("call function failed")   // 23
	ParseVariablesError = errors.New("parse variables failed") // 24
	FFmpegError         = errors.New("ffmpeg error")           // 25
	FFprobeError        = errors.New("ffprobe error")          // 26
)

// runner: [30, 40)
var (
	StartRunnerFailed   = errors.New("start runner failed")    // 30
	InitPluginFailed    = errors.New("init plugin failed")     // 31
	BuildGoPluginFailed = errors.New("build go plugin failed") // 32
	BuildPyPluginFailed = errors.New("build py plugin failed") // 33
	MaxRetryError       = errors.New("max retry error")        // 37
	InterruptError      = errors.New("interrupt error")        // 38
	TimeoutError        = errors.New("timeout error")          // 39
)

// summary: [40, 50)
var (
	GenSummaryFailed         = errors.New("generate summary failed") // 40
	CollectValidResultFailed = errors.New("no valid result")         // 44
	DownloadFailed           = errors.New("download failed")         // 48
	UploadFailed             = errors.New("upload failed")           // 49
)

// device related: [50, 70)
var (
	DeviceConnectionError    = errors.New("device general connection error")    // 50
	DeviceHTTPDriverError    = errors.New("device HTTP driver error")           // 51
	DeviceUSBDriverError     = errors.New("device USB driver error")            // 52
	DeviceUntrustedCertError = errors.New("device app certificate not trusted") // 53
	DeviceAppNotInstalled    = errors.New("device app not installed")           // 59
	DeviceGetInfoError       = errors.New("device get info error")              // 60
	DeviceConfigureError     = errors.New("device configure error")             // 61
	DeviceShellExecError     = errors.New("device shell exec error")            // 62
	DeviceOfflineError       = errors.New("device offline")                     // 63
	DeviceInstallFailed      = errors.New("device install app failed")          // 64
	DeviceScreenShotError    = errors.New("device screenshot error")            // 65
	DeviceCaptureLogError    = errors.New("device capture log error")           // 66
	DeviceUIResponseSlow     = errors.New("device UI response slow")            // 67
)

// UI automation related: [70, 80)
var (
	MobileUIDriverAppNotInstalled         = errors.New("mobile UI driver app not installed")         // 68
	MobileUIDriverAppCrashed              = errors.New("mobile UI driver app crashed")               // 69
	MobileUIDriverError                   = errors.New("mobile UI driver error")                     // 70
	MobileUILaunchAppError                = errors.New("mobile UI launch app error")                 // 71
	MobileUITapError                      = errors.New("mobile UI tap error")                        // 72
	MobileUISwipeError                    = errors.New("mobile UI swipe error")                      // 73
	MobileUIInputError                    = errors.New("mobile UI input error")                      // 74
	MobileUIValidationError               = errors.New("mobile UI validation error")                 // 75
	MobileUIAssertForegroundAppError      = errors.New("mobile UI assert foreground app error")      // 76
	MobileUIAssertForegroundActivityError = errors.New("mobile UI assert foreground activity error") // 77
	MobileUIPopupError                    = errors.New("mobile UI popup error")                      // 78
	LoopActionNotFoundError               = errors.New("loop action not found error")                // 79
)

// CV related: [80, 90)
var (
	CVEnvMissedError      = errors.New("CV env missed error")      // 80
	CVPrepareRequestError = errors.New("CV prepare request error") // 81
	CVRequestServiceError = errors.New("CV request service error") // 82
	CVParseResponseError  = errors.New("CV parse response error")  // 83
	CVResultNotFoundError = errors.New("CV result not found")      // 84

	StateUnknowError = errors.New("detect state failed") // 85
)

// trackings related: [90, 100)
var (
	TrackingGetError   = errors.New("get trackings failed")  // 90
	TrackingFomatError = errors.New("tracking format error") // 91
)

// risk control related: [100, 110)
var (
	RiskControlLogout            = errors.New("risk control logout")             // 100
	RiskControlSlideVerification = errors.New("risk control slide verification") // 101
	RiskControlAccountActivation = errors.New("risk control account activation") // 102
)

// LLM related: [110, 120)
var (
	LLMEnvMissedError              = errors.New("missed LLM env error")               // 110
	LLMPrepareRequestError         = errors.New("prepare LLM request error")          // 111
	LLMRequestServiceError         = errors.New("request LLM service error")          // 112
	LLMParsePlanningResponseError  = errors.New("parse LLM planning response error")  // 113
	LLMParseAssertionResponseError = errors.New("parse LLM assertion response error") // 114
	LLMParseQueryResponseError     = errors.New("parse LLM query response error")     // 115
)

var errorsMap = map[error]int{
	// environment
	ConfigureError:        3,
	UnauthorizedError:     4,
	NetworkError:          5,
	NetworkConfigureError: 6,
	PanicError:            7,
	InvalidPython3Venv:    9,

	// loader
	LoadFileError:            10,
	LoadJSONError:            11,
	LoadYAMLError:            12,
	LoadEnvError:             13,
	LoadCSVError:             14,
	InvalidCaseError:         15,
	UnsupportedFileExtension: 16,
	ReferencedFileNotFound:   17,
	InvalidPluginFile:        18,
	InvalidParamError:        19,

	// parser
	ParseError:          20,
	VariableNotFound:    21,
	ParseFunctionError:  22,
	CallFunctionError:   23,
	ParseVariablesError: 24,
	FFmpegError:         25,
	FFprobeError:        26,

	// runner
	StartRunnerFailed:   30,
	InitPluginFailed:    31,
	BuildGoPluginFailed: 32,
	BuildPyPluginFailed: 33,
	MaxRetryError:       37,
	InterruptError:      38,
	TimeoutError:        39,

	// summary
	GenSummaryFailed:         40,
	CollectValidResultFailed: 44,
	DownloadFailed:           48,
	UploadFailed:             49,

	// device related
	DeviceConnectionError:    50,
	DeviceHTTPDriverError:    51,
	DeviceUSBDriverError:     52,
	DeviceUntrustedCertError: 53,
	DeviceAppNotInstalled:    59,
	DeviceGetInfoError:       60,
	DeviceConfigureError:     61,
	DeviceShellExecError:     62,
	DeviceOfflineError:       63,
	DeviceInstallFailed:      64,
	DeviceScreenShotError:    65,
	DeviceCaptureLogError:    66,
	DeviceUIResponseSlow:     67,

	// UI automation related
	MobileUIDriverAppNotInstalled:         68,
	MobileUIDriverAppCrashed:              69,
	MobileUIDriverError:                   70,
	MobileUILaunchAppError:                71,
	MobileUITapError:                      72,
	MobileUISwipeError:                    73,
	MobileUIInputError:                    74,
	MobileUIValidationError:               75,
	MobileUIAssertForegroundAppError:      76,
	MobileUIAssertForegroundActivityError: 77,
	MobileUIPopupError:                    78,
	LoopActionNotFoundError:               79,

	// CV related
	CVEnvMissedError:      80,
	CVPrepareRequestError: 81,
	CVRequestServiceError: 82,
	CVParseResponseError:  83,
	CVResultNotFoundError: 84,

	StateUnknowError: 85,

	// LLM related
	LLMEnvMissedError:              110,
	LLMPrepareRequestError:         111,
	LLMRequestServiceError:         112,
	LLMParsePlanningResponseError:  113,
	LLMParseAssertionResponseError: 114,
	LLMParseQueryResponseError:     115,

	// trackings related
	TrackingGetError:   90,
	TrackingFomatError: 91,

	// risk control related
	RiskControlLogout:            100,
	RiskControlSlideVerification: 101,
	RiskControlAccountActivation: 102,
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
	return
}

func GetErrorByCode(errCode int) error {
	if errCode < 0 {
		return nil
	}
	for key, value := range errorsMap {
		if value == errCode {
			return key
		}
	}
	return nil
}
