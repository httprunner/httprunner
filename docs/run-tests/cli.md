
HttpRunner 在命令行中启动测试时，通过指定参数，可实现丰富的测试特性控制。

```text
$ hrun -h
usage: hrun [-h] [-V] [--log-level LOG_LEVEL] [--log-file LOG_FILE]
            [--dot-env-path DOT_ENV_PATH] [--report-template REPORT_TEMPLATE]
            [--report-dir REPORT_DIR] [--failfast] [--save-tests]
            [--startproject STARTPROJECT]
            [--validate [VALIDATE [VALIDATE ...]]]
            [--prettify [PRETTIFY [PRETTIFY ...]]]
            [testcase_paths [testcase_paths ...]]

One-stop solution for HTTP(S) testing.

positional arguments:
  testcase_paths        testcase file path

optional arguments:
  -h, --help            show this help message and exit
  -V, --version         show version
  --log-level LOG_LEVEL
                        Specify logging level, default is INFO.
  --log-file LOG_FILE   Write logs to specified file path.
  --dot-env-path DOT_ENV_PATH
                        Specify .env file path, which is useful for keeping
                        sensitive data.
  --report-template REPORT_TEMPLATE
                        specify report template path.
  --report-dir REPORT_DIR
                        specify report save directory.
  --failfast            Stop the test run on the first error or failure.
  --save-tests          Save loaded tests and parsed tests to JSON file.
  --startproject STARTPROJECT
                        Specify new project name.
  --validate [VALIDATE [VALIDATE ...]]
                        Validate JSON testcase format.
  --prettify [PRETTIFY [PRETTIFY ...]]
                        Prettify JSON testcase format.
```

## 指定测试用例路径

使用 HttpRunner 指定测试用例路径时，支持多种方式。

使用 hrun 命令外加单个测试用例文件的路径，运行单个测试用例，并生成一个测试报告文件：

```text
$ hrun filepath/testcase.yml
```

将多个测试用例文件放置到文件夹中，指定文件夹路径可将文件夹下所有测试用例作为测试用例集进行运行，并生成一个测试报告文件：

```text
$ hrun testcases_folder_path
```

## failfast

默认情况下，HttpRunner 会运行指定用例集中的所有测试用例，并统计测试结果。

> 对于某些依赖于执行顺序的测试用例，例如需要先登录成功才能执行后续接口请求的场景，当前面的测试用例执行失败后，后续的测试用例也都必将失败，因此没有继续执行的必要了。

若希望测试用例在运行过程中，遇到失败时不再继续运行后续用例，则可通过在命令中添加`--failfast`实现。

```text
$ hrun filepath/testcase.yml --failfast
```

## 日志级别

默认情况下，HttpRunner 运行时的日志级别为`INFO`，只会包含最基本的信息，包括用例名称、请求的URL和Method、响应结果的状态码、耗时和内容大小。

```text
$ hrun docs/data/demo-quickstart-6.json
INFO     Start to run testcase: testcase description
/api/get-token
INFO     POST http://127.0.0.1:5000/api/get-token
INFO     status_code: 200, response_time(ms): 9.08 ms, response_length: 46 bytes

.
/api/users/1548560655759
INFO     POST http://127.0.0.1:5000/api/users/1548560655759
INFO     status_code: 201, response_time(ms): 2.89 ms, response_length: 54 bytes

.

----------------------------------------------------------------------
Ran 2 tests in 0.019s

OK
INFO     Start to render Html report ...
INFO     Generated Html report: /Users/debugtalk/MyProjects/HttpRunner-dev/httprunner-docs-v2x/reports/1548560655.html
```

若需要查看到更详尽的信息，例如请求的参数和响应的详细内容，可以将日志级别设置为`DEBUG`，即在命令中添加`--log-level debug`。

```
$ hrun docs/data/demo-quickstart-6.json --log-level debug
INFO     Start to run testcase: testcase description
/api/get-token
INFO     POST http://127.0.0.1:5000/api/get-token
DEBUG    request kwargs(raw): {'headers': {'User-Agent': 'python-requests/2.18.4', 'device_sn': 'W5ACRDytKRQJPhC', 'os_platform': 'ios', 'app_version': '2.8.6', 'Content-Type': 'application/json'}, 'json': {'sign': '2e7c3b5d560a1c8a859edcb9c8b0d3f8349abeff'}, 'verify': True}
DEBUG    processed request:
> POST http://127.0.0.1:5000/api/get-token
> kwargs: {'headers': {'User-Agent': 'python-requests/2.18.4', 'device_sn': 'W5ACRDytKRQJPhC', 'os_platform': 'ios', 'app_version': '2.8.6', 'Content-Type': 'application/json'}, 'json': {'sign': '2e7c3b5d560a1c8a859edcb9c8b0d3f8349abeff'}, 'verify': True, 'timeout': 120}
DEBUG
================== request details ==================
url              : 'http://127.0.0.1:5000/api/get-token'
method           : 'POST'
headers          : {'User-Agent': 'python-requests/2.18.4', 'Accept-Encoding': 'gzip, deflate', 'Accept': '*/*', 'Connection': 'keep-alive', 'device_sn': 'W5ACRDytKRQJPhC', 'os_platform': 'ios', 'app_version': '2.8.6', 'Content-Type': 'application/json', 'Content-Length': '52'}
body             : b'{"sign": "2e7c3b5d560a1c8a859edcb9c8b0d3f8349abeff"}'

DEBUG
================== response details ==================
ok               : True
url              : 'http://127.0.0.1:5000/api/get-token'
status_code      : 200
reason           : 'OK'
cookies          : {}
encoding         : None
headers          : {'Content-Type': 'application/json', 'Content-Length': '46', 'Server': 'Werkzeug/0.14.1 Python/3.6.5+', 'Date': 'Sun, 27 Jan 2019 03:45:16 GMT'}
content_type     : 'application/json'
json             : {'success': True, 'token': 'o6uakmubLrCbpRRS'}

INFO     status_code: 200, response_time(ms): 9.28 ms, response_length: 46 bytes

DEBUG    start to extract from response object.
DEBUG    extract: content.token	=> o6uakmubLrCbpRRS
DEBUG    start to validate.
DEBUG    extract: status_code	=> 200
DEBUG    validate: status_code equals 200(int)	==> pass
DEBUG    extract: headers.Content-Type	=> application/json
DEBUG    validate: headers.Content-Type equals application/json(str)	==> pass
DEBUG    extract: content.success	=> True
DEBUG    validate: content.success equals True(bool)	==> pass
.
/api/users/1548560716736
INFO     POST http://127.0.0.1:5000/api/users/1548560716736
DEBUG    request kwargs(raw): {'headers': {'User-Agent': 'python-requests/2.18.4', 'device_sn': 'W5ACRDytKRQJPhC', 'token': 'o6uakmubLrCbpRRS', 'Content-Type': 'application/json'}, 'json': {'name': 'user1', 'password': '123456'}, 'verify': True}
DEBUG    processed request:
> POST http://127.0.0.1:5000/api/users/1548560716736
> kwargs: {'headers': {'User-Agent': 'python-requests/2.18.4', 'device_sn': 'W5ACRDytKRQJPhC', 'token': 'o6uakmubLrCbpRRS', 'Content-Type': 'application/json'}, 'json': {'name': 'user1', 'password': '123456'}, 'verify': True, 'timeout': 120}
DEBUG
================== request details ==================
url              : 'http://127.0.0.1:5000/api/users/1548560716736'
method           : 'POST'
headers          : {'User-Agent': 'python-requests/2.18.4', 'Accept-Encoding': 'gzip, deflate', 'Accept': '*/*', 'Connection': 'keep-alive', 'device_sn': 'W5ACRDytKRQJPhC', 'token': 'o6uakmubLrCbpRRS', 'Content-Type': 'application/json', 'Content-Length': '39'}
body             : b'{"name": "user1", "password": "123456"}'

DEBUG
================== response details ==================
ok               : True
url              : 'http://127.0.0.1:5000/api/users/1548560716736'
status_code      : 201
reason           : 'CREATED'
cookies          : {}
encoding         : None
headers          : {'Content-Type': 'application/json', 'Content-Length': '54', 'Server': 'Werkzeug/0.14.1 Python/3.6.5+', 'Date': 'Sun, 27 Jan 2019 03:45:16 GMT'}
content_type     : 'application/json'
json             : {'success': True, 'msg': 'user created successfully.'}

INFO     status_code: 201, response_time(ms): 2.77 ms, response_length: 54 bytes

DEBUG    start to validate.
DEBUG    extract: status_code	=> 201
DEBUG    validate: status_code equals 201(int)	==> pass
DEBUG    extract: headers.Content-Type	=> application/json
DEBUG    validate: headers.Content-Type equals application/json(str)	==> pass
DEBUG    extract: content.success	=> True
DEBUG    validate: content.success equals True(bool)	==> pass
DEBUG    extract: content.msg	=> user created successfully.
DEBUG    validate: content.msg equals user created successfully.(str)	==> pass
.

----------------------------------------------------------------------
Ran 2 tests in 0.022s

OK
DEBUG    No html report template specified, use default.
INFO     Start to render Html report ...
INFO     Generated Html report: /Users/debugtalk/MyProjects/HttpRunner-dev/httprunner-docs-v2x/reports/1548560716.html
```

## 保存详细过程数据

为了方便定位问题，运行测试时可指定 `--save-tests` 参数，即可将运行过程的中间数据保存为日志文件。

```text
$ hrun docs/data/demo-quickstart-6.json --save-tests
dump file: /Users/debugtalk/MyProjects/HttpRunner-dev/httprunner-docs-v2x/docs/data/logs/demo-quickstart-6.loaded.json
dump file: /Users/debugtalk/MyProjects/HttpRunner-dev/httprunner-docs-v2x/docs/data/logs/demo-quickstart-6.parsed.json
INFO     Start to run testcase: testcase description
/api/get-token
INFO     POST http://127.0.0.1:5000/api/get-token
INFO     status_code: 200, response_time(ms): 11.42 ms, response_length: 46 bytes

.
/api/users/1548560768589
INFO     POST http://127.0.0.1:5000/api/users/1548560768589
INFO     status_code: 201, response_time(ms): 2.8 ms, response_length: 54 bytes

.

----------------------------------------------------------------------
Ran 2 tests in 0.028s

OK
dump file: /Users/debugtalk/MyProjects/HttpRunner-dev/httprunner-docs-v2x/docs/data/logs/demo-quickstart-6.summary.json
INFO     Start to render Html report ...
INFO     Generated Html report: /Users/debugtalk/MyProjects/HttpRunner-dev/httprunner-docs-v2x/reports/1548560768.html
```

日志文件将保存在项目根目录的 `logs` 文件夹中，生成的文件有如下三个（XXX为测试用例名称）：

- `XXX.loaded.json`：测试用例加载后的数据结构内容，加载包括测试用例文件（YAML/JSON）、debugtalk.py、.env 等所有项目文件，例如 [`demo-quickstart-6.loaded.json`](/data/logs/demo-quickstart-6.loaded.json)
- `XXX.parsed.json`：测试用例解析后的数据结构内容，解析内容包括测试用例引用（API/testcase）、变量计算和替换、base_url 拼接等，例如 [`demo-quickstart-6.parsed.json`](/data/logs/demo-quickstart-6.parsed.json)
- `XXX.summary.json`：测试报告生成前的数据结构内容，例如 [`demo-quickstart-6.summary.json`](/data/logs/demo-quickstart-6.summary.json)
