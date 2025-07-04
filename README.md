# HttpRunner

[![Github Actions](https://github.com/httprunner/httprunner/actions/workflows/unittest.yml/badge.svg)](https://github.com/httprunner/httprunner/actions)
[![codecov](https://codecov.io/gh/httprunner/httprunner/branch/master/graph/badge.svg)](https://codecov.io/gh/httprunner/httprunner)
[![Go Reference](https://pkg.go.dev/badge/github.com/httprunner/httprunner.svg)](https://pkg.go.dev/github.com/httprunner/httprunner)
[![downloads](https://pepy.tech/badge/httprunner)](https://pepy.tech/project/httprunner)
[![TesterHome](https://img.shields.io/badge/TTF-TesterHome-2955C5.svg)](https://testerhome.com/github_statistics)

> ⚠️ HttpRunner v5 仅包含 Golang 版本，Python 版本的代码已迁移至 [httprunner/httprunner.py](https://github.com/httprunner/httprunner.py)

`HttpRunner`（简称 hrp） 是一个开源测试框架，诞生于 2017 年，最开始应用于 API 接口、性能测试，后面逐步进化为了一款通用、可拓展的测试框架。

在 2022 年，HttpRunner 开始新增支持 UI 自动化测试，当前已经支持了 Android/iOS/Harmony/Browser 多种系统平台，并在 v5 版本融入了大模型技术，成长成为了一款通用的智能自动化测试框架。

HttpRunner 相比其它 UI 自动化框架，主要特点包括：

- 采用纯视觉驱动方案（OCR/CV/VLM），追求通用性和尽可能低的性能损耗
- 多端统一 API，降低学习和横向拓展的成本
- 拥抱开源生态，充分复用开源组件
- Golang 技术栈，二进制分发部署

> [HttpRunner v5 用户指南（更新中）](https://debugtalk.feishu.cn/wiki/RqGuw17bsizGTik9WuNcGQyhnaf)
> [HttpRunner DeepWiki](https://deepwiki.com/httprunner/httprunner)

## 使用说明

HttpRunner v5 安装完成后，你将获得一个 `hrp` 命令行工具，执行 `hrp -h` 即可查看到参数帮助说明。

```text
$ hrp -h

██╗  ██╗████████╗████████╗██████╗ ██████╗ ██╗   ██╗███╗   ██╗███╗   ██╗███████╗██████╗
██║  ██║╚══██╔══╝╚══██╔══╝██╔══██╗██╔══██╗██║   ██║████╗  ██║████╗  ██║██╔════╝██╔══██╗
███████║   ██║      ██║   ██████╔╝██████╔╝██║   ██║██╔██╗ ██║██╔██╗ ██║█████╗  ██████╔╝
██╔══██║   ██║      ██║   ██╔═══╝ ██╔══██╗██║   ██║██║╚██╗██║██║╚██╗██║██╔══╝  ██╔══██╗
██║  ██║   ██║      ██║   ██║     ██║  ██║╚██████╔╝██║ ╚████║██║ ╚████║███████╗██║  ██║
╚═╝  ╚═╝   ╚═╝      ╚═╝   ╚═╝     ╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═══╝╚═╝  ╚═══╝╚══════╝╚═╝  ╚═╝

HttpRunner: Enjoy your All-in-One Testing Solution ✨ 🚀 ✨

💡 Simple Yet Powerful
   - Natural language driven test scenarios powered by LLM
   - User-friendly SDK API with IDE auto-completion
   - Intuitive GoTest/YAML/JSON/Text testcase format

📌 Comprehensive Testing Capabilities
   - UI Automation: Android/iOS/Harmony/Browser
   - API Testing: HTTP(S)/HTTP2/WebSocket/RPC
   - Load Testing: run API testcase concurrently with boomer

🧩 High Scalability
   - Plugin system for custom functions
   - Distributed testing support
   - Cross-platform: macOS/Linux/Windows

🛠 Easy Integration
   - CI/CD friendly with JSON logs and HTML reports
   - Rich ecosystem tools

Learn more:
Website: https://httprunner.com
GitHub: https://github.com/httprunner/httprunner

Copyright © 2017-present debugtalk. Apache-2.0 License.

Usage:
  hrp [command]

Available Commands:
  adb          simple utils for android device management
  build        Build plugin for testing
  completion   Generate the autocompletion script for the specified shell
  convert      Convert multiple source format to HttpRunner JSON/YAML/gotest/pytest cases
  help         Help about any command
  ios          simple utils for ios device management
  mcp-server   Start MCP server for UI automation
  mcphost      Start a chat session to interact with MCP tools
  pytest       Run API test with pytest
  report       Generate HTML report from test results
  run          Run API test with go engine
  server       Start hrp server
  startproject Create a scaffold project
  wiki         visit https://httprunner.com

Flags:
  -h, --help               help for hrp
      --log-json           set log to json format (default colorized console)
  -l, --log-level string   set log level (default "INFO")
      --venv string        specify python3 venv path
  -v, --version            version for hrp

Use "hrp [command] --help" for more information about a command.
```

## 用户案例

<a href="https://httprunner.com/docs/cases/dji-ibg"><img src="https://httprunner.com/image/logo/dji.jpeg" title="大疆 - 基于 HttpRunner 构建完整的自动化测试体系" width="60"></a>
<a href="https://httprunner.com/docs/cases/youmi"><img src="https://httprunner.com/image/logo/youmi.png" title="有米科技 - 基于 HttpRunner 建设自动化测试平台" width="60"></a>
<a href="https://httprunner.com/docs/cases/umcare"><img src="https://httprunner.com/image/logo/umcare.png" title="通用环球医疗 - 使用 HttpRunner 实践接口自动化测试" width="100"></a>
<a href="https://httprunner.com/docs/cases/mihoyo"><img src="https://httprunner.com/image/logo/miHoYo.png" title="米哈游 - 基于 HttpRunner 搭建接口自动化测试体系" width="100"></a>

## Subscribe

关注 HttpRunner 的微信公众号，第一时间获得最新资讯。

<img src="https://httprunner.com/image/qrcode.png" alt="HttpRunner" width="400">

[HttpRunner]: https://github.com/httprunner/httprunner
[boomer]: https://github.com/myzhan/boomer
[locust]: https://github.com/locustio/locust
[jmespath]: https://jmespath.org/
[allure]: https://docs.qameta.io/allure/
[HAR]: https://en.wikipedia.org/wiki/HAR_(file_format)
[hashicorp plugin]: https://github.com/hashicorp/go-plugin
[go plugin]: https://pkg.go.dev/plugin
[版本发布日志]: docs/CHANGELOG.md
[pushgateway]: https://github.com/prometheus/pushgateway
[survey]: https://wj.qq.com/s2/9699514/0d19/
[user-survey-report]: https://httprunner.com/blog/user-survey-report/
[English]: README.en.md
[pytest]: https://docs.pytest.org/
