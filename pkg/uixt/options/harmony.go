package options

type HarmonyDeviceConfig struct {
	ConnectKey string `json:"connect_key,omitempty" yaml:"connect_key,omitempty"`
	LogOn      bool   `json:"log_on,omitempty" yaml:"log_on,omitempty"`
}

func (dev *HarmonyDeviceConfig) Options() (deviceOptions []HarmonyDeviceOption) {
	if dev.ConnectKey != "" {
		deviceOptions = append(deviceOptions, WithConnectKey(dev.ConnectKey))
	}
	if dev.LogOn {
		deviceOptions = append(deviceOptions, WithLogOn(true))
	}
	return
}

func NewHarmonyDeviceConfig(options ...HarmonyDeviceOption) (device *HarmonyDeviceConfig) {
	device = &HarmonyDeviceConfig{}
	for _, option := range options {
		option(device)
	}
	return
}

type HarmonyDeviceOption func(*HarmonyDeviceConfig)

func WithConnectKey(connectKey string) HarmonyDeviceOption {
	return func(device *HarmonyDeviceConfig) {
		device.ConnectKey = connectKey
	}
}

func WithLogOn(logOn bool) HarmonyDeviceOption {
	return func(device *HarmonyDeviceConfig) {
		device.LogOn = logOn
	}
}
