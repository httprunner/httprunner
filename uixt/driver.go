package uixt

import (
	"bytes"
	_ "image/gif"
	_ "image/png"

	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/httprunner/httprunner/v5/uixt/types"
)

var (
	_ IDriver = (*ADBDriver)(nil)
	_ IDriver = (*UIA2Driver)(nil)
	_ IDriver = (*WDADriver)(nil)
	_ IDriver = (*HDCDriver)(nil)
	_ IDriver = (*BrowserDriver)(nil)

	// Ensure drivers implement SIMSupport interface
	_ SIMSupport = (*UIA2Driver)(nil)
	_ SIMSupport = (*WDADriver)(nil)
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
	PressButton(button types.DeviceButton) error

	// hover
	HoverBySelector(selector string, opts ...option.ActionOption) error
	// tap
	TapXY(x, y float64, opts ...option.ActionOption) error    // by percentage or absolute coordinate
	TapAbsXY(x, y float64, opts ...option.ActionOption) error // by absolute coordinate
	TapBySelector(text string, opts ...option.ActionOption) error
	DoubleTap(x, y float64, opts ...option.ActionOption) error // by absolute coordinate
	TouchAndHold(x, y float64, opts ...option.ActionOption) error
	// secondary click
	SecondaryClick(x, y float64) error
	SecondaryClickBySelector(selector string, options ...option.ActionOption) error
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

	// clipboard operations
	GetPasteboard() (string, error)
}

// SIMSupport interface defines simulated interaction methods
// Any driver that supports simulated touch and input should implement this interface
type SIMSupport interface {
	SIMClickAtPoint(x, y float64, opts ...option.ActionOption) error
	SIMSwipeWithDirection(direction string, fromX, fromY, simMinDistance, simMaxDistance float64, opts ...option.ActionOption) error
	SIMSwipeInArea(direction string, simAreaStartX, simAreaStartY, simAreaEndX, simAreaEndY, simMinDistance, simMaxDistance float64, opts ...option.ActionOption) error
	SIMSwipeFromPointToPoint(fromX, fromY, toX, toY float64, opts ...option.ActionOption) error
	SIMInput(text string, opts ...option.ActionOption) error
}
