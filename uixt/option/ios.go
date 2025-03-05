package option

type IOSDeviceOptions struct {
	UDID         string `json:"udid,omitempty" yaml:"udid,omitempty"`
	WDAPort      int    `json:"port,omitempty" yaml:"port,omitempty"`             // WDA remote port
	WDAMjpegPort int    `json:"mjpeg_port,omitempty" yaml:"mjpeg_port,omitempty"` // WDA remote MJPEG port
	LogOn        bool   `json:"log_on,omitempty" yaml:"log_on,omitempty"`
	LazySetup    bool   `json:"lazy_setup,omitempty" yaml:"lazy_setup,omitempty"` // lazy setup WDA

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
	if dev.WDAPort != 0 {
		deviceOptions = append(deviceOptions, WithWDAPort(dev.WDAPort))
	}
	if dev.WDAMjpegPort != 0 {
		deviceOptions = append(deviceOptions, WithWDAMjpegPort(dev.WDAMjpegPort))
	}
	if dev.LogOn {
		deviceOptions = append(deviceOptions, WithWDALogOn(true))
	}
	if dev.LazySetup {
		deviceOptions = append(deviceOptions, WithLazySetup(true))
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
	defaultWDAPort   = 8100
	defaultMjpegPort = 9100
)

const (
	// Changes the value of maximum depth for traversing elements source tree.
	// It may help to prevent out of memory or timeout errors while getting the elements source tree,
	// but it might restrict the depth of source tree.
	// A part of elements source tree might be lost if the value was too small. Defaults to 50
	defaultSnapshotMaxDepth = 10
	// Allows to customize accept/dismiss alert button selector.
	// It helps you to handle an arbitrary element as accept button in accept alert command.
	// The selector should be a valid class chain expression, where the search root is the alert element itself.
	// The default button location algorithm is used if the provided selector is wrong or does not match any element.
	// e.g. **/XCUIElementTypeButton[`label CONTAINS[c] ‘accept’`]
	acceptAlertButtonSelector  = "**/XCUIElementTypeButton[`label IN {'允许','好','仅在使用应用期间','稍后再说'}`]"
	dismissAlertButtonSelector = "**/XCUIElementTypeButton[`label IN {'不允许','暂不'}`]"
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

	if config.SnapshotMaxDepth == 0 {
		config.SnapshotMaxDepth = defaultSnapshotMaxDepth
	}
	if config.AcceptAlertButtonSelector == "" {
		config.AcceptAlertButtonSelector = acceptAlertButtonSelector
	}
	if config.DismissAlertButtonSelector == "" {
		config.DismissAlertButtonSelector = dismissAlertButtonSelector
	}

	return config
}

type IOSDeviceOption func(*IOSDeviceOptions)

func WithUDID(udid string) IOSDeviceOption {
	return func(device *IOSDeviceOptions) {
		device.UDID = udid
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

func WithLazySetup(lazySetup bool) IOSDeviceOption {
	return func(device *IOSDeviceOptions) {
		device.LazySetup = lazySetup
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
