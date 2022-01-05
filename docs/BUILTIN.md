# Builtin

## Assertion Methods

### Usage
In "teststeps" of each json/yaml testcase, the "validate" part contains four fields: "check", "assert", "expect" and 
"msg", when using assertion methods, method name should be put in "assert" field. The assertion result of "check" 
element will be checked out using the regulation you put in "assert" field and compared with the element in "expect"
field.

### Method List

- equals: assert the element to check equals the expected element.
- equal: alias for equals.
- greater_than: assert the element to check is greater than the expected element.
- less_than: assert the element to check is less than the expected element.
- greater_or_equals: assert the element to check is greater than or equal with the expected element.
- less_or_equals: assert the element to check is less than or equal with the expected element.
- not_equal: assert the element to check is not equal with the expected element.
- contained_by: assert the expected element contains the element to check.
- regex_match: assert the element to check matches the expected element using regex.
- type_match: assert the element to check matches the expected element in type.
- startswith: assert the element to check starts with the expected element.
- endswith: assert the element to check ends with the expected element.
- length_equals: assert the length of the element to check is equal with the expected element.
- length_equal: alias for length_equals.
- contains: assert the element to check contains the expected element.
- string_equals: assert the string is equal with the expected string.

## Common Functions

### Usage
The common functions are useful during the variables configuration, you can use "${FUNCTION_NAME}" to call the specific 
function to define variables.

### Function List
- get_timestamp: get the thirteen-digit timestamp of current time. (call without argument)
- sleep: sleep n seconds to simulate the thinking time. (call with one argument n)
- gen_random_string: get the n-digit random string. (call with one argument n)
- max: get the maximum of two numbers m and n. (call with two argument m and n)
- md5: get the MD5 of the input string s. (call with one argument s)



