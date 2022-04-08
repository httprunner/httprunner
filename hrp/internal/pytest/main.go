package pytest

import (
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func RunPytest(args []string) error {
	cmd := exec.Command("pytest", args...)
	log.Info().Str("cmd", cmd.String()).Msg("run pytest")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrap(err, "pytest running failed")
	}
	out := strings.TrimSpace(string(output))
	println(out)

	return nil
}
