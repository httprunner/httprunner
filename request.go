package httpboomer

func RunRequest(name string) *Request {
	return &Request{
		TStep: &TStep{
			Name:      name,
			Variables: make(Variables),
		},
	}
}

type Request struct {
	*TStep
}

func (req *Request) WithVariables(variables Variables) *Request {
	req.TStep.Variables = variables
	return req
}

func (req *Request) GET(url string) *RequestWithOptionalArgs {
	req.TStep.Request.Method = GET
	req.TStep.Request.URL = url
	return &RequestWithOptionalArgs{
		TStep: req.TStep,
	}
}

func (req *Request) HEAD(url string) *RequestWithOptionalArgs {
	req.TStep.Request.Method = HEAD
	req.TStep.Request.URL = url
	return &RequestWithOptionalArgs{
		TStep: req.TStep,
	}
}

func (req *Request) POST(url string) *RequestWithOptionalArgs {
	req.TStep.Request.Method = POST
	req.TStep.Request.URL = url
	return &RequestWithOptionalArgs{
		TStep: req.TStep,
	}
}

func (req *Request) PUT(url string) *RequestWithOptionalArgs {
	req.TStep.Request.Method = PUT
	req.TStep.Request.URL = url
	return &RequestWithOptionalArgs{
		TStep: req.TStep,
	}
}

func (req *Request) DELETE(url string) *RequestWithOptionalArgs {
	req.TStep.Request.Method = DELETE
	req.TStep.Request.URL = url
	return &RequestWithOptionalArgs{
		TStep: req.TStep,
	}
}

func (req *Request) OPTIONS(url string) *RequestWithOptionalArgs {
	req.TStep.Request.Method = OPTIONS
	req.TStep.Request.URL = url
	return &RequestWithOptionalArgs{
		TStep: req.TStep,
	}
}

func (req *Request) PATCH(url string) *RequestWithOptionalArgs {
	req.TStep.Request.Method = PATCH
	req.TStep.Request.URL = url
	return &RequestWithOptionalArgs{
		TStep: req.TStep,
	}
}

func (req *Request) Run() error {
	return req.TStep.Run()
}

// implements IStep interface
type RequestWithOptionalArgs struct {
	*TStep
}

func (req *RequestWithOptionalArgs) SetVerify(verify bool) *RequestWithOptionalArgs {
	req.TStep.Request.Verify = verify
	return req
}

func (req *RequestWithOptionalArgs) SetTimeout(timeout float32) *RequestWithOptionalArgs {
	req.TStep.Request.Timeout = timeout
	return req
}

func (req *RequestWithOptionalArgs) SetProxies(proxies map[string]string) *RequestWithOptionalArgs {
	// TODO
	return req
}

func (req *RequestWithOptionalArgs) SetAllowRedirects(allowRedirects bool) *RequestWithOptionalArgs {
	req.TStep.Request.AllowRedirects = allowRedirects
	return req
}

func (req *RequestWithOptionalArgs) SetAuth(auth map[string]string) *RequestWithOptionalArgs {
	// TODO
	return req
}

func (req *RequestWithOptionalArgs) WithParams(params Params) *RequestWithOptionalArgs {
	req.TStep.Request.Params = params
	return req
}

func (req *RequestWithOptionalArgs) WithHeaders(headers Headers) *RequestWithOptionalArgs {
	req.TStep.Request.Headers = headers
	return req
}

func (req *RequestWithOptionalArgs) WithCookies(cookies Cookies) *RequestWithOptionalArgs {
	req.TStep.Request.Cookies = cookies
	return req
}

func (req *RequestWithOptionalArgs) WithData(data interface{}) *RequestWithOptionalArgs {
	req.TStep.Request.Data = data
	return req
}

func (req *RequestWithOptionalArgs) WithJSON(json interface{}) *RequestWithOptionalArgs {
	req.TStep.Request.JSON = json
	return req
}

func (req *RequestWithOptionalArgs) Validate() *StepRequestValidation {
	return &StepRequestValidation{
		TStep: req.TStep,
	}
}

func (req *RequestWithOptionalArgs) ToStruct() *TStep {
	return req.TStep
}

func (req *RequestWithOptionalArgs) Run() error {
	return req.TStep.Run()
}
