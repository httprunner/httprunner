package hrp

import (
	"testing"
	"time"

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

func TestRunRequestStatOn(t *testing.T) {
	testcase := &TestCase{
		Config:    NewConfig("test").SetBaseURL("https://postman-echo.com"),
		TestSteps: []IStep{stepGET, stepPOSTData},
	}
	caseRunner, _ := NewRunner(t).SetHTTPStatOn().NewCaseRunner(testcase)
	sessionRunner := caseRunner.NewSession()
	if err := sessionRunner.Start(nil); err != nil {
		t.Fatal()
	}
	summary, _ := sessionRunner.GetSummary()

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

func TestRunCaseWithTimeout(t *testing.T) {
	r := NewRunner(t)

	// global timeout
	testcase1 := &TestCase{
		Config: NewConfig("TestCase1").
			SetTimeout(2 * time.Second). // set global timeout to 2s
			SetBaseURL("http://httpbin.org"),
		TestSteps: []IStep{
			NewStep("step1").
				GET("/delay/1").
				Validate().
				AssertEqual("status_code", 200, "check status code"),
		},
	}
	err := r.Run(testcase1)
	if !assert.NoError(t, err) { // assert no error
		t.FailNow()
	}

	testcase2 := &TestCase{
		Config: NewConfig("TestCase2").
			SetTimeout(2 * time.Second). // set global timeout to 2s
			SetBaseURL("http://httpbin.org"),
		TestSteps: []IStep{
			NewStep("step1").
				GET("/delay/3").
				Validate().
				AssertEqual("status_code", 200, "check status code"),
		},
	}
	err = r.Run(testcase2)
	if !assert.Error(t, err) { // assert error
		t.FailNow()
	}

	// step timeout
	testcase3 := &TestCase{
		Config: NewConfig("TestCase3").
			SetTimeout(2 * time.Second).
			SetBaseURL("http://httpbin.org"),
		TestSteps: []IStep{
			NewStep("step2").
				GET("/delay/3").
				SetTimeout(4*time.Second). // set step timeout to 4s
				Validate().
				AssertEqual("status_code", 200, "check status code"),
		},
	}
	err = r.Run(testcase3)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
}
