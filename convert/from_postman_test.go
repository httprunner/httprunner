package convert

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var collectionPath = "../tests/data/postman/postman_collection.json"

func TestLoadCollection(t *testing.T) {
	casePostman, err := loadCasePostman(collectionPath)
	assert.NoError(t, err)
	assert.Equal(t, "postman collection demo", casePostman.Info.Name)
}

func TestMakeTestCaseFromCollection(t *testing.T) {
	tCase, err := LoadPostmanCase(collectionPath)
	assert.NoError(t, err)
	// check name
	assert.Equal(t, "postman collection demo", tCase.Config.Name)
	// check method
	assert.EqualValues(t, "GET", tCase.Steps[0].Request.Method)
	assert.EqualValues(t, "POST", tCase.Steps[1].Request.Method)
	// check url
	assert.Equal(t, "https://postman-echo.com/get", tCase.Steps[0].Request.URL)
	assert.Equal(t, "https://postman-echo.com/post", tCase.Steps[1].Request.URL)
	// check params
	assert.Equal(t, "v1", tCase.Steps[0].Request.Params["k1"])
	// check cookies (pass, postman collection doesn't contain cookies)
	// check headers
	assert.Equal(t, "application/x-www-form-urlencoded", tCase.Steps[2].Request.Headers["Content-Type"])
	assert.Equal(t, "application/json", tCase.Steps[3].Request.Headers["Content-Type"])
	assert.Equal(t, "text/plain", tCase.Steps[4].Request.Headers["Content-Type"])
	assert.Equal(t, "HttpRunner", tCase.Steps[5].Request.Headers["User-Agent"])
	// check body
	assert.Equal(t, nil, tCase.Steps[0].Request.Body)
	assert.Equal(t, map[string]string{"k1": "v1", "k2": "v2"}, tCase.Steps[2].Request.Body)
	assert.Equal(t, map[string]interface{}{"k1": "v1", "k2": "v2"}, tCase.Steps[3].Request.Body)
	assert.Equal(t, "have a nice day", tCase.Steps[4].Request.Body)
	assert.Equal(t, nil, tCase.Steps[5].Request.Body)
}
