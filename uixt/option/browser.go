package option

func NewBrowserDeviceOptions(opts ...BrowserDeviceOption) *BrowserDeviceOptions {
	config := &BrowserDeviceOptions{}
	for _, opt := range opts {
		opt(config)
	}
	return config
}

type BrowserDeviceOptions struct {
	BrowserID   string `json:"browser_id,omitempty" yaml:"browser_id,omitempty"`
	LogOn       bool   `json:"log_on,omitempty" yaml:"log_on,omitempty"`
	IgnorePopup bool   `json:"ignore_popup,omitempty" yaml:"ignore_popup,omitempty"` // keep for compatibility
	Width       int    `json:"width,omitempty" yaml:"width,omitempty"`
	Height      int    `json:"height,omitempty" yaml:"height,omitempty"`
}

func (dev *BrowserDeviceOptions) Options() (deviceOptions []BrowserDeviceOption) {
	if dev.BrowserID != "" {
		deviceOptions = append(deviceOptions, WithBrowserID(dev.BrowserID))
	}
	if dev.LogOn {
		deviceOptions = append(deviceOptions, WithBrowserLogOn(true))
	}
	if dev.Width > 0 && dev.Height > 0 {
		deviceOptions = append(deviceOptions, WithBrowserPageSize(dev.Width, dev.Height))
	}
	return
}

type BrowserDeviceOption func(*BrowserDeviceOptions)

func WithBrowserID(serial string) BrowserDeviceOption {
	return func(device *BrowserDeviceOptions) {
		device.BrowserID = serial
	}
}

func WithBrowserLogOn(logOn bool) BrowserDeviceOption {
	return func(device *BrowserDeviceOptions) {
		device.LogOn = logOn
	}
}

func WithBrowserPageSize(width, height int) BrowserDeviceOption {
	return func(device *BrowserDeviceOptions) {
		device.Width = width
		device.Height = height
	}
}
