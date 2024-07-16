package server

// 常见的错误代码和消息
const (
	InternalServerErrorCode = 100001
	InternalServerErrorMsg  = "Invalid Server Error"

	InvalidParamErrorCode = 100002
	InvalidParamErrorMsg  = "Invalid %s Param"

	CodeNotFound = 1004
	MsgNotFound  = "Resource not found"
)

const (
	DeviceNotFoundCode = 110001
	DeviceNotFoundMsg  = "Device %s Not Found"
)
