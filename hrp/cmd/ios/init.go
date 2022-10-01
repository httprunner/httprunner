package ios

import "github.com/spf13/cobra"

var iosRootCmd = &cobra.Command{
	Use:   "ios",
	Short: "simple utils for ios device management",
}

func Init(rootCmd *cobra.Command) {
	rootCmd.AddCommand(iosRootCmd)
}
