package tests

import (
	"testing"

	"github.com/httprunner/httprunner/hrp"
)

func TestProtocol(t *testing.T) {
	testcase := &hrp.TestCase{
		Config: hrp.NewConfig("run request with different protocol types").
			SetBaseURL("https://postman-echo.com"),
		TestSteps: []hrp.IStep{
			hrp.NewStep("HTTP/1.1 get").
				GET("/get").
				WithParams(map[string]interface{}{"foo1": "foo1", "foo2": "foo2"}).
				WithHeaders(map[string]string{"User-Agent": "HttpRunnerPlus"}).
				Validate().
				AssertEqual("status_code", 200, "check status code").
				AssertEqual("proto", "HTTP/1.1", "check protocol type").
				AssertLengthEqual("body.args.foo1", 4, "check param foo1"),
			hrp.NewStep("HTTP/1.1 post").
				POST("/post").
				WithHeaders(map[string]string{"User-Agent": "HttpRunnerPlus"}).
				WithBody(map[string]interface{}{"foo1": "foo1", "foo2": "foo2"}).
				Validate().
				AssertEqual("status_code", 200, "check status code").
				AssertEqual("proto", "HTTP/1.1", "check protocol type").
				AssertLengthEqual("body.json.foo1", 4, "check body foo1"),
			hrp.NewStep("HTTP2.0 get").
				GET("/get").
				EnableHTTP2().
				WithParams(map[string]interface{}{"foo1": "foo1", "foo2": "foo2"}).
				WithHeaders(map[string]string{"User-Agent": "HttpRunnerPlus"}).
				Validate().
				AssertEqual("status_code", 200, "check status code").
				AssertEqual("proto", "HTTP/2.0", "check protocol type").
				AssertLengthEqual("body.args.foo1", 4, "check param foo1"),
			hrp.NewStep("HTTP/2.0 post").
				POST("/post").
				EnableHTTP2().
				WithHeaders(map[string]string{"User-Agent": "HttpRunnerPlus"}).
				WithBody(map[string]interface{}{"foo1": "foo1", "foo2": "foo2"}).
				Validate().
				AssertEqual("status_code", 200, "check status code").
				AssertEqual("proto", "HTTP/2.0", "check protocol type").
				AssertLengthEqual("body.json.foo1", 4, "check body foo1"),
		},
	}
	err := hrp.NewRunner(t).Run(testcase)
	if err != nil {
		t.Fatalf("run testcase error: %v", err)
	}
}
