package convert

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	collectionPath                = "../../../examples/data/postman/postman_collection.json"
	collectionProfileOverridePath = "../../../examples/data/postman/profile_override.yml"
	collectionProfilePath         = "../../../examples/data/postman/profile.yml"
)

var converterPostman = NewConverterPostman(NewTCaseConverter(collectionPath))

func TestPostman2JSON(t *testing.T) {
	jsonPath, err := converterPostman.ToJSON()
	if !assert.NoError(t, err) {
		t.Fatal()
	}
	if !assert.NotEmpty(t, jsonPath) {
		t.Fatal()
	}
}

func TestPostman2YAML(t *testing.T) {
	yamlPath, err := converterPostman.ToYAML()
	if !assert.NoError(t, err) {
		t.Fatal()
	}
	if !assert.NotEmpty(t, yamlPath) {
		t.Fatal()
	}
}

func TestLoadCollection(t *testing.T) {
	casePostman, err := converterPostman.load()
	if !assert.NoError(t, err) {
		t.Fatal(err)
	}
	if !assert.Equal(t, "postman collection demo", casePostman.Info.Name) {
		t.Fatal()
	}
}

func TestMakeTestCaseFromCollection(t *testing.T) {
	tCase, err := converterPostman.makeTestCase()
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
	// check cookies (pass, postman collection doesn't contains cookies)
	// check headers
	if !assert.Contains(t, tCase.TestSteps[1].Request.Headers["Content-Type"], "multipart/form-data") {
		t.Fatal()
	}
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
	if !assert.NotEmpty(t, tCase.TestSteps[1].Request.Body) {
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

func TestMakeTestCaseWithProfileOverride(t *testing.T) {
	tCaseConverter := NewTCaseConverter(collectionPath)
	tCaseConverter.SetProfile(collectionProfileOverridePath)
	c := NewConverterPostman(tCaseConverter)
	tCase, err := c.makeTestCase()
	if !assert.NoError(t, err) {
		t.Fatal()
	}
	for _, step := range tCase.TestSteps {
		if step.Request.Method == "GET" && !assert.Len(t, step.Request.Headers, 1) {
			t.Fatal()
		}
		if step.Request.Method == "POST" && !assert.Len(t, step.Request.Headers, 2) {
			t.Fatal()
		}
		if !assert.Equal(t, "all original headers will be overridden", step.Request.Headers["Header1"]) {
			t.Fatal()
		}
		if !assert.Len(t, step.Request.Cookies, 1) {
			t.Fatal()
		}
		if !assert.Equal(t, "all original cookies will be overridden", step.Request.Cookies["Cookie1"]) {
			t.Fatal()
		}
	}
}

func TestMakeTestCaseWithProfile(t *testing.T) {
	tCaseConverter := NewTCaseConverter(collectionPath)
	tCaseConverter.SetProfile(collectionProfilePath)
	c := NewConverterPostman(tCaseConverter)
	tCase, err := c.makeTestCase()
	if !assert.NoError(t, err) {
		t.Fatal()
	}
	// create cookies Cookie1 indicated in profile
	if !assert.Equal(t, "this cookie will be created or updated", tCase.TestSteps[0].Request.Cookies["Cookie1"]) {
		t.Fatal()
	}
	// update header User-Agent indicated in profile
	if !assert.Equal(t, "this header will be created or updated", tCase.TestSteps[5].Request.Headers["User-Agent"]) {
		t.Fatal()
	}
	// pass header Connection which is not indicated in profile
	if !assert.Equal(t, "close", tCase.TestSteps[5].Request.Headers["Connection"]) {
		t.Fatal()
	}
}
