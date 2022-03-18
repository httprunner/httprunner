package hrp

import (
	"math"
	"os"
	"os/exec"
	"testing"

	"github.com/rs/zerolog/log"
)

func buildHashicorpPlugin() {
	log.Info().Msg("[init] build hashicorp go plugin")
	cmd := exec.Command("go", "build",
		"-o", "examples/debugtalk.bin",
		"examples/plugin/hashicorp.go", "examples/plugin/debugtalk.go")
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func removeHashicorpPlugin() {
	log.Info().Msg("[teardown] remove hashicorp plugin")
	os.Remove("examples/debugtalk.bin")
}

func TestHttpRunnerWithGoPlugin(t *testing.T) {
	buildHashicorpPlugin()
	defer removeHashicorpPlugin()

	assertRunTestCases(t)
}

func TestHttpRunnerWithPythonPlugin(t *testing.T) {
	assertRunTestCases(t)
}

func assertRunTestCases(t *testing.T) {
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
			NewStep("TestCase3").CallRefCase(&TestCase{Config: NewConfig("TestCase3").SetBaseURL("http://httpbin.org"), TestSteps: []IStep{
				NewStep("ip").
					GET("/ip").
					Validate().
					AssertEqual("status_code", 200, "check status code").
					AssertEqual("headers.\"Content-Type\"", "application/json", "check http response Content-Type"),
			}}),
			NewStep("TestCase4").CallRefCase(&demoRefAPIYAMLPath),
			NewStep("TestCase5").CallRefCase(&demoTestCaseJSONPath),
		},
	}
	testcase2 := &TestCase{
		Config: NewConfig("TestCase2").SetWeight(3),
	}
	testcase3 := &TestCase{
		Config: NewConfig("TestCase1").
			SetBaseURL("https://postman-echo.com"),
		TestSteps: []IStep{
			NewStep("TestCase5").CallRefAPI(&demoAPIYAMLPath),
		},
	}
	testcase4 := &demoRefTestCaseJSONPath

	r := NewRunner(t)
	err := r.Run(testcase1, testcase2, testcase3, testcase4)
	if err != nil {
		t.Fatalf("run testcase error: %v", err)
	}
}

func TestInitRendezvous(t *testing.T) {
	rendezvousBonudaryTestcase := &TestCase{
		Config: NewConfig("run request with functions").
			SetBaseURL("https://postman-echo.com").
			WithVariables(map[string]interface{}{
				"n": 5,
				"a": 12.3,
				"b": 3.45,
			}),
		TestSteps: []IStep{
			NewStep("test negative number").
				Rendezvous("test negative number").
				WithUserNumber(-1),
			NewStep("test overflow number").
				Rendezvous("test overflow number").
				WithUserNumber(1000000),
			NewStep("test negative percent").
				Rendezvous("test very low percent").
				WithUserPercent(-0.5),
			NewStep("test very low percent").
				Rendezvous("test very low percent").
				WithUserPercent(0.00001),
			NewStep("test overflow percent").
				Rendezvous("test overflow percent").
				WithUserPercent(1.5),
			NewStep("test conflict params").
				Rendezvous("test conflict params").
				WithUserNumber(1).
				WithUserPercent(0.123),
			NewStep("test negative timeout").
				Rendezvous("test negative timeout").
				WithTimeout(-1000),
		},
	}

	type rendezvousParam struct {
		number  int64
		percent float32
		timeout int64
	}
	expectedRendezvousParams := []rendezvousParam{
		{number: 100, percent: 1, timeout: 5000},
		{number: 100, percent: 1, timeout: 5000},
		{number: 100, percent: 1, timeout: 5000},
		{number: 0, percent: 0.00001, timeout: 5000},
		{number: 100, percent: 1, timeout: 5000},
		{number: 100, percent: 1, timeout: 5000},
		{number: 100, percent: 1, timeout: 5000},
	}

	rendezvousList := initRendezvous(rendezvousBonudaryTestcase, 100)

	for i, r := range rendezvousList {
		if r.Number != expectedRendezvousParams[i].number {
			t.Fatalf("run rendezvous %v error: expected number: %v, real number: %v", r.Name, expectedRendezvousParams[i].number, r.Number)
		}
		if math.Abs(float64(r.Percent-expectedRendezvousParams[i].percent)) > 0.001 {
			t.Fatalf("run rendezvous %v error: expected percent: %v, real percent: %v", r.Name, expectedRendezvousParams[i].percent, r.Percent)
		}
		if r.Timeout != expectedRendezvousParams[i].timeout {
			t.Fatalf("run rendezvous %v error: expected timeout: %v, real timeout: %v", r.Name, expectedRendezvousParams[i].timeout, r.Timeout)
		}
	}
}

func TestGenHTMLReport(t *testing.T) {
	summary := newOutSummary()
	caseSummary1 := newSummary()
	caseSummary2 := newSummary()
	stepResult1 := &stepData{}
	stepResult2 := &stepData{
		Name:        "Test",
		StepType:    stepTypeRequest,
		Success:     false,
		ContentSize: 0,
		Attachment:  "err",
	}
	caseSummary1.Records = []*stepData{stepResult1, stepResult2, nil}
	summary.appendCaseSummary(caseSummary1)
	summary.appendCaseSummary(caseSummary2)
	err := summary.genHTMLReport()
	if err != nil {
		t.Error(err)
	}
}
