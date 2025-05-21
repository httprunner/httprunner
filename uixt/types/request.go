package types

type TapRequest struct {
	X        float64 `json:"x" binding:"required" desc:"X coordinate (0.0~1.0 for percent, or absolute pixel value)"`
	Y        float64 `json:"y" binding:"required" desc:"Y coordinate (0.0~1.0 for percent, or absolute pixel value)"`
	Duration float64 `json:"duration" desc:"Tap duration in seconds (optional)"`
}

type DragRequest struct {
	FromX         float64 `json:"from_x" binding:"required" desc:"Starting X-coordinate (percentage, 0.0 to 1.0)"`
	FromY         float64 `json:"from_y" binding:"required" desc:"Starting Y-coordinate (percentage, 0.0 to 1.0)"`
	ToX           float64 `json:"to_x" binding:"required" desc:"Ending X-coordinate (percentage, 0.0 to 1.0)"`
	ToY           float64 `json:"to_y" binding:"required" desc:"Ending Y-coordinate (percentage, 0.0 to 1.0)"`
	Duration      float64 `json:"duration" desc:"Swipe duration in milliseconds (optional)"`
	PressDuration float64 `json:"press_duration" desc:"Press duration in milliseconds (optional)"`
}

type SwipeRequest struct {
	Direction string `json:"direction" binding:"required" desc:"The direction of the swipe. Supported directions: up, down, left, right"`
}

type AppClearRequest struct {
	PackageName string `json:"packageName" binding:"required"`
}

type AppLaunchRequest struct {
	PackageName string `json:"packageName" binding:"required" desc:"The package name of the app to launch"`
}

type AppTerminateRequest struct {
	PackageName string `json:"packageName" binding:"required" desc:"The package name of the app to terminate"`
}

type PressButtonRequest struct {
	Button DeviceButton `json:"button" binding:"required" desc:"The button to press. Supported buttons: BACK (android only), HOME, VOLUME_UP, VOLUME_DOWN, ENTER."`
}
