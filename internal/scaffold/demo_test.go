package scaffold

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/rs/zerolog/log"

	"github.com/httprunner/hrp"
)

var (
	demoTestCaseJSONPath = "../../examples/demo.json"
	demoTestCaseYAMLPath = "../../examples/demo.yaml"
)

func buildHashicorpPlugin() {
	log.Info().Msg("[init] build hashicorp go plugin")
	cmd := exec.Command("go", "build",
		"-o", "../../examples/debugtalk.bin",
		"../../examples/plugin/hashicorp.go", "../../examples/plugin/debugtalk.go")
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func removeHashicorpPlugin() {
	log.Info().Msg("[teardown] remove hashicorp plugin")
	os.Remove("../../examples/debugtalk.bin")
}

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
	buildHashicorpPlugin()
	defer removeHashicorpPlugin()

	demoTestCase.Config.ToStruct().Path = "../../examples/debugtalk.bin"
	err := hrp.NewRunner(nil).Run(demoTestCase) // hrp.Run(demoTestCase)
	fmt.Println(err)
	// Output:
	// <nil>
}

func Example_jsonDemo() {
	buildHashicorpPlugin()
	defer removeHashicorpPlugin()

	testCase := &hrp.TestCasePath{Path: demoTestCaseJSONPath}
	err := hrp.NewRunner(nil).Run(testCase) // hrp.Run(testCase)
	fmt.Println(err)
	// Output:
	// <nil>
}

func Example_yamlDemo() {
	buildHashicorpPlugin()
	defer removeHashicorpPlugin()

	testCase := &hrp.TestCasePath{Path: demoTestCaseYAMLPath}
	err := hrp.NewRunner(nil).Run(testCase) // hrp.Run(testCase)
	fmt.Println(err)
	// Output:
	// <nil>
}
