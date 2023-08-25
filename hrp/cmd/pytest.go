package cmd

import (
	"strings"
	"time"

	"github.com/httprunner/funplugin/myexec"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v4/hrp/internal/code"
	"github.com/httprunner/httprunner/v4/hrp/internal/pytest"
	"github.com/httprunner/httprunner/v4/hrp/internal/sdk"
)

var pytestCmd = &cobra.Command{
	Use:                "pytest $path ...",
	Short:              "run API test with pytest",
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
		return pytest.RunPytest(args)
	},
}

func init() {
	rootCmd.AddCommand(pytestCmd)
}
