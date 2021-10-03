package examples

import (
	"testing"

	"github.com/httprunner/httpboomer"
)

func TestCaseCallFunction(t *testing.T) {
	testcase := &httpboomer.TestCase{
		Config: httpboomer.TConfig{
			Name:    "run request with functions",
			BaseURL: "https://postman-echo.com",
			Verify:  false,
		},
		TestSteps: []httpboomer.IStep{
			httpboomer.Step("get with params").
				WithVariables(map[string]interface{}{
					"n": 5,
					"a": 12.3,
					"b": 3.45,
				}).
				GET("/get").
				WithParams(map[string]interface{}{"foo1": "${gen_random_string($n)}", "foo2": "${max($a, $b)}"}).
				WithHeaders(map[string]string{"User-Agent": "HttpBoomer"}).
				Extract().
				WithJmesPath("body.args.foo1", "varFoo1").
				Validate().
				AssertLengthEqual("body.args.foo1", 5, "check args foo1").
				AssertEqual("body.args.foo2", "12.3", "check args foo2"), // notice: request params value will be converted to string
		},
	}

	err := httpboomer.Test(t, testcase)
	if err != nil {
		t.Fatalf("run testcase error: %v", err)
	}
}
