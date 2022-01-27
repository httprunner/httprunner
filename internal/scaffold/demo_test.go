package scaffold

import (
	"os"
	"os/exec"
	"testing"

	"github.com/rs/zerolog/log"

	"github.com/httprunner/hrp"
	"github.com/httprunner/hrp/internal/builtin"
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
	err := builtin.Dump2JSON(tCase, demoTestCaseJSONPath)
	if err != nil {
		t.Fail()
	}
	err = builtin.Dump2YAML(tCase, demoTestCaseYAMLPath)
	if err != nil {
		t.Fail()
	}
}

func TestExampleDemo(t *testing.T) {
	buildHashicorpPlugin()
	defer removeHashicorpPlugin()

	demoTestCase.Config.Path = "../../examples/debugtalk.bin"
	err := hrp.NewRunner(nil).Run(demoTestCase) // hrp.Run(demoTestCase)
	if err != nil {
		t.Fail()
	}
}

func TestJsonDemo(t *testing.T) {
	buildHashicorpPlugin()
	defer removeHashicorpPlugin()

	testCase := &hrp.TestCasePath{Path: demoTestCaseJSONPath}
	err := hrp.NewRunner(nil).Run(testCase) // hrp.Run(testCase)
	if err != nil {
		t.Fail()
	}
}

func TestYamlDemo(t *testing.T) {
	buildHashicorpPlugin()
	defer removeHashicorpPlugin()

	testCase := &hrp.TestCasePath{Path: demoTestCaseYAMLPath}
	err := hrp.NewRunner(nil).Run(testCase) // hrp.Run(testCase)
	if err != nil {
		t.Fail()
	}
}
