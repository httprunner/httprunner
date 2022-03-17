# Release History

## v0.7.0 (2022-03-15)

- feat: support API layer for testcase #94
- feat: support global headers for testcase #95
- feat: support call referenced testcase by path in YAML/JSON testcases
- fix: decode failure when content-encoding is deflate
- fix: unstable RPS when load testing in high concurrency

## v0.6.4 (2022-03-10)

- feat: both support gPRC(default) and net/rpc mode in hashicorp plugin, switch with environment `HRP_PLUGIN_TYPE`
- refactor: move submodule `plugin` to separate repo `github.com/httprunner/funplugin`
- refactor: replace builtin json library with `json-iterator/go` to improve performance

## v0.6.3 (2022-03-04)

- feat: support customized setup/teardown hooks (variable assignment not supported)
- feat: add flag `--log-plugin` to turn on plugin logging
- change: add short flag `-c` for `--continue-on-failure`
- change: use `--log-requests-off` flag to turn off request & response details logging
- fix: support posting body in json array format
- fix: testcase format compatibility with HttpRunner

## v0.6.2 (2022-02-22)

- feat: support text/html extraction with regex
- change: json unmarshal to json.Number when parsing data
- fix: omit pseudo header names for HTTP/1, e.g. :authority
- fix: generate `headers.\"Content-Type\"` in har2case
- fix: incorrect data type when extracting data using jmespath
- fix: decode response body in brotli/gzip/deflate formats
- fix: omit print request/response body for non-text content
- fix: parse data for request cookie value

## v0.6.1 (2022-02-17)

- change: json unmarshal to float64 when parsing data
- fix: set request Content-Type for posting json only when not specified
- fix: failed to generate API test report when data is null
- fix: panic when assertion function not exists
- fix: broadcast to all rendezvous at once when spawn done

## v0.6.0 (2022-02-08)

- feat: implement `rendezvous` mechanism for data driven
- feat: upload release artifacts to aliyun oss
- feat: dump tests summary for execution results
- feat: generate html report for API testing
- change: remove sentry sdk

## v0.5.3 (2022-01-25)

- change: download package assets from aliyun OSS
- fix: disable color logging on Windows
- fix: print stderr when exec command failed
- fix: build hashicorp plugin failed when creating scaffold

## v0.5.2 (2022-01-19)

- feat: support creating and calling custom functions with [hashicorp/go-plugin](https://github.com/hashicorp/go-plugin)
- feat: add scaffold demo with hashicorp plugin
- feat: report events for initializing plugin
- fix: log failures when the assertion failed

## v0.5.1 (2022-01-13)

- feat: support specifying running cycles for load testing
- fix: ensure last stats reported when stop running

## v0.5.0 (2022-01-08)

- feat: support creating and calling custom functions with [go plugin](https://pkg.go.dev/plugin)
- feat: install hrp with one shell command
- feat: add `startproject` sub-command for creating scaffold project
- feat: report GA event for loading go plugin

## v0.4.0 (2022-01-05)

- feat: implement `parameterize` mechanism for data driven
- feat: add multiple builtin assertion methods and builtin functions

## v0.3.1 (2021-12-30)

- fix: set ulimit to 10240 before load testing
- fix: concurrent map writes in load testing

## v0.3.0 (2021-12-24)

- feat: implement `transaction` mechanism for load test
- feat: continue running next step when failure occurs with `--continue-on-failure` flag, default to failfast
- feat: report GA events with version
- feat: run load test with the given limit and burst as rate limiter, use `--spawn-count`, `--spawn-rate` and `--request-increase-rate` flag
- feat: report runner state to prometheus
- refactor: fork [boomer] as submodule initially and made a lot of changes
- change: update API models

## v0.2.2 (2021-12-07)

- refactor: update models to make API more concise
- change: remove mkdocs, move to [repo](https://github.com/httprunner/httprunner.github.io)

## v0.2.1 (2021-12-02)

- feat: push load testing metrics to [Prometheus Pushgateway][pushgateway]
- feat: report events with Google Analytics

## v0.2.0 (2021-11-19)

- feat: deploy mkdocs to github pages when PR merged
- feat: release hrp cli binaries automatically with github actions
- feat: add Makefile for running unittest and building hrp cli binary

## v0.1.0 (2021-11-18)

- feat: full support for HTTP(S)/1.1 methods
- feat: integrate [zerolog](https://github.com/rs/zerolog) for logging, include json log and pretty color console log
- feat: implement `har2case` for converting HAR to JSON/YAML testcases
- feat: extract and validate json response with [`jmespath`][jmespath]
- feat: run JSON/YAML testcases with builtin functions
- feat: support testcase and teststep level variables mechanism
- feat: integrate [`boomer`][boomer] standalone mode for load testing
- docs: init documentation website with [`mkdocs`][mkdocs]
- docs: add project badges, including go report card, codecov, github actions, FOSSA, etc.
- test: add CI test with [github actions][github-actions]
- test: integrate [sentry sdk][sentry sdk] for event reporting and analysis

[jmespath]: https://jmespath.org/
[mkdocs]: https://www.mkdocs.org/
[github-actions]: https://github.com/httprunner/hrp/actions
[boomer]: github.com/myzhan/boomer
[sentry sdk]: https://github.com/getsentry/sentry-go
[pushgateway]: https://github.com/prometheus/pushgateway
