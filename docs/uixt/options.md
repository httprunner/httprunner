# 配置选项文档

## 概述

HttpRunner UIXT 提供了丰富的配置选项，支持设备配置、驱动配置、AI 服务配置等多个层面的定制化设置。本文档详细介绍所有可用的配置选项。

## 设备配置选项

### Android 设备配置

#### 基础选项

| 选项 | 类型 | 说明 | 默认值 | 示例 |
|------|------|------|--------|------|
| `WithSerialNumber` | string | 设备序列号 | 必需 | `"emulator-5554"` |
| `WithAdbLogOn` | bool | 启用 ADB 日志 | false | `true` |
| `WithReset` | bool | 重置设备状态 | false | `true` |

```go
device, err := uixt.NewAndroidDevice(
    option.WithSerialNumber("emulator-5554"),
    option.WithAdbLogOn(true),
    option.WithReset(true),
)
```

#### 网络选项

| 选项 | 类型 | 说明 | 默认值 | 示例 |
|------|------|------|--------|------|
| `WithSystemPort` | int | UiAutomator2 系统端口 | 8200 | `8200` |
| `WithDevicePort` | int | 设备端口 | 6790 | `6790` |
| `WithForwardPort` | int | 端口转发 | 0 | `8080` |
| `WithProxy` | string | 代理设置 | "" | `"http://proxy:8080"` |

```go
device, err := uixt.NewAndroidDevice(
    option.WithSerialNumber("device_serial"),
    option.WithSystemPort(8200),
    option.WithDevicePort(6790),
    option.WithForwardPort(8080),
    option.WithProxy("http://proxy.example.com:8080"),
)
```

#### 应用管理选项

| 选项 | 类型 | 说明 | 默认值 | 示例 |
|------|------|------|--------|------|
| `WithInstallApp` | string | 自动安装应用路径 | "" | `"/path/to/app.apk"` |
| `WithGrantPermissions` | bool | 自动授予权限 | false | `true` |
| `WithSkipServerInstallation` | bool | 跳过服务器安装 | false | `true` |
| `WithUiAutomator2Timeout` | int | UiAutomator2 超时(秒) | 60 | `120` |

```go
device, err := uixt.NewAndroidDevice(
    option.WithSerialNumber("device_serial"),
    option.WithInstallApp("/path/to/app.apk"),
    option.WithGrantPermissions(true),
    option.WithUiAutomator2Timeout(120),
)
```

### iOS 设备配置

#### 基础选项

| 选项 | 类型 | 说明 | 默认值 | 示例 |
|------|------|------|--------|------|
| `WithUDID` | string | 设备 UDID | 必需 | `"00008030-001234567890123A"` |
| `WithWDAPort` | int | WebDriverAgent 端口 | 8700 | `8700` |
| `WithWDAMjpegPort` | int | MJPEG 流端口 | 8800 | `8800` |

```go
device, err := uixt.NewIOSDevice(
    option.WithUDID("00008030-001234567890123A"),
    option.WithWDAPort(8700),
    option.WithWDAMjpegPort(8800),
)
```

#### WDA 配置选项

| 选项 | 类型 | 说明 | 默认值 | 示例 |
|------|------|------|--------|------|
| `WithResetHomeOnStartup` | bool | 启动时回到主屏 | true | `false` |
| `WithPreventWDAAttachments` | bool | 防止 WDA 附件 | false | `true` |
| `WithWDAStartupTimeout` | int | WDA 启动超时(秒) | 120 | `180` |
| `WithWDAConnectionTimeout` | int | WDA 连接超时(秒) | 60 | `90` |

```go
device, err := uixt.NewIOSDevice(
    option.WithUDID("device_udid"),
    option.WithResetHomeOnStartup(false),
    option.WithPreventWDAAttachments(true),
    option.WithWDAStartupTimeout(180),
    option.WithWDAConnectionTimeout(90),
)
```

### HarmonyOS 设备配置

| 选项 | 类型 | 说明 | 默认值 | 示例 |
|------|------|------|--------|------|
| `WithConnectKey` | string | 设备连接密钥 | 必需 | `"192.168.1.100:5555"` |
| `WithHDCLogOn` | bool | 启用 HDC 日志 | false | `true` |
| `WithSystemPort` | int | 系统端口 | 9200 | `9200` |

```go
device, err := uixt.NewHarmonyDevice(
    option.WithConnectKey("192.168.1.100:5555"),
    option.WithHDCLogOn(true),
    option.WithSystemPort(9200),
)
```

### Web 浏览器配置

#### 基础选项

| 选项 | 类型 | 说明 | 默认值 | 示例 |
|------|------|------|--------|------|
| `WithBrowserID` | string | 浏览器标识 | 必需 | `"chrome"` |
| `WithHeadless` | bool | 无头模式 | true | `false` |
| `WithWindowSize` | int, int | 窗口大小 | 1280x720 | `1920, 1080` |

```go
device, err := uixt.NewBrowserDevice(
    option.WithBrowserID("chrome"),
    option.WithHeadless(false),
    option.WithWindowSize(1920, 1080),
)
```

#### 高级选项

| 选项 | 类型 | 说明 | 默认值 | 示例 |
|------|------|------|--------|------|
| `WithUserAgent` | string | 自定义 User-Agent | 默认 | `"custom-agent"` |
| `WithProxy` | string | 代理地址 | 无 | `"http://proxy:8080"` |
| `WithExtensions` | []string | 扩展列表 | 无 | `[]string{"ext1", "ext2"}` |
| `WithDownloadDir` | string | 下载目录 | 默认 | `"/path/to/downloads"` |

```go
device, err := uixt.NewBrowserDevice(
    option.WithBrowserID("chrome"),
    option.WithUserAgent("custom-agent"),
    option.WithProxy("http://proxy:8080"),
    option.WithExtensions([]string{"extension1", "extension2"}),
    option.WithDownloadDir("/custom/download/path"),
)
```

## AI 服务配置

### LLM 服务配置

#### 基础配置

```go
// 使用单一模型
xtDriver, err := uixt.NewXTDriver(driver,
    option.WithLLMService(option.OPENAI_GPT_4O),
)
```

#### 高级配置

```go
// 混合模型配置
config := option.NewLLMServiceConfig(option.DOUBAO_1_5_THINKING_VISION_PRO_250428).
    WithPlannerModel(option.DOUBAO_1_5_UI_TARS_250328).
    WithAsserterModel(option.OPENAI_GPT_4O).
    WithQuerierModel(option.DEEPSEEK_R1_250528)

xtDriver, err := uixt.NewXTDriver(driver,
    option.WithLLMConfig(config),
)
```

#### 支持的模型

| 模型名称 | 特点 | 适用场景 |
|---------|------|----------|
| `DOUBAO_1_5_UI_TARS_250328` | UI 理解专业模型 | UI 元素识别和操作规划 |
| `DOUBAO_1_5_THINKING_VISION_PRO_250428` | 思考推理模型 | 复杂逻辑推理和断言 |
| `OPENAI_GPT_4O` | 高性能通用模型 | 全场景通用 |
| `DEEPSEEK_R1_250528` | 成本效益模型 | 大量查询场景 |

#### 推荐配置

```go
configs := option.RecommendedConfigurations()

// 混合优化配置（推荐）
config := configs["mixed_optimal"]

// 高性能配置
config := configs["high_performance"]

// 成本优化配置
config := configs["cost_effective"]

// UI 专注配置
config := configs["ui_focused"]

// 推理专注配置
config := configs["reasoning_focused"]
```

### CV 服务配置

| 选项 | 类型 | 说明 | 默认值 | 示例 |
|------|------|------|--------|------|
| `WithCVService` | CVServiceType | CV 服务类型 | 无 | `option.CVServiceTypeVEDEM` |

```go
xtDriver, err := uixt.NewXTDriver(driver,
    option.WithCVService(option.CVServiceTypeVEDEM),
)
```

## 操作配置选项

### 通用操作选项

#### 时间相关选项

| 选项 | 类型 | 说明 | 默认值 | 示例 |
|------|------|------|--------|------|
| `WithDuration` | time.Duration | 操作持续时间 | 默认 | `2*time.Second` |
| `WithTimeout` | time.Duration | 操作超时时间 | 30s | `60*time.Second` |
| `WithDelay` | time.Duration | 操作前延迟 | 0 | `500*time.Millisecond` |

```go
// 慢速滑动
err := driver.Swipe(0.5, 0.8, 0.5, 0.2,
    option.WithDuration(2*time.Second),
)

// 长按操作
err := driver.TouchAndHold(150, 300,
    option.WithDuration(3*time.Second),
)

// 带超时的操作
err := driver.TapBySelector("登录",
    option.WithTimeout(10*time.Second),
)
```

#### 精度相关选项

| 选项 | 类型 | 说明 | 默认值 | 示例 |
|------|------|------|--------|------|
| `WithSteps` | int | 滑动步数 | 默认 | `20` |
| `WithPressure` | float64 | 压力值(iOS) | 1.0 | `0.8` |
| `WithFrequency` | int | 操作频率 | 默认 | `60` |

```go
// 多步滑动
err := driver.Swipe(0.5, 0.8, 0.5, 0.2,
    option.WithSteps(50),
)

// 3D Touch (iOS)
err := driver.ForceTouch(100, 200,
    option.WithPressure(0.8),
)
```

### 截图选项

| 选项 | 类型 | 说明 | 默认值 | 示例 |
|------|------|------|--------|------|
| `WithQuality` | int | 图片质量 | 80 | `100` |
| `WithFormat` | string | 图片格式 | "png" | `"jpeg"` |
| `WithScale` | float64 | 缩放比例 | 1.0 | `0.5` |

```go
// 高质量截图
screenshot, err := driver.ScreenShot(
    option.WithQuality(100),
    option.WithFormat("png"),
)

// 缩放截图
screenshot, err := driver.ScreenShot(
    option.WithScale(0.5),
)
```

### 录制选项

| 选项 | 类型 | 说明 | 默认值 | 示例 |
|------|------|------|--------|------|
| `WithBitRate` | int | 比特率 | 4000000 | `8000000` |
| `WithVideoSize` | string | 视频尺寸 | 默认 | `"1280x720"` |
| `WithTimeLimit` | time.Duration | 录制时长 | 180s | `300*time.Second` |

```go
// 高质量录制
videoPath, err := driver.ScreenRecord(
    option.WithBitRate(8000000),
    option.WithVideoSize("1920x1080"),
    option.WithTimeLimit(300*time.Second),
)
```

### OCR 选项

| 选项 | 类型 | 说明 | 默认值 | 示例 |
|------|------|------|--------|------|
| `WithRegex` | bool | 使用正则表达式 | false | `true` |
| `WithIndex` | int | 文本索引 | 0 | `1` |
| `WithIgnoreCase` | bool | 忽略大小写 | false | `true` |

```go
// 正则表达式匹配
err := xtDriver.TapOCR(`\d{4}`,
    option.WithRegex(true),
)

// 选择第二个匹配项
err := xtDriver.TapOCR("按钮",
    option.WithIndex(1),
)

// 忽略大小写
err := xtDriver.TapOCR("LOGIN",
    option.WithIgnoreCase(true),
)
```

## 环境变量配置

### LLM 模型配置

#### 豆包模型

```bash
# 豆包思维视觉专业版
DOUBAO_1_5_THINKING_VISION_PRO_250428_BASE_URL=https://ark.cn-beijing.volces.com/api/v3
DOUBAO_1_5_THINKING_VISION_PRO_250428_API_KEY=your_doubao_api_key

# 豆包UI-TARS
DOUBAO_1_5_UI_TARS_250328_BASE_URL=https://ark.cn-beijing.volces.com/api/v3
DOUBAO_1_5_UI_TARS_250328_API_KEY=your_doubao_ui_tars_api_key
```

#### OpenAI 模型

```bash
# OpenAI GPT-4O
OPENAI_GPT_4O_BASE_URL=https://api.openai.com/v1
OPENAI_GPT_4O_API_KEY=your_openai_api_key
```

#### DeepSeek 模型

```bash
# DeepSeek
DEEPSEEK_R1_250528_BASE_URL=https://api.deepseek.com/v1
DEEPSEEK_R1_250528_API_KEY=your_deepseek_api_key
```

#### 默认配置

```bash
# 默认配置，当没有找到服务特定配置时使用
LLM_MODEL_NAME=doubao-1.5-thinking-vision-pro-250428
OPENAI_BASE_URL=https://ark.cn-beijing.volces.com/api/v3
OPENAI_API_KEY=your_default_api_key
```

### CV 服务配置

#### 火山引擎 VEDEM

```bash
# 火山引擎 VEDEM 配置
VEDEM_IMAGE_URL=https://visual.volcengineapi.com
VEDEM_IMAGE_AK=your_access_key
VEDEM_IMAGE_SK=your_secret_key
```

### 配置优先级

环境变量的加载优先级（从高到低）：

1. `.env` 文件（当前工作目录）
2. `~/.hrp/.env` 文件（全局用户配置）
3. 系统环境变量

```bash
# 项目级配置文件 .env
OPENAI_API_KEY=project_specific_key

# 用户级配置文件 ~/.hrp/.env
OPENAI_API_KEY=user_default_key

# 系统环境变量
export OPENAI_API_KEY=system_key
```

## 配置文件

### 项目配置文件

创建 `.env` 文件在项目根目录：

```bash
# .env
# LLM 服务配置
OPENAI_BASE_URL=https://api.openai.com/v1
OPENAI_API_KEY=your_openai_api_key

# CV 服务配置
VEDEM_IMAGE_URL=https://visual.volcengineapi.com
VEDEM_IMAGE_AK=your_access_key
VEDEM_IMAGE_SK=your_secret_key

# 设备配置
DEFAULT_ANDROID_SERIAL=emulator-5554
DEFAULT_IOS_UDID=00008030-001234567890123A
```

### 用户配置文件

创建 `~/.hrp/.env` 文件：

```bash
# ~/.hrp/.env
# 全局默认配置
OPENAI_API_KEY=your_global_api_key
VEDEM_IMAGE_AK=your_global_access_key
VEDEM_IMAGE_SK=your_global_secret_key
```

### YAML 配置文件

```yaml
# config.yaml
devices:
  android:
    serial: "emulator-5554"
    system_port: 8200
    device_port: 6790
    adb_log: true

  ios:
    udid: "00008030-001234567890123A"
    wda_port: 8700
    mjpeg_port: 8800
    reset_home: false

ai_services:
  llm:
    default_model: "doubao-1.5-thinking-vision-pro-250428"
    planner_model: "doubao-1.5-ui-tars-250328"
    asserter_model: "openai-gpt-4o"
    querier_model: "deepseek-r1-250528"

  cv:
    service_type: "vedem"

operations:
  default_timeout: 30
  screenshot_quality: 80
  video_bitrate: 4000000
```

## 动态配置

### 运行时配置

```go
// 运行时修改配置
func configureDriver(driver IDriver) error {
    // 设置超时
    driver.SetTimeout(60 * time.Second)

    // 设置重试次数
    driver.SetRetryCount(3)

    // 设置日志级别
    driver.SetLogLevel(log.DebugLevel)

    return nil
}
```

### 条件配置

```go
// 根据环境选择配置
func createDriverWithEnvironmentConfig(platform string) (*uixt.XTDriver, error) {
    var device uixt.IDevice
    var err error

    switch platform {
    case "android":
        if os.Getenv("CI") == "true" {
            // CI 环境使用模拟器
            device, err = uixt.NewAndroidDevice(
                option.WithSerialNumber("emulator-5554"),
                option.WithReset(true),
            )
        } else {
            // 本地环境使用真机
            device, err = uixt.NewAndroidDevice(
                option.WithSerialNumber(os.Getenv("ANDROID_SERIAL")),
                option.WithAdbLogOn(true),
            )
        }
    }

    if err != nil {
        return nil, err
    }

    driver, err := uixt.NewUIA2Driver(device)
    if err != nil {
        return nil, err
    }

    // 根据环境选择 AI 配置
    var aiOptions []option.AIServiceOption
    if os.Getenv("ENABLE_AI") == "true" {
        configs := option.RecommendedConfigurations()
        aiOptions = append(aiOptions, option.WithLLMConfig(configs["mixed_optimal"]))
        aiOptions = append(aiOptions, option.WithCVService(option.CVServiceTypeVEDEM))
    }

    return uixt.NewXTDriver(driver, aiOptions...)
}
```

## 配置验证

### 配置检查

```go
// 验证配置完整性
func validateConfiguration() error {
    // 检查必需的环境变量
    requiredEnvs := []string{
        "OPENAI_API_KEY",
        "VEDEM_IMAGE_AK",
        "VEDEM_IMAGE_SK",
    }

    for _, env := range requiredEnvs {
        if os.Getenv(env) == "" {
            return fmt.Errorf("required environment variable %s not set", env)
        }
    }

    // 检查设备连接
    devices, err := uixt.DiscoverAndroidDevices()
    if err != nil {
        return fmt.Errorf("failed to discover Android devices: %w", err)
    }

    if len(devices) == 0 {
        return fmt.Errorf("no Android devices found")
    }

    return nil
}
```

### 配置诊断

```go
// 配置诊断工具
func diagnoseConfiguration() {
    fmt.Println("=== Configuration Diagnosis ===")

    // 检查环境变量
    fmt.Println("\nEnvironment Variables:")
    envVars := []string{
        "OPENAI_BASE_URL", "OPENAI_API_KEY",
        "VEDEM_IMAGE_URL", "VEDEM_IMAGE_AK", "VEDEM_IMAGE_SK",
    }

    for _, env := range envVars {
        value := os.Getenv(env)
        if value != "" {
            fmt.Printf("  %s: %s\n", env, maskSensitive(value))
        } else {
            fmt.Printf("  %s: NOT SET\n", env)
        }
    }

    // 检查设备连接
    fmt.Println("\nDevice Status:")
    androidDevices, _ := uixt.DiscoverAndroidDevices()
    fmt.Printf("  Android devices: %d\n", len(androidDevices))

    iosDevices, _ := uixt.DiscoverIOSDevices()
    fmt.Printf("  iOS devices: %d\n", len(iosDevices))
}

func maskSensitive(value string) string {
    if len(value) <= 8 {
        return "***"
    }
    return value[:4] + "***" + value[len(value)-4:]
}
```

## 最佳实践

### 1. 配置分层

```go
// 分层配置管理
type Config struct {
    Device    DeviceConfig    `yaml:"device"`
    AI        AIConfig        `yaml:"ai"`
    Operation OperationConfig `yaml:"operation"`
}

type DeviceConfig struct {
    Platform string `yaml:"platform"`
    Serial   string `yaml:"serial"`
    Timeout  int    `yaml:"timeout"`
}

type AIConfig struct {
    LLMModel string `yaml:"llm_model"`
    CVService string `yaml:"cv_service"`
}

type OperationConfig struct {
    DefaultTimeout int `yaml:"default_timeout"`
    RetryCount     int `yaml:"retry_count"`
}
```

### 2. 配置验证

```go
// 配置验证
func (c *Config) Validate() error {
    if c.Device.Platform == "" {
        return fmt.Errorf("device platform is required")
    }

    if c.Device.Serial == "" {
        return fmt.Errorf("device serial is required")
    }

    if c.Operation.DefaultTimeout <= 0 {
        c.Operation.DefaultTimeout = 30 // 设置默认值
    }

    return nil
}
```

### 3. 配置热重载

```go
// 配置热重载
func watchConfigFile(configPath string, callback func(*Config)) {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        log.Fatal(err)
    }
    defer watcher.Close()

    err = watcher.Add(configPath)
    if err != nil {
        log.Fatal(err)
    }

    for {
        select {
        case event := <-watcher.Events:
            if event.Op&fsnotify.Write == fsnotify.Write {
                config, err := loadConfig(configPath)
                if err == nil {
                    callback(config)
                }
            }
        case err := <-watcher.Errors:
            log.Println("error:", err)
        }
    }
}
```

## 参考资料

- [环境变量最佳实践](https://12factor.net/config)
- [YAML 配置文件格式](https://yaml.org/)
- [Go 配置管理库 Viper](https://github.com/spf13/viper)