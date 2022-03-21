package hrp

import (
	"fmt"
	"math/rand"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/httprunner/hrp/internal/builtin"
	"github.com/httprunner/hrp/internal/version"
)

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
	Name              string                 `json:"name" yaml:"name"` // required
	Verify            bool                   `json:"verify,omitempty" yaml:"verify,omitempty"`
	BaseURL           string                 `json:"base_url,omitempty" yaml:"base_url,omitempty"`
	Headers           map[string]string      `json:"headers,omitempty" yaml:"headers,omitempty"`
	Variables         map[string]interface{} `json:"variables,omitempty" yaml:"variables,omitempty"`
	Parameters        map[string]interface{} `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	ParametersSetting *TParamsConfig         `json:"parameters_setting,omitempty" yaml:"parameters_setting,omitempty"`
	ThinkTime         *ThinkTimeConfig       `json:"think_time,omitempty" yaml:"think_time,omitempty"`
	Export            []string               `json:"export,omitempty" yaml:"export,omitempty"`
	Weight            int                    `json:"weight,omitempty" yaml:"weight,omitempty"`
	Path              string                 `json:"path,omitempty" yaml:"path,omitempty"` // testcase file path
}

type TParamsConfig struct {
	Strategy  interface{} `json:"strategy,omitempty" yaml:"strategy,omitempty"` // map[string]string、string
	Iteration int         `json:"iteration,omitempty" yaml:"iteration,omitempty"`
	Iterators []*Iterator `json:"parameterIterator,omitempty" yaml:"parameterIterator,omitempty"` // 保存参数的迭代器
}

const (
	strategyRandom     string = "random"
	strategySequential string = "Sequential"
)

type ThinkTimeConfig struct {
	Strategy string      `json:"strategy,omitempty" yaml:"strategy,omitempty"` // default、random、limit、multiply、ignore
	Setting  interface{} `json:"setting,omitempty" yaml:"setting,omitempty"`   // random(map): {"min_percentage": 0.5, "max_percentage": 1.5}; 10、multiply(float64): 1.5
	Limit    float64     `json:"limit,omitempty" yaml:"limit,omitempty"`       // limit think time no more than specific time, ignore if value <= 0
}

const (
	thinkTimeDefault          string = "default"           // as recorded
	thinkTimeRandomPercentage string = "random_percentage" // use random percentage of recorded think time
	thinkTimeMultiply         string = "multiply"          // multiply recorded think time
	thinkTimeIgnore           string = "ignore"            // ignore recorded think time
)

const (
	thinkTimeDefaultMultiply = 1
)

var (
	thinkTimeDefaultRandom = map[string]float64{"min_percentage": 0.5, "max_percentage": 1.5}
)

func (ttc *ThinkTimeConfig) checkThinkTime() {
	if ttc == nil {
		return
	}
	// unset strategy, set default strategy
	if ttc.Strategy == "" {
		ttc.Strategy = thinkTimeDefault
	}
	// check think time
	if ttc.Strategy == thinkTimeRandomPercentage {
		if ttc.Setting == nil || reflect.TypeOf(ttc.Setting).Kind() != reflect.Map {
			ttc.Setting = thinkTimeDefaultRandom
			return
		}
		value, ok := ttc.Setting.(map[string]interface{})
		if !ok {
			ttc.Setting = thinkTimeDefaultRandom
			return
		}
		if _, ok := value["min_percentage"]; !ok {
			ttc.Setting = thinkTimeDefaultRandom
			return
		}
		if _, ok := value["max_percentage"]; !ok {
			ttc.Setting = thinkTimeDefaultRandom
			return
		}
		left, err := builtin.Interface2Float64(value["min_percentage"])
		if err != nil {
			ttc.Setting = thinkTimeDefaultRandom
			return
		}
		right, err := builtin.Interface2Float64(value["max_percentage"])
		if err != nil {
			ttc.Setting = thinkTimeDefaultRandom
			return
		}
		ttc.Setting = map[string]float64{"min_percentage": left, "max_percentage": right}
	} else if ttc.Strategy == thinkTimeMultiply {
		if ttc.Setting == nil {
			ttc.Setting = float64(0) // default
			return
		}
		value, err := builtin.Interface2Float64(ttc.Setting)
		if err != nil {
			ttc.Setting = float64(0) // default
			return
		}
		ttc.Setting = value
	} else if ttc.Strategy != thinkTimeIgnore {
		// unrecognized strategy, set default strategy
		ttc.Strategy = thinkTimeDefault
	}
}

type paramsType []map[string]interface{}

type Iterator struct {
	sync.Mutex
	data      paramsType
	strategy  string // random, sequential
	iteration int
	index     int
}

func (params paramsType) Iterator() *Iterator {
	return &Iterator{
		data:      params,
		iteration: len(params),
		index:     0,
	}
}

func (iter *Iterator) HasNext() bool {
	if iter.iteration == -1 {
		return true
	}
	return iter.index < iter.iteration
}

func (iter *Iterator) Next() (value map[string]interface{}) {
	iter.Lock()
	defer iter.Unlock()
	if len(iter.data) == 0 {
		iter.index++
		return map[string]interface{}{}
	}
	if iter.strategy == strategyRandom {
		randSource := rand.New(rand.NewSource(time.Now().Unix()))
		randIndex := randSource.Intn(len(iter.data))
		value = iter.data[randIndex]
	} else {
		value = iter.data[iter.index%len(iter.data)]
	}
	iter.index++
	return value
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
	Json           interface{}            `json:"json,omitempty" yaml:"json,omitempty"`
	Data           interface{}            `json:"data,omitempty" yaml:"data,omitempty"`
	Timeout        float32                `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	AllowRedirects bool                   `json:"allow_redirects,omitempty" yaml:"allow_redirects,omitempty"`
	Verify         bool                   `json:"verify,omitempty" yaml:"verify,omitempty"`
}

type API struct {
	Name          string                 `json:"name" yaml:"name"` // required
	Request       *Request               `json:"request,omitempty" yaml:"request,omitempty"`
	Variables     map[string]interface{} `json:"variables,omitempty" yaml:"variables,omitempty"`
	SetupHooks    []string               `json:"setup_hooks,omitempty" yaml:"setup_hooks,omitempty"`
	TeardownHooks []string               `json:"teardown_hooks,omitempty" yaml:"teardown_hooks,omitempty"`
	Extract       map[string]string      `json:"extract,omitempty" yaml:"extract,omitempty"`
	Validators    []interface{}          `json:"validate,omitempty" yaml:"validate,omitempty"`
	Export        []string               `json:"export,omitempty" yaml:"export,omitempty"`
}

func (api *API) ToAPI() (*API, error) {
	return api, nil
}

// Validator represents validator for one HTTP response.
type Validator struct {
	Check   string      `json:"check" yaml:"check"` // get value with jmespath
	Assert  string      `json:"assert" yaml:"assert"`
	Expect  interface{} `json:"expect" yaml:"expect"`
	Message string      `json:"msg,omitempty" yaml:"msg,omitempty"` // optional
}

// IAPI represents interface for api,
// includes API and APIPath.
type IAPI interface {
	ToAPI() (*API, error)
}

// TStep represents teststep data structure.
// Each step maybe two different type: make one HTTP request or reference another testcase.
type TStep struct {
	Name            string                 `json:"name" yaml:"name"` // required
	Request         *Request               `json:"request,omitempty" yaml:"request,omitempty"`
	APIPath         string                 `json:"api,omitempty" yaml:"api,omitempty"`
	TestCasePath    string                 `json:"testcase,omitempty" yaml:"testcase,omitempty"`
	APIContent      IAPI                   `json:"api_content,omitempty" yaml:"api_content,omitempty"`
	TestCaseContent ITestCase              `json:"testcase_content,omitempty" yaml:"testcase_content,omitempty"`
	Transaction     *Transaction           `json:"transaction,omitempty" yaml:"transaction,omitempty"`
	Rendezvous      *Rendezvous            `json:"rendezvous,omitempty" yaml:"rendezvous,omitempty"`
	ThinkTime       *ThinkTime             `json:"think_time,omitempty" yaml:"think_time,omitempty"`
	Variables       map[string]interface{} `json:"variables,omitempty" yaml:"variables,omitempty"`
	SetupHooks      []string               `json:"setup_hooks,omitempty" yaml:"setup_hooks,omitempty"`
	TeardownHooks   []string               `json:"teardown_hooks,omitempty" yaml:"teardown_hooks,omitempty"`
	Extract         map[string]string      `json:"extract,omitempty" yaml:"extract,omitempty"`
	Validators      []interface{}          `json:"validate,omitempty" yaml:"validate,omitempty"`
	Export          []string               `json:"export,omitempty" yaml:"export,omitempty"`
}

type stepType string

const (
	stepTypeRequest     stepType = "request"
	stepTypeTestCase    stepType = "testcase"
	stepTypeTransaction stepType = "transaction"
	stepTypeRendezvous  stepType = "rendezvous"
	stepTypeThinkTime   stepType = "thinktime"
)

type ThinkTime struct {
	Time float64 `json:"time" yaml:"time"`
}

type transactionType string

const (
	transactionStart transactionType = "start"
	transactionEnd   transactionType = "end"
)

type Transaction struct {
	Name string          `json:"name" yaml:"name"`
	Type transactionType `json:"type" yaml:"type"`
}

const (
	defaultRendezvousTimeout int64   = 5000
	defaultRendezvousPercent float32 = 1.0
)

type Rendezvous struct {
	Name           string  `json:"name" yaml:"name"`                           // required
	Percent        float32 `json:"percent,omitempty" yaml:"percent,omitempty"` // default to 1(100%)
	Number         int64   `json:"number,omitempty" yaml:"number,omitempty"`
	Timeout        int64   `json:"timeout,omitempty" yaml:"timeout,omitempty"` // milliseconds
	cnt            int64
	releasedFlag   uint32
	spawnDoneFlag  uint32
	wg             sync.WaitGroup
	timerResetChan chan struct{}
	activateChan   chan struct{}
	releaseChan    chan struct{}
	once           *sync.Once
	lock           sync.Mutex
}

// TCase represents testcase data structure.
// Each testcase includes one public config and several sequential teststeps.
type TCase struct {
	Config    *TConfig `json:"config" yaml:"config"`
	TestSteps []*TStep `json:"teststeps" yaml:"teststeps"`
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
	Config    *TConfig
	TestSteps []IStep
}

func (tc *TestCase) ToTestCase() (*TestCase, error) {
	return tc, nil
}

func (tc *TestCase) ToTCase() (*TCase, error) {
	tCase := TCase{
		Config: tc.Config,
	}
	for _, step := range tc.TestSteps {
		tCase.TestSteps = append(tCase.TestSteps, step.ToStruct())
	}
	return &tCase, nil
}

type testCaseStat struct {
	Total   int `json:"total" yaml:"total"`
	Success int `json:"success" yaml:"success"`
	Fail    int `json:"fail" yaml:"fail"`
}

type testStepStat struct {
	Total     int `json:"total" yaml:"total"`
	Successes int `json:"successes" yaml:"successes"`
	Failures  int `json:"failures" yaml:"failures"`
}

type stat struct {
	TestCases testCaseStat `json:"testcases" yaml:"test_cases"`
	TestSteps testStepStat `json:"teststeps" yaml:"test_steps"`
}

type testCaseTime struct {
	StartAt  time.Time `json:"start_at,omitempty" yaml:"start_at,omitempty"`
	Duration float64   `json:"duration,omitempty" yaml:"duration,omitempty"`
}

type platform struct {
	HttprunnerVersion string `json:"httprunner_version" yaml:"httprunner_version"`
	GoVersion         string `json:"go_version" yaml:"go_version"`
	Platform          string `json:"platform" yaml:"platform"`
}

// Summary stores tests summary for current task execution, maybe include one or multiple testcases
type Summary struct {
	Success  bool               `json:"success" yaml:"success"`
	Stat     *stat              `json:"stat" yaml:"stat"`
	Time     *testCaseTime      `json:"time" yaml:"time"`
	Platform *platform          `json:"platform" yaml:"platform"`
	Details  []*testCaseSummary `json:"details" yaml:"details"`
}

func newOutSummary() *Summary {
	platForm := &platform{
		HttprunnerVersion: version.VERSION,
		GoVersion:         runtime.Version(),
		Platform:          fmt.Sprintf("%v-%v", runtime.GOOS, runtime.GOARCH),
	}
	return &Summary{
		Success: true,
		Stat:    &stat{},
		Time: &testCaseTime{
			StartAt: time.Now(),
		},
		Platform: platForm,
	}
}

func (s *Summary) appendCaseSummary(caseSummary *testCaseSummary) {
	s.Success = s.Success && caseSummary.Success
	s.Stat.TestCases.Total += 1
	s.Stat.TestSteps.Total += len(caseSummary.Records)
	if caseSummary.Success {
		s.Stat.TestCases.Success += 1
	} else {
		s.Stat.TestCases.Fail += 1
	}
	s.Stat.TestSteps.Successes += caseSummary.Stat.Successes
	s.Stat.TestSteps.Failures += caseSummary.Stat.Failures
	s.Details = append(s.Details, caseSummary)
	s.Success = s.Success && caseSummary.Success
}

type stepData struct {
	Name        string                 `json:"name" yaml:"name"`                                   // step name
	StepType    stepType               `json:"step_type" yaml:"step_type"`                         // step type, testcase/request/transaction/rendezvous
	Success     bool                   `json:"success" yaml:"success"`                             // step execution result
	Elapsed     int64                  `json:"elapsed_ms" yaml:"elapsed_ms"`                       // step execution time in millisecond(ms)
	Data        interface{}            `json:"data,omitempty" yaml:"data,omitempty"`               // session data or slice of step data
	ContentSize int64                  `json:"content_size" yaml:"content_size"`                   // response body length
	ExportVars  map[string]interface{} `json:"export_vars,omitempty" yaml:"export_vars,omitempty"` // extract variables
	Attachment  string                 `json:"attachment,omitempty" yaml:"attachment,omitempty"`   // step error information
}

type testCaseInOut struct {
	ConfigVars map[string]interface{} `json:"config_vars" yaml:"config_vars"`
	ExportVars map[string]interface{} `json:"export_vars" yaml:"export_vars"`
}

// testCaseSummary stores tests summary for one testcase
type testCaseSummary struct {
	Name    string         `json:"name" yaml:"name"`
	Success bool           `json:"success" yaml:"success"`
	CaseId  string         `json:"case_id,omitempty" yaml:"case_id,omitempty"` // TODO
	Stat    *testStepStat  `json:"stat" yaml:"stat"`
	Time    *testCaseTime  `json:"time" yaml:"time"`
	InOut   *testCaseInOut `json:"in_out" yaml:"in_out"`
	Log     string         `json:"log,omitempty" yaml:"log,omitempty"` // TODO
	Records []*stepData    `json:"records" yaml:"records"`
}

type validationResult struct {
	Validator
	CheckValue  interface{} `json:"check_value" yaml:"check_value"`
	CheckResult string      `json:"check_result" yaml:"check_result"`
}

type reqResps struct {
	Request  interface{} `json:"request" yaml:"request"`
	Response interface{} `json:"response" yaml:"response"`
}

type address struct {
	ClientIP   string `json:"client_ip,omitempty" yaml:"client_ip,omitempty"`
	ClientPort string `json:"client_port,omitempty" yaml:"client_port,omitempty"`
	ServerIP   string `json:"server_ip,omitempty" yaml:"server_ip,omitempty"`
	ServerPort string `json:"server_port,omitempty" yaml:"server_port,omitempty"`
}

type SessionData struct {
	Success    bool                `json:"success" yaml:"success"`
	ReqResps   *reqResps           `json:"req_resps" yaml:"req_resps"`
	Address    *address            `json:"address,omitempty" yaml:"address,omitempty"` // TODO
	Validators []*validationResult `json:"validators,omitempty" yaml:"validators,omitempty"`
}

func newSessionData() *SessionData {
	return &SessionData{
		Success:  false,
		ReqResps: &reqResps{},
	}
}
