package types

type TargetDeviceRequest struct {
	Platform string `json:"platform" binding:"required" desc:"Device platform: android/ios/browser"`
	Serial   string `json:"serial" binding:"required" desc:"Device serial/udid/browser id"`
}

type TapRequest struct {
	TargetDeviceRequest
	X        float64 `json:"x" binding:"required" desc:"X coordinate (0.0~1.0 for percent, or absolute pixel value)"`
	Y        float64 `json:"y" binding:"required" desc:"Y coordinate (0.0~1.0 for percent, or absolute pixel value)"`
	Duration float64 `json:"duration" desc:"Tap duration in seconds (optional)"`
}

type DragRequest struct {
	TargetDeviceRequest
	FromX         float64 `json:"from_x" binding:"required" desc:"Starting X-coordinate (percentage, 0.0 to 1.0)"`
	FromY         float64 `json:"from_y" binding:"required" desc:"Starting Y-coordinate (percentage, 0.0 to 1.0)"`
	ToX           float64 `json:"to_x" binding:"required" desc:"Ending X-coordinate (percentage, 0.0 to 1.0)"`
	ToY           float64 `json:"to_y" binding:"required" desc:"Ending Y-coordinate (percentage, 0.0 to 1.0)"`
	Duration      float64 `json:"duration" desc:"Swipe duration in milliseconds (optional)"`
	PressDuration float64 `json:"press_duration" desc:"Press duration in milliseconds (optional)"`
}

type SwipeRequest struct {
	TargetDeviceRequest
	Direction     string  `json:"direction" binding:"required" desc:"The direction of the swipe. Supported directions: up, down, left, right"`
	Duration      float64 `json:"duration" desc:"Swipe duration in milliseconds (optional)"`
	PressDuration float64 `json:"press_duration" desc:"Press duration in milliseconds (optional)"`
}

type InputRequest struct {
	TargetDeviceRequest
	Text      string `json:"text" binding:"required"`
	Frequency int    `json:"frequency"` // only iOS
}

type DeleteRequest struct {
	TargetDeviceRequest
	Count int `json:"count" binding:"required"`
}

type KeycodeRequest struct {
	TargetDeviceRequest
	Keycode int `json:"keycode" binding:"required"`
}

type AppInstallRequest struct {
	TargetDeviceRequest
	AppUrl             string `json:"appUrl" binding:"required"`
	MappingUrl         string `json:"mappingUrl"`
	ResourceMappingUrl string `json:"resourceMappingUrl"`
	PackageName        string `json:"packageName"`
}

type AppInfoRequest struct {
	TargetDeviceRequest
	PackageName string `form:"packageName" binding:"required"`
}

type AppUninstallRequest struct {
	TargetDeviceRequest
	PackageName string `json:"packageName" binding:"required"`
}

type AppClearRequest struct {
	TargetDeviceRequest
	PackageName string `json:"packageName" binding:"required"`
}

type AppLaunchRequest struct {
	TargetDeviceRequest
	PackageName string `json:"packageName" binding:"required" desc:"The package name of the app to launch"`
}

type AppTerminateRequest struct {
	TargetDeviceRequest
	PackageName string `json:"packageName" binding:"required" desc:"The package name of the app to terminate"`
}

type PressButtonRequest struct {
	TargetDeviceRequest
	Button DeviceButton `json:"button" binding:"required" desc:"The button to press. Supported buttons: BACK (android only), HOME, VOLUME_UP, VOLUME_DOWN, ENTER."`
}
