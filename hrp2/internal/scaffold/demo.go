package scaffold

import "github.com/httprunner/hrp"

var demoTestCase = &hrp.TestCase{
	Config: hrp.NewConfig("demo with complex mechanisms").
		SetBaseURL("https://postman-echo.com").
		WithVariables(map[string]interface{}{ // global level variables
			"n":       "${sum_ints(1, 2, 2)}",
			"a":       "${sum(10, 2.3)}",
			"b":       3.45,
			"varFoo1": "${gen_random_string($n)}",
			"varFoo2": "${max($a, $b)}", // 12.3; eval with built-in function
		}),
	TestSteps: []hrp.IStep{
		hrp.NewStep("transaction 1 start").StartTransaction("tran1"), // start transaction
		hrp.NewStep("get with params").
			WithVariables(map[string]interface{}{ // step level variables
				"n":       3,                // inherit config level variables if not set in step level, a/varFoo1
				"b":       34.5,             // override config level variable if existed, n/b/varFoo2
				"varFoo2": "${max($a, $b)}", // 34.5; override variable b and eval again
				"name":    "get with params",
			}).
			SetupHook("${setup_hook_example($name)}").
			GET("/get").
			TeardownHook("${teardown_hook_example($name)}").
			WithParams(map[string]interface{}{"foo1": "$varFoo1", "foo2": "$varFoo2"}). // request with params
			WithHeaders(map[string]string{"User-Agent": "HttpRunnerPlus"}).             // request with headers
			Extract().
			WithJmesPath("body.args.foo1", "varFoo1"). // extract variable with jmespath
			Validate().
			AssertEqual("status_code", 200, "check response status code").        // validate response status code
			AssertStartsWith("headers.\"Content-Type\"", "application/json", ""). // validate response header
			AssertLengthEqual("body.args.foo1", 5, "check args foo1").            // validate response body with jmespath
			AssertLengthEqual("$varFoo1", 5, "check args foo1").                  // assert with extracted variable from current step
			AssertEqual("body.args.foo2", "34.5", "check args foo2"),             // notice: request params value will be converted to string
		hrp.NewStep("transaction 1 end").EndTransaction("tran1"), // end transaction
		hrp.NewStep("post json data").
			POST("/post").
			WithBody(map[string]interface{}{
				"foo1": "$varFoo1",       // reference former extracted variable
				"foo2": "${max($a, $b)}", // 12.3; step level variables are independent, variable b is 3.45 here
			}).
			Validate().
			AssertEqual("status_code", 200, "check status code").
			AssertLengthEqual("body.json.foo1", 5, "check args foo1").
			AssertEqual("body.json.foo2", 12.3, "check args foo2"),
		hrp.NewStep("post form data").
			POST("/post").
			WithHeaders(map[string]string{"Content-Type": "application/x-www-form-urlencoded; charset=UTF-8"}).
			WithBody(map[string]interface{}{
				"foo1": "$varFoo1",       // reference former extracted variable
				"foo2": "${max($a, $b)}", // 12.3; step level variables are independent, variable b is 3.45 here
				"time": "${get_timestamp()}",
			}).
			Extract().
			WithJmesPath("body.form.time", "varTime").
			Validate().
			AssertEqual("status_code", 200, "check status code").
			AssertLengthEqual("body.form.foo1", 5, "check args foo1").
			AssertEqual("body.form.foo2", "12.3", "check args foo2"), // form data will be converted to string
		hrp.NewStep("get with timestamp").
			GET("/get").WithParams(map[string]interface{}{"time": "$varTime"}).
			Validate().
			AssertLengthEqual("body.args.time", 13, "check extracted var timestamp"),
	},
}

var demoTestCaseWithoutPlugin = &hrp.TestCase{
	Config: hrp.NewConfig("demo without custom function plugin").
		SetBaseURL("https://postman-echo.com").
		WithVariables(map[string]interface{}{ // global level variables
			"n":       5,
			"a":       12.3,
			"b":       3.45,
			"varFoo1": "${gen_random_string($n)}",
			"varFoo2": "${max($a, $b)}", // 12.3; eval with built-in function
		}),
	TestSteps: []hrp.IStep{
		hrp.NewStep("transaction 1 start").StartTransaction("tran1"), // start transaction
		hrp.NewStep("get with params").
			WithVariables(map[string]interface{}{ // step level variables
				"n":       3,                // inherit config level variables if not set in step level, a/varFoo1
				"b":       34.5,             // override config level variable if existed, n/b/varFoo2
				"varFoo2": "${max($a, $b)}", // 34.5; override variable b and eval again
				"name":    "get with params",
			}).
			GET("/get").
			WithParams(map[string]interface{}{"foo1": "$varFoo1", "foo2": "$varFoo2"}). // request with params
			WithHeaders(map[string]string{"User-Agent": "HttpRunnerPlus"}).             // request with headers
			Extract().
			WithJmesPath("body.args.foo1", "varFoo1"). // extract variable with jmespath
			Validate().
			AssertEqual("status_code", 200, "check response status code").        // validate response status code
			AssertStartsWith("headers.\"Content-Type\"", "application/json", ""). // validate response header
			AssertLengthEqual("body.args.foo1", 5, "check args foo1").            // validate response body with jmespath
			AssertLengthEqual("$varFoo1", 5, "check args foo1").                  // assert with extracted variable from current step
			AssertEqual("body.args.foo2", "34.5", "check args foo2"),             // notice: request params value will be converted to string
		hrp.NewStep("transaction 1 end").EndTransaction("tran1"), // end transaction
		hrp.NewStep("post json data").
			POST("/post").
			WithBody(map[string]interface{}{
				"foo1": "$varFoo1",       // reference former extracted variable
				"foo2": "${max($a, $b)}", // 12.3; step level variables are independent, variable b is 3.45 here
			}).
			Validate().
			AssertEqual("status_code", 200, "check status code").
			AssertLengthEqual("body.json.foo1", 5, "check args foo1").
			AssertEqual("body.json.foo2", 12.3, "check args foo2"),
		hrp.NewStep("post form data").
			POST("/post").
			WithHeaders(map[string]string{"Content-Type": "application/x-www-form-urlencoded; charset=UTF-8"}).
			WithBody(map[string]interface{}{
				"foo1": "$varFoo1",       // reference former extracted variable
				"foo2": "${max($a, $b)}", // 12.3; step level variables are independent, variable b is 3.45 here
				"time": "${get_timestamp()}",
			}).
			Extract().
			WithJmesPath("body.form.time", "varTime").
			Validate().
			AssertEqual("status_code", 200, "check status code").
			AssertLengthEqual("body.form.foo1", 5, "check args foo1").
			AssertEqual("body.form.foo2", "12.3", "check args foo2"), // form data will be converted to string
		hrp.NewStep("get with timestamp").
			GET("/get").WithParams(map[string]interface{}{"time": "$varTime"}).
			Validate().
			AssertLengthEqual("body.args.time", 13, "check extracted var timestamp"),
	},
}

// debugtalk.go
var demoGoPlugin = `package main

import (
	"fmt"

	"github.com/httprunner/funplugin/fungo"
)

func SumTwoInt(a, b int) int {
	return a + b
}

func SumInts(args ...int) int {
	var sum int
	for _, arg := range args {
		sum += arg
	}
	return sum
}

func Sum(args ...interface{}) (interface{}, error) {
	var sum float64
	for _, arg := range args {
		switch v := arg.(type) {
		case int:
			sum += float64(v)
		case float64:
			sum += v
		default:
			return nil, fmt.Errorf("unexpected type: %T", arg)
		}
	}
	return sum, nil
}

func SetupHookExample(args string) string {
	return fmt.Sprintf("step name: %v, setup...", args)
}

func TeardownHookExample(args string) string {
	return fmt.Sprintf("step name: %v, teardown...", args)
}

func main() {
	fungo.Register("sum_ints", SumInts)
	fungo.Register("sum_two_int", SumTwoInt)
	fungo.Register("sum", Sum)
	fungo.Register("setup_hook_example", SetupHookExample)
	fungo.Register("teardown_hook_example", TeardownHookExample)
	fungo.Serve()
}
`

// debugtalk.py
var demoPyPlugin = `import logging
from typing import List

import funppy


def sum(*args):
    result = 0
    for arg in args:
        result += arg
    return result

def sum_ints(*args: List[int]) -> int:
    result = 0
    for arg in args:
        result += arg
    return result

def sum_two_int(a: int, b: int) -> int:
    return a + b

def sum_two_string(a: str, b: str) -> str:
    return a + b

def sum_strings(*args: List[str]) -> str:
    result = ""
    for arg in args:
        result += arg
    return result

def concatenate(*args: List[str]) -> str:
    result = ""
    for arg in args:
        result += str(arg)
    return result

def setup_hook_example(name):
    logging.warning("setup_hook_example")
    return f"setup_hook_example: {name}"

def teardown_hook_example(name):
    logging.warning("teardown_hook_example")
    return f"teardown_hook_example: {name}"


if __name__ == '__main__':
    funppy.register("sum", sum)
    funppy.register("sum_ints", sum_ints)
    funppy.register("concatenate", concatenate)
    funppy.register("sum_two_int", sum_two_int)
    funppy.register("sum_two_string", sum_two_string)
    funppy.register("sum_strings", sum_strings)
    funppy.register("setup_hook_example", setup_hook_example)
    funppy.register("teardown_hook_example", teardown_hook_example)
    funppy.serve()
`

// .gitignore
var demoIgnoreContent = `.env
reports/
*.so
.vscode/
.idea/
.DS_Store
output/

# plugin
debugtalk.bin
debugtalk.so
`

// .env
var demoEnvContent = `USERNAME=debugtalk
PASSWORD=123456
`
