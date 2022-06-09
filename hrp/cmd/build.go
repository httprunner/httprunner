package cmd

import (
	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v4/hrp"
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
		return hrp.BuildPlugin(args[0], output)
	},
}

var output string

func init() {
	rootCmd.AddCommand(buildCmd)

	buildCmd.Flags().StringVarP(&output, "output", "o", "", "funplugin product output path, default: cwd")
}
