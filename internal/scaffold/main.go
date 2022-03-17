package scaffold

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/httprunner/hrp/internal/builtin"
	"github.com/httprunner/hrp/internal/ga"
	"github.com/rs/zerolog/log"
)

func CreateScaffold(projectName string) error {
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
	pluginDir := path.Join(projectName, "plugin")
	if err := builtin.CreateFolder(pluginDir); err != nil {
		return err
	}
	if err := builtin.CreateFolder(path.Join(projectName, "reports")); err != nil {
		return err
	}

	// create demo testcases
	tCase, _ := demoTestCase.ToTCase()
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

	// create debugtalk.go
	pluginFile := path.Join(pluginDir, "debugtalk.go")
	if err := builtin.CreateFile(pluginFile, demoPlugin); err != nil {
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

	// create .gitignore
	if err := builtin.CreateFile(path.Join(projectName, ".gitignore"), demoIgnoreContent); err != nil {
		return err
	}
	// create .env
	if err := builtin.CreateFile(path.Join(projectName, ".env"), demoEnvContent); err != nil {
		return err
	}

	return nil
}
