package httpboomer

type enumHTTPMethod string

const (
	GET     enumHTTPMethod = "GET"
	HEAD    enumHTTPMethod = "HEAD"
	POST    enumHTTPMethod = "POST"
	PUT     enumHTTPMethod = "PUT"
	DELETE  enumHTTPMethod = "DELETE"
	OPTIONS enumHTTPMethod = "OPTIONS"
	PATCH   enumHTTPMethod = "PATCH"
)

type TConfig struct {
	Name       string                 `json:"name"`
	Verify     bool                   `json:"verify,omitempty"`
	BaseURL    string                 `json:"base_url,omitempty"`
	Variables  map[string]interface{} `json:"variables,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	Export     []string               `json:"export,omitempty"`
	Weight     int                    `json:"weight,omitempty"`
}

type TRequest struct {
	Method         enumHTTPMethod         `json:"method"`
	URL            string                 `json:"url"`
	Params         map[string]interface{} `json:"params,omitempty"`
	Headers        map[string]string      `json:"headers,omitempty"`
	Cookies        map[string]string      `json:"cookies,omitempty"`
	Data           interface{}            `json:"data,omitempty"`
	JSON           interface{}            `json:"json,omitempty"`
	Timeout        float32                `json:"timeout,omitempty"`
	AllowRedirects bool                   `json:"allow_redirects,omitempty"`
	Verify         bool                   `json:"verify,omitempty"`
}

type TValidator struct {
	Check   string      `json:"check,omitempty"` // get value with jmespath
	Assert  string      `json:"assert,omitempty"`
	Expect  interface{} `json:"expect,omitempty"`
	Message string      `json:"msg,omitempty"`
}

type TStep struct {
	Name          string                 `json:"name"`
	Request       *TRequest              `json:"request,omitempty"`
	TestCase      *TestCase              `json:"testcase,omitempty"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	SetupHooks    []string               `json:"setup_hooks,omitempty"`
	TeardownHooks []string               `json:"teardown_hooks,omitempty"`
	Extract       map[string]string      `json:"extract,omitempty"`
	Validators    []TValidator           `json:"validate,omitempty"`
	Export        []string               `json:"export,omitempty"`
}

// used for testcase json loading and dumping
type TCase struct {
	Config    TConfig  `json:"config"`
	TestSteps []*TStep `json:"teststeps"`
}

// interface for all types of steps
type IStep interface {
	Name() string
	Type() string
	ToStruct() *TStep
}

// used for testcase runner
type TestCase struct {
	Config    TConfig
	TestSteps []IStep
}

type TestCaseSummary struct{}

type StepData struct {
	Name       string                 // step name
	Success    bool                   // step execution result
	ExportVars map[string]interface{} // extract variables
}
