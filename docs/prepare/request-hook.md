## 概述

HttpRunner 从 `1.4.5` 版本开始实现了全新的 hook 机制，可以在请求前和请求后调用钩子函数。

## 调用 hook 函数

hook 机制分为两个层级：

- 测试用例层面（testcase）
- 测试步骤层面（teststep）

### 测试用例层面（testcase）

在 YAML/JSON 测试用例的 `config` 中新增关键字 `setup_hooks` 和 `teardown_hooks`。

- setup_hooks: 在整个用例开始执行前触发 hook 函数，主要用于准备工作。
- teardown_hooks: 在整个用例结束执行后触发 hook 函数，主要用于测试后的清理工作。

```yaml
- config:
    name: basic test with httpbin
    request:
        base_url: http://127.0.0.1:3458/
    setup_hooks:
        - ${hook_print(setup)}
    teardown_hooks:
        - ${hook_print(teardown)}
```

### 测试步骤层面（teststep）

在 YAML/JSON 测试步骤的 `test` 中新增关键字 `setup_hooks` 和 `teardown_hooks`。

- setup_hooks: 在 HTTP 请求发送前执行 hook 函数，主要用于准备工作；也可以实现对请求的 request 内容进行预处理。
- teardown_hooks: 在 HTTP 请求发送后执行 hook 函数，主要用于测试后的清理工作；也可以实现对响应的 response 进行修改，例如进行加解密等处理。

```json
"test": {
    "name": "get token with $user_agent, $os_platform, $app_version",
    "request": {
        "url": "/api/get-token",
        "method": "POST",
        "headers": {
            "app_version": "$app_version",
            "os_platform": "$os_platform",
            "user_agent": "$user_agent"
        },
        "json": {
            "sign": "${get_sign($user_agent, $device_sn, $os_platform, $app_version)}"
        }
    },
    "validate": [
        {"eq": ["status_code", 200]}
    ],
    "setup_hooks": [
        "${setup_hook_prepare_kwargs($request)}",
        "${setup_hook_httpntlmauth($request)}"
    ],
    "teardown_hooks": [
        "${teardown_hook_sleep_N_secs($response, 2)}"
    ]
}
```

## 编写 hook 函数

hook 函数的定义放置在项目的 `debugtalk.py` 中，在 YAML/JSON 中调用 hook 函数仍然是采用 `${func($a, $b)}` 的形式。

对于测试用例层面的 hook 函数，与 YAML/JSON 中自定义的函数完全相同，可通过自定义参数传参的形式来实现灵活应用。

```python
def hook_print(msg):
    print(msg)
```

对于单个测试用例层面的 hook 函数，除了可传入自定义参数外，还可以传入与当前测试用例相关的信息，包括请求的 `$request` 和响应的 `$response`，用于实现更复杂场景的灵活应用。

### setup_hooks

在测试步骤层面的 setup_hooks 函数中，除了可传入自定义参数外，还可以传入 `$request`，该参数对应着当前测试步骤 request 的全部内容。因为 request 是可变参数类型（dict），因此该函数参数为引用传递，当我们需要对请求参数进行预处理时尤其有用。

e.g.

```python
def setup_hook_prepare_kwargs(request):
    if request["method"] == "POST":
        content_type = request.get("headers", {}).get("content-type")
        if content_type and "data" in request:
            # if request content-type is application/json, request data should be dumped
            if content_type.startswith("application/json") and isinstance(request["data"], (dict, list)):
                request["data"] = json.dumps(request["data"])

            if isinstance(request["data"], str):
                request["data"] = request["data"].encode('utf-8')

def setup_hook_httpntlmauth(request):
    if "httpntlmauth" in request:
        from requests_ntlm import HttpNtlmAuth
        auth_account = request.pop("httpntlmauth")
        request["auth"] = HttpNtlmAuth(
            auth_account["username"], auth_account["password"])
```

通过上述的 `setup_hook_prepare_kwargs` 函数，可以实现根据请求方法和请求的 Content-Type 来对请求的 data 进行加工处理；通过 `setup_hook_httpntlmauth` 函数，可以实现 HttpNtlmAuth 权限授权。

### teardown_hooks

在测试步骤层面的 teardown_hooks 函数中，除了可传入自定义参数外，还可以传入 `$response`，该参数对应着当前请求的响应实例（requests.Response）。

e.g.

```python
def teardown_hook_sleep_N_secs(response, n_secs):
    """ sleep n seconds after request
    """
    if response.status_code == 200:
        time.sleep(0.1)
    else:
        time.sleep(n_secs)
```

通过上述的 `teardown_hook_sleep_N_secs` 函数，可以根据接口响应的状态码来进行不同时间的延迟等待。

另外，在 teardown_hooks 函数中还可以对 response 进行修改。当我们需要先对响应内容进行处理（例如加解密、参数运算），再进行参数提取（extract）和校验（validate）时尤其有用。

例如在下面的测试步骤中，在执行测试后，通过 teardown_hooks 函数将响应结果的状态码和 headers 进行了修改，然后再进行了校验。

```yaml
- test:
    name: alter response
    request:
        url: /headers
        method: GET
    teardown_hooks:
        - ${alter_response($response)}
    validate:
        - eq: ["status_code", 500]
        - eq: ["headers.content-type", "html/text"]
```

```python
def alter_response(response):
    response.status_code = 500
    response.headers["Content-Type"] = "html/text"
```
