package cmd

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v4/hrp/internal/myexec"
	"github.com/httprunner/httprunner/v4/hrp/internal/pytest"
	"github.com/httprunner/httprunner/v4/hrp/internal/version"
)

var pytestCmd = &cobra.Command{
	Use:   "pytest $path ...",
	Short: "run API test with pytest",
	Args:  cobra.MinimumNArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		setLogLevel(logLevel)
	},
	DisableFlagParsing: true, // allow to pass any args to pytest
	RunE: func(cmd *cobra.Command, args []string) error {
		packages := []string{
			fmt.Sprintf("httprunner==%s", version.HttpRunnerMinimumVersion),
		}
		_, err := myexec.EnsurePython3Venv(venv, packages...)
		if err != nil {
			log.Error().Err(err).Msg("python3 venv is not ready")
			return err
		}
		return pytest.RunPytest(args)
	},
}

func init() {
	rootCmd.AddCommand(pytestCmd)
}
