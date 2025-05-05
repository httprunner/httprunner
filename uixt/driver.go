package uixt

import (
	"bytes"
	_ "image/gif"
	_ "image/png"

	"github.com/httprunner/httprunner/v5/uixt/ai"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/httprunner/httprunner/v5/uixt/types"
	"github.com/rs/zerolog/log"
)

var (
	_ IDriver = (*ADBDriver)(nil)
	_ IDriver = (*UIA2Driver)(nil)
	_ IDriver = (*WDADriver)(nil)
	_ IDriver = (*HDCDriver)(nil)
	_ IDriver = (*BrowserDriver)(nil)
)

// current implemeted driver: ADBDriver, UIA2Driver, WDADriver, HDCDriver
type IDriver interface {
	GetDevice() IDevice
	Setup() error
	TearDown() error

	// session
	InitSession(capabilities option.Capabilities) error
	GetSession() *DriverSession
	DeleteSession() error

	// device info and status
	Status() (types.DeviceStatus, error)
	DeviceInfo() (types.DeviceInfo, error)
	BatteryInfo() (types.BatteryInfo, error)
	ForegroundInfo() (app types.AppInfo, err error)
	WindowSize() (types.Size, error)
	ScreenShot(opts ...option.ActionOption) (*bytes.Buffer, error)
	ScreenRecord(opts ...option.ActionOption) (videoPath string, err error)
	Source(srcOpt ...option.SourceOption) (string, error)
	Orientation() (orientation types.Orientation, err error)
	Rotation() (rotation types.Rotation, err error)

	// config
	SetRotation(rotation types.Rotation) error
	SetIme(ime string) error

	// actions
	Home() error
	Unlock() error
	Back() error
	// tap
	TapXY(x, y float64, opts ...option.ActionOption) error     // by percentage or absolute coordinate
	TapAbsXY(x, y float64, opts ...option.ActionOption) error  // by absolute coordinate
	DoubleTap(x, y float64, opts ...option.ActionOption) error // by absolute coordinate
	TouchAndHold(x, y float64, opts ...option.ActionOption) error
	// swipe
	Drag(fromX, fromY, toX, toY float64, opts ...option.ActionOption) error
	Swipe(fromX, fromY, toX, toY float64, opts ...option.ActionOption) error // by percentage
	// input
	Input(text string, opts ...option.ActionOption) error
	Backspace(count int, opts ...option.ActionOption) error

	// app related
	AppLaunch(packageName string) error
	AppTerminate(packageName string) (bool, error)
	AppClear(packageName string) error

	// image related
	PushImage(localPath string) error
	PullImages(localDir string) error
	ClearImages() error

	// files related
	PushFile(localPath string, remoteDir string) error
	PullFiles(localDir string, remoteDirs ...string) error
	ClearFiles(paths ...string) error

	// triggers the log capture and returns the log entries
	StartCaptureLog(identifier ...string) error
	StopCaptureLog() (result interface{}, err error)
}

func NewXTDriver(driver IDriver, opts ...option.AIServiceOption) (*XTDriver, error) {
	driverExt := &XTDriver{
		IDriver: driver,
	}

	services := option.NewAIServiceOptions(opts...)

	var err error
	if services.CVService != "" {
		driverExt.CVService, err = ai.NewCVService(services.CVService)
		if err != nil {
			log.Error().Err(err).Msg("init vedem image service failed")
			return nil, err
		}
	}
	if services.LLMService != "" {
		driverExt.LLMService, err = ai.NewLLMService(services.LLMService)
		if err != nil {
			log.Error().Err(err).Msg("init llm service failed")
			return nil, err
		}
	}

	return driverExt, nil
}

// XTDriver = IDriver + AI
type XTDriver struct {
	IDriver
	CVService  ai.ICVService  // OCR/CV
	LLMService ai.ILLMService // LLM
}
