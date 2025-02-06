package option

type AndroidDeviceConfig struct {
	SerialNumber string `json:"serial,omitempty" yaml:"serial,omitempty"`
	STUB         bool   `json:"stub,omitempty" yaml:"stub,omitempty"`           // use stub
	UIA2         bool   `json:"uia2,omitempty" yaml:"uia2,omitempty"`           // use uiautomator2
	UIA2IP       string `json:"uia2_ip,omitempty" yaml:"uia2_ip,omitempty"`     // uiautomator2 server ip
	UIA2Port     int    `json:"uia2_port,omitempty" yaml:"uia2_port,omitempty"` // uiautomator2 server port
	LogOn        bool   `json:"log_on,omitempty" yaml:"log_on,omitempty"`
}

func (dev *AndroidDeviceConfig) Options() (deviceOptions []AndroidDeviceOption) {
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

func NewAndroidDeviceConfig(opts ...AndroidDeviceOption) *AndroidDeviceConfig {
	config := &AndroidDeviceConfig{}
	for _, opt := range opts {
		opt(config)
	}
	return config
}

type AndroidDeviceOption func(*AndroidDeviceConfig)

func WithSerialNumber(serial string) AndroidDeviceOption {
	return func(device *AndroidDeviceConfig) {
		device.SerialNumber = serial
	}
}

func WithUIA2(uia2On bool) AndroidDeviceOption {
	return func(device *AndroidDeviceConfig) {
		device.UIA2 = uia2On
	}
}

func WithStub(stubOn bool) AndroidDeviceOption {
	return func(device *AndroidDeviceConfig) {
		device.STUB = stubOn
	}
}

func WithUIA2IP(ip string) AndroidDeviceOption {
	return func(device *AndroidDeviceConfig) {
		device.UIA2IP = ip
	}
}

func WithUIA2Port(port int) AndroidDeviceOption {
	return func(device *AndroidDeviceConfig) {
		device.UIA2Port = port
	}
}

func WithAdbLogOn(logOn bool) AndroidDeviceOption {
	return func(device *AndroidDeviceConfig) {
		device.LogOn = logOn
	}
}
