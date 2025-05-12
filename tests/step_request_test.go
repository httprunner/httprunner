package tests

import (
	"testing"
	"time"

	hrp "github.com/httprunner/httprunner/v5"
	"github.com/stretchr/testify/assert"
)

var (
	stepGET = hrp.NewStep("get with params").
		GET("/get").
		WithParams(map[string]interface{}{"foo1": "bar1", "foo2": "bar2"}).
		WithHeaders(map[string]string{"User-Agent": "HttpRunnerPlus"}).
		WithCookies(map[string]string{"user": "debugtalk"}).
		Validate().
		AssertEqual("status_code", 200, "check status code").
		AssertEqual("headers.\"Content-Type\"", "application/json; charset=utf-8", "check header Content-Type").
		AssertEqual("body.args.foo1", "bar1", "check param foo1").
		AssertEqual("body.args.foo2", "bar2", "check param foo2")
	stepPOSTData = hrp.NewStep("post form data").
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
	assert.Equal(t, tStep.Request.Method, hrp.HTTP_GET)
	assert.Equal(t, tStep.Request.URL, "/get")

	assert.Equal(t, tStep.Request.Params["foo1"], "bar1")
	assert.Equal(t, tStep.Request.Params["foo2"], "bar2")

	assert.Equal(t, tStep.Request.Headers["User-Agent"], "HttpRunnerPlus")
	assert.Equal(t, tStep.Request.Cookies["user"], "debugtalk")

	validator, ok := tStep.Validators[0].(hrp.Validator)
	if !ok || validator.Check != "status_code" || validator.Expect != 200 {
		t.Fatalf("tStep.Validators mismatch")
	}
}

func TestRunRequestPostDataToStruct(t *testing.T) {
	tStep := stepPOSTData
	assert.Equal(t, tStep.Request.Method, hrp.HTTP_POST)
	assert.Equal(t, tStep.Request.URL, "/post")

	assert.Equal(t, tStep.Request.Params["foo1"], "bar1")
	assert.Equal(t, tStep.Request.Params["foo2"], "bar2")

	assert.Equal(t, tStep.Request.Headers["User-Agent"], "HttpRunnerPlus")
	assert.Equal(t, tStep.Request.Cookies["user"], "debugtalk")
	assert.Equal(t, tStep.Request.Body, "a=1&b=2")

	validator, ok := tStep.Validators[0].(hrp.Validator)
	if !ok || validator.Check != "status_code" || validator.Expect != 200 {
		t.Fatalf("tStep.Validators mismatch")
	}
}

func TestRunRequestStatOn(t *testing.T) {
	testcase := hrp.TestCase{
		Config:    hrp.NewConfig("test").SetBaseURL("https://postman-echo.com"),
		TestSteps: []hrp.IStep{stepGET, stepPOSTData},
	}
	caseRunner, _ := hrp.NewCaseRunner(testcase, hrp.NewRunner(t).SetHTTPStatOn())
	sessionRunner := caseRunner.NewSession()
	summary, err := sessionRunner.Start(nil)
	assert.Nil(t, err)

	stat := summary.Records[0].HttpStat
	assert.GreaterOrEqual(t, stat["DNSLookup"], int64(0))
	assert.Greater(t, stat["TCPConnection"], int64(0))
	assert.Greater(t, stat["TLSHandshake"], int64(0))
	assert.Greater(t, stat["ServerProcessing"], int64(0))
	assert.GreaterOrEqual(t, stat["ContentTransfer"], int64(0))
	assert.GreaterOrEqual(t, stat["NameLookup"], int64(0))
	assert.Greater(t, stat["Connect"], int64(0))
	assert.Greater(t, stat["Pretransfer"], int64(0))
	assert.Greater(t, stat["StartTransfer"], int64(0))
	assert.Greater(t, stat["Total"], int64(5))
	assert.Less(t, stat["Total"]-summary.Records[0].Elapsed, int64(3))

	// reuse connection
	stat = summary.Records[1].HttpStat
	assert.Equal(t, int64(0), stat["DNSLookup"])
	assert.Equal(t, int64(0), stat["TCPConnection"])
	assert.Equal(t, int64(0), stat["TLSHandshake"])
	assert.Greater(t, stat["ServerProcessing"], int64(1))
	assert.Equal(t, int64(0), stat["NameLookup"])
	assert.Equal(t, int64(0), stat["Connect"])
	assert.Equal(t, int64(0), stat["Pretransfer"])
	assert.Greater(t, stat["StartTransfer"], int64(0))
	assert.Greater(t, stat["Total"], int64(1))
	assert.Less(t, stat["Total"]-summary.Records[0].Elapsed, int64(100))
}

func TestRunCaseWithTimeout(t *testing.T) {
	r := hrp.NewRunner(t)

	// global timeout
	testcase1 := &hrp.TestCase{
		Config: hrp.NewConfig("TestCase1").
			SetRequestTimeout(10). // set global timeout to 10s
			SetBaseURL("https://postman-echo.com"),
		TestSteps: []hrp.IStep{
			hrp.NewStep("step1").
				GET("/delay/1").
				Validate().
				AssertEqual("status_code", 200, "check status code"),
		},
	}
	err := r.Run(testcase1)
	assert.Nil(t, err)

	testcase2 := &hrp.TestCase{
		Config: hrp.NewConfig("TestCase2").
			SetRequestTimeout(5). // set global timeout to 10s
			SetBaseURL("https://postman-echo.com"),
		TestSteps: []hrp.IStep{
			hrp.NewStep("step1").
				GET("/delay/10").
				Validate().
				AssertEqual("status_code", 200, "check status code"),
		},
	}
	err = r.Run(testcase2)
	assert.Error(t, err)

	// step timeout
	testcase3 := &hrp.TestCase{
		Config: hrp.NewConfig("TestCase3").
			SetRequestTimeout(10).
			SetBaseURL("https://postman-echo.com"),
		TestSteps: []hrp.IStep{
			hrp.NewStep("step2").
				GET("/delay/11").
				SetTimeout(15*time.Second). // set step timeout to 4s
				Validate().
				AssertEqual("status_code", 200, "check status code"),
		},
	}
	err = r.Run(testcase3)
	assert.Nil(t, err)
}
