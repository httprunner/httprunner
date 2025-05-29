# HttpRunner MCP Server 完整说明文档

## 📖 概述

HttpRunner MCP Server 是基于 Model Context Protocol (MCP) 协议实现的 UI 自动化测试服务器，它将 HttpRunner 的强大 UI 自动化能力通过标准化的 MCP 接口暴露给 AI 模型和其他客户端。

## 🎯 核心功能特性

### 1. 设备管理
- **设备发现**: 自动发现 Android/iOS 设备和模拟器
- **设备选择**: 支持通过序列号/UDID 选择特定设备
- **多平台支持**: Android、iOS、Harmony、Browser 全平台覆盖

### 2. 交互操作
- **点击操作**: 支持坐标点击、OCR 文本点击、CV 图像识别点击
- **滑动操作**: 方向滑动、坐标滑动、智能滑动查找
- **拖拽操作**: 精确的拖拽控制，支持反作弊
- **输入操作**: 文本输入、按键操作

### 3. 应用管理
- **应用控制**: 启动、终止、安装、卸载、清除数据
- **包名查询**: 获取设备上所有应用包名
- **前台应用**: 获取当前前台应用信息

### 4. 屏幕操作
- **截图功能**: 高质量屏幕截图，支持 Base64 编码
- **屏幕信息**: 获取屏幕尺寸、方向等信息
- **UI 层次**: 获取界面元素层次结构

### 5. 高级功能
- **AI 驱动**: 支持 AI 模型驱动的智能操作
- **反作弊机制**: 内置反作弊检测和规避
- **Web 自动化**: 支持浏览器自动化操作
- **时间控制**: 精确的等待和延时控制

## 🏗️ 架构设计

### 整体架构

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

#### 1. MCPServer4XTDriver
```go
type MCPServer4XTDriver struct {
    mcpServer     *server.MCPServer                // MCP 协议服务器
    mcpTools      []mcp.Tool                       // 注册的工具列表
    actionToolMap map[option.ActionName]ActionTool // 动作到工具的映射
}
```

#### 2. ActionTool 接口
```go
type ActionTool interface {
    Name() option.ActionName                                              // 工具名称
    Description() string                                                  // 工具描述
    Options() []mcp.ToolOption                                           // MCP 选项定义
    Implement() server.ToolHandlerFunc                                   // 工具实现逻辑
    ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) // 动作转换
}
```

## 🛠️ 实现思路

### 1. 纯 ActionTool 架构

采用纯 ActionTool 风格架构，每个 MCP 工具都是独立的结构体：

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
```

### 2. 统一参数处理

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

### 3. 设备管理策略

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

### 4. 错误处理机制

统一的错误处理和日志记录：

```go
if err != nil {
    log.Error().Err(err).Str("tool", toolName).Msg("tool execution failed")
    return mcp.NewToolResultError(fmt.Sprintf("操作失败: %s", err.Error())), nil
}
```

## 🔧 如何扩展接入新工具

### 步骤 1: 定义工具结构体

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

### 步骤 2: 定义 MCP 选项

```go
func (t *ToolLongPress) Options() []mcp.ToolOption {
    return []mcp.ToolOption{
        mcp.WithString("platform", mcp.Enum("android", "ios"), mcp.Description("设备平台")),
        mcp.WithString("serial", mcp.Description("设备序列号")),
        mcp.WithNumber("x", mcp.Description("X 坐标")),
        mcp.WithNumber("y", mcp.Description("Y 坐标")),
        mcp.WithNumber("duration", mcp.Description("长按持续时间(秒)")),
        mcp.WithBoolean("anti_risk", mcp.Description("是否启用反作弊")),
    }
}
```

### 步骤 3: 实现工具逻辑

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

### 步骤 4: 实现动作转换

```go
func (t *ToolLongPress) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
    if params, err := builtin.ConvertToFloat64Slice(action.Params); err == nil && len(params) >= 2 {
        arguments := map[string]any{
            "x": params[0],
            "y": params[1],
        }

        // 添加持续时间
        if len(params) > 2 {
            arguments["duration"] = params[2]
        }

        // 提取动作选项
        extractActionOptionsToArguments(action.GetOptions(), arguments)

        return buildMCPCallToolRequest(t.Name(), arguments), nil
    }
    return mcp.CallToolRequest{}, fmt.Errorf("invalid long press params: %v", action.Params)
}
```

### 步骤 5: 注册工具

在 `registerTools()` 方法中添加新工具：

```go
func (s *MCPServer4XTDriver) registerTools() {
    // ... 现有工具注册 ...

    // 注册新工具
    s.registerTool(&ToolLongPress{})

    // ... 其他工具 ...
}
```

### 步骤 6: 添加单元测试

```go
func TestToolLongPress(t *testing.T) {
    tool := &ToolLongPress{}

    // 测试工具基本信息
    assert.Equal(t, option.ACTION_LongPress, tool.Name())
    assert.Contains(t, tool.Description(), "长按")

    // 测试选项定义
    options := tool.Options()
    assert.NotEmpty(t, options)

    // 测试动作转换
    action := MobileAction{
        Method: option.ACTION_LongPress,
        Params: []float64{100, 200, 2.0}, // x, y, duration
        ActionOptions: option.ActionOptions{
            AntiRisk: true,
        },
    }

    request, err := tool.ConvertActionToCallToolRequest(action)
    assert.NoError(t, err)
    assert.Equal(t, string(option.ACTION_LongPress), request.Params.Name)
    assert.Equal(t, 100.0, request.Params.Arguments["x"])
    assert.Equal(t, 200.0, request.Params.Arguments["y"])
    assert.Equal(t, 2.0, request.Params.Arguments["duration"])
    assert.Equal(t, true, request.Params.Arguments["anti_risk"])
}
```

## 📋 工具开发最佳实践

### 1. 命名规范
- 工具结构体: `Tool{ActionName}`
- 常量定义: `ACTION_{ActionName}`
- 参数名称: 使用下划线分隔 (`from_x`, `to_y`)

### 2. 参数验证
```go
// 必需参数验证
if unifiedReq.Text == "" {
    return nil, fmt.Errorf("text parameter is required")
}

// 坐标参数验证
_, hasX := request.Params.Arguments["x"]
_, hasY := request.Params.Arguments["y"]
if !hasX || !hasY {
    return nil, fmt.Errorf("x and y coordinates are required")
}
```

### 3. 错误处理
```go
// 统一错误格式
if err != nil {
    return mcp.NewToolResultError(fmt.Sprintf("操作失败: %s", err.Error())), nil
}

// 成功结果
return mcp.NewToolResultText(fmt.Sprintf("操作成功: %s", details)), nil
```

### 4. 日志记录
```go
// 操作开始日志
log.Info().Str("action", "long_press").
    Float64("x", x).Float64("y", y).
    Msg("executing long press operation")

// 调试日志
log.Debug().Interface("arguments", arguments).
    Msg("parsed tool arguments")
```

### 5. 选项处理
```go
// 使用 extractActionOptionsToArguments 统一处理
extractActionOptionsToArguments(action.GetOptions(), arguments)

// 或手动添加特定选项
if unifiedReq.AntiRisk {
    opts = append(opts, option.WithAntiRisk(true))
}
```

## 🚀 高级特性

### 1. 反作弊支持
```go
// 在需要反作弊的操作中添加
if unifiedReq.AntiRisk {
    arguments := getCommonMCPArguments(driver)
    callMCPActionTool(driver, "evalpkgs", "set_touch_info", arguments)
}
```

### 2. 异步操作
```go
// 对于长时间运行的操作，使用 context 控制超时
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
```

### 3. 批量操作
```go
// 支持批量参数处理
for _, point := range unifiedReq.Points {
    err := driverExt.TapXY(point.X, point.Y, opts...)
    if err != nil {
        return mcp.NewToolResultError(fmt.Sprintf("批量操作失败: %s", err.Error())), nil
    }
}
```

## 📚 MCP Tools 快速参考

### 📱 设备管理工具

#### list_available_devices
**功能**: 发现所有可用的设备和模拟器
**参数**: 无
**返回**: JSON 格式的设备列表
```json
{
  "androidDevices": ["emulator-5554", "device-serial"],
  "iosDevices": ["iPhone-UDID", "simulator-UDID"]
}
```

#### select_device
**功能**: 选择要使用的设备
**参数**:
- `platform` (string): "android" | "ios" | "web" | "harmony"
- `serial` (string): 设备序列号或 UDID

---

### 👆 触摸操作工具

#### tap_xy
**功能**: 在相对坐标点击 (0-1 范围)
**参数**:
- `x` (number): X 坐标 (0.0-1.0)
- `y` (number): Y 坐标 (0.0-1.0)
- `duration` (number, 可选): 点击持续时间(秒)
- `anti_risk` (boolean, 可选): 启用反作弊

#### tap_abs_xy
**功能**: 在绝对像素坐标点击
**参数**:
- `x` (number): X 像素坐标
- `y` (number): Y 像素坐标
- `duration` (number, 可选): 点击持续时间(秒)
- `anti_risk` (boolean, 可选): 启用反作弊

#### tap_ocr
**功能**: 通过 OCR 识别文本并点击
**参数**:
- `text` (string): 要查找的文本
- `ignore_NotFoundError` (boolean, 可选): 忽略未找到错误
- `regex` (boolean, 可选): 使用正则表达式匹配

#### tap_cv
**功能**: 通过计算机视觉识别图像并点击
**参数**:
- `imagePath` (string): 模板图像路径
- `threshold` (number, 可选): 匹配阈值

#### double_tap_xy
**功能**: 在指定坐标双击
**参数**:
- `x` (number): X 坐标
- `y` (number): Y 坐标

---

### 🔄 手势操作工具

#### swipe
**功能**: 通用滑动 (自动检测方向或坐标)
**参数**: 支持方向滑动或坐标滑动两种模式

##### 方向滑动模式:
- `direction` (string): "up" | "down" | "left" | "right"
- `duration` (number, 可选): 滑动持续时间
- `press_duration` (number, 可选): 按压持续时间

##### 坐标滑动模式:
- `from_x` (number): 起始 X 坐标
- `from_y` (number): 起始 Y 坐标
- `to_x` (number): 结束 X 坐标
- `to_y` (number): 结束 Y 坐标

#### drag
**功能**: 拖拽操作
**参数**:
- `from_x` (number): 起始 X 坐标
- `from_y` (number): 起始 Y 坐标
- `to_x` (number): 结束 X 坐标
- `to_y` (number): 结束 Y 坐标
- `duration` (number, 可选): 拖拽持续时间(毫秒)

#### swipe_to_tap_app
**功能**: 滑动查找并点击应用
**参数**:
- `appName` (string): 应用名称
- `max_retry_times` (number, 可选): 最大重试次数
- `ignore_NotFoundError` (boolean, 可选): 忽略未找到错误

#### swipe_to_tap_text
**功能**: 滑动查找并点击文本
**参数**:
- `text` (string): 要查找的文本
- `max_retry_times` (number, 可选): 最大重试次数
- `regex` (boolean, 可选): 使用正则表达式

#### swipe_to_tap_texts
**功能**: 滑动查找并点击多个文本中的一个
**参数**:
- `texts` (array): 文本数组
- `max_retry_times` (number, 可选): 最大重试次数

---

### ⌨️ 输入操作工具

#### input
**功能**: 在当前焦点元素输入文本
**参数**:
- `text` (string): 要输入的文本

#### press_button
**功能**: 按设备按键
**参数**:
- `button` (string): 按键名称
  - Android: "BACK", "HOME", "VOLUME_UP", "VOLUME_DOWN", "ENTER"
  - iOS: "HOME", "VOLUME_UP", "VOLUME_DOWN"

#### home
**功能**: 按 Home 键
**参数**: 无

#### back
**功能**: 按返回键 (仅 Android)
**参数**: 无

---

### 📱 应用管理工具

#### list_packages
**功能**: 列出设备上所有应用包名
**参数**: 无

#### app_launch
**功能**: 启动应用
**参数**:
- `packageName` (string): 应用包名

#### app_terminate
**功能**: 终止应用
**参数**:
- `packageName` (string): 应用包名

#### app_install
**功能**: 安装应用
**参数**:
- `appUrl` (string): APK/IPA 文件路径或 URL

#### app_uninstall
**功能**: 卸载应用
**参数**:
- `packageName` (string): 应用包名

#### app_clear
**功能**: 清除应用数据
**参数**:
- `packageName` (string): 应用包名

---

### 📸 屏幕操作工具

#### screenshot
**功能**: 截取屏幕截图
**参数**: 无
**返回**: Base64 编码的图像数据

#### get_screen_size
**功能**: 获取屏幕尺寸
**参数**: 无
**返回**: 屏幕宽度和高度 (像素)

#### get_source
**功能**: 获取 UI 层次结构
**参数**:
- `packageName` (string, 可选): 指定应用包名

---

### ⏱️ 时间控制工具

#### sleep
**功能**: 等待指定秒数
**参数**:
- `seconds` (number): 等待秒数

#### sleep_ms
**功能**: 等待指定毫秒数
**参数**:
- `milliseconds` (number): 等待毫秒数

#### sleep_random
**功能**: 随机等待
**参数**:
- `params` (array): 随机参数数组

---

### 🛠️ 实用工具

#### set_ime
**功能**: 设置输入法
**参数**:
- `ime` (string): 输入法包名

#### close_popups
**功能**: 关闭弹窗
**参数**: 无

---

### 🌐 Web 操作工具

#### web_login_none_ui
**功能**: 无 UI 登录
**参数**:
- `packageName` (string): 应用包名
- `phoneNumber` (string, 可选): 手机号
- `captcha` (string, 可选): 验证码
- `password` (string, 可选): 密码

#### secondary_click
**功能**: 右键点击
**参数**:
- `x` (number): X 坐标
- `y` (number): Y 坐标

#### hover_by_selector
**功能**: 悬停在选择器元素上
**参数**:
- `selector` (string): CSS 选择器或 XPath

#### tap_by_selector
**功能**: 点击选择器元素
**参数**:
- `selector` (string): CSS 选择器或 XPath

#### secondary_click_by_selector
**功能**: 右键点击选择器元素
**参数**:
- `selector` (string): CSS 选择器或 XPath

#### web_close_tab
**功能**: 关闭浏览器标签页
**参数**:
- `tabIndex` (number): 标签页索引

---

### 🤖 AI 操作工具

#### ai_action
**功能**: AI 驱动的智能操作
**参数**:
- `prompt` (string): 自然语言指令

#### finished
**功能**: 标记任务完成
**参数**:
- `content` (string): 完成信息

---

### 📋 通用参数说明

#### 设备参数 (所有工具通用)
- `platform` (string): 设备平台
  - "android": Android 设备
  - "ios": iOS 设备
  - "web": Web 浏览器
  - "harmony": 鸿蒙设备
- `serial` (string): 设备标识符
  - Android: 设备序列号 (如 "emulator-5554")
  - iOS: 设备 UDID
  - Web: 浏览器会话 ID

#### 坐标参数
- **相对坐标**: 0.0-1.0 范围，相对于屏幕尺寸
- **绝对坐标**: 像素值，基于实际屏幕分辨率

#### 时间参数
- `duration`: 操作持续时间 (秒)
- `press_duration`: 按压持续时间 (秒)
- `milliseconds`: 毫秒数

#### 行为参数
- `anti_risk`: 启用反作弊检测
- `ignore_NotFoundError`: 忽略元素未找到错误
- `regex`: 使用正则表达式匹配
- `pre_mark_operation`: 启用操作前标记 (用于调试和可视化)
- `max_retry_times`: 最大重试次数
- `index`: 元素索引 (多个匹配时)

---

### 🔧 使用示例

#### 基本点击操作
```json
{
  "name": "tap_xy",
  "arguments": {
    "platform": "android",
    "serial": "emulator-5554",
    "x": 0.5,
    "y": 0.3
  }
}
```

#### 滑动操作
```json
{
  "name": "swipe",
  "arguments": {
    "platform": "android",
    "serial": "emulator-5554",
    "direction": "up",
    "duration": 0.5
  }
}
```

#### 应用启动
```json
{
  "name": "app_launch",
  "arguments": {
    "platform": "android",
    "serial": "emulator-5554",
    "packageName": "com.example.app"
  }
}
```

#### OCR 文本点击
```json
{
  "name": "tap_ocr",
  "arguments": {
    "platform": "android",
    "serial": "emulator-5554",
    "text": "登录",
    "ignore_NotFoundError": false
  }
}
```

---

### ⚠️ 注意事项

1. **设备连接**: 确保设备已连接并可访问
2. **权限要求**: 某些操作需要设备 root 或开发者权限
3. **坐标系统**: 注意相对坐标 (0-1) 和绝对坐标 (像素) 的区别
4. **平台差异**: 不同平台支持的功能可能有差异
5. **错误处理**: 建议启用适当的错误忽略选项
6. **性能考虑**: 避免过于频繁的操作，适当添加等待时间
