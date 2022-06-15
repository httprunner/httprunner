//go:build darwin || linux
// +build darwin linux

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

func isPython3(python string) bool {
	out, err := exec.Command(python, "--version").Output()
	if err != nil {
		return false
	}
	if strings.HasPrefix(string(out), "Python 3") {
		return true
	}
	return false
}

func getPython3Executable(venvDir string) string {
	return filepath.Join(venvDir, "bin", "python3")
}

func ensurePython3Venv(venv string, packages ...string) (python3 string, err error) {
	python3 = getPython3Executable(venv)

	log.Info().
		Str("python3", python3).
		Interface("packages", packages).
		Msg("ensure python3 venv")

	// check if python3 venv is available
	if !isPython3(python3) {
		// python3 venv not available, create one
		// check if system python3 is available
		if err := ExecCommand("python3", "--version"); err != nil {
			return "", errors.Wrap(err, "python3 not found")
		}

		// check if .venv exists
		if _, err := os.Stat(venv); err == nil {
			// .venv exists, remove first
			if err := ExecCommand("rm", "-rf", venv); err != nil {
				return "", errors.Wrap(err, "remove existed venv failed")
			}
		}

		// create python3 .venv
		if err := ExecCommand("python3", "-m", "venv", venv); err != nil {
			return "", errors.Wrap(err, "create python3 venv failed")
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
	cmd := exec.Command(cmdName, args...)
	log.Info().Str("cmd", cmd.String()).Msg("exec command")

	// add cmd dir path to $PATH
	if cmdDir := filepath.Dir(cmdName); cmdDir != "" {
		PATH := fmt.Sprintf("%s:%s", cmdDir, os.Getenv("PATH"))
		if err := os.Setenv("PATH", PATH); err != nil {
			log.Error().Err(err).Msg("set env $PATH failed")
			return err
		}
	}

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
