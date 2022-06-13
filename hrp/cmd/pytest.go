package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
	"github.com/httprunner/httprunner/v4/hrp/internal/pytest"
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
		err := builtin.PrepareVenv(venv)
		if err != nil {
			log.Error().Err(err).Msg("prepare python3 venv failed")
			return err
		}
		return pytest.RunPytest(args)
	},
}

func init() {
	rootCmd.AddCommand(pytestCmd)
}
