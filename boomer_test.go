package hrp

import (
	"testing"
	"time"
)

func TestBoomerStandaloneRun(t *testing.T) {
	testcase1 := &TestCase{
		Config: TConfig{
			Name:    "TestCase1",
			BaseURL: "http://httpbin.org",
		},
		TestSteps: []IStep{
			Step("headers").
				GET("/headers").
				Validate().
				AssertEqual("status_code", 200, "check status code").
				AssertEqual("headers.\"Content-Type\"", "application/json", "check http response Content-Type"),
			Step("user-agent").
				GET("/user-agent").
				Validate().
				AssertEqual("status_code", 200, "check status code").
				AssertEqual("headers.\"Content-Type\"", "application/json", "check http response Content-Type"),
			Step("TestCase3").CallRefCase(&TestCase{Config: TConfig{Name: "TestCase3"}}),
		},
	}
	testcase2 := &TestCasePath{demoTestCaseJSONPath}

	b := NewStandaloneBoomer(2, 1)
	go b.Run(testcase1, testcase2)
	time.Sleep(5 * time.Second)
	b.Quit()
}
