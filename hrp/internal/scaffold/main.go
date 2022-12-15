package scaffold

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/httprunner/funplugin/fungo"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp"
	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
	"github.com/httprunner/httprunner/v4/hrp/internal/env"
	"github.com/httprunner/httprunner/v4/hrp/internal/myexec"
	"github.com/httprunner/httprunner/v4/hrp/internal/sdk"
	"github.com/httprunner/httprunner/v4/hrp/internal/version"
)

type PluginType string

const (
	Empty  PluginType = "empty"
	Ignore PluginType = "ignore"
	Py     PluginType = "py"
	Go     PluginType = "go"
)

type ProjectInfo struct {
	ProjectName string    `json:"project_name,omitempty" yaml:"project_name,omitempty"`
	CreateTime  time.Time `json:"create_time,omitempty" yaml:"create_time,omitempty"`
	Version     string    `json:"hrp_version,omitempty" yaml:"hrp_version,omitempty"`
}

//go:embed templates/*
var templatesDir embed.FS

// CopyFile copies a file from templates dir to scaffold project
func CopyFile(templateFile, targetFile string) error {
	log.Info().Str("path", targetFile).Msg("create file")
	content, err := templatesDir.ReadFile(templateFile)
	if err != nil {
		return errors.Wrap(err, "template file not found")
	}

	err = os.WriteFile(targetFile, content, 0o644)
	if err != nil {
		log.Error().Err(err).Msg("create file failed")
		return err
	}
	return nil
}

func CreateScaffold(projectName string, pluginType PluginType, venv string, force bool) error {
	// report event
	sdk.SendEvent(sdk.EventTracking{
		Category: "Scaffold",
		Action:   "hrp startproject",
	})

	log.Info().
		Str("projectName", projectName).
		Str("pluginType", string(pluginType)).
		Bool("force", force).
		Msg("create new scaffold project")

	// check if projectName exists
	if _, err := os.Stat(projectName); err == nil {
		if !force {
			log.Warn().Str("projectName", projectName).
				Msg("project name already exists, please specify a new one.")
			return fmt.Errorf("project name already exists")
		}

		log.Warn().Str("projectName", projectName).
			Msg("project name already exists, remove first !!!")
		os.RemoveAll(projectName)
	}

	// create project folders
	if err := builtin.CreateFolder(projectName); err != nil {
		return err
	}
	if err := builtin.CreateFolder(filepath.Join(projectName, "har")); err != nil {
		return err
	}
	if err := builtin.CreateFile(filepath.Join(projectName, "har", ".keep"), ""); err != nil {
		return err
	}
	if err := builtin.CreateFolder(filepath.Join(projectName, "testcases")); err != nil {
		return err
	}
	if err := builtin.CreateFolder(filepath.Join(projectName, env.ResultsDirName)); err != nil {
		return err
	}
	if err := builtin.CreateFile(filepath.Join(projectName, env.ResultsDirName, ".keep"), ""); err != nil {
		return err
	}

	projectInfo := &ProjectInfo{
		ProjectName: filepath.Base(projectName),
		CreateTime:  time.Now(),
		Version:     version.VERSION,
	}

	// dump project information to file
	err := builtin.Dump2JSON(projectInfo, filepath.Join(projectName, "proj.json"))
	if err != nil {
		return err
	}

	// create .gitignore
	err = CopyFile("templates/gitignore", filepath.Join(projectName, ".gitignore"))
	if err != nil {
		return err
	}
	// create .env
	err = CopyFile("templates/env", filepath.Join(projectName, ".env"))
	if err != nil {
		return err
	}

	// create project testcases
	if pluginType == Empty {
		// create empty project
		err := CopyFile("templates/testcases/demo_empty_request.json",
			filepath.Join(projectName, "testcases", "requests.json"))
		if err != nil {
			return err
		}
		return nil
	} else if pluginType == Ignore {
		// create project without funplugin
		err := CopyFile("templates/testcases/demo_without_funplugin.json",
			filepath.Join(projectName, "testcases", "requests.json"))
		if err != nil {
			return err
		}
		log.Info().Msg("skip creating function plugin")
		return nil
	}

	// create project with funplugin
	err = CopyFile("templates/testcases/demo_with_funplugin.json",
		filepath.Join(projectName, "testcases", "demo.json"))
	if err != nil {
		return err
	}
	err = CopyFile("templates/testcases/demo_requests.json",
		filepath.Join(projectName, "testcases", "requests.json"))
	if err != nil {
		return err
	}
	err = CopyFile("templates/testcases/demo_requests.yml",
		filepath.Join(projectName, "testcases", "requests.yml"))
	if err != nil {
		return err
	}
	err = CopyFile("templates/testcases/demo_ref_testcase.yml",
		filepath.Join(projectName, "testcases", "ref_testcase.yml"))
	if err != nil {
		return err
	}

	// create debugtalk function plugin
	switch pluginType {
	case Py:
		return createPythonPlugin(projectName, venv)
	case Go:
		return createGoPlugin(projectName)
	}

	return nil
}

func createGoPlugin(projectName string) error {
	log.Info().Msg("start to create hashicorp go plugin")
	// check go sdk
	if err := myexec.RunCommand("go", "version"); err != nil {
		return errors.Wrap(err, "go sdk not installed")
	}

	// create debugtalk.go
	pluginDir := filepath.Join(projectName, "plugin")
	if err := builtin.CreateFolder(pluginDir); err != nil {
		return err
	}
	err := CopyFile("templates/plugin/debugtalk.go",
		filepath.Join(projectName, "plugin", hrp.PluginGoSourceFile))
	if err != nil {
		return errors.Wrap(err, "copy debugtalk.go failed")
	}

	return nil
}

func createPythonPlugin(projectName, venv string) error {
	log.Info().Msg("start to create hashicorp python plugin")

	// create debugtalk.py
	pluginFile := filepath.Join(projectName, hrp.PluginPySourceFile)
	err := CopyFile("templates/plugin/debugtalk.py", pluginFile)
	if err != nil {
		return errors.Wrap(err, "copy file failed")
	}

	packages := []string{
		fmt.Sprintf("funppy==%s", fungo.Version),
		fmt.Sprintf("httprunner==%s", version.HttpRunnerMinimumVersion),
	}
	_, err = myexec.EnsurePython3Venv(venv, packages...)
	if err != nil {
		return err
	}

	return nil
}
