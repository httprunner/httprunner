# HttpRunner

[![Github Actions](https://github.com/httprunner/httprunner/actions/workflows/unittest.yml/badge.svg)](https://github.com/httprunner/httprunner/actions)
[![codecov](https://codecov.io/gh/httprunner/httprunner/branch/master/graph/badge.svg)](https://codecov.io/gh/httprunner/httprunner)
[![Go Reference](https://pkg.go.dev/badge/github.com/httprunner/httprunner.svg)](https://pkg.go.dev/github.com/httprunner/httprunner)
[![downloads](https://pepy.tech/badge/httprunner)](https://pepy.tech/project/httprunner)
[![pypi version](https://img.shields.io/pypi/v/httprunner.svg)](https://pypi.python.org/pypi/httprunner)
[![pyversions](https://img.shields.io/pypi/pyversions/httprunner.svg)](https://pypi.python.org/pypi/httprunner)
[![TesterHome](https://img.shields.io/badge/TTF-TesterHome-2955C5.svg)](https://testerhome.com/github_statistics)

`HttpRunner` 是一个开源的 API 测试工具，支持 HTTP(S)/HTTP2/WebSocket/RPC 等网络协议，涵盖接口测试、性能测试、数字体验监测等测试类型。简单易用，功能强大，具有丰富的插件化机制和高度的可扩展能力。

> HttpRunner [用户调研问卷][survey] 持续收集中，我们将基于用户反馈动态调整产品特性和需求优先级。

![flow chart](https://httprunner.com/image/hrp-flow.jpg)

[版本发布日志] | [English]

## 设计理念

相比于其它 API 测试工具，HttpRunner 最大的不同在于设计理念。

- 约定大于配置：测试用例是标准结构化的，格式统一，方便协作和维护
- 标准开放：基于开放的标准，支持与 [HAR]/Postman/Swagger/Curl/JMeter 等工具对接，轻松实现用例生成和转换
- 一次投入多维复用：一套脚本可同时支持接口自动化测试、性能测试、数字体验监测等多种 API 测试需求
- 融入最佳工程实践：不仅仅是一款测试工具，在功能中融入最佳工程实践，实现面向网络协议的一站式测试解决方案

## 核心特性

- 网络协议：完整支持 HTTP(S)/HTTP2/WebSocket，可扩展支持 TCP/UDP/RPC 等更多协议
- 多格式可选：测试用例支持 YAML/JSON/go test/pytest 格式，并且支持格式互相转换
- 双执行引擎：同时支持 golang/python 两个执行引擎，兼具 go 的高性能和 [pytest] 的丰富生态
- 录制 & 生成：可使用 [HAR]/Postman/Swagger/curl 等生成测试用例；基于链式调用的方法提示也可快速编写测试用例
- 复杂场景：基于 variables/extract/validate/hooks 机制可以方便地创建任意复杂的测试场景
- 插件化机制：内置丰富的函数库，同时可以基于主流编程语言（go/python/java）编写自定义函数轻松实现更多能力
- 性能测试：无需额外工作即可实现压力测试；单机可轻松支撑 `1w+` VUM，结合分布式负载能力可实现海量发压
- 网络性能采集：在场景化接口测试的基础上，可额外采集网络链路性能指标（DNS 解析、TCP 连接、SSL 握手、网络传输等）
- 一键部署：采用二进制命令行工具分发，无需环境依赖，一条命令即可在 macOS/Linux/Windows 快速完成安装部署

## 用户声音

基于 252 份调研问卷的统计结果，HttpRunner 用户的整体满意度评分 `4.3/5`，最喜欢的特性包括：

- 简单易用：测试用例支持 YAML/JSON 标准化格式，可通过录制的方式快速生成用例，上手简单，使用方便
- 功能强大：支持灵活的自定义函数和 hook 机制，参数变量、数据驱动、结果断言等机制一应俱全，轻松适应各种复杂场景
- 设计理念：测试用例组织支持分层设计，格式统一，易于实现测试用例的维护和复用

更多内容详见 [HttpRunner 首轮用户调研报告（2022.02）][user-survey-report]

## 一键部署

HttpRunner 二进制命令行工具已上传至阿里云 OSS，在系统终端中执行如下命令可完成安装部署。

```bash
$ bash -c "$(curl -ksSL https://httprunner.com/script/install.sh)"
```

安装成功后，你将获得一个 `hrp` 命令行工具，执行 `hrp -h` 即可查看到参数帮助说明。

```text
$ hrp -h

██╗  ██╗████████╗████████╗██████╗ ██████╗ ██╗   ██╗███╗   ██╗███╗   ██╗███████╗██████╗
██║  ██║╚══██╔══╝╚══██╔══╝██╔══██╗██╔══██╗██║   ██║████╗  ██║████╗  ██║██╔════╝██╔══██╗
███████║   ██║      ██║   ██████╔╝██████╔╝██║   ██║██╔██╗ ██║██╔██╗ ██║█████╗  ██████╔╝
██╔══██║   ██║      ██║   ██╔═══╝ ██╔══██╗██║   ██║██║╚██╗██║██║╚██╗██║██╔══╝  ██╔══██╗
██║  ██║   ██║      ██║   ██║     ██║  ██║╚██████╔╝██║ ╚████║██║ ╚████║███████╗██║  ██║
╚═╝  ╚═╝   ╚═╝      ╚═╝   ╚═╝     ╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═══╝╚═╝  ╚═══╝╚══════╝╚═╝  ╚═╝

HttpRunner is an open source API testing tool that supports HTTP(S)/HTTP2/WebSocket/RPC
network protocols, covering API testing, performance testing and digital experience
monitoring (DEM) test types. Enjoy! ✨ 🚀 ✨

License: Apache-2.0
Website: https://httprunner.com
Github: https://github.com/httprunner/httprunner
Copyright 2017 debugtalk

Usage:
  hrp [command]

Available Commands:
  boom         run load test with boomer
  completion   generate the autocompletion script for the specified shell
  har2case     convert HAR to json/yaml testcase files
  help         Help about any command
  pytest       run API test with pytest
  run          run API test with go engine
  startproject create a scaffold project

Flags:
  -h, --help               help for hrp
      --log-json           set log to json format
  -l, --log-level string   set log level (default "INFO")
  -v, --version            version for hrp

Use "hrp [command] --help" for more information about a command.
```

## 用户案例

<a href="https://httprunner.com/docs/cases/dji-ibg"><img src="https://httprunner.com/image/logo/dji.jpeg" title="大疆 - 基于 HttpRunner 构建完整的自动化测试体系" width="60"></a>
<a href="https://httprunner.com/docs/cases/bytedance-feishu"><img src="https://httprunner.com/image/logo/feishu.jpeg" title="飞书 - 使用 HttpRunner 替换已有测试平台的执行引擎" width="60"></a>

## 赞助商

### 金牌赞助商

[<img src="https://httprunner.com/image/hogwarts.jpeg" alt="霍格沃兹测试开发学社" width="400">](https://ceshiren.com/)

> [霍格沃兹测试开发学社](http://qrcode.testing-studio.com/f?from=httprunner&url=https://ceshiren.com)是业界领先的测试开发技术高端教育品牌，隶属于[测吧（北京）科技有限公司](http://qrcode.testing-studio.com/f?from=httprunner&url=https://www.testing-studio.com) 。学院课程由一线大厂测试经理与资深测试开发专家参与研发，实战驱动。课程涵盖 web/app 自动化测试、接口测试、性能测试、安全测试、持续集成/持续交付/DevOps，测试左移&右移、精准测试、测试平台开发、测试管理等内容，帮助测试工程师实现测试开发技术转型。通过优秀的学社制度（奖学金、内推返学费、行业竞赛等多种方式）来实现学员、学社及用人企业的三方共赢。

> [进入测试开发技术能力测评!](http://qrcode.testing-studio.com/f?from=httprunner&url=https://ceshiren.com/t/topic/14940)

### 开源服务赞助商

[<img src="https://httprunner.com/image/sentry-logo-black.svg" alt="Sentry" width="150">](https://sentry.io/_/open-source/)

HttpRunner is in Sentry Sponsored plan.

## Subscribe

关注 HttpRunner 的微信公众号，第一时间获得最新资讯。

<img src="https://httprunner.com/image/qrcode.png" alt="HttpRunner" width="400">

如果你期望加入 HttpRunner 用户群，请看这里：[HttpRunner v4 用户交流群，它来啦！](https://httprunner.com/blog/join-chat-group)

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
