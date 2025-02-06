package uixt

import (
	"fmt"

	"code.byted.org/iesqa/ghdc"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/pkg/uixt/options"
)

var (
	HdcServerHost = "localhost"
	HdcServerPort = ghdc.HdcServerPort // 5037
)

type HarmonyDevice struct {
	*options.HarmonyDeviceConfig
	d *ghdc.Device
}

func NewHarmonyDevice(opts ...options.HarmonyDeviceOption) (device *HarmonyDevice, err error) {
	deviceConfig := options.NewHarmonyDeviceConfig(opts...)

	deviceList, err := GetHarmonyDevices(deviceConfig.ConnectKey)
	if err != nil {
		return nil, errors.Wrap(code.DeviceConnectionError, err.Error())
	}

	if deviceConfig.ConnectKey == "" && len(deviceList) > 1 {
		return nil, errors.Wrap(code.DeviceConnectionError, "more than one device connected, please specify the serial")
	}

	dev := deviceList[0]
	if deviceConfig.ConnectKey == "" {
		selectSerial := dev.Serial()
		deviceConfig.ConnectKey = selectSerial
		log.Warn().
			Str("connectKey", deviceConfig.ConnectKey).
			Msg("harmony ConnectKey is not specified, select the first one")
	}

	device = &HarmonyDevice{
		HarmonyDeviceConfig: deviceConfig,
		d:                   dev,
	}
	log.Info().Str("connectKey", device.ConnectKey).Msg("init harmony device")
	return device, nil
}

func GetHarmonyDevices(serial ...string) (devices []*ghdc.Device, err error) {
	var hdcClient ghdc.Client
	if hdcClient, err = ghdc.NewClientWith(HdcServerHost, HdcServerPort); err != nil {
		return nil, err
	}
	var deviceList []ghdc.Device

	if deviceList, err = hdcClient.DeviceList(); err != nil {
		return nil, err
	}

	// filter by serial
	for _, d := range deviceList {
		for _, s := range serial {
			if s != "" && s != d.Serial() {
				continue
			}
			devices = append(devices, &d)
		}
	}

	if len(devices) == 0 {
		var err error
		if serial == nil || (len(serial) == 1 && serial[0] == "") {
			err = fmt.Errorf("no harmony device found")
		} else {
			err = fmt.Errorf("no harmony device found for serial %v", serial)
		}
		return nil, err
	}
	return devices, nil
}

func (dev *HarmonyDevice) Init() error {
	return nil
}

func (dev *HarmonyDevice) UUID() string {
	return dev.ConnectKey
}

func (dev *HarmonyDevice) LogEnabled() bool {
	return dev.LogOn
}

func (dev *HarmonyDevice) NewDriver(opts ...options.DriverOption) (driverExt *DriverExt, err error) {
	driver, err := newHarmonyDriver(dev.d)
	if err != nil {
		log.Error().Err(err).Msg("failed to new harmony driver")
		return nil, err
	}

	driverExt, err = newDriverExt(dev, driver, opts...)
	if err != nil {
		return nil, err
	}

	return driverExt, nil
}

func (dev *HarmonyDevice) NewUSBDriver(opts ...options.DriverOption) (driver IWebDriver, err error) {
	harmonyDriver, err := newHarmonyDriver(dev.d)
	if err != nil {
		log.Error().Err(err).Msg("failed to new harmony driver")
		return nil, err
	}

	return harmonyDriver, nil
}

func (dev *HarmonyDevice) Install(appPath string, options ...InstallOption) error {
	return nil
}

func (dev *HarmonyDevice) Uninstall(packageName string) error {
	return nil
}

func (dev *HarmonyDevice) GetPackageInfo(packageName string) (AppInfo, error) {
	log.Warn().Msg("get package info not implemented for harmony device, skip")
	return AppInfo{}, nil
}

func (dev *HarmonyDevice) GetCurrentWindow() (WindowInfo, error) {
	return WindowInfo{}, nil
}
