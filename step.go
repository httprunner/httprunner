package hrp

import (
	"github.com/httprunner/httprunner/v5/uixt"
	"github.com/httprunner/httprunner/v5/uixt/option"
)

type StepType string

const (
	StepTypeRequest     StepType = "request"
	StepTypeAPI         StepType = "api"
	StepTypeTestCase    StepType = "testcase"
	StepTypeTransaction StepType = "transaction"
	StepTypeRendezvous  StepType = "rendezvous"
	StepTypeThinkTime   StepType = "thinktime"
	StepTypeWebSocket   StepType = "websocket"
	StepTypeAndroid     StepType = "android"
	StepTypeHarmony     StepType = "harmony"
	StepTypeIOS         StepType = "ios"
	stepTypeBrowser     StepType = "browser"
	StepTypeShell       StepType = "shell"
	StepTypeFunction    StepType = "function"

	stepTypeSuffixExtraction StepType = "_extraction"
	stepTypeSuffixValidation StepType = "_validation"
)

type StepConfig struct {
	StepName          string                 `json:"name" yaml:"name"` // required
	Variables         map[string]interface{} `json:"variables,omitempty" yaml:"variables,omitempty"`
	Parameters        map[string]interface{} `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	ParametersSetting *TParamsConfig         `json:"parameters_setting,omitempty" yaml:"parameters_setting,omitempty"`
	SetupHooks        []string               `json:"setup_hooks,omitempty" yaml:"setup_hooks,omitempty"`
	TeardownHooks     []string               `json:"teardown_hooks,omitempty" yaml:"teardown_hooks,omitempty"`
	Extract           map[string]string      `json:"extract,omitempty" yaml:"extract,omitempty"`
	Validators        []interface{}          `json:"validate,omitempty" yaml:"validate,omitempty"`
	StepExport        []string               `json:"export,omitempty" yaml:"export,omitempty"`
	Loops             int                    `json:"loops,omitempty" yaml:"loops,omitempty"`
	IgnorePopup       bool                   `json:"ignore_popup,omitempty" yaml:"ignore_popup,omitempty"`             // ignore popup for this step, keep for compatibility
	AutoPopupHandler  bool                   `json:"auto_popup_handler,omitempty" yaml:"auto_popup_handler,omitempty"` // enable auto popup handler for this step
}

// define struct for teststep
type TStep struct {
	StepConfig  `json:",inline" yaml:",inline"`
	Request     *Request         `json:"request,omitempty" yaml:"request,omitempty"`
	API         interface{}      `json:"api,omitempty" yaml:"api,omitempty"`           // *APIPath or *API
	TestCase    interface{}      `json:"testcase,omitempty" yaml:"testcase,omitempty"` // *TestCasePath or *TestCase
	Transaction *Transaction     `json:"transaction,omitempty" yaml:"transaction,omitempty"`
	Rendezvous  *Rendezvous      `json:"rendezvous,omitempty" yaml:"rendezvous,omitempty"`
	ThinkTime   *ThinkTime       `json:"think_time,omitempty" yaml:"think_time,omitempty"`
	WebSocket   *WebSocketAction `json:"websocket,omitempty" yaml:"websocket,omitempty"`
	Android     *MobileUI        `json:"android,omitempty" yaml:"android,omitempty"`
	Harmony     *MobileUI        `json:"harmony,omitempty" yaml:"harmony,omitempty"`
	IOS         *MobileUI        `json:"ios,omitempty" yaml:"ios,omitempty"`
	Browser     *MobileUI        `json:"browser,omitempty" yaml:"browser,omitempty"`
	Shell       *Shell           `json:"shell,omitempty" yaml:"shell,omitempty"`
}

// one step contains one or multiple actions
type ActionResult struct {
	option.MobileAction `json:",inline"`
	StartTime           int64                           `json:"start_time"`          // action start time in millisecond(ms)
	Elapsed             int64                           `json:"elapsed_ms"`          // action elapsed time(ms)
	Error               string                          `json:"error,omitempty"`     // action execution result
	Plannings           []*uixt.PlanningExecutionResult `json:"plannings,omitempty"` // store planning results for start_to_goal actions, which contains multiple sub-actions
	AIResult            *uixt.AIExecutionResult         `json:"ai_result,omitempty"` // store unified AI execution result for ai_query/ai_action/ai_assert actions
	uixt.SessionData                                    // store session data for other actions besides start_to_goal
	ExtraData           interface{}                     `json:"extra_data,omitempty"`
}

// one testcase contains one or multiple steps
type StepResult struct {
	Name        string                 `json:"name" yaml:"name"`                                     // step name
	StartTime   int64                  `json:"start_time" yaml:"time"`                               // step start time in millisecond(ms)
	StepType    StepType               `json:"step_type" yaml:"step_type"`                           // step type, testcase/request/transaction/rendezvous
	Success     bool                   `json:"success" yaml:"success"`                               // step execution result
	Elapsed     int64                  `json:"elapsed_ms" yaml:"elapsed_ms"`                         // step execution time in millisecond(ms)
	HttpStat    map[string]int64       `json:"httpstat,omitempty" yaml:"httpstat,omitempty"`         // httpstat in millisecond(ms)
	Data        interface{}            `json:"data,omitempty" yaml:"data,omitempty"`                 // step data
	ContentSize int64                  `json:"content_size,omitempty" yaml:"content_size,omitempty"` // response body length
	ExportVars  map[string]interface{} `json:"export_vars,omitempty" yaml:"export_vars,omitempty"`   // extract variables
	Actions     []*ActionResult        `json:"actions,omitempty" yaml:"actions,omitempty"`           // store action execution info
	Attachments interface{}            `json:"attachments,omitempty" yaml:"attachments,omitempty"`   // store extra step information, such as error message or screenshots
}

// IStep represents interface for all types for teststeps, includes:
// StepRequest, StepRequestWithOptionalArgs, StepRequestValidation, StepRequestExtraction,
// StepTestCaseWithOptionalArgs,
// StepTransaction, StepRendezvous, StepWebSocket.
type IStep interface {
	Name() string
	Type() StepType
	Config() *StepConfig
	Run(*SessionRunner) (*StepResult, error)
}
