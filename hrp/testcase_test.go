package hrp

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
)

const (
	hrpExamplesDir = "../examples/hrp"
)

// tmpl returns template file path
func tmpl(relativePath string) string {
	return filepath.Join("internal/scaffold/templates/", relativePath)
}

var (
	demoTestCaseWithPluginJSONPath    = tmpl("testcases/demo_with_funplugin.json")
	demoTestCaseWithPluginYAMLPath    = tmpl("testcases/demo_with_funplugin.yaml")
	demoTestCaseWithoutPluginJSONPath = tmpl("testcases/demo_without_funplugin.json")
	demoTestCaseWithoutPluginYAMLPath = tmpl("testcases/demo_without_funplugin.yaml")
	demoTestCaseWithRefAPIPath        = tmpl("testcases/demo_ref_api.json")
	demoAPIGETPath                    = tmpl("/api/get.yml")
)

var demoTestCaseWithThinkTimePath TestCasePath = hrpExamplesDir + "/think_time_test.json"

var demoTestCaseWithPlugin = &TestCase{
	Config: NewConfig("demo with complex mechanisms").
		SetBaseURL("https://postman-echo.com").
		WithVariables(map[string]interface{}{ // global level variables
			"n":       "${sum_ints(1, 2, 2)}",
			"a":       "${sum(10, 2.3)}",
			"b":       3.45,
			"varFoo1": "${gen_random_string($n)}",
			"varFoo2": "${max($a, $b)}", // 12.3; eval with built-in function
		}),
	TestSteps: []IStep{
		NewStep("transaction 1 start").StartTransaction("tran1"), // start transaction
		NewStep("get with params").
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
		NewStep("transaction 1 end").EndTransaction("tran1"), // end transaction
		NewStep("post json data").
			POST("/post").
			WithBody(map[string]interface{}{
				"foo1": "$varFoo1",       // reference former extracted variable
				"foo2": "${max($a, $b)}", // 12.3; step level variables are independent, variable b is 3.45 here
			}).
			Validate().
			AssertEqual("status_code", 200, "check status code").
			AssertLengthEqual("body.json.foo1", 5, "check args foo1").
			AssertEqual("body.json.foo2", 12.3, "check args foo2"),
		NewStep("post form data").
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
		NewStep("get with timestamp").
			GET("/get").WithParams(map[string]interface{}{"time": "$varTime"}).
			Validate().
			AssertLengthEqual("body.args.time", 13, "check extracted var timestamp"),
	},
}

var demoTestCaseWithoutPlugin = &TestCase{
	Config: NewConfig("demo without custom function plugin").
		SetBaseURL("https://postman-echo.com").
		WithVariables(map[string]interface{}{ // global level variables
			"n":       5,
			"a":       12.3,
			"b":       3.45,
			"varFoo1": "${gen_random_string($n)}",
			"varFoo2": "${max($a, $b)}", // 12.3; eval with built-in function
		}),
	TestSteps: []IStep{
		NewStep("transaction 1 start").StartTransaction("tran1"), // start transaction
		NewStep("get with params").
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
		NewStep("transaction 1 end").EndTransaction("tran1"), // end transaction
		NewStep("post json data").
			POST("/post").
			WithBody(map[string]interface{}{
				"foo1": "$varFoo1",       // reference former extracted variable
				"foo2": "${max($a, $b)}", // 12.3; step level variables are independent, variable b is 3.45 here
			}).
			Validate().
			AssertEqual("status_code", 200, "check status code").
			AssertLengthEqual("body.json.foo1", 5, "check args foo1").
			AssertEqual("body.json.foo2", 12.3, "check args foo2"),
		NewStep("post form data").
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
		NewStep("get with timestamp").
			GET("/get").WithParams(map[string]interface{}{"time": "$varTime"}).
			Validate().
			AssertLengthEqual("body.args.time", 13, "check extracted var timestamp"),
	},
}

func TestGenDemoTestCase(t *testing.T) {
	tCase := demoTestCaseWithPlugin.ToTCase()
	err := builtin.Dump2JSON(tCase, demoTestCaseWithPluginJSONPath)
	if err != nil {
		t.Fatal()
	}
	err = builtin.Dump2YAML(tCase, demoTestCaseWithPluginYAMLPath)
	if err != nil {
		t.Fatal()
	}

	tCase = demoTestCaseWithoutPlugin.ToTCase()
	err = builtin.Dump2JSON(tCase, demoTestCaseWithoutPluginJSONPath)
	if err != nil {
		t.Fatal()
	}
	err = builtin.Dump2YAML(tCase, demoTestCaseWithoutPluginYAMLPath)
	if err != nil {
		t.Fatal()
	}
}

func TestLoadCase(t *testing.T) {
	tcJSON := &TCase{}
	tcYAML := &TCase{}
	err := builtin.LoadFile(demoTestCaseWithPluginJSONPath, tcJSON)
	if !assert.NoError(t, err) {
		t.Fatal()
	}
	err = builtin.LoadFile(demoTestCaseWithPluginYAMLPath, tcYAML)
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	if !assert.Equal(t, tcJSON.Config.Name, tcYAML.Config.Name) {
		t.Fatal()
	}
	if !assert.Equal(t, tcJSON.Config.BaseURL, tcYAML.Config.BaseURL) {
		t.Fatal()
	}
	if !assert.Equal(t, tcJSON.TestSteps[1].Name, tcYAML.TestSteps[1].Name) {
		t.Fatal()
	}
	if !assert.Equal(t, tcJSON.TestSteps[1].Request, tcYAML.TestSteps[1].Request) {
		t.Fatal()
	}
}

func TestConvertCheckExpr(t *testing.T) {
	exprs := []struct {
		before string
		after  string
	}{
		// normal check expression
		{"a.b.c", "a.b.c"},
		{"a.\"b-c\".d", "a.\"b-c\".d"},
		{"a.b-c.d", "a.b-c.d"},
		{"body.args.a[-1]", "body.args.a[-1]"},
		// check expression using regex
		{"covering (.*) testing,", "covering (.*) testing,"},
		{" (.*) a-b-c", " (.*) a-b-c"},
		// abnormal check expression
		{"headers.Content-Type", "headers.\"Content-Type\""},
		{"headers.\"Content-Type", "headers.\"Content-Type\""},
		{"headers.Content-Type\"", "headers.\"Content-Type\""},
		{"headers.User-Agent", "headers.\"User-Agent\""},
	}
	for _, expr := range exprs {
		if !assert.Equal(t, expr.after, convertJmespathExpr(expr.before)) {
			t.Fatal()
		}
	}
}
