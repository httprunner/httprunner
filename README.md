
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

[<img src="docs/assets/hogwarts.png" alt="霍格沃兹测试学院" width="400">](https://testing-studio.com)

> [霍格沃兹测试学院](https://testing-studio.com) 是由测吧（北京）科技有限公司与知名软件测试社区 [TesterHome](https://testerhome.com/) 合作的高端教育品牌。由 BAT 一线**测试大咖执教**，提供**实战驱动**的接口自动化测试、移动自动化测试、性能测试、持续集成与 DevOps 等技术培训，以及测试开发优秀人才内推服务。[点击学习!](https://ke.qq.com/course/254956?flowToken=1014690)

霍格沃兹测试学院是 HttpRunner 的首家金牌赞助商。 

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


