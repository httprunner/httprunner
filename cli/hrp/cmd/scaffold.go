package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/httprunner/hrp/internal/scaffold"
)

var scaffoldCmd = &cobra.Command{
	Use:   "startproject $project_name",
	Short: "create a scaffold project",
	Args:  cobra.ExactValidArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		setLogLevel(logLevel)
	},
	Run: func(cmd *cobra.Command, args []string) {
		err := scaffold.CreateScaffold(args[0])
		if err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(scaffoldCmd)
}
