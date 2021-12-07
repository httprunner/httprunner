package hrp

import "fmt"

// NewConfig returns a new constructed testcase config with specified testcase name.
func NewConfig(name string) *TConfig {
	return &TConfig{
		Name:      name,
		Variables: make(map[string]interface{}),
	}
}

// WithVariables sets variables for current testcase.
func (c *TConfig) WithVariables(variables map[string]interface{}) *TConfig {
	c.Variables = variables
	return c
}

// SetBaseURL sets base URL for current testcase.
func (c *TConfig) SetBaseURL(baseURL string) *TConfig {
	c.BaseURL = baseURL
	return c
}

// SetVerifySSL sets whether to verify SSL for current testcase.
func (c *TConfig) SetVerifySSL(verify bool) *TConfig {
	c.Verify = verify
	return c
}

// WithParameters sets parameters for current testcase.
func (c *TConfig) WithParameters(parameters map[string]interface{}) *TConfig {
	c.Parameters = parameters
	return c
}

// ExportVars specifies variable names to export for current testcase.
func (c *TConfig) ExportVars(vars ...string) *TConfig {
	c.Export = vars
	return c
}

// SetWeight sets weight for current testcase, which is used in load testing.
func (c *TConfig) SetWeight(weight int) *TConfig {
	c.Weight = weight
	return c
}

// NewStep returns a new constructed teststep with specified step name.
func NewStep(name string) *TStep {
	return &TStep{
		Name:      name,
		Variables: make(map[string]interface{}),
	}
}

// WithVariables sets variables for current teststep.
func (s *TStep) WithVariables(variables map[string]interface{}) *TStep {
	s.Variables = variables
	return s
}

// SetupHook adds a setup hook for current teststep.
func (s *TStep) SetupHook(hook string) *TStep {
	s.SetupHooks = append(s.SetupHooks, hook)
	return s
}

// GET makes a HTTP GET request.
func (s *TStep) GET(url string) *requestWithOptionalArgs {
	s.Request = &Request{
		Method: httpGET,
		URL:    url,
	}
	return &requestWithOptionalArgs{
		step: s,
	}
}

// HEAD makes a HTTP HEAD request.
func (s *TStep) HEAD(url string) *requestWithOptionalArgs {
	s.Request = &Request{
		Method: httpHEAD,
		URL:    url,
	}
	return &requestWithOptionalArgs{
		step: s,
	}
}

// POST makes a HTTP POST request.
func (s *TStep) POST(url string) *requestWithOptionalArgs {
	s.Request = &Request{
		Method: httpPOST,
		URL:    url,
	}
	return &requestWithOptionalArgs{
		step: s,
	}
}

// PUT makes a HTTP PUT request.
func (s *TStep) PUT(url string) *requestWithOptionalArgs {
	s.Request = &Request{
		Method: httpPUT,
		URL:    url,
	}
	return &requestWithOptionalArgs{
		step: s,
	}
}

// DELETE makes a HTTP DELETE request.
func (s *TStep) DELETE(url string) *requestWithOptionalArgs {
	s.Request = &Request{
		Method: httpDELETE,
		URL:    url,
	}
	return &requestWithOptionalArgs{
		step: s,
	}
}

// OPTIONS makes a HTTP OPTIONS request.
func (s *TStep) OPTIONS(url string) *requestWithOptionalArgs {
	s.Request = &Request{
		Method: httpOPTIONS,
		URL:    url,
	}
	return &requestWithOptionalArgs{
		step: s,
	}
}

// PATCH makes a HTTP PATCH request.
func (s *TStep) PATCH(url string) *requestWithOptionalArgs {
	s.Request = &Request{
		Method: httpPATCH,
		URL:    url,
	}
	return &requestWithOptionalArgs{
		step: s,
	}
}

// CallRefCase calls a referenced testcase.
func (s *TStep) CallRefCase(tc *TestCase) *testcaseWithOptionalArgs {
	s.TestCase = tc
	return &testcaseWithOptionalArgs{
		step: s,
	}
}

// implements IStep interface
type requestWithOptionalArgs struct {
	step *TStep
}

// SetVerify sets whether to verify SSL for current HTTP request.
func (s *requestWithOptionalArgs) SetVerify(verify bool) *requestWithOptionalArgs {
	s.step.Request.Verify = verify
	return s
}

// SetTimeout sets timeout for current HTTP request.
func (s *requestWithOptionalArgs) SetTimeout(timeout float32) *requestWithOptionalArgs {
	s.step.Request.Timeout = timeout
	return s
}

// SetProxies sets proxies for current HTTP request.
func (s *requestWithOptionalArgs) SetProxies(proxies map[string]string) *requestWithOptionalArgs {
	// TODO
	return s
}

// SetAllowRedirects sets whether to allow redirects for current HTTP request.
func (s *requestWithOptionalArgs) SetAllowRedirects(allowRedirects bool) *requestWithOptionalArgs {
	s.step.Request.AllowRedirects = allowRedirects
	return s
}

// SetAuth sets auth for current HTTP request.
func (s *requestWithOptionalArgs) SetAuth(auth map[string]string) *requestWithOptionalArgs {
	// TODO
	return s
}

// WithParams sets HTTP request params for current step.
func (s *requestWithOptionalArgs) WithParams(params map[string]interface{}) *requestWithOptionalArgs {
	s.step.Request.Params = params
	return s
}

// WithHeaders sets HTTP request headers for current step.
func (s *requestWithOptionalArgs) WithHeaders(headers map[string]string) *requestWithOptionalArgs {
	s.step.Request.Headers = headers
	return s
}

// WithCookies sets HTTP request cookies for current step.
func (s *requestWithOptionalArgs) WithCookies(cookies map[string]string) *requestWithOptionalArgs {
	s.step.Request.Cookies = cookies
	return s
}

// WithBody sets HTTP request body for current step.
func (s *requestWithOptionalArgs) WithBody(body interface{}) *requestWithOptionalArgs {
	s.step.Request.Body = body
	return s
}

// TeardownHook adds a teardown hook for current teststep.
func (s *requestWithOptionalArgs) TeardownHook(hook string) *requestWithOptionalArgs {
	s.step.TeardownHooks = append(s.step.TeardownHooks, hook)
	return s
}

// Validate switches to step validation.
func (s *requestWithOptionalArgs) Validate() *stepRequestValidation {
	return &stepRequestValidation{
		step: s.step,
	}
}

// Extract switches to step extraction.
func (s *requestWithOptionalArgs) Extract() *stepRequestExtraction {
	s.step.Extract = make(map[string]string)
	return &stepRequestExtraction{
		step: s.step,
	}
}

func (s *requestWithOptionalArgs) name() string {
	if s.step.Name != "" {
		return s.step.Name
	}
	return fmt.Sprintf("%s %s", s.step.Request.Method, s.step.Request.URL)
}

func (s *requestWithOptionalArgs) getType() string {
	return fmt.Sprintf("request-%v", s.step.Request.Method)
}

func (s *requestWithOptionalArgs) toStruct() *TStep {
	return s.step
}

// implements IStep interface
type testcaseWithOptionalArgs struct {
	step *TStep
}

// TeardownHook adds a teardown hook for current teststep.
func (s *testcaseWithOptionalArgs) TeardownHook(hook string) *testcaseWithOptionalArgs {
	s.step.TeardownHooks = append(s.step.TeardownHooks, hook)
	return s
}

// Export specifies variable names to export from referenced testcase for current step.
func (s *testcaseWithOptionalArgs) Export(names ...string) *testcaseWithOptionalArgs {
	s.step.Export = append(s.step.Export, names...)
	return s
}

func (s *testcaseWithOptionalArgs) name() string {
	if s.step.Name != "" {
		return s.step.Name
	}
	return s.step.TestCase.Config.Name
}

func (s *testcaseWithOptionalArgs) getType() string {
	return "testcase"
}

func (s *testcaseWithOptionalArgs) toStruct() *TStep {
	return s.step
}
