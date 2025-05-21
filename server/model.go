package server

import (
	"github.com/httprunner/httprunner/v5/uixt/option"
)

type uploadRequest struct {
	X          float64 `json:"x"`
	Y          float64 `json:"y"`
	FileUrl    string  `json:"file_url"`
	FileFormat string  `json:"file_format"`
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
	Width   int `json:"width"`
	Height  int `json:"height"`
}
