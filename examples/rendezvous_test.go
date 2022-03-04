package examples

import (
	"testing"

	"github.com/httprunner/hrp"
	"github.com/httprunner/hrp/internal/builtin"
)

const rendezvousTestJSONPath = "rendezvous_test.json"

var rendezvousTestcase = &hrp.TestCase{
	Config: hrp.NewConfig("run request with functions").
		SetBaseURL("https://postman-echo.com").
		WithVariables(map[string]interface{}{
			"n": 5,
			"a": 12.3,
			"b": 3.45,
		}),
	TestSteps: []hrp.IStep{
		hrp.NewStep("waiting for all users in the beginning").
			Rendezvous("rendezvous0"),
		hrp.NewStep("rendezvous before get").
			Rendezvous("rendezvous1").
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
			Rendezvous("rendezvous2").
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
			Rendezvous("rendezvous3"),
	},
}

func TestRendezvous(t *testing.T) {
	err := hrp.NewRunner(t).Run(rendezvousTestcase)
	if err != nil {
		t.Fatalf("run testcase error: %v", err)
	}
}

func TestRendezvousDump2JSON(t *testing.T) {
	tCase, err := rendezvousTestcase.ToTCase()
	if err != nil {
		t.Fatalf("ToTCase error: %v", err)
	}
	err = builtin.Dump2JSON(tCase, rendezvousTestJSONPath)
	if err != nil {
		t.Fatalf("dump to json error: %v", err)
	}
}
