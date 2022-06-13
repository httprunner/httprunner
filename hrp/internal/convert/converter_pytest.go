package convert

import (
	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
)

func convert2PyTestScripts(paths ...string) error {
	args := append([]string{"make"}, paths...)
	return builtin.ExecPython3Command("httprunner", args...)
}
