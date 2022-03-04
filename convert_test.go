package hrp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	demoTestCaseJSONPath = "examples/demo.json"
	demoTestCaseYAMLPath = "examples/demo.yaml"
)

func TestLoadCase(t *testing.T) {
	tcJSON, err := loadFromJSON(demoTestCaseJSONPath)
	if !assert.NoError(t, err) {
		t.Fail()
	}
	tcYAML, err := loadFromYAML(demoTestCaseYAMLPath)
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
