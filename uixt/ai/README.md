# HttpRunner UIXT AI 模块

## 🚀 概述

HttpRunner UIXT AI 模块是一个集成了多种人工智能服务的 UI 自动化智能引擎，提供基于大语言模型（LLM）的智能规划、断言验证、信息查询、计算机视觉识别等功能，实现真正的智能化 UI 自动化测试。

## ✨ 核心特性

### 🎯 智能组件

- **智能规划器 (Planner)**: 基于视觉语言模型进行 UI 操作规划
- **智能断言器 (Asserter)**: 基于视觉语言模型进行断言验证
- **智能查询器 (Querier)**: 从屏幕截图中提取结构化信息
- **计算机视觉 (CV)**: OCR 文本识别、UI 元素检测、弹窗识别

### 🔧 灵活配置

- **统一 API**: 通过 `NewXTDriver` 统一初始化，无需额外函数
- **混合模型**: 支持为三个组件分别选择不同的最优模型
- **预设配置**: 提供多种推荐配置方案

## 📖 使用指南

### 基本用法

```go
import (
    "github.com/httprunner/httprunner/v5/uixt"
    "github.com/httprunner/httprunner/v5/uixt/option"
)

// 方式1: 使用单一模型
driver, err := uixt.NewXTDriver(mockDriver,
    option.WithLLMService(option.OPENAI_GPT_4O))

// 方式2: 使用高级配置 - 为不同组件选择不同模型
config := option.NewLLMServiceConfig(option.DOUBAO_1_5_THINKING_VISION_PRO_250428).
    WithPlannerModel(option.DOUBAO_1_5_UI_TARS_250328).  // UI理解用UI-TARS
    WithAsserterModel(option.OPENAI_GPT_4O).             // 推理用GPT-4O
    WithQuerierModel(option.DEEPSEEK_R1_250528)          // 查询用DeepSeek

driver, err := uixt.NewXTDriver(mockDriver,
    option.WithLLMConfig(config))

// 方式3: 使用推荐配置
configs := option.RecommendedConfigurations()
driver, err := uixt.NewXTDriver(mockDriver,
    option.WithLLMConfig(configs["mixed_optimal"]))
```

### 推荐配置方案

| 配置名称 | 说明 | 适用场景 |
|---------|------|----------|
| `cost_effective` | 成本优化配置 | 预算有限的项目 |
| `high_performance` | 高性能配置（全部使用GPT-4O） | 对准确性要求极高的场景 |
| `mixed_optimal` | 混合优化配置 | 平衡性能和成本的最佳选择 |
| `ui_focused` | UI专注配置（全部使用UI-TARS） | UI自动化专项测试 |
| `reasoning_focused` | 推理专注配置（全部使用豆包思考模型） | 复杂逻辑推理场景 |

### 支持的模型

| 模型名称 | 特点 | 适用组件 |
|---------|------|----------|
| `DOUBAO_1_5_UI_TARS_250328` | UI理解专业模型 | Planner |
| `DOUBAO_1_5_THINKING_VISION_PRO_250428` | 思考推理模型 | Asserter, Querier |
| `OPENAI_GPT_4O` | 高性能通用模型 | 全部组件 |
| `DEEPSEEK_R1_250528` | 成本效益模型 | Querier |

## 🔧 环境配置

### 多模型配置

支持为不同模型配置独立的环境变量：

```bash
# 豆包思维视觉专业版
DOUBAO_1_5_THINKING_VISION_PRO_250428_BASE_URL=https://ark.cn-beijing.volces.com/api/v3
DOUBAO_1_5_THINKING_VISION_PRO_250428_API_KEY=your_doubao_api_key

# 豆包UI-TARS
DOUBAO_1_5_UI_TARS_250328_BASE_URL=https://ark.cn-beijing.volces.com/api/v3
DOUBAO_1_5_UI_TARS_250328_API_KEY=your_doubao_ui_tars_api_key

# OpenAI GPT-4O
OPENAI_GPT_4O_BASE_URL=https://api.openai.com/v1
OPENAI_GPT_4O_API_KEY=your_openai_api_key

# DeepSeek
DEEPSEEK_R1_250528_BASE_URL=https://api.deepseek.com/v1
DEEPSEEK_R1_250528_API_KEY=your_deepseek_api_key
```

### 默认配置

```bash
# 默认配置，当没有找到服务特定配置时使用
LLM_MODEL_NAME=doubao-1.5-thinking-vision-pro-250428
OPENAI_BASE_URL=https://ark.cn-beijing.volces.com/api/v3
OPENAI_API_KEY=your_default_api_key
```

### 配置优先级

1. **服务特定配置**（最高优先级）：`{SERVICE_NAME}_BASE_URL`、`{SERVICE_NAME}_API_KEY`
2. **默认配置**：`OPENAI_BASE_URL`、`OPENAI_API_KEY`、`LLM_MODEL_NAME`

## 🏗️ 核心架构

### 整体架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   UI Driver     │    │   AI Module     │    │  LLM Services   │
│   (XTDriver)    │◄──►│   (ai package)  │◄──►│ (多模型支持)     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │  CV Services    │
                       │   (VEDEM)       │
                       └─────────────────┘
```

### 核心接口

```go
// LLM 服务接口
type ILLMService interface {
    Plan(ctx context.Context, opts *PlanningOptions) (*PlanningResult, error)
    Assert(ctx context.Context, opts *AssertOptions) (*AssertionResult, error)
    Query(ctx context.Context, opts *QueryOptions) (*QueryResult, error)
    RegisterTools(tools []*schema.ToolInfo) error
}

// 计算机视觉服务接口
type ICVService interface {
    ReadFromBuffer(imageBuf *bytes.Buffer, opts ...option.ActionOption) (*CVResult, error)
    ReadFromPath(imagePath string, opts ...option.ActionOption) (*CVResult, error)
}
```

## 💡 功能详解

### 1. 智能规划 (Planning)

基于视觉语言模型进行 UI 操作规划，将自然语言指令转换为具体的操作序列。

```go
// 规划选项
type PlanningOptions struct {
    UserInstruction string          `json:"user_instruction"` // 用户指令
    Message         *schema.Message `json:"message"`          // 消息内容
    Size            types.Size      `json:"size"`             // 屏幕尺寸
    ResetHistory    bool            `json:"reset_history"`    // 是否重置历史
}

// 规划结果
type PlanningResult struct {
    ToolCalls []schema.ToolCall  `json:"tool_calls"` // 工具调用序列
    Thought   string             `json:"thought"`    // 思考过程
    Content   string             `json:"content"`    // 响应内容
    Error     string             `json:"error,omitempty"`
    ModelName string             `json:"model_name"`
    Usage     *schema.TokenUsage `json:"usage,omitempty"`
}
```

**使用示例**：
```go
planResult, err := service.Plan(ctx, &ai.PlanningOptions{
    UserInstruction: "点击登录按钮",
    Message:         message,
    Size:           screenSize,
})
```

### 2. 智能断言 (Assertion)

基于视觉语言模型进行断言验证，支持自然语言描述的断言条件。

```go
// 断言选项
type AssertOptions struct {
    Assertion  string     `json:"assertion"`  // 断言条件
    Screenshot string     `json:"screenshot"` // 屏幕截图
    Size       types.Size `json:"size"`       // 屏幕尺寸
}

// 断言结果
type AssertionResult struct {
    Pass    bool   `json:"pass"`    // 是否通过
    Thought string `json:"thought"` // 推理过程
}
```

**使用示例**：
```go
assertResult, err := service.Assert(ctx, &ai.AssertOptions{
    Assertion:  "登录按钮应该可见",
    Screenshot: screenshot,
    Size:       screenSize,
})
```

### 3. 智能查询 (Query)

从屏幕截图中提取结构化信息，支持自定义输出格式。

```go
// 查询选项
type QueryOptions struct {
    Query        string      `json:"query"`                    // 查询指令
    Screenshot   string      `json:"screenshot"`               // 屏幕截图
    Size         types.Size  `json:"size"`                     // 屏幕尺寸
    OutputSchema interface{} `json:"outputSchema,omitempty"`   // 自定义输出格式
}

// 查询结果
type QueryResult struct {
    Content string      `json:"content"`           // 文本内容
    Thought string      `json:"thought"`           // 思考过程
    Data    interface{} `json:"data,omitempty"`    // 结构化数据
}
```

**基础查询示例**：
```go
result, err := service.Query(ctx, &ai.QueryOptions{
    Query:      "请描述这张图片中的内容",
    Screenshot: screenshot,
    Size:       screenSize,
})
```

**自定义格式查询示例**：
```go
type GameInfo struct {
    Content string   `json:"content"`
    Thought string   `json:"thought"`
    Rows    int      `json:"rows"`
    Cols    int      `json:"cols"`
    Icons   []string `json:"icons"`
}

result, err := service.Query(ctx, &ai.QueryOptions{
    Query:        "分析这个连连看游戏界面",
    Screenshot:   screenshot,
    Size:         screenSize,
    OutputSchema: GameInfo{},
})

// 直接类型断言获取结构化数据
if gameInfo, ok := result.Data.(*GameInfo); ok {
    fmt.Printf("游戏有 %d 行 %d 列\n", gameInfo.Rows, gameInfo.Cols)
}
```

### 4. 计算机视觉 (CV)

提供 OCR 文本识别、UI 元素检测、弹窗识别等计算机视觉功能。

```go
// CV 结果
type CVResult struct {
    URL               string             `json:"url,omitempty"`
    OCRResult         OCRResults         `json:"ocrResult,omitempty"`
    LiveType          string             `json:"liveType,omitempty"`
    LivePopularity    int64              `json:"livePopularity,omitempty"`
    UIResult          UIResultMap        `json:"uiResult,omitempty"`
    ClosePopupsResult *ClosePopupsResult `json:"closeResult,omitempty"`
}
```

**使用示例**：
```go
cvService, err := ai.NewCVService(option.CVServiceTypeVEDEM)
cvResult, err := cvService.ReadFromBuffer(imageBuffer)

// 处理 OCR 结果
ocrTexts := cvResult.OCRResult.ToOCRTexts()
targetText, err := ocrTexts.FindText("登录", option.WithRegex(false))
center := targetText.Center()
```

## 🎨 高级特性

### 1. 多模型适配

不同模型具有不同的优势，可以根据场景选择最适合的模型：

- **UI-TARS**: 专门针对 UI 自动化优化，理解界面元素能力强
- **GPT-4O**: 通用性强，推理能力优秀
- **豆包思考模型**: 支持深度思考，适合复杂场景分析
- **DeepSeek**: 成本效益高，适合大量查询场景

### 2. 坐标系统转换

支持多种坐标格式的智能转换：

- 相对坐标 (0-1000 范围) 转换为绝对像素坐标
- 支持 `<point>`、`<bbox>`、`[x,y,x,y]` 等多种格式
- 自动处理不同模型的坐标输出差异

### 3. 智能会话管理

- **对话历史**: 维护完整的对话上下文
- **内存优化**: 自动清理过期的对话记录
- **消息管理**: 智能管理用户图像消息和助手回复

### 4. 自定义输出格式

查询功能支持用户定义的复杂结构化输出格式：

```go
type UIAnalysisResult struct {
    Content    string      `json:"content"`
    Elements   []UIElement `json:"elements"`
    Statistics Statistics  `json:"statistics"`
}

type UIElement struct {
    Type        string      `json:"type"`
    Text        string      `json:"text"`
    BoundingBox BoundingBox `json:"boundingBox"`
    Clickable   bool        `json:"clickable"`
}
```

## 📋 配置参数

### 模型配置

| 参数 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| `BaseURL` | string | API 基础 URL | 从环境变量读取 |
| `APIKey` | string | API 密钥 | 从环境变量读取 |
| `Model` | string | 模型名称 | 从环境变量读取 |
| `Temperature` | float32 | 温度参数 | 0 |
| `TopP` | float32 | Top-P 参数 | 0.7 |
| `Timeout` | time.Duration | 请求超时 | 30s |

### 操作选项

| 组件 | 必需参数 | 可选参数 |
|------|----------|----------|
| **Planner** | `UserInstruction`, `Message`, `Size` | `ResetHistory` |
| **Asserter** | `Assertion`, `Screenshot`, `Size` | - |
| **Querier** | `Query`, `Screenshot`, `Size` | `OutputSchema` |

## ⚠️ 注意事项

### 1. 环境配置
- 确保所有必需的环境变量都已正确设置
- API 密钥需要有足够的权限和配额
- 支持多模型配置，可以同时配置多个服务

### 2. 图像格式
- 支持 Base64 编码的图像数据
- 推荐使用 JPEG 格式以减少数据传输量
- 图像尺寸信息必须准确提供

### 3. 坐标系统
- 不同模型使用不同的坐标系统
- 需要正确的屏幕尺寸信息进行坐标转换
- 系统会自动处理坐标格式差异

### 4. 性能考虑
- LLM 调用有延迟，适合异步处理
- 图像数据较大，注意网络传输优化
- 对话历史会占用内存，系统会自动清理

### 5. 错误处理
- 网络请求可能失败，需要适当的重试机制
- 模型输出格式可能不稳定，系统提供健壮的解析逻辑
- 建议在生产环境中添加监控和告警

## 🧪 测试数据

模块包含丰富的测试数据，位于 `testdata/` 目录：

- `xhs-feed.jpeg`: 小红书信息流界面
- `popup_risk_warning.png`: 风险警告弹窗
- `llk_*.png`: 连连看游戏界面
- `deepseek_*.png`: DeepSeek 应用界面
- `chat_list.jpeg`: 聊天列表界面

这些测试数据覆盖了各种典型的 UI 场景，用于验证 AI 模块的功能正确性。

## 🚀 快速开始

1. **配置环境变量**
   ```bash
   # 配置默认模型
   export OPENAI_BASE_URL=https://your-endpoint.com
   export OPENAI_API_KEY=your-api-key
   ```

2. **创建驱动**
   ```go
   driver, err := uixt.NewXTDriver(mockDriver,
       option.WithLLMService(option.DOUBAO_1_5_THINKING_VISION_PRO_250428))
   ```

3. **执行智能操作**
   ```go
   // 智能规划
   planResult, err := driver.LLMService.Plan(ctx, planningOpts)

   // 智能断言
   assertResult, err := driver.LLMService.Assert(ctx, assertOpts)

   // 智能查询
   queryResult, err := driver.LLMService.Query(ctx, queryOpts)
   ```

通过 HttpRunner UIXT AI 模块，您可以轻松实现智能化的 UI 自动化测试，大幅提升测试效率和准确性。