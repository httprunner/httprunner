package uixt

import (
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
)

// current implemeted device: IOSDevice, AndroidDevice, HarmonyDevice
type IDevice interface {
	UUID() string // ios udid or android serial

	Setup() error
	Teardown() error

	Install(appPath string, opts ...option.InstallOption) error
	Uninstall(packageName string) error

	GetPackageInfo(packageName string) (AppInfo, error)

	// TODO: remove?
	LogEnabled() bool
}
