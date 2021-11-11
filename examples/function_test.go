package examples

import (
	"testing"

	"github.com/httprunner/hrp"
)

func TestCaseCallFunction(t *testing.T) {
	testcase := &hrp.TestCase{
		Config: hrp.TConfig{
			Name:    "run request with functions",
			BaseURL: "https://postman-echo.com",
			Verify:  false,
			Variables: map[string]interface{}{
				"n": 5,
				"a": 12.3,
				"b": 3.45,
			},
		},
		TestSteps: []hrp.IStep{
			hrp.Step("get with params").
				GET("/get").
				WithParams(map[string]interface{}{"foo1": "${gen_random_string($n)}", "foo2": "${max($a, $b)}"}).
				WithHeaders(map[string]string{"User-Agent": "HttpRunnerPlus"}).
				Extract().
				WithJmesPath("body.args.foo1", "varFoo1").
				Validate().
				AssertEqual("status_code", 200, "check status code").
				AssertLengthEqual("body.args.foo1", 5, "check args foo1").
				AssertEqual("body.args.foo2", "12.3", "check args foo2"), // notice: request params value will be converted to string
			hrp.Step("post json data with functions").
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
