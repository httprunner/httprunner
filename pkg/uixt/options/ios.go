package options

type IOSDeviceConfig struct {
	UDID      string `json:"udid,omitempty" yaml:"udid,omitempty"`
	Port      int    `json:"port,omitempty" yaml:"port,omitempty"`             // WDA remote port
	MjpegPort int    `json:"mjpeg_port,omitempty" yaml:"mjpeg_port,omitempty"` // WDA remote MJPEG port
	STUB      bool   `json:"stub,omitempty" yaml:"stub,omitempty"`             // use stub
	LogOn     bool   `json:"log_on,omitempty" yaml:"log_on,omitempty"`

	// switch to iOS springboard before init WDA session
	ResetHomeOnStartup bool `json:"reset_home_on_startup,omitempty" yaml:"reset_home_on_startup,omitempty"`

	// config appium settings
	SnapshotMaxDepth           int    `json:"snapshot_max_depth,omitempty" yaml:"snapshot_max_depth,omitempty"`
	AcceptAlertButtonSelector  string `json:"accept_alert_button_selector,omitempty" yaml:"accept_alert_button_selector,omitempty"`
	DismissAlertButtonSelector string `json:"dismiss_alert_button_selector,omitempty" yaml:"dismiss_alert_button_selector,omitempty"`
}

func (dev *IOSDeviceConfig) Options() (deviceOptions []IOSDeviceOption) {
	if dev.UDID != "" {
		deviceOptions = append(deviceOptions, WithUDID(dev.UDID))
	}
	if dev.Port != 0 {
		deviceOptions = append(deviceOptions, WithWDAPort(dev.Port))
	}
	if dev.MjpegPort != 0 {
		deviceOptions = append(deviceOptions, WithWDAMjpegPort(dev.MjpegPort))
	}
	if dev.STUB {
		deviceOptions = append(deviceOptions, WithIOSStub(true))
	}
	if dev.LogOn {
		deviceOptions = append(deviceOptions, WithWDALogOn(true))
	}
	if dev.ResetHomeOnStartup {
		deviceOptions = append(deviceOptions, WithResetHomeOnStartup(true))
	}
	if dev.SnapshotMaxDepth != 0 {
		deviceOptions = append(deviceOptions, WithSnapshotMaxDepth(dev.SnapshotMaxDepth))
	}
	if dev.AcceptAlertButtonSelector != "" {
		deviceOptions = append(deviceOptions, WithAcceptAlertButtonSelector(dev.AcceptAlertButtonSelector))
	}
	if dev.DismissAlertButtonSelector != "" {
		deviceOptions = append(deviceOptions, WithDismissAlertButtonSelector(dev.DismissAlertButtonSelector))
	}
	return
}

func NewIOSDeviceConfig(opts ...IOSDeviceOption) *IOSDeviceConfig {
	config := &IOSDeviceConfig{}
	for _, opt := range opts {
		opt(config)
	}
	return config
}

type IOSDeviceOption func(*IOSDeviceConfig)

func WithUDID(udid string) IOSDeviceOption {
	return func(device *IOSDeviceConfig) {
		device.UDID = udid
	}
}

func WithWDAPort(port int) IOSDeviceOption {
	return func(device *IOSDeviceConfig) {
		device.Port = port
	}
}

func WithWDAMjpegPort(port int) IOSDeviceOption {
	return func(device *IOSDeviceConfig) {
		device.MjpegPort = port
	}
}

func WithWDALogOn(logOn bool) IOSDeviceOption {
	return func(device *IOSDeviceConfig) {
		device.LogOn = logOn
	}
}

func WithIOSStub(stub bool) IOSDeviceOption {
	return func(device *IOSDeviceConfig) {
		device.STUB = stub
	}
}

func WithResetHomeOnStartup(reset bool) IOSDeviceOption {
	return func(device *IOSDeviceConfig) {
		device.ResetHomeOnStartup = reset
	}
}

func WithSnapshotMaxDepth(depth int) IOSDeviceOption {
	return func(device *IOSDeviceConfig) {
		device.SnapshotMaxDepth = depth
	}
}

func WithAcceptAlertButtonSelector(selector string) IOSDeviceOption {
	return func(device *IOSDeviceConfig) {
		device.AcceptAlertButtonSelector = selector
	}
}

func WithDismissAlertButtonSelector(selector string) IOSDeviceOption {
	return func(device *IOSDeviceConfig) {
		device.DismissAlertButtonSelector = selector
	}
}
