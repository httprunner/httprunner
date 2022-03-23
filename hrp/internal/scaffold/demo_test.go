package scaffold

import (
	"os"
	"os/exec"
	"testing"

	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/hrp"
	"github.com/httprunner/httprunner/hrp/internal/builtin"
)

var (
	demoTestCaseJSONPath hrp.TestCasePath = "../../../examples/hrp/demo.json"
	demoTestCaseYAMLPath hrp.TestCasePath = "../../../examples/hrp/demo.yaml"
)

func buildHashicorpPlugin() {
	log.Info().Msg("[init] build hashicorp go plugin")
	cmd := exec.Command("go", "build",
		"-o", "../../../examples/hrp/debugtalk.bin",
		"../../../examples/hrp/plugin/hashicorp.go", "../../../examples/hrp/plugin/debugtalk.go")
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func removeHashicorpPlugin() {
	log.Info().Msg("[teardown] remove hashicorp plugin")
	os.Remove("../../../examples/hrp/debugtalk.bin")
}

func TestGenDemoTestCase(t *testing.T) {
	tCase, _ := demoTestCase.ToTCase()
	err := builtin.Dump2JSON(tCase, demoTestCaseJSONPath.ToString())
	if err != nil {
		t.Fail()
	}
	err = builtin.Dump2YAML(tCase, demoTestCaseYAMLPath.ToString())
	if err != nil {
		t.Fail()
	}
}

func TestExampleDemo(t *testing.T) {
	buildHashicorpPlugin()
	defer removeHashicorpPlugin()

	demoTestCase.Config.Path = "../../../examples/hrp/debugtalk.bin"
	err := hrp.NewRunner(nil).Run(demoTestCase) // hrp.Run(demoTestCase)
	if err != nil {
		t.Fail()
	}
}

func TestJsonDemo(t *testing.T) {
	buildHashicorpPlugin()
	defer removeHashicorpPlugin()

	err := hrp.NewRunner(nil).Run(&demoTestCaseJSONPath) // hrp.Run(testCase)
	if err != nil {
		t.Fail()
	}
}

func TestYamlDemo(t *testing.T) {
	buildHashicorpPlugin()
	defer removeHashicorpPlugin()

	err := hrp.NewRunner(nil).Run(&demoTestCaseYAMLPath) // hrp.Run(testCase)
	if err != nil {
		t.Fail()
	}
}
