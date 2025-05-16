package hrp

import (
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/httprunner/httprunner/v5/internal/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildURL(t *testing.T) {
	var preparedURL *url.URL

	preparedURL = buildURL("https://postman-echo.com", "/get", nil)
	assert.Equal(t, preparedURL.String(), "https://postman-echo.com/get/")
	preparedURL = buildURL("https://postman-echo.com", "get", nil)
	assert.Equal(t, preparedURL.String(), "https://postman-echo.com/get/")
	preparedURL = buildURL("https://postman-echo.com/", "/get", nil)
	assert.Equal(t, preparedURL.String(), "https://postman-echo.com/get/")

	preparedURL = buildURL("https://postman-echo.com/abc/", "/get?a=1&b=2", nil)
	assert.Equal(t, preparedURL.String(), "https://postman-echo.com/abc/get?a=1&b=2")
	preparedURL = buildURL("https://postman-echo.com/abc", "get?a=1&b=2", nil)
	assert.Equal(t, preparedURL.String(), "https://postman-echo.com/abc/get?a=1&b=2")

	// omit query string in base url
	preparedURL = buildURL("https://postman-echo.com/abc?x=6&y=9", "/get?a=1&b=2", nil)
	assert.Equal(t, preparedURL.String(), "https://postman-echo.com/abc/get?a=1&b=2")

	preparedURL = buildURL("", "https://postman-echo.com/get", nil)
	assert.Equal(t, preparedURL.String(), "https://postman-echo.com/get/")

	// notice: step request url > config base url
	preparedURL = buildURL("https://postman-echo.com", "https://httpbin.org/get", nil)
	assert.Equal(t, preparedURL.String(), "https://httpbin.org/get/")

	// websocket url
	preparedURL = buildURL("wss://ws.postman-echo.com/raw", "", nil)
	assert.Equal(t, preparedURL.String(), "wss://ws.postman-echo.com/raw/")

	preparedURL = buildURL("wss://ws.postman-echo.com", "/raw", nil)
	assert.Equal(t, preparedURL.String(), "wss://ws.postman-echo.com/raw/")

	preparedURL = buildURL("wss://ws.postman-echo.com/raw", "ws://echo.websocket.events", nil)
	assert.Equal(t, preparedURL.String(), "ws://echo.websocket.events/")

	queryParams := url.Values{}
	queryParams.Add("c", "3")
	queryParams.Add("d", "4")
	preparedURL = buildURL("https://postman-echo.com/", "/get/", queryParams)
	assert.Equal(t, preparedURL.String(), "https://postman-echo.com/get?c=3&d=4")
	preparedURL = buildURL("https://postman-echo.com/abc", "get?a=1&b=2", queryParams)
	assert.Equal(t, preparedURL.String(), "https://postman-echo.com/abc/get?a=1&b=2&c=3&d=4")
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
		assert.Len(t, varMatched, 3)
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
		assert.Len(t, varMatched, 0)
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
		assert.Len(t, varMatched, 3)
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
		assert.Len(t, varMatched, 0)
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
		{"abc$var_4", "abcmap[a:1]"}, // "abc{'a': 1}"
		{"abc$var_5", "abctrue"},     // "abcTrue"
	}

	parser := NewParser()
	for _, data := range testData {
		parsedData, err := parser.Parse(data.expr, variablesMapping)
		assert.Nil(t, err)
		assert.Equal(t, data.expect, parsedData)
	}
}

func TestParseDataStringWithUndefinedVariables(t *testing.T) {
	variablesMapping := map[string]interface{}{
		"var_1": "abc",
		"var_2": "def",
	}

	testData := []struct {
		expr   string
		expect interface{}
	}{
		{"/api/$SECRET_KEY", "/api/$SECRET_KEY"}, // raise error
	}

	parser := NewParser()
	for _, data := range testData {
		parsedData, err := parser.Parse(data.expr, variablesMapping)
		assert.Error(t, err)
		assert.Equal(t, data.expect, parsedData)
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

	parser := NewParser()
	for _, data := range testData {
		parsedData, err := parser.Parse(data.expr, variablesMapping)
		assert.Nil(t, err)
		assert.Equal(t, data.expect, parsedData)
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

	parser := NewParser()
	for _, data := range testData {
		parsedData, err := parser.Parse(data.expr, variablesMapping)
		assert.Nil(t, err)
		assert.Equal(t, data.expect, parsedData)
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

	parser := NewParser()
	for _, data := range testData {
		parsedHeaders, err := parser.ParseHeaders(data.rawHeaders, variablesMapping)
		assert.Nil(t, err)
		assert.Equal(t, data.expectHeaders, parsedHeaders)
	}
}

func TestMergeVariables(t *testing.T) {
	testData := []struct {
		stepVariables   map[string]interface{}
		configVariables map[string]interface{}
		expectVariables map[string]interface{}
	}{
		{
			map[string]interface{}{"base_url": "$base_url", "foo1": "bar1"},
			map[string]interface{}{"base_url": "https://httpbin.org", "foo1": "bar111"},
			map[string]interface{}{"base_url": "https://httpbin.org", "foo1": "bar1"},
		},
		{
			map[string]interface{}{"n": 3, "b": 34.5, "varFoo2": "${max($a, $b)}"},
			map[string]interface{}{"n": 5, "a": 12.3, "b": 3.45, "varFoo1": "7a6K3", "varFoo2": 12.3},
			map[string]interface{}{"n": 3, "a": 12.3, "b": 34.5, "varFoo1": "7a6K3", "varFoo2": "${max($a, $b)}"},
		},
	}

	for _, data := range testData {
		mergedVariables := mergeVariables(data.stepVariables, data.configVariables)
		assert.Equal(t, data.expectVariables, mergedVariables)
	}
}

func TestMergeMap(t *testing.T) {
	testData := []struct {
		m             map[string]string
		overriddenMap map[string]string
		expectMap     map[string]string
	}{
		{
			map[string]string{"Accept": "*/*", "Accept-Encoding": "gzip, deflate, br", "Connection": "close"},
			map[string]string{"Cache-Control": "no-cache", "Connection": "keep-alive"},
			map[string]string{"Accept": "*/*", "Accept-Encoding": "gzip, deflate, br", "Connection": "close", "Cache-Control": "no-cache"},
		},
		{
			map[string]string{"Host": "postman-echo.com", "Postman-Token": "ea19464c-ddd4-4724-abe9-5e2b254c2723"},
			map[string]string{"Host": "Postman-echo.com", "Connection": "keep-alive", "Postman-Token": "ea19464c-ddd4-4724-abe9-5e2b342c2723"},
			map[string]string{"Host": "postman-echo.com", "Postman-Token": "ea19464c-ddd4-4724-abe9-5e2b254c2723", "Connection": "keep-alive"},
		},
		{
			map[string]string{"Accept": "*/*", "Accept-Encoding": "gzip, deflate, br", "Connection": "close"},
			nil,
			map[string]string{"Accept": "*/*", "Accept-Encoding": "gzip, deflate, br", "Connection": "close"},
		},
		{
			nil,
			map[string]string{"Cache-Control": "no-cache", "Connection": "keep-alive"},
			map[string]string{"Cache-Control": "no-cache", "Connection": "keep-alive"},
		},
	}

	for _, data := range testData {
		mergedMap := mergeMap(data.m, data.overriddenMap)
		assert.Equal(t, data.expectMap, mergedMap)
	}
}

func TestMergeSlices(t *testing.T) {
	testData := []struct {
		slice           []string
		overriddenSlice []string
		expectSlice     []string
	}{
		{
			[]string{"${setup_hook_example1($name)}", "${setup_hook_example2($name)}"},
			[]string{"${setup_hook_example3($name)}", "${setup_hook_example4($name)}"},
			[]string{"${setup_hook_example1($name)}", "${setup_hook_example2($name)}", "${setup_hook_example3($name)}", "${setup_hook_example4($name)}"},
		},
		{
			[]string{"${setup_hook_example1($name)}", "${setup_hook_example2($name)}"},
			nil,
			[]string{"${setup_hook_example1($name)}", "${setup_hook_example2($name)}"},
		},
		{
			nil,
			[]string{"${setup_hook_example3($name)}", "${setup_hook_example4($name)}"},
			[]string{"${setup_hook_example3($name)}", "${setup_hook_example4($name)}"},
		},
	}

	for _, data := range testData {
		mergedSlice := mergeSlices(data.slice, data.overriddenSlice)
		assert.Equal(t, data.expectSlice, mergedSlice)
	}
}

func TestMergeValidators(t *testing.T) {
	testData := []struct {
		validators           []interface{}
		overriddenValidators []interface{}
		expectValidators     []interface{}
	}{
		{
			[]interface{}{Validator{Check: "status_code", Assert: "equals", Expect: 200, Message: "assert response status code"}},
			[]interface{}{Validator{Check: `headers."Content-Type"`, Assert: "equals", Expect: "application/json; charset=utf-8", Message: "assert response header Content-Typ"}},
			[]interface{}{
				Validator{Check: "status_code", Assert: "equals", Expect: 200, Message: "assert response status code"},
				Validator{Check: `headers."Content-Type"`, Assert: "equals", Expect: "application/json; charset=utf-8", Message: "assert response header Content-Typ"},
			},
		},
		{
			[]interface{}{Validator{Check: "status_code", Assert: "equals", Expect: 302, Message: "assert response status code"}},
			[]interface{}{Validator{Check: "status_code", Assert: "equals", Expect: 200, Message: "assert response status code"}},
			[]interface{}{Validator{Check: "status_code", Assert: "equals", Expect: 302, Message: "assert response status code"}},
		},
		{
			nil,
			[]interface{}{Validator{Check: "status_code", Assert: "equals", Expect: 200, Message: "assert response status code"}},
			[]interface{}{Validator{Check: "status_code", Assert: "equals", Expect: 200, Message: "assert response status code"}},
		},
		{
			[]interface{}{Validator{Check: "status_code", Assert: "equals", Expect: 302, Message: "assert response status code"}},
			nil,
			[]interface{}{Validator{Check: "status_code", Assert: "equals", Expect: 302, Message: "assert response status code"}},
		},
	}

	for _, data := range testData {
		mergedValidators := mergeValidators(data.validators, data.overriddenValidators)
		assert.Equal(t, data.expectValidators, mergedValidators)
	}
}

func TestCallBuiltinFunction(t *testing.T) {
	parser := NewParser()

	// call function without arguments
	_, err := parser.CallFunc("get_timestamp")
	assert.Nil(t, err)

	// call function with one argument
	timeStart := time.Now()
	_, err = parser.CallFunc("sleep", 1)
	assert.Nil(t, err)
	assert.Greater(t, time.Since(timeStart), time.Duration(1)*time.Second)

	// call function with one argument
	result, err := parser.CallFunc("gen_random_string", 10)
	assert.Nil(t, err)
	assert.Equal(t, 10, len(result.(string)))

	// call function with two argument
	result, err = parser.CallFunc("max", float64(10), 9.99)
	assert.Nil(t, err)
	assert.Equal(t, float64(10), result.(float64))
}

func TestCallMCPTool(t *testing.T) {
	// Create a new case runner for testing
	caseRunner, err := NewCaseRunner(TestCase{
		Config: &TConfig{
			MCPConfigPath: "pkg/mcphost/testdata/test.mcp.json",
		},
	}, nil)
	require.Nil(t, err)

	parser := caseRunner.GetParser()

	resp, err := parser.CallMCPTool("filesystem", "read_file",
		map[string]interface{}{"path": "internal/version/VERSION"})
	assert.Nil(t, err)
	t.Logf("resp: %v", resp)
	assert.Contains(t, resp, version.VERSION)
}

func TestLiteralEval(t *testing.T) {
	testData := []struct {
		expr   string
		expect interface{}
	}{
		{"123", 123},
		{"1.23", 1.23},
		{"-123", -123},
		{"-1.23", -1.23},
		{"abc", "abc"},
		{" a bc ", "a bc"},
		{" a $bc ", "a $bc"},
		{"$var", "$var"},
		{" $var ", "$var"},
		{" $var1 ", "$var1"},
		{"", ""},
	}

	for _, data := range testData {
		value, err := literalEval(data.expr)
		assert.Nil(t, err)
		assert.Equal(t, data.expect, value)
	}
}

func TestParseFunctionArguments(t *testing.T) {
	testData := []struct {
		expr   string
		expect interface{}
	}{
		{"", []interface{}{}},
		{"123", []interface{}{123}},
		{"1.23", []interface{}{1.23}},
		{"-123", []interface{}{-123}},
		{"-1.23", []interface{}{-1.23}},
		{"abc", []interface{}{"abc"}},
		{"$var", []interface{}{"$var"}},
		{"1,2", []interface{}{1, 2}},
		{"1,2.3", []interface{}{1, 2.3}},
		{"1, -2.3", []interface{}{1, -2.3}},
		{"1,,2", []interface{}{1, nil, 2}},
		{" $var1 , 2 ", []interface{}{"$var1", 2}},
	}

	for _, data := range testData {
		value, err := parseFunctionArguments(data.expr)
		assert.Nil(t, err)
		assert.Equal(t, data.expect, value)
	}
}

func TestParseDataStringWithFunctions(t *testing.T) {
	variablesMapping := map[string]interface{}{
		"n": 5,
		"a": 12.3,
		"b": 3.45,
	}

	testData1 := []struct {
		expr   string
		expect interface{}
	}{
		{"${gen_random_string(5)}", 5},
		{"${gen_random_string($n)}", 5},
		{"123${gen_random_string(5)}abc", 11},
		{"123${gen_random_string($n)}abc", 11},
	}

	parser := NewParser()
	for _, data := range testData1 {
		value, err := parser.Parse(data.expr, variablesMapping)
		assert.Nil(t, err)
		assert.Equal(t, data.expect, len(value.(string)))
	}

	testData2 := []struct {
		expr   string
		expect interface{}
	}{
		{"${max($a, $b)}", 12.3},
		{"abc${max($a, $b)}123", "abc12.3123"},
		{"abc${max($a, 3.45)}123", "abc12.3123"},
	}

	for _, data := range testData2 {
		value, err := parser.Parse(data.expr, variablesMapping)
		assert.Nil(t, err)
		assert.Equal(t, data.expect, value)
	}
}

func TestConvertString(t *testing.T) {
	testData := []struct {
		raw    interface{}
		expect interface{}
	}{
		{"", ""},
		{"abc", "abc"},
		{"123", "123"},
		{123, "123"},
		{1.23, "1.23"},
		{100000000000, "100000000000"}, // avoid exponential notation
		{100000000000.23, "100000000000.23"},
		{nil, "<nil>"},
	}

	for _, data := range testData {
		value := convertString(data.raw)
		assert.Equal(t, data.expect, value)
	}
}

func TestParseVariables(t *testing.T) {
	testData := []struct {
		rawVars    map[string]interface{}
		expectVars map[string]interface{}
	}{
		{
			map[string]interface{}{"varA": "$varB", "varB": "$varC", "varC": "123", "a": 1, "b": 2},
			map[string]interface{}{"varA": "123", "varB": "123", "varC": "123", "a": int64(1), "b": int64(2)},
		},
		{
			map[string]interface{}{"n": 34.5, "a": 12.3, "b": "$n", "varFoo2": "${max($a, $b)}"},
			map[string]interface{}{"n": 34.5, "a": 12.3, "b": 34.5, "varFoo2": 34.5},
		},
	}

	parser := NewParser()
	for _, data := range testData {
		value, err := parser.ParseVariables(data.rawVars)
		assert.Nil(t, err)
		assert.Equal(t, data.expectVars, value)
	}
}

func TestParseVariablesAbnormal(t *testing.T) {
	testData := []struct {
		rawVars    map[string]interface{}
		expectVars map[string]interface{}
	}{
		{ // self referenced variable $varA
			map[string]interface{}{"varA": "$varA"},
			map[string]interface{}{"varA": "$varA"},
		},
		{ // undefined variable $varC
			map[string]interface{}{"varA": "$varB", "varB": "$varC", "a": 1, "b": 2},
			map[string]interface{}{"varA": "$varB", "varB": "$varC", "a": 1, "b": 2},
		},
		{ // circular reference
			map[string]interface{}{"varA": "$varB", "varB": "$varA"},
			map[string]interface{}{"varA": "$varB", "varB": "$varA"},
		},
	}

	parser := NewParser()
	for _, data := range testData {
		value, err := parser.ParseVariables(data.rawVars)
		assert.Error(t, err)
		assert.Equal(t, data.expectVars, value)
	}
}

func TestExtractVariables(t *testing.T) {
	testData := []struct {
		raw        interface{}
		expectVars []string
	}{
		{nil, nil},
		{"/$var1/$var1", []string{"var1"}},
		{
			map[string]interface{}{"varA": "$varB", "varB": "$varC", "varC": "123"},
			[]string{"varB", "varC"},
		},
		{
			[]interface{}{"varA", "$varB", 123, "$varC", "123"},
			[]string{"varB", "varC"},
		},
		{ // nested map and slice
			map[string]interface{}{"varA": "$varB", "varB": map[string]interface{}{"C": "$varC", "D": []string{"$varE"}}},
			[]string{"varB", "varC", "varE"},
		},
	}

	for _, data := range testData {
		var varList []string
		for varName := range extractVariables(data.raw) {
			varList = append(varList, varName)
		}
		sort.Strings(varList)
		assert.Equal(t, data.expectVars, varList)
	}
}

func TestFindallVariables(t *testing.T) {
	testData := []struct {
		raw        string
		expectVars []string
	}{
		{"", nil},
		{"$variable", []string{"variable"}},
		{"${variable}123", []string{"variable"}},
		{"/blog/$postid", []string{"postid"}},
		{"/$var1/$var2", []string{"var1", "var2"}},
		{"/$var1/$var1", []string{"var1"}},
		{"abc", nil},
		{"Z:2>1*0*1+1$a", []string{"a"}},
		{"Z:2>1*0*1+1$$a", nil},
		{"Z:2>1*0*1+1$$$a", []string{"a"}},
		{"Z:2>1*0*1+1$$$$a", nil},
		{"Z:2>1*0*1+1$$a$b", []string{"b"}},
		{"Z:2>1*0*1+1$$a$$b", nil},
		{"Z:2>1*0*1+1$a$b", []string{"a", "b"}},
		{"Z:2>1*0*1+1$$1", nil},
		{"a$var", []string{"var"}},
		{"a$v b", []string{"v"}},
		{"${func()}", nil},
		{"a${func(1,2)}b", nil},
		{"${gen_md5($TOKEN, $data, $random)}", []string{"TOKEN", "data", "random"}},
	}

	for _, data := range testData {
		var varList []string
		for varName := range findallVariables(data.raw) {
			varList = append(varList, varName)
		}
		sort.Strings(varList)
		assert.Equal(t, data.expectVars, varList)
	}
}

func TestSearchJmespath(t *testing.T) {
	testText := `{"a": {"b": "foo"}, "c": "bar", "d": {"e": [{"f": "foo"}, {"f": "bar"}]}}`
	testData := []struct {
		raw      string
		expected string
	}{
		{"body.a.b", "foo"},
		{"body.c", "bar"},
		{"body.d.e[0].f", "foo"},
		{"body.d.e[1].f", "bar"},
	}
	resp := http.Response{}
	resp.Body = io.NopCloser(strings.NewReader(testText))
	respObj, err := newHttpResponseObject(t, NewParser(), &resp)
	require.Nil(t, err)
	for _, data := range testData {
		assert.Equal(t, data.expected, respObj.searchJmespath(data.raw))
	}
}

func TestSearchRegexp(t *testing.T) {
	testText := `
	<ul class="nav navbar-nav navbar-right">
	<li><a href="/order/addToCart" style="color: white"><i class="fa fa-shopping-cart fa-2x"></i><span class="badge">0</span></a></li>
	<li class="dropdown">
	  <a class="dropdown-toggle" data-toggle="dropdown" href="#" style="color: white">
		Leo   <i class="fa fa-cog fa-2x"></i><span class="caret"></span></a>
	  <ul class="dropdown-menu">
		<li><a href="/user/changePassword">Change Password</a></li>
		<li><a href="/user/addAddress">Shipping</a></li>
		<li><a href="/user/addCard">Payment</a></li>
		<li><a href="/order/orderHistory">Order History</a></li>
		<li><a href="/user/signOut">Sign Out</a></li>
	  </ul>
	</li>

	<li>&nbsp;&nbsp;&nbsp;</li>
	<li><a href="/user/signOut" style="color: white"><i class="fa fa-sign-out fa-2x"></i>
	  Sign Out</a></li>
  </ul>
`
	testData := []struct {
		raw      string
		expected string
	}{
		{"/user/signOut\">(.*)</a></li>", "Sign Out"},
		{"<li><a href=\"/user/(.*)\" style", "signOut"},
		{"		(.*)   <i class=\"fa fa-cog fa-2x\"></i>", "Leo"},
	}
	// new response object
	resp := http.Response{}
	resp.Body = io.NopCloser(strings.NewReader(testText))
	respObj, err := newHttpResponseObject(t, NewParser(), &resp)
	require.Nil(t, err)
	for _, data := range testData {
		assert.Equal(t, data.expected, respObj.searchRegexp(data.raw))
	}
}

func TestConvertCheckExpr(t *testing.T) {
	exprs := []struct {
		before string
		after  string
	}{
		// normal check expression
		{"a.b.c", "a.b.c"},
		{"a.\"b-c\".d", "a.\"b-c\".d"},
		{"a.b-c.d", "a.b-c.d"},
		{"body.args.a[-1]", "body.args.a[-1]"},
		// check expression using regex
		{"covering (.*) testing,", "covering (.*) testing,"},
		{" (.*) a-b-c", " (.*) a-b-c"},
		// abnormal check expression
		{"headers.Content-Type", "headers.\"Content-Type\""},
		{"headers.\"Content-Type", "headers.\"Content-Type\""},
		{"headers.Content-Type\"", "headers.\"Content-Type\""},
		{"headers.User-Agent", "headers.\"User-Agent\""},
	}
	for _, expr := range exprs {
		assert.Equal(t, expr.after, convertJmespathExpr(expr.before))
	}
}

func TestFindAllPythonFunctionNames(t *testing.T) {
	content := `
def test_1():	# exported function
    pass

def _test_2():	# exported function
    pass

def __test_3():	# private function
    pass

# def test_4():	# commented out function
#    pass

def Test5():	# exported function
    pass
`
	names, err := regexPyFunctionName.findAllFunctionNames(content)
	assert.Nil(t, err)
	assert.Contains(t, names, "test_1")
	assert.Contains(t, names, "Test5")
	assert.Contains(t, names, "_test_2")
	assert.NotContains(t, names, "__test_3")
	// commented out function
	assert.NotContains(t, names, "test_4")
}

func TestFindAllGoFunctionNames(t *testing.T) {
	content := `
func Test1() {	// exported function
	return
}

func testFunc2() {	// exported function
	return
}

func init() {	// private function
	return
}

func _Test3() { // exported function
	return
}

// func Test4() {	// commented out function
// 	return
// }
`
	names, err := regexGoFunctionName.findAllFunctionNames(content)
	assert.Nil(t, err)
	assert.Contains(t, names, "Test1")
	assert.Contains(t, names, "testFunc2")
	assert.NotContains(t, names, "init")
	assert.Contains(t, names, "_Test3")
	// commented out function
	assert.NotContains(t, names, "Test4")
}

func TestFindAllGoFunctionNamesAbnormal(t *testing.T) {
	content := `
func init() {	// private function
	return
}

func main() {	// should not define main() function
	return
}
`
	_, err := regexGoFunctionName.findAllFunctionNames(content)
	assert.Error(t, err)
}
