package pytest

import (
	"github.com/pkg/errors"

	"github.com/httprunner/httprunner/hrp/internal/builtin"
	"github.com/httprunner/httprunner/hrp/internal/sdk"
)

func RunPytest(args []string) error {
	sdk.SendEvent(sdk.EventTracking{
		Category: "RunAPITests",
		Action:   "hrp pytest",
	})

	python3, err := builtin.EnsurePython3Venv("httprunner")
	if err != nil {
		return errors.Wrap(err, "ensure python venv failed")
	}

	args = append([]string{"-m", "httprunner", "run"}, args...)
	return builtin.ExecCommand(python3, args...)
}
