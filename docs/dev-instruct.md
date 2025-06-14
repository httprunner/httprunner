# 代码阅读指南

## 核心数据结构

HttpRunner 以 `TestCase` 为核心，将任意测试场景抽象为有序步骤的集合。

```go
type TestCase struct {
	Config    IConfig `json:"config" yaml:"config"`
	TestSteps []IStep `json:"teststeps" yaml:"teststeps"`
}

type IConfig interface {
	Get() *TConfig
}
```

其中，测试步骤 `IStep` 采用了 `go interface` 的设计理念，支持进行任意拓展；步骤内容统一在 `Run` 方法中进行实现。

```go
type IStep interface {
	Name() string
	Type() StepType
	Config() *StepConfig
	Run(*SessionRunner) (*StepResult, error)
}
```

我们只需遵循 `IStep` 的接口定义，即可实现各种类型的测试步骤类型。当前 hrp 已支持的步骤类型包括：

- [request](step_request.go)：发起单次 HTTP 请求
- [api](step_api.go)：引用执行其它 API 文件
- [testcase](step_testcase.go)：引用执行其它测试用例文件
- [thinktime](step_thinktime.go)：思考时间，按照配置的逻辑进行等待
- [transaction](step_transaction.go)：事务机制，用于压测
- [rendezvous](step_rendezvous.go)：集合点机制，用于压测
- [websocket](step_websocket.go)：WebSocket 通信
- [android](step_ui.go)：Android UI 自动化
- [ios](step_ui.go)：iOS UI 自动化
- [harmony](step_ui.go)：Harmony UI 自动化
- [browser](step_ui.go)：浏览器 UI 自动化
- [shell](step_shell.go)：执行 shell 命令
- [function](step_function.go)：自定义函数调用

基于该机制，我们可以扩展支持新的协议类型，例如 HTTP2/WebSocket/RPC 等；同时也可以支持新的测试类型，例如 UI 自动化。甚至我们还可以在一个测试用例中混合调用多种不同的 Step 类型，例如实现 HTTP/RPC/UI 混合场景。

## 运行主流程

### 整体控制器 HRPRunner

执行测试时，会初始化一个 `HRPRunner`，用于控制测试的执行策略。

```go
type HRPRunner struct {
	t                *testing.T
	failfast         bool
	httpStatOn       bool
	requestsLogOn    bool
	pluginLogOn      bool
	venv             string
	saveTests        bool
	genHTMLReport    bool
	httpClient       *http.Client
	http2Client      *http.Client
	wsDialer         *websocket.Dialer
	caseTimeoutTimer *time.Timer    // case timeout timer
	interruptSignal  chan os.Signal // interrupt signal channel
}

func (r *HRPRunner) Run(testcases ...ITestCase) error
func NewCaseRunner(testcase TestCase, hrpRunner *HRPRunner) (*CaseRunner, error)
```

重点关注的方法：

- Run：测试执行的主入口，支持运行一个或多个测试用例
- NewCaseRunner：针对给定的测试用例初始化一个 CaseRunner

HRPRunner 支持多种配置选项：
- SetFailfast：配置是否在步骤失败时立即停止
- SetRequestsLogOn：开启请求响应详细日志
- SetHTTPStatOn：开启 HTTP 延迟统计
- SetPluginLogOn：开启插件日志
- SetProxyUrl：配置代理 URL，用于抓包调试
- SetRequestTimeout：配置全局请求超时
- SetCaseTimeout：配置测试用例超时
- GenHTMLReport：生成 HTML 测试报告

### 用例执行器 CaseRunner

针对每个测试用例，采用 CaseRunner 存储其公共信息，包括 plugin/parser

```go
type CaseRunner struct {
	TestCase // each testcase init its own CaseRunner

	hrpRunner *HRPRunner // all case runners share one HRPRunner
	parser    *Parser    // each CaseRunner init its own Parser

	parametersIterator *ParametersIterator
}

func (r *CaseRunner) NewSession() *SessionRunner
```

重点关注一个方法：

- NewSession：测试用例的每一次执行对应一个 SessionRunner

### SessionRunner

测试用例的具体执行都由 `SessionRunner` 完成，每个 session 实例中除了包含测试用例自身内容外，还会包含测试过程的 session 数据和最终测试结果 summary。

```go
type SessionRunner struct {
	caseRunner *CaseRunner // all session runners share one CaseRunner

	sessionVariables map[string]interface{} // testcase execution session variables
	summary          *TestCaseSummary       // record test case summary

	// transactions stores transaction timing info.
	// key is transaction name, value is map of transaction type and time, e.g. start time and end time.
	transactions map[string]map[TransactionType]time.Time

	// websocket session
	ws *wsSession
}

func (r *SessionRunner) Start(givenVars map[string]interface{}) (summary *TestCaseSummary, err error)
func (r *SessionRunner) RunStep(step IStep) (stepResult *StepResult, err error)
func (r *SessionRunner) ParseStep(step IStep) error
```

重点关注的方法：

- Start：启动执行用例，依次执行所有测试步骤
- RunStep：执行单个测试步骤，支持循环执行
- ParseStep：解析步骤配置，包括变量替换和验证器解析

```go
func (r *SessionRunner) Start(givenVars map[string]interface{}) (summary *TestCaseSummary, err error) {
	// report GA event
	sdk.SendGA4Event("hrp_session_runner_start", nil)

	config := r.caseRunner.TestCase.Config.Get()
	log.Info().Str("testcase", config.Name).Msg("run testcase start")

	// update config variables with given variables
	r.InitWithParameters(givenVars)

	defer func() {
		// release session resources
		r.ReleaseResources()

		summary = r.summary
		summary.Name = config.Name
		summary.Time.Duration = time.Since(summary.Time.StartAt).Seconds()
		// ... handle export variables and logs
	}()

	// run step in sequential order
	for _, step := range r.caseRunner.TestSteps {
		select {
		case <-r.caseRunner.hrpRunner.caseTimeoutTimer.C:
			log.Warn().Msg("timeout in session runner")
			return summary, errors.Wrap(code.TimeoutError, "session runner timeout")
		case <-r.caseRunner.hrpRunner.interruptSignal:
			log.Warn().Msg("interrupted in session runner")
			return summary, errors.Wrap(code.InterruptError, "session runner interrupted")
		default:
			_, err := r.RunStep(step)
			if err == nil {
				continue
			}
			// interrupted or timeout, abort running
			if errors.Is(err, code.InterruptError) || errors.Is(err, code.TimeoutError) {
				return summary, err
			}

			// check if failfast
			if r.caseRunner.hrpRunner.failfast {
				return summary, errors.Wrap(err, "abort running due to failfast setting")
			}
		}
	}

	log.Info().Str("testcase", config.Name).Msg("run testcase end")
	return summary, nil
}
```

在主流程中，SessionRunner 并不需要关注 step 的具体类型，统一都是调用 `r.RunStep(step)`，具体实现逻辑都在对应 step 的 `Run(*SessionRunner)` 方法中。

## 新增特性

### 1. 超时和中断处理

v5 版本增加了完善的超时和中断处理机制：
- 支持测试用例级别的超时控制
- 支持优雅的中断处理（SIGTERM, SIGINT）
- 在执行过程中实时检查超时和中断信号

### 2. 多平台 UI 自动化

统一的 UI 自动化接口，支持多个平台：
- **Android**：基于 ADB 和 UIAutomator2
- **iOS**：基于 WebDriverAgent (WDA)
- **Harmony**：基于 HDC (Harmony Device Connector)
- **Browser**：基于 WebDriver 协议

### 3. AI 集成

集成了大模型能力：
- 支持 AI 驱动的 UI 操作
- 通过 MCP (Model Context Protocol) 与大模型通信
- 支持自然语言描述的测试步骤

### 4. 增强的步骤配置

步骤配置支持更多选项：
```go
type StepConfig struct {
	StepName         string                 `json:"name" yaml:"name"` // required
	Variables        map[string]interface{} `json:"variables,omitempty" yaml:"variables,omitempty"`
	SetupHooks       []string               `json:"setup_hooks,omitempty" yaml:"setup_hooks,omitempty"`
	TeardownHooks    []string               `json:"teardown_hooks,omitempty" yaml:"teardown_hooks,omitempty"`
	Extract          map[string]string      `json:"extract,omitempty" yaml:"extract,omitempty"`
	Validators       []interface{}          `json:"validate,omitempty" yaml:"validate,omitempty"`
	StepExport       []string               `json:"export,omitempty" yaml:"export,omitempty"`
	Loops            *types.IntOrString     `json:"loops,omitempty" yaml:"loops,omitempty"`
	AutoPopupHandler bool                   `json:"auto_popup_handler,omitempty" yaml:"auto_popup_handler,omitempty"` // enable auto popup handler for this step
}
```

### 5. 协议支持扩展

除了 HTTP/HTTPS，还支持：
- HTTP/2 协议
- WebSocket 通信
- 自定义函数调用

### 6. 资源管理

增强的资源管理机制：
- 自动释放会话资源
- UI 驱动器缓存管理
- 日志收集和聚合

## UI 自动化步骤示例

### StepMobile 结构

UI 自动化步骤统一使用 `StepMobile` 结构：

```go
type StepMobile struct {
	StepConfig
	Mobile  *MobileUI `json:"mobile,omitempty" yaml:"mobile,omitempty"`
	Android *MobileUI `json:"android,omitempty" yaml:"android,omitempty"`
	Harmony *MobileUI `json:"harmony,omitempty" yaml:"harmony,omitempty"`
	IOS     *MobileUI `json:"ios,omitempty" yaml:"ios,omitempty"`
	Browser *MobileUI `json:"browser,omitempty" yaml:"browser,omitempty"`
}
```

### 常用 UI 操作方法

```go
// 基础操作
func (s *StepMobile) TapXY(x, y float64, opts ...option.ActionOption) *StepMobile
func (s *StepMobile) TapByOCR(ocrText string, opts ...option.ActionOption) *StepMobile
func (s *StepMobile) TapByCV(imagePath string, opts ...option.ActionOption) *StepMobile
func (s *StepMobile) AIAction(prompt string, opts ...option.ActionOption) *StepMobile

// 应用管理
func (s *StepMobile) AppLaunch(bundleId string) *StepMobile
func (s *StepMobile) AppTerminate(bundleId string) *StepMobile
func (s *StepMobile) InstallApp(path string) *StepMobile

// 滑动操作
func (s *StepMobile) Swipe(sx, sy, ex, ey float64, opts ...option.ActionOption) *StepMobile
func (s *StepMobile) SwipeUp(opts ...option.ActionOption) *StepMobile
func (s *StepMobile) SwipeDown(opts ...option.ActionOption) *StepMobile

// 输入操作
func (s *StepMobile) Input(text string, opts ...option.ActionOption) *StepMobile

// 等待操作
func (s *StepMobile) Sleep(nSeconds float64, startTime ...time.Time) *StepMobile
func (s *StepMobile) SleepRandom(params ...float64) *StepMobile

// 验证操作
func (s *StepMobile) Validate() *StepMobileUIValidation
```

### UI 验证方法

```go
// OCR 文本验证
func (s *StepMobileUIValidation) AssertOCRExists(expectedText string, msg ...string) *StepMobileUIValidation
func (s *StepMobileUIValidation) AssertOCRNotExists(expectedText string, msg ...string) *StepMobileUIValidation

// 图像验证
func (s *StepMobileUIValidation) AssertImageExists(expectedImagePath string, msg ...string) *StepMobileUIValidation
func (s *StepMobileUIValidation) AssertImageNotExists(expectedImagePath string, msg ...string) *StepMobileUIValidation

// AI 验证
func (s *StepMobileUIValidation) AssertAI(prompt string, msg ...string) *StepMobileUIValidation

// 应用状态验证
func (s *StepMobileUIValidation) AssertAppInForeground(packageName string, msg ...string) *StepMobileUIValidation
func (s *StepMobileUIValidation) AssertAppNotInForeground(packageName string, msg ...string) *StepMobileUIValidation
```

## 开发建议

### 1. 添加新的步骤类型

要添加新的步骤类型，需要：
1. 在 `step.go` 中定义新的 `StepType` 常量
2. 创建实现 `IStep` 接口的结构体
3. 在 `testcase.go` 的 `loadISteps` 方法中添加对应的处理逻辑

### 2. 扩展 UI 平台支持

要支持新的 UI 平台：
1. 在 `uixt/` 目录下实现对应的驱动器
2. 在 `StepMobile` 中添加新的平台字段
3. 在 `obj()` 方法中添加对应的处理逻辑

### 3. 调试技巧

- 使用 `SetRequestsLogOn()` 开启详细的请求日志
- 使用 `SetPluginLogOn()` 开启插件日志
- 使用 `SetProxyUrl()` 配置代理进行抓包分析
- 查看生成的 HTML 报告了解执行详情
