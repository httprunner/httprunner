package examples

import (
	"testing"

	"github.com/httprunner/hrp"
)

func TestCaseBasicRequest(t *testing.T) {
	testcase := &hrp.TestCase{
		Config: hrp.TConfig{
			Name:    "request methods testcase in hardcode",
			BaseURL: "https://postman-echo.com",
			Verify:  false,
		},
		TestSteps: []hrp.IStep{
			hrp.Step("get with params").
				GET("/get").
				WithParams(map[string]interface{}{"foo1": "bar1", "foo2": "bar2"}).
				WithHeaders(map[string]string{
					"User-Agent": "HttpRunnerPlus",
				}).
				Validate().
				AssertEqual("status_code", 200, "check status code").
				AssertEqual("headers.Connection", "keep-alive", "check header Connection").
				AssertEqual("headers.\"Content-Type\"", "application/json; charset=utf-8", "check header Content-Type").
				AssertEqual("body.args.foo1", "bar1", "check args foo1").
				AssertEqual("body.args.foo2", "bar2", "check args foo2"),
			hrp.Step("post raw text").
				POST("/post").
				WithHeaders(map[string]string{
					"User-Agent":   "HttpRunnerPlus",
					"Content-Type": "text/plain",
				}).
				WithBody("This is expected to be sent back as part of response body.").
				Validate().
				AssertEqual("status_code", 200, "check status code").
				AssertEqual("body.data", "This is expected to be sent back as part of response body.", "check data"),
			hrp.Step("post form data").
				POST("/post").
				WithHeaders(map[string]string{
					"User-Agent":   "HttpRunnerPlus",
					"Content-Type": "application/x-www-form-urlencoded",
				}).
				WithBody(map[string]interface{}{"foo1": "bar1", "foo2": "bar2"}).
				Validate().
				AssertEqual("status_code", 200, "check status code").
				AssertEqual("body.form.foo1", "bar1", "check form foo1").
				AssertEqual("body.form.foo2", "bar2", "check form foo2"),
			hrp.Step("post json data").
				POST("/post").
				WithHeaders(map[string]string{
					"User-Agent": "HttpRunnerPlus",
				}).
				WithBody(map[string]interface{}{"foo1": "bar1", "foo2": "bar2"}).
				Validate().
				AssertEqual("status_code", 200, "check status code").
				AssertEqual("body.json.foo1", "bar1", "check json foo1").
				AssertEqual("body.json.foo2", "bar2", "check json foo2"),
			hrp.Step("put request").
				PUT("/put").
				WithHeaders(map[string]string{
					"User-Agent":   "HttpRunnerPlus",
					"Content-Type": "text/plain",
				}).
				WithBody("This is expected to be sent back as part of response body.").
				Validate().
				AssertEqual("status_code", 200, "check status code").
				AssertEqual("body.data", "This is expected to be sent back as part of response body.", "check data"),
		},
	}

	err := hrp.Run(t, testcase)
	if err != nil {
		t.Fatalf("run testcase error: %v", err)
	}
}
