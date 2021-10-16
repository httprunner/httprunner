package hrp

import "fmt"

func Step(name string) *step {
	return &step{
		TStep: &TStep{
			Name:      name,
			Variables: make(map[string]interface{}),
		},
	}
}

type step struct {
	*TStep
}

func (s *step) WithVariables(variables map[string]interface{}) *step {
	s.TStep.Variables = variables
	return s
}

func (s *step) SetupHook(hook string) *step {
	s.TStep.SetupHooks = append(s.TStep.SetupHooks, hook)
	return s
}

func (s *step) GET(url string) *requestWithOptionalArgs {
	s.TStep.Request = &TRequest{
		Method: GET,
		URL:    url,
	}
	return &requestWithOptionalArgs{
		step: s.TStep,
	}
}

func (s *step) HEAD(url string) *requestWithOptionalArgs {
	s.TStep.Request = &TRequest{
		Method: HEAD,
		URL:    url,
	}
	return &requestWithOptionalArgs{
		step: s.TStep,
	}
}

func (s *step) POST(url string) *requestWithOptionalArgs {
	s.TStep.Request = &TRequest{
		Method: POST,
		URL:    url,
	}
	return &requestWithOptionalArgs{
		step: s.TStep,
	}
}

func (s *step) PUT(url string) *requestWithOptionalArgs {
	s.TStep.Request = &TRequest{
		Method: PUT,
		URL:    url,
	}
	return &requestWithOptionalArgs{
		step: s.TStep,
	}
}

func (s *step) DELETE(url string) *requestWithOptionalArgs {
	s.TStep.Request = &TRequest{
		Method: DELETE,
		URL:    url,
	}
	return &requestWithOptionalArgs{
		step: s.TStep,
	}
}

func (s *step) OPTIONS(url string) *requestWithOptionalArgs {
	s.TStep.Request = &TRequest{
		Method: OPTIONS,
		URL:    url,
	}
	return &requestWithOptionalArgs{
		step: s.TStep,
	}
}

func (s *step) PATCH(url string) *requestWithOptionalArgs {
	s.TStep.Request = &TRequest{
		Method: PATCH,
		URL:    url,
	}
	return &requestWithOptionalArgs{
		step: s.TStep,
	}
}

// call referenced testcase
func (s *step) CallRefCase(tc *TestCase) *testcaseWithOptionalArgs {
	s.TStep.TestCase = tc
	return &testcaseWithOptionalArgs{
		step: s.TStep,
	}
}

// implements IStep interface
type requestWithOptionalArgs struct {
	step *TStep
}

func (s *requestWithOptionalArgs) SetVerify(verify bool) *requestWithOptionalArgs {
	s.step.Request.Verify = verify
	return s
}

func (s *requestWithOptionalArgs) SetTimeout(timeout float32) *requestWithOptionalArgs {
	s.step.Request.Timeout = timeout
	return s
}

func (s *requestWithOptionalArgs) SetProxies(proxies map[string]string) *requestWithOptionalArgs {
	// TODO
	return s
}

func (s *requestWithOptionalArgs) SetAllowRedirects(allowRedirects bool) *requestWithOptionalArgs {
	s.step.Request.AllowRedirects = allowRedirects
	return s
}

func (s *requestWithOptionalArgs) SetAuth(auth map[string]string) *requestWithOptionalArgs {
	// TODO
	return s
}

func (s *requestWithOptionalArgs) WithParams(params map[string]interface{}) *requestWithOptionalArgs {
	s.step.Request.Params = params
	return s
}

func (s *requestWithOptionalArgs) WithHeaders(headers map[string]string) *requestWithOptionalArgs {
	s.step.Request.Headers = headers
	return s
}

func (s *requestWithOptionalArgs) WithCookies(cookies map[string]string) *requestWithOptionalArgs {
	s.step.Request.Cookies = cookies
	return s
}

func (s *requestWithOptionalArgs) WithBody(body interface{}) *requestWithOptionalArgs {
	s.step.Request.Body = body
	return s
}

func (s *requestWithOptionalArgs) TeardownHook(hook string) *requestWithOptionalArgs {
	s.step.TeardownHooks = append(s.step.TeardownHooks, hook)
	return s
}

func (s *requestWithOptionalArgs) Validate() *stepRequestValidation {
	return &stepRequestValidation{
		step: s.step,
	}
}

func (s *requestWithOptionalArgs) Extract() *stepRequestExtraction {
	s.step.Extract = make(map[string]string)
	return &stepRequestExtraction{
		step: s.step,
	}
}

func (s *requestWithOptionalArgs) Name() string {
	if s.step.Name != "" {
		return s.step.Name
	}
	return fmt.Sprintf("%s %s", s.step.Request.Method, s.step.Request.URL)
}

func (s *requestWithOptionalArgs) Type() string {
	return fmt.Sprintf("request-%v", s.step.Request.Method)
}

func (s *requestWithOptionalArgs) ToStruct() *TStep {
	return s.step
}

// implements IStep interface
type testcaseWithOptionalArgs struct {
	step *TStep
}

func (s *testcaseWithOptionalArgs) TeardownHook(hook string) *testcaseWithOptionalArgs {
	s.step.TeardownHooks = append(s.step.TeardownHooks, hook)
	return s
}

func (s *testcaseWithOptionalArgs) Export(names ...string) *testcaseWithOptionalArgs {
	s.step.Export = append(s.step.Export, names...)
	return s
}

func (s *testcaseWithOptionalArgs) Name() string {
	if s.step.Name != "" {
		return s.step.Name
	}
	return s.step.TestCase.Config.Name
}

func (s *testcaseWithOptionalArgs) Type() string {
	return "testcase"
}

func (s *testcaseWithOptionalArgs) ToStruct() *TStep {
	return s.step
}
