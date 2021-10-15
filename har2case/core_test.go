package har2case

import (
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var harPath string

func TestMain(m *testing.M) {
	harPath = "../examples/demo.har"

	// run all tests
	code := m.Run()
	defer os.Exit(code)

	// teardown
}

func TestGenJSON(t *testing.T) {
	jsonPath, err := NewHAR(harPath).GenJSON()
	log.Printf("jsonPath: %v, err: %v", jsonPath, err)
	if !assert.NoError(t, err) {
		t.Fail()
	}
	if !assert.NotEmpty(t, jsonPath) {
		t.Fail()
	}
}

func TestLoadHAR(t *testing.T) {
	har := NewHAR(harPath)
	h, err := har.load()
	if !assert.NoError(t, err) {
		t.Fail()
	}
	if !assert.Equal(t, "GET", h.Log.Entries[0].Request.Method) {
		t.Fail()
	}
	if !assert.Equal(t, "POST", h.Log.Entries[1].Request.Method) {
		t.Fail()
	}
}

func TestMakeTestCase(t *testing.T) {
	har := NewHAR(harPath)
	tCase, err := har.makeTestCase()
	if !assert.NoError(t, err) {
		t.Fail()
	}

	// make request method
	if !assert.EqualValues(t, "GET", tCase.TestSteps[0].Request.Method) {
		t.Fail()
	}
	if !assert.EqualValues(t, "POST", tCase.TestSteps[1].Request.Method) {
		t.Fail()
	}

	// make request url
	if !assert.Equal(t, "https://postman-echo.com/get", tCase.TestSteps[0].Request.URL) {
		t.Fail()
	}
	if !assert.Equal(t, "https://postman-echo.com/post", tCase.TestSteps[1].Request.URL) {
		t.Fail()
	}

	// make request params
	if !assert.Equal(t, "HDnY8", tCase.TestSteps[0].Request.Params["foo1"]) {
		t.Fail()
	}

	// make request cookies
	if !assert.NotEmpty(t, tCase.TestSteps[1].Request.Cookies["sails.sid"]) {
		t.Fail()
	}

	// make request headers
	if !assert.Equal(t, "HttpBoomer", tCase.TestSteps[0].Request.Headers["User-Agent"]) {
		t.Fail()
	}
	if !assert.Equal(t, "postman-echo.com", tCase.TestSteps[0].Request.Headers["Host"]) {
		t.Fail()
	}

	// make request data
	if !assert.Equal(t, map[string]interface{}{"foo1": "HDnY8", "foo2": 12.3}, tCase.TestSteps[1].Request.Body) {
		t.Fail()
	}
}
