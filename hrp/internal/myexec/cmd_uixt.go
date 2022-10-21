//go:build darwin || linux

package myexec

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

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
		if err := RunCommand("python3", "--version"); err != nil {
			return "", errors.Wrap(err, "python3 not found")
		}

		// check if .venv exists
		if _, err := os.Stat(venv); err == nil {
			// .venv exists, remove first
			if err := RunCommand("rm", "-rf", venv); err != nil {
				return "", errors.Wrap(err, "remove existed venv failed")
			}
		}

		// create python3 .venv
		if err := RunCommand("python3", "-m", "venv", venv); err != nil {
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

func Command(name string, arg ...string) *exec.Cmd {
	cmd := exec.Command(name, arg...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	return cmd
}

func KillProcessesByGpid(cmd *exec.Cmd) error {
	return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
}
