package adb

import "github.com/spf13/cobra"

var androidRootCmd = &cobra.Command{
	Use:              "adb",
	Short:            "simple utils for android device management",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {},
}

func Init(rootCmd *cobra.Command) {
	rootCmd.AddCommand(androidRootCmd)
}
