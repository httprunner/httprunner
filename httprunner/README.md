# 代码阅读指南（python 部分）

## 核心数据结构

HttpRunner 以 `TestCase` 为核心，将任意测试场景抽象为有序步骤的集合。

```py
class TestCase(BaseModel):
    config: TConfig
    teststeps: List[TStep]
```

针对每种测试步骤，统一继承自 `IStep`，并要求必须至少实现如下 4 个方法；步骤内容统一在 `run` 方法中进行实现。

```py
class IStep(object):

    def name(self) -> str:
        raise NotImplementedError

    def type(self) -> str:
        raise NotImplementedError

    def struct(self) -> TStep:
        raise NotImplementedError

    def run(self, runner) -> StepData:
        # runner: HttpRunner
        raise NotImplementedError
```

我们只需遵循 `IStep` 的接口定义，即可实现各种类型的测试步骤类型。当前 python 版本已支持的步骤类型包括：

- [request](step_request.py)：发起单次 HTTP 请求
- [testcase](step_testcase.py)：引用执行其它测试用例文件

基于该机制，我们可以扩展支持新的协议类型，例如 HTTP2/WebSocket/RPC 等；同时也可以支持新的测试类型，例如 UI 自动化。甚至我们还可以在一个测试用例中混合调用多种不同的 Step 类型，例如实现 HTTP/RPC/UI 混合场景。

## 用例编写

## 运行主流程

### 整体控制器 pytest

不同于 golang 版本，python 版本的控制逻辑都基于 `pytest` 的用例发现和执行机制。

- 如果是运行 JSON/YAML 格式的用例，hrp 会将用例转换为 pytest 支持的用例格式
- 如果是要自行编写 pytest 测试用例，需要遵循 HttpRunner 的格式要求

### pytest 用例格式要求

所有测试用例要求都继承自 `HttpRunner`，然后

结构如下所示：

```py
class TestCaseRequestWithFunctions(HttpRunner):

    config = (
        Config("request methods testcase with functions")
    )

    teststeps = [
        Step(
            RunRequest("get with params")...
        ),
        Step(
            RunRequest("post raw text")...
        ),
        Step(
            RunRequest("post form data")...
        ),
    ]
```

完整案例可参考：

- [request_with_functions_test.py](../examples/postman_echo/request_methods/request_with_functions_test.py)：用例中包含了 requests 的情况
- [request_with_testcase_reference_test.py](../examples/postman_echo/request_methods/request_with_testcase_reference_test.py)：用例中包含了引用其它测试用例的情况

### 用例执行器 SessionRunner

测试用例的具体执行都由 `SessionRunner` 完成，每个 TestCase 对应一个实例，在该实例中除了包含测试用例自身内容外，还会包含测试过程的 session 数据和最终测试结果 summary。

```py
class SessionRunner(object):
	config: Config
    teststeps: List[object]     # list of Step
    ...
```

重点关注一个方法：

- test_start：该方法将被 pytest 发现，作为启动执行入口，依次执行所有测试步骤

```go
def test_start(self, param: Dict = None) -> "SessionRunner":
    """main entrance, discovered by pytest"""
    self.__start_at = time.time()
    try:
        # run step in sequential order
        for step in self.teststeps:
            self.__run_step(step)
    finally:
        logger.info(f"generate testcase log: {self.__log_path}")

    self.__duration = time.time() - self.__start_at
```

在主流程中，SessionRunner 并不需要关注 step 的具体类型，统一都是调用 `step.run(self)`，具体实现逻辑都在对应 step 的 `run` 方法中。

```py
def run(self, runner: HttpRunner) -> StepData:
    return self.__step.run(runner)
```
