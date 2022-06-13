package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v4/hrp"
	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
)

var buildCmd = &cobra.Command{
	Use:   "build $path ...",
	Short: "build plugin for testing",
	Long:  `build python/go plugin for testing`,
	Example: `  $ hrp build plugin/debugtalk.go
  $ hrp build plugin/debugtalk.py`,
	Args: cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		setLogLevel(logLevel)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		err := builtin.PrepareVenv(venv)
		if err != nil {
			log.Error().Err(err).Msg("prepare python3 venv failed")
			return err
		}
		return hrp.BuildPlugin(args[0], output)
	},
}

var output string

func init() {
	rootCmd.AddCommand(buildCmd)

	buildCmd.Flags().StringVarP(&output, "output", "o", "", "funplugin product output path, default: cwd")
}
