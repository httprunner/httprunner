# Builtin

## Builtin assertions

HttpRunner+ validation should follow the following format. `check`, `assert` and `expect` are required field.

```json
{
    "check": "status_code", // target field, usually used with jmespath
    "assert": "equals", // assertion method, you can use builtin method or custom defined function
    "expect": 200, // expected value
    "msg": "check response status code" // optional, print this message if assertion failed
}
```

The `assert` method name will be mapped to a built-in function with the following function signature.

```go
func(t assert.TestingT, actual interface{}, expected interface{}, msgAndArgs ...interface{}) bool
```

Currently, HttpRunner+ has the following built-in assertion functions.

| `assert` | Description | A(check), B(expect) | examples |
| --- | --- | --- | --- |
| `eq`, `equals`, `equal` | value is equal | A == B | 9 eq 9 |
| `lt`, `less_than` | less than | A < B | 7 lt 8 |
| `le`, `less_or_equals` | less than or equals | A <= B | 7 le 8, 8 le 8 |
| `gt`, `greater_than` | greater than | A > B | 8 gt 7 |
| `ge`, `greater_or_equals` | greater than or equals | A >= B | 8 ge 7, 8 ge 8 |
| `ne`, `not_equal` | not equals | A != B | 6 ne 9 |
| `str_eq`, `string_equals` | string equals | str(A) == str(B) | 123 str_eq '123' |
| `len_eq`, `length_equals`, `length_equal` | length equals | len(A) == B | 'abc' len_eq 3, [1,2] len_eq 2 |
| `len_gt`, `count_gt`, `length_greater_than` | length greater than | len(A) > B | 'abc' len_gt 2, [1,2,3] len_gt 2 |
| `len_ge`, `count_ge`, `length_greater_or_equals` | length greater than or equals | len(A) >= B | 'abc' len_ge 3, [1,2,3] len_gt 3 |
| `len_lt`, `count_lt`, `length_less_than` | length less than | len(A) < B | 'abc' len_lt 4, [1,2,3] len_lt 4 |
| `len_le`, `count_le`, `length_less_or_equals` | length less than or equals | len(A) <= B | 'abc' len_le 3, [1,2,3] len_le 3 |
| `contains` | contains | [1, 2] contains 1 | 'abc' contains 'a', [1,2,3] len_lt 4 |
| `contained_by` | contained by | A in B | 'a' contained_by 'abc', 1 contained_by [1,2] |
| `type_match` | A and B are in the same type | type(A) == type(B) | 123 type_match 1 |
| `regex_match` | regex matches | re.match(B, A) | 'abcdef' regex_match 'a\w+d' |
| `startswith` | starts with | A.startswith(B) is True | 'abc' startswith 'ab' |
| `endswith` | ends with | A.endswith(B) is True | 'abc' endswith 'bc' |

## Builtin functions

HttpRunner v5 提供了丰富的内置函数，支持各种常用的数据处理和辅助功能。

| Name | Arguments | Description |
| --- | --- | --- |
| `get_timestamp` | () | 获取当前时间的13位时间戳 |
| `sleep` | (n int) | 休眠 n 秒，模拟思考时间 |
| `gen_random_string` | (n int) | 生成长度为 n 的随机字符串 |
| `random_int` | (max int) | 生成 0 到 max-1 的随机整数 |
| `random_range` | (a, b float64) | 生成 a 到 b 范围内的随机浮点数 |
| `max` | (a, b float64) | 返回两个数中的最大值 |
| `md5` | (str string) | 计算字符串的 MD5 哈希值 |
| `parameterize`, `P` | (filepath string) | 从 CSV 文件加载参数化数据 |
| `split_by_comma` | (s string) | 按逗号分割字符串 |
| `environ`, `ENV` | (key string) | 获取环境变量值 |
| `load_ws_message` | (filepath string) | 加载 WebSocket 消息数据 |
| `multipart_encoder` | (formMap map) | 编码 multipart/form-data 数据 |
| `multipart_content_type` | (writer *TFormDataWriter) | 获取 multipart 内容类型 |

### 使用示例

```yaml
# 在测试用例中使用内置函数
teststeps:
- name: test with builtin functions
  request:
    method: POST
    url: /api/users
    headers:
      X-Timestamp: "{{ get_timestamp() }}"
      X-Random-ID: "{{ gen_random_string(10) }}"
    json:
      username: "user_{{ random_int(1000) }}"
      password: "{{ md5('secret123') }}"
  validate:
    - check: status_code
      assert: eq
      expect: 201
```

### v5 版本增强

- 改进了文件上传处理，支持 `@` 标识符
- 增强了 multipart/form-data 编码功能
- 优化了环境变量访问性能
- 添加了更多数学计算函数
| `gen_random_string` | (n int) | get the n-digit random string. |
| `max` | (m,n int) | get the maximum of two numbers m and n. |
| `md5` | (s string) | get the MD5 of the input string s. |
