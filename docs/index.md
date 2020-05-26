
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

> This docs site is corresponding to the latest version `3.x`, for `2.x` you can reference [`archive link`](https://v2.httprunner.org/).

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

## Subscribe

关注 HttpRunner 的微信公众号，第一时间获得最新资讯。

![](/assets/qrcode.jpg)

[requests]: http://docs.python-requests.org/en/master/
[pytest]: https://docs.pytest.org/
[pydantic]: https://pydantic-docs.helpmanual.io/
[locust]: http://locust.io/
[jmespath]: https://jmespath.org/
[allure]: https://docs.qameta.io/allure/
[HAR]: http://httparchive.org/
