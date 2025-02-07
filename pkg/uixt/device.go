package uixt

import (
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
)

// current implemeted device: IOSDevice, AndroidDevice, HarmonyDevice
type IDevice interface {
	Setup() error
	Teardown() error

	UUID() string // ios udid or android serial
	LogEnabled() bool

	// TODO: remove
	NewDriver(...option.DriverOption) (driverExt *DriverExt, err error)

	Install(appPath string, opts ...option.InstallOption) error
	Uninstall(packageName string) error

	GetPackageInfo(packageName string) (AppInfo, error)
	GetCurrentWindow() (windowInfo WindowInfo, err error)
}
