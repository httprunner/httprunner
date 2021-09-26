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
	testcase2 := &TestCase{
		Config: TConfig{
			Name:   "TestCase2",
			Weight: 3,
		},
	}

	err := Test(t, testcase1, testcase2)
	if err != nil {
		t.Fatalf("run testcase error: %v", err)
	}
}
