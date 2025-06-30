package server

type HttpResponse struct {
	Code    int         `json:"errorCode"`
	Message string      `json:"message"`
	Result  interface{} `json:"result,omitempty"`
}
