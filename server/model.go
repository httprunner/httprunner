package server

import (
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
)

type TapRequest struct {
	X        float64               `json:"x" binding:"required"`
	Y        float64               `json:"y" binding:"required"`
	Duration float64               `json:"duration"`
	Options  *option.ActionOptions `json:"options,omitempty"`
}
type uploadRequest struct {
	X          float64 `json:"x"`
	Y          float64 `json:"y"`
	FileUrl    string  `json:"file_url"`
	FileFormat string  `json:"file_format"`
}
type DragRequest struct {
	FromX         float64               `json:"from_x" binding:"required"`
	FromY         float64               `json:"from_y" binding:"required"`
	ToX           float64               `json:"to_x" binding:"required"`
	ToY           float64               `json:"to_y" binding:"required"`
	Duration      float64               `json:"duration"`
	PressDuration float64               `json:"press_duration"`
	Options       *option.ActionOptions `json:"options,omitempty"`
}

type InputRequest struct {
	Text      string `json:"text" binding:"required"`
	Frequency int    `json:"frequency"` // only iOS
}

type DeleteRequest struct {
	Count int `json:"count" binding:"required"`
}

type KeycodeRequest struct {
	Keycode int `json:"keycode" binding:"required"`
}

type AppClearRequest struct {
	PackageName string `json:"packageName" binding:"required"`
}

type AppLaunchRequest struct {
	PackageName string `json:"packageName" binding:"required"`
}

type AppTerminalRequest struct {
	PackageName string `json:"packageName" binding:"required"`
}

type AppInstallRequest struct {
	AppUrl             string `json:"appUrl" binding:"required"`
	MappingUrl         string `json:"mappingUrl"`
	ResourceMappingUrl string `json:"resourceMappingUrl"`
	PackageName        string `json:"packageName"`
}

type AppInfoRequest struct {
	PackageName string `form:"packageName" binding:"required"`
}

type AppUninstallRequest struct {
	PackageName string `json:"packageName" binding:"required"`
}

type PushMediaRequest struct {
	ImageUrl string `json:"imageUrl" binding:"required_without=VideoUrl"`
	VideoUrl string `json:"videoUrl" binding:"required_without=ImageUrl"`
}

type OperateRequest struct {
	StepText string `json:"stepText" binding:"required"`
}

type HttpResponse struct {
	Code    int         `json:"errorCode"`
	Message string      `json:"errorMsg"`
	Result  interface{} `json:"result,omitempty"`
}

type ScreenRequest struct {
	Options *option.ScreenOptions `json:"options,omitempty"`
}

type UploadRequest struct {
	X          float64 `json:"x"`
	Y          float64 `json:"y"`
	FileUrl    string  `json:"file_url"`
	FileFormat string  `json:"file_format"`
}

type HoverRequest struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type ScrollRequest struct {
	Delta int `json:"delta"`
}

type CreateBrowserRequest struct {
	Timeout int `json:"timeout"`
}
