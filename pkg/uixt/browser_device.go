package uixt

import (
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
	"github.com/httprunner/httprunner/v5/pkg/uixt/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type BrowserDevice struct {
	BrowserId string `json:"browser_id,omitempty" yaml:"browser_id,omitempty"`
	LogOn     bool   `json:"log_on,omitempty" yaml:"log_on,omitempty"`
}
type BrowserDeviceOption func(*BrowserDevice)

func WithBrowserId(serial string) BrowserDeviceOption {
	return func(device *BrowserDevice) {
		device.BrowserId = serial
	}
}

func NewBrowserDevice(options ...BrowserDeviceOption) (device *BrowserDevice, err error) {
	device = &BrowserDevice{}
	for _, option := range options {
		option(device)
	}

	if device.BrowserId == "" {
		browserInfo, err := CreateBrowser(3600)
		if err != nil {
			log.Error().Err(err).Msg("failed to create browser")
			return nil, err
		}
		device.BrowserId = browserInfo.ContextId
	}

	return device, nil
}

func (dev *BrowserDevice) UUID() string {
	return dev.BrowserId
}

func (dev *BrowserDevice) Setup() error {
	return nil
}

func (dev *BrowserDevice) LogEnabled() bool {
	return dev.LogOn
}

func (dev *BrowserDevice) Teardown() error {
	return nil
}

func (dev *BrowserDevice) Install(appPath string, opts ...option.InstallOption) error {
	return errors.New("not support")
}

func (dev *BrowserDevice) Uninstall(packageName string) error {
	return errors.New("not support")
}

func (dev *BrowserDevice) GetPackageInfo(packageName string) (types.AppInfo, error) {
	return types.AppInfo{}, errors.New("not support")
}

func (dev *BrowserDevice) NewDriver() (driver IDriver, err error) {
	// var driver WebDriver
	driver, err = NewBrowserWebDriver(dev.UUID())
	if err != nil {
		return nil, err
	}
	return driver, nil
}
