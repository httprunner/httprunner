package option

type HarmonyDeviceOptions struct {
	ConnectKey string `json:"connect_key,omitempty" yaml:"connect_key,omitempty"`
	LogOn      bool   `json:"log_on,omitempty" yaml:"log_on,omitempty"`
}

func (dev *HarmonyDeviceOptions) Options() (deviceOptions []HarmonyDeviceOption) {
	if dev.ConnectKey != "" {
		deviceOptions = append(deviceOptions, WithConnectKey(dev.ConnectKey))
	}
	if dev.LogOn {
		deviceOptions = append(deviceOptions, WithLogOn(true))
	}
	return
}

func NewHarmonyDeviceOptions(opts ...HarmonyDeviceOption) (device *HarmonyDeviceOptions) {
	device = &HarmonyDeviceOptions{}
	for _, option := range opts {
		option(device)
	}
	return
}

type HarmonyDeviceOption func(*HarmonyDeviceOptions)

func WithConnectKey(connectKey string) HarmonyDeviceOption {
	return func(device *HarmonyDeviceOptions) {
		device.ConnectKey = connectKey
	}
}

func WithLogOn(logOn bool) HarmonyDeviceOption {
	return func(device *HarmonyDeviceOptions) {
		device.LogOn = logOn
	}
}
