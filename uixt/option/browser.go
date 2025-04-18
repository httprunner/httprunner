package option

func NewBrowserDeviceOptions(opts ...BrowserDeviceOption) *BrowserDeviceOptions {
	config := &BrowserDeviceOptions{}
	for _, opt := range opts {
		opt(config)
	}
	return config
}

type BrowserDeviceOptions struct {
	BrowserID string `json:"browser_id,omitempty" yaml:"browser_id,omitempty"`
	LogOn     bool   `json:"log_on,omitempty" yaml:"log_on,omitempty"`
}

type BrowserDeviceOption func(*BrowserDeviceOptions)

func WithBrowserID(serial string) BrowserDeviceOption {
	return func(device *BrowserDeviceOptions) {
		device.BrowserID = serial
	}
}
