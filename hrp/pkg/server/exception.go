package server

// 常见的错误代码和消息
const (
	InternalServerErrorCode = 100001
	InternalServerErrorMsg  = "Internal Server Error"

	InvalidParamErrorCode = 100002
	InvalidParamErrorMsg  = "Invalid %s Param"
)

const (
	DeviceNotFoundCode = 110001
	DeviceNotFoundMsg  = "Device %s Not Found"
)
