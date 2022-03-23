package examples

import (
	"testing"

	"github.com/httprunner/hrp"
)

func TestCaseConfigVariables(t *testing.T) {
	testcase := &hrp.TestCase{
		Config: hrp.NewConfig("run request with variables").
			SetBaseURL("https://postman-echo.com").
			WithVariables(map[string]interface{}{
				"var1":               "bar1",
				"agent":              "HttpRunnerPlus",
				"expectedStatusCode": 200,
			}).SetVerifySSL(false),
		TestSteps: []hrp.IStep{
			hrp.NewStep("get with params").
				GET("/get").
				WithParams(map[string]interface{}{"foo1": "$var1", "foo2": "bar2"}).
				WithHeaders(map[string]string{"User-Agent": "$agent"}).
				Validate().
				AssertEqual("status_code", "$expectedStatusCode", "check status code").
				AssertEqual("headers.\"Content-Type\"", "application/json; charset=utf-8", "check header Content-Type").
				AssertEqual("body.args.foo1", "bar1", "check args foo1").
				AssertEqual("body.args.foo2", "bar2", "check args foo2").
				AssertEqual("body.headers.\"user-agent\"", "HttpRunnerPlus", "check header user agent"),
		},
	}

	err := hrp.NewRunner(t).Run(testcase)
	if err != nil {
		t.Fatalf("run testcase error: %v", err)
	}
}

func TestCaseStepVariables(t *testing.T) {
	testcase := &hrp.TestCase{
		Config: hrp.NewConfig("run request with variables").
			SetBaseURL("https://postman-echo.com").
			SetVerifySSL(false),
		TestSteps: []hrp.IStep{
			hrp.NewStep("get with params").
				WithVariables(map[string]interface{}{
					"var1":               "bar1",
					"agent":              "HttpRunnerPlus",
					"expectedStatusCode": 200,
				}).
				GET("/get").
				WithParams(map[string]interface{}{"foo1": "$var1", "foo2": "bar2"}).
				WithHeaders(map[string]string{"User-Agent": "$agent"}).
				Validate().
				AssertEqual("status_code", "$expectedStatusCode", "check status code").
				AssertEqual("headers.\"Content-Type\"", "application/json; charset=utf-8", "check header Content-Type").
				AssertEqual("body.args.foo1", "bar1", "check args foo1").
				AssertEqual("body.args.foo2", "bar2", "check args foo2").
				AssertEqual("body.headers.\"user-agent\"", "HttpRunnerPlus", "check header user agent"),
		},
	}

	err := hrp.NewRunner(t).Run(testcase)
	if err != nil {
		t.Fatalf("run testcase error: %v", err)
	}
}

func TestCaseOverrideConfigVariables(t *testing.T) {
	testcase := &hrp.TestCase{
		Config: hrp.NewConfig("run request with variables").
			SetBaseURL("https://postman-echo.com").
			WithVariables(map[string]interface{}{
				"var1":               "bar0",
				"agent":              "HttpRunnerPlus",
				"expectedStatusCode": 200,
			}).SetVerifySSL(false),
		TestSteps: []hrp.IStep{
			hrp.NewStep("get with params").
				WithVariables(map[string]interface{}{
					"var1":  "bar1",   // override config variable
					"agent": "$agent", // reference config variable
					// expectedStatusCode, inherit config variable
				}).
				GET("/get").
				WithParams(map[string]interface{}{"foo1": "$var1", "foo2": "bar2"}).
				WithHeaders(map[string]string{"User-Agent": "$agent"}).
				Validate().
				AssertEqual("status_code", "$expectedStatusCode", "check status code").
				AssertEqual("headers.\"Content-Type\"", "application/json; charset=utf-8", "check header Content-Type").
				AssertEqual("body.args.foo1", "bar1", "check args foo1").
				AssertEqual("body.args.foo2", "bar2", "check args foo2").
				AssertEqual("body.headers.\"user-agent\"", "HttpRunnerPlus", "check header user agent"),
		},
	}

	err := hrp.NewRunner(t).Run(testcase)
	if err != nil {
		t.Fatalf("run testcase error: %v", err)
	}
}

func TestCaseParseVariables(t *testing.T) {
	testcase := &hrp.TestCase{
		Config: hrp.NewConfig("run request with functions").
			SetBaseURL("https://postman-echo.com").
			WithVariables(map[string]interface{}{
				"n":       5,
				"a":       12.3,
				"b":       3.45,
				"varFoo1": "${gen_random_string($n)}",
				"varFoo2": "${max($a, $b)}", // 12.3
			}).SetVerifySSL(false),
		TestSteps: []hrp.IStep{
			hrp.NewStep("get with params").
				WithVariables(map[string]interface{}{
					"n":       3,
					"b":       34.5,
					"varFoo2": "${max($a, $b)}", // 34.5
				}).
				GET("/get").
				WithParams(map[string]interface{}{"foo1": "$varFoo1", "foo2": "$varFoo2"}).
				WithHeaders(map[string]string{"User-Agent": "HttpRunnerPlus"}).
				Extract().
				WithJmesPath("body.args.foo1", "varFoo1").
				Validate().
				AssertEqual("status_code", 200, "check status code").
				AssertLengthEqual("body.args.foo1", 5, "check args foo1").
				AssertEqual("body.args.foo2", "34.5", "check args foo2"), // notice: request params value will be converted to string
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
