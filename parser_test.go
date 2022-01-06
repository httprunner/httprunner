package hrp

import (
	"sort"
	"testing"
	"time"

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
		{"abc$var_4", "abcmap[a:1]"}, // "abc{'a': 1}"
		{"abc$var_5", "abctrue"},     // "abcTrue"
	}

	parser := newParser()
	for _, data := range testData {
		parsedData, err := parser.parseData(data.expr, variablesMapping)
		if !assert.NoError(t, err) {
			t.Fail()
		}
		if !assert.Equal(t, data.expect, parsedData) {
			t.Fail()
		}
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

	parser := newParser()
	for _, data := range testData {
		parsedData, err := parser.parseData(data.expr, variablesMapping)
		if !assert.Error(t, err) {
			t.Fail()
		}
		if !assert.Equal(t, data.expect, parsedData) {
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

	parser := newParser()
	for _, data := range testData {
		parsedData, err := parser.parseData(data.expr, variablesMapping)
		if !assert.NoError(t, err) {
			t.Fail()
		}
		if !assert.Equal(t, data.expect, parsedData) {
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

	parser := newParser()
	for _, data := range testData {
		parsedData, err := parser.parseData(data.expr, variablesMapping)
		if !assert.NoError(t, err) {
			t.Fail()
		}
		if !assert.Equal(t, data.expect, parsedData) {
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

	parser := newParser()
	for _, data := range testData {
		parsedHeaders, err := parser.parseHeaders(data.rawHeaders, variablesMapping)
		if !assert.NoError(t, err) {
			t.Fail()
		}
		if !assert.Equal(t, data.expectHeaders, parsedHeaders) {
			t.Fail()
		}
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
		if !assert.Equal(t, data.expectVariables, mergedVariables) {
			t.Fail()
		}
	}
}

func TestCallBuiltinFunction(t *testing.T) {
	parser := newParser()

	// call function without arguments
	_, err := parser.callFunc("get_timestamp")
	if !assert.NoError(t, err) {
		t.Fail()
	}

	// call function with one argument
	timeStart := time.Now()
	_, err = parser.callFunc("sleep", 1)
	if !assert.NoError(t, err) {
		t.Fail()
	}
	if !assert.Greater(t, time.Since(timeStart), time.Duration(1)*time.Second) {
		t.Fail()
	}

	// call function with one argument
	result, err := parser.callFunc("gen_random_string", 10)
	if !assert.NoError(t, err) {
		t.Fail()
	}
	if !assert.Equal(t, 10, len(result.(string))) {
		t.Fail()
	}

	// call function with two argument
	result, err = parser.callFunc("max", float64(10), 9.99)
	if !assert.NoError(t, err) {
		t.Fail()
	}
	if !assert.Equal(t, float64(10), result.(float64)) {
		t.Fail()
	}
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
		if !assert.NoError(t, err) {
			t.Fail()
		}
		if !assert.Equal(t, data.expect, value) {
			t.Fail()
		}
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
		if !assert.NoError(t, err) {
			t.Fail()
		}
		if !assert.Equal(t, data.expect, value) {
			t.Fail()
		}
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

	parser := newParser()
	for _, data := range testData1 {
		value, err := parser.parseData(data.expr, variablesMapping)
		if !assert.NoError(t, err) {
			t.Fail()
		}
		if !assert.Equal(t, data.expect, len(value.(string))) {
			t.Fail()
		}
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
		value, err := parser.parseData(data.expr, variablesMapping)
		if !assert.NoError(t, err) {
			t.Fail()
		}
		if !assert.Equal(t, data.expect, value) {
			t.Fail()
		}
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
		{nil, "<nil>"},
	}

	for _, data := range testData {
		value := convertString(data.raw)
		if !assert.Equal(t, data.expect, value) {
			t.Fail()
		}
	}
}

func TestParseVariables(t *testing.T) {
	testData := []struct {
		rawVars    map[string]interface{}
		expectVars map[string]interface{}
	}{
		{
			map[string]interface{}{"varA": "$varB", "varB": "$varC", "varC": "123", "a": 1, "b": 2},
			map[string]interface{}{"varA": "123", "varB": "123", "varC": "123", "a": 1, "b": 2},
		},
		{
			map[string]interface{}{"n": 34.5, "a": 12.3, "b": "$n", "varFoo2": "${max($a, $b)}"},
			map[string]interface{}{"n": 34.5, "a": 12.3, "b": 34.5, "varFoo2": 34.5},
		},
	}

	parser := newParser()
	for _, data := range testData {
		value, err := parser.parseVariables(data.rawVars)
		if !assert.NoError(t, err) {
			t.Fail()
		}
		if !assert.Equal(t, data.expectVars, value) {
			t.Fail()
		}
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

	parser := newParser()
	for _, data := range testData {
		value, err := parser.parseVariables(data.rawVars)
		if !assert.Error(t, err) {
			t.Fail()
		}
		if !assert.Equal(t, data.expectVars, value) {
			t.Fail()
		}
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
		if !assert.Equal(t, data.expectVars, varList) {
			t.Fail()
		}
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
		if !assert.Equal(t, data.expectVars, varList) {
			t.Fail()
		}
	}
}

func TestParseParameters(t *testing.T) {
	testData := []struct {
		rawVars      map[string]interface{}
		expectLength int
	}{
		{
			map[string]interface{}{
				"username-password": "${parameterize(examples/account.csv)}",
				"user_agent":        []interface{}{"IOS/10.1", "IOS/10.2"}},
			6,
		},
		{
			map[string]interface{}{
				"username-password": [][]interface{}{{"test1", "111111"}, {"test2", "222222"}, {"test3", "333333"}},
				"user_agent":        []interface{}{"IOS/10.1", "IOS/10.2"},
				"app_version":       []interface{}{0.3}},
			6,
		},
		{
			map[string]interface{}{
				"username-password": [][]interface{}{{"test1", "111111"}, {"test2", "222222"}, {"test3", "333333"}},
				"user_agent":        []interface{}{"IOS/10.1", "IOS/10.2"},
				"app_version":       []interface{}{0.3, 0.4, 0.5}},
			18,
		},
		{
			map[string]interface{}{}, 0,
		},
		{
			nil, 0,
		},
	}
	for _, data := range testData {
		params, _ := parseParameters(data.rawVars, map[string]interface{}{})
		value := genCartesianProduct(params)
		if !assert.Len(t, value, data.expectLength) {
			t.Fail()
		}
	}
}

func TestParseParametersError(t *testing.T) {
	testData := []struct {
		rawVars map[string]interface{}
	}{
		{
			map[string]interface{}{
				"username_password": "${parameterize(examples/account.csv)}",
				"user_agent":        []interface{}{"IOS/10.1", "IOS/10.2"}},
		},
		{
			map[string]interface{}{
				"username-password": "${parameterize(examples/account.csv)}",
				"user-agent":        []interface{}{"IOS/10.1", "IOS/10.2"}},
		},
		{
			map[string]interface{}{
				"username-password": "${param(examples/account.csv)}",
				"user_agent":        []interface{}{"IOS/10.1", "IOS/10.2"}},
		},
	}
	for _, data := range testData {
		_, err := parseParameters(data.rawVars, map[string]interface{}{})
		if !assert.Error(t, err) {
			t.Fail()
		}
	}
}

func TestParseSlice(t *testing.T) {
	testData := []struct {
		rawVar1 string
		rawVar2 interface{}
		expect  []map[string]interface{}
	}{
		{
			"username-password",
			[]map[string]interface{}{{"username": "test1", "password": 111111, "other": "111"}, {"username": "test2", "password": 222222, "other": "222"}},
			[]map[string]interface{}{
				{"username": "test1", "password": 111111},
				{"username": "test2", "password": 222222},
			},
		},
		{
			"username-password",
			[][]string{{"test1", "111111"}, {"test2", "222222"}},
			[]map[string]interface{}{
				{"username": "test1", "password": "111111"},
				{"username": "test2", "password": "222222"},
			},
		},
		{
			"app_version",
			[]float64{3.1, 3.0},
			[]map[string]interface{}{
				{"app_version": 3.1},
				{"app_version": 3.0},
			},
		},
	}
	for _, data := range testData {
		value, _ := parseSlice(data.rawVar1, data.rawVar2)
		if !assert.Equal(t, data.expect, value) {
			t.Fail()
		}
	}
}

func TestParseSliceError(t *testing.T) {
	testData := []struct {
		rawVar1 string
		rawVar2 interface{}
	}{
		{
			"app_version",
			123,
		},
		{
			"app_version",
			"123",
		},
	}
	for _, data := range testData {
		_, err := parseSlice(data.rawVar1, data.rawVar2)
		if !assert.Error(t, err) {
			t.Fail()
		}
	}
}
