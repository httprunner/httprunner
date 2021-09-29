package postman_echo

import (
	"testing"

	"github.com/httprunner/httpboomer"
)

func TestCaseVariables(t *testing.T) {
	testcase := &httpboomer.TestCase{
		Config: httpboomer.TConfig{
			Name:    "run request with variables",
			BaseURL: "https://postman-echo.com",
			Verify:  false,
		},
		TestSteps: []httpboomer.IStep{
			httpboomer.Step("get with params").
				WithVariables(map[string]interface{}{"var1": "bar1", "expectedStatusCode": 200}).
				GET("/get").
				WithParams(map[string]interface{}{"foo1": "$var1", "foo2": "bar2"}).
				WithHeaders(map[string]string{"User-Agent": "HttpBoomer"}).
				Validate().
				AssertEqual("status_code", "$expectedStatusCode", "check status code").
				AssertEqual("headers.Connection", "keep-alive", "check header Connection").
				AssertEqual("headers.\"Content-Type\"", "application/json; charset=utf-8", "check header Content-Type").
				AssertEqual("body.args.foo1", "$var1", "check args foo1").
				AssertEqual("body.args.foo2", "bar2", "check args foo2"),
		},
	}

	err := httpboomer.Test(t, testcase)
	if err != nil {
		t.Fatalf("run testcase error: %v", err)
	}
}
