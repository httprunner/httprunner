package hrp

type StepType string

const (
	stepTypeRequest     StepType = "request"
	stepTypeAPI         StepType = "api"
	stepTypeTestCase    StepType = "testcase"
	stepTypeTransaction StepType = "transaction"
	stepTypeRendezvous  StepType = "rendezvous"
	stepTypeThinkTime   StepType = "thinktime"
	stepTypeWebSocket   StepType = "websocket"
	stepTypeAndroid     StepType = "android"
	stepTypeIOS         StepType = "ios"
)

type MobileMethod string

const (
	appInstall          MobileMethod = "install"
	appUninstall        MobileMethod = "uninstall"
	appStart            MobileMethod = "app_start"
	appLaunch           MobileMethod = "app_launch"            // 等待 app 打开并堵塞到 app 首屏加载完成，可以传入 app 的启动参数、环境变量
	appLaunchUnattached MobileMethod = "app_launch_unattached" // 只负责通知打开 app，不堵塞等待，不可传入启动参数
	appTerminate        MobileMethod = "app_terminate"
	appStop             MobileMethod = "app_stop"
	ctlScreenShot       MobileMethod = "screenshot"
	ctlSleep            MobileMethod = "sleep"
	ctlStartCamera      MobileMethod = "camera_start" // alias for app_launch camera
	ctlStopCamera       MobileMethod = "camera_stop"  // alias for app_terminate camera
	recordStart         MobileMethod = "record_start"
	recordStop          MobileMethod = "record_stop"

	// UI handling
	uiHome        MobileMethod = "home"
	uiTapXY       MobileMethod = "tap_xy"
	uiTap         MobileMethod = "tap"
	uiDoubleTapXY MobileMethod = "double_tap_xy"
	uiDoubleTap   MobileMethod = "double_tap"
	uiSwipe       MobileMethod = "swipe"
	uiInput       MobileMethod = "input"

	// UI validation
	uiSelectorName     string = "ui_name"
	uiSelectorOCR      string = "ui_ocr"
	assertionExists    string = "exists"
	assertionNotExists string = "not_exists"
)

type MobileAction struct {
	Method MobileMethod `json:"method" yaml:"method"`
	Params interface{}  `json:"params,omitempty" yaml:"params,omitempty"`
}

type StepResult struct {
	Name        string                 `json:"name" yaml:"name"`                                   // step name
	StepType    StepType               `json:"step_type" yaml:"step_type"`                         // step type, testcase/request/transaction/rendezvous
	Success     bool                   `json:"success" yaml:"success"`                             // step execution result
	Elapsed     int64                  `json:"elapsed_ms" yaml:"elapsed_ms"`                       // step execution time in millisecond(ms)
	HttpStat    map[string]int64       `json:"httpstat,omitempty" yaml:"httpstat,omitempty"`       // httpstat in millisecond(ms)
	Data        interface{}            `json:"data,omitempty" yaml:"data,omitempty"`               // session data or slice of step data
	ContentSize int64                  `json:"content_size" yaml:"content_size"`                   // response body length
	ExportVars  map[string]interface{} `json:"export_vars,omitempty" yaml:"export_vars,omitempty"` // extract variables
	Attachment  string                 `json:"attachment,omitempty" yaml:"attachment,omitempty"`   // step error information
}

// TStep represents teststep data structure.
// Each step maybe three different types: make one request or reference another api/testcase.
type TStep struct {
	Name          string                 `json:"name" yaml:"name"` // required
	Request       *Request               `json:"request,omitempty" yaml:"request,omitempty"`
	API           interface{}            `json:"api,omitempty" yaml:"api,omitempty"`           // *APIPath or *API
	TestCase      interface{}            `json:"testcase,omitempty" yaml:"testcase,omitempty"` // *TestCasePath or *TestCase
	Transaction   *Transaction           `json:"transaction,omitempty" yaml:"transaction,omitempty"`
	Rendezvous    *Rendezvous            `json:"rendezvous,omitempty" yaml:"rendezvous,omitempty"`
	ThinkTime     *ThinkTime             `json:"think_time,omitempty" yaml:"think_time,omitempty"`
	WebSocket     *WebSocketAction       `json:"websocket,omitempty" yaml:"websocket,omitempty"`
	Android       *AndroidStep           `json:"android,omitempty" yaml:"android,omitempty"`
	IOS           *IOSStep               `json:"ios,omitempty" yaml:"ios,omitempty"`
	Variables     map[string]interface{} `json:"variables,omitempty" yaml:"variables,omitempty"`
	SetupHooks    []string               `json:"setup_hooks,omitempty" yaml:"setup_hooks,omitempty"`
	TeardownHooks []string               `json:"teardown_hooks,omitempty" yaml:"teardown_hooks,omitempty"`
	Extract       map[string]string      `json:"extract,omitempty" yaml:"extract,omitempty"`
	Validators    []interface{}          `json:"validate,omitempty" yaml:"validate,omitempty"`
	Export        []string               `json:"export,omitempty" yaml:"export,omitempty"`
}

// IStep represents interface for all types for teststeps, includes:
// StepRequest, StepRequestWithOptionalArgs, StepRequestValidation, StepRequestExtraction,
// StepTestCaseWithOptionalArgs,
// StepTransaction, StepRendezvous, StepWebSocket.
type IStep interface {
	Name() string
	Type() StepType
	Struct() *TStep
	Run(*SessionRunner) (*StepResult, error)
}
