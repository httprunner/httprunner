package myexec

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/code"
	"github.com/httprunner/httprunner/v4/hrp/internal/env"
)

var python3Executable string = "python3" // system default python3

func isPython3(python string) bool {
	out, err := Command(python, "--version").Output()
	if err != nil {
		return false
	}
	if strings.HasPrefix(string(out), "Python 3") {
		return true
	}
	return false
}

// EnsurePython3Venv ensures python3 venv with specified packages
// venv should be directory path of target venv
func EnsurePython3Venv(venv string, packages ...string) (python3 string, err error) {
	// priority: specified > $HOME/.hrp/venv
	if venv == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", errors.Wrap(err, "get user home dir failed")
		}
		venv = filepath.Join(home, ".hrp", "venv")
	}
	python3, err = ensurePython3Venv(venv, packages...)
	if err != nil {
		return "", errors.Wrap(code.InvalidPython3Venv, err.Error())
	}
	python3Executable = python3
	log.Info().Str("Python3Executable", python3Executable).Msg("set python3 executable path")
	return python3, nil
}

func ExecPython3Command(cmdName string, args ...string) error {
	args = append([]string{"-m", cmdName}, args...)
	return RunCommand(python3Executable, args...)
}

func AssertPythonPackage(python3 string, pkgName, pkgVersion string) error {
	out, err := exec.Command(
		python3, "-c", fmt.Sprintf("import %s; print(%s.__version__)", pkgName, pkgName),
	).Output()
	if err != nil {
		return fmt.Errorf("python package %s not found", pkgName)
	}

	// do not check version if pkgVersion is empty
	if pkgVersion == "" {
		log.Info().Str("name", pkgName).Msg("python package is ready")
		return nil
	}

	// check package version equality
	version := strings.TrimSpace(string(out))
	if strings.TrimLeft(version, "v") != strings.TrimLeft(pkgVersion, "v") {
		return fmt.Errorf("python package %s version %s not matched, please upgrade to %s",
			pkgName, version, pkgVersion)
	}

	log.Info().Str("name", pkgName).Str("version", pkgVersion).Msg("python package is ready")
	return nil
}

func InstallPythonPackage(python3 string, pkg string) (err error) {
	var pkgName, pkgVersion string
	if strings.Contains(pkg, "==") {
		// funppy==0.5.0
		pkgInfo := strings.Split(pkg, "==")
		pkgName = pkgInfo[0]
		pkgVersion = pkgInfo[1]
	} else {
		// funppy
		pkgName = pkg
	}

	// check if package installed and version matched
	err = AssertPythonPackage(python3, pkgName, pkgVersion)
	if err == nil {
		return nil
	}

	// check if pip available
	err = RunCommand(python3, "-m", "pip", "--version")
	if err != nil {
		log.Warn().Msg("pip is not available")
		return errors.Wrap(err, "pip is not available")
	}

	log.Info().Str("pkgName", pkgName).Str("pkgVersion", pkgVersion).Msg("installing python package")

	// install package
	pypiIndexURL := env.PYPI_INDEX_URL
	if pypiIndexURL == "" {
		pypiIndexURL = "https://pypi.org/simple" // default
	}
	err = RunCommand(python3, "-m", "pip", "install", "--upgrade", pkg,
		"--index-url", pypiIndexURL,
		"--quiet", "--disable-pip-version-check")
	if err != nil {
		return errors.Wrap(err, "pip install package failed")
	}

	return AssertPythonPackage(python3, pkgName, pkgVersion)
}

func RunCommand(cmdName string, args ...string) error {
	cmd := Command(cmdName, args...)
	log.Info().Str("cmd", cmd.String()).Msg("exec command")

	// add cmd dir path to $PATH
	if cmdDir := filepath.Dir(cmdName); cmdDir != "" {
		path := fmt.Sprintf("%s:%s", cmdDir, env.PATH)
		if err := os.Setenv("PATH", path); err != nil {
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

func ExecCommandInDir(cmd *exec.Cmd, dir string) error {
	log.Info().Str("cmd", cmd.String()).Str("dir", dir).Msg("exec command")
	cmd.Dir = dir

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
