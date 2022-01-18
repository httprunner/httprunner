package hrp

import (
	"testing"
	"time"
)

func TestBoomerStandaloneRun(t *testing.T) {
	buildHashicorpPlugin()
	defer removeHashicorpPlugin()

	testcase1 := &TestCase{
		Config: NewConfig("TestCase1").SetBaseURL("http://httpbin.org"),
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
	testcase2 := &TestCasePath{demoTestCaseJSONPath}

	b := NewBoomer(2, 1)
	go b.Run(testcase1, testcase2)
	time.Sleep(5 * time.Second)
	b.Quit()
}
