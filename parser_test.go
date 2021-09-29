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
		"v":     4.5, // variable name with one character
		"_v":    6.9, // variable name starts with underscore
	}

	testData := []struct {
		expr   string
		expect interface{}
	}{
		// no variable
		{"var_1", "var_1"},
		{"$", "$"},
		{"var_1$", "var_1$"},
		{"var_1$123", "var_1$123"}, // variable should starts with a letter
		// single variable
		{"$var_1", "abc"},
		{"${var_1}", "abc"},
		{"$var_3", 123},
		{"$var_4", map[string]interface{}{"a": 1}},
		{"${var_4}", map[string]interface{}{"a": 1}},
		{"$var_5", true},
		{"$var_6", nil},
		{"$v", 4.5},
		{"var_1$_v", "var_16.9"},
		// single variable with prefix or suffix
		{"$var_1#XYZ", "abc#XYZ"},
		{"${var_1}#XYZ", "abc#XYZ"},
		{"ABC$var_1", "ABCabc"},
		{"ABC${var_1}", "ABCabc"},
		{"ABC$var_1/", "ABCabc/"},
		{"ABC${v}/", "ABC4.5/"},
		// multiple variables
		{"/$var_1/$var_2/var3", "/abc/def/var3"},
		{"/${var_1}/$var_2/var3", "/abc/def/var3"},
		{"ABC$var_1$var_3", "ABCabc123"},
		{"ABC$var_1${var_3}", "ABCabc123"},
		{"ABC$var_1/$var_3", "ABCabc/123"},
		{"ABC${var_1}/${var_3}", "ABCabc/123"},
		{"ABC$var_1/123$var_1/456", "ABCabc/123abc/456"},
		{"ABC$var_1/123${var_1}/456", "ABCabc/123abc/456"},
		{"ABC$var_1/$var_2/$var_1", "ABCabc/def/abc"},
		{"ABC$var_1/$var_2/${var_1}", "ABCabc/def/abc"},
		{"func1($var_1, $var_3)", "func1(abc, 123)"},
		{"func1($var_1, ${var_3})", "func1(abc, 123)"},
		// TODO: fix compatibility with python version
		{"abc$var_4", "abcmap[a:1]"},       // "abc{'a': 1}"
		{"abc$var_5", "abctrue"},           // "abcTrue"
		{"/api/$SECRET_KEY", "/api/<nil>"}, // raise error
	}

	for _, data := range testData {
		if !assert.Equal(t, data.expect, parseData(data.expr, variablesMapping)) {
			t.Fail()
		}
	}
}
func TestParseDataStringWithVariablesAbnormal(t *testing.T) {
	variablesMapping := map[string]interface{}{
		"var_1": "abc",
		"var_2": "def",
		"var_3": 123,
		"var_4": map[string]interface{}{"a": 1},
		"var_5": true,
		"var_6": nil,
		"v":     4.5, // variable name with one character
		"_v":    6.9, // variable name starts with underscore
	}

	testData := []struct {
		expr   string
		expect interface{}
	}{
		{"ABC$var_1{", "ABCabc{"},
		{"{ABC$var_1{}a}", "{ABCabc{}a}"},
		{"AB{C$var_1{}a}", "AB{Cabc{}a}"},
		{"ABC$var_1}", "ABCabc}"},
	}

	for _, data := range testData {
		if !assert.Equal(t, data.expect, parseData(data.expr, variablesMapping)) {
			t.Fail()
		}
	}
}
