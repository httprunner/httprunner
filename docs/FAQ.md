## 无法自动安装PyUnitReport依赖库

如果安装过程中出现如下报错：

```text
Downloading/unpacking PyUnitReport (from ApiTestEngine)
  Could not find any downloads that satisfy the requirement PyUnitReport (from ApiTestEngine)
```

那么需要先手动安装`PyUnitReport`，安装方式如下：

```bash
$ pip install git+https://github.com/debugtalk/PyUnitReport.git#egg=PyUnitReport
```

然后再重新安装`ApiTestEngine`即可。

```bash
$ pip install git+https://github.com/debugtalk/ApiTestEngine.git#egg=ApiTestEngine
```
