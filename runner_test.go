package hrp

import (
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

func TestHttpRunner(t *testing.T) {
	buildHashicorpPlugin()
	defer removeHashicorpPlugin()

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
