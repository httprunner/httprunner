package option

import "github.com/httprunner/httprunner/v5/pkg/gadb"

type AndroidDeviceOptions struct {
	SerialNumber string `json:"serial,omitempty" yaml:"serial,omitempty"`
	LogOn        bool   `json:"log_on,omitempty" yaml:"log_on,omitempty"`
	IgnorePopup  bool   `json:"ignore_popup,omitempty" yaml:"ignore_popup,omitempty"` // keep for compatibility

	// adb
	AdbServerHost string `json:"adb_server_host,omitempty" yaml:"adb_server_host,omitempty"`
	AdbServerPort int    `json:"adb_server_port,omitempty" yaml:"adb_server_port,omitempty"`

	// uiautomator2
	UIA2                      bool   `json:"uia2,omitempty" yaml:"uia2,omitempty"`           // use uiautomator2
	UIA2IP                    string `json:"uia2_ip,omitempty" yaml:"uia2_ip,omitempty"`     // uiautomator2 server ip
	UIA2Port                  int    `json:"uia2_port,omitempty" yaml:"uia2_port,omitempty"` // uiautomator2 server port
	UIA2ServerPackageName     string `json:"uia2_server_package_name,omitempty" yaml:"uia2_server_package_name,omitempty"`
	UIA2ServerTestPackageName string `json:"uia2_server_test_package_name,omitempty" yaml:"uia2_server_test_package_name,omitempty"`
}

func (dev *AndroidDeviceOptions) Options() (deviceOptions []AndroidDeviceOption) {
	if dev.SerialNumber != "" {
		deviceOptions = append(deviceOptions, WithSerialNumber(dev.SerialNumber))
	}
	if dev.UIA2 {
		deviceOptions = append(deviceOptions, WithUIA2(true))
	}
	if dev.UIA2IP != "" {
		deviceOptions = append(deviceOptions, WithUIA2IP(dev.UIA2IP))
	}
	if dev.UIA2Port != 0 {
		deviceOptions = append(deviceOptions, WithUIA2Port(dev.UIA2Port))
	}
	if dev.LogOn {
		deviceOptions = append(deviceOptions, WithAdbLogOn(true))
	}
	return
}

const (
	// adb server
	defaultAdbServerHost = "localhost"
	defaultAdbServerPort = gadb.AdbServerPort // 5037

	// uiautomator2 server
	defaultUIA2ServerHost            = "localhost"
	defaultUIA2ServerPort            = 6790
	defaultUIA2ServerPackageName     = "io.appium.uiautomator2.server"
	defaultUIA2ServerTestPackageName = "io.appium.uiautomator2.server.test"

	AdbKeyBoardPackageName = "com.android.adbkeyboard/.AdbIME"
	UnicodeImePackageName  = "io.appium.settings/.UnicodeIME"
)

func NewAndroidDeviceOptions(opts ...AndroidDeviceOption) *AndroidDeviceOptions {
	config := &AndroidDeviceOptions{}
	for _, opt := range opts {
		opt(config)
	}

	// adb default
	if config.AdbServerHost == "" {
		config.AdbServerHost = defaultAdbServerHost
	}
	if config.AdbServerPort == 0 {
		config.AdbServerPort = defaultAdbServerPort
	}

	// uiautomator2 default
	if config.UIA2IP == "" && config.UIA2Port == 0 {
		config.UIA2IP = defaultUIA2ServerHost
		config.UIA2Port = defaultUIA2ServerPort
	}
	if config.UIA2ServerPackageName == "" {
		config.UIA2ServerPackageName = defaultUIA2ServerPackageName
	}
	if config.UIA2ServerTestPackageName == "" {
		config.UIA2ServerTestPackageName = defaultUIA2ServerTestPackageName
	}

	return config
}

type AndroidDeviceOption func(*AndroidDeviceOptions)

func WithSerialNumber(serial string) AndroidDeviceOption {
	return func(device *AndroidDeviceOptions) {
		device.SerialNumber = serial
	}
}

func WithUIA2(uia2On bool) AndroidDeviceOption {
	return func(device *AndroidDeviceOptions) {
		device.UIA2 = uia2On
	}
}

func WithUIA2IP(ip string) AndroidDeviceOption {
	return func(device *AndroidDeviceOptions) {
		device.UIA2IP = ip
	}
}

func WithUIA2Port(port int) AndroidDeviceOption {
	return func(device *AndroidDeviceOptions) {
		device.UIA2Port = port
	}
}

func WithAdbLogOn(logOn bool) AndroidDeviceOption {
	return func(device *AndroidDeviceOptions) {
		device.LogOn = logOn
	}
}
