package hrp

import "fmt"

// NewConfig returns a new constructed testcase config with specified testcase name.
func NewConfig(name string) *Config {
	return &Config{
		cfg: &TConfig{
			Name:      name,
			Variables: make(map[string]interface{}),
		},
	}
}

type Config struct {
	cfg *TConfig
}

// WithVariables sets variables for current testcase.
func (c *Config) WithVariables(variables map[string]interface{}) *Config {
	c.cfg.Variables = variables
	return c
}

// SetBaseURL sets base URL for current testcase.
func (c *Config) SetBaseURL(baseURL string) *Config {
	c.cfg.BaseURL = baseURL
	return c
}

// SetVerifySSL sets whether to verify SSL for current testcase.
func (c *Config) SetVerifySSL(verify bool) *Config {
	c.cfg.Verify = verify
	return c
}

// WithParameters sets parameters for current testcase.
func (c *Config) WithParameters(parameters map[string]interface{}) *Config {
	c.cfg.Parameters = parameters
	return c
}

// ExportVars specifies variable names to export for current testcase.
func (c *Config) ExportVars(vars ...string) *Config {
	c.cfg.Export = vars
	return c
}

// SetWeight sets weight for current testcase, which is used in load testing.
func (c *Config) SetWeight(weight int) *Config {
	c.cfg.Weight = weight
	return c
}

// Name returns config name, this implements IConfig interface.
func (c *Config) Name() string {
	return c.cfg.Name
}

// ToStruct returns *TConfig, this implements IConfig interface.
func (c *Config) ToStruct() *TConfig {
	return c.cfg
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
func (s *TStep) GET(url string) *stepRequestWithOptionalArgs {
	s.Request = &Request{
		Method: httpGET,
		URL:    url,
	}
	return &stepRequestWithOptionalArgs{
		step: s,
	}
}

// HEAD makes a HTTP HEAD request.
func (s *TStep) HEAD(url string) *stepRequestWithOptionalArgs {
	s.Request = &Request{
		Method: httpHEAD,
		URL:    url,
	}
	return &stepRequestWithOptionalArgs{
		step: s,
	}
}

// POST makes a HTTP POST request.
func (s *TStep) POST(url string) *stepRequestWithOptionalArgs {
	s.Request = &Request{
		Method: httpPOST,
		URL:    url,
	}
	return &stepRequestWithOptionalArgs{
		step: s,
	}
}

// PUT makes a HTTP PUT request.
func (s *TStep) PUT(url string) *stepRequestWithOptionalArgs {
	s.Request = &Request{
		Method: httpPUT,
		URL:    url,
	}
	return &stepRequestWithOptionalArgs{
		step: s,
	}
}

// DELETE makes a HTTP DELETE request.
func (s *TStep) DELETE(url string) *stepRequestWithOptionalArgs {
	s.Request = &Request{
		Method: httpDELETE,
		URL:    url,
	}
	return &stepRequestWithOptionalArgs{
		step: s,
	}
}

// OPTIONS makes a HTTP OPTIONS request.
func (s *TStep) OPTIONS(url string) *stepRequestWithOptionalArgs {
	s.Request = &Request{
		Method: httpOPTIONS,
		URL:    url,
	}
	return &stepRequestWithOptionalArgs{
		step: s,
	}
}

// PATCH makes a HTTP PATCH request.
func (s *TStep) PATCH(url string) *stepRequestWithOptionalArgs {
	s.Request = &Request{
		Method: httpPATCH,
		URL:    url,
	}
	return &stepRequestWithOptionalArgs{
		step: s,
	}
}

// CallRefCase calls a referenced testcase.
func (s *TStep) CallRefCase(tc *TestCase) *stepTestCaseWithOptionalArgs {
	s.TestCase = tc
	return &stepTestCaseWithOptionalArgs{
		step: s,
	}
}

// implements IStep interface
type stepRequestWithOptionalArgs struct {
	step *TStep
}

// SetVerify sets whether to verify SSL for current HTTP request.
func (s *stepRequestWithOptionalArgs) SetVerify(verify bool) *stepRequestWithOptionalArgs {
	s.step.Request.Verify = verify
	return s
}

// SetTimeout sets timeout for current HTTP request.
func (s *stepRequestWithOptionalArgs) SetTimeout(timeout float32) *stepRequestWithOptionalArgs {
	s.step.Request.Timeout = timeout
	return s
}

// SetProxies sets proxies for current HTTP request.
func (s *stepRequestWithOptionalArgs) SetProxies(proxies map[string]string) *stepRequestWithOptionalArgs {
	// TODO
	return s
}

// SetAllowRedirects sets whether to allow redirects for current HTTP request.
func (s *stepRequestWithOptionalArgs) SetAllowRedirects(allowRedirects bool) *stepRequestWithOptionalArgs {
	s.step.Request.AllowRedirects = allowRedirects
	return s
}

// SetAuth sets auth for current HTTP request.
func (s *stepRequestWithOptionalArgs) SetAuth(auth map[string]string) *stepRequestWithOptionalArgs {
	// TODO
	return s
}

// WithParams sets HTTP request params for current step.
func (s *stepRequestWithOptionalArgs) WithParams(params map[string]interface{}) *stepRequestWithOptionalArgs {
	s.step.Request.Params = params
	return s
}

// WithHeaders sets HTTP request headers for current step.
func (s *stepRequestWithOptionalArgs) WithHeaders(headers map[string]string) *stepRequestWithOptionalArgs {
	s.step.Request.Headers = headers
	return s
}

// WithCookies sets HTTP request cookies for current step.
func (s *stepRequestWithOptionalArgs) WithCookies(cookies map[string]string) *stepRequestWithOptionalArgs {
	s.step.Request.Cookies = cookies
	return s
}

// WithBody sets HTTP request body for current step.
func (s *stepRequestWithOptionalArgs) WithBody(body interface{}) *stepRequestWithOptionalArgs {
	s.step.Request.Body = body
	return s
}

// TeardownHook adds a teardown hook for current teststep.
func (s *stepRequestWithOptionalArgs) TeardownHook(hook string) *stepRequestWithOptionalArgs {
	s.step.TeardownHooks = append(s.step.TeardownHooks, hook)
	return s
}

// Validate switches to step validation.
func (s *stepRequestWithOptionalArgs) Validate() *stepRequestValidation {
	return &stepRequestValidation{
		step: s.step,
	}
}

// Extract switches to step extraction.
func (s *stepRequestWithOptionalArgs) Extract() *stepRequestExtraction {
	s.step.Extract = make(map[string]string)
	return &stepRequestExtraction{
		step: s.step,
	}
}

func (s *stepRequestWithOptionalArgs) Name() string {
	if s.step.Name != "" {
		return s.step.Name
	}
	return fmt.Sprintf("%s %s", s.step.Request.Method, s.step.Request.URL)
}

func (s *stepRequestWithOptionalArgs) Type() string {
	return fmt.Sprintf("request-%v", s.step.Request.Method)
}

func (s *stepRequestWithOptionalArgs) ToStruct() *TStep {
	return s.step
}

// implements IStep interface
type stepTestCaseWithOptionalArgs struct {
	step *TStep
}

// TeardownHook adds a teardown hook for current teststep.
func (s *stepTestCaseWithOptionalArgs) TeardownHook(hook string) *stepTestCaseWithOptionalArgs {
	s.step.TeardownHooks = append(s.step.TeardownHooks, hook)
	return s
}

// Export specifies variable names to export from referenced testcase for current step.
func (s *stepTestCaseWithOptionalArgs) Export(names ...string) *stepTestCaseWithOptionalArgs {
	s.step.Export = append(s.step.Export, names...)
	return s
}

func (s *stepTestCaseWithOptionalArgs) Name() string {
	if s.step.Name != "" {
		return s.step.Name
	}
	return s.step.TestCase.Config.Name()
}

func (s *stepTestCaseWithOptionalArgs) Type() string {
	return "testcase"
}

func (s *stepTestCaseWithOptionalArgs) ToStruct() *TStep {
	return s.step
}

// implements IStep interface
type stepTransaction struct {
	step *TStep
}

func (s *stepTransaction) Name() string {
	if s.step.Name != "" {
		return s.step.Name
	}
	return fmt.Sprintf("transaction %s %s", s.step.Transaction.Name, s.step.Transaction.Type)
}

func (s *stepTransaction) Type() string {
	return "transaction"
}

func (s *stepTransaction) ToStruct() *TStep {
	return s.step
}

// implements IStep interface
type stepRendezvous struct {
	step *TStep
}

func (s *stepRendezvous) Name() string {
	if s.step.Name != "" {
		return s.step.Name
	}
	return s.step.Rendezvous.Name
}

func (s *stepRendezvous) Type() string {
	return "rendezvous"
}

func (s *stepRendezvous) ToStruct() *TStep {
	return s.step
}
