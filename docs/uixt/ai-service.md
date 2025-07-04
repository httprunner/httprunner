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

#### 高级查询场景

**UI 元素分析**：
```go
type UIAnalysis struct {
    Content  string      `json:"content"`
    Thought  string      `json:"thought"`
    Elements []UIElement `json:"elements"`
}

type UIElement struct {
    Type      string      `json:"type"`        // button, text, input等
    Text      string      `json:"text"`        // 文本内容
    BoundBox  BoundingBox `json:"boundBox"`    // 位置坐标
    Clickable bool        `json:"clickable"`   // 是否可点击
}

result, err := service.Query(ctx, &ai.QueryOptions{
    Query: `分析这张截图并提供结构化信息：
1. 识别界面类型和主要元素
2. 提取所有可交互元素的位置和属性
3. 统计各类元素的数量`,
    Screenshot:   screenshot,
    Size:         screenSize,
    OutputSchema: UIAnalysis{},
})
```

**网格游戏分析**：
```go
type GridGame struct {
    Content string     `json:"content"`
    Thought string     `json:"thought"`
    Grid    [][]Cell   `json:"grid"`       // 网格数据
    Stats   Statistics `json:"statistics"` // 统计信息
}

type Cell struct {
    Type  string `json:"type"`  // 单元格类型
    Value string `json:"value"` // 单元格值
    Row   int    `json:"row"`   // 行索引
    Col   int    `json:"col"`   // 列索引
}

result, err := service.Query(ctx, &ai.QueryOptions{
    Query:        "分析这个网格游戏的布局和状态",
    Screenshot:   screenshot,
    Size:         screenSize,
    OutputSchema: GridGame{},
})
```

**表单数据提取**：
```go
type FormAnalysis struct {
    Content string      `json:"content"`
    Thought string      `json:"thought"`
    Fields  []FormField `json:"fields"`
    Actions []Action    `json:"actions"`
}

type FormField struct {
    Label    string      `json:"label"`    // 字段标签
    Type     string      `json:"type"`     // 字段类型
    Value    string      `json:"value"`    // 当前值
    Required bool        `json:"required"` // 是否必填
    BoundBox BoundingBox `json:"boundBox"` // 位置
}

result, err := service.Query(ctx, &ai.QueryOptions{
    Query:        "提取表单中的所有字段信息和操作按钮",
    Screenshot:   screenshot,
    Size:         screenSize,
    OutputSchema: FormAnalysis{},
})
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

查询功能支持用户定义的复杂结构化输出格式，具有以下核心特性：

#### 自动类型转换
- 指定 `OutputSchema` 时，`QueryResult.Data` 自动转换为指定类型
- 支持直接类型断言：`result.Data.(*YourType)`
- 无需手动调用转换函数

#### 多级回退机制
1. 优先解析为指定的结构化类型
2. 失败时尝试通用JSON解析
3. 最终回退到纯文本响应

#### 向后兼容
- 不指定 `OutputSchema` 时行为不变
- 现有代码无需修改

**结构体设计最佳实践**：
```go
// 推荐：包含标准字段
type YourSchema struct {
    Content string `json:"content"` // 必须：人类可读描述
    Thought string `json:"thought"` // 必须：AI推理过程
    // 自定义字段...
    Data    CustomData `json:"data"`
}

// 使用描述性的JSON标签
type Element struct {
    Type     string `json:"elementType"`   // 清晰的字段名
    Position Point  `json:"gridPosition"`  // 描述性标签
    Visible  bool   `json:"isVisible"`     // 布尔值清晰性
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

# AI 功能使用指南

HttpRunner v5 提供了强大的 AI 功能，支持基于视觉语言模型（VLM）的智能化测试操作。

## 功能概述

HttpRunner v5 集成了多种 AI 功能：

- **AIAction**: 使用自然语言执行 UI 操作
- **AIAssert**: 使用自然语言进行断言验证
- **AIQuery**: 使用自然语言从屏幕中提取信息
- **StartToGoal**: 目标导向的智能操作序列

## AIQuery 功能详解

### 概述

AIQuery 是 HttpRunner v5 中新增的 AI 查询功能，允许用户使用自然语言从屏幕截图中提取信息。它基于视觉语言模型（VLM），能够理解屏幕内容并返回结构化的查询结果。

### 功能特点

- **自然语言查询**: 使用自然语言描述要查询的信息
- **智能屏幕分析**: 基于 AI 视觉模型分析屏幕内容
- **结构化输出**: 返回格式化的查询结果
- **多平台支持**: 支持 Android、iOS、Browser 等平台

### 基本用法

#### 1. 在测试步骤中使用 AIQuery

```go
// 基本查询示例
hrp.NewStep("Query Screen Content").
    Android().
    AIQuery("Please describe what is displayed on the screen")

// 提取特定信息
hrp.NewStep("Extract App List").
    Android().
    AIQuery("What apps are visible on the home screen? List them as a comma-separated string")

// UI 元素分析
hrp.NewStep("Analyze Buttons").
    Android().
    AIQuery("Are there any buttons visible? Describe their text and positions")
```

#### 2. 配置 LLM 服务

在使用 AIQuery 之前，需要配置 LLM 服务：

```go
testcase := &hrp.TestCase{
    Config: hrp.NewConfig("AIQuery Test").
        SetLLMService(option.OPENAI_GPT_4O), // 配置 LLM 服务
    TestSteps: []hrp.IStep{
        // 使用 AIQuery 的步骤
    },
}
```

#### 3. 支持的选项

AIQuery 支持以下选项：

```go
hrp.NewStep("Query with Options").
    Android().
    AIQuery("Describe the screen content",
        option.WithLLMService("openai_gpt_4o"),  // 指定 LLM 服务
        option.WithCVService("openai_gpt_4o"),   // 指定 CV 服务
        option.WithOutputSchema(CustomSchema{}), // 自定义输出格式
    )
```

#### 4. 自定义输出格式 (OutputSchema)

AIQuery 支持自定义输出格式，可以返回结构化数据：

```go
// 定义自定义输出格式
type GameAnalysis struct {
    Content     string   `json:"content"`     // 必须：人类可读描述
    Thought     string   `json:"thought"`     // 必须：AI推理过程
    GameType    string   `json:"game_type"`   // 游戏类型
    Rows        int      `json:"rows"`        // 行数
    Cols        int      `json:"cols"`        // 列数
    Icons       []string `json:"icons"`       // 图标类型
    TotalIcons  int      `json:"total_icons"` // 图标总数
}

// 使用自定义格式查询
hrp.NewStep("Analyze Game Interface").
    Android().
    AIQuery("分析这个连连看游戏界面，告诉我有多少行多少列，有哪些不同类型的图案",
        option.WithOutputSchema(GameAnalysis{}))
```

### 实际应用场景

#### 1. 游戏界面分析

```go
// 分析连连看游戏界面
hrp.NewStep("Analyze Game Board").
    Android().
    AIQuery("This is a LianLianKan (连连看) game interface. Please analyze: 1) How many rows and columns are there? 2) What types of icons are present?")
```

#### 2. 应用状态检查

```go
// 检查应用状态
hrp.NewStep("Check App State").
    Android().
    AIQuery("Is the login screen displayed? Are there any error messages visible?")
```

#### 3. 内容提取

```go
// 提取列表内容
hrp.NewStep("Extract List Items").
    Android().
    AIQuery("Extract all items from the list displayed on screen as a JSON array")
```

### 与其他 AI 功能的对比

| 功能 | 用途 | 返回值 | 使用场景 |
|------|------|--------|----------|
| AIAction | 执行操作 | 无 | 点击、输入、滑动等交互操作 |
| AIAssert | 断言验证 | 布尔值 | 验证界面状态、元素存在性 |
| AIQuery | 信息查询 | 字符串 | 提取屏幕信息、分析内容 |

### 最佳实践

#### 1. 明确的查询描述

```go
// 好的示例：具体明确
AIQuery("How many unread messages are shown in the notification badge?")

// 避免：过于模糊
AIQuery("Tell me about the screen")
```

#### 2. 结构化查询

```go
// 请求结构化输出
AIQuery("List all visible buttons with their text and approximate positions in JSON format")
```

#### 3. 上下文相关查询

```go
// 结合应用上下文
AIQuery("In this shopping app, what products are displayed in the current category? Include product names and prices")
```

### 错误处理

AIQuery 可能遇到的常见错误：

1. **LLM 服务未配置**: 确保在测试配置中设置了 LLM 服务
2. **网络连接问题**: 检查网络连接和 API 密钥配置
3. **屏幕截图失败**: 确保设备连接正常

### 注意事项

1. AIQuery 需要网络连接来访问 LLM 服务
2. 查询结果的准确性依赖于所使用的 LLM 模型
3. 建议在查询中使用具体、明确的描述以获得更好的结果
4. 对于复杂的信息提取，可以要求返回 JSON 格式的结构化数据

## StartToGoal 功能详解

### 概述

`StartToGoal` 是 HttpRunner v5 中的目标导向智能操作功能，它使用自然语言描述目标，然后自动规划和执行一系列操作来达成目标。该功能基于视觉语言模型（VLM）进行智能规划，能够理解屏幕内容并自动生成操作序列。

### 功能特点

- **目标导向**: 使用自然语言描述最终目标，AI 自动规划操作步骤
- **智能规划**: 基于屏幕内容进行上下文相关的操作规划
- **自动执行**: 自动执行规划的操作序列直到达成目标
- **灵活控制**: 支持多种控制选项如重试次数、超时时间等

### 基本用法

#### 1. 基本示例

```go
// 基本目标导向操作
results, err := driver.StartToGoal(ctx, "导航到设置页面并启用深色模式")

// 带选项的目标导向操作
results, err := driver.StartToGoal(ctx, "登录应用",
    option.WithMaxRetryTimes(3),
    option.WithIdentifier("user-login"),
)
```

#### 2. 在测试步骤中使用

```go
hrp.NewStep("Navigate to Settings").
    Android().
    StartToGoal("打开设置页面")

hrp.NewStep("Enable Feature").
    Android().
    StartToGoal("启用深色模式功能",
        option.WithMaxRetryTimes(3),
        option.WithIdentifier("enable-dark-mode"),
    )
```

### TimeLimit 时间限制功能

`StartToGoal` 支持 `TimeLimit` 选项，用于设置执行时间限制。这是一个重要的资源管理功能。

#### 功能特性

- **时间限制**: 支持设置执行时间上限（秒）
- **优雅停止**: 超出时间限制后停止执行，但返回成功状态
- **部分结果**: 即使达到时间限制，也会返回已完成的规划结果

#### 使用方法

##### 基本用法

```go
// 设置 30 秒时间限制
results, err := driver.StartToGoal(ctx, prompt, option.WithTimeLimit(30))
```

##### 与其他选项结合使用

```go
results, err := driver.StartToGoal(ctx, prompt,
    option.WithTimeLimit(45),           // 45秒时间限制
    option.WithMaxRetryTimes(3),        // 最大重试3次
    option.WithIdentifier("my-task"),   // 任务标识符
)
```

#### TimeLimit vs Timeout

| 特性 | TimeLimit | Timeout | Interrupt Signal |
|------|-----------|---------|------------------|
| 行为 | 优雅停止 | 强制取消 | 立即中断 |
| 返回值 | 成功 (err == nil) | 错误 (err != nil) | 错误 (err != nil) |
| 结果 | 返回部分结果 | 返回部分结果 | 返回部分结果 |
| 用途 | 资源管理，时间预算 | 防止无限等待 | 用户主动中断 |
| 优先级 | 中等 | 低 | 最高 |

#### 使用场景

##### 使用 TimeLimit 的场景：
- 需要在指定时间内完成尽可能多的任务
- 资源管理和时间预算控制
- 希望获得部分结果而不是完全失败
- 测试场景下的时间控制

##### 使用 Timeout 的场景：
- 防止无限等待
- 超时即视为失败的场景
- 需要严格的时间控制

##### Interrupt Signal 的特点：
- 用户主动中断（Ctrl+C）
- 优先级最高，立即生效
- 无论是否设置 TimeLimit，都返回错误
- 适用于需要立即停止的场景

#### 实现原理

1. **Context 复用**: `TimeLimit` 和 `Timeout` 复用相同的 context 超时机制
2. **模式标记**: 通过 `isTimeLimitMode` 标记区分当前是时间限制模式还是超时模式
3. **优先级处理**: 在 `ctx.Done()` 时按优先级检查取消原因
4. **结果收集**: 返回所有已完成的规划结果

**技术实现**：
```go
// 复用 timeout context 机制，用标记区分模式
var isTimeLimitMode bool
if options.TimeLimit > 0 {
    ctx, cancel = context.WithTimeout(ctx, time.Duration(options.TimeLimit)*time.Second)
    isTimeLimitMode = true
} else if options.Timeout > 0 {
    ctx, cancel = context.WithTimeout(ctx, time.Duration(options.Timeout)*time.Second)
}

// 按优先级检查取消原因
select {
case <-ctx.Done():
    cause := context.Cause(ctx)
    // 1. 中断信号优先级最高，始终返回错误
    if errors.Is(cause, code.InterruptError) {
        return allPlannings, errors.Wrap(cause, "StartToGoal interrupted")
    }
    // 2. TimeLimit 超时返回成功
    if isTimeLimitMode && errors.Is(cause, context.DeadlineExceeded) {
        return allPlannings, nil
    }
    // 3. 其他取消原因返回错误
    return allPlannings, errors.Wrap(cause, "StartToGoal cancelled")
}
```

#### 注意事项

1. **检测精度**: 时间限制的检测精度依赖于规划和工具调用的频率，基于 Go context 机制更加精确
2. **资源清理**: 即使达到时间限制，也会完成当前操作以确保资源正确清理
3. **结果可用性**: 返回的结果包含会话数据，可用于生成报告
4. **Context 复用**: `TimeLimit` 和 `Timeout` 复用相同的 context 超时机制，简化了实现
5. **优先级**: 如果同时设置了 `TimeLimit` 和 `Timeout`，`TimeLimit` 优先生效
6. **中断信号**: 用户中断信号（如 Ctrl+C）优先级最高，无论是否设置 `TimeLimit` 都会返回错误

### 支持的选项

`StartToGoal` 支持多种控制选项：

```go
// 全面的选项示例
results, err := driver.StartToGoal(ctx, prompt,
    option.WithTimeLimit(60),           // 时间限制（秒）
    option.WithTimeout(120),            // 超时时间（秒）
    option.WithMaxRetryTimes(5),        // 最大重试次数
    option.WithIdentifier("task-id"),   // 任务标识符
    option.WithLLMService("gpt-4o"),    // LLM 服务
    option.WithCVService("vedem"),      // CV 服务
    option.WithResetHistory(true),      // 重置对话历史
)
```

### 最佳实践

#### 1. 明确的目标描述

```go
// 好的示例：具体明确
StartToGoal("打开设置页面，找到显示选项，然后启用深色模式")

// 避免：过于模糊
StartToGoal("做一些设置")
```

#### 2. 合理的时间限制

```go
// 根据任务复杂度设置合理的时间限制
StartToGoal("完成用户注册流程", option.WithTimeLimit(120)) // 复杂任务
StartToGoal("点击登录按钮", option.WithTimeLimit(30))     // 简单任务
```

#### 3. 错误处理和重试

```go
// 设置重试机制
results, err := driver.StartToGoal(ctx, prompt,
    option.WithMaxRetryTimes(3),
    option.WithTimeLimit(90),
)

if err != nil {
    // 处理错误
    log.Printf("StartToGoal failed: %v", err)
    // 可以分析 results 中的部分结果
}
```

### 实际应用场景

#### 1. 复杂的操作流程

```go
// 完成整个购物流程
hrp.NewStep("Complete Purchase").
    Android().
    StartToGoal("搜索商品'手机'，选择第一个商品，添加到购物车，然后结账",
        option.WithTimeLimit(180),
        option.WithMaxRetryTimes(2),
    )
```

#### 2. 应用初始化设置

```go
// 首次使用应用的设置流程
hrp.NewStep("Initial Setup").
    Android().
    StartToGoal("跳过引导页，允许所有权限，然后进入主界面",
        option.WithTimeLimit(60),
    )
```

#### 3. 测试场景验证

```go
// 验证特定功能流程
hrp.NewStep("Verify Feature").
    Android().
    StartToGoal("验证分享功能是否正常工作",
        option.WithTimeLimit(45),
        option.WithIdentifier("share-test"),
    )
```

### 返回结果

`StartToGoal` 返回 `PlanningExecutionResult` 数组，包含详细的执行信息：

```go
type PlanningExecutionResult struct {
    PlanningResult ai.PlanningResult `json:"planning_result"`
    SubActions     []*SubActionResult `json:"sub_actions"`
    StartTime      int64             `json:"start_time"`
    Elapsed        int64             `json:"elapsed"`
}
```

可以通过返回结果分析执行过程：

```go
results, err := driver.StartToGoal(ctx, prompt, option.WithTimeLimit(60))
if err != nil {
    log.Printf("Task failed: %v", err)
}

// 分析执行结果
for i, result := range results {
    log.Printf("Planning %d: %s", i+1, result.PlanningResult.Thought)
    log.Printf("Actions executed: %d", len(result.SubActions))
    log.Printf("Elapsed time: %d ms", result.Elapsed)
}
```

## 完整示例

以下是一个完整的 AIQuery 使用示例：

```go
func TestAIQuery(t *testing.T) {
    testCase := &hrp.TestCase{
        Config: hrp.NewConfig("AIQuery Demo").
            SetLLMService(option.OPENAI_GPT_4O),
        TestSteps: []hrp.IStep{
            hrp.NewStep("Take Screenshot").
                Android().
                ScreenShot(),
            hrp.NewStep("Query Screen Content").
                Android().
                AIQuery("Please describe what is displayed on the screen and identify any interactive elements"),
            hrp.NewStep("Extract App Information").
                Android().
                AIQuery("What apps are visible on the screen? List them as a comma-separated string"),
            hrp.NewStep("Analyze UI Elements").
                Android().
                AIQuery("Are there any buttons or clickable elements visible? Describe their locations and purposes"),
        },
    }

    err := hrp.NewRunner(t).Run(testCase)
    assert.Nil(t, err)
}
```

## StartToGoal 完整示例

以下是 `StartToGoal` 功能的完整使用示例：

```go
func TestStartToGoal(t *testing.T) {
    testCase := &hrp.TestCase{
        Config: hrp.NewConfig("StartToGoal Demo").
            SetLLMService(option.OPENAI_GPT_4O),
        TestSteps: []hrp.IStep{
            hrp.NewStep("App Launch").
                Android().
                AppLaunch("com.example.app"),
            hrp.NewStep("Complete User Setup").
                Android().
                StartToGoal("跳过引导页，创建新用户账户",
                    option.WithTimeLimit(120),
                    option.WithMaxRetryTimes(3),
                ),
            hrp.NewStep("Navigate to Feature").
                Android().
                StartToGoal("导航到设置页面并启用深色模式",
                    option.WithTimeLimit(60),
                    option.WithIdentifier("enable-dark-mode"),
                ),
            hrp.NewStep("Complex Workflow").
                Android().
                StartToGoal("搜索'测试'，选择第一个结果，然后分享给朋友",
                    option.WithTimeLimit(180),
                    option.WithMaxRetryTimes(2),
                ),
        },
    }

    err := hrp.NewRunner(t).Run(testCase)
    assert.Nil(t, err)
}
```