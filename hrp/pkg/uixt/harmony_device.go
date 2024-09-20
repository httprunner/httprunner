package uixt

import (
	"fmt"

	"code.byted.org/iesqa/ghdc"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/code"
)

var (
	HdcServerHost = "localhost"
	HdcServerPort = ghdc.HdcServerPort // 5037
)

type HarmonyDevice struct {
	d           *ghdc.Device
	ConnectKey  string `json:"connect_key,omitempty" yaml:"connect_key,omitempty"`
	IgnorePopup bool   `json:"ignore_popup,omitempty" yaml:"ignore_popup,omitempty"`
	LogOn       bool   `json:"log_on,omitempty" yaml:"log_on,omitempty"`
}

type HarmonyDeviceOption func(*HarmonyDevice)

func WithConnectKey(connectKey string) HarmonyDeviceOption {
	return func(device *HarmonyDevice) {
		device.ConnectKey = connectKey
	}
}

func WithIgnorePopup(ignorePopup bool) HarmonyDeviceOption {
	return func(device *HarmonyDevice) {
		device.IgnorePopup = ignorePopup
	}
}

func WithLogOn(logOn bool) HarmonyDeviceOption {
	return func(device *HarmonyDevice) {
		device.LogOn = logOn
	}
}

func NewHarmonyDevice(options ...HarmonyDeviceOption) (device *HarmonyDevice, err error) {
	device = &HarmonyDevice{}
	for _, option := range options {
		option(device)
	}

	deviceList, err := GetHarmonyDevices(device.ConnectKey)
	if err != nil {
		return nil, errors.Wrap(code.DeviceConnectionError, err.Error())
	}

	if device.ConnectKey == "" && len(deviceList) > 1 {
		return nil, errors.Wrap(code.DeviceConnectionError, "more than one device connected, please specify the serial")
	}

	dev := deviceList[0]

	if device.ConnectKey == "" {
		selectSerial := dev.Serial()
		device.ConnectKey = selectSerial
		log.Warn().
			Str("connectKey", device.ConnectKey).
			Msg("harmony ConnectKey is not specified, select the first one")
	}

	device.d = dev
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

func (dev *HarmonyDevice) NewDriver(options ...DriverOption) (driverExt *DriverExt, err error) {
	driver, err := newHarmonyDriver(dev.d)
	if err != nil {
		log.Error().Err(err).Msg("failed to new harmony driver")
		return nil, err
	}

	driverExt, err = newDriverExt(dev, driver, options...)
	if err != nil {
		return nil, err
	}

	return driverExt, nil
}

func (dev *HarmonyDevice) NewUSBDriver(options ...DriverOption) (driver IWebDriver, err error) {
	harmonyDriver, err := newHarmonyDriver(dev.d)
	if err != nil {
		log.Error().Err(err).Msg("failed to new harmony driver")
		return nil, err
	}

	return harmonyDriver, nil
}

func (dev *HarmonyDevice) StartPerf() error {
	return nil
}

func (dev *HarmonyDevice) StopPerf() string {
	return ""
}

func (dev *HarmonyDevice) StartPcap() error {
	return nil
}

func (dev *HarmonyDevice) StopPcap() string {
	return ""
}

func (dev *HarmonyDevice) Install(appPath string, opts *InstallOptions) error {
	return nil
}

func (dev *HarmonyDevice) Uninstall(packageName string) error {
	return nil
}

func (dev *HarmonyDevice) GetPackageInfo(packageName string) (AppInfo, error) {
	return AppInfo{}, nil
}
