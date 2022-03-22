
# HttpRunner

[![downloads](https://pepy.tech/badge/httprunner)](https://pepy.tech/project/httprunner)
[![unittest](https://github.com/httprunner/httprunner/workflows/unittest/badge.svg
)](https://github.com/httprunner/httprunner/actions)
[![integration-test](https://github.com/httprunner/httprunner/workflows/integration_test/badge.svg
)](https://github.com/httprunner/httprunner/actions)
[![codecov](https://codecov.io/gh/httprunner/httprunner/branch/master/graph/badge.svg)](https://codecov.io/gh/httprunner/httprunner)
[![pypi version](https://img.shields.io/pypi/v/httprunner.svg)](https://pypi.python.org/pypi/httprunner)
[![pyversions](https://img.shields.io/pypi/pyversions/httprunner.svg)](https://pypi.python.org/pypi/httprunner)
[![TesterHome](https://img.shields.io/badge/TTF-TesterHome-2955C5.svg)](https://testerhome.com/github_statistics)

*HttpRunner* is a simple & elegant, yet powerful HTTP(S) testing framework. Enjoy! ✨ 🚀 ✨

> 欢迎参加 HttpRunner [用户调研问卷][survey]，你的反馈将帮助 HttpRunner 更好地成长！

## Design Philosophy

- Convention over configuration
- ROI matters
- Embrace open source, leverage [`requests`][requests], [`pytest`][pytest], [`pydantic`][pydantic], [`allure`][allure] and [`locust`][locust].

## Key Features

- [x] Inherit all powerful features of [`requests`][requests], just have fun to handle HTTP(S) in human way.
- [x] Define testcase in YAML or JSON format, run with [`pytest`][pytest] in concise and elegant manner.
- [x] Record and generate testcases with [`HAR`][HAR] support.
- [x] Supports `variables`/`extract`/`validate`/`hooks` mechanisms to create extremely complex test scenarios.
- [x] With `debugtalk.py` plugin, any function can be used in any part of your testcase.
- [x] With [`jmespath`][jmespath], extract and validate json response has never been easier.
- [x] With [`pytest`][pytest], hundreds of plugins are readily available.
- [x] With [`allure`][allure], test report can be pretty nice and powerful.
- [x] With reuse of [`locust`][locust], you can run performance test without extra work.
- [x] CLI command supported, perfect combination with `CI/CD`.

## Sponsors

Thank you to all our sponsors! ✨🍰✨ ([become a sponsor](sponsors.md))

### 金牌赞助商（Gold Sponsor）

[<img src="assets/hogwarts.jpeg" alt="霍格沃兹测试开发学社" width="400">](https://ceshiren.com/)

> [霍格沃兹测试开发学社](http://qrcode.testing-studio.com/f?from=httprunner&url=https://ceshiren.com)是业界领先的测试开发技术高端教育品牌，隶属于[测吧（北京）科技有限公司](http://qrcode.testing-studio.com/f?from=httprunner&url=https://www.testing-studio.com) 。学院课程由一线大厂测试经理与资深测试开发专家参与研发，实战驱动。课程涵盖 web/app 自动化测试、接口测试、性能测试、安全测试、持续集成/持续交付/DevOps，测试左移&右移、精准测试、测试平台开发、测试管理等内容，帮助测试工程师实现测试开发技术转型。通过优秀的学社制度（奖学金、内推返学费、行业竞赛等多种方式）来实现学员、学社及用人企业的三方共赢。

> [进入测试开发技术能力测评!](http://qrcode.testing-studio.com/f?from=httprunner&url=https://ceshiren.com/t/topic/14940)

### 开源服务赞助商（Open Source Sponsor）

[<img src="assets/sentry-logo-black.svg" alt="Sentry" width="150">](https://sentry.io/_/open-source/)

HttpRunner is in Sentry Sponsored plan.

## Subscribe

关注 HttpRunner 的微信公众号，第一时间获得最新资讯。

<img src="assets/qrcode.jpg" alt="HttpRunner" width="200">

如果你期望加入 HttpRunner 核心用户群，请填写[用户调研问卷][survey]并留下你的联系方式，作者将拉你进群。

[requests]: http://docs.python-requests.org/en/master/
[pytest]: https://docs.pytest.org/
[pydantic]: https://pydantic-docs.helpmanual.io/
[locust]: http://locust.io/
[jmespath]: https://jmespath.org/
[allure]: https://docs.qameta.io/allure/
[HAR]: http://httparchive.org/
[survey]: https://wj.qq.com/s2/9699514/0d19/
