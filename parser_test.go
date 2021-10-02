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

func TestRegexCompileVariable(t *testing.T) {
	testData := []string{
		"$var1",
		"${var1}",
		"$v",
		"var_1$_v",
		"${var_1}#XYZ",
		"func1($var_1, $var_3)",
	}

	for _, expr := range testData {
		varMatched := regexCompileVariable.FindStringSubmatch(expr)
		if !assert.Len(t, varMatched, 3) {
			t.Fail()
		}
	}
}

func TestRegexCompileAbnormalVariable(t *testing.T) {
	testData := []string{
		"var1",
		"${var1",
		"$123",
		"var_1$",
		"func1($123, var_3)",
	}

	for _, expr := range testData {
		varMatched := regexCompileVariable.FindStringSubmatch(expr)
		if !assert.Len(t, varMatched, 0) {
			t.Fail()
		}
	}
}

func TestRegexCompileFunction(t *testing.T) {
	testData := []string{
		"${func1()}",
		"${func1($a)}",
		"${func1($a, $b)}",
		"${func1($a, 123)}",
		"${func1(123, $b)}",
		"abc${func1(123, $b)}123",
	}

	for _, expr := range testData {
		varMatched := regexCompileFunction.FindStringSubmatch(expr)
		if !assert.Len(t, varMatched, 3) {
			t.Fail()
		}
	}
}

func TestRegexCompileAbnormalFunction(t *testing.T) {
	testData := []string{
		"${func1()",
		"${func1(}",
		"${func1)}",
		"$func1()}",
		"${1func1()}", // function name can not start with number
		"${func1($a}",
		"abc$func1(123, $b)}123",
		// "${func1($a $b)}",
		// "${func1($a, $123)}",
		// "${func1(123 $b)}",
	}

	for _, expr := range testData {
		varMatched := regexCompileFunction.FindStringSubmatch(expr)
		if !assert.Len(t, varMatched, 0) {
			t.Fail()
		}
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
		{"$", "$"},
		{"var_1$", "var_1$"},
		{"var_1$123", "var_1$123"},        // variable should starts with a letter
		{"ABC$var_1{", "ABCabc{"},         // {
		{"ABC$var_1}", "ABCabc}"},         // }
		{"{ABC$var_1{}a}", "{ABCabc{}a}"}, // {xx}
		{"AB{C$var_1{}a}", "AB{Cabc{}a}"}, // {xx{}x}
		{"ABC$$var_1{", "ABC$var_1{"},     // $$
		{"ABC$$$var_1{", "ABC$abc{"},      // $$$
		{"ABC$$$$var_1{", "ABC$$var_1{"},  // $$$$
		{"ABC$var_1${", "ABCabc${"},       // ${
		{"ABC$var_1${a", "ABCabc${a"},     // ${
		{"ABC$var_1$}a", "ABCabc$}a"},     // $}
		{"ABC$var_1}{a", "ABCabc}{a"},     // }{
		{"ABC$var_1{}a", "ABCabc{}a"},     // {}
	}

	for _, data := range testData {
		if !assert.Equal(t, data.expect, parseData(data.expr, variablesMapping)) {
			t.Fail()
		}
	}
}

func TestParseDataMapWithVariables(t *testing.T) {
	variablesMapping := map[string]interface{}{
		"var1": "foo1",
		"val1": 200,
		"var2": 123, // key is int
	}

	testData := []struct {
		expr   map[string]interface{}
		expect interface{}
	}{
		{map[string]interface{}{"key": "$var1"}, map[string]interface{}{"key": "foo1"}},
		{map[string]interface{}{"foo1": "$val1", "foo2": "bar2"}, map[string]interface{}{"foo1": 200, "foo2": "bar2"}},
		// parse map key, key is string
		{map[string]interface{}{"$var1": "$val1"}, map[string]interface{}{"foo1": 200}},
		// parse map key, key is int
		{map[string]interface{}{"$var2": "$val1"}, map[string]interface{}{"123": 200}},
	}

	for _, data := range testData {
		if !assert.Equal(t, data.expect, parseData(data.expr, variablesMapping)) {
			t.Fail()
		}
	}
}

func TestParseHeaders(t *testing.T) {
	variablesMapping := map[string]interface{}{
		"var1": "foo1",
		"val1": 200,
		"var2": 123, // key is int
		"val2": nil, // value is nil
	}

	testData := []struct {
		rawHeaders    map[string]string
		expectHeaders map[string]string
	}{
		{map[string]string{"key": "$var1"}, map[string]string{"key": "foo1"}},
		{map[string]string{"foo1": "$val1", "foo2": "bar2"}, map[string]string{"foo1": "200", "foo2": "bar2"}},
		// parse map key, key is string
		{map[string]string{"$var1": "$val1"}, map[string]string{"foo1": "200"}},
		// parse map key, key is int
		{map[string]string{"$var2": "$val1"}, map[string]string{"123": "200"}},
		// parse map key & value, key is int, value is nil
		{map[string]string{"$var2": "$val2"}, map[string]string{"123": "<nil>"}},
	}

	for _, data := range testData {
		if !assert.Equal(t, data.expectHeaders, parseHeaders(data.rawHeaders, variablesMapping)) {
			t.Fail()
		}
	}
}

func TestMergeVariables(t *testing.T) {
	stepVariables := map[string]interface{}{
		"base_url": "$base_url",
		"foo1":     "bar1",
	}
	configVariables := map[string]interface{}{
		"base_url": "https://httpbin.org",
		"foo1":     "bar111",
	}
	mergedVariables := mergeVariables(stepVariables, configVariables)
	expectVariables := map[string]interface{}{
		"base_url": "https://httpbin.org",
		"foo1":     "bar1",
	}
	if !assert.Equal(t, expectVariables, mergedVariables) {
		t.Fail()
	}
}
