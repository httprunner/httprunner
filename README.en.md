# HttpRunner

[![Github Actions](https://github.com/httprunner/httprunner/actions/workflows/unittest.yml/badge.svg)](https://github.com/httprunner/httprunner/actions)
[![codecov](https://codecov.io/gh/httprunner/httprunner/branch/master/graph/badge.svg)](https://codecov.io/gh/httprunner/httprunner)
[![Go Reference](https://pkg.go.dev/badge/github.com/httprunner/httprunner.svg)](https://pkg.go.dev/github.com/httprunner/httprunner)
[![downloads](https://pepy.tech/badge/httprunner)](https://pepy.tech/project/httprunner)
[![TesterHome](https://img.shields.io/badge/TTF-TesterHome-2955C5.svg)](https://testerhome.com/github_statistics)

> ⚠️ HttpRunner v5 only includes the Golang version, and the Python version of the code has been migrated to [httprunner/httprunner.py](https://github.com/httprunner/httprunner.py)

`HttpRunner` (also known as hrp) is an open-source testing framework that was born in 2017. Initially, it was used for API interface and performance testing, and later evolved into a versatile and extensible testing framework.

In 2022, HttpRunner began to support UI automation testing, currently supporting multiple system platforms such as Android/iOS/Harmony/Browser, and integrated large model technology in v5.

Compared to other UI automation frameworks, HttpRunner's main features include:

- Pure visual-driven solution (OCR/CV/LLM), pursuing universality and minimal performance loss
- Unified API across multiple platforms, reducing learning and horizontal expansion costs
- Embracing the open-source ecosystem, fully reusing open-source components

> [HttpRunner v5 用户指南（更新中）](https://debugtalk.feishu.cn/wiki/RqGuw17bsizGTik9WuNcGQyhnaf)
> [HttpRunner DeepWiki](https://deepwiki.com/httprunner/httprunner)

## Usage
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
  build        build plugin for testing
  completion   Generate the autocompletion script for the specified shell
  convert      convert multiple source format to HttpRunner JSON/YAML/gotest/pytest cases
  help         Help about any command
  ios          simple utils for ios device management
  pytest       run API test with pytest
  run          run API test with go engine
  server       start hrp server
  startproject create a scaffold project
  wiki         visit https://httprunner.com

Flags:
  -h, --help               help for hrp
      --log-json           set log to json format (default colorized console)
  -l, --log-level string   set log level (default "INFO")
      --venv string        specify python3 venv path
  -v, --version            version for hrp

Use "hrp [command] --help" for more information about a command.
```

## User Cases

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
[HAR]: http://httparchive.org/
[hashicorp plugin]: https://github.com/hashicorp/go-plugin
[go plugin]: https://pkg.go.dev/plugin
[CHANGELOG]: docs/CHANGELOG.md
[pushgateway]: https://github.com/prometheus/pushgateway
[survey]: https://wj.qq.com/s2/9699514/0d19/
[中文]: README.md
