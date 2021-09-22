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
	Verify     bool                   `json:"verify"`
	BaseURL    string                 `json:"base_url"`
	Variables  map[string]interface{} `json:"variables"`
	Parameters map[string]interface{} `json:"parameters"`
	Export     []string               `json:"export"`
	Weight     int                    `json:"weight"`
}

type TRequest struct {
	Method         enumHTTPMethod         `json:"method"`
	URL            string                 `json:"url"`
	Params         map[string]interface{} `json:"params"`
	Headers        map[string]string      `json:"headers"`
	Cookies        map[string]string      `json:"cookies"`
	Data           interface{}            `json:"data"`
	JSON           interface{}            `json:"json"`
	Timeout        float32                `json:"timeout"`
	AllowRedirects bool                   `json:"allow_redirects"`
	Verify         bool                   `json:"verify"`
}

type TValidator struct {
	Check      string // get value with jmespath
	Comparator string
	Expect     interface{}
	Message    string
}

type TStep struct {
	Name          string                 `json:"name"`
	Request       *TRequest              `json:"request"`
	TestCase      *TestCase              `json:"testcase"`
	Variables     map[string]interface{} `json:"variables"`
	SetupHooks    []string               `json:"setup_hooks"`
	TeardownHooks []string               `json:"teardown_hooks"`
	Extract       map[string]string      `json:"extract"`
	Validators    []TValidator           `json:"validators"`
	Export        []string               `json:"export"`
}

// interface for all types of steps
type IStep interface {
	Name() string
	Type() string
	Run(config *TConfig) error
}

type TestCase struct {
	Config    TConfig `json:"config"`
	TestSteps []IStep `json:"teststeps"`
}

type TestCaseSummary struct{}
