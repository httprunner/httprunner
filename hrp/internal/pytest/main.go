package pytest

import (
	"github.com/httprunner/httprunner/v4/hrp/internal/myexec"
	"github.com/httprunner/httprunner/v4/hrp/internal/sdk"
)

func RunPytest(args []string) error {
	sdk.SendEvent(sdk.EventTracking{
		Category: "RunAPITests",
		Action:   "hrp pytest",
	})

	args = append([]string{"run"}, args...)
	return myexec.ExecPython3Command("httprunner", args...)
}
