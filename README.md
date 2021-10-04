# HttpBoomer

[![Go Reference](https://pkg.go.dev/badge/github.com/httprunner/httpboomer.svg)](https://pkg.go.dev/github.com/httprunner/httpboomer)
[![Github Actions](https://github.com/httprunner/HttpBoomer/actions/workflows/main.yml/badge.svg)](https://github.com/httprunner/HttpBoomer/actions)
[![codecov](https://codecov.io/gh/httprunner/HttpBoomer/branch/main/graph/badge.svg?token=HPCQWCD7KO)](https://codecov.io/gh/httprunner/HttpBoomer)
[![Go Report Card](https://goreportcard.com/badge/github.com/httprunner/HttpBoomer)](https://goreportcard.com/report/github.com/httprunner/HttpBoomer)
[![FOSSA Status](https://app.fossa.com/api/projects/custom%2B27856%2Fgithub.com%2Fhttprunner%2FHttpBoomer.svg?type=shield)](https://app.fossa.com/reports/fb0e64a7-7dcf-48bb-8de9-8f0e016b903b)

> HttpBoomer = [HttpRunner] + [Boomer]

HttpBoomer is a golang implementation of [HttpRunner]. Ideally, HttpBoomer will be fully compatible with HttpRunner, including testcase format and usage. What's more, HttpBoomer will integrate Boomer natively to be a better load generator for [locust].

## Key Features

- [x] Full support for HTTP(S) requests, more protocols are also in the plan.
- [ ] Testcases can be described in multiple formats, `YAML`/`JSON`/`Golang`, and they are interchangeable.
- [ ] With [`HAR`][HAR] support, you can use Charles/Fiddler/Chrome/etc as a script recording generator.
- [x] Supports `variables`/`extract`/`validate`/`hooks` mechanisms to create extremely complex test scenarios.
- [ ] Built-in integration of rich functions, and you can also use [`go plugin`][plugin] to create and call custom functions.
- [x] Inherit all powerful features of [`Boomer`][Boomer] and [`locust`][locust], you can run `load test` without extra work.
- [ ] Use it as a `CLI tool` or as a `library` are both supported.

## Quick Start

[HttpRunner]: https://github.com/httprunner/httprunner
[Boomer]: https://github.com/myzhan/boomer
[locust]: https://github.com/locustio/locust
[jmespath]: https://jmespath.org/
[allure]: https://docs.qameta.io/allure/
[HAR]: http://httparchive.org/
[plugin]: https://pkg.go.dev/plugin
