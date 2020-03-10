
HttpRunner 从 `1.3.1` 版本开始，支持对 JSON 格式测试用例的内容进行样式美化功能。

## JSON 格式美化

与 YAML 格式不同，JSON 格式不强制要求缩进和换行，这有点类似于 C 语言和 Python 语言的差异。

例如，`demo-quickstart.json`文件也可以改写为如下形式。

```json
[{"config": {"name": "testcase description","variables": [],"request": {"base_url": "","headers": {"User-Agent": "python-requests/2.18.4"}}}},{"test": {"name": "/api/get-token","request": {"url": "http://127.0.0.1:5000/api/get-token","headers": {"device_sn": "FwgRiO7CNA50DSU","user_agent": "iOS/10.3","os_platform": "ios","app_version": "2.8.6","Content-Type": "application/json"},"method": "POST","json": {"sign": "9c0c7e51c91ae963c833a4ccbab8d683c4a90c98"}},"validate": [{"eq": ["status_code",200]},{"eq": ["headers.Content-Type","application/json"]},{"eq": ["content.success",true]},{"eq": ["content.token","baNLX1zhFYP11Seb"]}]}},{"test": {"name": "/api/users/1000","request": {"url": "http://127.0.0.1:5000/api/users/1000","headers": {"device_sn": "FwgRiO7CNA50DSU","token": "baNLX1zhFYP11Seb","Content-Type": "application/json"},"method": "POST","json": {"name": "user1","password": "123456"}},"validate": [{"eq": ["status_code",201]},{"eq": ["headers.Content-Type","application/json"]},{"eq": ["content.success",true]},{"eq": ["content.msg","user created successfully."]}]}}]
```

虽然上面 JSON 格式的测试用例也能正常执行，但测试用例文件的可读性太差，不利于阅读和维护。

针对该需求，可使用 `--prettify` 参数对 JSON 格式用例文件进行样式美化。

可指定单个 JSON 用例文件路径。

```bash
$ hrun --prettify docs/data/demo-quickstart.json
Start to prettify JSON file: docs/data/demo-quickstart.json
success: docs/data/demo-quickstart.pretty.json
```

也可指定多个 JSON 用例文件路径。

```bash
$ hrun --prettify docs/data/demo-quickstart.json docs/data/demo-quickstart.yml docs/data/demo-quickstart-0.json
WARNING  Only JSON file format can be prettified, skip: docs/data/demo-quickstart.yml
Start to prettify JSON file: docs/data/demo-quickstart.json
success: docs/data/demo-quickstart.pretty.json
Start to prettify JSON file: docs/data/demo-quickstart-0.json
success: docs/data/demo-quickstart-0.pretty.json
```

如上所示，当传入的文件后缀不是`.json`，HttpRunner 会打印 WARNING 信息，并跳过检测。

若转换成功，则打印美化后的文件路径；若 JSON 文件格式存在异常，则打印详细的报错信息，精确到错误在文件中出现的行和列。
