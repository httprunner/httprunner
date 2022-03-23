package examples

import (
	"testing"

	"github.com/httprunner/hrp"
)

func TestCaseCallFunction(t *testing.T) {
	testcase := &hrp.TestCase{
		Config: hrp.NewConfig("run request with functions").
			SetBaseURL("https://postman-echo.com").
			WithVariables(map[string]interface{}{
				"n": 5,
				"a": 12.3,
				"b": 3.45,
			}).
			SetVerifySSL(false),
		TestSteps: []hrp.IStep{
			hrp.NewStep("get with params").
				GET("/get").
				WithParams(map[string]interface{}{"foo1": "${gen_random_string($n)}", "foo2": "${max($a, $b)}", "foo3": "Foo3"}).
				WithHeaders(map[string]string{"User-Agent": "HttpRunnerPlus"}).
				Extract().
				WithJmesPath("body.args.foo1", "varFoo1").
				Validate().
				AssertEqual("status_code", 200, "check status code").
				AssertLengthEqual("body.args.foo1", 5, "check args foo1").
				AssertEqual("body.args.foo2", "12.3", "check args foo2").
				AssertTypeMatch("body.args.foo3", "str", "check args foo3 is type string").
				AssertStringEqual("body.args.foo3", "foo3", "check args foo3 case-insensitivity").
				AssertContains("body.args.foo3", "Foo", "check contains ").
				AssertContainedBy("body.args.foo3", "this is Foo3 test", "check contained by"), // notice: request params value will be converted to string
			hrp.NewStep("post json data with functions").
				POST("/post").
				WithHeaders(map[string]string{"User-Agent": "HttpRunnerPlus"}).
				WithBody(map[string]interface{}{"foo1": "${gen_random_string($n)}", "foo2": "${max($a, $b)}"}).
				Validate().
				AssertEqual("status_code", 200, "check status code").
				AssertLengthEqual("body.json.foo1", 5, "check args foo1").
				AssertEqual("body.json.foo2", 12.3, "check args foo2"),
		},
	}

	err := hrp.NewRunner(t).Run(testcase)
	if err != nil {
		t.Fatalf("run testcase error: %v", err)
	}
}
