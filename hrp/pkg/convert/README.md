# hrp convert

## 快速上手

```shell
$ hrp convert -h
convert multiple source format to HttpRunner JSON/YAML/gotest/pytest cases

Usage:
  hrp convert $path... [flags]

Flags:
      --from-har            load from HAR format
      --from-json           load from json case format (default true)
      --from-postman        load from postman format
      --from-yaml           load from yaml case format
  -h, --help                help for convert
  -d, --output-dir string   specify output directory
  -p, --profile string      specify profile path to override headers and cookies
      --to-json             convert to JSON case scripts (default true)
      --to-pytest           convert to pytest scripts
      --to-yaml             convert to YAML case scripts

Global Flags:
      --log-json           set log to json format
  -l, --log-level string   set log level (default "INFO")
      --venv string        specify python3 venv path
```

`hrp convert` 指令用于将 HAR/Postman/JMeter/Swagger 文件或 curl/Apache ab 指令转化为 HttpRunner JSON/YAML/gotest/pytest 形态的测试用例，同时也支持 HttpRunner 测试用例各个形态之间的相互转化。

该指令所有选项的详细说明如下：

- `--to-json / --to-yaml / --to-gotest / --to-pytest` 用于将输入转化为对应形态的 HttpRunner 测试用例，四个选项中最多只能指定一个，如果不指定则默认会将输入转化为 JSON 形态的测试用例
- `--output-dir` 后接测试用例的期望输出目录的路径，用于将转换生成的测试用例输出到对应的文件夹；默认输出的文件夹为源文件所在的文件夹
- `--profile` 后接 profile 配置文件的路径，目前支持替换（不存在则会创建）或者覆盖输入的外部脚本/测试用例中的 `Headers` 和 `Cookies` 信息，profile 文件的后缀可以为 `json/yaml/yml`，下面给出两类 profile 配置文件的示例：

- 根据 profile 替换指定的 `Headers` 和 `Cookies` 信息

```yaml
headers:
  Header1: "this header will be created or updated"
cookies:
  Cookie1: "this cookie will be created or updated"
```

- 根据 profile 覆盖原有的 `Headers` 和 `Cookies` 信息

```yaml
override: true
headers:
  Header1: "all original headers will be overridden"
cookies:
  Cookie1: "all original cookies will be overridden"
```

## 注意事项

1. 输出的测试用例文件名格式为 `源文件名称（不带拓展名）` + `_test` + `.json/.yaml/.go/.py 后缀`，如果该文件已经存在则会进行覆盖
2. 在 profile 文件中，指定 `override` 字段为 `false/true` 可以选择修改模式为替换/覆盖。需要注意的是，如果不指定该字段则 profile 的默认修改模式为替换模式
3. 输入为 JSON/YAML 测试用例时，良好兼容 Golang/Python 双引擎的请求体、断言格式细微差异，输出的 JSON/YAML 则统一采用 Golang 引擎的风格


## 转换流程图

`hrp convert` 的转换过程流程图如下：
![flow chart](asset/flowgram.png)

## 开发进度

`hrp convert` 当前的开发进度如下：

| from \ to | JSON | YAML | GoTest | PyTest |
|:---------:|:----:|:----:|:------:|:------:|
|    HAR    |  ✅   |  ✅   |   ❌    |   ✅    |
|  Postman  |  ✅   |  ✅   |   ❌    |   ✅    |
|  JMeter   |  ❌   |  ❌   |   ❌    |   ❌    |
|  Swagger  |  ❌   |  ❌   |   ❌    |   ❌    |
|   curl    |  ✅   |  ✅   |   ❌    |   ✅    |
| Apache ab |  ❌   |  ❌   |   ❌    |   ❌    |
|   JSON    |  ✅   |  ✅   |   ❌    |   ✅    |
|   YAML    |  ✅   |  ✅   |   ❌    |   ✅    |
|  GoTest   |  ❌   |  ❌   |   ❌    |   ❌    |
|  PyTest   |  ❌   |  ❌   |   ❌    |   ❌    |