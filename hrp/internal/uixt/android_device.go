package uixt

type AndroidDevice struct {
	SerialNumber string `json:"serial,omitempty" yaml:"serial,omitempty"`
	Port         int    `json:"port,omitempty" yaml:"port,omitempty"`
	LogOn        bool   `json:"log_on,omitempty" yaml:"log_on,omitempty"`
}

func (o AndroidDevice) UUID() string {
	return o.SerialNumber
}

func InitUIAClient(device *AndroidDevice) (*DriverExt, error) {
	return nil, nil
}
