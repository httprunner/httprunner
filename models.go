package hrp

const (
	httpGET     string = "GET"
	httpHEAD    string = "HEAD"
	httpPOST    string = "POST"
	httpPUT     string = "PUT"
	httpDELETE  string = "DELETE"
	httpOPTIONS string = "OPTIONS"
	httpPATCH   string = "PATCH"
)

// TConfig represents config data structure for testcase.
// Each testcase should contain one config part.
type TConfig struct {
	Name       string                 `json:"name" yaml:"name"` // required
	Verify     bool                   `json:"verify,omitempty" yaml:"verify,omitempty"`
	BaseURL    string                 `json:"base_url,omitempty" yaml:"base_url,omitempty"`
	Variables  map[string]interface{} `json:"variables,omitempty" yaml:"variables,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Export     []string               `json:"export,omitempty" yaml:"export,omitempty"`
	Weight     int                    `json:"weight,omitempty" yaml:"weight,omitempty"`
}

// Request represents HTTP request data structure.
// This is used for teststep.
type Request struct {
	Method         string                 `json:"method" yaml:"method"` // required
	URL            string                 `json:"url" yaml:"url"`       // required
	Params         map[string]interface{} `json:"params,omitempty" yaml:"params,omitempty"`
	Headers        map[string]string      `json:"headers,omitempty" yaml:"headers,omitempty"`
	Cookies        map[string]string      `json:"cookies,omitempty" yaml:"cookies,omitempty"`
	Body           interface{}            `json:"body,omitempty" yaml:"body,omitempty"`
	Timeout        float32                `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	AllowRedirects bool                   `json:"allow_redirects,omitempty" yaml:"allow_redirects,omitempty"`
	Verify         bool                   `json:"verify,omitempty" yaml:"verify,omitempty"`
}

// Validator represents validator for one HTTP response.
type Validator struct {
	Check   string      `json:"check" yaml:"check"` // get value with jmespath
	Assert  string      `json:"assert" yaml:"assert"`
	Expect  interface{} `json:"expect" yaml:"expect"`
	Message string      `json:"msg,omitempty" yaml:"msg,omitempty"` // optional
}

// TStep represents teststep data structure.
// Each step maybe two different type: make one HTTP request or reference another testcase.
type TStep struct {
	Name          string                 `json:"name" yaml:"name"` // required
	Request       *Request               `json:"request,omitempty" yaml:"request,omitempty"`
	TestCase      *TestCase              `json:"testcase,omitempty" yaml:"testcase,omitempty"`
	Variables     map[string]interface{} `json:"variables,omitempty" yaml:"variables,omitempty"`
	SetupHooks    []string               `json:"setup_hooks,omitempty" yaml:"setup_hooks,omitempty"`
	TeardownHooks []string               `json:"teardown_hooks,omitempty" yaml:"teardown_hooks,omitempty"`
	Extract       map[string]string      `json:"extract,omitempty" yaml:"extract,omitempty"`
	Validators    []Validator            `json:"validate,omitempty" yaml:"validate,omitempty"`
	Export        []string               `json:"export,omitempty" yaml:"export,omitempty"`
}

// TCase represents testcase data structure.
// Each testcase includes one public config and several sequential teststeps.
type TCase struct {
	Config    *TConfig `json:"config" yaml:"config"`
	TestSteps []*TStep `json:"teststeps" yaml:"teststeps"`
}

// IStep represents interface for all types for teststeps.
type IStep interface {
	Name() string
	Type() string
	ToStruct() *TStep
}

// ITestCase represents interface for all types for testcases.
type ITestCase interface {
	ToTestCase() (*TestCase, error)
	ToTCase() (*TCase, error)
}

// TestCase is a container for one testcase.
// used for testcase runner
type TestCase struct {
	Config    *TConfig
	TestSteps []IStep
}

func (tc *TestCase) ToTestCase() (*TestCase, error) {
	return tc, nil
}

type TestCasePath struct {
	Path string
}

type testCaseSummary struct{}

type stepData struct {
	name           string                 // step name
	success        bool                   // step execution result
	responseLength int64                  // response body length
	exportVars     map[string]interface{} // extract variables
}
