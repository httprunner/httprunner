# Summary 数据结构说明文档

## 概述

HttpRunner 的 Summary 数据结构用于存储测试执行的完整汇总信息，包括测试结果、统计数据、时间信息、平台信息以及详细的测试步骤记录。本文档基于 `summary.go` 和相关代码的最新定义进行详细说明。

## 数据结构层次关系

```
Summary (根结构)
├── Success (bool)
├── Stat (统计信息)
│   ├── TestCases (测试用例统计)
│   └── TestSteps (测试步骤统计)
├── Time (时间信息)
├── Platform (平台信息)
└── Details (测试用例详情列表)
    └── TestCaseSummary (单个测试用例汇总)
        ├── Stat (步骤统计)
        ├── Time (用例时间)
        ├── InOut (输入输出)
        ├── Logs (日志)
        └── Records (步骤记录)
            └── StepResult (步骤结果)
                ├── Data (步骤数据)
                │   ├── ReqResps (请求响应)
                │   └── Validators (验证器)
                ├── Actions (操作列表)
                │   ├── Requests (请求记录)
                │   ├── Plannings (AI规划)
                │   │   ├── ToolCalls (工具调用)
                │   │   ├── Usage (模型使用统计)
                │   │   ├── ScreenResult (屏幕结果)
                │   │   └── SubActions (子操作)
                │   ├── AIResult (统一AI操作结果)
                │   │   ├── QueryResult (查询结果)
                │   │   ├── PlanningResult (规划结果)
                │   │   └── AssertionResult (断言结果)
                │   └── ScreenResults (屏幕截图)
                └── Attachments (附件信息)
```

## 详细数据结构说明

### 1. Summary (主汇总结构)

| 字段名 | 类型 | JSON标签 | 说明 |
|--------|------|----------|------|
| Success | bool | `success` | 整体测试执行是否成功 |
| Stat | *Stat | `stat` | 汇总统计信息 |
| Time | *TestCaseTime | `time` | 整体执行时间信息 |
| Platform | *Platform | `platform` | 平台和版本信息 |
| Details | []*TestCaseSummary | `details` | 各个测试用例的详细信息 |
| rootDir | string | - | 根目录路径（私有字段） |

**示例数据**:
```json
{
    "success": true,
    "stat": { /* 统计信息 */ },
    "time": { /* 时间信息 */ },
    "platform": { /* 平台信息 */ },
    "details": [ /* 测试用例详情 */ ]
}
```

### 2. Stat (统计信息)

| 字段名 | 类型 | JSON标签 | 说明 |
|--------|------|----------|------|
| TestCases | TestCaseStat | `testcases` | 测试用例统计 |
| TestSteps | TestStepStat | `teststeps` | 测试步骤统计 |

### 3. TestCaseStat (测试用例统计)

| 字段名 | 类型 | JSON标签 | 说明 |
|--------|------|----------|------|
| Total | int | `total` | 测试用例总数 |
| Success | int | `success` | 成功的测试用例数 |
| Fail | int | `fail` | 失败的测试用例数 |

**示例数据**:
```json
{
    "testcases": {
        "total": 1,
        "success": 1,
        "fail": 0
    }
}
```

### 4. TestStepStat (测试步骤统计)

| 字段名 | 类型 | JSON标签 | 说明 |
|--------|------|----------|------|
| Total | int | `total` | 测试步骤总数 |
| Successes | int | `successes` | 成功的步骤数 |
| Failures | int | `failures` | 失败的步骤数 |
| Actions | map[option.ActionName]int | `actions` | 各种操作的统计计数 |

**示例数据**:
```json
{
    "teststeps": {
        "total": 5,
        "successes": 5,
        "failures": 0,
        "actions": {}
    }
}
```

### 5. TestCaseTime (时间信息)

| 字段名 | 类型 | JSON标签 | 说明 |
|--------|------|----------|------|
| StartAt | time.Time | `start_at,omitempty` | 开始时间 |
| Duration | float64 | `duration,omitempty` | 持续时间（秒） |

**示例数据**:
```json
{
    "time": {
        "start_at": "2025-06-23T13:41:06.150641+08:00",
        "duration": 188.998332334
    }
}
```

### 6. Platform (平台信息)

| 字段名 | 类型 | JSON标签 | 说明 |
|--------|------|----------|------|
| HttprunnerVersion | string | `httprunner_version` | HttpRunner 版本号 |
| GoVersion | string | `go_version` | Go 语言版本 |
| Platform | string | `platform` | 操作系统平台信息 |

**示例数据**:
```json
{
    "platform": {
        "httprunner_version": "v5.0.0-beta-2506222254",
        "go_version": "go1.24.1",
        "platform": "darwin-arm64"
    }
}
```

### 7. TestCaseSummary (测试用例汇总)

| 字段名 | 类型 | JSON标签 | 说明 |
|--------|------|----------|------|
| Name | string | `name` | 测试用例名称 |
| Success | bool | `success` | 用例执行是否成功 |
| CaseId | string | `case_id,omitempty` | 用例ID（可选） |
| Stat | *TestStepStat | `stat` | 该用例的步骤统计 |
| Time | *TestCaseTime | `time` | 该用例的时间信息 |
| InOut | *TestCaseInOut | `in_out` | 输入输出信息 |
| Logs | []interface{} | `logs,omitempty` | 日志信息 |
| Records | []*StepResult | `records` | 步骤执行记录 |
| RootDir | string | `root_dir,omitempty` | 根目录路径 |

### 8. TestCaseInOut (输入输出信息)

| 字段名 | 类型 | JSON标签 | 说明 |
|--------|------|----------|------|
| ConfigVars | map[string]interface{} | `config_vars` | 配置变量 |
| ExportVars | map[string]interface{} | `export_vars` | 导出变量 |

**示例数据**:
```json
{
    "in_out": {
        "config_vars": {
            "OPENAI_API_KEY": "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",
            "dramaName": "涂山赊刀",
            "userName": "青榕小剧场"
        },
        "export_vars": {}
    }
}
```

### 9. StepResult (步骤结果)

步骤结果是 Records 数组中的元素，包含每个测试步骤的详细执行信息：

| 字段名 | 类型 | JSON标签 | 说明 |
|--------|------|----------|------|
| Name | string | `name` | 步骤名称 |
| Identifier | string | `identifier,omitempty` | 步骤标识符 |
| StartTime | int64 | `start_time` | 开始时间（Unix时间戳，毫秒） |
| StepType | StepType | `step_type` | 步骤类型（如 "android_validation", "android"） |
| Success | bool | `success` | 步骤执行是否成功 |
| Elapsed | int64 | `elapsed_ms` | 执行耗时（毫秒） |
| HttpStat | map[string]int64 | `httpstat,omitempty` | HTTP统计信息（毫秒） |
| Data | interface{} | `data,omitempty` | 步骤相关数据 |
| ContentSize | int64 | `content_size,omitempty` | 响应体长度 |
| ExportVars | map[string]interface{} | `export_vars,omitempty` | 提取的变量 |
| Actions | []*ActionResult | `actions,omitempty` | 执行的操作列表 |
| Attachments | interface{} | `attachments,omitempty` | 附件信息（如截图等） |

**示例数据**:
```json
{
    "name": "启动快手 app",
    "start_time": 1750657267057,
    "step_type": "android_validation",
    "success": true,
    "elapsed_ms": 8797,
    "data": { /* 步骤数据 */ },
    "actions": [ /* 操作列表 */ ],
    "attachments": { /* 附件信息 */ }
}
```

### 10. ActionResult (操作结果)

每个步骤可能包含多个操作，每个操作的详细执行信息：

| 字段名 | 类型 | JSON标签 | 说明 |
|--------|------|----------|------|
| MobileAction | option.MobileAction | `,inline` | 移动端操作信息（内联） |
| StartTime | int64 | `start_time` | 操作开始时间（Unix时间戳，毫秒） |
| Elapsed | int64 | `elapsed_ms` | 操作耗时（毫秒） |
| Error | string | `error,omitempty` | 操作执行错误信息 |
| Plannings | []*uixt.PlanningExecutionResult | `plannings,omitempty` | AI规划执行结果（用于start_to_goal操作） |
| AIResult | *uixt.AIExecutionResult | `ai_result,omitempty` | 统一AI执行结果（用于ai_query/ai_action/ai_assert操作） |
| SessionData | uixt.SessionData | - | 会话数据（内联，包含请求和屏幕截图信息） |

### 11. AIExecutionResult (统一AI执行结果)

这是所有AI操作（ai_query、ai_action、ai_assert）的统一结果结构：

| 字段名 | 类型 | JSON标签 | 说明 |
|--------|------|----------|------|
| Type | string | `type` | 操作类型："query"、"action"、"assert" |
| ModelCallElapsed | int64 | `model_call_elapsed` | 模型调用耗时（毫秒） |
| ScreenshotElapsed | int64 | `screenshot_elapsed` | 截图耗时（毫秒） |
| ImagePath | string | `image_path` | 截图文件路径 |
| Resolution | *types.Size | `resolution` | 屏幕分辨率 |
| QueryResult | *ai.QueryResult | `query_result,omitempty` | 查询操作结果（仅query类型） |
| PlanningResult | *ai.PlanningResult | `planning_result,omitempty` | 规划操作结果（仅action类型） |
| AssertionResult | *ai.AssertionResult | `assertion_result,omitempty` | 断言操作结果（仅assert类型） |
| Error | string | `error,omitempty` | 操作失败的错误信息 |

**示例数据**:
```json
{
    "type": "query",
    "model_call_elapsed": 1234,
    "screenshot_elapsed": 567,
    "image_path": "/path/to/screenshot.png",
    "resolution": {"width": 1080, "height": 1920},
    "query_result": { /* 查询结果详情 */ }
}
```

### 12. QueryResult (查询结果)

用于ai_query操作的具体结果：

| 字段名 | 类型 | JSON标签 | 说明 |
|--------|------|----------|------|
| Content | string | `content` | 提取的内容/信息 |
| Thought | string | `thought` | AI的推理过程 |
| Data | interface{} | `data,omitempty` | 结构化数据（当提供OutputSchema时） |
| ModelName | string | `model_name` | 使用的模型名称 |
| Usage | *schema.TokenUsage | `usage,omitempty` | token使用统计 |

**示例数据**:
```json
{
    "content": "搜索框位于屏幕右上角",
    "thought": "通过分析截图，我看到了页面右上角有一个搜索图标",
    "model_name": "doubao-1.5-thinking-vision-pro-250428",
    "usage": {
        "prompt_tokens": 1234,
        "completion_tokens": 56,
        "total_tokens": 1290
    }
}
```

### 13. PlanningResult (规划结果)

用于ai_action操作的具体结果：

| 字段名 | 类型 | JSON标签 | 说明 |
|--------|------|----------|------|
| ToolCalls | []schema.ToolCall | `tool_calls` | 工具调用列表 |
| Thought | string | `thought` | AI的思考过程 |
| Content | string | `content` | 模型的原始内容 |
| Error | string | `error,omitempty` | 规划错误信息 |
| ModelName | string | `model_name` | 使用的模型名称 |
| Usage | *schema.TokenUsage | `usage,omitempty` | token使用统计 |

**示例数据**:
```json
{
    "tool_calls": [
        {
            "id": "tap_xy_1750657286",
            "type": "function",
            "function": {
                "name": "uixt__tap_xy",
                "arguments": "{\"x\":1107.6,\"y\":232.4}"
            }
        }
    ],
    "thought": "点击页面右上角的搜索图标，打开搜索界面",
    "model_name": "doubao-1.5-thinking-vision-pro-250428",
    "usage": {
        "prompt_tokens": 2199,
        "completion_tokens": 135,
        "total_tokens": 2334
    }
}
```

### 14. AssertionResult (断言结果)

用于ai_assert操作的具体结果：

| 字段名 | 类型 | JSON标签 | 说明 |
|--------|------|----------|------|
| Pass | bool | `pass` | 断言是否通过 |
| Thought | string | `thought` | AI的推理过程 |
| ModelName | string | `model_name` | 使用的模型名称 |
| Usage | *schema.TokenUsage | `usage,omitempty` | token使用统计 |

**示例数据**:
```json
{
    "pass": true,
    "thought": "根据截图分析，当前页面确实显示了搜索结果",
    "model_name": "doubao-1.5-thinking-vision-pro-250428",
    "usage": {
        "prompt_tokens": 1500,
        "completion_tokens": 45,
        "total_tokens": 1545
    }
}
```

### 15. SessionData (会话数据)

| 字段名 | 类型 | JSON标签 | 说明 |
|--------|------|----------|------|
| ReqResps | *ReqResps | `req_resps` | 请求响应数据 |
| Address | *Address | `address,omitempty` | 网络地址信息 |
| Validators | []*ValidationResult | `validators,omitempty` | 验证结果列表 |

### 16. ReqResps (请求响应)

| 字段名 | 类型 | JSON标签 | 说明 |
|--------|------|----------|------|
| Request | interface{} | `request` | 请求信息 |
| Response | interface{} | `response` | 响应信息 |

### 17. ValidationResult (验证结果)

| 字段名 | 类型 | JSON标签 | 说明 |
|--------|------|----------|------|
| Check | string | `check` | 验证检查项 |
| Assert | string | `assert` | 断言类型 |
| Expect | interface{} | `expect` | 期望值 |
| Msg | string | `msg` | 验证消息 |
| CheckValue | interface{} | `check_value` | 实际检查值 |
| CheckResult | string | `check_result` | 检查结果（"pass"/"fail"） |

**示例数据**:
```json
{
    "check": "ui_foreground_app",
    "assert": "equal",
    "expect": "com.smile.gifmaker",
    "msg": "app [com.smile.gifmaker] should be in foreground",
    "check_value": null,
    "check_result": "pass"
}
```

### 18. PlanningExecutionResult (规划执行结果)

用于复杂的start_to_goal操作，包含规划和执行的完整信息：

| 字段名 | 类型 | JSON标签 | 说明 |
|--------|------|----------|------|
| PlanningResult | ai.PlanningResult | - | 继承的规划结果字段 |
| ScreenshotElapsed | int64 | `screenshot_elapsed_ms` | 截图耗时（毫秒） |
| ImagePath | string | `image_path` | 截图文件路径 |
| Resolution | *types.Size | `resolution` | 图像分辨率 |
| ScreenResult | *ScreenResult | `screen_result` | 完整屏幕结果数据 |
| ModelCallElapsed | int64 | `model_call_elapsed_ms` | 模型调用耗时（毫秒） |
| ToolCallsCount | int | `tool_calls_count` | 生成的工具调用数量 |
| ActionNames | []string | `action_names` | 解析的操作名称列表 |
| StartTime | int64 | `start_time` | 规划开始时间 |
| Elapsed | int64 | `elapsed_ms` | 规划耗时（毫秒） |
| SubActions | []*SubActionResult | `sub_actions,omitempty` | 此规划生成的子操作 |

### 19. ToolCall (工具调用)

AI规划中的工具调用信息：

| 字段名 | 类型 | JSON标签 | 说明 |
|--------|------|----------|------|
| ID | string | `id` | 工具调用唯一标识 |
| Type | string | `type` | 调用类型（通常为 "function"） |
| Function | Function | `function` | 函数调用详情 |

### 20. Function (函数调用)

| 字段名 | 类型 | JSON标签 | 说明 |
|--------|------|----------|------|
| Name | string | `name` | 函数名称（如 "uixt__tap_xy"） |
| Arguments | string | `arguments` | 函数参数（JSON字符串格式） |

### 21. TokenUsage (token使用统计)

| 字段名 | 类型 | JSON标签 | 说明 |
|--------|------|----------|------|
| PromptTokens | int | `prompt_tokens` | 输入token数量 |
| CompletionTokens | int | `completion_tokens` | 输出token数量 |
| TotalTokens | int | `total_tokens` | 总token数量 |

### 22. Size (分辨率)

| 字段名 | 类型 | JSON标签 | 说明 |
|--------|------|----------|------|
| Width | int | `width` | 屏幕宽度（像素） |
| Height | int | `height` | 屏幕高度（像素） |

### 23. ScreenResult (屏幕结果)

| 字段名 | 类型 | JSON标签 | 说明 |
|--------|------|----------|------|
| ImagePath | string | `image_path` | 截图文件路径 |
| Resolution | Size | `resolution` | 屏幕分辨率 |
| UploadedURL | string | `uploaded_url` | 上传后的URL（通常为空） |
| Texts | []Text | `texts` | 识别的文本信息（可为null） |
| Icons | []Icon | `icons` | 识别的图标信息（可为null） |
| Tags | []Tag | `tags` | 识别的标签信息（可为null） |

### 24. SubActionResult (子操作结果)

规划中实际执行的具体操作：

| 字段名 | 类型 | JSON标签 | 说明 |
|--------|------|----------|------|
| ActionName | string | `action_name` | 操作名称（如 "uixt__tap_xy"） |
| Arguments | interface{} | `arguments,omitempty` | 操作参数 |
| StartTime | int64 | `start_time` | 操作开始时间 |
| Elapsed | int64 | `elapsed_ms` | 操作耗时 |
| Error | error | `error,omitempty` | 操作执行错误 |
| SessionData | SessionData | - | 会话数据（内联） |

**示例数据**:
```json
{
    "action_name": "uixt__tap_xy",
    "arguments": {"x": 1107.6, "y": 232.4},
    "start_time": 1750657286274,
    "elapsed_ms": 319,
    "requests": [ /* 请求记录 */ ],
    "screen_results": [ /* 屏幕截图 */ ]
}
```

## 重要架构变更说明

### AI操作统一架构

HttpRunner v5 引入了统一的AI操作架构：

1. **统一结果结构**: `AIExecutionResult` 作为所有AI操作的统一容器
2. **类型区分**: 通过 `Type` 字段区分不同的AI操作类型
3. **具体结果**: 根据操作类型，在对应的结果字段中存储具体数据
4. **统一时间统计**: 所有AI操作都包含模型调用和截图的时间统计
5. **统一错误处理**: 通过 `Error` 字段统一处理所有AI操作的错误

### 数据类型和时间格式

#### 时间戳格式
- **Unix时间戳（毫秒）**: 用于 `start_time` 字段，如 `1750657267057`
- **ISO时间格式**: 用于 `start_at` 和 `request_time` 字段，如 `"2025-06-23T13:41:06.150641+08:00"`

#### 耗时统计
- **毫秒级**: `elapsed_ms`, `model_call_elapsed`, `screenshot_elapsed` 等
- **秒级**: `duration` 字段使用浮点数表示秒

#### 状态标识
- **布尔值**: `success`, `pass` 表示操作或测试是否成功
- **字符串**: `check_result` 使用 "pass"/"fail" 表示验证结果

这个数据结构设计充分考虑了现代UI自动化测试的需求，特别是AI驱动的测试场景。通过统一的AI操作架构和详细的嵌套字段定义，为测试结果的分析、报告生成和调试提供了完整的数据基础。