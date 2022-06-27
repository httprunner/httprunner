package convert

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var collectionPath = "../../../examples/data/postman/postman_collection.json"

func TestLoadCollection(t *testing.T) {
	casePostman, err := loadCasePostman(collectionPath)
	if !assert.NoError(t, err) {
		t.Fatal(err)
	}
	if !assert.Equal(t, "postman collection demo", casePostman.Info.Name) {
		t.Fatal()
	}
}

func TestMakeTestCaseFromCollection(t *testing.T) {
	tCase, err := LoadPostmanCase(collectionPath)
	if !assert.NoError(t, err) {
		t.Fatal()
	}
	// check name
	if !assert.Equal(t, "postman collection demo", tCase.Config.Name) {
		t.Fatal()
	}
	// check method
	if !assert.EqualValues(t, "GET", tCase.TestSteps[0].Request.Method) {
		t.Fatal()
	}
	if !assert.EqualValues(t, "POST", tCase.TestSteps[1].Request.Method) {
		t.Fatal()
	}
	// check url
	if !assert.Equal(t, "https://postman-echo.com/get", tCase.TestSteps[0].Request.URL) {
		t.Fatal()
	}
	if !assert.Equal(t, "https://postman-echo.com/post", tCase.TestSteps[1].Request.URL) {
		t.Fatal()
	}
	// check params
	if !assert.Equal(t, "v1", tCase.TestSteps[0].Request.Params["k1"]) {
		t.Fatal()
	}
	// check cookies (pass, postman collection doesn't contain cookies)
	// check headers
	if !assert.Equal(t, "application/x-www-form-urlencoded", tCase.TestSteps[2].Request.Headers["Content-Type"]) {
		t.Fatal()
	}
	if !assert.Equal(t, "application/json", tCase.TestSteps[3].Request.Headers["Content-Type"]) {
		t.Fatal()
	}
	if !assert.Equal(t, "text/plain", tCase.TestSteps[4].Request.Headers["Content-Type"]) {
		t.Fatal()
	}
	if !assert.Equal(t, "HttpRunner", tCase.TestSteps[5].Request.Headers["User-Agent"]) {
		t.Fatal()
	}
	// check body
	if !assert.Equal(t, nil, tCase.TestSteps[0].Request.Body) {
		t.Fatal()
	}
	if !assert.Equal(t, map[string]string{"k1": "v1", "k2": "v2"}, tCase.TestSteps[2].Request.Body) {
		t.Fatal()
	}
	if !assert.Equal(t, map[string]interface{}{"k1": "v1", "k2": "v2"}, tCase.TestSteps[3].Request.Body) {
		t.Fatal()
	}
	if !assert.Equal(t, "have a nice day", tCase.TestSteps[4].Request.Body) {
		t.Fatal()
	}
	if !assert.Equal(t, nil, tCase.TestSteps[5].Request.Body) {
		t.Fatal()
	}
}
