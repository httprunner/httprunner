package hrp

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLocatePlugin(t *testing.T) {
	cwd, _ := os.Getwd()
	_, err := locatePlugin(cwd)
	if !assert.Error(t, err) {
		t.Fail()
	}

	_, err = locatePlugin("")
	if !assert.Error(t, err) {
		t.Fail()
	}

	startPath := "examples/debugtalk.so"
	_, err = locatePlugin(startPath)
	if !assert.Nil(t, err) {
		t.Fail()
	}

	startPath = "examples/demo.json"
	_, err = locatePlugin(startPath)
	if !assert.Nil(t, err) {
		t.Fail()
	}

	startPath = "examples/"
	_, err = locatePlugin(startPath)
	if !assert.Nil(t, err) {
		t.Fail()
	}

	startPath = "examples/plugin/debugtalk.go"
	_, err = locatePlugin(startPath)
	if !assert.Nil(t, err) {
		t.Fail()
	}

	startPath = "/abc"
	_, err = locatePlugin(startPath)
	if !assert.Error(t, err) {
		t.Fail()
	}
}

func TestHttpRunner(t *testing.T) {
	testcase1 := &TestCase{
		Config: NewConfig("TestCase1").
			SetBaseURL("http://httpbin.org"),
		TestSteps: []IStep{
			NewStep("headers").
				GET("/headers").
				Validate().
				AssertEqual("status_code", 200, "check status code").
				AssertEqual("headers.\"Content-Type\"", "application/json", "check http response Content-Type"),
			NewStep("user-agent").
				GET("/user-agent").
				Validate().
				AssertEqual("status_code", 200, "check status code").
				AssertEqual("headers.\"Content-Type\"", "application/json", "check http response Content-Type"),
			NewStep("TestCase3").CallRefCase(&TestCase{Config: NewConfig("TestCase3")}),
		},
	}
	testcase2 := &TestCase{
		Config: NewConfig("TestCase2").SetWeight(3),
	}
	testcase3 := &TestCasePath{demoTestCaseJSONPath}

	err := NewRunner(t).Run(testcase1, testcase2, testcase3)
	if err != nil {
		t.Fatalf("run testcase error: %v", err)
	}
}
