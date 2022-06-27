package convert

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/httprunner/httprunner/v4/hrp"
)

var harPath = "../../../examples/data/har/demo.har"

var caseHar *CaseHar

func init() {
	caseHar, _ = loadCaseHAR(harPath)
}

func TestLoadHAR(t *testing.T) {
	caseHAR, err := loadCaseHAR(harPath)
	if !assert.NoError(t, err) {
		t.Fatal()
	}
	if !assert.Equal(t, "GET", caseHAR.Log.Entries[0].Request.Method) {
		t.Fatal()
	}
	if !assert.Equal(t, "POST", caseHAR.Log.Entries[1].Request.Method) {
		t.Fatal()
	}
}

func TestLoadTCaseFromHAR(t *testing.T) {
	tCase, err := LoadHARCase(harPath)
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	// make request method
	if !assert.EqualValues(t, "GET", tCase.TestSteps[0].Request.Method) {
		t.Fatal()
	}
	if !assert.EqualValues(t, "POST", tCase.TestSteps[1].Request.Method) {
		t.Fatal()
	}

	// make request url
	if !assert.Equal(t, "https://postman-echo.com/get", tCase.TestSteps[0].Request.URL) {
		t.Fatal()
	}
	if !assert.Equal(t, "https://postman-echo.com/post", tCase.TestSteps[1].Request.URL) {
		t.Fatal()
	}

	// make request params
	if !assert.Equal(t, "HDnY8", tCase.TestSteps[0].Request.Params["foo1"]) {
		t.Fatal()
	}

	// make request cookies
	if !assert.NotEmpty(t, tCase.TestSteps[1].Request.Cookies["sails.sid"]) {
		t.Fatal()
	}

	// make request headers
	if !assert.Equal(t, "HttpRunnerPlus", tCase.TestSteps[0].Request.Headers["User-Agent"]) {
		t.Fatal()
	}
	if !assert.Equal(t, "postman-echo.com", tCase.TestSteps[0].Request.Headers["Host"]) {
		t.Fatal()
	}

	// make request data
	if !assert.Equal(t, nil, tCase.TestSteps[0].Request.Body) {
		t.Fatal()
	}
	if !assert.Equal(t, map[string]interface{}{"foo1": "HDnY8", "foo2": 12.3}, tCase.TestSteps[1].Request.Body) {
		t.Fatal()
	}
	if !assert.Equal(t, map[string]string{"foo1": "HDnY8", "foo2": "12.3"}, tCase.TestSteps[2].Request.Body) {
		t.Fatal()
	}

	// make validators
	validator, ok := tCase.TestSteps[0].Validators[0].(hrp.Validator)
	if !ok || !assert.Equal(t, "status_code", validator.Check) {
		t.Fatal()
	}
	validator, ok = tCase.TestSteps[0].Validators[1].(hrp.Validator)
	if !ok || !assert.Equal(t, "headers.\"Content-Type\"", validator.Check) {
		t.Fatal()
	}
	validator, ok = tCase.TestSteps[0].Validators[2].(hrp.Validator)
	if !ok || !assert.Equal(t, "body.url", validator.Check) {
		t.Fatal()
	}
}

func TestMakeRequestURL(t *testing.T) {
	entry := &Entry{
		Request: Request{
			URL: "http://127.0.0.1:8080/api/login",
		},
	}
	step, err := caseHar.prepareTestStep(entry)
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	if !assert.Equal(t, "http://127.0.0.1:8080/api/login", step.Request.URL) {
		t.Fatal()
	}
}

func TestMakeRequestHeaders(t *testing.T) {
	entry := &Entry{
		Request: Request{
			Method: "POST",
			Headers: []NVP{
				{Name: "Content-Type", Value: "application/json; charset=utf-8"},
			},
		},
	}
	step, err := caseHar.prepareTestStep(entry)
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	if !assert.Equal(t, map[string]string{
		"Content-Type": "application/json; charset=utf-8",
	}, step.Request.Headers) {
		t.Fatal()
	}
}

func TestMakeRequestCookies(t *testing.T) {
	entry := &Entry{
		Request: Request{
			Method: "POST",
			Cookies: []Cookie{
				{Name: "abc", Value: "123"},
				{Name: "UserName", Value: "leolee"},
			},
		},
	}
	step, err := caseHar.prepareTestStep(entry)
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	if !assert.Equal(t, map[string]string{
		"abc":      "123",
		"UserName": "leolee",
	}, step.Request.Cookies) {
		t.Fatal()
	}
}

func TestMakeRequestDataParams(t *testing.T) {
	entry := &Entry{
		Request: Request{
			Method: "POST",
			PostData: PostData{
				MimeType: "application/x-www-form-urlencoded; charset=utf-8",
				Params: []PostParam{
					{Name: "a", Value: "1"},
					{Name: "b", Value: "2"},
				},
			},
		},
	}
	step, err := caseHar.prepareTestStep(entry)
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	if !assert.Equal(t, map[string]string{"a": "1", "b": "2"}, step.Request.Body) {
		t.Fatal()
	}
}

func TestMakeRequestDataJSON(t *testing.T) {
	entry := &Entry{
		Request: Request{
			Method: "POST",
			PostData: PostData{
				MimeType: "application/json; charset=utf-8",
				Text:     "{\"a\":\"1\",\"b\":\"2\"}",
			},
		},
	}
	step, err := caseHar.prepareTestStep(entry)
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	if !assert.Equal(t, map[string]interface{}{"a": "1", "b": "2"}, step.Request.Body) {
		t.Fatal()
	}
}

func TestMakeRequestDataTextEmpty(t *testing.T) {
	entry := &Entry{
		Request: Request{
			Method: "POST",
			PostData: PostData{
				MimeType: "application/json; charset=utf-8",
				Text:     "",
			},
		},
	}
	step, err := caseHar.prepareTestStep(entry)
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	if !assert.Equal(t, nil, step.Request.Body) { // TODO
		t.Fatal()
	}
}

func TestMakeValidate(t *testing.T) {
	entry := &Entry{
		Response: Response{
			Status: 200,
			Headers: []NVP{
				{Name: "Content-Type", Value: "application/json; charset=utf-8"},
			},
			Content: Content{
				Size:     71,
				MimeType: "application/json; charset=utf-8",
				// map[Code:200 IsSuccess:true Message:<nil> Value:map[BlnResult:true]]
				Text:     "eyJJc1N1Y2Nlc3MiOnRydWUsIkNvZGUiOjIwMCwiTWVzc2FnZSI6bnVsbCwiVmFsdWUiOnsiQmxuUmVzdWx0Ijp0cnVlfX0=",
				Encoding: "base64",
			},
		},
	}
	step, err := caseHar.prepareTestStep(entry)
	if !assert.NoError(t, err) {
		t.Fatal()
	}
	validator, ok := step.Validators[0].(hrp.Validator)
	if !ok {
		t.Fatal()
	}
	if !assert.Equal(t, validator,
		hrp.Validator{
			Check:   "status_code",
			Expect:  200,
			Assert:  "equals",
			Message: "assert response status code",
		}) {
		t.Fatal()
	}

	validator, ok = step.Validators[1].(hrp.Validator)
	if !ok {
		t.Fatal()
	}
	if !assert.Equal(t, validator,
		hrp.Validator{
			Check:   "headers.\"Content-Type\"",
			Expect:  "application/json; charset=utf-8",
			Assert:  "equals",
			Message: "assert response header Content-Type",
		}) {
		t.Fatal()
	}

	validator, ok = step.Validators[2].(hrp.Validator)
	if !ok {
		t.Fatal()
	}
	if !assert.Equal(t, validator,
		hrp.Validator{
			Check:   "body.Code",
			Expect:  float64(200), // TODO
			Assert:  "equals",
			Message: "assert response body Code",
		}) {
		t.Fatal()
	}
}
