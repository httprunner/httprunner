package httpboomer

import (
	"testing"
)

func TestHttpRunner(t *testing.T) {
	testcase1 := &TestCase{
		Config: TConfig{
			Name:    "TestCase1",
			BaseURL: "http://httpbin.org",
		},
		TestSteps: []IStep{
			RunRequest("headers").
				GET("/headers").
				Validate().
				AssertEqual("status_code", 200, "check status code").
				AssertEqual("headers.Host", "httpbin.org", "check http response host"),
			RunRequest("user-agent").
				GET("/user-agent").
				Validate().
				AssertEqual("status_code", 200, "check status code").
				AssertEqual("body.\"user-agent\"", "python-requests", "check User-Agent"),
			RunTestCase("TestCase3").WithVariables(Variables{"var1": "value1"}),
		},
	}
	testcase2 := &TestCase{
		Config: TConfig{
			Name:   "TestCase2",
			Weight: 3,
		},
	}

	err := HttpRunner().Run(testcase1, testcase2)
	if err != nil {
		t.Fatalf("run testcase error: %v", err)
	}
}
