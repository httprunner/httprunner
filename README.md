# ApiTestEngine

[![Build Status](https://travis-ci.org/debugtalk/ApiTestEngine.svg?branch=master)](https://travis-ci.org/debugtalk/ApiTestEngine)
[![Coverage Status](https://coveralls.io/repos/github/debugtalk/ApiTestEngine/badge.svg?branch=master)](https://coveralls.io/github/debugtalk/ApiTestEngine?branch=master)

## 核心特性

- 支持API接口的多种请求方法，包括 GET/POST/HEAD/PUT/DELETE 等
- 测试用例与代码分离，测试用例维护方式简洁优雅，支持`YAML/JSON`
- 测试用例描述方式具有表现力，可采用简洁的方式描述输入参数和预期输出结果
- 接口测试用例具有可复用性，便于创建复杂测试场景
- 测试执行方式简单灵活，支持单接口调用测试、批量接口调用测试、定时任务执行测试
- 测试结果统计报告简洁清晰，附带详尽日志记录，包括接口请求耗时、请求响应数据等
- 身兼多职，同时实现接口管理、接口自动化测试、接口性能测试（结合Locust）
- 具有可扩展性，便于扩展实现Web平台化

[《背景介绍》](docs/background.md) [《特性拆解介绍》](docs/features-intro.md)

## Install

```bash
$ pip install -r requirements.txt
```

Run unittest to make sure everything is OK.

```bash
$ python -m unittest discover
```

## 编写测试用例

推荐采用`YAML`格式编写测试用例。

如下是一个典型的接口测试用例示例。具体的编写方式请阅读详细文档。

```python
- config:
    name: "create user testsets."
    requires:
        - random
        - string
        - hashlib
    function_binds:
        gen_random_string: "lambda str_len: ''.join(random.choice(string.ascii_letters + string.digits) for _ in range(str_len))"
        gen_md5: "lambda *str_args: hashlib.md5(''.join(str_args).encode('utf-8')).hexdigest()"
    variable_binds:
        - TOKEN: debugtalk
        - data: ""
        - random: ${gen_random_string(5)}
        - authorization: ${gen_md5($TOKEN, $data, $random)}
    request:
        base_url: http://127.0.0.1:5000

- test:
    name: create user which does not exist
    variable_binds:
        - data: '{"name": "user", "password": "123456"}'
    request:
        url: /api/users/1000
        method: POST
        headers:
            Content-Type: application/json
            authorization: $authorization
            random: $random
        data: $data
    validators:
        - {"check": "status_code", "comparator": "eq", "expected": 201}
        - {"check": "content.success", "comparator": "eq", "expected": true}

- test:
    name: create user which does exist
    variable_binds:
        - data: '{"name": "user", "password": "123456"}'
        - expected_status_code: 500
    request:
        url: /api/users/1000
        method: POST
        headers:
            Content-Type: application/json
            authorization: $authorization
            random: $random
        data: $data
    validators:
        - {"check": "status_code", "comparator": "eq", "expected": 500}
        - {"check": "content.success", "comparator": "eq", "expected": false}
```

## 运行测试用例

`ApiTestEngine`可指定运行特定的测试用例文件，或运行指定目录下的所有测试用例。

```bash
$ python main.py --testcase-path filepath/testcase.yml

$ python main.py --testcase-path testcases_folder_path
```

## Supported Python Versions

Python `2.7`, `3.3`, `3.4`, `3.5`, and `3.6`.

## 阅读更多

- [《接口自动化测试的最佳工程实践（ApiTestEngine）》](http://debugtalk.com/post/ApiTestEngine-api-test-best-practice/)
- [《ApiTestEngine 演进之路（0）开发未动，测试先行》](http://debugtalk.com/post/ApiTestEngine-0-setup-CI-test/)
- [《ApiTestEngine 演进之路（1）搭建基础框架》](http://debugtalk.com/post/ApiTestEngine-1-setup-basic-framework/)
- [《ApiTestEngine 演进之路（2）探索优雅的测试用例描述方式》](http://debugtalk.com/post/ApiTestEngine-2-best-testcase-description/)
- [《ApiTestEngine 演进之路（3）测试用例中实现 Python 函数的定义》](http://debugtalk.com/post/ApiTestEngine-3-define-functions-in-yaml-testcases/)
