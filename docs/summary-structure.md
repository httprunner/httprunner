# Summary 数据结构说明文档

## 概述

HttpRunner 的 Summary 数据结构用于存储测试执行的完整汇总信息，包括测试结果、统计数据、时间信息、平台信息以及详细的测试步骤记录。本文档基于 `summary.go` 的代码定义和实际执行产物进行详细说明。

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
| name | string | `name` | 步骤名称 |
| start_time | int64 | `start_time` | 开始时间（Unix时间戳，毫秒） |
| step_type | string | `step_type` | 步骤类型（如 "android_validation", "android"） |
| success | bool | `success` | 步骤执行是否成功 |
| elapsed_ms | int | `elapsed_ms` | 执行耗时（毫秒） |
| data | *SessionData | `data` | 步骤相关数据（包含请求响应和验证结果） |
| actions | []Action | `actions` | 执行的操作列表 |
| attachments | map[string]interface{} | `attachments` | 附件信息（如截图等） |

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

### 10. SessionData (步骤数据)

| 字段名 | 类型 | JSON标签 | 说明 |
|--------|------|----------|------|
| ReqResps | *ReqResps | `req_resps` | 请求响应数据 |
| Address | *Address | `address,omitempty` | 网络地址信息 |
| Validators | []*ValidationResult | `validators,omitempty` | 验证结果列表 |

### 11. ReqResps (请求响应)

| 字段名 | 类型 | JSON标签 | 说明 |
|--------|------|----------|------|
| Request | interface{} | `request` | 请求信息 |
| Response | interface{} | `response` | 响应信息 |

### 12. ValidationResult (验证结果)

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

### 13. Action (操作信息)

每个步骤可能包含多个操作：

| 字段名 | 类型 | JSON标签 | 说明 |
|--------|------|----------|------|
| method | string | `method` | 操作方法名（如 "app_launch", "start_to_goal"） |
| params | interface{} | `params` | 操作参数 |
| start_time | int64 | `start_time` | 操作开始时间（Unix时间戳，毫秒） |
| elapsed_ms | int | `elapsed_ms` | 操作耗时（毫秒） |
| requests | []Request | `requests` | HTTP请求记录（如果有） |
| plannings | []Planning | `plannings` | AI规划信息（UI自动化场景） |
| screen_results | []ScreenResult | `screen_results` | 屏幕截图结果 |

**示例数据**:
```json
{
    "method": "start_to_goal",
    "params": "搜索「青榕小剧场」，切换到「用户」搜索结果页，点击进入第一个搜索结果的用户个人主页",
    "start_time": 1750657275855,
    "elapsed_ms": 109543,
    "plannings": [ /* AI规划列表 */ ]
}
```

### 14. Request (请求记录)

ADB或HTTP请求的详细记录：

| 字段名 | 类型 | JSON标签 | 说明 |
|--------|------|----------|------|
| request_method | string | `request_method` | 请求方法（如 "adb", "http"） |
| request_url | string | `request_url` | 请求URL或命令 |
| request_body | string | `request_body` | 请求体或命令参数 |
| request_time | string | `request_time` | 请求时间（ISO格式） |
| response_status | int | `response_status` | 响应状态码 |
| response_duration_ms | int | `response_duration(ms)` | 响应耗时（毫秒） |
| response_body | string | `response_body` | 响应内容 |
| success | bool | `success` | 请求是否成功 |

**示例数据**:
```json
{
    "request_method": "adb",
    "request_url": "monkey",
    "request_body": "-p com.smile.gifmaker -c android.intent.category.LAUNCHER 1",
    "request_time": "2025-06-23T13:41:07.200504+08:00",
    "response_status": 0,
    "response_duration(ms)": 566,
    "response_body": "Events injected: 1\n## Network stats: elapsed time=45ms",
    "success": true
}
```

### 15. Planning (AI规划信息)

UI自动化测试中的AI规划详情：

| 字段名 | 类型 | JSON标签 | 说明 |
|--------|------|----------|------|
| tool_calls | []ToolCall | `tool_calls` | 工具调用信息 |
| thought | string | `thought` | AI的思考过程 |
| content | string | `content` | 规划内容（JSON格式的操作描述） |
| model_name | string | `model_name` | 使用的AI模型名称 |
| usage | Usage | `usage` | 模型使用统计 |
| screenshot_elapsed_ms | int | `screenshot_elapsed_ms` | 截图耗时（毫秒） |
| image_path | string | `image_path` | 截图文件路径 |
| resolution | Resolution | `resolution` | 屏幕分辨率 |
| screen_result | ScreenResult | `screen_result` | 屏幕分析结果 |
| model_call_elapsed_ms | int | `model_call_elapsed_ms` | 模型调用耗时（毫秒） |
| tool_calls_count | int | `tool_calls_count` | 工具调用次数 |
| action_names | []string | `action_names` | 执行的操作名称列表 |
| start_time | int64 | `start_time` | 规划开始时间 |
| elapsed_ms | int | `elapsed_ms` | 规划总耗时 |
| sub_actions | []SubAction | `sub_actions` | 子操作列表 |

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
    "thought": "点击页面右上角的搜索图标，打开搜索界面以进行后续搜索操作。",
    "model_name": "doubao-1.5-thinking-vision-pro-250428",
    "usage": {
        "prompt_tokens": 2199,
        "completion_tokens": 135,
        "total_tokens": 2334
    }
}
```

### 16. ToolCall (工具调用)

AI规划中的工具调用信息：

| 字段名 | 类型 | JSON标签 | 说明 |
|--------|------|----------|------|
| id | string | `id` | 工具调用唯一标识 |
| type | string | `type` | 调用类型（通常为 "function"） |
| function | Function | `function` | 函数调用详情 |

### 17. Function (函数调用)

| 字段名 | 类型 | JSON标签 | 说明 |
|--------|------|----------|------|
| name | string | `name` | 函数名称（如 "uixt__tap_xy"） |
| arguments | string | `arguments` | 函数参数（JSON字符串格式） |

### 18. Usage (模型使用统计)

| 字段名 | 类型 | JSON标签 | 说明 |
|--------|------|----------|------|
| prompt_tokens | int | `prompt_tokens` | 输入token数量 |
| completion_tokens | int | `completion_tokens` | 输出token数量 |
| total_tokens | int | `total_tokens` | 总token数量 |

### 19. Resolution (分辨率)

| 字段名 | 类型 | JSON标签 | 说明 |
|--------|------|----------|------|
| width | int | `width` | 屏幕宽度（像素） |
| height | int | `height` | 屏幕高度（像素） |

### 20. ScreenResult (屏幕结果)

| 字段名 | 类型 | JSON标签 | 说明 |
|--------|------|----------|------|
| image_path | string | `image_path` | 截图文件路径 |
| resolution | Resolution | `resolution` | 屏幕分辨率 |
| uploaded_url | string | `uploaded_url` | 上传后的URL（通常为空） |
| texts | []Text | `texts` | 识别的文本信息（可为null） |
| icons | []Icon | `icons` | 识别的图标信息（可为null） |
| tags | []Tag | `tags` | 识别的标签信息（可为null） |

### 21. SubAction (子操作)

规划中实际执行的具体操作：

| 字段名 | 类型 | JSON标签 | 说明 |
|--------|------|----------|------|
| action_name | string | `action_name` | 操作名称（如 "uixt__tap_xy"） |
| arguments | string | `arguments` | 操作参数（JSON字符串） |
| start_time | int64 | `start_time` | 操作开始时间 |
| elapsed_ms | int | `elapsed_ms` | 操作耗时 |
| requests | []Request | `requests` | 相关的请求记录 |
| screen_results | []ScreenResult | `screen_results` | 操作后的屏幕截图 |

**示例数据**:
```json
{
    "action_name": "uixt__tap_xy",
    "arguments": "{\"x\":1107.6,\"y\":232.4}",
    "start_time": 1750657286274,
    "elapsed_ms": 319,
    "requests": [ /* 请求记录 */ ],
    "screen_results": [ /* 屏幕截图 */ ]
}
```

### 22. Attachments (附件信息)

步骤执行过程中产生的附件，主要是截图：

| 字段名 | 类型 | JSON标签 | 说明 |
|--------|------|----------|------|
| screen_results | []ScreenResult | `screen_results` | 屏幕截图列表 |

## 数据类型层次关系

### 时间戳格式
- **Unix时间戳（毫秒）**: 用于 `start_time` 字段，如 `1750657267057`
- **ISO时间格式**: 用于 `start_at` 和 `request_time` 字段，如 `"2025-06-23T13:41:06.150641+08:00"`

### 耗时统计
- **毫秒级**: `elapsed_ms`, `response_duration(ms)`, `screenshot_elapsed_ms`, `model_call_elapsed_ms`
- **秒级**: `duration` 字段使用浮点数表示秒

### 状态标识
- **布尔值**: `success`, `Success` 表示操作或测试是否成功
- **字符串**: `check_result` 使用 "pass"/"fail" 表示验证结果

这个数据结构设计充分考虑了测试执行的各种场景，特别是UI自动化测试中的复杂交互和AI规划过程，为测试结果的分析和报告提供了完整的数据基础。通过详细的嵌套字段定义，开发者可以精确理解和使用每个数据元素，实现更强大的测试分析和报告功能。