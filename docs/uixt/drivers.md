# 驱动接口文档

## 概述

HttpRunner UIXT 提供统一的驱动接口 `IDriver`，支持多种平台的 UI 自动化操作。每个平台都有专门的驱动实现，但对外提供相同的接口，确保跨平台的一致性。

## IDriver 核心接口

### 接口定义

```go
type IDriver interface {
    // 设备管理
    GetDevice() IDevice
    Setup() error
    TearDown() error

    // 会话管理
    InitSession(capabilities option.Capabilities) error
    GetSession() *DriverSession
    DeleteSession() error

    // 设备信息和状态
    Status() (types.DeviceStatus, error)
    DeviceInfo() (types.DeviceInfo, error)
    BatteryInfo() (types.BatteryInfo, error)
    ForegroundInfo() (app types.AppInfo, err error)
    WindowSize() (types.Size, error)
    ScreenShot(opts ...option.ActionOption) (*bytes.Buffer, error)
    ScreenRecord(opts ...option.ActionOption) (videoPath string, err error)
    Source(srcOpt ...option.SourceOption) (string, error)
    Orientation() (orientation types.Orientation, err error)
    Rotation() (rotation types.Rotation, err error)

    // 配置
    SetRotation(rotation types.Rotation) error
    SetIme(ime string) error

    // 基础操作
    Home() error
    Unlock() error
    Back() error
    PressButton(button types.DeviceButton) error

    // 悬停操作
    HoverBySelector(selector string, opts ...option.ActionOption) error

    // 点击操作
    TapXY(x, y float64, opts ...option.ActionOption) error
    TapAbsXY(x, y float64, opts ...option.ActionOption) error
    TapBySelector(text string, opts ...option.ActionOption) error
    DoubleTap(x, y float64, opts ...option.ActionOption) error
    TouchAndHold(x, y float64, opts ...option.ActionOption) error

    // 右键操作
    SecondaryClick(x, y float64) error
    SecondaryClickBySelector(selector string, options ...option.ActionOption) error

    // 滑动操作
    Drag(fromX, fromY, toX, toY float64, opts ...option.ActionOption) error
    Swipe(fromX, fromY, toX, toY float64, opts ...option.ActionOption) error

    // 输入操作
    Input(text string, opts ...option.ActionOption) error
    Backspace(count int, opts ...option.ActionOption) error

    // 应用管理
    AppLaunch(packageName string) error
    AppTerminate(packageName string) (bool, error)
    AppClear(packageName string) error

    // 文件管理
    PushImage(localPath string) error
    PullImages(localDir string) error
    ClearImages() error
    PushFile(localPath string, remoteDir string) error
    PullFiles(localDir string, remoteDirs ...string) error
    ClearFiles(paths ...string) error

    // 日志管理
    StartCaptureLog(identifier ...string) error
    StopCaptureLog() (result interface{}, err error)
}
```

## Android 驱动

### ADBDriver

基于 ADB (Android Debug Bridge) 的基础驱动，提供设备管理和基础操作。

```go
// 创建 ADB 驱动
device, err := uixt.NewAndroidDevice(option.WithSerialNumber("device_serial"))
driver, err := uixt.NewADBDriver(device)
```

#### 特色功能

- **应用管理**: 安装、卸载、启动、终止应用
- **文件传输**: 推送和拉取文件
- **Shell 命令**: 执行 Android shell 命令
- **日志捕获**: 实时捕获系统日志
- **屏幕录制**: 录制屏幕视频
- **系统设置**: 网络、权限、系统配置

#### 使用示例

```go
// 应用管理
err = driver.InstallApp("/path/to/app.apk")
err = driver.UninstallApp("com.example.app")
err = driver.AppLaunch("com.example.app")
terminated, err := driver.AppTerminate("com.example.app")
err = driver.AppClear("com.example.app")

// 文件操作
err = driver.PushFile("/local/path/file.txt", "/sdcard/")
err = driver.PullFiles("/local/dir", "/sdcard/Download")

// Shell 命令
output, err := driver.Shell("pm list packages")
output, err := driver.Shell("dumpsys battery")

// 日志捕获
err = driver.StartCaptureLog("main", "system")
logs, err := driver.StopCaptureLog()

// 权限管理
err = driver.GrantPermission("com.example.app", "android.permission.CAMERA")
err = driver.RevokePermission("com.example.app", "android.permission.CAMERA")

// 系统设置
err = driver.EnableWiFi()
err = driver.ConnectWiFi("SSID", "password")
err = driver.EnableMobileData()
```

### UIA2Driver

基于 UiAutomator2 的高级驱动，提供完整的 UI 自动化功能。

```go
// 创建 UIA2 驱动
device, err := uixt.NewAndroidDevice(option.WithSerialNumber("device_serial"))
driver, err := uixt.NewUIA2Driver(device)
```

#### 特色功能

- **UI 元素定位**: 支持多种选择器
- **手势操作**: 点击、滑动、拖拽等
- **输入操作**: 文本输入、按键操作
- **屏幕操作**: 截图、录制、旋转
- **页面源码**: 获取 UI 层次结构
- **等待机制**: 元素等待和条件等待

#### 选择器类型

```go
// 文本选择器
err = driver.TapBySelector("text=登录")
err = driver.TapBySelector("textContains=登")
err = driver.TapBySelector("textMatches=登.*")

// 资源ID选择器
err = driver.TapBySelector("resource-id=com.example:id/login_button")
err = driver.TapBySelector("resourceId=login_button")

// 类名选择器
err = driver.TapBySelector("className=android.widget.Button")

// 描述选择器
err = driver.TapBySelector("description=登录按钮")
err = driver.TapBySelector("contentDescription=登录按钮")

// 组合选择器
err = driver.TapBySelector("className=android.widget.Button,text=登录")

// XPath 选择器
err = driver.TapBySelector("xpath=//android.widget.Button[@text='登录']")
```

#### 使用示例

```go
// UI 操作
err = driver.TapXY(0.5, 0.5)                    // 相对坐标点击
err = driver.TapAbsXY(500, 800)                 // 绝对坐标点击
err = driver.TapBySelector("text=登录")          // 通过文本点击
err = driver.DoubleTap(0.5, 0.5)               // 双击
err = driver.TouchAndHold(0.5, 0.5)            // 长按

// 滑动操作
err = driver.Swipe(0.5, 0.8, 0.5, 0.2)         // 滑动
err = driver.Drag(0.2, 0.5, 0.8, 0.5)          // 拖拽

// 输入操作
err = driver.Input("Hello World")
err = driver.Backspace(5)
err = driver.PressButton(types.DeviceButtonBack)

// 屏幕操作
screenshot, err := driver.ScreenShot()
videoPath, err := driver.ScreenRecord()
source, err := driver.Source()

// 等待操作
err = driver.WaitForElement("text=登录", 10*time.Second)
err = driver.WaitForElementGone("text=加载中", 30*time.Second)
```

## iOS 驱动

### WDADriver

基于 WebDriverAgent 的 iOS 驱动，提供完整的 iOS UI 自动化功能。

```go
// 创建 WDA 驱动
device, err := uixt.NewIOSDevice(option.WithUDID("device_udid"))
driver, err := uixt.NewWDADriver(device)
```

#### 特色功能

- **原生 iOS 支持**: 支持 iOS 原生应用和系统应用
- **多点触控**: 支持复杂手势和多指操作
- **应用管理**: 启动、终止、安装、卸载应用
- **性能监控**: 获取应用性能数据和系统信息
- **弹窗处理**: 自动处理系统弹窗和权限请求
- **屏幕录制**: 支持高质量屏幕录制

#### 选择器类型

```go
// 文本选择器
err = driver.TapBySelector("label=登录")
err = driver.TapBySelector("name=登录按钮")

// 类型选择器
err = driver.TapBySelector("type=XCUIElementTypeButton")
err = driver.TapBySelector("className=XCUIElementTypeButton")

// 可访问性标识符
err = driver.TapBySelector("id=login_button")
err = driver.TapBySelector("accessibilityId=login_button")

// 值选择器
err = driver.TapBySelector("value=用户名")

// 组合选择器
err = driver.TapBySelector("type=XCUIElementTypeButton,label=登录")

// XPath 选择器
err = driver.TapBySelector("xpath=//XCUIElementTypeButton[@label='登录']")

// 谓词选择器
err = driver.TapBySelector("predicate=label CONTAINS '登录'")
err = driver.TapBySelector("predicate=type == 'XCUIElementTypeButton' AND visible == 1")
```

#### 使用示例

```go
// 应用管理
err = driver.AppLaunch("com.apple.mobilesafari")
err = driver.AppLaunch("com.example.app")
terminated, err := driver.AppTerminate("com.example.app")
err = driver.AppActivate("com.example.app") // 激活后台应用

// 手势操作
err = driver.TapXY(0.5, 0.5)                // 点击
err = driver.DoubleTap(100, 200)            // 双击
err = driver.TouchAndHold(150, 300)         // 长按
err = driver.Swipe(0.5, 0.8, 0.5, 0.2)     // 滑动
err = driver.Drag(0.2, 0.5, 0.8, 0.5)      // 拖拽

// 输入操作
err = driver.Input("Hello World")
err = driver.Backspace(5)
err = driver.ClearText()

// 设备操作
err = driver.Home()                         // 回到主屏
err = driver.Back()                         // 返回（如果支持）
err = driver.SetRotation(types.RotationLandscape)

// 屏幕操作
screenshot, err := driver.ScreenShot()
err = driver.StartScreenRecord()
videoPath, err := driver.StopScreenRecord()
source, err := driver.Source()

// 等待操作
err = driver.WaitForElement("label=登录", 10*time.Second)
err = driver.WaitForElementGone("label=加载中", 30*time.Second)
```

#### iOS 特有功能

```go
// Siri 操作
err = driver.ActivateSiri("打开设置")
err = driver.ActivateSiri("发送消息给张三")

// 3D Touch / Force Touch
err = driver.ForceTouch(100, 200, 0.8)      // 压力值 0.0-1.0
err = driver.ForceTouchBySelector("label=应用图标", 0.8)

// 设备控制
err = driver.Lock()                         // 锁定设备
err = driver.Unlock()                       // 解锁设备
err = driver.Shake()                        // 摇晃设备

// 音量控制
err = driver.VolumeUp()                     // 音量增加
err = driver.VolumeDown()                   // 音量减少
err = driver.SetVolume(0.5)                 // 设置音量 (0.0-1.0)

// 弹窗处理
err = driver.AcceptAlert()                  // 接受弹窗
err = driver.DismissAlert()                 // 关闭弹窗
alertText, err := driver.GetAlertText()     // 获取弹窗文本

// 键盘操作
err = driver.HideKeyboard()                 // 隐藏键盘
isVisible, err := driver.IsKeyboardShown()  // 检查键盘是否显示

// 应用状态
state, err := driver.GetAppState("com.example.app")
// 0: not installed, 1: not running, 2: running in background, 4: running in foreground

// 设备信息
battery, err := driver.BatteryInfo()
orientation, err := driver.Orientation()
size, err := driver.WindowSize()
```

## HarmonyOS 驱动

### HDCDriver

基于 HDC (HarmonyOS Device Connector) 的鸿蒙驱动，提供完整的 HarmonyOS UI 自动化功能。

```go
// 创建 HDC 驱动
device, err := uixt.NewHarmonyDevice(option.WithConnectKey("device_key"))
driver, err := uixt.NewHDCDriver(device)
```

#### 特色功能

- **原生鸿蒙支持**: 支持 HarmonyOS 应用和系统应用
- **分布式操作**: 支持多设备协同和跨设备操作
- **原子化服务**: 支持轻量级应用和服务
- **ArkUI 支持**: 支持 ArkUI 框架的组件识别
- **多模态交互**: 支持语音、手势等多种交互方式

#### 选择器类型

```go
// 文本选择器
err = driver.TapBySelector("text=登录")
err = driver.TapBySelector("textContains=登")

// 组件类型选择器
err = driver.TapBySelector("type=Button")
err = driver.TapBySelector("className=ohos.agp.components.Button")

// ID 选择器
err = driver.TapBySelector("id=login_button")
err = driver.TapBySelector("resourceId=login_button")

// 描述选择器
err = driver.TapBySelector("description=登录按钮")
err = driver.TapBySelector("contentDescription=登录按钮")

// 组合选择器
err = driver.TapBySelector("type=Button,text=登录")

// XPath 选择器
err = driver.TapBySelector("xpath=//Button[@text='登录']")
```

#### 使用示例

```go
// 基础操作
err = driver.TapXY(0.5, 0.5)                // 点击
err = driver.DoubleTap(0.5, 0.5)            // 双击
err = driver.TouchAndHold(0.5, 0.5)         // 长按
err = driver.Swipe(0.2, 0.8, 0.8, 0.2)     // 滑动
err = driver.Drag(0.2, 0.5, 0.8, 0.5)      // 拖拽

// 输入操作
err = driver.Input("测试文本")
err = driver.Backspace(5)
err = driver.PressButton(types.DeviceButtonBack)

// 应用管理
err = driver.AppLaunch("com.huawei.hmos.example")
err = driver.AppLaunch("com.example.harmony.app")
terminated, err := driver.AppTerminate("com.example.app")
err = driver.AppClear("com.example.app")

// 屏幕操作
screenshot, err := driver.ScreenShot()
videoPath, err := driver.ScreenRecord()
source, err := driver.Source()

// 等待操作
err = driver.WaitForElement("text=登录", 10*time.Second)
err = driver.WaitForElementGone("text=加载中", 30*time.Second)
```

#### HarmonyOS 特有功能

```go
// 分布式操作
err = driver.ConnectDistributedDevice("target_device_id")
err = driver.DisconnectDistributedDevice("target_device_id")

// 跨设备应用迁移
err = driver.MigrateApp("com.example.app", "target_device_id")

// 原子化服务
err = driver.LaunchAtomicService("service_id", map[string]interface{}{
    "param1": "value1",
    "param2": "value2",
})
err = driver.StopAtomicService("service_id")

// 多模态交互
err = driver.VoiceCommand("打开设置")
err = driver.GestureCommand("swipe_up")

// 系统设置
err = driver.EnableDistributedCapability()
err = driver.DisableDistributedCapability()

// 性能监控
performance, err := driver.GetPerformanceData()
memory, err := driver.GetMemoryInfo()
cpu, err := driver.GetCPUInfo()

// 设备信息
info, err := driver.DeviceInfo()
battery, err := driver.BatteryInfo()
```

## Web 驱动

### BrowserDriver

基于 WebDriver 协议的浏览器驱动，支持多种浏览器的 Web 自动化测试。

```go
// 创建浏览器驱动
device, err := uixt.NewBrowserDevice(option.WithBrowserID("chrome"))
driver, err := uixt.NewBrowserDriver(device)
```

#### 特色功能

- **多浏览器支持**: Chrome、Firefox、Safari、Edge
- **JavaScript 执行**: 执行自定义脚本和异步脚本
- **多标签页管理**: 创建、切换、关闭标签页
- **Cookie 管理**: 获取、设置、删除 Cookie
- **文件上传下载**: 支持文件操作
- **网络监控**: 监控网络请求和响应
- **移动端模拟**: 模拟移动设备和触摸操作

#### 选择器类型

```go
// CSS 选择器
err = driver.TapBySelector("#login-button")
err = driver.TapBySelector(".btn-primary")
err = driver.TapBySelector("button[type='submit']")

// XPath 选择器
err = driver.TapBySelector("xpath=//button[@id='login']")
err = driver.TapBySelector("xpath=//div[contains(@class, 'login')]//button")

// 文本选择器
err = driver.TapBySelector("text=登录")
err = driver.TapBySelector("linkText=点击这里")
err = driver.TapBySelector("partialLinkText=点击")

// 标签名选择器
err = driver.TapBySelector("tagName=button")
err = driver.TapBySelector("tagName=input")

// 属性选择器
err = driver.TapBySelector("name=username")
err = driver.TapBySelector("className=btn")
```

#### 使用示例

```go
// 页面导航
err = driver.NavigateTo("https://example.com")
err = driver.Refresh()
err = driver.GoBack()
err = driver.GoForward()

// 元素操作
err = driver.TapBySelector("#login-button")
err = driver.DoubleTap(100, 200)
err = driver.TouchAndHold(150, 300)
err = driver.Input("username")
err = driver.Backspace(5)

// 滑动和拖拽
err = driver.Swipe(0.5, 0.8, 0.5, 0.2)
err = driver.Drag(0.2, 0.5, 0.8, 0.5)

// 屏幕操作
screenshot, err := driver.ScreenShot()
err = driver.StartScreenRecord()
videoPath, err := driver.StopScreenRecord()

// JavaScript 执行
result, err := driver.ExecuteScript("return document.title;")
err = driver.ExecuteAsyncScript("callback(arguments[0]);", "test")

// 标签页管理
err = driver.NewTab()
err = driver.CloseTab(1)
err = driver.SwitchToTab(0)

// 等待操作
err = driver.WaitForElement("#element", 10*time.Second)
err = driver.WaitForElementGone("#loading", 30*time.Second)
err = driver.WaitForPageLoad(30*time.Second)
```

#### Web 特有功能

```go
// Cookie 操作
cookies, err := driver.GetCookies()
err = driver.SetCookie("name", "value", "domain.com")
err = driver.DeleteCookie("name")
err = driver.DeleteAllCookies()

// 窗口管理
err = driver.SetWindowSize(1920, 1080)
size, err := driver.GetWindowSize()
err = driver.Maximize()
err = driver.Minimize()
err = driver.Fullscreen()

// 页面信息
title, err := driver.GetTitle()
url, err := driver.GetCurrentURL()
source, err := driver.GetPageSource()

// 框架操作
err = driver.SwitchToFrame("frame_name")
err = driver.SwitchToFrameByIndex(0)
err = driver.SwitchToDefaultContent()

// 弹窗处理
err = driver.AcceptAlert()
err = driver.DismissAlert()
alertText, err := driver.GetAlertText()
err = driver.SendAlertText("input text")

// 文件操作
err = driver.UploadFile("#file-input", "/path/to/file.txt")
downloadPath, err := driver.DownloadFile("https://example.com/file.pdf")

// 网络监控
err = driver.StartNetworkMonitoring()
requests, err := driver.GetNetworkRequests()
err = driver.StopNetworkMonitoring()

// 移动端模拟
err = driver.SetMobileEmulation("iPhone 12")
err = driver.SetUserAgent("Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)")

// 性能监控
metrics, err := driver.GetPerformanceMetrics()
logs, err := driver.GetBrowserLogs()

// 截图和录制
fullPageScreenshot, err := driver.FullPageScreenShot()
elementScreenshot, err := driver.ElementScreenShot("#element")

// 元素信息
isVisible, err := driver.IsElementVisible("#element")
isEnabled, err := driver.IsElementEnabled("#button")
text, err := driver.GetElementText("#element")
value, err := driver.GetElementValue("#input")
attribute, err := driver.GetElementAttribute("#element", "class")

// 表单操作
err = driver.SelectOption("#select", "option_value")
err = driver.CheckCheckbox("#checkbox")
err = driver.UncheckCheckbox("#checkbox")
err = driver.SelectRadioButton("#radio")

// 滚动操作
err = driver.ScrollToElement("#element")
err = driver.ScrollToTop()
err = driver.ScrollToBottom()
err = driver.ScrollBy(0, 500)
```

## 扩展驱动 (XTDriver)

### 概述

`XTDriver` 是对基础驱动的扩展，集成了 AI 能力和 MCP 工具系统。

```go
// 创建扩展驱动
baseDriver, err := uixt.NewUIA2Driver(device)
xtDriver, err := uixt.NewXTDriver(baseDriver,
    option.WithCVService(option.CVServiceTypeVEDEM),
    option.WithLLMService(option.OPENAI_GPT_4O),
)
```

### 核心组件

```go
type XTDriver struct {
    IDriver                              // 基础驱动能力
    CVService  ai.ICVService             // 计算机视觉服务
    LLMService ai.ILLMService            // 大语言模型服务
    client     *MCPClient4XTDriver       // MCP 客户端
}
```

### AI 增强功能

#### 智能操作

```go
// 使用自然语言执行操作
result, err := xtDriver.LLMService.Plan(ctx, &ai.PlanningOptions{
    UserInstruction: "点击登录按钮并输入用户名",
    Message:         message,
    Size:           screenSize,
})

// 执行规划的操作
for _, toolCall := range result.ToolCalls {
    // 自动执行工具调用
}
```

#### 智能识别

```go
// OCR 文本识别
cvResult, err := xtDriver.CVService.ReadFromBuffer(screenshot)
ocrTexts := cvResult.OCRResult.ToOCRTexts()

// 查找特定文本
targetText, err := ocrTexts.FindText("登录")
center := targetText.Center()

// 点击识别的文本
err = xtDriver.TapAbsXY(center.X, center.Y)
```

#### 智能断言

```go
// 使用自然语言进行断言
assertResult, err := xtDriver.LLMService.Assert(ctx, &ai.AssertOptions{
    Assertion:  "页面应该显示用户已登录",
    Screenshot: screenshot,
    Size:       screenSize,
})

if assertResult.Pass {
    fmt.Println("断言通过")
} else {
    fmt.Printf("断言失败: %s\n", assertResult.Thought)
}
```

### MCP 工具集成

```go
// 执行 MCP 工具
result, err := xtDriver.ExecuteAction(ctx, option.MobileAction{
    Method: option.ActionTapXY,
    Params: map[string]interface{}{
        "x": 0.5,
        "y": 0.5,
    },
})
```

## 驱动选择指南

### 平台对应关系

| 平台 | 推荐驱动 | 备选驱动 | 说明 |
|------|----------|----------|------|
| Android | UIA2Driver | ADBDriver | UIA2 提供完整 UI 功能，ADB 提供基础操作 |
| iOS | WDADriver | - | 唯一选择，基于 WebDriverAgent |
| HarmonyOS | HDCDriver | - | 原生鸿蒙支持 |
| Web | BrowserDriver | - | 支持所有主流浏览器 |

### 选择建议

#### 功能需求

- **基础操作**: ADBDriver (Android)
- **完整 UI 自动化**: UIA2Driver (Android), WDADriver (iOS)
- **AI 增强**: XTDriver (所有平台)
- **Web 自动化**: BrowserDriver

#### 性能考虑

- **速度优先**: ADBDriver < UIA2Driver < WDADriver
- **稳定性**: WDADriver > UIA2Driver > ADBDriver
- **功能完整性**: XTDriver > 平台驱动 > 基础驱动

## 驱动配置

### 通用配置

```go
// 超时配置
driver.SetTimeout(30 * time.Second)

// 重试配置
driver.SetRetryCount(3)
driver.SetRetryInterval(1 * time.Second)

// 日志配置
driver.SetLogLevel(log.DebugLevel)
driver.EnableActionLog(true)
```

### 平台特定配置

#### Android 配置

```go
// UiAutomator2 配置
driver.SetUiAutomator2Config(uia2.Config{
    WaitForIdleTimeout:    10 * time.Second,
    WaitForSelectorTimeout: 20 * time.Second,
    ActionAcknowledgmentTimeout: 3 * time.Second,
})

// ADB 配置
driver.SetADBConfig(adb.Config{
    CommandTimeout: 30 * time.Second,
    ShellTimeout:   60 * time.Second,
})
```

#### iOS 配置

```go
// WebDriverAgent 配置
driver.SetWDAConfig(wda.Config{
    ConnectionTimeout: 60 * time.Second,
    CommandTimeout:    30 * time.Second,
    SnapshotTimeout:   15 * time.Second,
})
```

#### Web 配置

```go
// WebDriver 配置
driver.SetWebDriverConfig(webdriver.Config{
    PageLoadTimeout:    30 * time.Second,
    ScriptTimeout:      10 * time.Second,
    ImplicitWaitTimeout: 5 * time.Second,
})
```

## 最佳实践

### 1. 驱动生命周期管理

```go
func useDriver() error {
    // 创建驱动
    driver, err := createDriver()
    if err != nil {
        return err
    }

    // 初始化
    err = driver.Setup()
    if err != nil {
        return err
    }
    defer driver.TearDown() // 确保清理

    // 使用驱动
    return performOperations(driver)
}
```

### 2. 错误处理

```go
// 带重试的操作
func tapWithRetry(driver IDriver, x, y float64) error {
    maxRetries := 3
    for i := 0; i < maxRetries; i++ {
        err := driver.TapXY(x, y)
        if err == nil {
            return nil
        }

        // 检查是否是临时错误
        if isTemporaryError(err) {
            time.Sleep(time.Duration(i+1) * time.Second)
            continue
        }

        return err
    }
    return fmt.Errorf("operation failed after %d retries", maxRetries)
}
```

### 3. 性能优化

```go
// 批量操作
func performBatchOperations(driver IDriver, operations []Operation) error {
    // 开始批量模式
    driver.BeginBatch()
    defer driver.EndBatch()

    for _, op := range operations {
        err := op.Execute(driver)
        if err != nil {
            return err
        }
    }

    return nil
}
```

### 4. 跨平台兼容

```go
// 平台适配
func performPlatformSpecificOperation(driver IDriver) error {
    switch d := driver.(type) {
    case *UIA2Driver:
        // Android 特定操作
        return d.AndroidSpecificMethod()
    case *WDADriver:
        // iOS 特定操作
        return d.IOSSpecificMethod()
    case *BrowserDriver:
        // Web 特定操作
        return d.WebSpecificMethod()
    default:
        // 通用操作
        return driver.TapXY(0.5, 0.5)
    }
}
```

## 故障排除

### 常见问题

#### 驱动初始化失败

```go
// 检查设备连接
status, err := driver.Status()
if err != nil {
    log.Error("Device not connected: %v", err)
    return err
}

// 检查驱动服务
if !driver.IsServiceRunning() {
    err = driver.StartService()
    if err != nil {
        log.Error("Failed to start driver service: %v", err)
        return err
    }
}
```

#### 操作超时

```go
// 增加超时时间
driver.SetTimeout(60 * time.Second)

// 等待元素出现
err = driver.WaitForElement("selector", 30*time.Second)
if err != nil {
    log.Error("Element not found: %v", err)
    return err
}
```

#### 内存泄漏

```go
// 定期清理资源
func periodicCleanup(driver IDriver) {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()

    for range ticker.C {
        driver.ClearCache()
        runtime.GC()
    }
}
```

## 参考资料

- [UiAutomator2 文档](https://github.com/appium/appium-uiautomator2-driver)
- [WebDriverAgent 文档](https://github.com/appium/WebDriverAgent)
- [WebDriver 规范](https://w3c.github.io/webdriver/)
- [Android ADB 文档](https://developer.android.com/studio/command-line/adb)