//go:build windows
// +build windows

package builtin

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func init() {
	// use python instead of python3 if python3 is not available
	if !isPython3(python3Executable) {
		python3Executable = "python"
	}
}

func isPython3(python string) bool {
	out, err := exec.Command("cmd", "/c", python, "--version").Output()
	if err != nil {
		return false
	}
	if strings.HasPrefix(string(out), "Python 3") {
		return true
	}
	return false
}

func getPython3Executable(venvDir string) string {
	python := filepath.Join(venvDir, "Scripts", "python3.exe")
	if isPython3(python) {
		return python
	}
	return filepath.Join(venvDir, "Scripts", "python.exe")
}

func ensurePython3Venv(venvDir string, packages ...string) (python3 string, err error) {
	python3 = getPython3Executable(venvDir)
	log.Info().
		Str("python3", python3).
		Interface("packages", packages).
		Msg("ensure python3 venv")

	systemPython := "python3"

	// check if python3 venv is available
	if !isPython3(python3) {
		// python3 venv not available, create one
		// check if system python3 is available
		log.Warn().Str("pythonPath", python3).Msg("python3 venv is not available, try to check system python3")
		if !isPython3(systemPython) {
			if !isPython3("python") {
				return "", errors.Wrap(err, "python3 not found")
			}
			systemPython = "python"
		}

		// check if .venv exists
		if _, err := os.Stat(venvDir); err == nil {
			// .venv exists, remove first
			if err := ExecCommand("del", "/q", venvDir); err != nil {
				return "", errors.Wrap(err, "remove existed venv failed")
			}
		}

		// create python3 .venv
		// notice: --symlinks should be specified for windows
		// https://github.com/actions/virtual-environments/issues/2690
		if err := ExecCommand(systemPython, "-m", "venv", "--symlinks", venvDir); err != nil {
			// fix: failed to symlink on Windows
			log.Warn().Msg("failed to create python3 .venv by using --symlinks, try to use --copies")
			if err := ExecCommand(systemPython, "-m", "venv", "--copies", venvDir); err != nil {
				return "", errors.Wrap(err, "create python3 venv failed")
			}
		}

		// fix: python3 doesn't exist in .venv on Windows
		if _, err := os.Stat(python3); err != nil {
			log.Warn().Msg("python3 doesn't exist, try to link python")
			err := os.Link(filepath.Join(venvDir, "Scripts", "python.exe"), python3)
			if err != nil {
				return "", errors.Wrap(err, "python3 doesn't exist in .venv")
			}
		}
	}

	// install default python packages
	for _, pkg := range packages {
		err := InstallPythonPackage(python3, pkg)
		if err != nil {
			return "", errors.Wrap(err, fmt.Sprintf("pip install %s failed", pkg))
		}
	}

	return python3, nil
}

func ExecCommand(cmdName string, args ...string) error {
	// "cmd /c" carries out the command specified by string and then stops
	// refer: https://docs.microsoft.com/en-us/windows-server/administration/windows-commands/cmd
	cmdStr := fmt.Sprintf("%s %s", cmdName, strings.Join(args, " "))
	cmd := exec.Command("cmd", "/c", cmdStr)
	log.Info().Str("cmd", cmd.String()).Msg("exec command")

	// print output with colors
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		log.Error().Err(err).Msg("exec command failed")
		return err
	}

	return nil
}
