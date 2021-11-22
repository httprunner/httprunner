package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/httprunner/hrp"
	"github.com/httprunner/hrp/internal/version"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "hrp",
	Short: "One-stop solution for HTTP(S) testing.",
	Long: `hrp (HttpRunner+) is the next generation for HttpRunner. Enjoy! ✨ 🚀 ✨

License: Apache-2.0
Github: https://github.com/httprunner/hrp
Copyright 2021 debugtalk`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if !logJSON {
			hrp.SetLogPretty()
		}
		hrp.SetLogLevel(logLevel)
	},
	Version: version.VERSION,
}

var (
	logLevel string
	logJSON  bool
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	RootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "INFO", "set log level")
	RootCmd.PersistentFlags().BoolVar(&logJSON, "log-json", false, "set log to json format")

	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
