package hrp

import (
	"testing"
)

var (
	stepGET = NewStep("get with params").
		GET("/get").
		WithParams(map[string]interface{}{"foo1": "bar1", "foo2": "bar2"}).
		WithHeaders(map[string]string{"User-Agent": "HttpRunnerPlus"}).
		WithCookies(map[string]string{"user": "debugtalk"}).
		Validate().
		AssertEqual("status_code", 200, "check status code").
		AssertEqual("headers.\"Content-Type\"", "application/json; charset=utf-8", "check header Content-Type").
		AssertEqual("body.args.foo1", "bar1", "check param foo1").
		AssertEqual("body.args.foo2", "bar2", "check param foo2")
	stepPOSTData = NewStep("post form data").
			POST("/post").
			WithParams(map[string]interface{}{"foo1": "bar1", "foo2": "bar2"}).
			WithHeaders(map[string]string{"User-Agent": "HttpRunnerPlus", "Content-Type": "application/x-www-form-urlencoded"}).
			WithBody("a=1&b=2").
			WithCookies(map[string]string{"user": "debugtalk"}).
			Validate().
			AssertEqual("status_code", 200, "check status code")
)

func TestRunRequestGetToStruct(t *testing.T) {
	tStep := stepGET.step
	if tStep.Request.Method != httpGET {
		t.Fatalf("tStep.Request.Method != GET")
	}
	if tStep.Request.URL != "/get" {
		t.Fatalf("tStep.Request.URL != '/get'")
	}
	if tStep.Request.Params["foo1"] != "bar1" || tStep.Request.Params["foo2"] != "bar2" {
		t.Fatalf("tStep.Request.Params mismatch")
	}
	if tStep.Request.Headers["User-Agent"] != "HttpRunnerPlus" {
		t.Fatalf("tStep.Request.Headers mismatch")
	}
	if tStep.Request.Cookies["user"] != "debugtalk" {
		t.Fatalf("tStep.Request.Cookies mismatch")
	}
	validator, ok := tStep.Validators[0].(Validator)
	if !ok || validator.Check != "status_code" || validator.Expect != 200 {
		t.Fatalf("tStep.Validators mismatch")
	}
}

func TestRunRequestPostDataToStruct(t *testing.T) {
	tStep := stepPOSTData.step
	if tStep.Request.Method != httpPOST {
		t.Fatalf("tStep.Request.Method != POST")
	}
	if tStep.Request.URL != "/post" {
		t.Fatalf("tStep.Request.URL != '/post'")
	}
	if tStep.Request.Params["foo1"] != "bar1" || tStep.Request.Params["foo2"] != "bar2" {
		t.Fatalf("tStep.Request.Params mismatch")
	}
	if tStep.Request.Headers["User-Agent"] != "HttpRunnerPlus" {
		t.Fatalf("tStep.Request.Headers mismatch")
	}
	if tStep.Request.Cookies["user"] != "debugtalk" {
		t.Fatalf("tStep.Request.Cookies mismatch")
	}
	if tStep.Request.Body != "a=1&b=2" {
		t.Fatalf("tStep.Request.Data mismatch")
	}
	validator, ok := tStep.Validators[0].(Validator)
	if !ok || validator.Check != "status_code" || validator.Expect != 200 {
		t.Fatalf("tStep.Validators mismatch")
	}
}

func TestRunRequestRun(t *testing.T) {
	testcase := &TestCase{
		Config:    NewConfig("test").SetBaseURL("https://postman-echo.com"),
		TestSteps: []IStep{stepGET, stepPOSTData},
	}
	runner := NewRunner(t).SetRequestsLogOn().newCaseRunner(testcase)
	if _, err := runner.runStep(0, testcase.Config); err != nil {
		t.Fatalf("tStep.Run() error: %s", err)
	}
	if _, err := runner.runStep(1, testcase.Config); err != nil {
		t.Fatalf("tStepPOSTData.Run() error: %s", err)
	}
}
