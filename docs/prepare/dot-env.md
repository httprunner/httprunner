
## 环境变量的作用

在自动化测试中，有时需要借助环境变量实现某些特定的目的，常见的场景包括：

- 切换测试环境
- 切换测试配置
- 存储敏感数据（从[信息安全](/prepare/security/)的角度出发）

## 设置环境变量

### 在终端中预设环境变量

使用环境变量之前，需要先在系统中设置环境变量名称和值，传统的方式为使用 export 命令（Windows系统中使用 set 命令）：

```bash
$ export UserName=admin
$ echo $UserName
admin
$ export Password=123456
$ echo $Password
123456
```

然后，在程序中就可以对系统中的环境变量进行读取。

```bash
$ python
>>> import os
>>> os.environ["UserName"]
'admin'
```

### 通过 .env 文件设置环境变量

除了这种方式，HttpRunner 还借鉴了 pipenv [加载 `.env` 的方式][pipenv_load_env]。

默认情况下，在自动化测试项目的根目录中，创建 `.env` 文件，并将敏感数据信息放置到其中，存储采用 `name=value` 的格式：

```bash
$ cat .env
UserName=admin
Password=123456
PROJECT_KEY=ABCDEFGH
```

同时，`.env` 文件不应该添加到代码仓库中，建议将 `.env` 加入到 `.gitignore` 中。

HttpRunner 运行时，会自动将 `.env` 文件中的内容加载到运行时（RunTime）的环境变量中，然后在运行时中就可以对环境变量进行读取了。

若需加载不位于自动化项目根目录中的 `.env`，或者其它名称的 `.env` 文件（例如 `production.env`），可以采用 `--dot-env-path` 参数指定文件路径：

```bash
$ hrun /path/to/testcase.yml --dot-env-path /path/to/.env --log-level debug
INFO     Loading environment variables from /path/to/.env
DEBUG    Loaded variable: UserName
DEBUG    Loaded variable: Password
DEBUG    Loaded variable: PROJECT_KEY
...
```

## 引用环境变量

在 HttpRunner 中内置了函数 `environ`（简称 `ENV`），可用于在 YAML/JSON 脚本中直接引用环境变量。

```yaml
- test:
    name: login
    request:
        url: http://host/api/login
        method: POST
        headers:
            Content-Type: application/json
        json:
            username: ${ENV(UserName)}
            password: ${ENV(Password)}
        validate:
            - eq: [status_code, 200]
```

若还需对读取的环境变量做进一步处理，则可以在 `debugtalk.py` 通过 Python 内置的函数 `os.environ` 对环境变量进行引用，然后再实现处理逻辑。

例如，若发起请求的密码需要先与密钥进行拼接并生成 MD5，那么就可以在 `debugtalk.py` 文件中实现如下函数：

```python
import os

def get_encrypt_password():
    raw_passwd = os.environ["Password"]
    PROJECT_KEY = os.environ["PROJECT_KEY"])
    password = (raw_passwd + PROJECT_KEY).encode('ascii')
    return hmac.new(password, hashlib.sha1).hexdigest()
```

然后，在 YAML/JSON 格式的测试用例中，就可以通过`${func()}`的方式引用环境变量的值了。

```yaml
- test:
    name: login
    request:
        url: http://host/api/login
        method: POST
        headers:
            Content-Type: application/json
        json:
            username: ${ENV(UserName)}
            password: ${get_encrypt_password()}
        validate:
            - eq: [status_code, 200]
```

[pipenv_load_env]: https://docs.pipenv.org/advanced/#automatic-loading-of-env
