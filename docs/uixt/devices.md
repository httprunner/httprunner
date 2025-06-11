# 设备管理文档

## 概述

HttpRunner UIXT 提供统一的设备管理接口，支持 Android、iOS、HarmonyOS 和 Web 浏览器等多种平台的设备发现、连接和管理。

## 设备接口

### IDevice 核心接口

所有设备都实现统一的 `IDevice` 接口：

```go
type IDevice interface {
    UUID() string                    // 设备唯一标识
    NewDriver(driverType DriverType) (IDriver, error) // 创建驱动
}
```

## Android 设备

### 环境准备

#### Android SDK 安装

```bash
# 下载并安装 Android SDK
export ANDROID_HOME=/path/to/android-sdk
export PATH=$PATH:$ANDROID_HOME/platform-tools
export PATH=$PATH:$ANDROID_HOME/tools

# 验证安装
adb version
```

#### 真机配置

1. **开启开发者选项**
   - 进入设置 → 关于手机
   - 连续点击版本号 7 次

2. **启用 USB 调试**
   - 进入设置 → 开发者选项
   - 开启 USB 调试

3. **连接设备**
   ```bash
   # 连接设备并授权
   adb devices

   # 如果显示 unauthorized，在设备上点击允许
   ```

#### 模拟器配置

```bash
# 创建 AVD
avdmanager create avd -n test_device -k "system-images;android-30;google_apis;x86_64"

# 启动模拟器
emulator -avd test_device

# 验证连接
adb devices
```

### 设备创建

```go
import "github.com/httprunner/httprunner/v5/uixt/option"

// 基础创建
device, err := uixt.NewAndroidDevice(
    option.WithSerialNumber("device_serial"),
)

// 高级配置
device, err := uixt.NewAndroidDevice(
    option.WithSerialNumber("emulator-5554"),
    option.WithAdbLogOn(true),                    // 启用 ADB 日志
    option.WithReset(true),                       // 重置设备状态
    option.WithSystemPort(8200),                  // 系统端口
    option.WithDevicePort(6790),                  // 设备端口
    option.WithForwardPort(8080),                 // 端口转发
    option.WithInstallApp("/path/to/app.apk"),    // 安装应用
    option.WithGrantPermissions(true),            // 授予权限
    option.WithSkipServerInstallation(false),     // 跳过服务器安装
    option.WithUiAutomator2Timeout(60),          // UiAutomator2 超时
)
```

### 设备发现

```go
// 发现所有连接的 Android 设备
devices, err := uixt.DiscoverAndroidDevices()
for _, device := range devices {
    fmt.Printf("Found device: %s\n", device.UUID())
}

// 发现模拟器
emulators, err := uixt.DiscoverAndroidEmulators()
```

### 配置选项

| 选项 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| `WithSerialNumber` | string | 设备序列号 | 必需 |
| `WithAdbLogOn` | bool | 启用 ADB 日志 | false |
| `WithReset` | bool | 重置设备状态 | false |
| `WithSystemPort` | int | UiAutomator2 系统端口 | 8200 |
| `WithDevicePort` | int | 设备端口 | 6790 |
| `WithForwardPort` | int | 端口转发 | 0 |
| `WithInstallApp` | string | 安装应用路径 | "" |
| `WithGrantPermissions` | bool | 自动授予权限 | false |
| `WithSkipServerInstallation` | bool | 跳过服务器安装 | false |
| `WithUiAutomator2Timeout` | int | UiAutomator2 超时(秒) | 60 |

### Android 特有功能

#### 应用管理

```go
// 应用安装
err = driver.InstallApp("/path/to/app.apk")
err = driver.InstallApp("/path/to/app.apk", option.WithForceInstall(true))

// 应用卸载
err = driver.UninstallApp("com.example.app")
err = driver.UninstallApp("com.example.app", option.WithKeepData(true))

// 应用信息
appInfo, err := driver.GetAppInfo("com.example.app")
installed, err := driver.IsAppInstalled("com.example.app")
permissions, err := driver.GetAppPermissions("com.example.app")
```

#### 权限管理

```go
// 授予权限
err = driver.GrantPermission("com.example.app", "android.permission.CAMERA")

// 撤销权限
err = driver.RevokePermission("com.example.app", "android.permission.CAMERA")

// 批量授予权限
permissions := []string{
    "android.permission.CAMERA",
    "android.permission.RECORD_AUDIO",
    "android.permission.ACCESS_FINE_LOCATION",
}
err = driver.GrantPermissions("com.example.app", permissions)
```

#### 系统设置

```go
// WiFi 操作
err = driver.EnableWiFi()
err = driver.DisableWiFi()
err = driver.ConnectWiFi("SSID", "password")

// 移动数据操作
err = driver.EnableMobileData()
err = driver.DisableMobileData()

// 飞行模式
err = driver.EnableAirplaneMode()
err = driver.DisableAirplaneMode()
```

### 设备信息

```go
// 获取设备信息
info, err := device.DeviceInfo()
fmt.Printf("Device: %s %s\n", info.Brand, info.Model)
fmt.Printf("Android: %s\n", info.Version)

// 获取电池信息
battery, err := device.BatteryInfo()
fmt.Printf("Battery: %d%%\n", battery.Level)

// 获取屏幕尺寸
size, err := device.WindowSize()
fmt.Printf("Screen: %dx%d\n", size.Width, size.Height)
```

## iOS 设备

### 环境准备

#### Xcode 和开发者工具

```bash
# 安装 Xcode（从 App Store）
# 安装命令行工具
xcode-select --install

# 安装 ios-deploy
npm install -g ios-deploy

# 验证安装
ios-deploy --version
```

#### WebDriverAgent 配置

```bash
# 克隆 WebDriverAgent
git clone https://github.com/appium/WebDriverAgent.git
cd WebDriverAgent

# 配置开发者证书
# 在 Xcode 中打开 WebDriverAgent.xcodeproj
# 设置开发团队和签名证书

# 构建并安装到设备
xcodebuild -project WebDriverAgent.xcodeproj -scheme WebDriverAgentRunner -destination 'id=device_udid' test
```

#### 真机配置

1. **启用开发者模式**
   - 连接设备到 Mac
   - 在设备上信任开发者证书

2. **设备信任**
   - 设置 → 通用 → VPN与设备管理
   - 信任开发者应用

3. **获取设备 UDID**
   ```bash
   # 使用 Xcode
   xcrun simctl list devices

   # 使用 idevice_id
   idevice_id -l
   ```

#### 模拟器配置

```bash
# 列出可用的模拟器
xcrun simctl list devices

# 创建新模拟器
xcrun simctl create "iPhone 14" "iPhone 14" "iOS 16.0"

# 启动模拟器
xcrun simctl boot "iPhone 14"

# 安装应用到模拟器
xcrun simctl install booted /path/to/app.app
```

### 设备创建

```go
// 基础创建
device, err := uixt.NewIOSDevice(
    option.WithUDID("device_udid"),
)

// 高级配置
device, err := uixt.NewIOSDevice(
    option.WithUDID("00008030-001234567890123A"),
    option.WithWDAPort(8700),                     // WDA 端口
    option.WithWDAMjpegPort(8800),               // MJPEG 端口
    option.WithResetHomeOnStartup(false),        // 启动时不回到主屏
    option.WithPreventWDAAttachments(true),      // 防止 WDA 附件
    option.WithWDAStartupTimeout(120),           // WDA 启动超时
    option.WithWDAConnectionTimeout(60),         // WDA 连接超时
    option.WithWDACommandTimeout(30),            // WDA 命令超时
    option.WithAcceptAlerts(true),               // 自动接受弹窗
    option.WithDismissAlerts(false),             // 自动关闭弹窗
)
```

### 设备发现

```go
// 发现所有连接的 iOS 设备
devices, err := uixt.DiscoverIOSDevices()
for _, device := range devices {
    fmt.Printf("Found device: %s\n", device.UUID())
}

// 发现模拟器
simulators, err := uixt.DiscoverIOSSimulators()

// 按条件筛选设备
realDevices, err := uixt.DiscoverIOSDevices(uixt.DeviceFilter{
    DeviceType: "real",
    IOSVersion: "16.0+",
})
```

### 配置选项

| 选项 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| `WithUDID` | string | 设备 UDID | 必需 |
| `WithWDAPort` | int | WebDriverAgent 端口 | 8700 |
| `WithWDAMjpegPort` | int | MJPEG 流端口 | 8800 |
| `WithResetHomeOnStartup` | bool | 启动时回到主屏 | true |
| `WithPreventWDAAttachments` | bool | 防止 WDA 附件 | false |
| `WithWDAStartupTimeout` | int | WDA 启动超时(秒) | 120 |
| `WithWDAConnectionTimeout` | int | WDA 连接超时(秒) | 60 |
| `WithWDACommandTimeout` | int | WDA 命令超时(秒) | 30 |
| `WithAcceptAlerts` | bool | 自动接受弹窗 | false |
| `WithDismissAlerts` | bool | 自动关闭弹窗 | false |

### iOS 特有功能

#### WebDriverAgent 管理

```go
// 启动 WDA
err = device.StartWDA()

// 停止 WDA
err = device.StopWDA()

// 检查 WDA 状态
isRunning := device.IsWDARunning()

// 重启 WDA
err = device.RestartWDA()

// 获取 WDA 状态
status, err := device.GetWDAStatus()
```

#### 应用管理

```go
// 应用安装（需要开发者证书）
err = driver.InstallApp("/path/to/app.ipa")

// 应用卸载
err = driver.UninstallApp("com.example.app")

// 应用信息
appInfo, err := driver.GetAppInfo("com.example.app")
installed, err := driver.IsAppInstalled("com.example.app")

// 应用状态
state, err := driver.GetAppState("com.example.app")
// 0: not installed, 1: not running, 2: running in background, 4: running in foreground
```

#### 系统操作

```go
// Siri 操作
err = driver.ActivateSiri("打开设置")

// 锁定/解锁
err = driver.Lock()
err = driver.Unlock()

// 摇晃设备
err = driver.Shake()

// 音量控制
err = driver.VolumeUp()
err = driver.VolumeDown()

// 截图和录制
screenshot, err := driver.ScreenShot()
err = driver.StartScreenRecord()
videoPath, err := driver.StopScreenRecord()
```

#### 设备信息

```go
// 获取设备信息
info, err := device.DeviceInfo()
fmt.Printf("Device: %s %s\n", info.Model, info.Name)
fmt.Printf("iOS: %s\n", info.Version)

// 获取电池信息
battery, err := device.BatteryInfo()
fmt.Printf("Battery: %d%%, State: %s\n", battery.Level, battery.State)

// 获取屏幕信息
size, err := device.WindowSize()
scale, err := device.GetScreenScale()
```

## HarmonyOS 设备

### 环境准备

#### HarmonyOS SDK 安装

```bash
# 下载并安装 HarmonyOS SDK
export HARMONY_HOME=/path/to/harmony-sdk
export PATH=$PATH:$HARMONY_HOME/toolchains

# 验证安装
hdc version
```

#### 设备配置

1. **开启开发者模式**
   - 进入设置 → 关于手机
   - 连续点击版本号 7 次

2. **启用 USB 调试**
   - 进入设置 → 系统和更新 → 开发人员选项
   - 开启 USB 调试

3. **连接设备**
   ```bash
   # 连接设备并授权
   hdc list targets

   # 如果显示 unauthorized，在设备上点击允许
   ```

### 设备创建

```go
// 基础创建
device, err := uixt.NewHarmonyDevice(
    option.WithConnectKey("device_connect_key"),
)

// 高级配置
device, err := uixt.NewHarmonyDevice(
    option.WithConnectKey("192.168.1.100:5555"),
    option.WithHDCLogOn(true),                    // 启用 HDC 日志
    option.WithSystemPort(9200),                  // 系统端口
    option.WithDevicePort(6790),                  // 设备端口
    option.WithHDCTimeout(60),                    // HDC 超时
)
```

### 设备发现

```go
// 发现所有连接的 HarmonyOS 设备
devices, err := uixt.DiscoverHarmonyDevices()
for _, device := range devices {
    fmt.Printf("Found device: %s\n", device.UUID())
}

// 网络设备发现
networkDevices, err := uixt.DiscoverHarmonyNetworkDevices()
```

### 配置选项

| 选项 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| `WithConnectKey` | string | 设备连接密钥 | 必需 |
| `WithHDCLogOn` | bool | 启用 HDC 日志 | false |
| `WithSystemPort` | int | 系统端口 | 9200 |
| `WithDevicePort` | int | 设备端口 | 6790 |
| `WithHDCTimeout` | int | HDC 超时(秒) | 60 |

### HarmonyOS 特有功能

#### 应用管理

```go
// 应用安装
err = driver.InstallApp("/path/to/app.hap")

// 应用卸载
err = driver.UninstallApp("com.example.harmony.app")

// 应用信息
appInfo, err := driver.GetAppInfo("com.example.harmony.app")
installed, err := driver.IsAppInstalled("com.example.harmony.app")
```

#### 分布式操作

```go
// 设备协同
err = driver.ConnectDistributedDevice("target_device_id")
err = driver.DisconnectDistributedDevice("target_device_id")

// 跨设备应用迁移
err = driver.MigrateApp("com.example.app", "target_device_id")
```

#### 原子化服务

```go
// 启动原子化服务
err = driver.LaunchAtomicService("service_id", map[string]interface{}{
    "param1": "value1",
    "param2": "value2",
})

// 停止原子化服务
err = driver.StopAtomicService("service_id")
```

## Web 浏览器设备

### 环境准备

#### 浏览器驱动安装

```bash
# Chrome
# 下载 ChromeDriver 并添加到 PATH
wget https://chromedriver.storage.googleapis.com/latest/chromedriver_mac64.zip
unzip chromedriver_mac64.zip
mv chromedriver /usr/local/bin/

# Firefox
# 下载 GeckoDriver
wget https://github.com/mozilla/geckodriver/releases/download/v0.33.0/geckodriver-v0.33.0-macos.tar.gz
tar -xzf geckodriver-v0.33.0-macos.tar.gz
mv geckodriver /usr/local/bin/

# Safari (macOS only)
# 启用开发者菜单
# Safari → 偏好设置 → 高级 → 在菜单栏中显示"开发"菜单
# 开发 → 允许远程自动化

# Edge
# 下载 EdgeDriver
wget https://msedgedriver.azureedge.net/latest/edgedriver_mac64.zip
```

### 设备创建

```go
// Chrome 浏览器
device, err := uixt.NewBrowserDevice(
    option.WithBrowserID("chrome"),
)

// 高级配置
device, err := uixt.NewBrowserDevice(
    option.WithBrowserID("chrome"),
    option.WithHeadless(false),                   // 非无头模式
    option.WithWindowSize(1920, 1080),           // 窗口大小
    option.WithUserAgent("custom-agent"),         // 自定义 User-Agent
    option.WithProxy("http://proxy:8080"),        // 代理设置
    option.WithExtensions([]string{"ext1", "ext2"}), // 扩展
    option.WithDownloadDir("/path/to/downloads"), // 下载目录
    option.WithIncognito(true),                   // 隐私模式
)
```

### 支持的浏览器

| 浏览器 | ID | 驱动 | 说明 |
|--------|----|----- |------|
| Chrome | `chrome` | ChromeDriver | Google Chrome |
| Firefox | `firefox` | GeckoDriver | Mozilla Firefox |
| Safari | `safari` | SafariDriver | Apple Safari (macOS) |
| Edge | `edge` | EdgeDriver | Microsoft Edge |

### 配置选项

| 选项 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| `WithBrowserID` | string | 浏览器标识 | 必需 |
| `WithHeadless` | bool | 无头模式 | true |
| `WithWindowSize` | int, int | 窗口大小 | 1280x720 |
| `WithUserAgent` | string | User-Agent | 默认 |
| `WithProxy` | string | 代理地址 | 无 |
| `WithExtensions` | []string | 扩展列表 | 无 |
| `WithDownloadDir` | string | 下载目录 | 默认 |
| `WithIncognito` | bool | 隐私模式 | false |

### Web 特有功能

#### 页面管理

```go
// 页面导航
err = driver.NavigateTo("https://example.com")
err = driver.Refresh()
err = driver.GoBack()
err = driver.GoForward()

// 页面信息
title, err := driver.GetTitle()
url, err := driver.GetCurrentURL()
source, err := driver.GetPageSource()
```

#### 标签页管理

```go
// 标签页操作
err = driver.NewTab()
err = driver.CloseTab(1)
err = driver.SwitchToTab(0)

// 获取标签页信息
tabs, err := driver.GetTabs()
currentTab, err := driver.GetCurrentTab()
```

#### Cookie 管理

```go
// Cookie 操作
cookies, err := driver.GetCookies()
err = driver.SetCookie("name", "value", "domain.com")
err = driver.DeleteCookie("name")
err = driver.DeleteAllCookies()
```

#### JavaScript 执行

```go
// 执行 JavaScript
result, err := driver.ExecuteScript("return document.title;")
err = driver.ExecuteAsyncScript("callback(arguments[0]);", "test")

// 注入脚本
err = driver.InjectScript("console.log('injected');")
```

## 设备管理工具

### 设备发现工具

```go
// 发现所有平台的设备
allDevices, err := uixt.DiscoverAllDevices()
for platform, devices := range allDevices {
    fmt.Printf("Platform: %s\n", platform)
    for _, device := range devices {
        fmt.Printf("  Device: %s\n", device.UUID())
    }
}

// 按平台发现
androidDevices, err := uixt.DiscoverDevicesByPlatform("android")
iosDevices, err := uixt.DiscoverDevicesByPlatform("ios")
```

### 设备选择工具

```go
// 交互式设备选择
device, err := uixt.SelectDeviceInteractively()

// 按条件选择设备
device, err := uixt.SelectDevice(uixt.DeviceFilter{
    Platform: "android",
    Model:    "Pixel",
    Online:   true,
    Version:  "11+",
})

// 智能选择最佳设备
device, err := uixt.SelectBestDevice(uixt.DevicePreference{
    PreferReal:      true,  // 优先真机
    PreferHighRes:   true,  // 优先高分辨率
    PreferNewVersion: true, // 优先新版本
})
```

### 设备健康检查

```go
// 检查设备健康状态
health, err := device.HealthCheck()
if health.IsHealthy {
    fmt.Println("Device is healthy")
} else {
    fmt.Printf("Device issues: %v\n", health.Issues)
}

// 修复设备问题
err = device.Repair()

// 设备诊断
diagnosis, err := device.Diagnose()
fmt.Printf("Diagnosis: %s\n", diagnosis.Report)
```

## 设备状态管理

### 设备状态

```go
// 获取设备状态
status, err := device.Status()
fmt.Printf("Status: %s\n", status.State) // online, offline, unauthorized

// 等待设备就绪
err = device.WaitForReady(30 * time.Second)

// 检查设备连接
isConnected := device.IsConnected()

// 设备可用性检查
isAvailable := device.IsAvailable()
```

### 设备重置

```go
// 软重置（重启应用）
err = device.SoftReset()

// 硬重置（重启设备）
err = device.HardReset()

// 恢复出厂设置（仅 Android）
err = device.FactoryReset()

// 清理设备缓存
err = device.ClearCache()
```

## 多设备管理

### 设备池

```go
// 创建设备池
pool := uixt.NewDevicePool()

// 添加设备到池
pool.AddDevice(androidDevice)
pool.AddDevice(iosDevice)
pool.AddDevice(harmonyDevice)

// 从池中获取可用设备
device, err := pool.AcquireDevice(uixt.DeviceFilter{
    Platform: "android",
})
defer pool.ReleaseDevice(device)

// 并行执行任务
results := pool.ExecuteParallel(func(device IDevice) interface{} {
    // 在设备上执行任务
    return performTask(device)
})

// 设备池统计
stats := pool.GetStats()
fmt.Printf("Total: %d, Available: %d, InUse: %d\n",
    stats.Total, stats.Available, stats.InUse)
```

### 设备同步

```go
// 同步多个设备的操作
sync := uixt.NewDeviceSync()
sync.AddDevice(device1)
sync.AddDevice(device2)
sync.AddDevice(device3)

// 同步执行操作
err = sync.Execute(func(device IDevice) error {
    return device.TapXY(0.5, 0.5)
})

// 等待所有设备完成
err = sync.WaitForAll(30 * time.Second)
```

### 设备集群

```go
// 创建设备集群
cluster := uixt.NewDeviceCluster()

// 添加设备组
cluster.AddGroup("android", androidDevices)
cluster.AddGroup("ios", iosDevices)

// 按组执行任务
results, err := cluster.ExecuteByGroup("android", func(device IDevice) interface{} {
    return performAndroidTask(device)
})

// 负载均衡
device, err := cluster.GetLeastBusyDevice()
```

## 最佳实践

### 1. 设备选择策略

```go
// 优先选择真机，其次模拟器
func selectBestDevice() (IDevice, error) {
    // 先尝试真机
    devices, err := uixt.DiscoverAndroidDevices()
    if err == nil && len(devices) > 0 {
        return devices[0], nil
    }

    // 再尝试模拟器
    emulators, err := uixt.DiscoverAndroidEmulators()
    if err == nil && len(emulators) > 0 {
        return emulators[0], nil
    }

    return nil, fmt.Errorf("no available devices")
}
```

### 2. 设备资源管理

```go
// 使用 defer 确保资源释放
func useDevice() error {
    device, err := uixt.NewAndroidDevice(option.WithSerialNumber("device_serial"))
    if err != nil {
        return err
    }
    defer device.Cleanup() // 确保清理资源

    // 使用设备...
    return nil
}
```

### 3. 错误处理和重试

```go
// 带重试的设备操作
func performWithRetry(device IDevice, operation func() error) error {
    maxRetries := 3
    for i := 0; i < maxRetries; i++ {
        err := operation()
        if err == nil {
            return nil
        }

        // 检查是否是设备连接问题
        if isDeviceConnectionError(err) {
            // 尝试重新连接
            device.Reconnect()
        }

        time.Sleep(time.Duration(i+1) * time.Second)
    }
    return fmt.Errorf("operation failed after %d retries", maxRetries)
}
```

### 4. 设备监控

```go
// 设备监控
func monitorDevice(device IDevice) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        status, err := device.Status()
        if err != nil {
            log.Error("Failed to get device status: %v", err)
            continue
        }

        if status.State != "online" {
            log.Warn("Device %s is %s", device.UUID(), status.State)
            // 尝试修复
            device.Repair()
        }
    }
}
```

## 故障排除

### 常见问题

#### Android 设备

1. **设备未识别**
   ```bash
   # 检查 ADB 连接
   adb devices

   # 重启 ADB 服务
   adb kill-server
   adb start-server

   # 检查驱动程序
   # Windows: 更新设备驱动
   # macOS: 检查系统偏好设置中的安全性设置
   ```

2. **UiAutomator2 启动失败**
   ```bash
   # 检查端口占用
   netstat -an | grep 8200

   # 清理应用数据
   adb shell pm clear io.appium.uiautomator2.server
   adb shell pm clear io.appium.uiautomator2.server.test

   # 重新安装服务
   adb uninstall io.appium.uiautomator2.server
   adb uninstall io.appium.uiautomator2.server.test
   ```

3. **权限问题**
   ```bash
   # 检查 USB 调试权限
   adb shell settings get global development_settings_enabled

   # 授予应用权限
   adb shell pm grant com.example.app android.permission.CAMERA
   ```

#### iOS 设备

1. **WDA 启动失败**
   ```bash
   # 检查开发者证书
   security find-identity -v -p codesigning

   # 重新安装 WDA
   xcodebuild -project WebDriverAgent.xcodeproj -scheme WebDriverAgentRunner -destination 'id=device_udid' test

   # 检查设备信任
   # 设置 → 通用 → VPN与设备管理 → 信任开发者应用
   ```

2. **设备信任问题**
   - 在设备上信任开发者证书
   - 检查设备是否已解锁
   - 确保设备已配对

3. **网络连接问题**
   ```bash
   # 检查端口转发
   iproxy 8700 8700 device_udid

   # 测试 WDA 连接
   curl http://localhost:8700/status
   ```

#### HarmonyOS 设备

1. **HDC 连接失败**
   ```bash
   # 检查 HDC 连接
   hdc list targets

   # 重启 HDC 服务
   hdc kill
   hdc start

   # 检查网络连接（网络调试）
   ping device_ip
   ```

2. **应用安装失败**
   ```bash
   # 检查应用签名
   hdc shell bm dump -a

   # 清理应用数据
   hdc shell bm uninstall -n com.example.app
   ```

#### Web 浏览器

1. **驱动版本不匹配**
   ```bash
   # 检查浏览器版本
   google-chrome --version
   firefox --version

   # 更新驱动程序
   # 确保驱动版本与浏览器版本匹配
   ```

2. **端口冲突**
   ```bash
   # 查找占用端口的进程
   lsof -i :4444

   # 终止进程
   kill -9 <pid>
   ```

#### 通用问题

1. **端口冲突**
   ```bash
   # 查找占用端口的进程
   lsof -i :8700

   # 终止进程
   kill -9 <pid>

   # 使用不同端口
   device, err := uixt.NewIOSDevice(
       option.WithUDID("device_udid"),
       option.WithWDAPort(8701),
   )
   ```

2. **权限问题**
   ```bash
   # 检查文件权限
   ls -la /path/to/device/files

   # 修改权限
   chmod +x /path/to/executable

   # macOS 安全设置
   # 系统偏好设置 → 安全性与隐私 → 隐私 → 辅助功能
   ```

3. **内存不足**
   ```bash
   # 检查系统资源
   top
   free -h

   # 清理设备缓存
   device.ClearCache()

   # 重启设备
   device.HardReset()
   ```

## 参考资料

- [Android Debug Bridge (ADB)](https://developer.android.com/studio/command-line/adb)
- [WebDriverAgent](https://github.com/appium/WebDriverAgent)
- [HarmonyOS HDC](https://developer.harmonyos.com/cn/docs/documentation/doc-guides/ohos-debugging-and-testing-0000001263040487)
- [WebDriver 协议](https://w3c.github.io/webdriver/)
- [ChromeDriver](https://chromedriver.chromium.org/)
- [GeckoDriver](https://github.com/mozilla/geckodriver)