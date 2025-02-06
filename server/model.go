package server

import "github.com/httprunner/httprunner/v5/pkg/uixt/option"

type HttpResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"msg"`
	Result  interface{} `json:"result,omitempty"`
}

type TapRequest struct {
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	Text string  `json:"text"`

	Options *option.ActionOptions `json:"options,omitempty"`
}

type DragRequest struct {
	FromX float64 `json:"from_x"`
	FromY float64 `json:"from_y"`
	ToX   float64 `json:"to_x"`
	ToY   float64 `json:"to_y"`

	Options *option.ActionOptions `json:"options,omitempty"`
}

type InputRequest struct {
	Text      string `json:"text"`
	Frequency int    `json:"frequency"` // only iOS
}

type ScreenRequest struct {
	Options *option.ActionOptions `json:"options,omitempty"`
}

type KeycodeRequest struct {
	Keycode int `json:"keycode"`
}

type AppClearRequest struct {
	PackageName string `json:"packageName"`
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
	Password    string `json:"password"`
}

type LogoutRequest struct {
	PackageName string `json:"packageName"`
}
