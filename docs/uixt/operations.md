# 操作指南文档

## 概述

HttpRunner UIXT 提供了丰富的 UI 操作接口，支持触摸、滑动、输入、应用管理等各种操作。本文档详细介绍每种操作的使用方法和最佳实践。

## 基础操作

### 点击操作

#### 相对坐标点击

使用 0-1 范围的相对坐标进行点击，适用于不同屏幕尺寸的设备。

```go
// 点击屏幕中心
err := driver.TapXY(0.5, 0.5)

// 点击右上角
err := driver.TapXY(0.9, 0.1)

// 点击左下角
err := driver.TapXY(0.1, 0.9)
```

#### 绝对坐标点击

使用像素坐标进行精确点击。

```go
// 点击绝对坐标 (500, 800)
err := driver.TapAbsXY(500, 800)

// 获取屏幕尺寸后计算坐标
size, err := driver.WindowSize()
if err == nil {
    centerX := float64(size.Width) / 2
    centerY := float64(size.Height) / 2
    err = driver.TapAbsXY(centerX, centerY)
}
```

#### 选择器点击

通过文本或其他选择器进行点击。

```go
// 通过文本点击
err := driver.TapBySelector("登录")
err := driver.TapBySelector("text=登录")

// 通过资源ID点击（Android）
err := driver.TapBySelector("resource-id=com.example:id/login_button")

// 通过XPath点击（Web）
err := driver.TapBySelector("//button[@id='login']")

// 通过CSS选择器点击（Web）
err := driver.TapBySelector("#login-button")
```

#### 双击操作

```go
// 双击指定坐标
err := driver.DoubleTap(100, 200)

// 双击相对坐标
err := driver.DoubleTap(0.5, 0.5)
```

#### 长按操作

```go
// 长按指定坐标
err := driver.TouchAndHold(150, 300)

// 带选项的长按
err := driver.TouchAndHold(150, 300,
    option.WithDuration(2*time.Second),
)
```

### 滑动操作

#### 基础滑动

```go
// 从下往上滑动（向上滚动）
err := driver.Swipe(0.5, 0.8, 0.5, 0.2)

// 从上往下滑动（向下滚动）
err := driver.Swipe(0.5, 0.2, 0.5, 0.8)

// 从右往左滑动（向左翻页）
err := driver.Swipe(0.8, 0.5, 0.2, 0.5)

// 从左往右滑动（向右翻页）
err := driver.Swipe(0.2, 0.5, 0.8, 0.5)
```

#### 带选项的滑动

```go
// 慢速滑动
err := driver.Swipe(0.5, 0.8, 0.5, 0.2,
    option.WithDuration(2*time.Second),
)

// 快速滑动
err := driver.Swipe(0.5, 0.8, 0.5, 0.2,
    option.WithDuration(200*time.Millisecond),
)

// 多步滑动
err := driver.Swipe(0.5, 0.8, 0.5, 0.2,
    option.WithSteps(20),
)
```

#### 拖拽操作

```go
// 拖拽元素从一个位置到另一个位置
err := driver.Drag(0.2, 0.3, 0.8, 0.7)

// 带持续时间的拖拽
err := driver.Drag(0.2, 0.3, 0.8, 0.7,
    option.WithDuration(1*time.Second),
)
```

### 输入操作

#### 文本输入

```go
// 基础文本输入
err := driver.Input("Hello World")

// 输入中文
err := driver.Input("你好世界")

// 输入特殊字符
err := driver.Input("user@example.com")
err := driver.Input("P@ssw0rd123!")
```

#### 退格操作

```go
// 删除一个字符
err := driver.Backspace(1)

// 删除多个字符
err := driver.Backspace(5)

// 清空输入框（删除大量字符）
err := driver.Backspace(100)
```

#### 输入法设置

```go
// 设置输入法（Android）
err := driver.SetIme("com.google.android.inputmethod.latin/.LatinIME")

// 设置中文输入法
err := driver.SetIme("com.sohu.inputmethod.sogou/.SogouIME")
```

### 按键操作

#### 系统按键

```go
// Home 键
err := driver.Home()

// Back 键（Android）
err := driver.Back()

// 通用按键操作
err := driver.PressButton(types.DeviceButtonHome)
err := driver.PressButton(types.DeviceButtonBack)
err := driver.PressButton(types.DeviceButtonVolumeUp)
err := driver.PressButton(types.DeviceButtonVolumeDown)
```

#### 特殊按键

```go
// 电源键
err := driver.PressButton(types.DeviceButtonPower)

// 菜单键
err := driver.PressButton(types.DeviceButtonMenu)

// 搜索键
err := driver.PressButton(types.DeviceButtonSearch)
```

## 高级操作

### 智能操作

#### OCR 识别点击

```go
// 通过 OCR 识别文本并点击
err := xtDriver.TapOCR("登录")

// 使用正则表达式匹配
err := xtDriver.TapOCR(`\d{4}`, option.WithRegex(true))

// 选择特定索引的文本
err := xtDriver.TapOCR("按钮", option.WithIndex(1))
```

#### 计算机视觉点击

```go
// 通过 CV 识别 UI 元素并点击
err := xtDriver.TapCV("button", "登录按钮")

// 识别图标并点击
err := xtDriver.TapCV("icon", "设置图标")
```

#### 智能滑动查找

```go
// 滑动查找应用并点击
err := xtDriver.SwipeToTapApp("微信")

// 滑动查找文本并点击
err := xtDriver.SwipeToTapText("设置")

// 滑动查找多个文本中的一个
err := xtDriver.SwipeToTapTexts([]string{"登录", "Sign In", "ログイン"})
```

### 组合操作

#### 登录流程

```go
func performLogin(driver IDriver, username, password string) error {
    // 1. 点击用户名输入框
    err := driver.TapBySelector("用户名")
    if err != nil {
        return err
    }

    // 2. 输入用户名
    err = driver.Input(username)
    if err != nil {
        return err
    }

    // 3. 点击密码输入框
    err = driver.TapBySelector("密码")
    if err != nil {
        return err
    }

    // 4. 输入密码
    err = driver.Input(password)
    if err != nil {
        return err
    }

    // 5. 点击登录按钮
    err = driver.TapBySelector("登录")
    if err != nil {
        return err
    }

    return nil
}
```

#### 列表滚动查找

```go
func findInList(driver IDriver, targetText string) error {
    maxSwipes := 10

    for i := 0; i < maxSwipes; i++ {
        // 尝试点击目标文本
        err := driver.TapBySelector(targetText)
        if err == nil {
            return nil // 找到并点击成功
        }

        // 向上滑动继续查找
        err = driver.Swipe(0.5, 0.8, 0.5, 0.2)
        if err != nil {
            return err
        }

        // 等待滑动完成
        time.Sleep(500 * time.Millisecond)
    }

    return fmt.Errorf("text '%s' not found after %d swipes", targetText, maxSwipes)
}
```

#### 表单填写

```go
func fillForm(driver IDriver, formData map[string]string) error {
    for fieldName, value := range formData {
        // 点击字段
        err := driver.TapBySelector(fieldName)
        if err != nil {
            return fmt.Errorf("failed to tap field %s: %w", fieldName, err)
        }

        // 清空现有内容
        err = driver.Backspace(50)
        if err != nil {
            return fmt.Errorf("failed to clear field %s: %w", fieldName, err)
        }

        // 输入新值
        err = driver.Input(value)
        if err != nil {
            return fmt.Errorf("failed to input value for field %s: %w", fieldName, err)
        }
    }

    return nil
}
```

## 应用管理

### 应用生命周期

#### 启动应用

```go
// 启动应用
err := driver.AppLaunch("com.example.app")

// 启动系统应用
err := driver.AppLaunch("com.android.settings")  // Android 设置
err := driver.AppLaunch("com.apple.Preferences") // iOS 设置
```

#### 终止应用

```go
// 终止应用
terminated, err := driver.AppTerminate("com.example.app")
if err != nil {
    return err
}

if terminated {
    fmt.Println("App terminated successfully")
} else {
    fmt.Println("App was not running")
}
```

#### 清理应用数据

```go
// 清理应用数据和缓存（Android）
err := driver.AppClear("com.example.app")
```

### 应用信息

#### 获取前台应用

```go
// 获取当前前台应用信息
appInfo, err := driver.ForegroundInfo()
if err != nil {
    return err
}

fmt.Printf("Current app: %s (%s)\n", appInfo.Name, appInfo.PackageName)
```

#### 列出已安装应用

```go
// 列出所有已安装的应用（需要扩展功能）
packages, err := xtDriver.ListPackages()
if err != nil {
    return err
}

for _, pkg := range packages {
    fmt.Printf("Package: %s\n", pkg)
}
```

## 屏幕操作

### 截图操作

#### 基础截图

```go
// 获取屏幕截图
screenshot, err := driver.ScreenShot()
if err != nil {
    return err
}

// 保存截图到文件
err = ioutil.WriteFile("screenshot.png", screenshot.Bytes(), 0644)
```

#### 带选项的截图

```go
// 高质量截图
screenshot, err := driver.ScreenShot(
    option.WithQuality(100),
)

// 指定格式截图
screenshot, err := driver.ScreenShot(
    option.WithFormat("jpeg"),
)
```

### 屏幕录制

```go
// 开始录制
videoPath, err := driver.ScreenRecord(
    option.WithDuration(30*time.Second),
    option.WithBitRate(4000000),
)
if err != nil {
    return err
}

fmt.Printf("Video saved to: %s\n", videoPath)
```

### 屏幕信息

#### 获取屏幕尺寸

```go
// 获取屏幕尺寸
size, err := driver.WindowSize()
if err != nil {
    return err
}

fmt.Printf("Screen size: %dx%d\n", size.Width, size.Height)
```

#### 获取屏幕方向

```go
// 获取当前方向
orientation, err := driver.Orientation()
if err != nil {
    return err
}

fmt.Printf("Orientation: %s\n", orientation)

// 获取旋转角度
rotation, err := driver.Rotation()
if err != nil {
    return err
}

fmt.Printf("Rotation: %d degrees\n", rotation)
```

#### 设置屏幕方向

```go
// 设置为横屏
err := driver.SetRotation(types.RotationLandscape)

// 设置为竖屏
err := driver.SetRotation(types.RotationPortrait)

// 设置为倒置横屏
err := driver.SetRotation(types.RotationLandscapeFlipped)
```

## 文件操作

### 文件传输

#### 推送文件到设备

```go
// 推送单个文件
err := driver.PushFile("/local/path/file.txt", "/sdcard/Download/")

// 推送图片
err := driver.PushImage("/local/path/image.jpg")
```

#### 从设备拉取文件

```go
// 拉取文件到本地
err := driver.PullFiles("/local/download/", "/sdcard/Download/")

// 拉取图片
err := driver.PullImages("/local/images/")
```

#### 清理文件

```go
// 清理指定路径的文件
err := driver.ClearFiles("/sdcard/Download/temp.txt")

// 清理图片
err := driver.ClearImages()
```

## Web 操作

### 页面导航

```go
// 导航到URL（仅Web驱动）
if webDriver, ok := driver.(*BrowserDriver); ok {
    err := webDriver.NavigateTo("https://example.com")

    // 刷新页面
    err = webDriver.Refresh()

    // 后退
    err = webDriver.GoBack()

    // 前进
    err = webDriver.GoForward()
}
```

### 元素操作

#### 悬停操作

```go
// 悬停在元素上（主要用于Web）
err := driver.HoverBySelector("#menu-item")

// 悬停在坐标上
err := driver.HoverXY(0.5, 0.3)
```

#### 右键点击

```go
// 右键点击坐标
err := driver.SecondaryClick(100, 200)

// 右键点击元素
err := driver.SecondaryClickBySelector("#context-menu-target")
```

### JavaScript 执行

```go
// 执行JavaScript（仅Web驱动）
if webDriver, ok := driver.(*BrowserDriver); ok {
    result, err := webDriver.ExecuteScript("return document.title;")
    if err == nil {
        fmt.Printf("Page title: %s\n", result)
    }

    // 执行复杂脚本
    script := `
        var element = document.getElementById('target');
        element.style.backgroundColor = 'red';
        return element.innerText;
    `
    result, err = webDriver.ExecuteScript(script)
}
```

## 等待和同步

### 显式等待

```go
// 等待元素出现
err := waitForElement(driver, "登录", 10*time.Second)

func waitForElement(driver IDriver, selector string, timeout time.Duration) error {
    deadline := time.Now().Add(timeout)

    for time.Now().Before(deadline) {
        err := driver.TapBySelector(selector)
        if err == nil {
            return nil // 元素找到
        }

        time.Sleep(500 * time.Millisecond)
    }

    return fmt.Errorf("element '%s' not found within %v", selector, timeout)
}
```

### 条件等待

```go
// 等待条件满足
err := waitForCondition(func() bool {
    // 检查某个条件
    appInfo, err := driver.ForegroundInfo()
    return err == nil && appInfo.PackageName == "com.target.app"
}, 30*time.Second)

func waitForCondition(condition func() bool, timeout time.Duration) error {
    deadline := time.Now().Add(timeout)

    for time.Now().Before(deadline) {
        if condition() {
            return nil
        }
        time.Sleep(1 * time.Second)
    }

    return fmt.Errorf("condition not met within %v", timeout)
}
```

### 智能等待

```go
// 等待页面加载完成
func waitForPageLoad(driver IDriver) error {
    // 等待一段时间让页面开始加载
    time.Sleep(1 * time.Second)

    // 连续检查页面是否稳定
    var lastScreenshot []byte
    stableCount := 0

    for i := 0; i < 10; i++ {
        screenshot, err := driver.ScreenShot()
        if err != nil {
            return err
        }

        currentScreenshot := screenshot.Bytes()

        if lastScreenshot != nil && bytes.Equal(lastScreenshot, currentScreenshot) {
            stableCount++
            if stableCount >= 3 {
                return nil // 页面稳定
            }
        } else {
            stableCount = 0
        }

        lastScreenshot = currentScreenshot
        time.Sleep(1 * time.Second)
    }

    return fmt.Errorf("page did not stabilize")
}
```

## 错误处理

### 重试机制

```go
// 带重试的操作
func performWithRetry(operation func() error, maxRetries int) error {
    var lastErr error

    for i := 0; i < maxRetries; i++ {
        err := operation()
        if err == nil {
            return nil
        }

        lastErr = err

        // 指数退避
        waitTime := time.Duration(math.Pow(2, float64(i))) * time.Second
        time.Sleep(waitTime)
    }

    return fmt.Errorf("operation failed after %d retries: %w", maxRetries, lastErr)
}

// 使用示例
err := performWithRetry(func() error {
    return driver.TapBySelector("登录")
}, 3)
```

### 异常恢复

```go
// 操作失败时的恢复策略
func performWithRecovery(driver IDriver, operation func() error) error {
    err := operation()
    if err == nil {
        return nil
    }

    // 尝试恢复策略
    log.Warn().Err(err).Msg("operation failed, attempting recovery")

    // 策略1: 返回主屏幕
    if err := driver.Home(); err != nil {
        log.Error().Err(err).Msg("failed to go home")
    }

    // 策略2: 等待一段时间
    time.Sleep(2 * time.Second)

    // 策略3: 重新尝试操作
    return operation()
}
```

## 性能优化

### 批量操作

```go
// 批量执行操作以提高性能
func performBatchOperations(driver IDriver, operations []func() error) error {
    // 如果驱动支持批量模式
    if batchDriver, ok := driver.(interface{ BeginBatch(); EndBatch() }); ok {
        batchDriver.BeginBatch()
        defer batchDriver.EndBatch()
    }

    for i, operation := range operations {
        err := operation()
        if err != nil {
            return fmt.Errorf("batch operation %d failed: %w", i, err)
        }
    }

    return nil
}
```

### 缓存优化

```go
// 缓存屏幕截图以避免重复获取
type ScreenshotCache struct {
    screenshot *bytes.Buffer
    timestamp  time.Time
    ttl        time.Duration
}

func (c *ScreenshotCache) GetScreenshot(driver IDriver) (*bytes.Buffer, error) {
    if c.screenshot != nil && time.Since(c.timestamp) < c.ttl {
        return c.screenshot, nil
    }

    screenshot, err := driver.ScreenShot()
    if err != nil {
        return nil, err
    }

    c.screenshot = screenshot
    c.timestamp = time.Now()

    return screenshot, nil
}
```

## 最佳实践

### 1. 操作前检查

```go
// 操作前检查设备状态
func checkDeviceReady(driver IDriver) error {
    status, err := driver.Status()
    if err != nil {
        return fmt.Errorf("failed to get device status: %w", err)
    }

    if status.State != "online" {
        return fmt.Errorf("device not ready: %s", status.State)
    }

    return nil
}
```

### 2. 操作后验证

```go
// 操作后验证结果
func tapAndVerify(driver IDriver, selector string, expectedResult func() bool) error {
    err := driver.TapBySelector(selector)
    if err != nil {
        return err
    }

    // 等待操作生效
    time.Sleep(1 * time.Second)

    // 验证结果
    if !expectedResult() {
        return fmt.Errorf("tap operation did not produce expected result")
    }

    return nil
}
```

### 3. 资源清理

```go
// 确保资源清理
func performOperationWithCleanup(driver IDriver, operation func() error) error {
    // 记录初始状态
    initialApp, _ := driver.ForegroundInfo()

    defer func() {
        // 恢复到初始状态
        if initialApp != nil {
            driver.AppLaunch(initialApp.PackageName)
        }
    }()

    return operation()
}
```

### 4. 日志记录

```go
// 详细的操作日志
func loggedTap(driver IDriver, x, y float64) error {
    log.Info().
        Float64("x", x).
        Float64("y", y).
        Msg("performing tap operation")

    start := time.Now()
    err := driver.TapXY(x, y)
    elapsed := time.Since(start)

    if err != nil {
        log.Error().
            Err(err).
            Float64("x", x).
            Float64("y", y).
            Dur("elapsed", elapsed).
            Msg("tap operation failed")
    } else {
        log.Info().
            Float64("x", x).
            Float64("y", y).
            Dur("elapsed", elapsed).
            Msg("tap operation completed")
    }

    return err
}
```

## 参考资料

- [Android UiAutomator2 文档](https://developer.android.com/training/testing/ui-automator)
- [iOS WebDriverAgent 文档](https://github.com/appium/WebDriverAgent)
- [WebDriver 规范](https://w3c.github.io/webdriver/)
- [Appium 文档](https://appium.io/docs/)