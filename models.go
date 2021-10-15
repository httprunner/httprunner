package httpboomer

type EnumHTTPMethod string

const (
	GET     EnumHTTPMethod = "GET"
	HEAD    EnumHTTPMethod = "HEAD"
	POST    EnumHTTPMethod = "POST"
	PUT     EnumHTTPMethod = "PUT"
	DELETE  EnumHTTPMethod = "DELETE"
	OPTIONS EnumHTTPMethod = "OPTIONS"
	PATCH   EnumHTTPMethod = "PATCH"
)

type TConfig struct {
	Name       string                 `json:"name" yaml:"name"`
	Verify     bool                   `json:"verify,omitempty" yaml:"verify,omitempty"`
	BaseURL    string                 `json:"base_url,omitempty" yaml:"base_url,omitempty"`
	Variables  map[string]interface{} `json:"variables,omitempty" yaml:"variables,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Export     []string               `json:"export,omitempty" yaml:"export,omitempty"`
	Weight     int                    `json:"weight,omitempty" yaml:"weight,omitempty"`
}

type TRequest struct {
	Method         EnumHTTPMethod         `json:"method" yaml:"method"`
	URL            string                 `json:"url" yaml:"url"`
	Params         map[string]interface{} `json:"params,omitempty" yaml:"params,omitempty"`
	Headers        map[string]string      `json:"headers,omitempty" yaml:"headers,omitempty"`
	Cookies        map[string]string      `json:"cookies,omitempty" yaml:"cookies,omitempty"`
	Body           interface{}            `json:"body,omitempty" yaml:"body,omitempty"`
	Timeout        float32                `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	AllowRedirects bool                   `json:"allow_redirects,omitempty" yaml:"allow_redirects,omitempty"`
	Verify         bool                   `json:"verify,omitempty" yaml:"verify,omitempty"`
}

type TValidator struct {
	Check   string      `json:"check,omitempty" yaml:"check,omitempty"` // get value with jmespath
	Assert  string      `json:"assert,omitempty" yaml:"assert,omitempty"`
	Expect  interface{} `json:"expect,omitempty" yaml:"expect,omitempty"`
	Message string      `json:"msg,omitempty" yaml:"msg,omitempty"`
}

type TStep struct {
	Name          string                 `json:"name" yaml:"name"`
	Request       *TRequest              `json:"request,omitempty" yaml:"request,omitempty"`
	TestCase      *TestCase              `json:"testcase,omitempty" yaml:"testcase,omitempty"`
	Variables     map[string]interface{} `json:"variables,omitempty" yaml:"variables,omitempty"`
	SetupHooks    []string               `json:"setup_hooks,omitempty" yaml:"setup_hooks,omitempty"`
	TeardownHooks []string               `json:"teardown_hooks,omitempty" yaml:"teardown_hooks,omitempty"`
	Extract       map[string]string      `json:"extract,omitempty" yaml:"extract,omitempty"`
	Validators    []TValidator           `json:"validate,omitempty" yaml:"validate,omitempty"`
	Export        []string               `json:"export,omitempty" yaml:"export,omitempty"`
}

// used for testcase json loading and dumping
type TCase struct {
	Config    TConfig  `json:"config" yaml:"config"`
	TestSteps []*TStep `json:"teststeps" yaml:"teststeps"`
}

// interface for all types of steps
type IStep interface {
	Name() string
	Type() string
	ToStruct() *TStep
}

type ITestCase interface {
	ToTestCase() (*TestCase, error)
	ToTCase() (*TCase, error)
}

// used for testcase runner
type TestCase struct {
	Config    TConfig
	TestSteps []IStep
}

func (tc *TestCase) ToTestCase() (*TestCase, error) {
	return tc, nil
}

type TestCasePath struct {
	Path string
}

type TestCaseSummary struct{}

type StepData struct {
	Name       string                 // step name
	Success    bool                   // step execution result
	ExportVars map[string]interface{} // extract variables
}
