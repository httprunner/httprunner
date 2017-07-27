# ApiTestEngine

[![Build Status](https://travis-ci.org/debugtalk/ApiTestEngine.svg?branch=master)](https://travis-ci.org/debugtalk/ApiTestEngine)
[![Coverage Status](https://coveralls.io/repos/github/debugtalk/ApiTestEngine/badge.svg?branch=master)](https://coveralls.io/github/debugtalk/ApiTestEngine?branch=master)

## Design Philosophy

Take full reuse of Python's existing powerful libraries: [`Requests`][requests], [`unittest`][unittest] and [`Locust`][Locust]. And achieve the goal of API automation test, production environment monitoring, and API performance test, with a concise and  elegant manner.

## Key Features

- Inherit all powerful features of [`Requests`][requests], just have fun to handle HTTP in human way.
- Define testcases in YAML or JSON format in concise and elegant manner.
- Supports `function`/`variable`/`extract`/`validate` mechanisms to create full test scenarios.
- Testcases can be run in diverse ways, with single testset, multiple testsets, or whole project folder.
- Test report is concise and clear, with detailed log records. See [`PyUnitReport`][PyUnitReport].
- Perfect combination with [Jenkins][Jenkins], running continuous test and production monitoring.
- With reuse of [`Locust`][Locust], you can run performance test without extra work.
- It is extensible to facilitate the implementation of web platform with [`Flask`][flask] framework.

[*`Background Introduction (中文版)`*](docs/background-CN.md) | [*`Feature Descriptions (中文版)`*](docs/feature-descriptions-CN.md)

## Installation/Upgrade

```bash
$ pip install git+https://github.com/debugtalk/ApiTestEngine.git#egg=ApiTestEngine
```

To upgrade all specified packages to the newest available version, you should add the `-U` option.

```bash
$ pip install -U git+https://github.com/debugtalk/ApiTestEngine.git#egg=ApiTestEngine
```

If there is a problem with the installation or upgrade, you can check the [`FAQ`](docs/FAQ.md).

To ensure the installation or upgrade is successful, you can execute command `ate -V` to see if you can get the correct version number.

```text
$ ate -V
0.2.2
```

Execute the command `ate -h` to view command help.

```text
$ ate -h
usage: ate [-h] [-V] [--log-level LOG_LEVEL] [--report-name REPORT_NAME]
           [testset_paths [testset_paths ...]]

Api Test Engine.

positional arguments:
  testset_paths         testset file path

optional arguments:
  -h, --help            show this help message and exit
  -V, --version         show version
  --log-level LOG_LEVEL
                        Specify logging level, default is INFO.
  --report-name REPORT_NAME
                        Specify report name, default is generated time.
```

## Write testcases

It is recommended to write testcases in `YAML` format.

And here is testset example of typical scenario: get token at the beginning, and each subsequent requests should take the token in the headers.

```yaml
- config:
    name: "create user testsets."
    import_module_functions:
        - tests.data.custom_functions
    variable_binds:
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
    variable_binds:
        - sign: ${get_sign($user_agent, $device_sn, $os_platform, $app_version)}
    request:
        url: /api/get-token
        method: POST
        headers:
            user_agent: $user_agent
            device_sn: $device_sn
            os_platform: $os_platform
            app_version: $app_version
        json:
            sign: $sign
    extract_binds:
        - token: content.token
    validators:
        - {"check": "status_code", "comparator": "eq", "expected": 200}
        - {"check": "content.token", "comparator": "len_eq", "expected": 16}

- test:
    name: create user which does not exist
    variable_binds:
        - user_name: "user1"
        - user_password: "123456"
    request:
        url: /api/users/1000
        method: POST
        headers:
            token: $token
        json:
            name: $user_name
            password: $user_password
    validators:
        - {"check": "status_code", "comparator": "eq", "expected": 201}
        - {"check": "content.success", "comparator": "eq", "expected": true}
```

For detailed regulations of writing testcases, you can read the specification.

## Run testcases

`ApiTestEngine` can run testcases in diverse ways.

You can run single testset by specifying testset file path.

```text
$ ate filepath/testcase.yml
```

You can also run several testsets by specifying multiple testset file paths.

```text
$ ate filepath1/testcase1.yml filepath2/testcase2.yml
```

If you want to run testsets of a whole project, you can achieve this goal by specifying the project folder path.

```text
$ ate testcases_folder_path
```

## Supported Python Versions

Python `2.7`, `3.3`, `3.4`, `3.5`, and `3.6`.

## To learn more ...

- [《接口自动化测试的最佳工程实践（ApiTestEngine）》](http://debugtalk.com/post/ApiTestEngine-api-test-best-practice/)
- [《ApiTestEngine 演进之路（0）开发未动，测试先行》](http://debugtalk.com/post/ApiTestEngine-0-setup-CI-test/)
- [《ApiTestEngine 演进之路（1）搭建基础框架》](http://debugtalk.com/post/ApiTestEngine-1-setup-basic-framework/)
- [《ApiTestEngine 演进之路（2）探索优雅的测试用例描述方式》](http://debugtalk.com/post/ApiTestEngine-2-best-testcase-description/)
- [《ApiTestEngine 演进之路（3）测试用例中实现 Python 函数的定义》](http://debugtalk.com/post/ApiTestEngine-3-define-functions-in-yaml-testcases/)
- [《ApiTestEngine 演进之路（4）测试用例中实现 Python 函数的调用》](http://debugtalk.com/post/ApiTestEngine-4-call-functions-in-yaml-testcases/)


[requests]: http://docs.python-requests.org/en/master/
[unittest]: https://docs.python.org/3/library/unittest.html
[Locust]: http://locust.io/
[flask]: http://flask.pocoo.org/
[PyUnitReport]: https://github.com/debugtalk/PyUnitReport
[Jenkins]: https://jenkins.io/index.html