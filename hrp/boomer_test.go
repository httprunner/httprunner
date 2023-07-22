package hrp

import (
	"testing"
	"time"
)

func TestBoomerStandaloneRun(t *testing.T) {
	buildHashicorpGoPlugin()
	defer removeHashicorpGoPlugin()

	testcase1 := &TestCase{
		Config: NewConfig("TestCase1").SetBaseURL("https://postman-echo.com"),
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
	testcase2 := TestCasePath(demoTestCaseWithPluginJSONPath)

	b := NewStandaloneBoomer(2, 1)
	go b.Run(testcase1, &testcase2)
	time.Sleep(5 * time.Second)
	b.Quit()
}
