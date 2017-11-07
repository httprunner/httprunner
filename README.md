# HttpRunner

[![license](https://img.shields.io/github/license/HttpRunner/HttpRunner.svg)](https://github.com/HttpRunner/HttpRunner/blob/master/LICENSE)
[![Build Status](https://travis-ci.org/debugtalk/HttpRunner.svg?branch=master)](https://travis-ci.org/HttpRunner/HttpRunner)
[![Coverage Status](https://coveralls.io/repos/github/debugtalk/HttpRunner/badge.svg?branch=master)](https://coveralls.io/github/debugtalk/HttpRunner?branch=master)
[![PyPI](https://img.shields.io/pypi/v/HttpRunner.svg)](https://pypi.python.org/pypi/HttpRunner)
[![PyPI](https://img.shields.io/pypi/pyversions/HttpRunner.svg)](https://pypi.python.org/pypi/HttpRunner)

New name for `ApiTestEngine`.

## Design Philosophy

Take full reuse of Python's existing powerful libraries: [`Requests`][requests], [`unittest`][unittest] and [`Locust`][Locust]. And achieve the goal of API automation test, production environment monitoring, and API performance test, with a concise and elegant manner.

## Key Features

- Inherit all powerful features of [`Requests`][requests], just have fun to handle HTTP in human way.
- Define testcases in YAML or JSON format in concise and elegant manner.
- Supports `function`/`variable`/`extract`/`validate` mechanisms to create full test scenarios.
- With `debugtalk.py` plugin, module functions can be auto-discovered in recursive upward directories.
- Testcases can be run in diverse ways, with single testset, multiple testsets, or entire project folder.
- Test report is concise and clear, with detailed log records. See [`PyUnitReport`][PyUnitReport].
- With reuse of [`Locust`][Locust], you can run performance test without extra work.
- CLI command supported, perfect combination with [Jenkins][Jenkins].

[*`Background Introduction (中文版)`*](docs/background-CN.md) | [*`Feature Descriptions (中文版)`*](docs/feature-descriptions-CN.md)

## Installation/Upgrade

```bash
$ pip install HttpRunner
```

To upgrade all specified packages to the newest available version, you should add the `-U` option.

If there is a problem with the installation or upgrade, you can check the [`FAQ`](docs/FAQ.md).

To ensure the installation or upgrade is successful, you can execute command `httprunner -V` to see if you can get the correct version number.

```text
$ httprunner -V
HttpRunner version: 0.8.0
```

Execute the command `httprunner -h` to view command help.

```text
$ httprunner -h
usage: httprunner [-h] [-V] [--log-level LOG_LEVEL] [--report-name REPORT_NAME]
           [--failfast] [--startproject STARTPROJECT]
           [testset_paths [testset_paths ...]]

HttpRunner.

positional arguments:
  testset_paths         testset file path

optional arguments:
  -h, --help            show this help message and exit
  -V, --version         show version
  --log-level LOG_LEVEL
                        Specify logging level, default is INFO.
  --report-name REPORT_NAME
                        Specify report name, default is generated time.
  --failfast            Stop the test run on the first error or failure.
  --startproject STARTPROJECT
                        Specify new project name.
```

## Write testcases

It is recommended to write testcases in `YAML` format.

And here is testset example of typical scenario: get `token` at the beginning, and each subsequent requests should take the `token` in the headers.

```yaml
- config:
    name: "create user testsets."
    variables:
        - user_agent: 'iOS/10.3'
        - device_sn: ${gen_random_string(15)}
        - os_platform: 'ios'
        - app_version: '2.8.6'
    request:
        base_url: http://127.0.0.1:5000
        headers:
            Content-Type: application/json
            device_sn: $device_sn

- test:
    name: get token
    request:
        url: /api/get-token
        method: POST
        headers:
            user_agent: $user_agent
            device_sn: $device_sn
            os_platform: $os_platform
            app_version: $app_version
        json:
            sign: ${get_sign($user_agent, $device_sn, $os_platform, $app_version)}
    extract:
        - token: content.token
    validate:
        - {"check": "status_code", "comparator": "eq", "expected": 200}
        - {"check": "content.token", "comparator": "len_eq", "expected": 16}

- test:
    name: create user which does not exist
    request:
        url: /api/users/1000
        method: POST
        headers:
            token: $token
        json:
            name: "user1"
            password: "123456"
    validate:
        - {"check": "status_code", "comparator": "eq", "expected": 201}
        - {"check": "content.success", "comparator": "eq", "expected": true}
```

Function invoke is supported in `YAML/JSON` format testcases, such as `gen_random_string` and `get_sign` above. This mechanism relies on the `debugtak.py` hot plugin, with which we can define functions in `debugtak.py` file, and then functions can be auto discovered and invoked in runtime.

For detailed regulations of writing testcases, you can read the [`QuickStart`][quickstart] documents.

## Run testcases

`HttpRunner` can run testcases in diverse ways.

You can run single testset by specifying testset file path.

```text
$ httprunner filepath/testcase.yml
```

You can also run several testsets by specifying multiple testset file paths.

```text
$ httprunner filepath1/testcase1.yml filepath2/testcase2.yml
```

If you want to run testsets of a whole project, you can achieve this goal by specifying the project folder path.

```text
$ httprunner testcases_folder_path
```

When you do continuous integration test or production environment monitoring with `Jenkins`, you may need to send test result notification. For instance, you can send email with mailgun service as below.

```text
$ httprunner filepath/testcase.yml --report-name ${BUILD_NUMBER} \
    --mailgun-smtp-username "qa@debugtalk.com" \
    --mailgun-smtp-password "12345678" \
    --email-sender excited@samples.mailgun.org \
    --email-recepients ${MAIL_RECEPIENTS} \
    --jenkins-job-name ${JOB_NAME} \
    --jenkins-job-url ${JOB_URL} \
    --jenkins-build-number ${BUILD_NUMBER}
```

## Performance test

With reuse of [`Locust`][Locust], you can run performance test without extra work.

```bash
$ locusts -V
[2017-08-26 23:45:42,246] bogon/INFO/stdout: Locust 0.8a2
[2017-08-26 23:45:42,246] bogon/INFO/stdout:
```

For full usage, you can run `locusts -h` to see help, and you will find that it is the same with `locust -h`.

The only difference is the `-f` argument. If you specify `-f` with a Python locustfile, it will be the same as `locust`, while if you specify `-f` with a `YAML/JSON` testcase file, it will convert to Python locustfile first and then pass to `locust`.

```bash
$ locusts -f examples/first-testcase.yml
[2017-08-18 17:20:43,915] Leos-MacBook-Air.local/INFO/locust.main: Starting web monitor at *:8089
[2017-08-18 17:20:43,918] Leos-MacBook-Air.local/INFO/locust.main: Starting Locust 0.8a2
```

In this case, you can reuse all features of [`Locust`][Locust].

That’s not all about it. With the argument `--full-speed`, you can even start locust with master and several slaves (default to cpu cores number) at one time, which means you can leverage all cpus of your machine.

```bash
$ locusts -f examples/first-testcase.yml --full-speed
[2017-08-26 23:51:47,071] bogon/INFO/locust.main: Starting web monitor at *:8089
[2017-08-26 23:51:47,075] bogon/INFO/locust.main: Starting Locust 0.8a2
[2017-08-26 23:51:47,078] bogon/INFO/locust.main: Starting Locust 0.8a2
[2017-08-26 23:51:47,080] bogon/INFO/locust.main: Starting Locust 0.8a2
[2017-08-26 23:51:47,083] bogon/INFO/locust.main: Starting Locust 0.8a2
[2017-08-26 23:51:47,084] bogon/INFO/locust.runners: Client 'bogon_656e0af8e968a8533d379dd252422ad3' reported as ready. Currently 1 clients ready to swarm.
[2017-08-26 23:51:47,085] bogon/INFO/locust.runners: Client 'bogon_09f73850252ee4ec739ed77d3c4c6dba' reported as ready. Currently 2 clients ready to swarm.
[2017-08-26 23:51:47,084] bogon/INFO/locust.main: Starting Locust 0.8a2
[2017-08-26 23:51:47,085] bogon/INFO/locust.runners: Client 'bogon_869f7ed671b1a9952b56610f01e2006f' reported as ready. Currently 3 clients ready to swarm.
[2017-08-26 23:51:47,085] bogon/INFO/locust.runners: Client 'bogon_80a804cda36b80fac17b57fd2d5e7cdb' reported as ready. Currently 4 clients ready to swarm.
```

![](docs/locusts-full-speed.jpg)

Enjoy!

## Supported Python Versions

Python `2.7`, `3.4`, `3.5` and `3.6`.

`HttpRunner` has been tested on `macOS`, `Linux` and `Windows` platforms.

## Development

To develop or debug `HttpRunner`, you can install relevant requirements and use `main-ate.py` or `main-locust.py` as entrances.

```bash
$ pip install -r requirements.txt
$ python main-ate -h
$ python main-locust -h
```

## To learn more ...

- [《接口自动化测试的最佳工程实践（ApiTestEngine）》](http://debugtalk.com/post/ApiTestEngine-api-test-best-practice/)
- [`ApiTestEngine QuickStart`][quickstart]
- [《ApiTestEngine 演进之路（0）开发未动，测试先行》](http://debugtalk.com/post/ApiTestEngine-0-setup-CI-test/)
- [《ApiTestEngine 演进之路（1）搭建基础框架》](http://debugtalk.com/post/ApiTestEngine-1-setup-basic-framework/)
- [《ApiTestEngine 演进之路（2）探索优雅的测试用例描述方式》](http://debugtalk.com/post/ApiTestEngine-2-best-testcase-description/)
- [《ApiTestEngine 演进之路（3）测试用例中实现 Python 函数的定义》](http://debugtalk.com/post/ApiTestEngine-3-define-functions-in-yaml-testcases/)
- [《ApiTestEngine 演进之路（4）测试用例中实现 Python 函数的调用》](http://debugtalk.com/post/ApiTestEngine-4-call-functions-in-yaml-testcases/)
- [《ApiTestEngine 集成 Locust 实现更好的性能测试体验》](http://debugtalk.com/post/apitestengine-supersede-locust/)
- [《约定大于配置：ApiTestEngine实现热加载机制》](http://debugtalk.com/post/apitestengine-hot-plugin/)

[requests]: http://docs.python-requests.org/en/master/
[unittest]: https://docs.python.org/3/library/unittest.html
[Locust]: http://locust.io/
[flask]: http://flask.pocoo.org/
[PyUnitReport]: https://github.com/debugtalk/PyUnitReport
[Jenkins]: https://jenkins.io/index.html
[quickstart]: docs/quickstart.md