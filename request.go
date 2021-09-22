package httpboomer

func RunRequest(name string) *Request {
	return &Request{
		TStep: &TStep{
			Name:      name,
			Request:   &TRequest{},
			Variables: make(Variables),
		},
	}
}

type Request struct {
	*TStep
}

func (r *Request) WithVariables(variables Variables) *Request {
	r.TStep.Variables = variables
	return r
}

func (r *Request) GET(url string) *RequestWithOptionalArgs {
	r.TStep.Request.Method = GET
	r.TStep.Request.URL = url
	return &RequestWithOptionalArgs{
		TStep: r.TStep,
	}
}

func (r *Request) HEAD(url string) *RequestWithOptionalArgs {
	r.TStep.Request.Method = HEAD
	r.TStep.Request.URL = url
	return &RequestWithOptionalArgs{
		TStep: r.TStep,
	}
}

func (r *Request) POST(url string) *RequestWithOptionalArgs {
	r.TStep.Request.Method = POST
	r.TStep.Request.URL = url
	return &RequestWithOptionalArgs{
		TStep: r.TStep,
	}
}

func (r *Request) PUT(url string) *RequestWithOptionalArgs {
	r.TStep.Request.Method = PUT
	r.TStep.Request.URL = url
	return &RequestWithOptionalArgs{
		TStep: r.TStep,
	}
}

func (r *Request) DELETE(url string) *RequestWithOptionalArgs {
	r.TStep.Request.Method = DELETE
	r.TStep.Request.URL = url
	return &RequestWithOptionalArgs{
		TStep: r.TStep,
	}
}

func (r *Request) OPTIONS(url string) *RequestWithOptionalArgs {
	r.TStep.Request.Method = OPTIONS
	r.TStep.Request.URL = url
	return &RequestWithOptionalArgs{
		TStep: r.TStep,
	}
}

func (r *Request) PATCH(url string) *RequestWithOptionalArgs {
	r.TStep.Request.Method = PATCH
	r.TStep.Request.URL = url
	return &RequestWithOptionalArgs{
		TStep: r.TStep,
	}
}

// implements IStep interface
type RequestWithOptionalArgs struct {
	*TStep
}

func (r *RequestWithOptionalArgs) SetVerify(verify bool) *RequestWithOptionalArgs {
	r.TStep.Request.Verify = verify
	return r
}

func (r *RequestWithOptionalArgs) SetTimeout(timeout float32) *RequestWithOptionalArgs {
	r.TStep.Request.Timeout = timeout
	return r
}

func (r *RequestWithOptionalArgs) SetProxies(proxies map[string]string) *RequestWithOptionalArgs {
	// TODO
	return r
}

func (r *RequestWithOptionalArgs) SetAllowRedirects(allowRedirects bool) *RequestWithOptionalArgs {
	r.TStep.Request.AllowRedirects = allowRedirects
	return r
}

func (r *RequestWithOptionalArgs) SetAuth(auth map[string]string) *RequestWithOptionalArgs {
	// TODO
	return r
}

func (r *RequestWithOptionalArgs) WithParams(params Params) *RequestWithOptionalArgs {
	r.TStep.Request.Params = params
	return r
}

func (r *RequestWithOptionalArgs) WithHeaders(headers Headers) *RequestWithOptionalArgs {
	r.TStep.Request.Headers = headers
	return r
}

func (r *RequestWithOptionalArgs) WithCookies(cookies Cookies) *RequestWithOptionalArgs {
	r.TStep.Request.Cookies = cookies
	return r
}

func (r *RequestWithOptionalArgs) WithData(data interface{}) *RequestWithOptionalArgs {
	r.TStep.Request.Data = data
	return r
}

func (r *RequestWithOptionalArgs) WithJSON(json interface{}) *RequestWithOptionalArgs {
	r.TStep.Request.JSON = json
	return r
}

func (r *RequestWithOptionalArgs) Validate() *StepRequestValidation {
	return &StepRequestValidation{
		TStep: r.TStep,
	}
}

func (r *RequestWithOptionalArgs) ToStruct() *TStep {
	return r.TStep
}

func (r *RequestWithOptionalArgs) Run() error {
	return r.TStep.Run()
}
