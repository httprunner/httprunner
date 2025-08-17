# HttpRunner UIXT 模块

## 🚀 概述

HttpRunner UIXT（UI eXtended Testing）是 HttpRunner v5.0 的核心 UI 自动化测试模块，提供统一的 API 接口支持多种平台的 UI 自动化测试，并集成了先进的 AI 能力，实现真正的智能化 UI 自动化测试。

### 核心特性

- **🎯 跨平台支持**: Android、iOS、HarmonyOS、Web 浏览器统一接口
- **🤖 AI 智能化**: 集成大语言模型和计算机视觉，支持自然语言驱动的 UI 操作
- **🔧 MCP 协议**: 基于 Model Context Protocol 的标准化工具接口
- **📱 多设备管理**: 支持真机、模拟器、浏览器的统一管理
- **🎨 丰富操作**: 触摸、滑动、输入、应用管理等完整操作集
- **📊 智能识别**: OCR 文本识别、UI 元素检测、弹窗识别

## 🏗️ 核心架构

### 整体架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                        HttpRunner UIXT v5                      │
├─────────────────────────────────────────────────────────────────┤
│                      XTDriver (扩展驱动)                        │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │   IDriver       │  │   AI Services   │  │   MCP Server    │  │
│  │   (核心驱动)     │  │   (AI 能力)     │  │   (工具协议)     │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘  │
├─────────────────────────────────────────────────────────────────┤
│                        设备驱动层                                │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │  Android Driver │  │   iOS Driver    │  │  Browser Driver │  │
│  │  (ADB/UIA2)     │  │     (WDA)       │  │   (WebDriver)   │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘  │
│  ┌─────────────────┐                                            │
│  │ Harmony Driver  │                                            │
│  │     (HDC)       │                                            │
│  └─────────────────┘                                            │
├─────────────────────────────────────────────────────────────────┤
│                        设备层                                   │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │  Android Device │  │   iOS Device    │  │  Browser Device │  │
│  │   (真机/模拟器)   │  │   (真机/模拟器)   │  │    (浏览器)      │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘  │
│  ┌─────────────────┐                                            │
│  │ Harmony Device  │                                            │
│  │   (真机/模拟器)   │                                            │
│  └─────────────────┘                                            │
└─────────────────────────────────────────────────────────────────┘
```

### v5 版本新特性

#### 🤖 AI 驱动的智能化
- **LLM 集成**: 支持 OpenAI GPT-4O、豆包模型、DeepSeek 等多种大语言模型
- **智能规划**: AI 自动分析屏幕截图并生成操作序列
- **智能断言**: 基于自然语言描述进行断言验证
- **智能查询**: 从界面中提取结构化信息

#### 🔌 MCP 协议支持
- **标准化接口**: 基于 Model Context Protocol 的工具调用
- **可扩展性**: 支持动态注册新的 AI 工具
- **跨平台兼容**: 统一的 AI 模型交互协议

#### 📱 增强的平台支持
- **Harmony 平台**: 新增完整的 HarmonyOS 支持
- **浏览器增强**: 改进的 WebDriver 集成
- **驱动缓存**: 智能会话管理和资源复用

### 核心设计思路

#### 1. 分层架构设计
- **设备层**: 抽象不同平台的设备管理
- **驱动层**: 统一不同平台的操作接口
- **扩展层**: 提供 AI 和高级功能
- **协议层**: 标准化的工具调用接口

#### 2. 接口统一化
所有平台都实现相同的 `IDriver` 接口，确保操作的一致性：

```go
type IDriver interface {
    // 设备信息和状态
    Status() (types.DeviceStatus, error)
    DeviceInfo() (types.DeviceInfo, error)
    WindowSize() (types.Size, error)
    ScreenShot(opts ...option.ActionOption) (*bytes.Buffer, error)

    // 基础操作
    TapXY(x, y float64, opts ...option.ActionOption) error
    Swipe(fromX, fromY, toX, toY float64, opts ...option.ActionOption) error
    Input(text string, opts ...option.ActionOption) error

    // 应用管理
    AppLaunch(packageName string) error
    AppTerminate(packageName string) (bool, error)

    // ... 更多操作
}
```

#### 3. AI 能力集成
通过 `XTDriver` 扩展驱动集成 AI 服务：

```go
type XTDriver struct {
    IDriver                    // 基础驱动能力
    CVService  ai.ICVService   // 计算机视觉服务
    LLMService ai.ILLMService  // 大语言模型服务
}
```

#### 4. MCP 工具化
将所有操作封装为 MCP 工具，支持 AI 模型直接调用：

```go
type ActionTool interface {
    Name() option.ActionName
    Description() string
    Options() []mcp.ToolOption
    Implement() server.ToolHandlerFunc
}
```

## 📖 支持平台

### Android 平台
- **驱动方式**: ADB + UiAutomator2
- **支持设备**: 真机、模拟器
- **最低版本**: Android 5.0+
- **特色功能**: 应用管理、文件传输、日志捕获

### iOS 平台
- **驱动方式**: WebDriverAgent (WDA)
- **支持设备**: 真机、模拟器
- **最低版本**: iOS 10.0+
- **特色功能**: 应用管理、图片传输、性能监控

### HarmonyOS 平台
- **驱动方式**: HDC (HarmonyOS Device Connector)
- **支持设备**: 真机、模拟器
- **最低版本**: HarmonyOS 2.0+
- **特色功能**: 原生鸿蒙应用支持

### Web 浏览器
- **驱动方式**: WebDriver 协议
- **支持浏览器**: Chrome、Firefox、Safari、Edge
- **特色功能**: 多标签页管理、JavaScript 执行

## 🚀 快速开始

### 1. 环境准备

#### Android 环境
```bash
# 安装 Android SDK
export ANDROID_HOME=/path/to/android-sdk
export PATH=$PATH:$ANDROID_HOME/platform-tools

# 启用 USB 调试
adb devices
```

#### iOS 环境
```bash
# 安装 Xcode 和 WebDriverAgent
# 配置开发者证书
# 启动 WDA 服务
```

#### AI 服务配置
```bash
# 配置大语言模型服务
export OPENAI_BASE_URL=https://api.openai.com/v1
export OPENAI_API_KEY=your_api_key

# 配置计算机视觉服务
export VEDEM_IMAGE_URL=https://visual.volcengineapi.com
export VEDEM_IMAGE_AK=your_access_key
export VEDEM_IMAGE_SK=your_secret_key
```

### 2. 基础使用

#### 创建设备和驱动
```go
package main

import (
    "github.com/httprunner/httprunner/v5/uixt"
    "github.com/httprunner/httprunner/v5/uixt/option"
)

func main() {
    // 创建 Android 设备
    device, err := uixt.NewAndroidDevice(
        option.WithSerialNumber("your_device_serial"),
    )
    if err != nil {
        panic(err)
    }

    // 创建基础驱动
    driver, err := uixt.NewUIA2Driver(device)
    if err != nil {
        panic(err)
    }

    // 创建扩展驱动（集成 AI 能力）
    xtDriver, err := uixt.NewXTDriver(driver,
        option.WithCVService(option.CVServiceTypeVEDEM),
        option.WithLLMService(option.OPENAI_GPT_4O),
    )
    if err != nil {
        panic(err)
    }

    // 初始化会话
    err = xtDriver.Setup()
    if err != nil {
        panic(err)
    }
    defer xtDriver.TearDown()
}
```

#### 基础操作示例
```go
// 获取屏幕截图
screenshot, err := xtDriver.ScreenShot()

// 点击操作
err = xtDriver.TapXY(0.5, 0.5) // 相对坐标 (50%, 50%)

// 滑动操作
err = xtDriver.Swipe(0.5, 0.8, 0.5, 0.2) // 从下往上滑动

// 输入文本
err = xtDriver.Input("Hello World")

// 启动应用
err = xtDriver.AppLaunch("com.example.app")
```

#### AI 智能操作
```go
import "context"

// 使用自然语言执行操作
result, err := xtDriver.LLMService.Plan(context.Background(), &ai.PlanningOptions{
    UserInstruction: "点击登录按钮",
    Message: message,
    Size: screenSize,
})

// 智能断言
assertResult, err := xtDriver.LLMService.Assert(context.Background(), &ai.AssertOptions{
    Assertion: "登录按钮应该可见",
    Screenshot: screenshot,
    Size: screenSize,
})

// 智能查询
queryResult, err := xtDriver.LLMService.Query(context.Background(), &ai.QueryOptions{
    Query: "提取页面中的所有文本内容",
    Screenshot: screenshot,
    Size: screenSize,
})
```

### 3. 高级配置

#### 混合模型配置
```go
// 为不同组件配置不同的最优模型
config := option.NewLLMServiceConfig(option.DOUBAO_1_5_THINKING_VISION_PRO_250428).
    WithPlannerModel(option.DOUBAO_1_5_UI_TARS_250328).  // UI理解用UI-TARS
    WithAsserterModel(option.OPENAI_GPT_4O).             // 推理用GPT-4O
    WithQuerierModel(option.DEEPSEEK_R1_250528)          // 查询用DeepSeek

xtDriver, err := uixt.NewXTDriver(driver,
    option.WithLLMConfig(config),
)
```

#### 使用推荐配置
```go
configs := option.RecommendedConfigurations()
xtDriver, err := uixt.NewXTDriver(driver,
    option.WithLLMConfig(configs["mixed_optimal"]),
)
```

## 📚 详细文档

### 核心文档

- **[设备管理](devices.md)** - 设备发现、连接、配置和管理
- **[驱动接口](drivers.md)** - 各平台驱动的功能和使用方法
- **[操作指南](operations.md)** - 详细的 UI 操作使用指南
- **[配置选项](options.md)** - 完整的配置参数说明

### AI 和工具

- **[AI 模块](ai.md)** - LLM 和 CV 服务的集成使用、智能规划、断言、查询
- **[MCP 工具](mcp-tools.md)** - MCP 协议和工具系统详解

### 快速导航

| 文档 | 内容概述 |
|------|----------|
| [设备管理](devices.md) | 设备发现、连接、多设备管理、故障排除、平台特有功能 |
| [驱动接口](drivers.md) | IDriver 接口、平台驱动、XTDriver 扩展、选择器类型 |
| [操作指南](operations.md) | 点击、滑动、输入、应用管理、屏幕操作 |
| [AI 模块](ai.md) | 智能规划、智能断言、智能查询、CV 识别、多模型配置 |
| [MCP 工具](mcp-tools.md) | 工具分类、实现方式、扩展开发 |
| [配置选项](options.md) | 设备配置、AI 配置、环境变量、最佳实践 |

## 🔧 依赖项目

### 核心依赖
- [electricbubble/gwda](https://github.com/electricbubble/gwda) - iOS WebDriverAgent 客户端
- [electricbubble/guia2](https://github.com/electricbubble/guia2) - Android UiAutomator2 客户端
- [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) - MCP 协议 Go 实现

### AI 服务依赖
- [cloudwego/eino](https://github.com/cloudwego/eino) - 统一的 LLM 接口
- 火山引擎 VEDEM - 计算机视觉服务
- OpenAI GPT-4O - 大语言模型服务
- 豆包系列模型 - 专业 UI 自动化模型

## 🤝 贡献指南

我们欢迎社区贡献！请查看以下资源：

- [贡献指南](CONTRIBUTING.md) - 如何参与项目贡献
- [开发环境搭建](development.md) - 开发环境配置
- [代码规范](coding-standards.md) - 代码风格和规范
- [测试指南](testing.md) - 测试编写和执行

## 📄 许可证

本项目采用 Apache 2.0 许可证，详情请查看 [LICENSE](LICENSE) 文件。

## 🙏 致谢

感谢以下开源项目的贡献：
- [appium-uiautomator2-server](https://github.com/appium/appium-uiautomator2-server) - Android 自动化基础
- [appium/WebDriverAgent](https://github.com/appium/WebDriverAgent) - iOS 自动化基础
- [danielpaulus/go-ios](https://github.com/danielpaulus/go-ios) - iOS 客户端库

---

**HttpRunner UIXT** - 让 UI 自动化测试更智能、更简单！
