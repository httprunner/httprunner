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

![flow chart](docs/assets/hrp-flow.jpg)

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
# backup
$ bash -c "$(curl -ksSL https://httprunner.oss-cn-beijing.aliyuncs.com/install.sh)"
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
Copyright 2021 debugtalk

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

## Subscribe

关注 HttpRunner 的微信公众号，第一时间获得最新资讯。

<img src="docs/assets/qrcode.jpg" alt="HttpRunner" width="200">

如果你期望加入 HttpRunner 核心用户群，请填写[用户调研问卷][survey]并留下你的联系方式，作者将拉你进群。

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
