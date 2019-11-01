## 介绍

在自动化测试中，经常会遇到如下场景：

- 测试搜索功能，只有一个搜索输入框，但有 10 种不同类型的搜索关键字；
- 测试账号登录功能，需要输入用户名和密码，按照等价类划分后有 20 种组合情况。

这里只是随意找了两个典型的例子，相信大家都有遇到过很多类似的场景。总结下来，就是在我们的自动化测试脚本中存在参数，并且我们需要采用不同的参数去运行。

经过概括，参数基本上分为两种类型：

- 单个独立参数：例如前面的第一种场景，我们只需要变换搜索关键字这一个参数
- 多个具有关联性的参数：例如前面的第二种场景，我们需要变换用户名和密码两个参数，并且这两个参数需要关联组合

然后，对于参数而言，我们可能具有一个参数列表，在脚本运行时需要按照不同的规则去取值，例如顺序取值、随机取值、循环取值等等。

这就是典型的参数化和数据驱动。

如需了解 HttpRunner 参数化数据驱动机制的实现原理和技术细节，可前往阅读[《HttpRunner 实现参数化数据驱动机制》](http://debugtalk.com/post/httprunner-data-driven/)。

## 测试用例集（testsuite）准备

从 2.0.0 版本开始，HttpRunner 不再支持在测试用例文件中进行参数化配置；参数化的功能需要在 testsuite 中实现。变更的目的是让测试用例（testcase）的概念更纯粹，关于测试用例和测试用例集的概念定义，详见[《测试用例组织》](/prepare/parameters/)。

参数化机制需要在测试用例集（testsuite）中实现。如需实现数据驱动机制，需要创建一个 testsuite，在 testsuite 中引用测试用例，并定义参数化配置。

测试用例集（testsuite）的格式如下所示：

```yaml
config:
    name: testsuite description

testcases:
    testcase1_name:
        testcase: /path/to/testcase1

    testcase2_name:
        testcase: /path/to/testcase2
```

需要注意的是，testsuite 和 testcase 的格式存在较大区别，详见[《测试用例组织》](/prepare/testcase-structure/)。


## 参数配置概述

如需对某测试用例（testcase）实现参数化数据驱动，需要使用 `parameters` 关键字，定义参数名称并指定数据源取值方式。

参数名称的定义分为两种情况：

- 独立参数单独进行定义；
- 多个参数具有关联性的参数需要将其定义在一起，采用短横线（`-`）进行连接。

数据源指定支持三种方式：

- 在 YAML/JSON 中直接指定参数列表：该种方式最为简单易用，适合参数列表比较小的情况
- 通过内置的 parameterize（可简写为P）函数引用 CSV 文件：该种方式需要准备 CSV 数据文件，适合数据量比较大的情况
- 调用 debugtalk.py 中自定义的函数生成参数列表：该种方式最为灵活，可通过自定义 Python 函数实现任意场景的数据驱动机制，当需要动态生成参数列表时也需要选择该种方式

三种方式可根据实际项目需求进行灵活选择，同时支持多种方式的组合使用。假如测试用例中定义了多个参数，那么测试用例在运行时会对参数进行笛卡尔积组合，覆盖所有参数组合情况。

使用方式概览如下：

```yaml
config:
    name: "demo"

testcases:
    testcase1_name:
        testcase: /path/to/testcase1
        parameters:
            user_agent: ["iOS/10.1", "iOS/10.2", "iOS/10.3"]
            user_id: ${P(user_id.csv)}
            username-password: ${get_account(10)}
```


## 参数配置详解

将参数名称定义和数据源指定方式进行组合，共有 6 种形式。现分别针对每一类情况进行详细说明。

### 独立参数 & 直接指定参数列表

对于参数列表比较小的情况，最简单的方式是直接在 YAML/JSON 中指定参数列表内容。

例如，对于独立参数 `user_id`，参数列表为 `[1001, 1002, 1003, 1004]`，那么就可以按照如下方式进行配置：

```yaml
config:
    name: testcase description

testcases:
    create user:
        testcase: demo-quickstart-6.yml
        parameters:
            user_id: [1001, 1002, 1003, 1004]
```

进行该配置后，测试用例在运行时就会对 user_id 实现数据驱动，即分别使用 `[1001, 1002, 1003, 1004]` 四个值运行测试用例。

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

### 关联参数 & 直接指定参数列表

对于具有关联性的多个参数，例如 username 和 password，那么就可以按照如下方式进行配置：

```yaml
config:
    name: "demo"

testcases:
    testcase1_name:
        testcase: /path/to/testcase1
        parameters:
            username-password:
                - ["user1", "111111"]
                - ["user2", "222222"]
                - ["user3", "333333"]
```

进行该配置后，测试用例在运行时就会对 username 和 password 实现数据驱动，即分别使用 `{"username": "user1", "password": "111111"}`、`{"username": "user2", "password": "222222"}`、`{"username": "user3", "password": "333333"}` 运行 3 次测试，并且保证参数值总是成对使用。

### 独立参数 & 引用 CSV 文件

对于已有参数列表，并且数据量比较大的情况，比较适合的方式是将参数列表值存储在 CSV 数据文件中。

对于 CSV 数据文件，需要遵循如下几项约定的规则：

- CSV 文件中的第一行必须为参数名称，从第二行开始为参数值，每个（组）值占一行；
- 若同一个 CSV 文件中具有多个参数，则参数名称和数值的间隔符需实用英文逗号；
- 在 YAML/JSON 文件引用 CSV 文件时，文件路径为基于项目根目录（debugtalk.py 所在路径）的相对路径。

例如，user_id 的参数取值范围为 1001～2000，那么我们就可以创建 user_id.csv，并且在文件中按照如下形式进行描述。

```csv
user_id
1001
1002
...
1999
2000
```

然后在 YAML/JSON 测试用例文件中，就可以通过内置的 `parameterize`（可简写为 `P`）函数引用 CSV 文件。

假设项目的根目录下有 data 文件夹，user_id.csv 位于其中，那么 user_id.csv 的引用描述如下：

```yaml
config:
    name: "demo"

testcases:
    testcase1_name:
        testcase: /path/to/testcase1
        parameters:
            user_id: ${P(data/user_id.csv)}
```

即 `P` 函数的参数（CSV 文件路径）是相对于项目根目录的相对路径。当然，这里也可以使用 CSV 文件在系统中的绝对路径，不过这样的话在项目路径变动时就会出现问题，因此推荐使用相对路径的形式。

### 关联参数 & 引用 CSV 文件

对于具有关联性的多个参数，例如 username 和 password，那么就可以创建 [account.csv](/data/account.csv)，并在文件中按照如下形式进行描述。

```csv
username,password
test1,111111
test2,222222
test3,333333
```

然后在 YAML/JSON 测试用例文件中，就可以通过内置的 `parameterize`（可简写为 `P`）函数引用 CSV 文件。

假设项目的根目录下有 data 文件夹，account.csv 位于其中，那么 account.csv 的引用描述如下：

```yaml
config:
    name: "demo"

testcases:
    testcase1_name:
        testcase: /path/to/testcase1
        parameters:
            username-password: ${P(data/account.csv)}
```

需要说明的是，在 parameters 中指定的参数名称必须与 CSV 文件中第一行的参数名称一致，顺序可以不一致，参数个数也可以不一致。

例如，在 [account.csv](/data/account.csv) 文件中可以包含多个参数，username、password、phone、age：

```csv
username,password,phone,age
test1,111111,18600000001,21
test2,222222,18600000002,22
test3,333333,18600000003,23
```

而在 YAML/JSON 测试用例文件中指定参数时，可以只使用部分参数，并且参数顺序无需与 CSV 文件中参数名称的顺序一致。

```yaml
config:
    name: "demo"

testcases:
    testcase1_name:
        testcase: /path/to/testcase1
        parameters:
            phone-username: ${P(account.csv)}
```

### 独立参数 & 引用自定义函数

对于没有现成参数列表，或者需要更灵活的方式动态生成参数的情况，可以通过在 debugtalk.py 中自定义函数生成参数列表，并在 YAML/JSON 引用自定义函数的方式。

例如，若需对 user_id 进行参数化数据驱动，参数取值范围为 1001～1004，那么就可以在 debugtalk.py 中定义一个函数，返回参数列表。

```python
def get_user_id():
    return [
        {"user_id": 1001},
        {"user_id": 1002},
        {"user_id": 1003},
        {"user_id": 1004}
    ]
```

然后，在 YAML/JSON 的 parameters 中就可以通过调用自定义函数的形式来指定数据源。

```yaml
config:
    name: "demo"

testcases:
    testcase1_name:
        testcase: /path/to/testcase1
        parameters:
            user_id: ${get_user_id()}
```

另外，通过函数的传参机制，还可以实现更灵活的参数生成功能，在调用函数时指定需要生成的参数个数。

### 关联参数 & 引用自定义函数

对于具有关联性的多个参数，实现方式也类似。

例如，在 debugtalk.py 中定义函数 get_account，生成指定数量的账号密码参数列表。

```python
def get_account(num):
    accounts = []
    for index in range(1, num+1):
        accounts.append(
            {"username": "user%s" % index, "password": str(index) * 6},
        )

    return accounts
```

那么在 YAML/JSON 的 parameters 中就可以调用自定义函数生成指定数量的参数列表。

```yaml
config:
    name: "demo"

testcases:
    testcase1_name:
        testcase: /path/to/testcase1
        parameters:
            username-password: ${get_account(10)}
```

> 需要注意的是，在自定义函数中，生成的参数列表必须为 `list of dict` 的数据结构，该设计主要是为了与 CSV 文件的处理机制保持一致。


## 参数化运行

完成以上参数定义和数据源准备工作之后，参数化运行与普通测试用例的运行完全一致。

采用 hrun 命令运行自动化测试：

```bash
$ hrun tests/data/demo_parameters.yml
```

采用 locusts 命令运行性能测试：

```bash
$ locusts -f tests/data/demo_parameters.yml
```

区别在于，自动化测试时遍历一遍后会终止执行，性能测试时每个并发用户都会循环遍历所有参数。

## 案例演示

假设我们有一个获取 token 的[测试用例](/data/demo-testcase-get-token.yml)。

<details>
<summary>点击查看 YAML 测试用例</summary>
```yaml
- config:
    name: get token
    base_url: http://127.0.0.1:5000
    variables:
        device_sn: ${gen_random_string(15)}
        os_platform: 'ios'
        app_version: '2.8.6'

- test:
    name: get token with $device_sn, $os_platform, $app_version
    request:
        headers:
            Content-Type: application/json
            User-Agent: python-requests/2.18.4
            app_version: $app_version
            device_sn: $device_sn
            os_platform: $os_platform
        json:
            sign: ${get_sign($device_sn, $os_platform, $app_version)}
        method: POST
        url: /api/get-token
    extract:
        token: content.token
    validate:
        - eq: [status_code, 200]
        - eq: [headers.Content-Type, application/json]
        - eq: [content.success, true]
```
</details>

如果我们需要使用 device_sn、app_version 和 os_platform 这三个参数来进行参数化数据驱动，那么就可以创建一个 [testsuite](/data/demo-parameters-get-token.yml)，并且进行参数化配置。

```yaml
config:
    name: get token with parameters

testcases:
    get token with $user_agent, $app_version, $os_platform:
        testcase: demo-testcase-get-token.yml
        parameters:
            user_agent: ["iOS/10.1", "iOS/10.2", "iOS/10.3"]
            app_version: ${P(app_version.csv)}
            os_platform: ${get_os_platform()}
```

其中，`user_agent` 使用了直接指定参数列表的形式。

[app_version](/data/app_version.csv) 通过 CSV 文件进行参数配置，对应的文件内容为：

```csv
app_version
2.8.5
2.8.6
```

os_platform 使用自定义函数的形式生成参数列表，对应的函数内容为：

```python
def get_os_platform():
    return [
        {"os_platform": "ios"},
        {"os_platform": "android"}
    ]
```

那么，经过笛卡尔积组合，应该总共有 `3*2*2=12` 种参数组合情况。


<details>
<summary>点击查看完整运行日志</summary>
```text
$ hrun docs/data/demo-parameters-get-token.yml
INFO     Start to run testcase: get token with iOS/10.1, 2.8.5, ios
get token with PBJda7SXM2ReWlu, ios, 2.8.5
INFO     POST http://127.0.0.1:5000/api/get-token
INFO     status_code: 200, response_time(ms): 10.66 ms, response_length: 46 bytes

.

----------------------------------------------------------------------
Ran 1 test in 0.026s

OK
INFO     Start to run testcase: get token with iOS/10.1, 2.8.5, android
get token with PBJda7SXM2ReWlu, android, 2.8.5
INFO     POST http://127.0.0.1:5000/api/get-token
INFO     status_code: 200, response_time(ms): 3.03 ms, response_length: 46 bytes

.

----------------------------------------------------------------------
Ran 1 test in 0.004s

OK
INFO     Start to run testcase: get token with iOS/10.1, 2.8.6, ios
get token with PBJda7SXM2ReWlu, ios, 2.8.6
INFO     POST http://127.0.0.1:5000/api/get-token
INFO     status_code: 200, response_time(ms): 10.76 ms, response_length: 46 bytes

.

----------------------------------------------------------------------
Ran 1 test in 0.012s

OK
INFO     Start to run testcase: get token with iOS/10.1, 2.8.6, android
get token with PBJda7SXM2ReWlu, android, 2.8.6
INFO     POST http://127.0.0.1:5000/api/get-token
INFO     status_code: 200, response_time(ms): 4.49 ms, response_length: 46 bytes

.

----------------------------------------------------------------------
Ran 1 test in 0.006s

OK
INFO     Start to run testcase: get token with iOS/10.2, 2.8.5, ios
get token with PBJda7SXM2ReWlu, ios, 2.8.5
INFO     POST http://127.0.0.1:5000/api/get-token
INFO     status_code: 200, response_time(ms): 4.39 ms, response_length: 46 bytes

.

----------------------------------------------------------------------
Ran 1 test in 0.006s

OK
INFO     Start to run testcase: get token with iOS/10.2, 2.8.5, android
get token with PBJda7SXM2ReWlu, android, 2.8.5
INFO     POST http://127.0.0.1:5000/api/get-token
INFO     status_code: 200, response_time(ms): 4.04 ms, response_length: 46 bytes

.

----------------------------------------------------------------------
Ran 1 test in 0.005s

OK
INFO     Start to run testcase: get token with iOS/10.2, 2.8.6, ios
get token with PBJda7SXM2ReWlu, ios, 2.8.6
INFO     POST http://127.0.0.1:5000/api/get-token
INFO     status_code: 200, response_time(ms): 3.44 ms, response_length: 46 bytes

.

----------------------------------------------------------------------
Ran 1 test in 0.004s

OK
INFO     Start to run testcase: get token with iOS/10.2, 2.8.6, android
get token with PBJda7SXM2ReWlu, android, 2.8.6
INFO     POST http://127.0.0.1:5000/api/get-token
INFO     status_code: 200, response_time(ms): 4.03 ms, response_length: 46 bytes

.

----------------------------------------------------------------------
Ran 1 test in 0.005s

OK
INFO     Start to run testcase: get token with iOS/10.3, 2.8.5, ios
get token with PBJda7SXM2ReWlu, ios, 2.8.5
INFO     POST http://127.0.0.1:5000/api/get-token
INFO     status_code: 200, response_time(ms): 5.14 ms, response_length: 46 bytes

.

----------------------------------------------------------------------
Ran 1 test in 0.008s

OK
INFO     Start to run testcase: get token with iOS/10.3, 2.8.5, android
get token with PBJda7SXM2ReWlu, android, 2.8.5
INFO     POST http://127.0.0.1:5000/api/get-token
INFO     status_code: 200, response_time(ms): 7.62 ms, response_length: 46 bytes

.

----------------------------------------------------------------------
Ran 1 test in 0.010s

OK
INFO     Start to run testcase: get token with iOS/10.3, 2.8.6, ios
get token with PBJda7SXM2ReWlu, ios, 2.8.6
INFO     POST http://127.0.0.1:5000/api/get-token
INFO     status_code: 200, response_time(ms): 4.88 ms, response_length: 46 bytes

.

----------------------------------------------------------------------
Ran 1 test in 0.006s

OK
INFO     Start to run testcase: get token with iOS/10.3, 2.8.6, android
get token with PBJda7SXM2ReWlu, android, 2.8.6
INFO     POST http://127.0.0.1:5000/api/get-token
INFO     status_code: 200, response_time(ms): 5.41 ms, response_length: 46 bytes

.

----------------------------------------------------------------------
Ran 1 test in 0.008s

OK
INFO     Start to render Html report ...
INFO     Generated Html report: /Users/debugtalk/MyProjects/HttpRunner-dev/httprunner-docs-v2x/reports/1551950193.html
```
</details>
