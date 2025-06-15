package cmd

import (
	"strings"
	"time"

	"github.com/httprunner/funplugin/myexec"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/sdk"
)

var CmdPytest = &cobra.Command{
	Use:                "pytest $path ...",
	Short:              "Run API test with pytest",
	Args:               cobra.MinimumNArgs(1),
	DisableFlagParsing: true, // allow to pass any args to pytest
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		startTime := time.Now()
		defer func() {
			sdk.SendGA4Event("hrp_pytest", map[string]interface{}{
				"args":                 strings.Join(args, "-"),
				"success":              err == nil,
				"engagement_time_msec": time.Since(startTime).Milliseconds(),
			})
		}()

		packages := []string{"httprunner"}
		_, err = myexec.EnsurePython3Venv(venv, packages...)
		if err != nil {
			log.Error().Err(err).Msg("python3 venv is not ready")
			return errors.Wrap(code.InvalidPython3Venv, err.Error())
		}
		return runPytest(args)
	},
}

func runPytest(args []string) error {
	args = append([]string{"run"}, args...)
	return myexec.ExecPython3Command("httprunner", args...)
}
