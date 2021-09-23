package postman_echo

import (
	"testing"

	"github.com/httprunner/httpboomer"
)

func TestCaseHardcode(t *testing.T) {
	testcase := &httpboomer.TestCase{
		Config: httpboomer.TConfig{
			Name:    "request methods testcase in hardcode",
			BaseURL: "https://postman-echo.com",
			Verify:  false,
		},
		TestSteps: []httpboomer.IStep{
			httpboomer.Step("get with params").
				GET("/get").
				WithParams(map[string]interface{}{"foo1": "bar1", "foo2": "bar2"}).
				WithHeaders(map[string]string{"User-Agent": "HttpBoomer"}).
				Validate().
				AssertEqual("status_code", 200, "check status code"),
			httpboomer.Step("post raw text").
				POST("/post").
				WithHeaders(map[string]string{"User-Agent": "HttpBoomer", "Content-Type": "text/plain"}).
				WithData("This is expected to be sent back as part of response body.").
				Validate().
				AssertEqual("status_code", 200, "check status code"),
		},
	}

	err := httpboomer.Test(testcase)
	if err != nil {
		t.Fatalf("run testcase error: %v", err)
	}
}
