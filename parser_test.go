package httpboomer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildURL(t *testing.T) {
	var url string
	url = buildURL("https://postman-echo.com", "/get")
	if url != "https://postman-echo.com/get" {
		t.Fatalf("buildURL error, %s != 'https://postman-echo.com/get'", url)
	}

	url = buildURL("https://postman-echo.com/abc/", "/get?a=1&b=2")
	if url != "https://postman-echo.com/get?a=1&b=2" {
		t.Fatalf("buildURL error, %s != 'https://postman-echo.com/get'", url)
	}

	url = buildURL("", "https://postman-echo.com/get")
	if url != "https://postman-echo.com/get" {
		t.Fatalf("buildURL error, %s != 'https://postman-echo.com/get'", url)
	}

	// notice: step request url > config base url
	url = buildURL("https://postman-echo.com", "https://httpbin.org/get")
	if url != "https://httpbin.org/get" {
		t.Fatalf("buildURL error, %s != 'https://httpbin.org/get'", url)
	}
}

func TestParseDataStringWithVariables(t *testing.T) {
	variablesMapping := map[string]interface{}{
		"var_1": "abc",
		"var_2": "def",
		"var_3": 123,
		"var_4": map[string]interface{}{"a": 1},
		"var_5": true,
		"var_6": nil,
	}

	if !assert.Equal(t, "abc", parseData("$var_1", variablesMapping)) {
		t.Fail()
	}
	if !assert.Equal(t, "abc", parseData("${var_1}", variablesMapping)) {
		t.Fail()
	}
	if !assert.Equal(t, "var_1", parseData("var_1", variablesMapping)) {
		t.Fail()
	}
	if !assert.Equal(t, "abc#XYZ", parseData("$var_1#XYZ", variablesMapping)) {
		t.Fail()
	}
	if !assert.Equal(t, "abc#XYZ", parseData("${var_1}#XYZ", variablesMapping)) {
		t.Fail()
	}
	if !assert.Equal(t, "/abc/def/var3", parseData("/$var_1/$var_2/var3", variablesMapping)) {
		t.Fail()
	}
	if !assert.Equal(t, 123, parseData("$var_3", variablesMapping)) {
		t.Fail()
	}
	if !assert.Equal(t, map[string]interface{}{"a": 1}, parseData("$var_4", variablesMapping)) {
		t.Fail()
	}
	if !assert.Equal(t, true, parseData("$var_5", variablesMapping)) {
		t.Fail()
	}
	if !assert.Equal(t, nil, parseData("$var_6", variablesMapping)) {
		t.Fail()
	}
	// TODO: fix compatibility with python version, "abc{'a': 1}"
	if !assert.Equal(t, "abcmap[a:1]", parseData("abc$var_4", variablesMapping)) {
		t.Fail()
	}
	// TODO: fix compatibility with python version, "abcTrue"
	if !assert.Equal(t, "abctrue", parseData("abc$var_5", variablesMapping)) {
		t.Fail()
	}
	// TODO: fix compatibility with python version, raise exception
	if !assert.Equal(t, "/api/<nil>", parseData("/api/$SECRET_KEY", variablesMapping)) {
		t.Fail()
	}
}
