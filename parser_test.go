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
}
