package pytest

import (
	"github.com/httprunner/httprunner/v4/hrp/internal/myexec"
	"github.com/httprunner/httprunner/v4/hrp/internal/sdk"
)

func RunPytest(args []string) error {
	sdk.SendGA4Event("hrp_pytest", nil)

	args = append([]string{"run"}, args...)
	return myexec.ExecPython3Command("httprunner", args...)
}
