# Release History

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

- feat: push load testing metrics to Prometheus Pushgateway
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
