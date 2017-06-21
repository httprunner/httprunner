# ApiTestEngine

[![Build Status](https://travis-ci.org/debugtalk/ApiTestEngine.svg?branch=master)](https://travis-ci.org/debugtalk/ApiTestEngine)

## 核心特性

- 支持API接口的多种请求方法，包括 GET/POST/HEAD/PUT/DELETE 等
- 测试用例与代码分离，测试用例维护方式简洁优雅，支持`YAML`
- 测试用例描述方式具有表现力，可采用简洁的方式描述输入参数和预期输出结果
- 接口测试用例具有可复用性，便于创建复杂测试场景
- 测试执行方式简单灵活，支持单接口调用测试、批量接口调用测试、定时任务执行测试
- 测试结果统计报告简洁清晰，附带详尽日志记录，包括接口请求耗时、请求响应数据等
- 身兼多职，同时实现接口管理、接口自动化测试、接口性能测试（结合Locust）
- 具有可扩展性，便于扩展实现Web平台化

## Install

```bash
$ pip install -r requirements.txt
```

## Supported Python Versions

Python 2.7, 3.3, 3.4, 3.5, and 3.6.

## 阅读更多

- [《背景介绍》](docs/background.md)
- [《特性拆解介绍》](docs/features-intro.md)
- [《接口自动化测试的最佳工程实践（ApiTestEngine）》](http://debugtalk.com/post/ApiTestEngine-api-test-best-practice/)
