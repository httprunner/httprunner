本文将通过一个简单的示例来展示 HttpRunner 的核心功能使用方法。

## 案例介绍

该案例作为被测服务，主要有两类接口：

- 权限校验，获取 token
- 支持 CRUD 操作的 RESTful APIs，所有接口的请求头域中都必须包含有效的 token

案例的实现形式为 flask 应用服务（[api_server.py](data/api_server.py)），启动方式如下：

```text
$ export FLASK_APP=docs/data/api_server.py
$ export FLASK_ENV=development
$ flask run
 * Serving Flask app "docs/data/api_server.py" (lazy loading)
 * Environment: development
 * Debug mode: on
 * Running on http://127.0.0.1:5000/ (Press CTRL+C to quit)
 * Restarting with stat
 * Debugger is active!
 * Debugger PIN: 989-476-348
```

服务启动成功后，我们就可以开始对其进行测试了。

## 测试准备

### 抓包分析

在开始测试之前，我们需要先了解接口的请求和响应细节，而最佳的方式就是采用 `Charles Proxy` 或者 `Fiddler` 这类网络抓包工具进行抓包分析。

例如，在本案例中，我们先进行权限校验，然后成功创建一个用户，对应的网络抓包内容如下图所示：

![](./images/demo-quickstart-http-1.jpg)

![](./images/demo-quickstart-http-2.jpg)

通过抓包，我们可以看到具体的接口信息，包括请求的URL、Method、headers、参数和响应内容等内容，基于这些信息，我们就可以开始编写测试用例了。

### 生成测试用例

为了简化测试用例的编写工作，HttpRunner 实现了测试用例生成的功能。

首先，需要将抓取得到的数据包导出为 HAR 格式的文件，假设导出的文件名称为 [demo-quickstart.har](data/demo-quickstart.har)。

然后，在命令行终端中运行如下命令，即可将 demo-quickstart.har 转换为 HttpRunner 的测试用例文件。

```bash
$ har2case docs/data/demo-quickstart.har -2y
INFO:root:Start to generate testcase.
INFO:root:dump testcase to YAML format.
INFO:root:Generate YAML testcase successfully: docs/data/demo-quickstart.yml
```

使用 `har2case` 转换脚本时默认转换为 JSON 格式，加上 `-2y` 参数后转换为 YAML 格式。两种格式完全等价，YAML 格式更简洁，JSON 格式支持的工具更丰富，大家可根据个人喜好进行选择。关于 [har2case][har2case] 的详细使用说明，请查看[《录制生成测试用例》](/prepare/record/)。

经过转换，在源 demo-quickstart.har 文件的同级目录下生成了相同文件名称的 YAML 格式测试用例文件 [demo-quickstart.yml](data/demo-quickstart.yml)，其内容如下：

```yaml
- config:
    name: testcase description
    variables: {}

- test:
    name: /api/get-token
    request:
        headers:
            Content-Type: application/json
            User-Agent: python-requests/2.18.4
            app_version: 2.8.6
            device_sn: FwgRiO7CNA50DSU
            os_platform: ios
        json:
            sign: 9c0c7e51c91ae963c833a4ccbab8d683c4a90c98
        method: POST
        url: http://127.0.0.1:5000/api/get-token
    validate:
        - eq: [status_code, 200]
        - eq: [headers.Content-Type, application/json]
        - eq: [content.success, true]
        - eq: [content.token, baNLX1zhFYP11Seb]

- test:
    name: /api/users/1000
    request:
        headers:
            Content-Type: application/json
            User-Agent: python-requests/2.18.4
            device_sn: FwgRiO7CNA50DSU
            token: baNLX1zhFYP11Seb
        json:
            name: user1
            password: '123456'
        method: POST
        url: http://127.0.0.1:5000/api/users/1000
    validate:
        - eq: [status_code, 201]
        - eq: [headers.Content-Type, application/json]
        - eq: [content.success, true]
        - eq: [content.msg, user created successfully.]
```

现在我们只需要知道如下几点：

- 每个 YAML/JSON 文件对应一个测试用例（testcase）
- 每个测试用例为一个`list of dict`结构，其中可能包含全局配置项（config）和若干个测试步骤（test）
- `config` 为全局配置项，作用域为整个测试用例
- `test` 对应单个测试步骤，作用域仅限于本身

如上便是 HttpRunner 测试用例的基本结构。

关于测试用例的更多内容，请查看[《测试用例结构描述》](/prepare/testcase-structure/)。

### 首次运行测试用例

测试用例就绪后，我们可以开始调试运行了。

为了演示测试用例文件的迭代优化过程，我们先将 demo-quickstart.json 重命名为 [demo-quickstart-0.json](data/demo-quickstart-0.json)（对应的 YAML 格式：[demo-quickstart-0.yml](data/demo-quickstart-0.yml)）。

运行测试用例的命令为`hrun`，后面直接指定测试用例文件的路径即可。

```text
$ hrun docs/data/demo-quickstart-0.yml
INFO     Start to run testcase: testcase description
/api/get-token
INFO     POST http://127.0.0.1:5000/api/get-token
INFO     status_code: 200, response_time(ms): 9.26 ms, response_length: 46 bytes

ERROR    validate: content.token equals baNLX1zhFYP11Seb(str)	==> fail
tXGuSQgOCVXcltkz(str) equals baNLX1zhFYP11Seb(str)
ERROR    ******************************** DETAILED REQUEST & RESPONSE ********************************
====== request details ======
url: http://127.0.0.1:5000/api/get-token
method: POST
headers: {'Content-Type': 'application/json', 'User-Agent': 'python-requests/2.18.4', 'app_version': '2.8.6', 'device_sn': 'FwgRiO7CNA50DSU', 'os_platform': 'ios'}
json: {'sign': '9c0c7e51c91ae963c833a4ccbab8d683c4a90c98'}
verify: True

====== response details ======
status_code: 200
headers: {'Content-Type': 'application/json', 'Content-Length': '46', 'Server': 'Werkzeug/0.14.1 Python/3.7.0', 'Date': 'Sat, 26 Jan 2019 14:43:55 GMT'}
body: '{"success": true, "token": "tXGuSQgOCVXcltkz"}'

F
/api/users/1000
INFO     POST http://127.0.0.1:5000/api/users/1000
ERROR    403 Client Error: FORBIDDEN for url: http://127.0.0.1:5000/api/users/1000
ERROR    validate: status_code equals 201(int)	==> fail
403(int) equals 201(int)
ERROR    validate: content.success equals True(bool)	==> fail
False(bool) equals True(bool)
ERROR    validate: content.msg equals user created successfully.(str)	==> fail
Authorization failed!(str) equals user created successfully.(str)
ERROR    ******************************** DETAILED REQUEST & RESPONSE ********************************
====== request details ======
url: http://127.0.0.1:5000/api/users/1000
method: POST
headers: {'Content-Type': 'application/json', 'User-Agent': 'python-requests/2.18.4', 'device_sn': 'FwgRiO7CNA50DSU', 'token': 'baNLX1zhFYP11Seb'}
json: {'name': 'user1', 'password': '123456'}
verify: True

====== response details ======
status_code: 403
headers: {'Content-Type': 'application/json', 'Content-Length': '50', 'Server': 'Werkzeug/0.14.1 Python/3.7.0', 'Date': 'Sat, 26 Jan 2019 14:43:55 GMT'}
body: '{"success": false, "msg": "Authorization failed!"}'

F

======================================================================
FAIL: test_0000_000 (httprunner.api.TestSequense)
/api/get-token
----------------------------------------------------------------------
Traceback (most recent call last):
  File "/Users/debugtalk/.pyenv/versions/3.6-dev/lib/python3.6/site-packages/httprunner/api.py", line 54, in test
    test_runner.run_test(test_dict)
httprunner.exceptions.ValidationFailure: validate: content.token equals baNLX1zhFYP11Seb(str)	==> fail
tXGuSQgOCVXcltkz(str) equals baNLX1zhFYP11Seb(str)

During handling of the above exception, another exception occurred:

Traceback (most recent call last):
  File "/Users/debugtalk/.pyenv/versions/3.6-dev/lib/python3.6/site-packages/httprunner/api.py", line 56, in test
    self.fail(str(ex))
AssertionError: validate: content.token equals baNLX1zhFYP11Seb(str)	==> fail
tXGuSQgOCVXcltkz(str) equals baNLX1zhFYP11Seb(str)

======================================================================
FAIL: test_0001_000 (httprunner.api.TestSequense)
/api/users/1000
----------------------------------------------------------------------
Traceback (most recent call last):
  File "/Users/debugtalk/.pyenv/versions/3.6-dev/lib/python3.6/site-packages/httprunner/api.py", line 54, in test
    test_runner.run_test(test_dict)
httprunner.exceptions.ValidationFailure: validate: status_code equals 201(int)	==> fail
403(int) equals 201(int)
validate: content.success equals True(bool)	==> fail
False(bool) equals True(bool)
validate: content.msg equals user created successfully.(str)	==> fail
Authorization failed!(str) equals user created successfully.(str)

During handling of the above exception, another exception occurred:

Traceback (most recent call last):
  File "/Users/debugtalk/.pyenv/versions/3.6-dev/lib/python3.6/site-packages/httprunner/api.py", line 56, in test
    self.fail(str(ex))
AssertionError: validate: status_code equals 201(int)	==> fail
403(int) equals 201(int)
validate: content.success equals True(bool)	==> fail
False(bool) equals True(bool)
validate: content.msg equals user created successfully.(str)	==> fail
Authorization failed!(str) equals user created successfully.(str)

----------------------------------------------------------------------
Ran 2 tests in 0.026s

FAILED (failures=2)
INFO     Start to render Html report ...
INFO     Generated Html report: /Users/debugtalk/MyProjects/HttpRunner-dev/httprunner-docs-v2x/reports/1548513835.html
```

非常不幸，两个接口的测试用例均运行失败了。

## 优化测试用例

从两个测试步骤的报错信息和堆栈信息（Traceback）可以看出，第一个步骤失败的原因是获取的 token 与预期值不一致，第二个步骤失败的原因是请求权限校验失败（403）。

接下来我们将逐步进行进行优化。

### 调整校验器

默认情况下，[har2case][har2case] 生成用例时，若 HTTP 请求的响应内容为 JSON 格式，则会将第一层级中的所有`key-value`转换为 validator。

例如上面的第一个测试步骤，生成的 validator 为：

```json
"validate": [
    {"eq": ["status_code", 200]},
    {"eq": ["headers.Content-Type", "application/json"]},
    {"eq": ["content.success", true]},
    {"eq": ["content.token", "baNLX1zhFYP11Seb"]}
]
```

运行测试用例时，就会对上面的各个项进行校验。

问题在于，请求`/api/get-token`接口时，每次生成的 token 都会是不同的，因此将生成的 token 作为校验项的话，校验自然就无法通过了。

正确的做法是，在测试步骤的 validate 中应该去掉这类动态变化的值。

去除该项后，将用例另存为 [demo-quickstart-1.json](data/demo-quickstart-1.json)（对应的 YAML 格式：[demo-quickstart-1.yml](data/demo-quickstart-1.yml)）。

再次运行测试用例，运行结果如下：

```text
$ hrun docs/data/demo-quickstart-1.yml
INFO     Start to run testcase: testcase description
/api/get-token
INFO     POST http://127.0.0.1:5000/api/get-token
INFO     status_code: 200, response_time(ms): 6.61 ms, response_length: 46 bytes

.
/api/users/1000
INFO     POST http://127.0.0.1:5000/api/users/1000
ERROR    403 Client Error: FORBIDDEN for url: http://127.0.0.1:5000/api/users/1000
ERROR    validate: status_code equals 201(int)	==> fail
403(int) equals 201(int)
ERROR    validate: content.success equals True(bool)	==> fail
False(bool) equals True(bool)
ERROR    validate: content.msg equals user created successfully.(str)	==> fail
Authorization failed!(str) equals user created successfully.(str)
ERROR    ******************************** DETAILED REQUEST & RESPONSE ********************************
====== request details ======
url: http://127.0.0.1:5000/api/users/1000
method: POST
headers: {'Content-Type': 'application/json', 'User-Agent': 'python-requests/2.18.4', 'device_sn': 'FwgRiO7CNA50DSU', 'token': 'baNLX1zhFYP11Seb'}
json: {'name': 'user1', 'password': '123456'}
verify: True

====== response details ======
status_code: 403
headers: {'Content-Type': 'application/json', 'Content-Length': '50', 'Server': 'Werkzeug/0.14.1 Python/3.7.0', 'Date': 'Sat, 26 Jan 2019 14:45:34 GMT'}
body: '{"success": false, "msg": "Authorization failed!"}'

F

======================================================================
FAIL: test_0001_000 (httprunner.api.TestSequense)
/api/users/1000
----------------------------------------------------------------------
Traceback (most recent call last):
  File "/Users/debugtalk/.pyenv/versions/3.6-dev/lib/python3.6/site-packages/httprunner/api.py", line 54, in test
    test_runner.run_test(test_dict)
httprunner.exceptions.ValidationFailure: validate: status_code equals 201(int)	==> fail
403(int) equals 201(int)
validate: content.success equals True(bool)	==> fail
False(bool) equals True(bool)
validate: content.msg equals user created successfully.(str)	==> fail
Authorization failed!(str) equals user created successfully.(str)

During handling of the above exception, another exception occurred:

Traceback (most recent call last):
  File "/Users/debugtalk/.pyenv/versions/3.6-dev/lib/python3.6/site-packages/httprunner/api.py", line 56, in test
    self.fail(str(ex))
AssertionError: validate: status_code equals 201(int)	==> fail
403(int) equals 201(int)
validate: content.success equals True(bool)	==> fail
False(bool) equals True(bool)
validate: content.msg equals user created successfully.(str)	==> fail
Authorization failed!(str) equals user created successfully.(str)

----------------------------------------------------------------------
Ran 2 tests in 0.018s

FAILED (failures=1)
INFO     Start to render Html report ...
INFO     Generated Html report: /Users/debugtalk/MyProjects/HttpRunner-dev/httprunner-docs-v2x/reports/1548513934.html
```

经过修改，第一个测试步骤已经运行成功了，第二个步骤仍然运行失败（403），还是因为权限校验的原因。

### 参数关联

我们继续查看 [demo-quickstart-1.json](data/demo-quickstart-1.json)，会发现第二个测试步骤的请求 headers 中的 token 仍然是硬编码的，即抓包时获取到的值。在我们再次运行测试用例时，这个 token 已经失效了，所以会出现 403 权限校验失败的问题。

正确的做法是，我们应该在每次运行测试用例的时候，先动态获取到第一个测试步骤中的 token，然后在后续测试步骤的请求中使用前面获取到的 token。

在 HttpRunner 中，支持参数提取（`extract`）和参数引用的功能（`$var`）。

在测试步骤（test）中，若需要从响应结果中提取参数，则可使用 `extract` 关键字。extract 的列表中可指定一个或多个需要提取的参数。

在提取参数时，当 HTTP 的请求响应结果为 JSON 格式，则可以采用`.`运算符的方式，逐级往下获取到参数值；响应结果的整体内容引用方式为 content 或者 body。

例如，第一个接口`/api/get-token`的响应结果为：

```json
{"success": true, "token": "ZQkYhbaQ6q8UFFNE"}
```

那么要获取到 token 参数，就可以使用 content.token 的方式；具体的写法如下：

```json
"extract": [
  {"token": "content.token"}
]
```

其中，token 作为提取后的参数名称，可以在后续使用 `$token` 进行引用。

```json
"headers": {
  "device_sn": "FwgRiO7CNA50DSU",
  "token": "$token",
  "Content-Type": "application/json"
}
```

修改后的测试用例另存为 [demo-quickstart-2.json](data/demo-quickstart-2.json)（对应的 YAML 格式：[demo-quickstart-2.yml](data/demo-quickstart-2.yml)）。

再次运行测试用例，运行结果如下：

```text
$ hrun docs/data/demo-quickstart-2.yml
INFO     Start to run testcase: testcase description
/api/get-token
INFO     POST http://127.0.0.1:5000/api/get-token
INFO     status_code: 200, response_time(ms): 8.32 ms, response_length: 46 bytes

.
/api/users/1000
INFO     POST http://127.0.0.1:5000/api/users/1000
INFO     status_code: 201, response_time(ms): 3.02 ms, response_length: 54 bytes

.

----------------------------------------------------------------------
Ran 2 tests in 0.019s

OK
INFO     Start to render Html report ...
INFO     Generated Html report: /Users/debugtalk/MyProjects/HttpRunner-dev/httprunner-docs-v2x/reports/1548514191.html
```

经过修改，第二个测试步骤也运行成功了。

### base_url

虽然测试步骤运行都成功了，但是仍然有继续优化的地方。

继续查看 [demo-quickstart-2.json](data/demo-quickstart-2.json)，我们会发现在每个测试步骤的 URL 中，都采用的是完整的描述（host+path），但大多数情况下同一个用例中的 host 都是相同的，区别仅在于 path 部分。

因此，我们可以将各个测试步骤（test） URL 的 `base_url` 抽取出来，放到全局配置模块（config）中，在测试步骤中的 URL 只保留 PATH 部分。

```yaml
- config:
    name: testcase description
    base_url: http://127.0.0.1:5000

- test:
    name: get token
    request:
        url: /api/get-token
```

调整后的测试用例另存为 [demo-quickstart-3.json](data/demo-quickstart-3.json)（对应的 YAML 格式：[demo-quickstart-3.yml](data/demo-quickstart-3.yml)）。

重启 flask 应用服务后再次运行测试用例，所有的测试步骤仍然运行成功。

### 变量的申明和引用

继续查看 [demo-quickstart-3.json](data/demo-quickstart-3.json)，我们会发现测试用例中存在较多硬编码的参数，例如 app_version、device_sn、os_platform、user_id 等。

大多数情况下，我们可以不用修改这些硬编码的参数，测试用例也能正常运行。但是为了更好地维护测试用例，例如同一个参数值在测试步骤中出现多次，那么比较好的做法是，将这些参数定义为变量，然后在需要参数的地方进行引用。

在 HttpRunner 中，支持变量申明（`variables`）和引用（`$var`）的机制。在 config 和 test 中均可以通过 `variables` 关键字定义变量，然后在测试步骤中可以通过 `$ + 变量名称` 的方式引用变量。区别在于，在 config 中定义的变量为全局的，整个测试用例（testcase）的所有地方均可以引用；在 test 中定义的变量作用域仅局限于当前测试步骤（teststep）。

对上述各个测试步骤中硬编码的参数进行变量申明和引用调整后，新的测试用例另存为 [demo-quickstart-4.json](data/demo-quickstart-4.json)（对应的 YAML 格式：[demo-quickstart-4.yml](data/demo-quickstart-4.yml)）。

重启 flask 应用服务后再次运行测试用例，所有的测试步骤仍然运行成功。

### 抽取公共变量

查看 [demo-quickstart-4.json](data/demo-quickstart-4.json) 可以看出，两个测试步骤中都定义了 device_sn。针对这类公共的参数，我们可以将其统一定义在 config 的 variables 中，在测试步骤中就不用再重复定义。

```yaml
- config:
    name: testcase description
    base_url: http://127.0.0.1:5000
    variables:
        device_sn: FwgRiO7CNA50DSU
```

调整后的测试用例见 [demo-quickstart-5.json](data/demo-quickstart-5.json)（对应的 YAML 格式：[demo-quickstart-5.yml](data/demo-quickstart-5.yml)）。

### 实现动态运算逻辑

在 [demo-quickstart-5.yml](data/demo-quickstart-5.yml) 中，参数 device_sn 代表的是设备的 SN 编码，虽然采用硬编码的方式暂时不影响测试用例的运行，但这与真实的用户场景不大相符。

假设 device_sn 的格式为 15 长度的字符串，那么我们就可以在每次运行测试用例的时候，针对 device_sn 生成一个 15 位长度的随机字符串。与此同时，sign 字段是根据 headers 中的各个字段拼接后生成得到的 MD5 值，因此在 device_sn 变动后，sign 也应该重新进行计算，否则就会再次出现签名校验失败的问题。

然而，HttpRunner 的测试用例都是采用 YAML/JSON 格式进行描述的，在文本格式中如何执行代码运算呢？

HttpRunner 的实现方式为，支持热加载的插件机制（`debugtalk.py`），可以在 YAML/JSON 中调用 Python 函数。

具体地做法，我们可以在测试用例文件的同级或其父级目录中创建一个 debugtalk.py 文件，然后在其中定义相关的函数和变量。

例如，针对 device_sn 的随机字符串生成功能，我们可以定义一个 gen_random_string 函数；针对 sign 的签名算法，我们可以定义一个 get_sign 函数。

```python
import hashlib
import hmac
import random
import string

SECRET_KEY = "DebugTalk"

def gen_random_string(str_len):
    random_char_list = []
    for _ in range(str_len):
        random_char = random.choice(string.ascii_letters + string.digits)
        random_char_list.append(random_char)

    random_string = ''.join(random_char_list)
    return random_string

def get_sign(*args):
    content = ''.join(args).encode('ascii')
    sign_key = SECRET_KEY.encode('ascii')
    sign = hmac.new(sign_key, content, hashlib.sha1).hexdigest()
    return sign
```

然后，我们在 YAML/JSON 测试用例文件中，就可以对定义的函数进行调用，对定义的变量进行引用了。引用变量的方式仍然与前面讲的一样，采用`$ + 变量名称`的方式；调用函数的方式为`${func($var)}`。

例如，生成 15 位长度的随机字符串并赋值给 device_sn 的代码为：

```json
"variables": [
  {"device_sn": "${gen_random_string(15)}"}
]
```

使用 $user_agent、$device_sn、$os_platform、$app_version 根据签名算法生成 sign 值的代码为：

```json
"json": {
  "sign": "${get_sign($user_agent, $device_sn, $os_platform, $app_version)}"
}
```

对测试用例进行上述调整后，另存为 [demo-quickstart-6.json](data/demo-quickstart-6.json)（对应的 YAML 格式：[demo-quickstart-6.yml](data/demo-quickstart-6.yml)）。

重启 flask 应用服务后再次运行测试用例，所有的测试步骤仍然运行成功。

### 参数化数据驱动

> 请确保你使用的 HttpRunner 版本号不低于 2.0.0

在 [demo-quickstart-6.yml](data/demo-quickstart-6.yml) 中，user_id 仍然是写死的值，假如我们需要创建 user_id 为 1001～1004 的用户，那我们只能不断地去修改 user_id，然后运行测试用例，重复操作 4 次？或者我们在测试用例文件中将创建用户的 test 复制 4 份，然后在每一份里面分别使用不同的 user_id ？

很显然，不管是采用上述哪种方式，都会很繁琐，并且也无法应对灵活多变的测试需求。

针对这类需求，HttpRunner 支持参数化数据驱动的功能。

在 HttpRunner 中，若要采用数据驱动的方式来运行测试用例，需要创建一个文件，对测试用例进行引用，并使用 `parameters` 关键字定义参数并指定数据源取值方式。

例如，我们需要在创建用户的接口中对 user_id 进行参数化，参数化列表为 1001～1004，并且取值方式为顺序取值，那么最简单的描述方式就是直接指定参数列表。具体的编写方式为，新建一个测试场景文件 [demo-quickstart-7.yml](data/demo-quickstart-7.yml)（对应的 JSON 格式：[demo-quickstart-7.json](data/demo-quickstart-7.json)），内容如下所示：

```yaml
config:
    name: testcase description

testcases:
    create user:
        testcase: demo-quickstart-6.yml
        parameters:
            user_id: [1001, 1002, 1003, 1004]
```

仅需如上配置，针对 user_id 的参数化数据驱动就完成了。

重启 flask 应用服务后再次运行测试用例，测试用例运行情况如下所示：

<details>
<summary>点击查看运行日志</summary>

```text
$ hrun docs/data/demo-quickstart-7.json
INFO     Start to run testcase: create user 1001
/api/get-token
INFO     POST http://127.0.0.1:5000/api/get-token
INFO     status_code: 200, response_time(ms): 8.95 ms, response_length: 46 bytes

.
/api/users/1001
INFO     POST http://127.0.0.1:5000/api/users/1001
INFO     status_code: 201, response_time(ms): 3.02 ms, response_length: 54 bytes

.

----------------------------------------------------------------------
Ran 2 tests in 0.021s

OK
INFO     Start to run testcase: create user 1002
/api/get-token
INFO     POST http://127.0.0.1:5000/api/get-token
INFO     status_code: 200, response_time(ms): 2.78 ms, response_length: 46 bytes

.
/api/users/1002
INFO     POST http://127.0.0.1:5000/api/users/1002
INFO     status_code: 201, response_time(ms): 2.84 ms, response_length: 54 bytes

.

----------------------------------------------------------------------
Ran 2 tests in 0.007s

OK
INFO     Start to run testcase: create user 1003
/api/get-token
INFO     POST http://127.0.0.1:5000/api/get-token
INFO     status_code: 200, response_time(ms): 2.92 ms, response_length: 46 bytes

.
/api/users/1003
INFO     POST http://127.0.0.1:5000/api/users/1003
INFO     status_code: 201, response_time(ms): 5.56 ms, response_length: 54 bytes

.

----------------------------------------------------------------------
Ran 2 tests in 0.011s

OK
INFO     Start to run testcase: create user 1004
/api/get-token
INFO     POST http://127.0.0.1:5000/api/get-token
INFO     status_code: 200, response_time(ms): 5.25 ms, response_length: 46 bytes

.
/api/users/1004
INFO     POST http://127.0.0.1:5000/api/users/1004
INFO     status_code: 201, response_time(ms): 7.02 ms, response_length: 54 bytes

.

----------------------------------------------------------------------
Ran 2 tests in 0.016s

OK
INFO     Start to render Html report ...
INFO     Generated Html report: /Users/debugtalk/MyProjects/HttpRunner-dev/httprunner-docs-v2x/reports/1548518757.html
```

</details>

可以看出，测试用例总共运行了 4 次，并且每次运行时都是采用的不同 user_id。

关于参数化数据驱动，这里只描述了最简单的场景和使用方式，如需了解更多，请进一步阅读[《数据驱动使用手册》](/prepare/parameters/)。

[har2case]: https://github.com/HttpRunner/har2case

## 查看测试报告

在每次使用 hrun 命令运行测试用例后，均会生成一份 HTML 格式的测试报告。报告文件位于 reports 目录下，文件名称为测试用例的开始运行时间。

例如，在运行完 [demo-quickstart-1.json](data/demo-quickstart-1.json) 后，将生成如下形式的测试报告：

![](./images/report-demo-quickstart-1-overview.jpg)

![](./images/report-demo-quickstart-1-log2.jpg)

![](./images/report-demo-quickstart-1-traceback.jpg)

关于测试报告的详细内容，请查看[《测试报告》](/run-tests/report/)部分。

## 总结

到此为止，HttpRunner 的核心功能就介绍完了，掌握本文中的功能特性，足以帮助你应对日常项目工作中至少 80% 的自动化测试需求。

当然，HttpRunner 不止于此，如需挖掘 HttpRunner 的更多特性，实现更复杂场景的自动化测试需求，可继续阅读后续文档。
