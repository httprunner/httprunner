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
	Path       string                 `json:"path,omitempty" yaml:"path,omitempty"` // testcase file path
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
	Transaction   *Transaction           `json:"transaction,omitempty" yaml:"transaction,omitempty"`
	Rendezvous    *Rendezvous            `json:"rendezvous,omitempty" yaml:"rendezvous,omitempty"`
	Variables     map[string]interface{} `json:"variables,omitempty" yaml:"variables,omitempty"`
	SetupHooks    []string               `json:"setup_hooks,omitempty" yaml:"setup_hooks,omitempty"`
	TeardownHooks []string               `json:"teardown_hooks,omitempty" yaml:"teardown_hooks,omitempty"`
	Extract       map[string]string      `json:"extract,omitempty" yaml:"extract,omitempty"`
	Validators    []Validator            `json:"validate,omitempty" yaml:"validate,omitempty"`
	Export        []string               `json:"export,omitempty" yaml:"export,omitempty"`
}

type stepType string

const (
	stepTypeRequest     stepType = "request"
	stepTypeTestCase    stepType = "testcase"
	stepTypeTransaction stepType = "transaction"
	stepTypeRendezvous  stepType = "rendezvous"
)

type transactionType string

const (
	transactionStart transactionType = "start"
	transactionEnd   transactionType = "end"
)

type Transaction struct {
	Name string          `json:"name" yaml:"name"`
	Type transactionType `json:"type" yaml:"type"`
}
type Rendezvous struct {
	Name    string  `json:"name" yaml:"name"`                           // required
	Percent float32 `json:"percent,omitempty" yaml:"percent,omitempty"` // default to 1(100%)
	Number  int64   `json:"number,omitempty" yaml:"number,omitempty"`
	Timeout int64   `json:"timeout,omitempty" yaml:"timeout,omitempty"` // milliseconds
}

// TCase represents testcase data structure.
// Each testcase includes one public config and several sequential teststeps.
type TCase struct {
	Config    *TConfig `json:"config" yaml:"config"`
	TestSteps []*TStep `json:"teststeps" yaml:"teststeps"`
}

// IConfig represents interface for testcase config,
// includes Config.
type IConfig interface {
	Name() string
	ToStruct() *TConfig
}

// IStep represents interface for all types for teststeps, includes:
// StepRequest, StepRequestWithOptionalArgs, StepRequestValidation, StepRequestExtraction,
// StepTestCaseWithOptionalArgs,
// StepTransaction, StepRendezvous.
type IStep interface {
	Name() string
	Type() string
	ToStruct() *TStep
}

// ITestCase represents interface for testcases,
// includes TestCase and TestCasePath.
type ITestCase interface {
	ToTestCase() (*TestCase, error)
	ToTCase() (*TCase, error)
}

// TestCase is a container for one testcase, which is used for testcase runner.
// TestCase implements ITestCase interface.
type TestCase struct {
	Config    IConfig
	TestSteps []IStep
}

func (tc *TestCase) ToTestCase() (*TestCase, error) {
	return tc, nil
}

func (tc *TestCase) ToTCase() (*TCase, error) {
	tCase := TCase{
		Config: tc.Config.ToStruct(),
	}
	for _, step := range tc.TestSteps {
		tCase.TestSteps = append(tCase.TestSteps, step.ToStruct())
	}
	return &tCase, nil
}

type testCaseSummary struct{}

type stepData struct {
	name        string                 // step name
	stepType    stepType               // step type, testcase/request/transaction/rendezvous
	success     bool                   // step execution result
	elapsed     int64                  // step execution time in millisecond(ms)
	contentSize int64                  // response body length
	exportVars  map[string]interface{} // extract variables
}
