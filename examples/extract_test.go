package examples

import (
	"testing"

	"github.com/httprunner/httpboomer"
)

// reference extracted variables for validation in the same step
func TestCaseExtractStepSingle(t *testing.T) {
	testcase := &httpboomer.TestCase{
		Config: httpboomer.TConfig{
			Name:    "run request with variables",
			BaseURL: "https://postman-echo.com",
			Verify:  false,
		},
		TestSteps: []httpboomer.IStep{
			httpboomer.Step("get with params").
				WithVariables(map[string]interface{}{
					"var1":               "bar1",
					"agent":              "HttpBoomer",
					"expectedStatusCode": 200,
				}).
				GET("/get").
				WithParams(map[string]interface{}{"foo1": "$var1", "foo2": "bar2"}).
				WithHeaders(map[string]string{"User-Agent": "$agent"}).
				Extract().
				WithJmesPath("status_code", "statusCode").
				WithJmesPath("headers.\"Content-Type\"", "contentType").
				WithJmesPath("body.args.foo1", "varFoo1").
				Validate().
				AssertEqual("$statusCode", "$expectedStatusCode", "check status code").                      // assert with extracted variable from current step
				AssertEqual("$contentType", "application/json; charset=utf-8", "check header Content-Type"). // assert with extracted variable from current step
				AssertEqual("$varFoo1", "bar1", "check args foo1").                                          // assert with extracted variable from current step
				AssertEqual("body.args.foo2", "bar2", "check args foo2").
				AssertEqual("body.headers.\"user-agent\"", "HttpBoomer", "check header user agent"),
		},
	}

	err := httpboomer.Run(t, testcase)
	if err != nil {
		t.Fatalf("run testcase error: %v", err)
	}
}

// reference extracted variables from previous step
func TestCaseExtractStepAssociation(t *testing.T) {
	testcase := &httpboomer.TestCase{
		Config: httpboomer.TConfig{
			Name:    "run request with variables",
			BaseURL: "https://postman-echo.com",
			Verify:  false,
		},
		TestSteps: []httpboomer.IStep{
			httpboomer.Step("get with params").
				WithVariables(map[string]interface{}{
					"var1":  "bar1",
					"agent": "HttpBoomer",
				}).
				GET("/get").
				WithParams(map[string]interface{}{"foo1": "$var1", "foo2": "bar2"}).
				WithHeaders(map[string]string{"User-Agent": "$agent"}).
				Extract().
				WithJmesPath("status_code", "statusCode").
				WithJmesPath("headers.\"Content-Type\"", "contentType").
				WithJmesPath("body.args.foo1", "varFoo1").
				Validate().
				AssertEqual("$statusCode", 200, "check status code").
				AssertEqual("headers.Connection", "keep-alive", "check header Connection").
				AssertEqual("$contentType", "application/json; charset=utf-8", "check header Content-Type").
				AssertEqual("$varFoo1", "bar1", "check args foo1").
				AssertEqual("body.args.foo2", "bar2", "check args foo2").
				AssertEqual("body.headers.\"user-agent\"", "HttpBoomer", "check header user agent"),
			httpboomer.Step("post json data").
				POST("/post").
				WithHeaders(map[string]string{"User-Agent": "HttpBoomer"}).
				WithBody(map[string]interface{}{"foo1": "bar1", "foo2": "bar2"}).
				Validate().
				AssertEqual("status_code", "$statusCode", "check status code"). // assert with extracted variable from previous step
				AssertEqual("$varFoo1", "bar1", "check json foo1").             // assert with extracted variable from previous step
				AssertEqual("body.json.foo2", "bar2", "check json foo2"),
		},
	}

	err := httpboomer.Run(t, testcase)
	if err != nil {
		t.Fatalf("run testcase error: %v", err)
	}
}
