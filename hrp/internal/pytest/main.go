package pytest

import (
	"fmt"

	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
	"github.com/httprunner/httprunner/v4/hrp/internal/sdk"
	"github.com/httprunner/httprunner/v4/hrp/internal/version"
)

func RunPytest(args []string) error {
	sdk.SendEvent(sdk.EventTracking{
		Category: "RunAPITests",
		Action:   "hrp pytest",
	})

	httprunner := fmt.Sprintf("httprunner>=%s", version.HttpRunnerMinVersion)
	python3, err := builtin.EnsurePython3Venv(httprunner)
	if err != nil {
		return err
	}

	args = append([]string{"-m", "httprunner", "run"}, args...)
	return builtin.ExecCommand(python3, args...)
}
