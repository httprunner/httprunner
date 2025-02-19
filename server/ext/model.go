package server_ext

type AppInstallRequest struct {
	AppUrl             string `json:"appUrl" binding:"required"`
	MappingUrl         string `json:"mappingUrl"`
	ResourceMappingUrl string `json:"resourceMappingUrl"`
	PackageName        string `json:"packageName"`
}

type LoginRequest struct {
	PackageName string `json:"packageName"`
	PhoneNumber string `json:"phoneNumber"`
	Captcha     string `json:"captcha" binding:"required_without=Password"`
	Password    string `json:"password" binding:"required_without=Captcha"`
}

type LogoutRequest struct {
	PackageName string `json:"packageName"`
}

type HttpResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"msg"`
	Result  interface{} `json:"result,omitempty"`
}
