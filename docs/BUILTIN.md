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

| Name | Arguments | Description |
| --- | --- | --- |
| `get_timestamp` | () | get the thirteen-digit timestamp of current time. |
| `sleep` | (n int) | sleep n seconds to simulate the thinking time. |
| `gen_random_string` | (n int) | get the n-digit random string. |
| `max` | (m,n int) | get the maximum of two numbers m and n. |
| `md5` | (s string) | get the MD5 of the input string s. |
