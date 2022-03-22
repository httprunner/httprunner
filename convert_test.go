package hrp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	demoTestCaseJSONPath    TestCasePath = "examples/demo.json"
	demoTestCaseYAMLPath    TestCasePath = "examples/demo.yaml"
	demoRefAPIYAMLPath      TestCasePath = "examples/ref_api_test.yaml"
	demoRefTestCaseJSONPath TestCasePath = "examples/ref_testcase_test.json"
	demoThinkTimeJsonPath   TestCasePath = "examples/think_time_test.json"
	demoAPIYAMLPath         APIPath      = "examples/api/put.yml"
)

func TestLoadCase(t *testing.T) {
	tcJSON := &TCase{}
	tcYAML := &TCase{}
	err := loadFromJSON(demoTestCaseJSONPath.ToString(), tcJSON)
	if !assert.NoError(t, err) {
		t.Fail()
	}
	err = loadFromYAML(demoTestCaseYAMLPath.ToString(), tcYAML)
	if !assert.NoError(t, err) {
		t.Fail()
	}

	if !assert.Equal(t, tcJSON.Config.Name, tcYAML.Config.Name) {
		t.Fail()
	}
	if !assert.Equal(t, tcJSON.Config.BaseURL, tcYAML.Config.BaseURL) {
		t.Fail()
	}
	if !assert.Equal(t, tcJSON.TestSteps[1].Name, tcYAML.TestSteps[1].Name) {
		t.Fail()
	}
	if !assert.Equal(t, tcJSON.TestSteps[1].Request, tcYAML.TestSteps[1].Request) {
		t.Fail()
	}
}

func Test_convertCheckExpr(t *testing.T) {
	exprs := []struct {
		before string
		after  string
	}{
		// normal check expression
		{"a.b.c", "a.b.c"},
		{"headers.\"Content-Type\"", "headers.\"Content-Type\""},
		// check expression using regex
		{"covering (.*) testing,", "covering (.*) testing,"},
		{" (.*) a-b-c", " (.*) a-b-c"},
		// abnormal check expression
		{"-", "\"-\""},
		{"b-c", "\"b-c\""},
		{"a.b-c.d", "a.\"b-c\".d"},
		{"a-b.c-d", "\"a-b\".\"c-d\""},
		{"\"a-b\".c-d", "\"a-b\".\"c-d\""},
		{"headers.Content-Type", "headers.\"Content-Type\""},
		{"body.I-am-a-Key.name", "body.\"I-am-a-Key\".name"},
	}
	for _, expr := range exprs {
		if !assert.Equal(t, convertCheckExpr(expr.before), expr.after) {
			t.Fail()
		}
	}
}
