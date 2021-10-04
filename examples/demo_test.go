package examples

import (
	"testing"

	"github.com/httprunner/httpboomer"
)

func TestCaseDemo(t *testing.T) {
	testcase := &httpboomer.TestCase{
		Config: httpboomer.TConfig{
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
		TestSteps: []httpboomer.IStep{
			httpboomer.Step("get with params").
				WithVariables(map[string]interface{}{ // step level variables
					"n":       3,                // inherit config level variables if not set in step level, a/varFoo1
					"b":       34.5,             // override config level variable if existed, n/b/varFoo2
					"varFoo2": "${max($a, $b)}", // 34.5; override variable b and eval again
				}).
				GET("/get").
				WithParams(map[string]interface{}{"foo1": "$varFoo1", "foo2": "$varFoo2"}). // request with params
				WithHeaders(map[string]string{"User-Agent": "HttpBoomer"}).                 // request with headers
				Extract().
				WithJmesPath("body.args.foo1", "varFoo1"). // extract variable with jmespath
				Validate().
				AssertEqual("status_code", 200, "check response status code").        // validate response status code
				AssertStartsWith("headers.\"Content-Type\"", "application/json", ""). // validate response header
				AssertLengthEqual("body.args.foo1", 5, "check args foo1").            // validate response body with jmespath
				AssertLengthEqual("$varFoo1", 5, "check args foo1").                  // assert with extracted variable from current step
				AssertEqual("body.args.foo2", "34.5", "check args foo2"),             // notice: request params value will be converted to string
			httpboomer.Step("post json data").
				POST("/post").
				WithJSON(map[string]interface{}{
					"foo1": "$varFoo1",       // reference former extracted variable
					"foo2": "${max($a, $b)}", // 12.3; step level variables are independent, variable b is 3.45 here
				}).
				Validate().
				AssertEqual("status_code", 200, "check status code").
				AssertLengthEqual("body.json.foo1", 5, "check args foo1").
				AssertEqual("body.json.foo2", 12.3, "check args foo2"),
		},
	}

	err := httpboomer.Test(t, testcase)
	if err != nil {
		t.Fatalf("run testcase error: %v", err)
	}
}
