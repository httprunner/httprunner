package cmd

import (
	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v4/hrp/internal/build"
)

var buildCmd = &cobra.Command{
	Use:   "build $path ...",
	Short: "build plugin for testing",
	Long:  `build python/go plugin for testing`,
	Example: `  $ hrp build plugin/debugtalk.go
  $ hrp build plugin/debugtalk.py`,
	Args: cobra.MinimumNArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		setLogLevel(logLevel)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return build.Run(args)
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
