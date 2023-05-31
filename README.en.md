# HttpRunner

[![Github Actions](https://github.com/httprunner/httprunner/actions/workflows/unittest.yml/badge.svg)](https://github.com/httprunner/httprunner/actions)
[![codecov](https://codecov.io/gh/httprunner/httprunner/branch/master/graph/badge.svg)](https://codecov.io/gh/httprunner/httprunner)
[![Go Reference](https://pkg.go.dev/badge/github.com/httprunner/httprunner.svg)](https://pkg.go.dev/github.com/httprunner/httprunner)
[![downloads](https://pepy.tech/badge/httprunner)](https://pepy.tech/project/httprunner)
[![pypi version](https://img.shields.io/pypi/v/httprunner.svg)](https://pypi.python.org/pypi/httprunner)
[![pyversions](https://img.shields.io/pypi/pyversions/httprunner.svg)](https://pypi.python.org/pypi/httprunner)
[![TesterHome](https://img.shields.io/badge/TTF-TesterHome-2955C5.svg)](https://testerhome.com/github_statistics)

`HttpRunner` is an open source API testing tool that supports HTTP(S)/HTTP2/WebSocket/RPC network protocols, covering API testing, performance testing and digital experience monitoring (DEM) test types. Simple and easy to use, powerful, with rich plug-in mechanism and high scalability.

> HttpRunner [用户调研问卷][survey] 持续收集中，我们将基于用户反馈动态调整产品特性和需求优先级。

![flow chart](https://httprunner.com/image/hrp-flow.jpg)

[CHANGELOG] | [中文]

## Key Features

### API Testing

- [x] Full support for HTTP(S)/1.1 and HTTP/2 requests.
- [ ] Support more protocols, WebSocket, TCP, RPC etc.
- [x] Testcases can be described in multiple formats, `YAML`/`JSON`/`Golang`, and they are interchangeable.
- [x] Use Charles/Fiddler/Chrome/etc to record HTTP requests and generate testcases from exported [`HAR`][HAR].
- [x] Supports `variables`/`extract`/`validate`/`hooks` mechanisms to create extremely complex test scenarios.
- [x] Data driven with `parameterize` mechanism, supporting sequential/random/unique strategies to select data.
- [ ] Built-in 100+ commonly used functions for ease, including md5sum, max/min, sleep, gen_random_string etc.
- [x] Create and call custom functions with `plugin` mechanism, support [hashicorp plugin] and [go plugin].
- [x] Generate html reports with rich test results.
- [x] Using it as a `CLI tool` or a `library` are both supported.

### Load Testing

Base on the API testing testcases, you can run professional load testing without extra work.

- [x] Inherit all powerful features of [`locust`][locust] and [`boomer`][boomer].
- [x] Report performance metrics to [prometheus pushgateway][pushgateway].
- [x] Use `transaction` to define a set of end-user actions that represent the real user activities.
- [x] Use `rendezvous` points to force Vusers to perform tasks concurrently during test execution.
- [x] Load testing with specified concurrent users or constant RPS, also supports spawn rate.
- [ ] Support mixed-scenario testing with custom weight.
- [ ] Simulate browser's HTTP parallel connections.
- [ ] IP spoofing.
- [ ] Run in distributed mode to generate unlimited RPS.

### Digital Experience Monitoring (DEM)

You can also monitor online services for digital experience assessments.

- [ ] HTTP(S) latency statistics including DNSLookup, TCP connections, SSL handshakes, content transfers, etc.
- [ ] `ping` indicators including latency, throughput and packets loss.
- [ ] traceroute
- [ ] DNS monitoring

## Install

You can install HttpRunner via one curl command.

```bash
$ bash -c "$(curl -ksSL https://httprunner.com/script/install.sh)"
```

Then you will get a `hrp` CLI tool.

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
  adb          simple utils for android device management
  boom         run load test with boomer
  build        build plugin for testing
  completion   Generate the autocompletion script for the specified shell
  convert      convert multiple source format to HttpRunner JSON/YAML/gotest/pytest cases
  help         Help about any command
  ios          simple utils for ios device management
  pytest       run API test with pytest
  run          run API test with go engine
  startproject create a scaffold project
  wiki         visit https://httprunner.com

Flags:
  -h, --help               help for hrp
      --log-json           set log to json format
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

## Sponsor

[<img src="https://testing-studio.com/img/icon.png" alt="霍格沃兹测试开发学社" width="500">](https://qrcode.testing-studio.com/f?from=HttpRunner&url=https://testing-studio.com/)

> 霍格沃兹测试开发学社是中国软件测试开发高端教育品牌，产品由国内顶尖软件测试开发技术专家携手打造，为企业与个人提供专业的技能培训与咨询、测试工具与测试平台、测试外包与测试众包服务。领域涵盖 App/Web 自动化测试、接口自动化测试、性能测试、安全测试、持续交付/DevOps、测试左移、测试右移、精准测试、测试平台开发、测试管理等方向。-> [**联系我们**](http://qrcode.testing-studio.com/f?from=HttpRunner&url=https://ceshiren.com/t/topic/23745)

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
