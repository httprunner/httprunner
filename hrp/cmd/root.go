package cmd

import (
	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v4/hrp"
	"github.com/httprunner/httprunner/v4/hrp/cmd/adb"
	"github.com/httprunner/httprunner/v4/hrp/cmd/ios"
	"github.com/httprunner/httprunner/v4/hrp/internal/code"
	"github.com/httprunner/httprunner/v4/hrp/internal/version"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "hrp",
	Short: "Next-Generation API Testing Solution.",
	Long: `
██╗  ██╗████████╗████████╗██████╗ ██████╗ ██╗   ██╗███╗   ██╗███╗   ██╗███████╗██████╗
██║  ██║╚══██╔══╝╚══██╔══╝██╔══██╗██╔══██╗██║   ██║████╗  ██║████╗  ██║██╔════╝██╔══██╗
███████║   ██║      ██║   ██████╔╝██████╔╝██║   ██║██╔██╗ ██║██╔██╗ ██║█████╗  ██████╔╝
██╔══██║   ██║      ██║   ██╔═══╝ ██╔══██╗██║   ██║██║╚██╗██║██║╚██╗██║██╔══╝  ██╔══██╗
██║  ██║   ██║      ██║   ██║     ██║  ██║╚██████╔╝██║ ╚████║██║ ╚████║███████╗██║  ██║
╚═╝  ╚═╝   ╚═╝      ╚═╝   ╚═╝     ╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═══╝╚═╝  ╚═══╝╚══════╝╚═╝  ╚═╝

HttpRunner is an open source API testing tool that supports HTTP(S)/HTTP2/WebSocket/RPC
network protocols, covering API testing, performance testing and digital experience
monitoring (DEM) test types. Enjoy! ✨ 🚀 ✨

License: Apache-2.0
Website: https://httprunner.com
Github: https://github.com/httprunner/httprunner
Copyright 2017 debugtalk`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		hrp.InitLogger(logLevel, logJSON)
	},
	Version:          version.VERSION,
	TraverseChildren: true, // parses flags on all parents before executing child command
	SilenceUsage:     true, // silence usage when an error occurs
}

var (
	logLevel string
	logJSON  bool
	venv     string
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() int {
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "INFO", "set log level")
	rootCmd.PersistentFlags().BoolVar(&logJSON, "log-json", false, "set log to json format (default colorized console)")
	rootCmd.PersistentFlags().StringVar(&venv, "venv", "", "specify python3 venv path")

	ios.Init(rootCmd)
	adb.Init(rootCmd)

	err := rootCmd.Execute()
	return code.GetErrorCode(err)
}
