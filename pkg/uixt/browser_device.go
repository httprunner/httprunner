package uixt

import (
	"bytes"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
	"github.com/httprunner/httprunner/v5/pkg/uixt/types"
)

type BrowserDevice struct {
	Options *option.BrowserDeviceOptions
}

func NewBrowserDevice(opts ...option.BrowserDeviceOption) (device *BrowserDevice, err error) {
	options := &option.BrowserDeviceOptions{}
	for _, option := range opts {
		option(options)
	}

	if options.BrowserID == "" {
		browserInfo, err := CreateBrowser(3600)
		if err != nil {
			log.Error().Err(err).Msg("failed to create browser")
			return nil, err
		}
		options.BrowserID = browserInfo.ContextId
	}

	device = &BrowserDevice{
		Options: options,
	}
	log.Info().Str("browserID", device.Options.BrowserID).Msg("init browser device")

	return device, nil
}

func (dev *BrowserDevice) UUID() string {
	return dev.Options.BrowserID
}

func (dev *BrowserDevice) Setup() error {
	return nil
}

func (dev *BrowserDevice) LogEnabled() bool {
	return dev.Options.LogOn
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

func (dev *BrowserDevice) ScreenShot() (*bytes.Buffer, error) {
	return nil, errors.New("not support")
}
