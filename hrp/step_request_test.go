package hrp

import (
	"io"
	"net/http"
	"strings"
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
	tStep := stepGET
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
	tStep := stepPOSTData
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
	testcase := TestCase{
		Config:    NewConfig("test").SetBaseURL("https://postman-echo.com"),
		TestSteps: []IStep{stepGET, stepPOSTData},
	}
	caseRunner, _ := NewRunner(t).SetHTTPStatOn().NewCaseRunner(testcase)
	sessionRunner := caseRunner.NewSession()
	summary, err := sessionRunner.Start(nil)
	if err != nil {
		t.Fatal()
	}

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
	if !assert.Greater(t, stat["ServerProcessing"], int64(0)) {
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
	if !assert.Less(t, stat["Total"]-summary.Records[0].Elapsed, int64(100)) {
		t.Fatal()
	}
}

func TestRunCaseWithTimeout(t *testing.T) {
	r := NewRunner(t)

	// global timeout
	testcase1 := &TestCase{
		Config: NewConfig("TestCase1").
			SetRequestTimeout(10). // set global timeout to 10s
			SetBaseURL("https://postman-echo.com"),
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
			SetRequestTimeout(5). // set global timeout to 10s
			SetBaseURL("https://postman-echo.com"),
		TestSteps: []IStep{
			NewStep("step1").
				GET("/delay/10").
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
			SetRequestTimeout(10).
			SetBaseURL("https://postman-echo.com"),
		TestSteps: []IStep{
			NewStep("step2").
				GET("/delay/11").
				SetTimeout(15*time.Second). // set step timeout to 4s
				Validate().
				AssertEqual("status_code", 200, "check status code"),
		},
	}
	err = r.Run(testcase3)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
}

func TestSearchJmespath(t *testing.T) {
	testText := `{"a": {"b": "foo"}, "c": "bar", "d": {"e": [{"f": "foo"}, {"f": "bar"}]}}`
	testData := []struct {
		raw      string
		expected string
	}{
		{"body.a.b", "foo"},
		{"body.c", "bar"},
		{"body.d.e[0].f", "foo"},
		{"body.d.e[1].f", "bar"},
	}
	resp := http.Response{}
	resp.Body = io.NopCloser(strings.NewReader(testText))
	respObj, err := newHttpResponseObject(t, newParser(), &resp)
	if err != nil {
		t.Fatal()
	}
	for _, data := range testData {
		if !assert.Equal(t, data.expected, respObj.searchJmespath(data.raw)) {
			t.Fatal()
		}
	}
}

func TestSearchRegexp(t *testing.T) {
	testText := `
	<ul class="nav navbar-nav navbar-right">
	<li><a href="/order/addToCart" style="color: white"><i class="fa fa-shopping-cart fa-2x"></i><span class="badge">0</span></a></li>
	<li class="dropdown">
	  <a class="dropdown-toggle" data-toggle="dropdown" href="#" style="color: white">
		Leo   <i class="fa fa-cog fa-2x"></i><span class="caret"></span></a>
	  <ul class="dropdown-menu">
		<li><a href="/user/changePassword">Change Password</a></li>
		<li><a href="/user/addAddress">Shipping</a></li>
		<li><a href="/user/addCard">Payment</a></li>
		<li><a href="/order/orderHistory">Order History</a></li>
		<li><a href="/user/signOut">Sign Out</a></li>
	  </ul>
	</li>

	<li>&nbsp;&nbsp;&nbsp;</li>
	<li><a href="/user/signOut" style="color: white"><i class="fa fa-sign-out fa-2x"></i>
	  Sign Out</a></li>
  </ul>
`
	testData := []struct {
		raw      string
		expected string
	}{
		{"/user/signOut\">(.*)</a></li>", "Sign Out"},
		{"<li><a href=\"/user/(.*)\" style", "signOut"},
		{"		(.*)   <i class=\"fa fa-cog fa-2x\"></i>", "Leo"},
	}
	// new response object
	resp := http.Response{}
	resp.Body = io.NopCloser(strings.NewReader(testText))
	respObj, err := newHttpResponseObject(t, newParser(), &resp)
	if err != nil {
		t.Fatal()
	}
	for _, data := range testData {
		if !assert.Equal(t, data.expected, respObj.searchRegexp(data.raw)) {
			t.Fatal()
		}
	}
}
