package tests

import (
	"testing"

	"github.com/httprunner/httprunner/v4/hrp"
)

func TestRendezvous(t *testing.T) {
	testcase := &hrp.TestCase{
		Config: hrp.NewConfig("run request with rendezvous").
			SetBaseURL("https://postman-echo.com").
			WithVariables(map[string]interface{}{
				"n": 5,
				"a": 12.3,
				"b": 3.45,
			}),
		TestSteps: []hrp.IStep{
			hrp.NewStep("waiting for all users in the beginning").
				SetRendezvous("rendezvous0"),
			hrp.NewStep("rendezvous before get").
				SetRendezvous("rendezvous1").
				WithUserNumber(50).
				WithTimeout(3000),
			hrp.NewStep("get with params").
				GET("/get").
				WithParams(map[string]interface{}{"foo1": "foo1", "foo2": "foo2"}).
				WithHeaders(map[string]string{"User-Agent": "HttpRunnerPlus"}).
				Extract().
				WithJmesPath("body.args.foo1", "varFoo1").
				Validate().
				AssertEqual("status_code", 200, "check status code"),
			hrp.NewStep("rendezvous before post").
				SetRendezvous("rendezvous2").
				WithUserNumber(20).
				WithTimeout(2000),
			hrp.NewStep("post json data with functions").
				POST("/post").
				WithHeaders(map[string]string{"User-Agent": "HttpRunnerPlus"}).
				WithBody(map[string]interface{}{"foo1": "foo1", "foo2": "foo2"}).
				Validate().
				AssertEqual("status_code", 200, "check status code").
				AssertLengthEqual("body.json.foo1", 4, "check args foo1").
				AssertEqual("body.json.foo2", "foo2", "check args foo2"),
			hrp.NewStep("waiting for all users in the end").
				SetRendezvous("rendezvous3"),
		},
	}
	err := hrp.NewRunner(t).Run(testcase)
	if err != nil {
		t.Fatalf("run testcase error: %v", err)
	}
}
