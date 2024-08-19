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
}
