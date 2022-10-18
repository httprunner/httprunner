# 代码阅读指南（golang 部分）

## 核心数据结构

HttpRunner 以 `TestCase` 为核心，将任意测试场景抽象为有序步骤的集合。

```go
type TestCase struct {
	Config    *TConfig
	TestSteps []IStep
}
```

其中，测试步骤 `IStep` 采用了 `go interface` 的设计理念，支持进行任意拓展；步骤内容统一在 `Run` 方法中进行实现。

```go
type IStep interface {
	Name() string
	Type() StepType
	Struct() *TStep
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

基于该机制，我们可以扩展支持新的协议类型，例如 HTTP2/WebSocket/RPC 等；同时也可以支持新的测试类型，例如 UI 自动化。甚至我们还可以在一个测试用例中混合调用多种不同的 Step 类型，例如实现 HTTP/RPC/UI 混合场景。

## 运行主流程

### 整体控制器 HRPRunner

执行接口测试时，会初始化一个 `HRPRunner`，用于控制测试的执行策略。

```go
type HRPRunner struct {
	t             *testing.T
	failfast      bool
	requestsLogOn bool
	pluginLogOn   bool
	saveTests     bool
	genHTMLReport bool
	client        *http.Client
}

func (r *HRPRunner) Run(testcases ...ITestCase) error
func (r *HRPRunner) NewCaseRunner(testcase *TestCase) (*CaseRunner, error)
```

重点关注两个方法：

- Run：测试执行的主入口，支持运行一个或多个测试用例
- NewCaseRunner：针对给定的测试用例初始化一个 CaseRunner

### 用例执行器 CaseRunner

针对每个测试用例，采用 CaseRunner 存储其公共信息，包括 plugin/parser

```go
type CaseRunner struct {
	testCase  *TestCase
	hrpRunner *HRPRunner
	parser    *Parser

	parsedConfig       *TConfig
	parametersIterator *ParametersIterator
	rootDir            string // project root dir
}

func (r *CaseRunner) NewSession() *SessionRunner {
```

重点关注一个方法：

- NewSession：测试用例的每一次执行对应一个 SessionRunner

### SessionRunner

测试用例的具体执行都由 `SessionRunner` 完成，每个 session 实例中除了包含测试用例自身内容外，还会包含测试过程的 session 数据和最终测试结果 summary。

```go
type SessionRunner struct {
	caseRunner       *CaseRunner
	sessionVariables map[string]interface{}
	transactions      map[string]map[transactionType]time.Time
	startTime         time.Time                  // record start time of the testcase
	summary           *TestCaseSummary           // record test case summary
}

func (r *SessionRunner) Start(givenVars map[string]interface{}) error
```

重点关注一个方法：

- Start：启动执行用例，依次执行所有测试步骤

```go
func (r *SessionRunner) Start(givenVars map[string]interface{}) error {
	...
	r.resetSession()

	r.InitWithParameters(givenVars)

	// run step in sequential order
	for _, step := range r.testCase.TestSteps {
		// parse step

		// run step
		stepResult, err := step.Run(r)

		// update summary
		r.summary.Records = append(r.summary.Records, stepResult)

		// update extracted variables
		for k, v := range stepResult.ExportVars {
			r.sessionVariables[k] = v
		}

		// check if failfast
		if err != nil && r.caseRunner.hrpRunner.failfast {
			return errors.Wrap(err, "abort running due to failfast setting")
		}
	}
	...
}
```

在主流程中，SessionRunner 并不需要关注 step 的具体类型，统一都是调用 `step.Run(r)`，具体实现逻辑都在对应 step 的 `Run(*SessionRunner)` 方法中。
