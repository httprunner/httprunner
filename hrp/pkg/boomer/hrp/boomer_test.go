package hrpboomer

import (
	"testing"
	"time"

	"github.com/httprunner/httprunner/v4/hrp"
)

func TestBoomerStandaloneRun(t *testing.T) {
	testcase1 := &hrp.TestCase{
		Config: hrp.NewConfig("TestCase1").SetBaseURL("https://postman-echo.com"),
		TestSteps: []hrp.IStep{
			hrp.NewStep("headers").
				GET("/headers").
				Validate().
				AssertEqual("status_code", 200, "check status code").
				AssertEqual("headers.\"Content-Type\"", "application/json", "check http response Content-Type"),
			hrp.NewStep("user-agent").
				GET("/user-agent").
				Validate().
				AssertEqual("status_code", 200, "check status code").
				AssertEqual("headers.\"Content-Type\"", "application/json", "check http response Content-Type"),
			hrp.NewStep("TestCase3").CallRefCase(&hrp.TestCase{Config: hrp.NewConfig("TestCase3")}),
		},
	}

	b := NewStandaloneBoomer(2, 1)
	go b.Run(testcase1)
	time.Sleep(5 * time.Second)
	b.Quit()
}
