package hrp

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
	runner := NewRunner(t).SetRequestsLogOn()
	sessionRunner, _ := runner.NewSessionRunner(testcase)

	if _, err := stepGET.Run(sessionRunner); err != nil {
		t.Fatalf("stepGET.Run() error: %v", err)
	}
	if _, err := stepPOSTData.Run(sessionRunner); err != nil {
		t.Fatalf("stepPOSTData.Run() error: %v", err)
	}
}

func TestRunRequestStatOn(t *testing.T) {
	testcase := &TestCase{
		Config:    NewConfig("test").SetBaseURL("https://postman-echo.com"),
		TestSteps: []IStep{stepGET, stepPOSTData},
	}
	runner := NewRunner(t).SetHTTPStatOn()
	sessionRunner, _ := runner.NewSessionRunner(testcase)
	if err := sessionRunner.Start(nil); err != nil {
		t.Fatal()
	}
	summary := sessionRunner.GetSummary()

	stat := summary.Records[0].HttpStat
	if !assert.GreaterOrEqual(t, stat["DNSLookup"], int64(0)) {
		t.Fatal()
	}
	if !assert.Greater(t, stat["TCPConnection"], int64(0)) {
		t.Fatal()
	}
	if !assert.Greater(t, stat["TLSHandshake"], int64(0)) {
		t.Fatal()
	}
	if !assert.Greater(t, stat["ServerProcessing"], int64(1)) {
		t.Fatal()
	}
	if !assert.GreaterOrEqual(t, stat["ContentTransfer"], int64(0)) {
		t.Fatal()
	}
	if !assert.GreaterOrEqual(t, stat["NameLookup"], int64(0)) {
		t.Fatal()
	}
	if !assert.Greater(t, stat["Connect"], int64(0)) {
		t.Fatal()
	}
	if !assert.Greater(t, stat["Pretransfer"], int64(0)) {
		t.Fatal()
	}
	if !assert.Greater(t, stat["StartTransfer"], int64(0)) {
		t.Fatal()
	}
	if !assert.Greater(t, stat["Total"], int64(5)) {
		t.Fatal()
	}
	if !assert.Less(t, stat["Total"]-summary.Records[0].Elapsed, int64(3)) {
		t.Fatal()
	}

	// reuse connection
	stat = summary.Records[1].HttpStat
	if !assert.Equal(t, int64(0), stat["DNSLookup"]) {
		t.Fatal()
	}
	if !assert.Equal(t, int64(0), stat["TCPConnection"]) {
		t.Fatal()
	}
	if !assert.Equal(t, int64(0), stat["TLSHandshake"]) {
		t.Fatal()
	}
	if !assert.Greater(t, stat["ServerProcessing"], int64(1)) {
		t.Fatal()
	}
	if !assert.Equal(t, int64(0), stat["NameLookup"]) {
		t.Fatal()
	}
	if !assert.Equal(t, int64(0), stat["Connect"]) {
		t.Fatal()
	}
	if !assert.Equal(t, int64(0), stat["Pretransfer"]) {
		t.Fatal()
	}
	if !assert.Greater(t, stat["StartTransfer"], int64(0)) {
		t.Fatal()
	}
	if !assert.Greater(t, stat["Total"], int64(1)) {
		t.Fatal()
	}
	if !assert.Less(t, stat["Total"]-summary.Records[0].Elapsed, int64(3)) {
		t.Fatal()
	}
}
