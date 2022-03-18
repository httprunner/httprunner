package scaffold

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/funplugin/shared"
	"github.com/httprunner/hrp"
	"github.com/httprunner/hrp/internal/builtin"
	"github.com/httprunner/hrp/internal/ga"
)

type PluginType uint

const (
	Ignore PluginType = iota
	Py
	Go
)

func CreateScaffold(projectName string, pluginType PluginType) error {
	// report event
	ga.SendEvent(ga.EventTracking{
		Category: "Scaffold",
		Action:   "hrp startproject",
	})

	// check if projectName exists
	if _, err := os.Stat(projectName); err == nil {
		log.Warn().Str("projectName", projectName).
			Msg("project name already exists, please specify a new one.")
		return fmt.Errorf("project name already exists")
	}

	log.Info().Str("projectName", projectName).Msg("create new scaffold project")

	// create project folders
	if err := builtin.CreateFolder(projectName); err != nil {
		return err
	}
	if err := builtin.CreateFolder(path.Join(projectName, "har")); err != nil {
		return err
	}
	if err := builtin.CreateFolder(path.Join(projectName, "testcases")); err != nil {
		return err
	}
	if err := builtin.CreateFolder(path.Join(projectName, "reports")); err != nil {
		return err
	}

	// create demo testcases
	var tCase *hrp.TCase
	if pluginType == Ignore {
		tCase, _ = demoTestCaseWithoutPlugin.ToTCase()
	} else {
		tCase, _ = demoTestCase.ToTCase()
	}
	err := builtin.Dump2JSON(tCase, path.Join(projectName, "testcases", "demo.json"))
	if err != nil {
		log.Error().Err(err).Msg("create demo.json testcase failed")
		return err
	}
	err = builtin.Dump2YAML(tCase, path.Join(projectName, "testcases", "demo.yaml"))
	if err != nil {
		log.Error().Err(err).Msg("create demo.yml testcase failed")
		return err
	}

	// create .gitignore
	if err := builtin.CreateFile(path.Join(projectName, ".gitignore"), demoIgnoreContent); err != nil {
		return err
	}
	// create .env
	if err := builtin.CreateFile(path.Join(projectName, ".env"), demoEnvContent); err != nil {
		return err
	}

	// create debugtalk function plugin
	switch pluginType {
	case Ignore:
		log.Info().Msg("skip creating function plugin")
		return nil
	case Py:
		return createPythonPlugin(projectName)
	case Go:
		return createGoPlugin(projectName)
	}

	return nil
}

func createGoPlugin(projectName string) error {
	log.Info().Msg("start to create hashicorp go plugin")
	// check go sdk
	if err := builtin.ExecCommand(exec.Command("go", "version"), projectName); err != nil {
		return errors.Wrap(err, "go sdk not installed")
	}

	// create debugtalk.go
	pluginDir := path.Join(projectName, "plugin")
	if err := builtin.CreateFolder(pluginDir); err != nil {
		return err
	}
	pluginFile := path.Join(pluginDir, "debugtalk.go")
	if err := builtin.CreateFile(pluginFile, demoGoPlugin); err != nil {
		return err
	}

	// create go mod
	if err := builtin.ExecCommand(exec.Command("go", "mod", "init", "plugin"), pluginDir); err != nil {
		return err
	}

	// download plugin dependency
	if err := builtin.ExecCommand(exec.Command("go", "get", "github.com/httprunner/funplugin"), pluginDir); err != nil {
		return err
	}

	// build plugin debugtalk.bin
	if err := builtin.ExecCommand(exec.Command("go", "build", "-o", path.Join("..", "debugtalk.bin"), "debugtalk.go"), pluginDir); err != nil {
		return err
	}

	return nil
}

func createPythonPlugin(projectName string) error {
	log.Info().Msg("start to create hashicorp python plugin")

	// create debugtalk.py
	pluginFile := path.Join(projectName, "debugtalk.py")
	if err := builtin.CreateFile(pluginFile, demoPyPlugin); err != nil {
		return err
	}

	// create python venv
	if _, err := shared.PreparePython3Venv(pluginFile); err != nil {
		return err
	}

	return nil
}
