# ApiTestEngine

[![Build Status](https://travis-ci.org/debugtalk/ApiTestEngine.svg?branch=master)](https://travis-ci.org/debugtalk/ApiTestEngine)
[![Coverage Status](https://coveralls.io/repos/github/debugtalk/ApiTestEngine/badge.svg?branch=master)](https://coveralls.io/github/debugtalk/ApiTestEngine?branch=master)

## Key Features

- 支持API接口的多种请求方法，包括 GET/POST/HEAD/PUT/DELETE 等
- 测试用例与代码分离，测试用例维护方式简洁优雅，支持`YAML/JSON`
- 测试用例描述方式具有表现力，可采用简洁的方式描述输入参数和预期输出结果
- 接口测试用例具有可复用性，便于创建复杂测试场景
- 测试执行方式简单灵活，支持单接口调用测试、批量接口调用测试、定时任务执行测试
- 测试结果统计报告简洁清晰，附带详尽日志记录，包括接口请求耗时、请求响应数据等
- 身兼多职，同时实现接口管理、接口自动化测试、接口性能测试（结合Locust）
- 具有可扩展性，便于扩展实现Web平台化

[《背景介绍》](docs/background.md) [《特性拆解介绍》](docs/features-intro.md)

## Installation

```bash
$ pip install git+https://github.com/debugtalk/ApiTestEngine.git#egg=ApiTestEngine
```

If there is a problem with the installation, you can check the [`FAQ`](docs/FAQ.md).

To ensure the installation is successful, you can excuting command `ate -V` to see if you can get the version number.

```text
$ ate -V
0.1.0
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

## To Learn more ...

- [《接口自动化测试的最佳工程实践（ApiTestEngine）》](http://debugtalk.com/post/ApiTestEngine-api-test-best-practice/)
- [《ApiTestEngine 演进之路（0）开发未动，测试先行》](http://debugtalk.com/post/ApiTestEngine-0-setup-CI-test/)
- [《ApiTestEngine 演进之路（1）搭建基础框架》](http://debugtalk.com/post/ApiTestEngine-1-setup-basic-framework/)
- [《ApiTestEngine 演进之路（2）探索优雅的测试用例描述方式》](http://debugtalk.com/post/ApiTestEngine-2-best-testcase-description/)
- [《ApiTestEngine 演进之路（3）测试用例中实现 Python 函数的定义》](http://debugtalk.com/post/ApiTestEngine-3-define-functions-in-yaml-testcases/)
- [《ApiTestEngine 演进之路（4）测试用例中实现 Python 函数的调用》](http://debugtalk.com/post/ApiTestEngine-4-call-functions-in-yaml-testcases/)
