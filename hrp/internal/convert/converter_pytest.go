package convert

import (
	"fmt"

	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
	"github.com/httprunner/httprunner/v4/hrp/internal/version"
)

func convert2PyTestScripts(paths ...string) error {
	httprunner := fmt.Sprintf("httprunner>=%s", version.HttpRunnerMinVersion)
	python3, err := builtin.EnsurePython3Venv(httprunner)
	if err != nil {
		return err
	}

	args := append([]string{"-m", "httprunner", "make"}, paths...)
	return builtin.ExecCommand(python3, args...)
}
