package uixt

import (
	"code.byted.org/iesqa/ghdc"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/pkg/uixt/ai"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
	"github.com/httprunner/httprunner/v5/pkg/uixt/types"
)

type HarmonyDevice struct {
	*ghdc.Device
	Options *option.HarmonyDeviceOptions
}

func NewHarmonyDevice(opts ...option.HarmonyDeviceOption) (device *HarmonyDevice, err error) {
	deviceConfig := option.NewHarmonyDeviceOptions(opts...)

	// get all attached android devices
	hdcClient, err := ghdc.NewClientWith(option.HdcServerHost, option.HdcServerPort)
	if err != nil {
		return nil, err
	}
	devices, err := hdcClient.DeviceList()
	if err != nil {
		return nil, err
	}
	if len(devices) == 0 {
		return nil, errors.Wrapf(code.DeviceConnectionError,
			"no attached harmony devices")
	}

	// filter device by serial
	var harmonyDevice *ghdc.Device
	if deviceConfig.ConnectKey == "" {
		if len(devices) > 1 {
			return nil, errors.Wrap(code.DeviceConnectionError,
				"more than one device connected, please specify the serial")
		}
		harmonyDevice = &devices[0]
		deviceConfig.ConnectKey = harmonyDevice.Serial()
		log.Warn().Str("serial", deviceConfig.ConnectKey).
			Msg("harmony ConnectKey is not specified, select the attached one")
	} else {
		for _, d := range devices {
			if d.Serial() == deviceConfig.ConnectKey {
				harmonyDevice = &d
				break
			}
		}
		if harmonyDevice == nil {
			return nil, errors.Wrapf(code.DeviceConnectionError,
				"harmony device %s not attached", harmonyDevice.Serial())
		}
	}

	device = &HarmonyDevice{
		Options: deviceConfig,
		Device:  harmonyDevice,
	}
	log.Info().Str("connectKey", device.Options.ConnectKey).Msg("init harmony device")
	return device, nil
}

func (dev *HarmonyDevice) Setup() error {
	return nil
}

func (dev *HarmonyDevice) Teardown() error {
	return nil
}

func (dev *HarmonyDevice) UUID() string {
	return dev.Options.ConnectKey
}

func (dev *HarmonyDevice) LogEnabled() bool {
	return dev.Options.LogOn
}

func (dev *HarmonyDevice) Install(appPath string, opts ...option.InstallOption) error {
	return nil
}

func (dev *HarmonyDevice) Uninstall(packageName string) error {
	return nil
}

func (dev *HarmonyDevice) GetPackageInfo(packageName string) (types.AppInfo, error) {
	log.Warn().Msg("get package info not implemented for harmony device, skip")
	return types.AppInfo{}, nil
}

func (dev *HarmonyDevice) NewDriver() (IDriverExt, error) {
	// init harmony driver
	driver, err := NewHDCDriver(dev)
	if err != nil {
		return nil, errors.Wrap(err, "init harmony driver failed")
	}
	driverExt, err := NewDriverExt(driver, ai.WithCVService(ai.CVServiceTypeVEDEM))
	if err != nil {
		return nil, errors.Wrap(err, "init harmony driver ext failed")
	}

	// setup driver
	if err := driverExt.GetDriver().Setup(); err != nil {
		return nil, err
	}

	return driverExt, nil
}
