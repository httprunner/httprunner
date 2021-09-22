package httpboomer

type Variables map[string]interface{}
type Params map[string]interface{}
type Headers map[string]string
type Cookies map[string]string

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
	Name       string    `json:"name"`
	Verify     bool      `json:"verify"`
	BaseURL    string    `json:"base_url"`
	Variables  Variables `json:"variables"`
	Parameters Variables `json:"parameters"`
	Export     []string  `json:"export"`
	Weight     int       `json:"weight"`
}

type TRequest struct {
	Method         enumHTTPMethod `json:"method"`
	URL            string         `json:"url"`
	Params         Params         `json:"params"`
	Headers        Headers        `json:"headers"`
	Cookies        Cookies        `json:"cookies"`
	Data           interface{}    `json:"data"`
	JSON           interface{}    `json:"json"`
	Timeout        float32        `json:"timeout"`
	AllowRedirects bool           `json:"allow_redirects"`
	Verify         bool           `json:"verify"`
}

type TValidator struct {
	Check      string // get value with jmespath
	Comparator string
	Expect     interface{}
	Message    string
}

type TStep struct {
	Name          string            `json:"name"`
	Request       *TRequest         `json:"request"`
	TestCase      *TestCase         `json:"testcase"`
	Variables     Variables         `json:"variables"`
	SetupHooks    []string          `json:"setup_hooks"`
	TeardownHooks []string          `json:"teardown_hooks"`
	Extract       map[string]string `json:"extract"`
	Validators    []TValidator      `json:"validators"`
	Export        []string          `json:"export"`
}

// interface for all types of steps
type IStep interface {
	Name() string
	Type() string
	Run() error
}

type TestCase struct {
	Config    TConfig `json:"config"`
	TestSteps []IStep `json:"teststeps"`
}

type TestCaseSummary struct{}
