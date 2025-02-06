package pytest

import (
	"github.com/httprunner/funplugin/myexec"
)

func RunPytest(args []string) error {
	args = append([]string{"run"}, args...)
	return myexec.ExecPython3Command("httprunner", args...)
}
