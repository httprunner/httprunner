# HttpRunner MCP Server 完整说明文档

## 📖 概述

HttpRunner MCP Server 是基于 Model Context Protocol (MCP) 协议实现的 UI 自动化测试服务器，将 HttpRunner 的强大 UI 自动化能力通过标准化的 MCP 接口暴露给 AI 模型和其他客户端，支持移动端和 Web 端的 UI 自动化任务。

## 🏗️ 架构设计

### 整体架构

采用纯 ActionTool 架构，每个 UI 操作都作为独立的工具实现：

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   MCP Client    │    │   MCP Server    │    │  XTDriver Core  │
│   (AI Model)    │◄──►│  (mcp_server)   │◄──►│   (UI Engine)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │  Device Layer   │
                       │ Android/iOS/Web │
                       └─────────────────┘
```

### 核心组件

#### MCPServer4XTDriver
MCP 协议服务器主体：

```go
type MCPServer4XTDriver struct {
    mcpServer     *server.MCPServer                // MCP 协议服务器
    mcpTools      []mcp.Tool                       // 注册的工具列表
    actionToolMap map[option.ActionName]ActionTool // 动作到工具的映射
}
```

#### ActionTool 接口
所有 MCP 工具的统一契约：

```go
type ActionTool interface {
    Name() option.ActionName                                              // 工具名称
    Description() string                                                  // 工具描述
    Options() []mcp.ToolOption                                           // MCP 选项定义
    Implement() server.ToolHandlerFunc                                   // 工具实现逻辑
    ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) // 动作转换
}
```

### 模块化架构

MCP 工具按功能类别拆分为多个文件：

- **mcp_server.go**: 核心服务器实现和工具注册
- **mcp_tools_device.go**: 设备管理工具
- **mcp_tools_touch.go**: 触摸操作工具
- **mcp_tools_swipe.go**: 滑动和拖拽操作工具
- **mcp_tools_input.go**: 输入和 IME 工具
- **mcp_tools_button.go**: 按键操作工具
- **mcp_tools_app.go**: 应用管理工具
- **mcp_tools_screen.go**: 屏幕操作工具
- **mcp_tools_utility.go**: 实用工具（睡眠、弹窗等）
- **mcp_tools_web.go**: Web 操作工具
- **mcp_tools_ai.go**: AI 驱动操作工具

### 架构特点

- **完全解耦**: 每个工具独立实现，无依赖关系
- **统一接口**: 所有工具遵循相同的 ActionTool 接口
- **模块化组织**: 按功能分类的清晰文件结构
- **直接调用**: `MCP Request -> ActionTool.Implement() -> Driver Method`

## 📋 响应格式

### 扁平化响应结构

所有工具使用统一的扁平化响应格式，所有字段在同一层级：

```json
{
    "action": "list_packages",
    "success": true,
    "message": "Found 5 installed packages",
    "packages": ["com.example.app1", "com.example.app2"],
    "count": 2
}
```

### 标准字段

每个响应包含三个标准字段：
- **action**: 执行的操作名称
- **success**: 操作是否成功（布尔值）
- **message**: 人类可读的结果描述

### 工具特定字段

每个工具根据功能返回特定数据字段，与标准字段在同一层级。

### 响应创建

统一的响应创建函数：

```go
func NewMCPSuccessResponse(message string, actionTool ActionTool) *mcp.CallToolResult
```

该函数自动：
- 提取操作名称
- 设置成功状态
- 使用反射提取工具字段
- 创建扁平化响应

### 工具结构定义

工具结构体只包含返回数据字段：

```go
type ToolListPackages struct {
    Packages []string `json:"packages" desc:"List of installed app package names on the device"`
    Count    int      `json:"count" desc:"Number of installed packages"`
}
```

### 自动模式生成

使用反射自动生成返回模式：

```go
func GenerateReturnSchema(toolStruct interface{}) map[string]string
```

## 🎯 功能特性

### 支持的操作类别

#### 设备管理（mcp_tools_device.go）
- **list_available_devices**: 发现 Android/iOS 设备和模拟器
- **select_device**: 通过平台和序列号选择特定设备

#### 触摸操作（mcp_tools_touch.go）
- **tap_xy**: 在相对坐标点击 (0-1 范围)
- **tap_abs_xy**: 在绝对像素坐标点击
- **tap_ocr**: 通过 OCR 识别文本并点击
- **tap_cv**: 通过计算机视觉识别元素并点击
- **double_tap_xy**: 在坐标处双击

#### 手势操作（mcp_tools_swipe.go）
- **swipe**: 通用滑动，自动检测方向或坐标
- **swipe_direction**: 方向滑动 (上/下/左/右)
- **swipe_coordinate**: 基于坐标的精确滑动控制
- **drag**: 两点间的拖拽操作
- **swipe_to_tap_app**: 滑动查找并点击应用
- **swipe_to_tap_text**: 滑动查找并点击文本
- **swipe_to_tap_texts**: 滑动查找并点击多个文本中的一个

#### 输入操作（mcp_tools_input.go）
- **input**: 在焦点元素上输入文本
- **set_ime**: 设置输入法编辑器

#### 按键操作（mcp_tools_button.go）
- **press_button**: 按设备按键 (home、back、音量等)
- **home**: 按 home 键
- **back**: 按 back 键

#### 应用管理（mcp_tools_app.go）
- **list_packages**: 列出所有已安装应用
- **app_launch**: 通过包名启动应用
- **app_terminate**: 终止运行中的应用
- **app_install**: 从 URL/路径安装应用
- **app_uninstall**: 通过包名卸载应用
- **app_clear**: 清除应用数据和缓存

#### 屏幕操作（mcp_tools_screen.go）
- **screenshot**: 捕获屏幕为 Base64 编码图像
- **get_screen_size**: 获取设备屏幕尺寸
- **get_source**: 获取 UI 层次结构/源码

#### 实用工具操作（mcp_tools_utility.go）
- **sleep**: 等待指定秒数
- **sleep_ms**: 等待指定毫秒数
- **sleep_random**: 基于参数的随机等待
- **close_popups**: 关闭弹窗/对话框

#### Web 操作（mcp_tools_web.go）
- **web_login_none_ui**: 执行无 UI 交互的登录
- **secondary_click**: 在指定坐标右键点击
- **hover_by_selector**: 通过 CSS 选择器/XPath 悬停元素
- **tap_by_selector**: 通过 CSS 选择器/XPath 点击元素
- **secondary_click_by_selector**: 通过选择器右键点击元素
- **web_close_tab**: 通过索引关闭浏览器标签页

#### AI 操作（mcp_tools_ai.go）
- **start_to_goal**: 使用自然语言描述开始到目标的任务
- **ai_action**: 使用自然语言提示执行 AI 驱动的动作
- **finished**: 标记任务完成并返回结果消息

### 关键特性

#### 反作弊支持
为敏感操作内置反检测机制：
- 真实时间的触摸模拟
- 设备指纹掩码
- 行为模式随机化

#### 统一参数处理
所有工具通过 `parseActionOptions()` 使用一致的参数解析：
- 类型安全的 JSON 编组/解组
- 自动验证和错误处理
- 支持复杂嵌套参数

#### 设备抽象
无缝的多平台支持：
- Android 设备（通过 ADB）
- iOS 设备（通过 go-ios）
- Web 浏览器（通过 WebDriver）
- Harmony OS 设备

#### 错误处理
全面的错误管理：
- 结构化错误响应
- 带上下文的详细日志记录
- 优雅的故障恢复

## 📖 使用指南

### 创建和启动服务器

```go
// 创建和启动 MCP 服务器
server := NewMCPServer()
err := server.Start() // 阻塞并通过 stdio 提供 MCP 协议服务
```

### 客户端交互流程
1. **初始化连接**: 建立 MCP 协议连接
2. **工具发现**: 客户端查询可用工具列表
3. **工具调用**: 客户端调用特定工具执行操作
4. **响应处理**: 服务器返回结构化响应

### 工具实现模式

每个工具遵循一致的实现模式：

```go
type ToolExample struct {
    // Return data fields - these define the structure of data returned by this tool
    Field1 string `json:"field1" desc:"Description of field1"`
    Field2 int    `json:"field2" desc:"Description of field2"`
}

func (t *ToolExample) Name() option.ActionName {
    return option.ACTION_Example
}

func (t *ToolExample) Description() string {
    return "Description of what this tool does"
}

func (t *ToolExample) Options() []mcp.ToolOption {
    unifiedReq := &option.ActionOptions{}
    return unifiedReq.GetMCPOptions(option.ACTION_Example)
}

func (t *ToolExample) Implement() server.ToolHandlerFunc {
    return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        // Setup driver
        driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
        if err != nil {
            return nil, fmt.Errorf("setup driver failed: %w", err)
        }

        // Parse parameters
        unifiedReq, err := parseActionOptions(request.Params.Arguments)
        if err != nil {
            return nil, err
        }

        // Execute business logic
        // ... implementation ...

        // Create response
        message := "Operation completed successfully"
        returnData := ToolExample{
            Field1: "value1",
            Field2: 42,
        }

        return NewMCPSuccessResponse(message, &returnData), nil
    }
}

func (t *ToolExample) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
    // Convert action to MCP request
    arguments := map[string]any{
        "param1": action.Params,
    }
    return buildMCPCallToolRequest(t.Name(), arguments), nil
}
```

### 参数处理

#### 统一参数结构
所有工具使用 `option.ActionOptions` 结构进行参数处理：

```go
type ActionOptions struct {
    // Common fields
    Platform string `json:"platform,omitempty"`
    Serial   string `json:"serial,omitempty"`

    // Action-specific fields
    Text     string  `json:"text,omitempty"`
    X        float64 `json:"x,omitempty"`
    Y        float64 `json:"y,omitempty"`
    // ... more fields
}
```

#### 参数解析
使用 `parseActionOptions()` 函数进行类型安全的参数解析：

```go
unifiedReq, err := parseActionOptions(request.Params.Arguments)
if err != nil {
    return nil, err
}
```

### 错误处理

#### 错误响应
使用 `NewMCPErrorResponse()` 创建错误响应：

```go
if err != nil {
    return NewMCPErrorResponse(fmt.Sprintf("Operation failed: %s", err.Error())), nil
}
```

#### 错误响应格式
```json
{
    "success": false,
    "message": "Error description"
}
```

## 🔧 开发指南

### 添加新工具

1. **定义工具结构体**：
```go
type ToolNewFeature struct {
    // Return data fields
    Result string `json:"result" desc:"Description of result"`
}
```

2. **实现 ActionTool 接口**：
```go
func (t *ToolNewFeature) Name() option.ActionName {
    return option.ACTION_NewFeature
}

func (t *ToolNewFeature) Description() string {
    return "Description of the new feature"
}

func (t *ToolNewFeature) Options() []mcp.ToolOption {
    unifiedReq := &option.ActionOptions{}
    return unifiedReq.GetMCPOptions(option.ACTION_NewFeature)
}

func (t *ToolNewFeature) Implement() server.ToolHandlerFunc {
    // Implementation logic
}

func (t *ToolNewFeature) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
    // Conversion logic
}
```

3. **注册工具**：
在 `mcp_server.go` 的 `NewMCPServer()` 函数中添加：

```go
&ToolNewFeature{},
```

### 测试工具

#### 单元测试
```go
func TestToolNewFeature(t *testing.T) {
    tool := &ToolNewFeature{}

    // Test Name
    assert.Equal(t, option.ACTION_NewFeature, tool.Name())

    // Test Description
    assert.NotEmpty(t, tool.Description())

    // Test Options
    options := tool.Options()
    assert.NotEmpty(t, options)

    // Test schema generation
    schema := GenerateReturnSchema(tool)
    assert.Contains(t, schema, "result")
}
```

#### 集成测试
```go
func TestToolNewFeatureIntegration(t *testing.T) {
    // Create mock request
    request := mcp.CallToolRequest{
        Params: mcp.CallToolRequestParams{
            Arguments: map[string]any{
                "param1": "value1",
            },
        },
    }

    // Execute tool
    tool := &ToolNewFeature{}
    handler := tool.Implement()
    result, err := handler(context.Background(), request)

    // Verify result
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

### 最佳实践

#### 工具设计
- **单一职责**: 每个工具只负责一个特定功能
- **清晰命名**: 使用描述性的工具名称
- **完整文档**: 提供详细的描述和参数说明
- **错误处理**: 提供有意义的错误消息

#### 响应设计
- **一致性**: 所有工具使用相同的响应格式
- **信息丰富**: 返回足够的信息供客户端使用
- **类型安全**: 使用适当的数据类型
- **描述性**: 提供清晰的字段描述

#### 性能优化
- **延迟加载**: 只在需要时初始化资源
- **资源复用**: 复用驱动程序连接
- **错误快速失败**: 尽早检测和报告错误
- **日志记录**: 提供适当的日志级别

## 📊 工具统计

### 总计
- **总工具数**: 40+ 个
- **文件数**: 9 个工具文件
- **支持平台**: Android、iOS、Web、Harmony OS

### 按类别分布
- **设备管理**: 2 个工具
- **触摸操作**: 5 个工具
- **手势操作**: 7 个工具
- **输入操作**: 2 个工具
- **按键操作**: 3 个工具
- **应用管理**: 6 个工具
- **屏幕操作**: 3 个工具
- **实用工具**: 4 个工具
- **Web 操作**: 6 个工具
- **AI 操作**: 3 个工具

## 🚀 性能特性

### 优化成果
- **代码减少**: 相比原始实现减少约 70% 的样板代码
- **一致性**: 100% 的工具使用统一响应格式
- **自动化**: 完全自动化的模式生成
- **类型安全**: 保持完整的类型安全性
- **零手动定义**: 无需手动定义响应模式

### 架构优势
- **极简化**: 单函数调用创建响应
- **可维护性**: 清晰的代码结构和分离关注点
- **开发体验**: 直观的 API 和最小认知开销
- **自文档化**: 代码即文档的设计

## 📝 总结

HttpRunner MCP Server 提供了一个强大、灵活且易于使用的 UI 自动化平台。通过采用扁平化响应格式和自动化模式生成，实现了极简化的架构，同时保持了完整的功能性和类型安全性。

该架构的主要优势：
- **统一性**: 所有工具遵循相同的模式
- **简洁性**: 最小化的样板代码
- **可扩展性**: 易于添加新功能
- **可维护性**: 清晰的代码组织
- **性能**: 优化的响应创建和处理

无论是进行移动应用测试、Web 自动化还是 AI 驱动的 UI 操作，HttpRunner MCP Server 都提供了必要的工具和基础设施来支持各种自动化需求。
