# HttpRunner

[![Github Actions](https://github.com/httprunner/httprunner/actions/workflows/unittest.yml/badge.svg)](https://github.com/httprunner/httprunner/actions)
[![codecov](https://codecov.io/gh/httprunner/httprunner/branch/master/graph/badge.svg)](https://codecov.io/gh/httprunner/httprunner)
[![Go Report Card](https://goreportcard.com/badge/github.com/httprunner/httprunner)](https://goreportcard.com/report/github.com/httprunner/httprunner)
[![Go Reference](https://pkg.go.dev/badge/github.com/httprunner/httprunner.svg)](https://pkg.go.dev/github.com/httprunner/httprunner)
[![downloads](https://pepy.tech/badge/httprunner)](https://pepy.tech/project/httprunner)
[![pypi version](https://img.shields.io/pypi/v/httprunner.svg)](https://pypi.python.org/pypi/httprunner)
[![pyversions](https://img.shields.io/pypi/pyversions/httprunner.svg)](https://pypi.python.org/pypi/httprunner)
[![TesterHome](https://img.shields.io/badge/TTF-TesterHome-2955C5.svg)](https://testerhome.com/github_statistics)

HttpRunner aims to be a one-stop solution for HTTP(S) testing, covering API testing, load testing and digital experience monitoring (DEM).

See [CHANGELOG].

> HttpRunner [用户调研问卷][survey] 持续收集中，我们将基于用户反馈动态调整产品特性和需求优先级。

## Key Features

![flow chart](docs/assets/flow.jpg)

### API Testing

- [x] Full support for HTTP(S)/1.1 requests.
- [ ] Support more protocols, HTTP/2, WebSocket, TCP, RPC etc.
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

## Sponsors

Thank you to all our sponsors! ✨🍰✨ ([become a sponsor](sponsors.md))

### 金牌赞助商（Gold Sponsor）

[<img src="docs/assets/hogwarts.jpeg" alt="霍格沃兹测试开发学社" width="400">](https://ceshiren.com/)

> [霍格沃兹测试开发学社](http://qrcode.testing-studio.com/f?from=httprunner&url=https://ceshiren.com)是业界领先的测试开发技术高端教育品牌，隶属于[测吧（北京）科技有限公司](http://qrcode.testing-studio.com/f?from=httprunner&url=https://www.testing-studio.com) 。学院课程由一线大厂测试经理与资深测试开发专家参与研发，实战驱动。课程涵盖 web/app 自动化测试、接口测试、性能测试、安全测试、持续集成/持续交付/DevOps，测试左移&右移、精准测试、测试平台开发、测试管理等内容，帮助测试工程师实现测试开发技术转型。通过优秀的学社制度（奖学金、内推返学费、行业竞赛等多种方式）来实现学员、学社及用人企业的三方共赢。

> [进入测试开发技术能力测评!](http://qrcode.testing-studio.com/f?from=httprunner&url=https://ceshiren.com/t/topic/14940)

### 开源服务赞助商（Open Source Sponsor）

[<img src="docs/assets/sentry-logo-black.svg" alt="Sentry" width="150">](https://sentry.io/_/open-source/)

HttpRunner is in Sentry Sponsored plan.

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
