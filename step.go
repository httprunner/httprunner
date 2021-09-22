package httpboomer

import "fmt"

func Step(name string) *step {
	return &step{
		runner: defaultRunner,
		TStep: &TStep{
			Name:      name,
			Request:   &TRequest{},
			Variables: make(Variables),
		},
	}
}

type step struct {
	runner *Runner
	*TStep
}

func (s *step) WithRunner(runner *Runner) *step {
	s.runner = runner
	return s
}

func (s *step) WithVariables(variables Variables) *step {
	s.TStep.Variables = variables
	return s
}

func (s *step) SetupHook(hook string) *step {
	s.TStep.SetupHooks = append(s.TStep.SetupHooks, hook)
	return s
}

func (s *step) GET(url string) *requestWithOptionalArgs {
	s.TStep.Request.Method = GET
	s.TStep.Request.URL = url
	return &requestWithOptionalArgs{
		runner: s.runner,
		step:   s.TStep,
	}
}

func (s *step) HEAD(url string) *requestWithOptionalArgs {
	s.TStep.Request.Method = HEAD
	s.TStep.Request.URL = url
	return &requestWithOptionalArgs{
		runner: s.runner,
		step:   s.TStep,
	}
}

func (s *step) POST(url string) *requestWithOptionalArgs {
	s.TStep.Request.Method = POST
	s.TStep.Request.URL = url
	return &requestWithOptionalArgs{
		runner: s.runner,
		step:   s.TStep,
	}
}

func (s *step) PUT(url string) *requestWithOptionalArgs {
	s.TStep.Request.Method = PUT
	s.TStep.Request.URL = url
	return &requestWithOptionalArgs{
		runner: s.runner,
		step:   s.TStep,
	}
}

func (s *step) DELETE(url string) *requestWithOptionalArgs {
	s.TStep.Request.Method = DELETE
	s.TStep.Request.URL = url
	return &requestWithOptionalArgs{
		runner: s.runner,
		step:   s.TStep,
	}
}

func (s *step) OPTIONS(url string) *requestWithOptionalArgs {
	s.TStep.Request.Method = OPTIONS
	s.TStep.Request.URL = url
	return &requestWithOptionalArgs{
		runner: s.runner,
		step:   s.TStep,
	}
}

func (s *step) PATCH(url string) *requestWithOptionalArgs {
	s.TStep.Request.Method = PATCH
	s.TStep.Request.URL = url
	return &requestWithOptionalArgs{
		runner: s.runner,
		step:   s.TStep,
	}
}

// call referenced testcase
func (s *step) CallRefCase(tc *TestCase) *testcaseWithOptionalArgs {
	s.TStep.TestCase = tc
	return &testcaseWithOptionalArgs{
		runner: s.runner,
		step:   s.TStep,
	}
}

// implements IStep interface
type requestWithOptionalArgs struct {
	runner *Runner
	step   *TStep
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

func (s *requestWithOptionalArgs) WithParams(params Params) *requestWithOptionalArgs {
	s.step.Request.Params = params
	return s
}

func (s *requestWithOptionalArgs) WithHeaders(headers Headers) *requestWithOptionalArgs {
	s.step.Request.Headers = headers
	return s
}

func (s *requestWithOptionalArgs) WithCookies(cookies Cookies) *requestWithOptionalArgs {
	s.step.Request.Cookies = cookies
	return s
}

func (s *requestWithOptionalArgs) WithData(data interface{}) *requestWithOptionalArgs {
	s.step.Request.Data = data
	return s
}

func (s *requestWithOptionalArgs) WithJSON(json interface{}) *requestWithOptionalArgs {
	s.step.Request.JSON = json
	return s
}

func (s *requestWithOptionalArgs) TeardownHook(hook string) *requestWithOptionalArgs {
	s.step.TeardownHooks = append(s.step.TeardownHooks, hook)
	return s
}

func (s *requestWithOptionalArgs) Validate() *stepRequestValidation {
	return &stepRequestValidation{
		runner: s.runner,
		step:   s.step,
	}
}

func (s *requestWithOptionalArgs) Extract() *stepRequestExtraction {
	return &stepRequestExtraction{
		runner: s.runner,
		step:   s.step,
	}
}

func (s *requestWithOptionalArgs) Name() string {
	return s.step.Name
}

func (s *requestWithOptionalArgs) Type() string {
	return fmt.Sprintf("request-%v", s.step.Request.Method)
}

func (s *requestWithOptionalArgs) Run() error {
	return s.runner.runStep(s.step)
}

// implements IStep interface
type testcaseWithOptionalArgs struct {
	runner *Runner
	step   *TStep
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
	return s.step.Name
}

func (s *testcaseWithOptionalArgs) Type() string {
	return "testcase"
}

func (s *testcaseWithOptionalArgs) Run() error {
	return s.runner.runCase(s.step.TestCase)
}
