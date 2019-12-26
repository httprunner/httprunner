
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

- Embrace open source, stand on giants' shoulders, like [`Requests`][Requests], [`unittest`][unittest] and [`Locust`][Locust].
- Convention over configuration.
- Pursuit of high rewards, write once and achieve a variety of testing needs

## Key Features

- Inherit all powerful features of [`Requests`][Requests], just have fun to handle HTTP(S) in human way.
- Define testcases in YAML or JSON format in concise and elegant manner.
- Record and generate testcases with [`HAR`][HAR] support. see [`har2case`][har2case].
- Supports `variables`/`extract`/`validate` mechanisms to create full test scenarios.
- Supports perfect hook mechanism.
- With `debugtalk.py` plugin, very easy to implement complex logic in testcase.
- Testcases can be run in diverse ways, with single testcase, multiple testcases, or entire project folder.
- Test report is concise and clear, with detailed log records.
- With reuse of [`Locust`][Locust], you can run performance test without extra work.
- CLI command supported, perfect combination with `CI/CD`.

## Documentation

HttpRunner is rich documented.

- [`中文用户使用手册`][user-docs-zh]
- [`开发历程记录博客`][development-blogs]
- [CHANGELOG](docs/CHANGELOG.md)

## Sponsors

Thank you to all our sponsors! ✨🍰✨ ([become a sponsor](docs/sponsors.md))

### 金牌赞助商（Gold Sponsor）

[<img src="docs/assets/hogwarts.png" alt="霍格沃兹测试学院" width="400">](https://testing-studio.com)

> [霍格沃兹测试学院](https://testing-studio.com) 是由测吧（北京）科技有限公司与知名软件测试社区 [TesterHome](https://testerhome.com/) 合作的高端教育品牌。由 BAT 一线**测试大咖执教**，提供**实战驱动**的接口自动化测试、移动自动化测试、性能测试、持续集成与 DevOps 等技术培训，以及测试开发优秀人才内推服务。[点击学习!](https://ke.qq.com/course/254956?flowToken=1014690)

霍格沃兹测试学院是 HttpRunner 的首家金牌赞助商。 

### 开源服务赞助商（Open Source Sponsor）

[<img src="docs/assets/sentry-logo-black.svg" alt="Sentry" width="150">](https://sentry.io/_/open-source/)

HttpRunner is in Sentry Sponsored plan.

## How to Contribute

1. Check for [open issues](https://github.com/httprunner/httprunner/issues) or [open a fresh issue](https://github.com/httprunner/httprunner/issues/new/choose) to start a discussion around a feature idea or a bug.
2. Fork [the repository](https://github.com/httprunner/httprunner) on GitHub to start making your changes to the **master** branch (or branch off of it). You also need to comply with the [development rules](https://github.com/httprunner/docs/blob/master/en/docs/dev-rules.md).
3. Write a test which shows that the bug was fixed or that the feature works as expected.
4. Send a pull request, you will then become a [contributor](https://github.com/httprunner/httprunner/graphs/contributors) after it gets merged and published.

## Subscribe

关注 HttpRunner 的微信公众号，第一时间获得最新资讯。

![](docs/assets/qrcode.jpg)

[Requests]: http://docs.python-requests.org/en/master/
[unittest]: https://docs.python.org/3/library/unittest.html
[Locust]: http://locust.io/
[har2case]: https://github.com/httprunner/har2case
[user-docs-zh]: http://docs.httprunner.org/
[development-blogs]: http://debugtalk.com/tags/httprunner/
[HAR]: http://httparchive.org/
[Swagger]: https://swagger.io/

