# HttpRunner MCP Server 完整说明文档

## 📖 概述

HttpRunner MCP Server 是基于 Model Context Protocol (MCP) 协议实现的 UI 自动化测试服务器，它将 HttpRunner 的强大 UI 自动化能力通过标准化的 MCP 接口暴露给 AI 模型和其他客户端，使其能够执行移动端和 Web 端的 UI 自动化任务。

## 🏗️ 架构设计

### 整体架构

MCP 服务器采用纯 ActionTool 架构，其中每个 UI 操作都作为独立的工具实现，符合 ActionTool 接口规范：

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
管理 MCP 协议通信和工具注册的主要服务器结构体：

```go
type MCPServer4XTDriver struct {
    mcpServer     *server.MCPServer                // MCP 协议服务器
    mcpTools      []mcp.Tool                       // 注册的工具列表
    actionToolMap map[option.ActionName]ActionTool // 动作到工具的映射
}
```

#### ActionTool 接口
定义所有 MCP 工具的契约：

```go
type ActionTool interface {
    Name() option.ActionName                                              // 工具名称
    Description() string                                                  // 工具描述
    Options() []mcp.ToolOption                                           // MCP 选项定义
    Implement() server.ToolHandlerFunc                                   // 工具实现逻辑
    ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) // 动作转换
    ReturnSchema() map[string]string                                     // 返回值结构描述
}
```

### 模块化架构

为了更好的代码组织和维护，MCP 工具按功能类别拆分为多个文件：

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

#### 纯 ActionTool 架构实现
- **每个 MCP 工具都是实现 ActionTool 接口的独立结构体**
- **操作逻辑直接嵌入在每个工具的 Implement() 方法中**
- **工具间无中间动作方法或耦合关系**
- **完全解耦，摆脱了原有大型 switch-case DoAction 方法**

#### 架构流程
```
MCP Request -> ActionTool.Implement() -> Direct Driver Method Call
```

#### 架构优势
- **真正的 ActionTool 接口一致性**: 所有工具保持一致
- **完全解耦**: 无方法间依赖关系
- **模块化组织**: 按功能分类的文件结构
- **简化错误处理**: 每个工具独立的错误处理和日志记录
- **易于扩展**: 新功能易于扩展

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
- **ai_action**: 使用自然语言提示执行 AI 驱动的动作
- **finished**: 标记任务完成并返回结果消息

### 关键特性

#### 反作弊支持
为敏感操作内置反检测机制：
- 真实时间的触摸模拟
- 设备指纹掩码
- 行为模式随机化

#### 统一参数处理
所有工具通过 parseActionOptions() 使用一致的参数解析：
- 类型安全的 JSON 编组/解组
- 自动验证和错误处理
- 支持复杂嵌套参数

#### 设备抽象
无缝的多平台支持：
- 通过 ADB 支持 Android 设备
- 通过 go-ios 支持 iOS 设备
- 通过 WebDriver 支持 Web 浏览器
- 支持 Harmony OS 设备

#### 错误处理
全面的错误管理：
- 结构化错误响应
- 带上下文的详细日志记录
- 优雅的故障恢复

## 📖 使用指南

### 创建和启动服务器

#### NewMCPServer 函数
该函数创建一个新的 XTDriver MCP 服务器并注册所有工具：

- **MCP 协议服务器**: 具有 uixt 功能
- **版本信息**: 来自 HttpRunner
- **工具功能**: 为性能考虑禁用 (设置为 false)
- **预注册工具**: 所有可用的 UI 自动化工具

#### 使用示例
```go
// 创建和启动 MCP 服务器
server := NewMCPServer()
err := server.Start() // 阻塞并通过 stdio 提供 MCP 协议服务
```

#### 客户端交互流程
1. **初始化连接**: 建立 MCP 协议连接
2. **列出可用工具**: 获取所有注册的工具列表
3. **调用工具**: 使用参数调用特定工具
4. **接收结果**: 获取结构化的操作结果

## 🛠️ 实现原理

### 统一参数处理

使用 `parseActionOptions` 函数统一处理 MCP 请求参数：

```go
func parseActionOptions(arguments map[string]any) (*option.ActionOptions, error) {
    b, err := json.Marshal(arguments)
    if err != nil {
        return nil, fmt.Errorf("marshal arguments failed: %w", err)
    }

    var actionOptions option.ActionOptions
    if err := json.Unmarshal(b, &actionOptions); err != nil {
        return nil, fmt.Errorf("unmarshal to ActionOptions failed: %w", err)
    }

    return &actionOptions, nil
}
```

### 设备管理策略

通过 `setupXTDriver` 函数实现设备的统一管理：

```go
func setupXTDriver(ctx context.Context, arguments map[string]any) (*XTDriver, error) {
    // 1. 解析设备参数
    platform := arguments["platform"].(string)
    serial := arguments["serial"].(string)

    // 2. 获取或创建驱动器
    driverExt, err := GetOrCreateXTDriver(
        option.WithPlatform(platform),
        option.WithSerial(serial),
    )

    return driverExt, err
}
```

### 工具实现模式

每个 MCP 工具都遵循统一的实现模式：

```go
type ToolTapXY struct{}

func (t *ToolTapXY) Name() option.ActionName {
    return option.ACTION_TapXY
}

func (t *ToolTapXY) Implement() server.ToolHandlerFunc {
    return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        // 1. 设置驱动器
        driverExt, err := setupXTDriver(ctx, request.Params.Arguments)

        // 2. 解析参数
        unifiedReq, err := parseActionOptions(request.Params.Arguments)

        // 3. 执行操作
        err = driverExt.TapXY(unifiedReq.X, unifiedReq.Y, opts...)

        // 4. 返回结果
        return mcp.NewToolResultText("操作成功"), nil
    }
}

func (t *ToolTapXY) ReturnSchema() map[string]string {
    return map[string]string{
        "message": "string: Success message confirming tap operation at specified coordinates",
    }
}
```

### 错误处理机制

统一的错误处理和日志记录：

```go
if err != nil {
    log.Error().Err(err).Str("tool", toolName).Msg("tool execution failed")
    return mcp.NewToolResultError(fmt.Sprintf("操作失败: %s", err.Error())), nil
}
```

### 工具注册机制

在 `mcp_server.go` 的 `registerTools()` 方法中统一注册所有工具：

```go
func (s *MCPServer4XTDriver) registerTools() {
    // Device Tools
    s.registerTool(&ToolListAvailableDevices{})
    s.registerTool(&ToolSelectDevice{})

    // Touch Tools
    s.registerTool(&ToolTapXY{})
    s.registerTool(&ToolTapAbsXY{})
    s.registerTool(&ToolTapByOCR{})
    s.registerTool(&ToolTapByCV{})
    s.registerTool(&ToolDoubleTapXY{})

    // Swipe Tools
    s.registerTool(&ToolSwipe{})
    s.registerTool(&ToolSwipeDirection{})
    s.registerTool(&ToolSwipeCoordinate{})
    s.registerTool(&ToolSwipeToTapApp{})
    s.registerTool(&ToolSwipeToTapText{})
    s.registerTool(&ToolSwipeToTapTexts{})
    s.registerTool(&ToolDrag{})

    // Input Tools
    s.registerTool(&ToolInput{})
    s.registerTool(&ToolSetIme{})

    // Button Tools
    s.registerTool(&ToolPressButton{})
    s.registerTool(&ToolHome{})
    s.registerTool(&ToolBack{})

    // App Tools
    s.registerTool(&ToolListPackages{})
    s.registerTool(&ToolLaunchApp{})
    s.registerTool(&ToolTerminateApp{})
    s.registerTool(&ToolAppInstall{})
    s.registerTool(&ToolAppUninstall{})
    s.registerTool(&ToolAppClear{})

    // Screen Tools
    s.registerTool(&ToolScreenShot{})
    s.registerTool(&ToolGetScreenSize{})
    s.registerTool(&ToolGetSource{})

    // Utility Tools
    s.registerTool(&ToolSleep{})
    s.registerTool(&ToolSleepMS{})
    s.registerTool(&ToolSleepRandom{})
    s.registerTool(&ToolClosePopups{})

    // Web Tools
    s.registerTool(&ToolWebLoginNoneUI{})
    s.registerTool(&ToolSecondaryClick{})
    s.registerTool(&ToolHoverBySelector{})
    s.registerTool(&ToolTapBySelector{})
    s.registerTool(&ToolSecondaryClickBySelector{})
    s.registerTool(&ToolWebCloseTab{})

    // AI Tools
    s.registerTool(&ToolAIAction{})
    s.registerTool(&ToolFinished{})
}
```

## 🔧 扩展开发

### 添加新工具的步骤

1. **选择合适的文件**: 根据功能类别选择对应的 `mcp_tools_*.go` 文件
2. **定义工具结构体**: 实现 ActionTool 接口
3. **实现所有必需方法**: Name、Description、Options、Implement、ConvertActionToCallToolRequest、ReturnSchema
4. **在 registerTools() 方法中注册工具**
5. **添加全面的单元测试**
6. **更新文档**

### 开发示例：长按操作工具

假设要在 `mcp_tools_touch.go` 中添加长按操作：

#### 步骤 1: 定义工具结构体

```go
// 新工具：长按操作
type ToolLongPress struct{}

func (t *ToolLongPress) Name() option.ActionName {
    return option.ACTION_LongPress // 需要在 option 包中定义
}

func (t *ToolLongPress) Description() string {
    return "在指定坐标执行长按操作"
}
```

#### 步骤 2: 定义 MCP 选项

```go
func (t *ToolLongPress) Options() []mcp.ToolOption {
    unifiedReq := &option.ActionOptions{}
    return unifiedReq.GetMCPOptions(option.ACTION_LongPress)
}
```

#### 步骤 3: 实现工具逻辑

```go
func (t *ToolLongPress) Implement() server.ToolHandlerFunc {
    return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        // 1. 设置驱动器
        driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
        if err != nil {
            return nil, fmt.Errorf("setup driver failed: %w", err)
        }

        // 2. 解析参数
        unifiedReq, err := parseActionOptions(request.Params.Arguments)
        if err != nil {
            return nil, err
        }

        // 3. 参数验证
        if unifiedReq.X == 0 || unifiedReq.Y == 0 {
            return nil, fmt.Errorf("x and y coordinates are required")
        }

        // 4. 构建选项
        opts := []option.ActionOption{}
        if unifiedReq.Duration > 0 {
            opts = append(opts, option.WithDuration(unifiedReq.Duration))
        }
        if unifiedReq.AntiRisk {
            opts = append(opts, option.WithAntiRisk(true))
        }

        // 5. 执行操作
        log.Info().Float64("x", unifiedReq.X).Float64("y", unifiedReq.Y).
            Float64("duration", unifiedReq.Duration).Msg("executing long press")

        err = driverExt.LongPress(unifiedReq.X, unifiedReq.Y, opts...)
        if err != nil {
            return mcp.NewToolResultError(fmt.Sprintf("长按操作失败: %s", err.Error())), nil
        }

        // 6. 返回结果
        return mcp.NewToolResultText(fmt.Sprintf("成功在坐标 (%.2f, %.2f) 执行长按操作",
            unifiedReq.X, unifiedReq.Y)), nil
    }
}
```

#### 步骤 4: 实现动作转换和返回值结构

```go
func (t *ToolLongPress) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
    if params, err := builtin.ConvertToFloat64Slice(action.Params); err == nil && len(params) >= 2 {
        arguments := map[string]any{
            "x": params[0],
            "y": params[1],
        }
        if len(params) > 2 {
            arguments["duration"] = params[2]
        }
        extractActionOptionsToArguments(action.GetOptions(), arguments)
        return buildMCPCallToolRequest(t.Name(), arguments), nil
    }
    return mcp.CallToolRequest{}, fmt.Errorf("invalid long press params: %v", action.Params)
}

func (t *ToolLongPress) ReturnSchema() map[string]string {
    return map[string]string{
        "message":  "string: Success message confirming long press operation",
        "x":        "float64: X coordinate where long press was performed",
        "y":        "float64: Y coordinate where long press was performed",
        "duration": "float64: Duration of the long press in seconds",
    }
}
```

#### 步骤 5: 注册工具

在 `mcp_server.go` 的 `registerTools()` 方法中添加：

```go
// Touch Tools
s.registerTool(&ToolTapXY{})
s.registerTool(&ToolTapAbsXY{})
s.registerTool(&ToolTapByOCR{})
s.registerTool(&ToolTapByCV{})
s.registerTool(&ToolDoubleTapXY{})
s.registerTool(&ToolLongPress{}) // 新增长按工具
```

### 开发最佳实践

#### 文件组织规范
- **按功能分类**: 将相关工具放在同一个文件中
- **命名一致性**: 文件名使用 `mcp_tools_{category}.go` 格式
- **工具命名**: 结构体使用 `Tool{ActionName}` 格式

#### 参数验证
```go
// 必需参数验证
if unifiedReq.Text == "" {
    return nil, fmt.Errorf("text parameter is required")
}

// 坐标参数验证
if unifiedReq.X == 0 || unifiedReq.Y == 0 {
    return nil, fmt.Errorf("x and y coordinates are required")
}
```

#### 错误处理
```go
// 统一错误格式
if err != nil {
    return mcp.NewToolResultError(fmt.Sprintf("操作失败: %s", err.Error())), nil
}

// 成功结果
return mcp.NewToolResultText(fmt.Sprintf("操作成功: %s", details)), nil
```

#### 日志记录
```go
// 操作开始日志
log.Info().Str("action", "long_press").
    Float64("x", x).Float64("y", y).
    Msg("executing long press operation")

// 调试日志
log.Debug().Interface("arguments", arguments).
    Msg("parsed tool arguments")
```

#### 返回值类型规范
```go
// 标准返回值类型前缀
"message": "string: 描述信息"
"x": "float64: X坐标值"
"count": "int: 数量"
"success": "bool: 成功状态"
"items": "[]string: 字符串数组"
"data": "object: 复杂对象"
```

## 🚀 性能与安全

### 性能考虑

- **驱动器实例缓存**: 为提高效率，驱动器实例被缓存和重用
- **参数解析优化**: 参数解析经过优化以最小化 JSON 开销
- **超时控制**: 超时控制防止操作挂起
- **资源清理**: 资源清理确保内存效率
- **模块化加载**: 按需加载工具模块，减少内存占用

### 安全注意事项

- **设备操作权限**: 所有设备操作都需要明确权限
- **输入验证**: 输入验证防止注入攻击
- **敏感操作保护**: 敏感操作支持反检测措施
- **审计日志**: 审计日志跟踪所有工具执行

### 高级特性

#### 反作弊支持
```go
// 在需要反作弊的操作中添加
if unifiedReq.AntiRisk {
    arguments := getCommonMCPArguments(driver)
    callMCPActionTool(driver, "evalpkgs", "set_touch_info", arguments)
}
```

#### 异步操作
```go
// 对于长时间运行的操作，使用 context 控制超时
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
```

#### 批量操作
```go
// 支持批量参数处理
for _, point := range unifiedReq.Points {
    err := driverExt.TapXY(point.X, point.Y, opts...)
    if err != nil {
        return mcp.NewToolResultError(fmt.Sprintf("批量操作失败: %s", err.Error())), nil
    }
}
```

---

## 📚 总结

HttpRunner MCP Server 通过模块化的架构设计，将 UI 自动化功能按类别拆分为多个文件，每个文件专注于特定的功能领域。这种设计不仅提高了代码的可维护性和可扩展性，还使得开发者能够更容易地理解和贡献代码。

### 核心优势

1. **模块化架构**: 按功能分类的文件组织，便于维护和扩展
2. **统一接口**: 所有工具都实现相同的 ActionTool 接口
3. **类型安全**: 强类型的参数处理和返回值定义
4. **完整文档**: 每个工具都有详细的参数和返回值说明
5. **易于测试**: 独立的工具实现便于单元测试

该实现为 UI 自动化测试提供了一个完整、可扩展且高性能的 MCP 服务器解决方案。
