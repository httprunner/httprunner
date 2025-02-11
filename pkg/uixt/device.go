package uixt

import (
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
	"github.com/httprunner/httprunner/v5/pkg/uixt/types"
)

// current implemeted device: IOSDevice, AndroidDevice, HarmonyDevice
type IDevice interface {
	UUID() string // ios udid or android serial
	NewDriver() (driver IDriver, err error)

	Setup() error
	Teardown() error

	Install(appPath string, opts ...option.InstallOption) error
	Uninstall(packageName string) error

	GetPackageInfo(packageName string) (types.AppInfo, error)

	// TODO: remove?
	LogEnabled() bool
}
