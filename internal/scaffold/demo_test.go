package scaffold

import (
	"fmt"
	"testing"

	"github.com/httprunner/hrp"
)

var (
	demoTestCaseJSONPath = "../../examples/demo.json"
	demoTestCaseYAMLPath = "../../examples/demo.yaml"
)

func TestGenDemoTestCase(t *testing.T) {
	tCase, _ := demoTestCase.ToTCase()
	err := tCase.Dump2JSON(demoTestCaseJSONPath)
	if err != nil {
		t.Fail()
	}
	err = tCase.Dump2YAML(demoTestCaseYAMLPath)
	if err != nil {
		t.Fail()
	}
}

func Example_demo() {
	err := hrp.NewRunner(nil).Run(demoTestCase) // hrp.Run(demoTestCase)
	fmt.Println(err)
	// Output:
	// <nil>
}

func Example_jsonDemo() {
	testCase := &hrp.TestCasePath{Path: demoTestCaseJSONPath}
	err := hrp.NewRunner(nil).Run(testCase) // hrp.Run(testCase)
	fmt.Println(err)
	// Output:
	// <nil>
}

func Example_yamlDemo() {
	testCase := &hrp.TestCasePath{Path: demoTestCaseYAMLPath}
	err := hrp.NewRunner(nil).Run(testCase) // hrp.Run(testCase)
	fmt.Println(err)
	// Output:
	// <nil>
}
