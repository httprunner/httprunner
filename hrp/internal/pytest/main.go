package pytest

import (
	"github.com/httprunner/httprunner/v4/hrp/internal/myexec"
)

func RunPytest(args []string) error {
	args = append([]string{"run"}, args...)
	return myexec.ExecPython3Command("httprunner", args...)
}
