package examples

import (
	"testing"

	"github.com/httprunner/hrp"
)

func TestCaseValidateStep(t *testing.T) {
	testcase := &hrp.TestCase{
		Config: hrp.NewConfig("run request with validation").
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
				Extract().
				WithJmesPath("body.args.foo1", "varFoo1").
				Validate().
				AssertEqual("status_code", "$expectedStatusCode", "check status code").                                  // assert status code
				AssertEqual("headers.\"Content-Type\"", "application/json; charset=utf-8", "check header Content-Type"). // assert response header, with double quotes
				AssertEqual("body.args.foo1", "bar1", "check args foo1").                                                // assert response json body with jmespath
				AssertEqual("body.args.foo2", "bar2", "check args foo2").
				AssertEqual("body.headers.\"user-agent\"", "HttpRunnerPlus", "check header user agent"),
			hrp.NewStep("get with params").
				WithVariables(map[string]interface{}{
					"var1":  "bar1",
					"agent": "HttpRunnerPlus",
				}).
				GET("/get").
				WithParams(map[string]interface{}{"foo1": "$var1", "foo2": "bar2"}).
				WithHeaders(map[string]string{"User-Agent": "$agent"}).
				Extract().
				WithJmesPath("status_code", "statusCode").
				WithJmesPath("headers.\"Content-Type\"", "contentType").
				Validate().
				AssertEqual("$statusCode", 200, "check status code").                                        // assert with extracted variable from current step
				AssertEqual("$contentType", "application/json; charset=utf-8", "check header Content-Type"). // assert with extracted variable from current step
				AssertEqual("$varFoo1", "bar1", "check args foo1").                                          // assert with extracted variable from previous step
				AssertEqual("body.args.foo2", "bar2", "check args foo2"),                                    // assert response json body with jmespath
		},
	}

	err := hrp.NewRunner(t).Run(testcase)
	if err != nil {
		t.Fatalf("run testcase error: %v", err)
	}
}
