package tests

import (
	"path/filepath"
	"testing"

	hrp "github.com/httprunner/httprunner/v5"
)

const (
	hrpExamplesDir = "../examples/hrp"
)

// tmpl returns template file path
func tmpl(relativePath string) string {
	return filepath.Join("../internal/scaffold/templates/", relativePath)
}

var (
	demoTestCaseWithPluginJSONPath    = tmpl("testcases/demo_with_funplugin.json")
	demoTestCaseWithPluginYAMLPath    = tmpl("testcases/demo_with_funplugin.yaml")
	demoTestCaseWithoutPluginJSONPath = tmpl("testcases/demo_without_funplugin.json")
	demoTestCaseWithoutPluginYAMLPath = tmpl("testcases/demo_without_funplugin.yaml")
	demoTestCaseWithRefAPIPath        = tmpl("testcases/demo_ref_api.json")
	demoAPIGETPath                    = tmpl("/api/get.yml")
)

var demoTestCaseWithThinkTimePath hrp.TestCasePath = hrpExamplesDir + "/think_time_test.json"

var demoTestCaseWithPlugin = &hrp.TestCase{
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

func TestGenDemoTestCase(t *testing.T) {
	err := demoTestCaseWithPlugin.Dump2JSON(demoTestCaseWithPluginJSONPath)
	if err != nil {
		t.Fatal()
	}
	err = demoTestCaseWithPlugin.Dump2YAML(demoTestCaseWithPluginYAMLPath)
	if err != nil {
		t.Fatal()
	}

	err = demoTestCaseWithoutPlugin.Dump2JSON(demoTestCaseWithoutPluginJSONPath)
	if err != nil {
		t.Fatal()
	}
	err = demoTestCaseWithoutPlugin.Dump2YAML(demoTestCaseWithoutPluginYAMLPath)
	if err != nil {
		t.Fatal()
	}
}
