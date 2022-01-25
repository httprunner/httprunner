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
func NewStep(name string) *StepRequest {
	return &StepRequest{
		step: &TStep{
			Name:      name,
			Variables: make(map[string]interface{}),
		},
	}
}

type StepRequest struct {
	step *TStep
}

// WithVariables sets variables for current teststep.
func (s *StepRequest) WithVariables(variables map[string]interface{}) *StepRequest {
	s.step.Variables = variables
	return s
}

// SetupHook adds a setup hook for current teststep.
func (s *StepRequest) SetupHook(hook string) *StepRequest {
	s.step.SetupHooks = append(s.step.SetupHooks, hook)
	return s
}

// GET makes a HTTP GET request.
func (s *StepRequest) GET(url string) *StepRequestWithOptionalArgs {
	s.step.Request = &Request{
		Method: httpGET,
		URL:    url,
	}
	return &StepRequestWithOptionalArgs{
		step: s.step,
	}
}

// HEAD makes a HTTP HEAD request.
func (s *StepRequest) HEAD(url string) *StepRequestWithOptionalArgs {
	s.step.Request = &Request{
		Method: httpHEAD,
		URL:    url,
	}
	return &StepRequestWithOptionalArgs{
		step: s.step,
	}
}

// POST makes a HTTP POST request.
func (s *StepRequest) POST(url string) *StepRequestWithOptionalArgs {
	s.step.Request = &Request{
		Method: httpPOST,
		URL:    url,
	}
	return &StepRequestWithOptionalArgs{
		step: s.step,
	}
}

// PUT makes a HTTP PUT request.
func (s *StepRequest) PUT(url string) *StepRequestWithOptionalArgs {
	s.step.Request = &Request{
		Method: httpPUT,
		URL:    url,
	}
	return &StepRequestWithOptionalArgs{
		step: s.step,
	}
}

// DELETE makes a HTTP DELETE request.
func (s *StepRequest) DELETE(url string) *StepRequestWithOptionalArgs {
	s.step.Request = &Request{
		Method: httpDELETE,
		URL:    url,
	}
	return &StepRequestWithOptionalArgs{
		step: s.step,
	}
}

// OPTIONS makes a HTTP OPTIONS request.
func (s *StepRequest) OPTIONS(url string) *StepRequestWithOptionalArgs {
	s.step.Request = &Request{
		Method: httpOPTIONS,
		URL:    url,
	}
	return &StepRequestWithOptionalArgs{
		step: s.step,
	}
}

// PATCH makes a HTTP PATCH request.
func (s *StepRequest) PATCH(url string) *StepRequestWithOptionalArgs {
	s.step.Request = &Request{
		Method: httpPATCH,
		URL:    url,
	}
	return &StepRequestWithOptionalArgs{
		step: s.step,
	}
}

// CallRefCase calls a referenced testcase.
func (s *StepRequest) CallRefCase(tc *TestCase) *StepTestCaseWithOptionalArgs {
	s.step.TestCase = tc
	return &StepTestCaseWithOptionalArgs{
		step: s.step,
	}
}

// StartTransaction starts a transaction.
func (s *StepRequest) StartTransaction(name string) *StepTransaction {
	s.step.Transaction = &Transaction{
		Name: name,
		Type: transactionStart,
	}
	return &StepTransaction{
		step: s.step,
	}
}

// EndTransaction ends a transaction.
func (s *StepRequest) EndTransaction(name string) *StepTransaction {
	s.step.Transaction = &Transaction{
		Name: name,
		Type: transactionEnd,
	}
	return &StepTransaction{
		step: s.step,
	}
}

// StepRequestWithOptionalArgs implements IStep interface.
type StepRequestWithOptionalArgs struct {
	step *TStep
}

// SetVerify sets whether to verify SSL for current HTTP request.
func (s *StepRequestWithOptionalArgs) SetVerify(verify bool) *StepRequestWithOptionalArgs {
	s.step.Request.Verify = verify
	return s
}

// SetTimeout sets timeout for current HTTP request.
func (s *StepRequestWithOptionalArgs) SetTimeout(timeout float32) *StepRequestWithOptionalArgs {
	s.step.Request.Timeout = timeout
	return s
}

// SetProxies sets proxies for current HTTP request.
func (s *StepRequestWithOptionalArgs) SetProxies(proxies map[string]string) *StepRequestWithOptionalArgs {
	// TODO
	return s
}

// SetAllowRedirects sets whether to allow redirects for current HTTP request.
func (s *StepRequestWithOptionalArgs) SetAllowRedirects(allowRedirects bool) *StepRequestWithOptionalArgs {
	s.step.Request.AllowRedirects = allowRedirects
	return s
}

// SetAuth sets auth for current HTTP request.
func (s *StepRequestWithOptionalArgs) SetAuth(auth map[string]string) *StepRequestWithOptionalArgs {
	// TODO
	return s
}

// WithParams sets HTTP request params for current step.
func (s *StepRequestWithOptionalArgs) WithParams(params map[string]interface{}) *StepRequestWithOptionalArgs {
	s.step.Request.Params = params
	return s
}

// WithHeaders sets HTTP request headers for current step.
func (s *StepRequestWithOptionalArgs) WithHeaders(headers map[string]string) *StepRequestWithOptionalArgs {
	s.step.Request.Headers = headers
	return s
}

// WithCookies sets HTTP request cookies for current step.
func (s *StepRequestWithOptionalArgs) WithCookies(cookies map[string]string) *StepRequestWithOptionalArgs {
	s.step.Request.Cookies = cookies
	return s
}

// WithBody sets HTTP request body for current step.
func (s *StepRequestWithOptionalArgs) WithBody(body interface{}) *StepRequestWithOptionalArgs {
	s.step.Request.Body = body
	return s
}

// TeardownHook adds a teardown hook for current teststep.
func (s *StepRequestWithOptionalArgs) TeardownHook(hook string) *StepRequestWithOptionalArgs {
	s.step.TeardownHooks = append(s.step.TeardownHooks, hook)
	return s
}

// Validate switches to step validation.
func (s *StepRequestWithOptionalArgs) Validate() *StepRequestValidation {
	return &StepRequestValidation{
		step: s.step,
	}
}

// Extract switches to step extraction.
func (s *StepRequestWithOptionalArgs) Extract() *StepRequestExtraction {
	s.step.Extract = make(map[string]string)
	return &StepRequestExtraction{
		step: s.step,
	}
}

func (s *StepRequestWithOptionalArgs) Name() string {
	if s.step.Name != "" {
		return s.step.Name
	}
	return fmt.Sprintf("%s %s", s.step.Request.Method, s.step.Request.URL)
}

func (s *StepRequestWithOptionalArgs) Type() string {
	return fmt.Sprintf("request-%v", s.step.Request.Method)
}

func (s *StepRequestWithOptionalArgs) ToStruct() *TStep {
	return s.step
}

// StepTestCaseWithOptionalArgs implements IStep interface.
type StepTestCaseWithOptionalArgs struct {
	step *TStep
}

// TeardownHook adds a teardown hook for current teststep.
func (s *StepTestCaseWithOptionalArgs) TeardownHook(hook string) *StepTestCaseWithOptionalArgs {
	s.step.TeardownHooks = append(s.step.TeardownHooks, hook)
	return s
}

// Export specifies variable names to export from referenced testcase for current step.
func (s *StepTestCaseWithOptionalArgs) Export(names ...string) *StepTestCaseWithOptionalArgs {
	s.step.Export = append(s.step.Export, names...)
	return s
}

func (s *StepTestCaseWithOptionalArgs) Name() string {
	if s.step.Name != "" {
		return s.step.Name
	}
	return s.step.TestCase.Config.Name
}

func (s *StepTestCaseWithOptionalArgs) Type() string {
	return "testcase"
}

func (s *StepTestCaseWithOptionalArgs) ToStruct() *TStep {
	return s.step
}

// StepTransaction implements IStep interface.
type StepTransaction struct {
	step *TStep
}

func (s *StepTransaction) Name() string {
	if s.step.Name != "" {
		return s.step.Name
	}
	return fmt.Sprintf("transaction %s %s", s.step.Transaction.Name, s.step.Transaction.Type)
}

func (s *StepTransaction) Type() string {
	return "transaction"
}

func (s *StepTransaction) ToStruct() *TStep {
	return s.step
}

// StepRendezvous implements IStep interface.
type StepRendezvous struct {
	step *TStep
}

func (s *StepRendezvous) Name() string {
	if s.step.Name != "" {
		return s.step.Name
	}
	return s.step.Rendezvous.Name
}

func (s *StepRendezvous) Type() string {
	return "rendezvous"
}

func (s *StepRendezvous) ToStruct() *TStep {
	return s.step
}

// Rendezvous creates a new rendezvous
func (s *StepRequest) Rendezvous(name string) *StepRendezvous {
	s.step.Rendezvous = &Rendezvous{
		Name: name,
	}
	return &StepRendezvous{
		step: s.step,
	}
}

// WithUserNumber sets the user number needed to release the current rendezvous
func (s *StepRendezvous) WithUserNumber(number int64) *StepRendezvous {
	s.step.Rendezvous.Number = number
	return s
}

// WithUserPercent sets the user percent needed to release the current rendezvous
func (s *StepRendezvous) WithUserPercent(percent float32) *StepRendezvous {
	s.step.Rendezvous.Percent = percent
	return s
}

// WithTimeout sets the timeout of duration between each user arriving at the current rendezvous
func (s *StepRendezvous) WithTimeout(timeout int64) *StepRendezvous {
	s.step.Rendezvous.Timeout = timeout
	return s
}
