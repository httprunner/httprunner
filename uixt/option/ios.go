package option

type IOSDeviceOptions struct {
	UDID         string `json:"udid,omitempty" yaml:"udid,omitempty"`
	Wireless     bool   `json:"wireless,omitempty" yaml:"wireless,omitempty"`
	WDAPort      int    `json:"port,omitempty" yaml:"port,omitempty"`             // WDA remote port
	WDAMjpegPort int    `json:"mjpeg_port,omitempty" yaml:"mjpeg_port,omitempty"` // WDA remote MJPEG port
	LogOn        bool   `json:"log_on,omitempty" yaml:"log_on,omitempty"`
	IgnorePopup  bool   `json:"ignore_popup,omitempty" yaml:"ignore_popup,omitempty"` // keep for compatibility

	// switch to iOS springboard before init WDA session
	ResetHomeOnStartup bool `json:"reset_home_on_startup,omitempty" yaml:"reset_home_on_startup,omitempty"`

	// config appium settings
	SnapshotMaxDepth           int    `json:"snapshot_max_depth,omitempty" yaml:"snapshot_max_depth,omitempty"`
	AcceptAlertButtonSelector  string `json:"accept_alert_button_selector,omitempty" yaml:"accept_alert_button_selector,omitempty"`
	DismissAlertButtonSelector string `json:"dismiss_alert_button_selector,omitempty" yaml:"dismiss_alert_button_selector,omitempty"`
}

func (dev *IOSDeviceOptions) Options() (deviceOptions []IOSDeviceOption) {
	if dev.UDID != "" {
		deviceOptions = append(deviceOptions, WithUDID(dev.UDID))
	}
	if dev.Wireless {
		deviceOptions = append(deviceOptions, WithWireless(true))
	}
	if dev.WDAPort != 0 {
		deviceOptions = append(deviceOptions, WithWDAPort(dev.WDAPort))
	}
	if dev.WDAMjpegPort != 0 {
		deviceOptions = append(deviceOptions, WithWDAMjpegPort(dev.WDAMjpegPort))
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

const (
	defaultWDAPort   = 8700
	defaultMjpegPort = 8800
)

func NewIOSDeviceOptions(opts ...IOSDeviceOption) *IOSDeviceOptions {
	config := &IOSDeviceOptions{}
	for _, opt := range opts {
		opt(config)
	}

	if config.WDAPort == 0 {
		config.WDAPort = defaultWDAPort
	}
	if config.WDAMjpegPort == 0 {
		config.WDAMjpegPort = defaultMjpegPort
	}

	return config
}

type IOSDeviceOption func(*IOSDeviceOptions)

func WithUDID(udid string) IOSDeviceOption {
	return func(device *IOSDeviceOptions) {
		device.UDID = udid
	}
}

func WithWireless(on bool) IOSDeviceOption {
	return func(device *IOSDeviceOptions) {
		device.Wireless = on
	}
}

func WithWDAPort(port int) IOSDeviceOption {
	return func(device *IOSDeviceOptions) {
		device.WDAPort = port
	}
}

func WithWDAMjpegPort(port int) IOSDeviceOption {
	return func(device *IOSDeviceOptions) {
		device.WDAMjpegPort = port
	}
}

func WithWDALogOn(logOn bool) IOSDeviceOption {
	return func(device *IOSDeviceOptions) {
		device.LogOn = logOn
	}
}

func WithResetHomeOnStartup(reset bool) IOSDeviceOption {
	return func(device *IOSDeviceOptions) {
		device.ResetHomeOnStartup = reset
	}
}

func WithSnapshotMaxDepth(depth int) IOSDeviceOption {
	return func(device *IOSDeviceOptions) {
		device.SnapshotMaxDepth = depth
	}
}

func WithAcceptAlertButtonSelector(selector string) IOSDeviceOption {
	return func(device *IOSDeviceOptions) {
		device.AcceptAlertButtonSelector = selector
	}
}

func WithDismissAlertButtonSelector(selector string) IOSDeviceOption {
	return func(device *IOSDeviceOptions) {
		device.DismissAlertButtonSelector = selector
	}
}
