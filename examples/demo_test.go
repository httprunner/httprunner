package examples

import (
	"fmt"
	"testing"

	"github.com/httprunner/hrp"
)

var demoTestCase = &hrp.TestCase{
	Config: hrp.TConfig{
		Name:    "demo with complex mechanisms",
		BaseURL: "https://postman-echo.com",
		Variables: map[string]interface{}{ // global level variables
			"n":       5,
			"a":       12.3,
			"b":       3.45,
			"varFoo1": "${gen_random_string($n)}",
			"varFoo2": "${max($a, $b)}", // 12.3; eval with built-in function
		},
	},
	TestSteps: []hrp.IStep{
		hrp.Step("get with params").
			WithVariables(map[string]interface{}{ // step level variables
				"n":       3,                // inherit config level variables if not set in step level, a/varFoo1
				"b":       34.5,             // override config level variable if existed, n/b/varFoo2
				"varFoo2": "${max($a, $b)}", // 34.5; override variable b and eval again
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
		hrp.Step("post json data").
			POST("/post").
			WithBody(map[string]interface{}{
				"foo1": "$varFoo1",       // reference former extracted variable
				"foo2": "${max($a, $b)}", // 12.3; step level variables are independent, variable b is 3.45 here
			}).
			Validate().
			AssertEqual("status_code", 200, "check status code").
			AssertLengthEqual("body.json.foo1", 5, "check args foo1").
			AssertEqual("body.json.foo2", 12.3, "check args foo2"),
		hrp.Step("post form data").
			POST("/post").
			WithHeaders(map[string]string{"Content-Type": "application/x-www-form-urlencoded; charset=UTF-8"}).
			WithBody(map[string]interface{}{
				"foo1": "$varFoo1",       // reference former extracted variable
				"foo2": "${max($a, $b)}", // 12.3; step level variables are independent, variable b is 3.45 here
			}).
			Validate().
			AssertEqual("status_code", 200, "check status code").
			AssertLengthEqual("body.form.foo1", 5, "check args foo1").
			AssertEqual("body.form.foo2", "12.3", "check args foo2"), // form data will be converted to string
	},
}

var (
	demoTestCaseJSONPath = "demo.json"
	demoTestCaseYAMLPath = "demo.yaml"
)

func TestGenDemoTestCase(t *testing.T) {
	tCase, _ := demoTestCase.ToTCase()
	err := tCase.Dump2JSON(demoTestCaseJSONPath)
	if err != nil {
		t.Fail()
	}
	err = tCase.Dump2YAML(demoTestCaseYAMLPath)
	if err != nil {
		t.Fail()
	}
}

func Example_TestCase() {
	err := hrp.NewRunner(nil).Run(demoTestCase) // hrp.Run(demoTestCase)
	fmt.Println(err)
	// Output:
	// <nil>
}

func Example_JSONTestCase() {
	testCase := &hrp.TestCasePath{Path: demoTestCaseJSONPath}
	err := hrp.NewRunner(nil).Run(testCase) // hrp.Run(testCase)
	fmt.Println(err)
	// Output:
	// <nil>
}

func Example_YAMTestCase() {
	testCase := &hrp.TestCasePath{Path: demoTestCaseYAMLPath}
	err := hrp.NewRunner(nil).Run(testCase) // hrp.Run(testCase)
	fmt.Println(err)
	// Output:
	// <nil>
}
