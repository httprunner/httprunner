## 运行环境

HttpRunner 是一个基于 Python 开发的测试框架，可以运行在 macOS、Linux、Windows 系统平台上。

**Python 版本**：HttpRunner 支持 Python 3.5 及以上的所有版本，并使用 Travis-CI 进行了[持续集成测试][travis-ci]，测试覆盖的版本包括 2.7/3.5/3.6/3.7/3.8。虽然 HttpRunner 暂时保留了对 Python 2.7 的兼容支持，但强烈建议使用 Python 3.6 及以上版本。

**操作系统**：推荐使用 macOS/Linux。

## 安装方式

HttpRunner 的稳定版本托管在 PyPI 上，可以使用 `pip` 进行安装。

```bash
$ pip install httprunner
```

如果你需要使用最新的开发版本，那么可以采用项目的 GitHub 仓库地址进行安装：

```bash
$ pip install git+https://github.com/HttpRunner/HttpRunner.git@master
```

## 版本升级

假如你之前已经安装过了 HttpRunner，现在需要升级到最新版本，那么你可以使用`-U`参数。该参数对以上三种安装方式均生效。

```bash
$ pip install -U HttpRunner
$ pip install -U git+https://github.com/HttpRunner/HttpRunner.git@master
```

## 安装校验

在 HttpRunner 安装成功后，系统中会新增如下 5 个命令：

- `httprunner`: 核心命令
- `ate`: 曾经用过的命令（当时框架名称为 ApiTestEngine），功能与 httprunner 完全相同
- `hrun`: httprunner 的缩写，功能与 httprunner 完全相同
- `locusts`: 基于 [Locust][Locust] 实现[性能测试](run-tests/load-test.md)
- [`har2case`][har2case]: 辅助工具，可将标准通用的 HAR 格式（HTTP Archive）转换为`YAML/JSON`格式的测试用例

httprunner、hrun、ate 三个命令完全等价，功能特性完全相同，个人推荐使用`hrun`命令。

运行如下命令，若正常显示版本号，则说明 HttpRunner 安装成功。

```text
$ hrun -V
2.4.1

$ har2case -V
0.3.1
```

## 开发者模式

默认情况下，安装 HttpRunner 的时候只会安装运行 HttpRunner 的必要依赖库。

如果你不仅仅是使用 HttpRunner，还需要对 HttpRunner 进行开发调试（debug），那么就需要进行如下操作。

HttpRunner 使用 [poetry][poetry] 对依赖包进行管理，若你还没有安装 poetry，需要先执行如下命令进行按照：

```bash
$ curl -sSL https://raw.githubusercontent.com/python-poetry/poetry/master/get-poetry.py | python
```

获取 HttpRunner 源码：

```bash
$ git clone https://github.com/HttpRunner/HttpRunner.git
```

进入仓库目录，安装所有依赖：

```bash
$ poetry install
```

运行单元测试，若测试全部通过，则说明环境正常。

```bash
$ poetry run python -m unittest discover
```

查看 HttpRunner 的依赖情况：

```text
$ poetry show
certifi           2019.9.11 Python package for providing Mozilla's CA Bundle.
chardet           3.0.4     Universal encoding detector for Python 2 and 3
click             7.0       Composable command line interface toolkit
colorama          0.4.1     Cross-platform colored terminal text.
colorlog          4.0.2     Log formatting with colors!
coverage          4.5.4     Code coverage measurement for Python
coveralls         1.8.2     Show coverage stats online via coveralls.io
docopt            0.6.2     Pythonic argument parser, that will make you smile
filetype          1.0.5     Infer file type and MIME type of any file/buffer. No external dependencies.
flask             0.12.4    A microframework based on Werkzeug, Jinja2 and good intentions
har2case          0.3.1     Convert HAR(HTTP Archive) to YAML/JSON testcases for HttpRunner.
idna              2.8       Internationalized Domain Names in Applications (IDNA)
itsdangerous      1.1.0     Various helpers to pass data to untrusted environments and back.
jinja2            2.10.3    A very fast and expressive template engine.
jsonpath          0.82      An XPath for JSON
markupsafe        1.1.1     Safely add untrusted strings to HTML/XML markup.
pyyaml            5.1.2     YAML parser and emitter for Python
requests          2.22.0    Python HTTP for Humans.
requests-toolbelt 0.9.1     A utility belt for advanced users of python-requests
urllib3           1.25.6    HTTP library with thread-safe connection pooling, file post, and more.
werkzeug          0.16.0    The comprehensive WSGI web application library.
```

调试运行方式：

```bash
# 调试运行 hrun
$ poetry run python -m httprunner -h

# 调试运行 locusts
$ pipenv run python -m httprunner.plugins.locusts -h
```

## Docker

TODO

[travis-ci]: https://travis-ci.org/HttpRunner/HttpRunner
[Locust]: http://locust.io/
[har2case]: https://github.com/HttpRunner/har2case
[poetry]: https://github.com/sdispater/poetry