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
