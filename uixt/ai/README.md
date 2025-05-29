# HttpRunner AI 模块文档

## 📖 概述

HttpRunner AI 模块是一个集成了多种人工智能服务的 UI 自动化智能引擎，提供基于大语言模型（LLM）的智能规划、断言验证、计算机视觉识别等功能，实现真正的智能化 UI 自动化测试。

## 🎯 核心功能

### 1. 智能规划 (Planning)
- **视觉语言模型驱动**: 基于屏幕截图和自然语言指令生成操作序列
- **多模型支持**: 支持 UI-TARS、豆包视觉等多种专业模型
- **上下文感知**: 维护对话历史，支持多轮交互规划
- **动作解析**: 将模型输出解析为标准化的工具调用

### 2. 智能断言 (Assertion)
- **视觉验证**: 基于屏幕截图验证断言条件
- **自然语言断言**: 支持自然语言描述的断言条件
- **结构化输出**: 返回标准化的断言结果和推理过程

### 3. 计算机视觉 (Computer Vision)
- **OCR 文本识别**: 提取屏幕中的文本内容和位置信息
- **UI 元素检测**: 识别界面中的图标、按钮等 UI 元素
- **弹窗检测**: 自动识别和定位弹窗及关闭按钮
- **坐标转换**: 支持相对坐标和绝对坐标的转换

### 4. 会话管理 (Session Management)
- **对话历史**: 维护完整的对话上下文
- **消息管理**: 智能管理用户图像消息和助手回复
- **历史清理**: 自动清理过期的对话记录

## 🏗️ 架构设计

### 整体架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   UI Driver     │    │   AI Module     │    │  LLM Services   │
│   (XTDriver)    │◄──►│   (ai package)  │◄──►│ (OpenAI/豆包)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │  CV Services    │
                       │   (VEDEM)       │
                       └─────────────────┘
```

### 核心接口

#### ILLMService - LLM 服务接口
```go
type ILLMService interface {
    Call(ctx context.Context, opts *PlanningOptions) (*PlanningResult, error)
    Assert(ctx context.Context, opts *AssertOptions) (*AssertionResult, error)
}
```

#### IPlanner - 规划器接口
```go
type IPlanner interface {
    Call(ctx context.Context, opts *PlanningOptions) (*PlanningResult, error)
}
```

#### IAsserter - 断言器接口
```go
type IAsserter interface {
    Assert(ctx context.Context, opts *AssertOptions) (*AssertionResult, error)
}
```

#### ICVService - 计算机视觉服务接口
```go
type ICVService interface {
    ReadFromBuffer(imageBuf *bytes.Buffer, opts ...option.ActionOption) (*CVResult, error)
    ReadFromPath(imagePath string, opts ...option.ActionOption) (*CVResult, error)
}
```

## 🔧 主要组件

### 1. AI 服务管理器 (ai.go)

**功能**: 统一管理 LLM 服务，提供规划和断言功能的统一入口

**核心类型**:
```go
type combinedLLMService struct {
    planner  IPlanner  // 提供规划功能
    asserter IAsserter // 提供断言功能
}

type ModelConfig struct {
    *openai.ChatModelConfig
    ModelType option.LLMServiceType
}
```

**主要功能**:
- 模型配置管理和验证
- 环境变量读取和验证
- API 密钥安全处理
- 多模型类型支持

**支持的模型类型**:
- `LLMServiceTypeUITARS`: UI-TARS 专业 UI 自动化模型
- `LLMServiceTypeDoubaoVL`: 豆包视觉语言模型

### 2. 智能规划器 (planner.go)

**功能**: 基于视觉语言模型进行 UI 操作规划

**核心类型**:
```go
type Planner struct {
    modelConfig *ModelConfig
    model       model.ToolCallingChatModel
    parser      LLMContentParser
    history     ConversationHistory
}

type PlanningOptions struct {
    UserInstruction string          `json:"user_instruction"`
    Message         *schema.Message `json:"message"`
    Size            types.Size      `json:"size"`
}

type PlanningResult struct {
    ToolCalls     []schema.ToolCall `json:"tool_calls"`
    ActionSummary string            `json:"summary"`
    Thought       string            `json:"thought"`
    Content       string            `json:"content"`
    Error         string            `json:"error,omitempty"`
}
```

**工作流程**:
1. 接收用户指令和屏幕截图
2. 构建包含系统提示词的对话历史
3. 调用视觉语言模型生成响应
4. 解析模型输出为标准化工具调用
5. 更新对话历史以支持多轮交互

**特性**:
- 支持工具注册和函数调用
- 智能对话历史管理
- 多种输出格式解析
- 详细的日志记录

### 3. 智能断言器 (asserter.go)

**功能**: 基于视觉语言模型进行断言验证

**核心类型**:
```go
type Asserter struct {
    modelConfig  *ModelConfig
    model        model.ToolCallingChatModel
    systemPrompt string
    history      ConversationHistory
}

type AssertOptions struct {
    Assertion  string     `json:"assertion"`
    Screenshot string     `json:"screenshot"`
    Size       types.Size `json:"size"`
}

type AssertionResult struct {
    Pass    bool   `json:"pass"`
    Thought string `json:"thought"`
}
```

**工作流程**:
1. 接收断言条件和屏幕截图
2. 构建断言验证提示词
3. 调用视觉语言模型进行判断
4. 解析模型输出为结构化结果
5. 返回断言通过状态和推理过程

**特性**:
- 结构化 JSON 输出格式
- 自然语言断言支持
- 详细的推理过程记录
- 多模型适配

### 4. 内容解析器 (parser_*.go)

**功能**: 将不同模型的输出解析为标准化的工具调用格式

#### JSONContentParser (parser_default.go)
- 适用于支持 JSON 格式输出的通用模型
- 解析标准 JSON 格式的动作序列
- 支持坐标归一化和参数处理

#### UITARSContentParser (parser_ui_tars.go)
- 专门适配 UI-TARS 模型的 Thought/Action 格式
- 支持多种坐标格式解析 (`<point>`, `<bbox>`, `[x,y,x,y]`)
- 智能参数名称映射和归一化
- 相对坐标到绝对坐标转换

**核心功能**:
```go
type LLMContentParser interface {
    SystemPrompt() string
    Parse(content string, size types.Size) (*PlanningResult, error)
}

type Action struct {
    ActionType   string         `json:"action_type"`
    ActionInputs map[string]any `json:"action_inputs"`
}
```

**解析特性**:
- 多种坐标格式支持
- 智能参数映射
- 坐标系统转换
- 错误处理和验证

### 5. 计算机视觉服务 (cv.go)

**功能**: 提供图像识别和分析能力

**核心类型**:
```go
type CVResult struct {
    URL               string             `json:"url,omitempty"`
    OCRResult         OCRResults         `json:"ocrResult,omitempty"`
    LiveType          string             `json:"liveType,omitempty"`
    LivePopularity    int64              `json:"livePopularity,omitempty"`
    UIResult          UIResultMap        `json:"uiResult,omitempty"`
    ClosePopupsResult *ClosePopupsResult `json:"closeResult,omitempty"`
}

type OCRText struct {
    Text    string          `json:"text"`
    RectStr string          `json:"rect"`
    Rect    image.Rectangle `json:"-"`
}

type UIResult struct {
    Box
}

type ClosePopupsResult struct {
    Type      string `json:"type"`
    PopupArea Box    `json:"popupArea"`
    CloseArea Box    `json:"closeArea"`
    Text      string `json:"text"`
}
```

**主要功能**:
- **OCR 文本识别**: 提取文本内容和精确位置
- **UI 元素检测**: 识别按钮、图标等界面元素
- **弹窗检测**: 自动识别弹窗和关闭按钮
- **区域过滤**: 支持指定区域的元素筛选
- **坐标计算**: 提供中心点和随机点计算

**OCR 功能特性**:
- 文本精确定位
- 正则表达式匹配
- 索引选择支持
- 区域范围过滤

### 6. 会话管理器 (session.go)

**功能**: 管理 AI 对话的历史记录和上下文

**核心类型**:
```go
type ConversationHistory []*schema.Message
```

**管理策略**:
- **用户消息**: 最多保留 4 条用户图像消息
- **助手消息**: 最多保留 10 条助手回复
- **自动清理**: 超出限制时自动删除最旧的消息
- **系统消息**: 始终保留系统提示词

**功能特性**:
- 智能消息管理
- 内存优化
- 日志记录和调试
- 敏感信息脱敏

## 🚀 使用指南

### 1. 环境配置

设置必要的环境变量：

```bash
export OPENAI_BASE_URL="https://your-api-endpoint"
export OPENAI_API_KEY="your-api-key"
export LLM_MODEL_NAME="your-model-name"
```

### 2. 创建 LLM 服务

```go
// 创建 UI-TARS 服务
llmService, err := ai.NewLLMService(option.LLMServiceTypeUITARS)
if err != nil {
    log.Fatal().Err(err).Msg("failed to create LLM service")
}

// 创建豆包视觉服务
llmService, err := ai.NewLLMService(option.LLMServiceTypeDoubaoVL)
if err != nil {
    log.Fatal().Err(err).Msg("failed to create LLM service")
}
```

### 3. 智能规划使用

```go
// 准备规划选项
planningOpts := &ai.PlanningOptions{
    UserInstruction: "点击登录按钮",
    Message: &schema.Message{
        Role: schema.User,
        MultiContent: []schema.ChatMessagePart{
            {
                Type: schema.ChatMessagePartTypeImageURL,
                ImageURL: &schema.ChatMessageImageURL{
                    URL: "data:image/jpeg;base64," + base64Screenshot,
                },
            },
        },
    },
    Size: types.Size{Width: 1080, Height: 1920},
}

// 执行规划
result, err := llmService.Call(ctx, planningOpts)
if err != nil {
    log.Error().Err(err).Msg("planning failed")
    return
}

// 处理规划结果
for _, toolCall := range result.ToolCalls {
    log.Info().Str("action", toolCall.Function.Name).
        Interface("args", toolCall.Function.Arguments).
        Msg("planned action")
}
```

### 4. 智能断言使用

```go
// 准备断言选项
assertOpts := &ai.AssertOptions{
    Assertion:  "登录按钮应该可见",
    Screenshot: "data:image/jpeg;base64," + base64Screenshot,
    Size:       types.Size{Width: 1080, Height: 1920},
}

// 执行断言
result, err := llmService.Assert(ctx, assertOpts)
if err != nil {
    log.Error().Err(err).Msg("assertion failed")
    return
}

// 检查断言结果
if result.Pass {
    log.Info().Str("thought", result.Thought).Msg("assertion passed")
} else {
    log.Warn().Str("thought", result.Thought).Msg("assertion failed")
}
```

### 5. 计算机视觉使用

```go
// 创建 CV 服务
cvService, err := ai.NewCVService(option.CVServiceTypeVEDEM)
if err != nil {
    log.Fatal().Err(err).Msg("failed to create CV service")
}

// 从图像缓冲区读取
cvResult, err := cvService.ReadFromBuffer(imageBuffer)
if err != nil {
    log.Error().Err(err).Msg("CV analysis failed")
    return
}

// 处理 OCR 结果
ocrTexts := cvResult.OCRResult.ToOCRTexts()
for _, ocrText := range ocrTexts {
    log.Info().Str("text", ocrText.Text).
        Str("rect", ocrText.RectStr).
        Msg("found text")
}

// 查找特定文本
targetText, err := ocrTexts.FindText("登录", option.WithRegex(false))
if err != nil {
    log.Error().Err(err).Msg("text not found")
    return
}

// 获取文本中心点
center := targetText.Center()
log.Info().Float64("x", center.X).Float64("y", center.Y).
    Msg("text center coordinates")
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

### 规划选项

| 参数 | 类型 | 说明 | 必需 |
|------|------|------|------|
| `UserInstruction` | string | 用户指令 | ✓ |
| `Message` | *schema.Message | 消息内容 | ✓ |
| `Size` | types.Size | 屏幕尺寸 | ✓ |

### 断言选项

| 参数 | 类型 | 说明 | 必需 |
|------|------|------|------|
| `Assertion` | string | 断言条件 | ✓ |
| `Screenshot` | string | Base64 截图 | ✓ |
| `Size` | types.Size | 屏幕尺寸 | ✓ |

## 🔍 高级特性

### 1. 多模型适配

AI 模块支持多种不同的语言模型，每种模型都有其特定的优势：

- **UI-TARS**: 专门针对 UI 自动化优化的模型，支持 Thought/Action 格式
- **豆包视觉**: 通用视觉语言模型，支持结构化 JSON 输出

### 2. 坐标系统转换

支持多种坐标格式的智能转换：

```go
// 相对坐标 (0-1000 范围) 转换为绝对像素坐标
func convertRelativeToAbsolute(relativeCoord float64, isXCoord bool, size types.Size) float64 {
    if isXCoord {
        return math.Round((relativeCoord/DefaultFactor*float64(size.Width))*10) / 10
    }
    return math.Round((relativeCoord/DefaultFactor*float64(size.Height))*10) / 10
}
```

### 3. 智能参数映射

自动处理不同模型输出格式的参数名称映射：

```go
func normalizeParameterName(paramName string) string {
    switch paramName {
    case "start_point":
        return "start_box"
    case "end_point":
        return "end_box"
    case "point":
        return "start_box"
    default:
        return paramName
    }
}
```

### 4. 对话历史优化

智能管理对话历史，平衡上下文完整性和内存使用：

- 用户图像消息限制：4 条
- 助手回复消息限制：10 条
- 自动清理策略：FIFO (先进先出)

## ⚠️ 注意事项

### 1. 环境变量配置
- 确保所有必需的环境变量都已正确设置
- API 密钥需要有足够的权限和配额
- 模型名称必须与服务类型匹配

### 2. 图像格式要求
- 支持 Base64 编码的图像数据
- 推荐使用 JPEG 格式以减少数据传输量
- 图像尺寸信息必须准确提供

### 3. 坐标系统
- UI-TARS 使用 1000x1000 相对坐标系统
- 需要正确的屏幕尺寸信息进行坐标转换
- 注意不同模型的坐标格式差异

### 4. 错误处理
- 网络请求可能失败，需要适当的重试机制
- 模型输出格式可能不稳定，需要健壮的解析逻辑
- 资源使用需要监控，避免内存泄漏

### 5. 性能考虑
- LLM 调用有延迟，适合异步处理
- 图像数据较大，注意网络传输优化
- 对话历史会占用内存，需要定期清理

## 🧪 测试数据

模块包含丰富的测试数据，位于 `testdata/` 目录：

- `xhs-feed.jpeg`: 小红书信息流界面
- `popup_risk_warning.png`: 风险警告弹窗
- `llk_*.png`: 连连看游戏界面
- `deepseek_*.png`: DeepSeek 应用界面
- `chat_list.jpeg`: 聊天列表界面

这些测试数据覆盖了各种典型的 UI 场景，用于验证 AI 模块的功能正确性。

## 📈 扩展开发

### 添加新的模型支持

1. 在 `option` 包中定义新的模型类型
2. 实现对应的 `LLMContentParser`
3. 在 `GetModelConfig` 中添加模型验证逻辑
4. 更新系统提示词和输出格式

### 添加新的 CV 服务

1. 实现 `ICVService` 接口
2. 在 `NewCVService` 中添加服务创建逻辑
3. 定义服务特定的配置和选项
4. 添加相应的测试用例

### 优化解析逻辑

1. 扩展坐标格式支持
2. 改进参数映射规则
3. 增强错误处理机制
4. 优化性能和内存使用

通过这些扩展点，AI 模块可以持续演进，支持更多的模型和服务，提供更强大的智能化 UI 自动化能力。