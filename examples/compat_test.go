package examples

import (
	"testing"

	"github.com/httprunner/hrp"
)

const demoHttpRunnerJSONPath = "../examples/demo_httprunner.json"
const demoHttpRunnerYAMLPath = "../examples/demo_httprunner.yaml"

func TestCompatTestCase(t *testing.T) {
	testcaseFromJSON := &hrp.TestCasePath{Path: demoHttpRunnerJSONPath}
	err := hrp.NewRunner(t).Run(testcaseFromJSON)
	if err != nil {
		t.Fatalf("run testcase error: %v", err)
	}

	testcaseFromYAML := &hrp.TestCasePath{Path: demoHttpRunnerYAMLPath}
	err = hrp.NewRunner(t).Run(testcaseFromYAML)
	if err != nil {
		t.Fatalf("run testcase error: %v", err)
	}
}
