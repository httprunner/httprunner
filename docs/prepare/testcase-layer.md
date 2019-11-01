

## 测试用例分层模型

在自动化测试领域，自动化测试用例的可维护性是极其重要的因素，直接关系到自动化测试能否持续有效地在项目中开展。

概括来说，测试用例分层机制的核心是将接口定义、测试步骤、测试用例、测试场景进行分离，单独进行描述和维护，从而尽可能地减少自动化测试用例的维护成本。

逻辑关系图如下所示：

![](../images/testcase-layer.png)

同时，强调如下几点核心概念：

- 测试用例（testcase）应该是完整且独立的，每条测试用例应该是都可以独立运行的
- 测试用例是测试步骤（teststep）的 `有序` 集合，每一个测试步骤对应一个 API 的请求描述
- 测试用例集（testsuite）是测试用例的 `无序` 集合，集合中的测试用例应该都是相互独立，不存在先后依赖关系的；如果确实存在先后依赖关系，那就需要在测试用例中完成依赖的处理

如果对于上述第三点感觉难以理解，不妨看下上图中的示例：

- testcase1 依赖于 testcase2，那么就可以在测试步骤（teststep12）中对 testcase2 进行引用，然后 testcase1 就是完整且可独立运行的；
- 在 testsuite 中，testcase1 与 testcase2 相互独立，运行顺序就不再有先后依赖关系了。

## 分层描述详解

理解了测试用例分层模型，接下来我们再来看下在分层模型下，接口、测试用例、测试用例集的描述形式。

### 接口定义（API）

为了更好地对接口描述进行管理，推荐使用独立的文件对接口描述进行存储，即每个文件对应一个接口描述。

接口定义描述的主要内容包括：**name**、variables、**request**、base_url、validate 等，形式如下：

```yaml
name: get headers
base_url: http://httpbin.org
variables:
    expected_status_code: 200
request:
    url: /headers
    method: GET
validate:
    - eq: ["status_code", $expected_status_code]
    - eq: [content.headers.Host, "httpbin.org"]
```

其中，name 和 request 部分是必须的，request 中的描述形式与 [requests.request](http://docs.python-requests.org/en/master/api/) 完全相同。

另外，API 描述需要尽量保持完整，做到可以单独运行。如果在接口描述中存在变量引用的情况，可在 variables 中对参数进行定义。通过这种方式，可以很好地实现单个接口的调试。

```bash
$ hrun api/get_headers.yml
INFO     Start to run testcase: get headers
headers
INFO     GET http://httpbin.org/headers
INFO     status_code: 200, response_time(ms): 477.32 ms, response_length: 157 bytes

.

----------------------------------------------------------------------
Ran 1 test in 0.478s

OK
```

### 测试用例（testcase）

#### 引用接口定义

有了接口的定义描述后，我们编写测试场景时就可以直接引用接口定义了。

在测试步骤（teststep）中，可通过 `api` 字段引用接口定义，引用方式为对应 API 文件的路径，绝对路径或相对路径均可。推荐使用相对路径，路径基准为项目根目录，即 `debugtalk.py` 所在的目录路径。

```yaml
- config:
    name: "setup and reset all."
    variables:
        user_agent: 'iOS/10.3'
        device_sn: "TESTCASE_SETUP_XXX"
        os_platform: 'ios'
        app_version: '2.8.6'
    base_url: "http://127.0.0.1:5000"
    verify: False
    output:
        - session_token

- test:
    name: get token (setup)
    api: api/get_token.yml
    variables:
        user_agent: 'iOS/10.3'
        device_sn: $device_sn
        os_platform: 'ios'
        app_version: '2.8.6'
    extract:
        - session_token: content.token
    validate:
        - eq: ["status_code", 200]
        - len_eq: ["content.token", 16]

- test:
    name: reset all users
    api: api/reset_all.yml
    variables:
        token: $session_token
```

若需要控制或改变接口定义中的参数值，可在测试步骤中指定 variables 参数，覆盖 API 中的 variables 实现。

同样地，在测试步骤中定义 validate 后，也会与 API 中的 validate 合并覆盖。因此推荐的做法是，在 API 定义中的 validate 只描述最基本的校验项，例如 status_code，对于与业务逻辑相关的更多校验项，在测试步骤的 validate 中进行描述。

#### 引用测试用例

在测试用例的测试步骤中，除了可以引用接口定义，还可以引用其它测试用例。通过这种方式，可以在避免重复描述的同时，解决测试用例的依赖关系，从而保证每个测试用例都是独立可运行的。

在测试步骤（teststep）中，可通过 `testcase` 字段引用其它测试用例，引用方式为对应测试用例文件的路径，绝对路径或相对路径均可。推荐使用相对路径，路径基准为项目根目录，即 `debugtalk.py` 所在的目录路径。

例如，在上面的测试用例（"setup and reset all."）中，实现了对获取 token 功能的测试；同时，在很多其它功能中都会依赖于获取 token 的功能，如果将该功能的测试步骤脚本拷贝到其它功能的测试用例中，那么就会存在大量重复，当需要对该部分进行修改时就需要修改所有地方，显然不便于维护。

比较好的做法是，在其它功能的测试用例（如创建用户）中，引用获取 token 功能的测试用例（testcases/setup.yml）作为一个测试步骤，从而创建用户（"create user and check result."）这个测试用例也变得独立可运行了。

```yaml
- config:
    name: "create user and check result."
    id: create_user
    base_url: "http://127.0.0.1:5000"
    variables:
        uid: 9001
        device_sn: "TESTCASE_CREATE_XXX"
    output:
        - session_token

- test:
    name: setup and reset all (override) for $device_sn.
    testcase: testcases/setup.yml
    output:
        - session_token

- test:
    name: create user and check result.
    variables:
        token: $session_token
    testcase: testcases/deps/check_and_create.yml
```

### 测试用例集（testsuite）

当测试用例数量比较多以后，为了方便管理和实现批量运行，通常需要使用测试用例集来对测试用例进行组织。

在前文的测试用例分层模型中也强调了，测试用例集（testsuite）是测试用例的 `无序` 集合，集合中的测试用例应该都是相互独立，不存在先后依赖关系的；如果确实存在先后依赖关系，那就需要在测试用例中完成依赖的处理。

因为是 `无序` 集合，因此测试用例集的描述形式会与测试用例有些不同，在每个测试用例集文件中，第一层级存在两类字段：

- config: 测试用例集的总体配置参数
- testcases: 值为字典结构（无序），key 为测试用例的名称，value 为测试用例的内容；在引用测试用例时也可以指定 variables，实现对引用测试用例中 variables 的覆盖。

#### 非参数化场景

```yaml
config:
    name: create users with uid
    variables:
        device_sn: ${gen_random_string(15)}
        var_a: ${gen_random_string(5)}
        var_b: $var_a
    base_url: "http://127.0.0.1:5000"

testcases:
    create user 1000 and check result.:
        testcase: testcases/create_user.yml
        variables:
            uid: 1000
            var_c: ${gen_random_string(5)}
            var_d: $var_c

    create user 1001 and check result.:
        testcase: testcases/create_user.yml
        variables:
            uid: 1001
            var_c: ${gen_random_string(5)}
            var_d: $var_c
```

#### 参数化场景（parameters）

对于参数化场景，可通过 parameters 实现，描述形式如下所示。

```yaml
config:
    name: create users with parameters
    variables:
        device_sn: ${gen_random_string(15)}
    base_url: "http://127.0.0.1:5000"

testcases:
    create user $uid and check result for $device_sn.:
        testcase: testcases/create_user.yml
        variables:
            uid: 1000
            device_sn: TESTSUITE_XXX
        parameters:
            uid: [101, 102, 103]
            device_sn: [TESTSUITE_X1, TESTSUITE_X2]
```

参数化后，parameters 中的变量将采用笛卡尔积组合形成参数列表，依次覆盖 variables 中的参数，驱动测试用例的运行。

## 文件目录结构管理 && 脚手架工具

在对测试用例文件进行组织管理时，对于文件的存储位置均没有要求和限制，在引用时只需要指定对应的文件路径即可。但从约定大于配置的角度，最好是按照推荐的文件夹名称进行存储管理，并可通过子目录实现项目模块分类管理。

推荐的方式汇总如下：

- `debugtalk.py` 放置在项目根目录下，假设为 `PRJ_ROOT_DIR`
- `.env` 放置在项目根目录下，路径为 `PRJ_ROOT_DIR/.env`
- 接口定义（API）放置在 `PRJ_ROOT_DIR/api/` 目录下
- 测试用例（testcase）放置在 `PRJ_ROOT_DIR/testcases/` 目录下
- 测试用例集（testsuite）文件必须放置在 `PRJ_ROOT_DIR/testsuites/` 目录下
- data 文件夹：存储参数化文件，或者项目依赖的文件，路径为 `PRJ_ROOT_DIR/data/`
- reports 文件夹：存储 HTML 测试报告，生成路径为 `PRJ_ROOT_DIR/reports/`

目录结构如下所示：

```bash
$ tree tests
tests
├── .env
├── data
│   ├── app_version.csv
│   └── account.csv
├── api
│   ├── create_user.yml
│   ├── get_headers.yml
│   ├── get_token.yml
│   ├── get_user.yml
│   └── reset_all.yml
├── debugtalk.py
├── testcases
│   ├── create_user.yml
│   ├── deps
│   │   └── check_and_create.yml
│   └── setup.yml
└── testsuites
    ├── create_users.yml
    └── create_users_with_parameters.yml
```

**项目脚手架**

同时，在 `HttpRunner` 中实现了一个脚手架工具，可以快速创建项目的目录结构。该想法来源于 `Django` 的 `django-admin.py startproject project_name`。

使用方式也与 `Django` 类似，只需要通过 `--startproject` 指定新项目的名称即可。

```bash
$ hrun --startproject demo
Start to create new project: demo
CWD: /Users/debugtalk/MyProjects/examples

created folder: demo
created folder: demo/api
created folder: demo/testcases
created folder: demo/testsuites
created folder: demo/reports
created file: demo/debugtalk.py
created file: demo/.env
```


## 相关参考

- [《HttpRunner 的测试用例分层机制（已过期）》](/post/HttpRunner-testcase-layer)
- 测试用例分层详细示例：[HttpRunner/tests](https://github.com/HttpRunner/HttpRunner/tree/master/tests)
