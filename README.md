# hrp (HttpRunner+)

[![Go Reference](https://pkg.go.dev/badge/github.com/httprunner/hrp.svg)](https://pkg.go.dev/github.com/httprunner/hrp)
[![Github Actions](https://github.com/httprunner/hrp/actions/workflows/unittest.yml/badge.svg)](https://github.com/httprunner/hrp/actions)
[![codecov](https://codecov.io/gh/httprunner/hrp/branch/main/graph/badge.svg?token=HPCQWCD7KO)](https://codecov.io/gh/httprunner/hrp)
[![Go Report Card](https://goreportcard.com/badge/github.com/httprunner/hrp)](https://goreportcard.com/report/github.com/httprunner/hrp)
[![FOSSA Status](https://app.fossa.com/api/projects/custom%2B27856%2Fgithub.com%2Fhttprunner%2Fhrp.svg?type=shield)](https://app.fossa.com/reports/c2742455-c8ab-4b13-8fd7-4a35ba0b2840)

`hrp` is a golang implementation of [HttpRunner]. Ideally, hrp will be fully compatible with HttpRunner, including testcase format and usage. What's more, hrp will integrate Boomer natively to be a better load generator for [locust].

## Key Features

![flow chart](docs/assets/flow.jpg)

- [x] Full support for HTTP(S) requests, more protocols are also in the plan.
- [x] Testcases can be described in multiple formats, `YAML`/`JSON`/`Golang`, and they are interchangeable.
- [x] With [`HAR`][HAR] support, you can use Charles/Fiddler/Chrome/etc as a script recording generator.
- [x] Supports `variables`/`extract`/`validate`/`hooks` mechanisms to create extremely complex test scenarios.
- [ ] Built-in integration of rich functions, and you can also use [`go plugin`][plugin] to create and call custom functions.
- [x] Inherit all powerful features of [`Boomer`][Boomer] and [`locust`][locust], you can run `load test` without extra work.
- [x] Use it as a `CLI tool` or as a `library` are both supported.

See [CHANGELOG].

## Quick Start

### use as CLI tool

```bash
$ go get -u github.com/httprunner/hrp/hrp
```

Since installed, you will get a `hrp` command with multiple sub-commands.

```text
$ hrp -h
hrp (HttpRunner+) is the one-stop solution for HTTP(S) testing. Enjoy! ✨ 🚀 ✨

License: Apache-2.0
Github: https://github.com/httprunner/hrp
Copyright 2021 debugtalk

Usage:
  hrp [command]

Available Commands:
  boom        run load test with boomer
  completion  generate the autocompletion script for the specified shell
  har2case    Convert HAR to json/yaml testcase files
  help        Help about any command
  run         run API test

Flags:
  -h, --help               help for hrp
      --log-json           set log to json format
  -l, --log-level string   set log level (default "INFO")
  -v, --version            version for hrp

Use "hrp [command] --help" for more information about a command.
```

You can use `hrp run` command to run HttpRunner JSON/YAML testcases. The following is an example running [examples/demo.json][demo.json]

<details>
<summary>$ hrp run examples/demo.json</summary>

```text
9:22PM INF Set log to color console other than JSON format.
9:22PM INF Set log level to INFO
9:22PM INF [init] SetDebug debug=true
9:22PM INF load json testcase path=/Users/debugtalk/MyProjects/HttpRunner-dev/hrp/examples/demo.json
9:22PM INF call function success arguments=[5] funcName=gen_random_string output=rWRNY
9:22PM INF call function success arguments=[12.3,3.45] funcName=max output=12.3
9:22PM INF run testcase start testcase="demo with complex mechanisms"
9:22PM INF run step start step="get with params"
9:22PM INF call function success arguments=[12.3,34.5] funcName=max output=34.5
-------------------- request --------------------
GET /get?foo1=rWRNY&foo2=34.5 HTTP/1.1
Host: postman-echo.com
User-Agent: HttpRunnerPlus


==================== response ===================
HTTP/1.1 200 OK
Content-Length: 304
Connection: keep-alive
Content-Type: application/json; charset=utf-8
Date: Tue, 07 Dec 2021 13:22:50 GMT
Etag: W/"130-gmtE0VWiyE0mXUGoJe5AyhMQ2ig"
Set-Cookie: sails.sid=s%3AEWPwP8H-nbpSrCseeulwDQ8OEtRy1pGu.aHV6KrEIiFgaJsUAuDmmmJCYiV6XkrHLS%2Fd9g9vtZQw; Path=/; HttpOnly
Vary: Accept-Encoding

{"args":{"foo1":"rWRNY","foo2":"34.5"},"headers":{"x-forwarded-proto":"https","x-forwarded-port":"443","host":"postman-echo.com","x-amzn-trace-id":"Root=1-61af602a-5eea88ee21122daf4e8dfe95","user-agent":"HttpRunnerPlus","accept-encoding":"gzip"},"url":"https://postman-echo.com/get?foo1=rWRNY&foo2=34.5"}
--------------------------------------------------
9:22PM INF extract value from=body.args.foo1 value=rWRNY
9:22PM INF set variable value=rWRNY variable=varFoo1
9:22PM INF validate status_code assertMethod=equals checkValue=200 expectValue=200 result=true
9:22PM INF validate headers."Content-Type" assertMethod=startswith checkValue="application/json; charset=utf-8" expectValue=application/json result=true
9:22PM INF validate body.args.foo1 assertMethod=length_equals checkValue=rWRNY expectValue=5 result=true
9:22PM INF validate $varFoo1 assertMethod=length_equals checkValue=rWRNY expectValue=5 result=true
9:22PM INF validate body.args.foo2 assertMethod=equals checkValue=34.5 expectValue=34.5 result=true
9:22PM INF run step end exportVars={"varFoo1":"rWRNY"} step="get with params" success=true
9:22PM INF run step start step="post json data"
9:22PM INF call function success arguments=[12.3,3.45] funcName=max output=12.3
-------------------- request --------------------
POST /post HTTP/1.1
Host: postman-echo.com
Content-Type: application/json; charset=UTF-8

{"foo1":"rWRNY","foo2":12.3}
==================== response ===================
HTTP/1.1 200 OK
Content-Length: 424
Connection: keep-alive
Content-Type: application/json; charset=utf-8
Date: Tue, 07 Dec 2021 13:22:50 GMT
Etag: W/"1a8-5fCAlcltnCS4Ed/6OxpH9i9dlKs"
Set-Cookie: sails.sid=s%3As1b8P7f8sc3JRNumS-XJrzbwb5oxdkOs.pXRRifddVUiWuzAxwBikBxf3ayM8OahgDDzP7kSnMCc; Path=/; HttpOnly
Vary: Accept-Encoding

{"args":{},"data":{"foo1":"rWRNY","foo2":12.3},"files":{},"form":{},"headers":{"x-forwarded-proto":"https","x-forwarded-port":"443","host":"postman-echo.com","x-amzn-trace-id":"Root=1-61af602a-54fcb6412d2d064822bcdd5f","content-length":"28","user-agent":"Go-http-client/1.1","content-type":"application/json; charset=UTF-8","accept-encoding":"gzip"},"json":{"foo1":"rWRNY","foo2":12.3},"url":"https://postman-echo.com/post"}
--------------------------------------------------
9:22PM INF validate status_code assertMethod=equals checkValue=200 expectValue=200 result=true
9:22PM INF validate body.json.foo1 assertMethod=length_equals checkValue=rWRNY expectValue=5 result=true
9:22PM INF validate body.json.foo2 assertMethod=equals checkValue=12.3 expectValue=12.3 result=true
9:22PM INF run step end exportVars=null step="post json data" success=true
9:22PM INF run step start step="post form data"
9:22PM INF call function success arguments=[12.3,3.45] funcName=max output=12.3
-------------------- request --------------------
POST /post HTTP/1.1
Host: postman-echo.com
Content-Type: application/x-www-form-urlencoded; charset=UTF-8

foo1=rWRNY&foo2=12.3
==================== response ===================
HTTP/1.1 200 OK
Content-Length: 445
Connection: keep-alive
Content-Type: application/json; charset=utf-8
Date: Tue, 07 Dec 2021 13:22:50 GMT
Etag: W/"1bd-V7gWOjKCZvyBWVyqprN77w2dmXE"
Set-Cookie: sails.sid=s%3Aj4sUA8hI4rAt9JMq1m4k_chSDlfkAEBV.ZfisF4bIH2e7iBY6%2BSHqUbHNBbhCzZi%2Fu4byLDdxy%2B4; Path=/; HttpOnly
Vary: Accept-Encoding

{"args":{},"data":"","files":{},"form":{"foo1":"rWRNY","foo2":"12.3"},"headers":{"x-forwarded-proto":"https","x-forwarded-port":"443","host":"postman-echo.com","x-amzn-trace-id":"Root=1-61af602a-2cc056eb54ba2f0c6850d84a","content-length":"20","user-agent":"Go-http-client/1.1","content-type":"application/x-www-form-urlencoded; charset=UTF-8","accept-encoding":"gzip"},"json":{"foo1":"rWRNY","foo2":"12.3"},"url":"https://postman-echo.com/post"}
--------------------------------------------------
9:22PM INF validate status_code assertMethod=equals checkValue=200 expectValue=200 result=true
9:22PM INF validate body.form.foo1 assertMethod=length_equals checkValue=rWRNY expectValue=5 result=true
9:22PM INF validate body.form.foo2 assertMethod=equals checkValue=12.3 expectValue=12.3 result=true
9:22PM INF run step end exportVars=null step="post form data" success=true
9:22PM INF run testcase end testcase="demo with complex mechanisms"
```
</details>

### use as library

Beside using `hrp` as a CLI tool, you can also use it as golang library.

```bash
$ go get -u github.com/httprunner/hrp
```

This is an example of `HttpRunner+` testcase. You can find more in the [`examples`][examples] directory.


<details>
<summary>demo</summary>

```go
import (
    "testing"

    "github.com/httprunner/hrp"
)

func TestCaseDemo(t *testing.T) {
    demoTestCase := &hrp.TestCase{
        Config: hrp.NewConfig("demo with complex mechanisms").
            SetBaseURL("https://postman-echo.com").
            WithVariables(map[string]interface{}{ // global level variables
                "n":       5,
                "a":       12.3,
                "b":       3.45,
                "varFoo1": "${gen_random_string($n)}",
                "varFoo2": "${max($a, $b)}", // 12.3; eval with built-in function
            }),
        TestSteps: []hrp.IStep{
            hrp.NewStep("get with params").
                WithVariables(map[string]interface{}{ // step level variables
                    "n":       3,                // inherit config level variables if not set in step level, a/varFoo1
                    "b":       34.5,             // override config level variable if existed, n/b/varFoo2
                    "varFoo2": "${max($a, $b)}", // 34.5; override variable b and eval again
                }).
                GET("/get").
                WithParams(map[string]interface{}{"foo1": "$varFoo1", "foo2": "$varFoo2"}). // request with params
                WithHeaders(map[string]string{"User-Agent": "HttpRunnerPlus"}).             // request with headers
                Extract().
                WithJmesPath("body.args.foo1", "varFoo1"). // extract variable with jmespath
                Validate().
                AssertEqual("status_code", 200, "check response status code").        // validate response status code
                AssertStartsWith("headers.\"Content-Type\"", "application/json", ""). // validate response header
                AssertLengthEqual("body.args.foo1", 5, "check args foo1").            // validate response body with jmespath
                AssertLengthEqual("$varFoo1", 5, "check args foo1").                  // assert with extracted variable from current step
                AssertEqual("body.args.foo2", "34.5", "check args foo2"),             // notice: request params value will be converted to string
            hrp.NewStep("post json data").
                POST("/post").
                WithBody(map[string]interface{}{
                    "foo1": "$varFoo1",       // reference former extracted variable
                    "foo2": "${max($a, $b)}", // 12.3; step level variables are independent, variable b is 3.45 here
                }).
                Validate().
                AssertEqual("status_code", 200, "check status code").
                AssertLengthEqual("body.json.foo1", 5, "check args foo1").
                AssertEqual("body.json.foo2", 12.3, "check args foo2"),
            hrp.NewStep("post form data").
                POST("/post").
                WithHeaders(map[string]string{"Content-Type": "application/x-www-form-urlencoded; charset=UTF-8"}).
                WithBody(map[string]interface{}{
                    "foo1": "$varFoo1",       // reference former extracted variable
                    "foo2": "${max($a, $b)}", // 12.3; step level variables are independent, variable b is 3.45 here
                }).
                Validate().
                AssertEqual("status_code", 200, "check status code").
                AssertLengthEqual("body.form.foo1", 5, "check args foo1").
                AssertEqual("body.form.foo2", "12.3", "check args foo2"), // form data will be converted to string
        },
    }

    err := hrp.NewRunner(nil).Run(demoTestCase) // hrp.Run(demoTestCase)
    if err != nil {
        t.Fatalf("run testcase error: %v", err)
    }
}
```
</details>

## Sponsors

Thank you to all our sponsors! ✨🍰✨ ([become a sponsor](sponsors.md))

### Gold Sponsor

[<img src="docs/assets/hogwarts.jpeg" alt="霍格沃兹测试开发学社" width="400">](https://ceshiren.com/)

> [霍格沃兹测试开发学社](http://qrcode.testing-studio.com/f?from=httprunner&url=https://ceshiren.com)是业界领先的测试开发技术高端教育品牌，隶属于[测吧（北京）科技有限公司](http://qrcode.testing-studio.com/f?from=httprunner&url=https://www.testing-studio.com) 。学院课程由一线大厂测试经理与资深测试开发专家参与研发，实战驱动。课程涵盖 web/app 自动化测试、接口测试、性能测试、安全测试、持续集成/持续交付/DevOps，测试左移&右移、精准测试、测试平台开发、测试管理等内容，帮助测试工程师实现测试开发技术转型。通过优秀的学社制度（奖学金、内推返学费、行业竞赛等多种方式）来实现学员、学社及用人企业的三方共赢。

> [进入测试开发技术能力测评!](http://qrcode.testing-studio.com/f?from=httprunner&url=https://ceshiren.com/t/topic/14940)

### Open Source Sponsor

[<img src="docs/assets/sentry-logo-black.svg" alt="Sentry" width="150">](https://sentry.io/_/open-source/)

## Subscribe

关注 HttpRunner 的微信公众号，第一时间获得最新资讯。

<img src="docs/assets/qrcode.jpg" alt="HttpRunner" width="200">

[HttpRunner]: https://github.com/httprunner/httprunner
[Boomer]: https://github.com/myzhan/boomer
[locust]: https://github.com/locustio/locust
[jmespath]: https://jmespath.org/
[allure]: https://docs.qameta.io/allure/
[HAR]: http://httparchive.org/
[plugin]: https://pkg.go.dev/plugin
[demo.json]: https://github.com/httprunner/hrp/blob/main/examples/demo.json
[examples]: https://github.com/httprunner/hrp/blob/main/examples/
[CHANGELOG]: docs/CHANGELOG.md