package cmd

import (
	"github.com/spf13/cobra"

	hrp "github.com/httprunner/httprunner/v5"
	"github.com/httprunner/httprunner/v5/cmd/adb"
	"github.com/httprunner/httprunner/v5/cmd/ios"
	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/version"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "hrp",
	Short: "All-in-One Testing Framework for API, UI and Performance",
	Long: `
██╗  ██╗████████╗████████╗██████╗ ██████╗ ██╗   ██╗███╗   ██╗███╗   ██╗███████╗██████╗
██║  ██║╚══██╔══╝╚══██╔══╝██╔══██╗██╔══██╗██║   ██║████╗  ██║████╗  ██║██╔════╝██╔══██╗
███████║   ██║      ██║   ██████╔╝██████╔╝██║   ██║██╔██╗ ██║██╔██╗ ██║█████╗  ██████╔╝
██╔══██║   ██║      ██║   ██╔═══╝ ██╔══██╗██║   ██║██║╚██╗██║██║╚██╗██║██╔══╝  ██╔══██╗
██║  ██║   ██║      ██║   ██║     ██║  ██║╚██████╔╝██║ ╚████║██║ ╚████║███████╗██║  ██║
╚═╝  ╚═╝   ╚═╝      ╚═╝   ╚═╝     ╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═══╝╚═╝  ╚═══╝╚══════╝╚═╝  ╚═╝

HttpRunner: Enjoy your All-in-One Testing Solution ✨ 🚀 ✨

💡 Simple Yet Powerful
   - Natural language driven test scenarios powered by LLM
   - User-friendly SDK API with IDE auto-completion
   - Intuitive GoTest/YAML/JSON/Text testcase format

📌 Comprehensive Testing Capabilities
   - UI Automation: Android/iOS/Harmony/Browser
   - API Testing: HTTP(S)/HTTP2/WebSocket/RPC
   - Load Testing: run API testcase concurrently with boomer

🧩 High Scalability
   - Plugin system for custom functions
   - Distributed testing support
   - Cross-platform: macOS/Linux/Windows

🛠 Easy Integration
   - CI/CD friendly with JSON logs and HTML reports
   - Rich ecosystem tools

Learn more:
Website: https://httprunner.com
GitHub: https://github.com/httprunner/httprunner

Copyright © 2017-present debugtalk. Apache-2.0 License.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		hrp.InitLogger(logLevel, logJSON)
	},
	Version:          version.GetVersionInfo(),
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
