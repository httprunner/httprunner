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

type PushMediaRequest struct {
	ImageUrl string `json:"imageUrl" binding:"required_without=VideoUrl"`
	VideoUrl string `json:"videoUrl" binding:"required_without=ImageUrl"`
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
