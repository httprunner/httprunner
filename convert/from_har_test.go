package convert

import (
	"testing"

	hrp "github.com/httprunner/httprunner/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var harPath = "../tests/data/har/demo.har"

func TestLoadHAR(t *testing.T) {
	caseHAR, err := loadCaseHAR(harPath)
	require.NoError(t, err)
	assert.Equal(t, "GET", caseHAR.Log.Entries[0].Request.Method)
	assert.Equal(t, "POST", caseHAR.Log.Entries[1].Request.Method)
}

func TestLoadTCaseFromHAR(t *testing.T) {
	tCase, err := LoadHARCase(harPath)
	assert.NoError(t, err)

	// make request method
	assert.EqualValues(t, "GET", tCase.Steps[0].Request.Method)
	assert.EqualValues(t, "POST", tCase.Steps[1].Request.Method)

	// make request url
	assert.Equal(t, "https://postman-echo.com/get", tCase.Steps[0].Request.URL)
	assert.Equal(t, "https://postman-echo.com/post", tCase.Steps[1].Request.URL)

	// make request params
	assert.Equal(t, "HDnY8", tCase.Steps[0].Request.Params["foo1"])

	// make request cookies
	assert.NotEmpty(t, tCase.Steps[1].Request.Cookies["sails.sid"])

	// make request headers
	assert.Equal(t, "HttpRunnerPlus", tCase.Steps[0].Request.Headers["User-Agent"])
	assert.Equal(t, "postman-echo.com", tCase.Steps[0].Request.Headers["Host"])

	// make request data
	assert.Equal(t, nil, tCase.Steps[0].Request.Body)
	assert.Equal(t, map[string]interface{}{"foo1": "HDnY8", "foo2": 12.3}, tCase.Steps[1].Request.Body)
	assert.Equal(t, map[string]string{"foo1": "HDnY8", "foo2": "12.3"}, tCase.Steps[2].Request.Body)

	// make validators
	validator, ok := tCase.Steps[0].Validators[0].(hrp.Validator)
	if !ok || !assert.Equal(t, "status_code", validator.Check) {
		t.Fatal()
	}
	validator, ok = tCase.Steps[0].Validators[1].(hrp.Validator)
	if !ok || !assert.Equal(t, "headers.\"Content-Type\"", validator.Check) {
		t.Fatal()
	}
	validator, ok = tCase.Steps[0].Validators[2].(hrp.Validator)
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
	caseHar, err := loadCaseHAR(harPath)
	require.NoError(t, err)
	step, err := caseHar.prepareTestStep(entry)
	assert.NoError(t, err)

	assert.Equal(t, "http://127.0.0.1:8080/api/login", step.Request.URL)
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
	caseHar, err := loadCaseHAR(harPath)
	require.NoError(t, err)
	step, err := caseHar.prepareTestStep(entry)
	assert.NoError(t, err)

	assert.Equal(t, map[string]string{
		"Content-Type": "application/json; charset=utf-8",
	}, step.Request.Headers)
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
	caseHar, err := loadCaseHAR(harPath)
	require.NoError(t, err)
	step, err := caseHar.prepareTestStep(entry)
	assert.NoError(t, err)

	assert.Equal(t, map[string]string{
		"abc":      "123",
		"UserName": "leolee",
	}, step.Request.Cookies)
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
	caseHar, err := loadCaseHAR(harPath)
	require.NoError(t, err)
	step, err := caseHar.prepareTestStep(entry)
	assert.NoError(t, err)

	assert.Equal(t, map[string]string{"a": "1", "b": "2"}, step.Request.Body)
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
	caseHar, err := loadCaseHAR(harPath)
	require.NoError(t, err)
	step, err := caseHar.prepareTestStep(entry)
	assert.NoError(t, err)

	assert.Equal(t, map[string]interface{}{"a": "1", "b": "2"}, step.Request.Body)
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
	caseHar, err := loadCaseHAR(harPath)
	require.NoError(t, err)
	step, err := caseHar.prepareTestStep(entry)
	assert.NoError(t, err)
	assert.Equal(t, nil, step.Request.Body)
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
	caseHar, err := loadCaseHAR(harPath)
	require.NoError(t, err)
	step, err := caseHar.prepareTestStep(entry)
	assert.NoError(t, err)
	validator, ok := step.Validators[0].(hrp.Validator)
	assert.True(t, ok)
	assert.Equal(t, validator,
		hrp.Validator{
			Check:   "status_code",
			Expect:  200,
			Assert:  "equals",
			Message: "assert response status code",
		})

	validator, ok = step.Validators[1].(hrp.Validator)
	assert.True(t, ok)
	assert.Equal(t, validator,
		hrp.Validator{
			Check:   "headers.\"Content-Type\"",
			Expect:  "application/json; charset=utf-8",
			Assert:  "equals",
			Message: "assert response header Content-Type",
		})

	validator, ok = step.Validators[2].(hrp.Validator)
	assert.True(t, ok)
	assert.Equal(t, validator,
		hrp.Validator{
			Check:   "body.Code",
			Expect:  float64(200), // TODO
			Assert:  "equals",
			Message: "assert response body Code",
		})
}
