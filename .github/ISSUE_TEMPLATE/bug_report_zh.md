---
name: Bug 反馈（中文）
about: 提交 bug 反馈
title: ''
labels: Pending
assignees: debugtalk
---

## 问题描述

请对遇到的 bug 进行简要描述。

## 版本信息

请提供如下版本信息：

 - 操作系统类型: [e.g. macos, Linux, Windows]
 - Python 版本 [e.g. 3.6]
 - HttpRunner 版本 [e.g. 2.1.2]
 - **设备 ID**: [e.g. 190070690681122]

获取方式：

在 Python 交互式 shell 中输入如下命令进行获取：

```bash
>>> import uuid; print(uuid.getnode())
190070690681122
```

## 项目文件内容（非必须）

如果可能，提供项目测试用例文件原始内容可加快 bug 定位和修复速度。

提示：请注意在去除项目敏感信息（IP、账号密码、密钥等）后再进行上传。

## 运行命令 && 堆栈信息

请提供在命令行中运行测试时所在的目录和命令，以及报错时的详细堆栈内容。

```bash
$ pwd
/Users/debugtalk/MyProjects/HttpRunner-dev/httprunner/tests
$ hrun testcases/setup.yml
INFO     Loading environment variables from /Users/debugtalk/MyProjects/HttpRunner-dev/HttpRunner/tests/.env
ERROR    !!!!!!!!!! exception stage: load tests !!!!!!!!!!
ModuleNotFoundError: No module named 'tests.api_server'
```
