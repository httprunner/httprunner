# HttpRunner

[![LICENSE](https://img.shields.io/github/license/HttpRunner/HttpRunner.svg)](https://github.com/HttpRunner/HttpRunner/blob/master/LICENSE) [![travis-ci](https://travis-ci.org/HttpRunner/HttpRunner.svg?branch=master)](https://travis-ci.org/HttpRunner/HttpRunner) [![coveralls](https://coveralls.io/repos/github/HttpRunner/HttpRunner/badge.svg?branch=master)](https://coveralls.io/github/HttpRunner/HttpRunner?branch=master) [![pypi version](https://img.shields.io/pypi/v/HttpRunner.svg)](https://pypi.python.org/pypi/HttpRunner) [![pyversions](https://img.shields.io/pypi/pyversions/HttpRunner.svg)](https://pypi.python.org/pypi/HttpRunner)

Httprunner is an HTTP(S) protocol-oriented universal testing framework. Write testing scripts in `YAML/JSON` once, you can then achieve automated testing, performance testing, online monitoring, continuous integration and other testing needs.

Former name: `ApiTestEngine`.

## Design Philosophy

- Take full reuse of Python's existing powerful libraries: [`Requests`][Requests], [`unittest`][unittest] and [`Locust`][Locust].
- Convention over configuration.
- Pursuit of high rewards, write once and achieve a variety of testing needs

## Key Features

- Inherit all powerful features of [`Requests`][Requests], just have fun to handle HTTP(S) in human way.
- Define testcases in YAML or JSON format in concise and elegant manner.
- Record and generate testcases with [`HAR`][HAR] support. see [`har2case`][har2case].
- Supports `function`/`variable`/`extract`/`validate` mechanisms to create full test scenarios.
- Supports perfect hook mechanism.
- With `debugtalk.py` plugin, module functions can be auto-discovered in recursive upward directories.
- Testcases can be run in diverse ways, with single testset, multiple testsets, or entire project folder.
- Test report is concise and clear, with detailed log records.
- With reuse of [`Locust`][Locust], you can run performance test without extra work.
- CLI command supported, perfect combination with `CI/CD`.

## Documentation

HttpRunner is rich documented.

- [`User documentation in English`][user-docs-en]
- [`中文用户使用手册`][user-docs-zh]
- [`开发历程记录博客`][development-blogs]

## Subscribe

关注 HttpRunner 的微信公众号，第一时间获得最新资讯。

![][qrcode_for_httprunner]

[Requests]: http://docs.python-requests.org/en/master/
[unittest]: https://docs.python.org/3/library/unittest.html
[Locust]: http://locust.io/
[PyUnitReport]: https://github.com/HttpRunner/PyUnitReport
[Jenkins]: https://jenkins.io/index.html
[har2case]: https://github.com/HttpRunner/har2case
[user-docs-en]: http://httprunner.org/
[user-docs-zh]: http://cn.httprunner.org/
[development-blogs]: http://debugtalk.com/tags/HttpRunner/
[HAR]: http://httparchive.org/
[Swagger]: https://swagger.io/
[Postman Collection Format]: http://blog.getpostman.com/2015/06/05/travelogue-of-postman-collection-format-v2/
[qrcode_for_httprunner]: https://raw.githubusercontent.com/HttpRunner/HttpRunner/master/docs/images/qrcode_for_httprunner.jpg
