
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

## Design Philosophy

- Convention over configuration
- ROI matters
- Embrace open source, leverage [`requests`][requests], [`pytest`][pytest], [`pydantic`][pydantic], [`allure`][allure] and [`locust`][locust].

## Key Features

- Inherit all powerful features of [`requests`][requests], just have fun to handle HTTP(S) in human way.
- Define testcase in YAML or JSON format, run with [`pytest`][pytest] in concise and elegant manner.
- Record and generate testcases with [`HAR`][HAR] support.
- Supports `variables`/`extract`/`validate`/`hooks` mechanisms to create extremely complex test scenarios.
- With `debugtalk.py` plugin, any function can be used in any part of your testcase.
- With [`jmespath`][jmespath], extract and validate json response has never been easier.
- With [`pytest`][pytest], hundreds of plugins are readily available.
- With [`allure`][allure], test report can be pretty nice and powerful.
- With reuse of [`locust`][locust], you can run performance test without extra work.
- CLI command supported, perfect combination with `CI/CD`.

## Sponsors

Thank you to all our sponsors! ✨🍰✨ ([become a sponsor](docs/sponsors.md))

### 金牌赞助商（Gold Sponsor）

[<img src="docs/assets/hogwarts.jpeg" alt="霍格沃兹测试学院" width="500">](https://ceshiren.com/)

> [霍格沃兹测试开发学社](http://qrcode.testing-studio.com/f?from=httprunner&url=https://ceshiren.com)是业界领先的测试开发技术高端教育品牌，隶属于[测吧（北京）科技有限公司](http://qrcode.testing-studio.com/f?from=httprunner&url=https://www.testing-studio.com) 。入学会先进行技术能力测评，因材施教，帮助测试工程师实现从手工到测试开发技术转型。通过优秀的学社制度（奖学金制度、内推返学费制度、行业竞赛等多种方式）来实现学员、学社及用人企业的三方共赢。

> 学院课程由一线大厂测试经理与资深测试开发专家参与研发，以实战驱动为导向，紧贴互联网名企的用人需求。课程方向涵盖移动app自动化测试、接口自动化测试、性能测试、安全测试、持续集成/持续交付/DevOps，测试左移、测试右移、精准测试、测试平台开发、测试管理等内容，全面提升测试开发工程师的技术实力。

> [进入测试开发技术能力测评!](http://qrcode.testing-studio.com/f?from=httprunner&url=https://ceshiren.com/t/topic/14940)

### 开源服务赞助商（Open Source Sponsor）

[<img src="docs/assets/sentry-logo-black.svg" alt="Sentry" width="150">](https://sentry.io/_/open-source/)

HttpRunner is in Sentry Sponsored plan.

## Subscribe

关注 HttpRunner 的微信公众号，第一时间获得最新资讯。

![](docs/assets/qrcode.jpg)

[requests]: http://docs.python-requests.org/en/master/
[pytest]: https://docs.pytest.org/
[pydantic]: https://pydantic-docs.helpmanual.io/
[locust]: http://locust.io/
[jmespath]: https://jmespath.org/
[allure]: https://docs.qameta.io/allure/
[HAR]: http://httparchive.org/


