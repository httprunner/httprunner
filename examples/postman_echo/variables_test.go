package postman_echo

import (
	"testing"

	"github.com/httprunner/httpboomer"
)

func TestCaseConfigVariables(t *testing.T) {
	testcase := &httpboomer.TestCase{
		Config: httpboomer.TConfig{
			Name:    "run request with variables",
			BaseURL: "https://postman-echo.com",
			Variables: map[string]interface{}{
				"var1":               "bar1",
				"agent":              "HttpBoomer",
				"expectedStatusCode": 200,
			},
			Verify: false,
		},
		TestSteps: []httpboomer.IStep{
			httpboomer.Step("get with params").
				GET("/get").
				WithParams(map[string]interface{}{"foo1": "$var1", "foo2": "bar2"}).
				WithHeaders(map[string]string{"User-Agent": "$agent"}).
				Validate().
				AssertEqual("status_code", "$expectedStatusCode", "check status code").
				AssertEqual("headers.Connection", "keep-alive", "check header Connection").
				AssertEqual("headers.\"Content-Type\"", "application/json; charset=utf-8", "check header Content-Type").
				AssertEqual("body.args.foo1", "bar1", "check args foo1").
				AssertEqual("body.args.foo2", "bar2", "check args foo2").
				AssertEqual("body.headers.\"user-agent\"", "HttpBoomer", "check header user agent"),
		},
	}

	err := httpboomer.Test(t, testcase)
	if err != nil {
		t.Fatalf("run testcase error: %v", err)
	}
}

func TestCaseStepVariables(t *testing.T) {
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
				Validate().
				AssertEqual("status_code", "$expectedStatusCode", "check status code").
				AssertEqual("headers.Connection", "keep-alive", "check header Connection").
				AssertEqual("headers.\"Content-Type\"", "application/json; charset=utf-8", "check header Content-Type").
				AssertEqual("body.args.foo1", "bar1", "check args foo1").
				AssertEqual("body.args.foo2", "bar2", "check args foo2").
				AssertEqual("body.headers.\"user-agent\"", "HttpBoomer", "check header user agent"),
		},
	}

	err := httpboomer.Test(t, testcase)
	if err != nil {
		t.Fatalf("run testcase error: %v", err)
	}
}

func TestCaseOverrideConfigVariables(t *testing.T) {
	testcase := &httpboomer.TestCase{
		Config: httpboomer.TConfig{
			Name:    "run request with variables",
			BaseURL: "https://postman-echo.com",
			Variables: map[string]interface{}{
				"var1":               "bar0",
				"agent":              "HttpBoomer",
				"expectedStatusCode": 200,
			},
			Verify: false,
		},
		TestSteps: []httpboomer.IStep{
			httpboomer.Step("get with params").
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
				AssertEqual("headers.Connection", "keep-alive", "check header Connection").
				AssertEqual("headers.\"Content-Type\"", "application/json; charset=utf-8", "check header Content-Type").
				AssertEqual("body.args.foo1", "bar1", "check args foo1").
				AssertEqual("body.args.foo2", "bar2", "check args foo2").
				AssertEqual("body.headers.\"user-agent\"", "HttpBoomer", "check header user agent"),
		},
	}

	err := httpboomer.Test(t, testcase)
	if err != nil {
		t.Fatalf("run testcase error: %v", err)
	}
}
