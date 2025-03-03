package uixt

import (
	"bytes"
	_ "image/gif"
	_ "image/png"
	"time"

	"github.com/httprunner/httprunner/v5/pkg/uixt/ai"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
	"github.com/httprunner/httprunner/v5/pkg/uixt/types"
)

var (
	_ IDriver = (*ADBDriver)(nil)
	_ IDriver = (*UIA2Driver)(nil)
	_ IDriver = (*WDADriver)(nil)
	_ IDriver = (*HDCDriver)(nil)
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
	ScreenRecord(duration time.Duration) (videoPath string, err error)
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
	TapXY(x, y float64, opts ...option.ActionOption) error     // by percentage
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
	ClearImages() error

	// triggers the log capture and returns the log entries
	StartCaptureLog(identifier ...string) error
	StopCaptureLog() (result interface{}, err error)
}

func NewXTDriver(driver IDriver, opts ...ai.AIServiceOption) *XTDriver {
	services := ai.NewAIService(opts...)
	driverExt := &XTDriver{
		IDriver:    driver,
		CVService:  services.ICVService,
		LLMService: services.ILLMService,

		screenResults: make([]*ScreenResult, 0),
	}
	return driverExt
}

// XTDriver = IDriver + AI
type XTDriver struct {
	IDriver
	CVService  ai.ICVService  // OCR/CV
	LLMService ai.ILLMService // LLM

	// cache screenshot results
	screenResults []*ScreenResult
}

func (dExt *XTDriver) GetIDriver() IDriver {
	return dExt.IDriver
}

func (dExt *XTDriver) GetWebDriver() IBrowserWebDriver {
	return dExt.GetIDriver().(*BrowserWebDriver)
}

type IXTDriver interface {
	IDriver
	GetIDriver() IDriver
	GetWebDriver() IBrowserWebDriver
	GetScreenResult(opts ...option.ActionOption) (screenResult *ScreenResult, err error)
	DoAction(action MobileAction) (err error)
}

type IBrowserWebDriver interface {
	IDriver
	Hover(x, y float64) (err error)
	RightClick(x, y float64) (err error)
	Scroll(delta int) (err error)
	// TODO: move x,y parameters to option
	UploadFile(x, y float64, FileUrl, FileFormat string) (err error)
}
