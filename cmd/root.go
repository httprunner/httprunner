package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/httprunner/httpboomer"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "httpboomer",
	Short: "httpboomer = httprunner + boomer",
	Long: `HttpBoomer is a golang implementation of HttpRunner.
Ideally, HttpBoomer will be fully compatible with HttpRunner, including testcase format and usage.
What's more, HttpBoomer will integrate Boomer natively to be a better load generator for locust.`,
	Version: httpboomer.VERSION,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) {},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
