package httpboomer

import (
	"testing"
)

var (
	tStepGET = RunRequest("get with params").
			GET("/get").
			WithParams(Params{"foo1": "bar1", "foo2": "bar2"}).
			WithHeaders(Headers{"User-Agent": "HttpBoomer"}).
			WithCookies(Cookies{"user": "debugtalk"}).
			Validate().
			AssertEqual("status_code", 200, "check status code")
	tStepPOSTData = RunRequest("post form data").
			POST("/post").
			WithParams(Params{"foo1": "bar1", "foo2": "bar2"}).
			WithHeaders(Headers{"User-Agent": "HttpBoomer", "Content-Type": "application/x-www-form-urlencoded"}).
			WithData("a=1&b=2").
			WithCookies(Cookies{"user": "debugtalk"}).
			Validate().
			AssertEqual("status_code", 200, "check status code")
)

func TestRunRequestGetToStruct(t *testing.T) {
	tStep := tStepGET.ToStruct()
	if tStep.Request.Method != GET {
		t.Fatalf("tStep.Request.Method != GET")
	}
	if tStep.Request.URL != "/get" {
		t.Fatalf("tStep.Request.URL != '/get'")
	}
	if tStep.Request.Params["foo1"] != "bar1" || tStep.Request.Params["foo2"] != "bar2" {
		t.Fatalf("tStep.Request.Params mismatch")
	}
	if tStep.Request.Headers["User-Agent"] != "HttpBoomer" {
		t.Fatalf("tStep.Request.Headers mismatch")
	}
	if tStep.Request.Cookies["user"] != "debugtalk" {
		t.Fatalf("tStep.Request.Cookies mismatch")
	}
	if tStep.Validators[0].Check != "status_code" || tStep.Validators[0].Expect != 200 {
		t.Fatalf("tStep.Validators mismatch")
	}
}

func TestRunRequestPostDataToStruct(t *testing.T) {
	tStep := tStepPOSTData.ToStruct()
	if tStep.Request.Method != POST {
		t.Fatalf("tStep.Request.Method != POST")
	}
	if tStep.Request.URL != "/post" {
		t.Fatalf("tStep.Request.URL != '/post'")
	}
	if tStep.Request.Params["foo1"] != "bar1" || tStep.Request.Params["foo2"] != "bar2" {
		t.Fatalf("tStep.Request.Params mismatch")
	}
	if tStep.Request.Headers["User-Agent"] != "HttpBoomer" {
		t.Fatalf("tStep.Request.Headers mismatch")
	}
	if tStep.Request.Cookies["user"] != "debugtalk" {
		t.Fatalf("tStep.Request.Cookies mismatch")
	}
	if tStep.Request.Data != "a=1&b=2" {
		t.Fatalf("tStep.Request.Data mismatch")
	}
	if tStep.Validators[0].Check != "status_code" || tStep.Validators[0].Expect != 200 {
		t.Fatalf("tStep.Validators mismatch")
	}
}

func TestRunRequestRun(t *testing.T) {
	if err := tStepGET.Run(); err != nil {
		t.Fatalf("tStep.Run() error: %s", err)
	}
	if err := tStepPOSTData.Run(); err != nil {
		t.Fatalf("tStepPOSTData.Run() error: %s", err)
	}
}
