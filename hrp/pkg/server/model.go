package server

type HttpResponse struct {
	Result    interface{} `json:"result,omitempty"`
	ErrorCode int         `json:"errorCode"`
	ErrorMsg  string      `json:"errorMsg"`
}

type TapRequest struct {
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Duration float64 `json:"duration"`
}

type DragRequest struct {
	FromX    float64 `json:"from_x"`
	FromY    float64 `json:"from_y"`
	ToX      float64 `json:"to_x"`
	ToY      float64 `json:"to_y"`
	Duration float64 `json:"duration"`
}

type InputRequest struct {
	Text      string `json:"text"`
	Frequency int    `json:"frequency"` // only iOS
}

type KeycodeRequest struct {
	Keycode int `json:"keycode"`
}

type AppLaunchRequest struct {
	PackageName string `json:"packageName"`
}

type AppTerminalRequest struct {
	PackageName string `json:"packageName"`
}

type LoginRequest struct {
	PackageName string `json:"packageName"`
	PhoneNumber string `json:"phoneNumber"`
	Captcha     string `json:"captcha"`
}
