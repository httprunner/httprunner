package hrp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadTestCases(t *testing.T) {
	// load test cases from folder path
	tc := TestCasePath("../examples/demo-with-py-plugin/testcases/")
	testCases, err := LoadTestCases(&tc)
	if !assert.Nil(t, err) {
		t.Fatal()
	}
	if !assert.Equal(t, 4, len(testCases)) {
		t.Fatal()
	}

	// load test cases from folder path, including sub folders
	tc = TestCasePath("../examples/demo-with-py-plugin/")
	testCases, err = LoadTestCases(&tc)
	if !assert.Nil(t, err) {
		t.Fatal()
	}
	if !assert.Equal(t, 4, len(testCases)) {
		t.Fatal()
	}

	// load test cases from single file path
	tc = TestCasePath(demoTestCaseWithPluginJSONPath)
	testCases, err = LoadTestCases(&tc)
	if !assert.Nil(t, err) {
		t.Fatal()
	}
	if !assert.Equal(t, 1, len(testCases)) {
		t.Fatal()
	}

	// load test cases from TestCase instance
	testcase := &TestCase{
		Config: NewConfig("TestCase").SetWeight(3),
	}
	testCases, err = LoadTestCases(testcase)
	if !assert.Nil(t, err) {
		t.Fatal()
	}
	if !assert.Equal(t, len(testCases), 1) {
		t.Fatal()
	}

	// load test cases from TestCaseJSON
	testcaseJSON := TestCaseJSON(`
	{
		"config":{"name":"TestCaseJSON"},
		"teststeps":[
			{"name": "step1", "request":{"url": "https://httpbin.org/get"}},
			{"name": "step2", "shell":{"string": "ls -l"}}
		]
	}`)
	testCases, err = LoadTestCases(&testcaseJSON)
	if !assert.Nil(t, err) {
		t.Fatal()
	}
	if !assert.Equal(t, len(testCases), 1) {
		t.Fatal()
	}
}

func TestLoadCase(t *testing.T) {
	tcJSON := &TestCaseDef{}
	tcYAML := &TestCaseDef{}
	err := LoadFileObject(demoTestCaseWithPluginJSONPath, tcJSON)
	if !assert.NoError(t, err) {
		t.Fatal()
	}
	err = LoadFileObject(demoTestCaseWithPluginYAMLPath, tcYAML)
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	if !assert.Equal(t, tcJSON.Config.Name, tcYAML.Config.Name) {
		t.Fatal()
	}
	if !assert.Equal(t, tcJSON.Config.BaseURL, tcYAML.Config.BaseURL) {
		t.Fatal()
	}
	if !assert.Equal(t, tcJSON.Steps[1].StepName, tcYAML.Steps[1].StepName) {
		t.Fatal()
	}
	if !assert.Equal(t, tcJSON.Steps[1].Request, tcJSON.Steps[1].Request) {
		t.Fatal()
	}
}
