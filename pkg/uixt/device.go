package uixt

import (
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
)

// current implemeted device: IOSDevice, AndroidDevice, HarmonyDevice
type IDevice interface {
	Init() error  // init android device
	UUID() string // ios udid or android serial
	LogEnabled() bool

	// TODO: add ctx to NewDriver
	NewDriver(...option.DriverOption) (driverExt *DriverExt, err error)

	Install(appPath string, opts ...option.InstallOption) error
	Uninstall(packageName string) error

	GetPackageInfo(packageName string) (AppInfo, error)
	GetCurrentWindow() (windowInfo WindowInfo, err error)

	// Teardown() error
}

func NewDriver(device IDevice, opts ...option.DriverOption) (driver IWebDriver, err error) {
	return
}
