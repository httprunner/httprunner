# MCP 工具文档

## 概述

HttpRunner UIXT 基于 Model Context Protocol (MCP) 协议实现了标准化的工具接口，将所有 UI 操作封装为 MCP 工具，支持 AI 模型直接调用，实现真正的智能化 UI 自动化。

## MCP 架构

### 整体架构

```
┌─────────────────────────────────────────────────────────────────┐
│                        MCP 生态系统                             │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐  │
│  │   MCP Client    │    │   MCP Server    │    │  Tool Registry  │  │
│  │   (AI Model)    │◄──►│  (UIXT Server)  │◄──►│   (工具注册)     │  │
│  └─────────────────┘    └─────────────────┘    └─────────────────┘  │
├─────────────────────────────────────────────────────────────────┤
│                        工具层                                   │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │  Device Tools   │  │  Action Tools   │  │   AI Tools      │  │
│  │   (设备工具)     │  │   (操作工具)     │  │   (AI工具)      │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘  │
├─────────────────────────────────────────────────────────────────┤
│                      底层驱动                                   │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │  Android Driver │  │   iOS Driver    │  │  Browser Driver │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
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

## 工具分类

### 设备管理工具 (mcp_tools_device.go)

#### list_available_devices
发现可用的设备和模拟器。

```json
{
  "name": "uixt__list_available_devices",
  "description": "List all available devices including Android devices, iOS devices, and simulators",
  "inputSchema": {
    "type": "object",
    "properties": {},
    "required": []
  }
}
```

**响应示例**：
```json
{
  "action": "list_available_devices",
  "success": true,
  "message": "Found 3 available devices",
  "devices": [
    {
      "platform": "android",
      "serial": "emulator-5554",
      "name": "Android Emulator",
      "status": "online"
    }
  ],
  "count": 3
}
```

#### select_device
选择特定的设备进行操作。

```json
{
  "name": "uixt__select_device",
  "description": "Select a specific device by platform and serial number",
  "inputSchema": {
    "type": "object",
    "properties": {
      "platform": {
        "type": "string",
        "description": "Device platform (android, ios, browser, harmony)"
      },
      "serial": {
        "type": "string",
        "description": "Device serial number or identifier"
      }
    },
    "required": ["platform", "serial"]
  }
}
```

### 触摸操作工具 (mcp_tools_touch.go)

#### tap_xy
在相对坐标位置点击（0-1 范围）。

```json
{
  "name": "uixt__tap_xy",
  "description": "Tap at relative coordinates (0-1 range)",
  "inputSchema": {
    "type": "object",
    "properties": {
      "x": {
        "type": "number",
        "description": "X coordinate (0-1 range)"
      },
      "y": {
        "type": "number",
        "description": "Y coordinate (0-1 range)"
      }
    },
    "required": ["x", "y"]
  }
}
```

#### tap_abs_xy
在绝对像素坐标位置点击。

```json
{
  "name": "uixt__tap_abs_xy",
  "description": "Tap at absolute pixel coordinates",
  "inputSchema": {
    "type": "object",
    "properties": {
      "x": {
        "type": "number",
        "description": "Absolute X coordinate in pixels"
      },
      "y": {
        "type": "number",
        "description": "Absolute Y coordinate in pixels"
      }
    },
    "required": ["x", "y"]
  }
}
```

#### tap_ocr
通过 OCR 识别文本并点击。

```json
{
  "name": "uixt__tap_ocr",
  "description": "Find text using OCR and tap on it",
  "inputSchema": {
    "type": "object",
    "properties": {
      "text": {
        "type": "string",
        "description": "Text to find and tap"
      },
      "regex": {
        "type": "boolean",
        "description": "Whether to use regex matching"
      },
      "index": {
        "type": "integer",
        "description": "Index of text occurrence to tap (0-based)"
      }
    },
    "required": ["text"]
  }
}
```

#### tap_cv
通过计算机视觉识别 UI 元素并点击。

```json
{
  "name": "uixt__tap_cv",
  "description": "Find UI element using computer vision and tap on it",
  "inputSchema": {
    "type": "object",
    "properties": {
      "element_type": {
        "type": "string",
        "description": "Type of UI element to find"
      },
      "description": {
        "type": "string",
        "description": "Description of the element"
      }
    },
    "required": ["element_type"]
  }
}
```

### 滑动操作工具 (mcp_tools_swipe.go)

#### swipe
通用滑动操作，自动检测方向或坐标。

```json
{
  "name": "uixt__swipe",
  "description": "Perform swipe gesture with automatic direction or coordinate detection",
  "inputSchema": {
    "type": "object",
    "properties": {
      "direction": {
        "type": "string",
        "description": "Swipe direction (up, down, left, right)"
      },
      "from_x": {
        "type": "number",
        "description": "Start X coordinate (0-1 range)"
      },
      "from_y": {
        "type": "number",
        "description": "Start Y coordinate (0-1 range)"
      },
      "to_x": {
        "type": "number",
        "description": "End X coordinate (0-1 range)"
      },
      "to_y": {
        "type": "number",
        "description": "End Y coordinate (0-1 range)"
      }
    }
  }
}
```

#### swipe_to_tap_app
滑动查找并点击应用。

```json
{
  "name": "uixt__swipe_to_tap_app",
  "description": "Swipe to find and tap on an app",
  "inputSchema": {
    "type": "object",
    "properties": {
      "app_name": {
        "type": "string",
        "description": "Name of the app to find and tap"
      },
      "max_swipes": {
        "type": "integer",
        "description": "Maximum number of swipes to perform"
      }
    },
    "required": ["app_name"]
  }
}
```

### 输入操作工具 (mcp_tools_input.go)

#### input
在焦点元素上输入文本。

```json
{
  "name": "uixt__input",
  "description": "Input text into the focused element",
  "inputSchema": {
    "type": "object",
    "properties": {
      "text": {
        "type": "string",
        "description": "Text to input"
      }
    },
    "required": ["text"]
  }
}
```

#### set_ime
设置输入法编辑器。

```json
{
  "name": "uixt__set_ime",
  "description": "Set the Input Method Editor (IME)",
  "inputSchema": {
    "type": "object",
    "properties": {
      "ime": {
        "type": "string",
        "description": "IME package name or identifier"
      }
    },
    "required": ["ime"]
  }
}
```

### 按键操作工具 (mcp_tools_button.go)

#### press_button
按设备按键。

```json
{
  "name": "uixt__press_button",
  "description": "Press a device button",
  "inputSchema": {
    "type": "object",
    "properties": {
      "button": {
        "type": "string",
        "description": "Button name (home, back, volume_up, volume_down, etc.)"
      }
    },
    "required": ["button"]
  }
}
```

### 应用管理工具 (mcp_tools_app.go)

#### list_packages
列出所有已安装的应用包。

```json
{
  "name": "uixt__list_packages",
  "description": "List all installed app packages on the device",
  "inputSchema": {
    "type": "object",
    "properties": {},
    "required": []
  }
}
```

#### app_launch
启动应用。

```json
{
  "name": "uixt__app_launch",
  "description": "Launch an app by package name",
  "inputSchema": {
    "type": "object",
    "properties": {
      "package_name": {
        "type": "string",
        "description": "Package name of the app to launch"
      }
    },
    "required": ["package_name"]
  }
}
```

#### app_terminate
终止应用。

```json
{
  "name": "uixt__app_terminate",
  "description": "Terminate a running app",
  "inputSchema": {
    "type": "object",
    "properties": {
      "package_name": {
        "type": "string",
        "description": "Package name of the app to terminate"
      }
    },
    "required": ["package_name"]
  }
}
```

### 屏幕操作工具 (mcp_tools_screen.go)

#### screenshot
捕获屏幕截图。

```json
{
  "name": "uixt__screenshot",
  "description": "Take a screenshot of the device screen",
  "inputSchema": {
    "type": "object",
    "properties": {},
    "required": []
  }
}
```

**响应示例**：
```json
{
  "action": "screenshot",
  "success": true,
  "message": "Screenshot captured successfully",
  "screenshot": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA...",
  "width": 1080,
  "height": 1920
}
```

#### get_screen_size
获取屏幕尺寸。

```json
{
  "name": "uixt__get_screen_size",
  "description": "Get the screen size of the device",
  "inputSchema": {
    "type": "object",
    "properties": {},
    "required": []
  }
}
```

### 实用工具 (mcp_tools_utility.go)

#### sleep
等待指定秒数。

```json
{
  "name": "uixt__sleep",
  "description": "Sleep for specified number of seconds",
  "inputSchema": {
    "type": "object",
    "properties": {
      "seconds": {
        "type": "number",
        "description": "Number of seconds to sleep"
      }
    },
    "required": ["seconds"]
  }
}
```

#### close_popups
关闭弹窗或对话框。

```json
{
  "name": "uixt__close_popups",
  "description": "Close popups or dialogs on the screen",
  "inputSchema": {
    "type": "object",
    "properties": {},
    "required": []
  }
}
```

### Web 操作工具 (mcp_tools_web.go)

#### secondary_click
在指定坐标右键点击。

```json
{
  "name": "uixt__secondary_click",
  "description": "Perform secondary click (right-click) at coordinates",
  "inputSchema": {
    "type": "object",
    "properties": {
      "x": {
        "type": "number",
        "description": "X coordinate for secondary click"
      },
      "y": {
        "type": "number",
        "description": "Y coordinate for secondary click"
      }
    },
    "required": ["x", "y"]
  }
}
```

#### hover_by_selector
通过选择器悬停元素。

```json
{
  "name": "uixt__hover_by_selector",
  "description": "Hover over element by CSS selector or XPath",
  "inputSchema": {
    "type": "object",
    "properties": {
      "selector": {
        "type": "string",
        "description": "CSS selector or XPath of the element"
      }
    },
    "required": ["selector"]
  }
}
```

### AI 操作工具 (mcp_tools_ai.go)

#### start_to_goal
使用自然语言描述执行从开始到目标的任务。

```json
{
  "name": "uixt__start_to_goal",
  "description": "Execute a task from start to goal using natural language description",
  "inputSchema": {
    "type": "object",
    "properties": {
      "goal": {
        "type": "string",
        "description": "Natural language description of the goal"
      }
    },
    "required": ["goal"]
  }
}
```

#### ai_action
使用自然语言提示执行 AI 驱动的动作。

```json
{
  "name": "uixt__ai_action",
  "description": "Execute AI-driven action using natural language prompt",
  "inputSchema": {
    "type": "object",
    "properties": {
      "prompt": {
        "type": "string",
        "description": "Natural language prompt for the action"
      }
    },
    "required": ["prompt"]
  }
}
```

## 工具实现

### ActionTool 实现示例

```go
// 点击工具实现
type ToolTapXY struct {
    X float64 `json:"x" desc:"X coordinate (0-1 range)"`
    Y float64 `json:"y" desc:"Y coordinate (0-1 range)"`
}

func (t *ToolTapXY) Name() option.ActionName {
    return option.ActionTapXY
}

func (t *ToolTapXY) Description() string {
    return "Tap at relative coordinates (0-1 range)"
}

func (t *ToolTapXY) Options() []mcp.ToolOption {
    return []mcp.ToolOption{
        {
            Name:        "x",
            Type:        "number",
            Description: "X coordinate (0-1 range)",
            Required:    true,
        },
        {
            Name:        "y",
            Type:        "number",
            Description: "Y coordinate (0-1 range)",
            Required:    true,
        },
    }
}

func (t *ToolTapXY) Implement() server.ToolHandlerFunc {
    return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        // 解析参数
        x, ok := req.Params.Arguments["x"].(float64)
        if !ok {
            return mcp.NewToolResultError("invalid x coordinate"), nil
        }

        y, ok := req.Params.Arguments["y"].(float64)
        if !ok {
            return mcp.NewToolResultError("invalid y coordinate"), nil
        }

        // 执行操作
        err := GetXTDriverFromContext(ctx).TapXY(x, y)
        if err != nil {
            return mcp.NewToolResultError(fmt.Sprintf("tap failed: %v", err)), nil
        }

        // 设置响应数据
        t.X = x
        t.Y = y

        return NewMCPSuccessResponse(
            fmt.Sprintf("Tapped at coordinates (%.2f, %.2f)", x, y),
            t,
        ), nil
    }
}
```

### 响应格式

所有工具使用统一的扁平化响应格式：

```go
func NewMCPSuccessResponse(message string, actionTool ActionTool) *mcp.CallToolResult {
    response := map[string]interface{}{
        "action":  string(actionTool.Name()),
        "success": true,
        "message": message,
    }

    // 使用反射提取工具字段
    toolValue := reflect.ValueOf(actionTool)
    if toolValue.Kind() == reflect.Ptr {
        toolValue = toolValue.Elem()
    }

    toolType := toolValue.Type()
    for i := 0; i < toolValue.NumField(); i++ {
        field := toolType.Field(i)
        jsonTag := field.Tag.Get("json")
        if jsonTag != "" && jsonTag != "-" {
            fieldName := strings.Split(jsonTag, ",")[0]
            response[fieldName] = toolValue.Field(i).Interface()
        }
    }

    return &mcp.CallToolResult{
        Content: []mcp.Content{
            {
                Type: mcp.ContentTypeText,
                Text: toJSONString(response),
            },
        },
    }
}
```

## 工具注册

### 服务器初始化

```go
func NewMCPServer() *MCPServer4XTDriver {
    server := &MCPServer4XTDriver{
        mcpTools:      make([]mcp.Tool, 0),
        actionToolMap: make(map[option.ActionName]ActionTool),
    }

    // 注册所有工具
    server.registerDeviceTools()
    server.registerTouchTools()
    server.registerSwipeTools()
    server.registerInputTools()
    server.registerButtonTools()
    server.registerAppTools()
    server.registerScreenTools()
    server.registerUtilityTools()
    server.registerWebTools()
    server.registerAITools()

    return server
}
```

### 工具注册方法

```go
func (s *MCPServer4XTDriver) registerTool(tool ActionTool) {
    // 创建 MCP 工具定义
    mcpTool := mcp.Tool{
        Name:        fmt.Sprintf("uixt__%s", tool.Name()),
        Description: tool.Description(),
        InputSchema: map[string]interface{}{
            "type":       "object",
            "properties": generateProperties(tool.Options()),
            "required":   getRequiredFields(tool.Options()),
        },
    }

    // 注册到服务器
    s.mcpTools = append(s.mcpTools, mcpTool)
    s.actionToolMap[tool.Name()] = tool
}
```

## 工具调用

### 客户端调用

```go
// 通过 MCP 客户端调用工具
func callTool(client client.MCPClient, toolName string, args map[string]interface{}) (*mcp.CallToolResult, error) {
    req := mcp.CallToolRequest{
        Params: mcp.CallToolParams{
            Name:      fmt.Sprintf("uixt__%s", toolName),
            Arguments: args,
        },
    }

    return client.CallTool(context.Background(), req)
}

// 使用示例
result, err := callTool(client, "tap_xy", map[string]interface{}{
    "x": 0.5,
    "y": 0.5,
})
```

### 服务器处理

```go
func (s *MCPServer4XTDriver) CallTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    // 提取工具名称
    toolName := strings.TrimPrefix(req.Params.Name, "uixt__")
    actionName := option.ActionName(toolName)

    // 查找工具
    tool, exists := s.actionToolMap[actionName]
    if !exists {
        return mcp.NewToolResultError(fmt.Sprintf("tool %s not found", toolName)), nil
    }

    // 执行工具
    handler := tool.Implement()
    return handler(ctx, req)
}
```

## 扩展开发

### 创建自定义工具

```go
// 1. 定义工具结构
type ToolCustomAction struct {
    Parameter1 string `json:"parameter1" desc:"Description of parameter1"`
    Parameter2 int    `json:"parameter2" desc:"Description of parameter2"`
}

// 2. 实现 ActionTool 接口
func (t *ToolCustomAction) Name() option.ActionName {
    return option.ActionName("custom_action")
}

func (t *ToolCustomAction) Description() string {
    return "Perform a custom action"
}

func (t *ToolCustomAction) Options() []mcp.ToolOption {
    return []mcp.ToolOption{
        {
            Name:        "parameter1",
            Type:        "string",
            Description: "Description of parameter1",
            Required:    true,
        },
        {
            Name:        "parameter2",
            Type:        "integer",
            Description: "Description of parameter2",
            Required:    false,
        },
    }
}

func (t *ToolCustomAction) Implement() server.ToolHandlerFunc {
    return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        // 解析参数
        param1, ok := req.Params.Arguments["parameter1"].(string)
        if !ok {
            return mcp.NewToolResultError("invalid parameter1"), nil
        }

        param2, _ := req.Params.Arguments["parameter2"].(float64)

        // 执行自定义逻辑
        err := performCustomAction(param1, int(param2))
        if err != nil {
            return mcp.NewToolResultError(fmt.Sprintf("custom action failed: %v", err)), nil
        }

        // 设置响应数据
        t.Parameter1 = param1
        t.Parameter2 = int(param2)

        return NewMCPSuccessResponse("Custom action completed", t), nil
    }
}

// 3. 注册工具
func (s *MCPServer4XTDriver) registerCustomTools() {
    s.registerTool(&ToolCustomAction{})
}
```

### 工具分组

```go
// 按功能分组注册工具
func (s *MCPServer4XTDriver) registerToolGroup(groupName string, tools []ActionTool) {
    for _, tool := range tools {
        // 添加分组前缀
        mcpTool := mcp.Tool{
            Name:        fmt.Sprintf("uixt__%s__%s", groupName, tool.Name()),
            Description: fmt.Sprintf("[%s] %s", groupName, tool.Description()),
            InputSchema: generateInputSchema(tool),
        }

        s.mcpTools = append(s.mcpTools, mcpTool)
        s.actionToolMap[tool.Name()] = tool
    }
}
```

## 最佳实践

### 1. 工具设计原则

```go
// 单一职责：每个工具只做一件事
type ToolSinglePurpose struct {
    // 明确的参数定义
    TargetText string `json:"target_text" desc:"Text to search for"`
}

// 参数验证：在工具实现中验证参数
func (t *ToolSinglePurpose) Implement() server.ToolHandlerFunc {
    return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        // 参数验证
        if err := t.validateParameters(req.Params.Arguments); err != nil {
            return mcp.NewToolResultError(err.Error()), nil
        }

        // 执行逻辑
        return t.execute(ctx, req)
    }
}
```

### 2. 错误处理

```go
// 统一的错误处理
func handleToolError(err error, toolName string) *mcp.CallToolResult {
    if err == nil {
        return nil
    }

    // 记录错误日志
    log.Error().Err(err).Str("tool", toolName).Msg("tool execution failed")

    // 返回用户友好的错误信息
    return mcp.NewToolResultError(fmt.Sprintf("Tool %s failed: %v", toolName, err))
}
```

### 3. 性能优化

```go
// 工具执行缓存
type ToolCache struct {
    cache map[string]*mcp.CallToolResult
    mutex sync.RWMutex
}

func (c *ToolCache) GetOrExecute(key string, executor func() (*mcp.CallToolResult, error)) (*mcp.CallToolResult, error) {
    c.mutex.RLock()
    if result, exists := c.cache[key]; exists {
        c.mutex.RUnlock()
        return result, nil
    }
    c.mutex.RUnlock()

    // 执行工具
    result, err := executor()
    if err != nil {
        return nil, err
    }

    // 缓存结果
    c.mutex.Lock()
    c.cache[key] = result
    c.mutex.Unlock()

    return result, nil
}
```

### 4. 工具组合

```go
// 复合工具：组合多个基础工具
type ToolComposite struct {
    Steps []ToolStep `json:"steps" desc:"Sequence of tool steps"`
}

type ToolStep struct {
    Tool      string                 `json:"tool"`
    Arguments map[string]interface{} `json:"arguments"`
}

func (t *ToolComposite) Implement() server.ToolHandlerFunc {
    return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        results := make([]interface{}, 0, len(t.Steps))

        for i, step := range t.Steps {
            // 执行每个步骤
            result, err := executeToolStep(ctx, step)
            if err != nil {
                return mcp.NewToolResultError(fmt.Sprintf("step %d failed: %v", i+1, err)), nil
            }
            results = append(results, result)
        }

        return NewMCPSuccessResponse("Composite tool completed", t), nil
    }
}
```

## 故障排除

### 常见问题

#### 工具注册失败

```go
// 检查工具注册
func validateToolRegistration(server *MCPServer4XTDriver) error {
    tools := server.ListTools()
    if len(tools) == 0 {
        return fmt.Errorf("no tools registered")
    }

    // 检查必需工具
    requiredTools := []string{"tap_xy", "screenshot", "app_launch"}
    for _, required := range requiredTools {
        found := false
        for _, tool := range tools {
            if strings.HasSuffix(tool.Name, required) {
                found = true
                break
            }
        }
        if !found {
            return fmt.Errorf("required tool %s not found", required)
        }
    }

    return nil
}
```

#### 工具调用失败

```go
// 调试工具调用
func debugToolCall(req mcp.CallToolRequest) {
    log.Debug().
        Str("tool", req.Params.Name).
        Interface("arguments", req.Params.Arguments).
        Msg("tool call debug")

    // 验证参数类型
    for key, value := range req.Params.Arguments {
        log.Debug().
            Str("param", key).
            Str("type", fmt.Sprintf("%T", value)).
            Interface("value", value).
            Msg("parameter debug")
    }
}
```

#### 性能问题

```go
// 监控工具性能
func monitorToolPerformance(toolName string, executor func() (*mcp.CallToolResult, error)) (*mcp.CallToolResult, error) {
    start := time.Now()

    result, err := executor()

    elapsed := time.Since(start)
    log.Info().
        Str("tool", toolName).
        Dur("elapsed", elapsed).
        Bool("success", err == nil).
        Msg("tool performance")

    if elapsed > 5*time.Second {
        log.Warn().
            Str("tool", toolName).
            Dur("elapsed", elapsed).
            Msg("slow tool execution")
    }

    return result, err
}
```

## 参考资料

- [Model Context Protocol 规范](https://modelcontextprotocol.io/docs/)
- [MCP Go 实现](https://github.com/mark3labs/mcp-go)
- [HttpRunner UIXT MCP 服务器文档](mcp_server.md)