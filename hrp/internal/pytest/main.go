package pytest

import (
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/hrp/internal/builtin"
)

func RunPytest(args []string) error {
	python3, err := builtin.EnsurePython3Venv("httprunner")
	if err != nil {
		return errors.Wrap(err, "ensure python venv failed")
	}

	args = append([]string{"-m", "httprunner", "run"}, args...)
	cmd := exec.Command(python3, args...)
	log.Info().Str("cmd", cmd.String()).Msg("run pytest")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrap(err, "pytest running failed")
	}
	out := strings.TrimSpace(string(output))
	println(out)

	return nil
}
