package cmd

import (
	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v4/hrp"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run $path...",
	Short: "run API test with go engine",
	Long:  `run yaml/json testcase files for API test`,
	Example: `  $ hrp run demo.json	# run specified json testcase file
  $ hrp run demo.yaml	# run specified yaml testcase file
  $ hrp run examples/	# run testcases in specified folder`,
	Args: cobra.MinimumNArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		setLogLevel(logLevel)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var paths []hrp.ITestCase
		for _, arg := range args {
			path := hrp.TestCasePath(arg)
			paths = append(paths, &path)
		}
		runner := makeHRPRunner()
		return runner.Run(paths...)
	},
}

var (
	continueOnFailure bool
	requestsLogOff    bool
	httpStatOn        bool
	pluginLogOn       bool
	proxyUrl          string
	saveTests         bool
	genHTMLReport     bool
)

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().BoolVarP(&continueOnFailure, "continue-on-failure", "c", false, "continue running next step when failure occurs")
	runCmd.Flags().BoolVar(&requestsLogOff, "log-requests-off", false, "turn off request & response details logging")
	runCmd.Flags().BoolVar(&httpStatOn, "http-stat", false, "turn on HTTP latency stat (DNSLookup, TCP Connection, etc.)")
	runCmd.Flags().BoolVar(&pluginLogOn, "log-plugin", false, "turn on plugin logging")
	runCmd.Flags().StringVarP(&proxyUrl, "proxy-url", "p", "", "set proxy url")
	runCmd.Flags().BoolVarP(&saveTests, "save-tests", "s", false, "save tests summary")
	runCmd.Flags().BoolVarP(&genHTMLReport, "gen-html-report", "g", false, "generate html report")
}

func makeHRPRunner() *hrp.HRPRunner {
	runner := hrp.NewRunner(nil).
		SetFailfast(!continueOnFailure).
		SetSaveTests(saveTests)
	if genHTMLReport {
		runner.GenHTMLReport()
	}
	if !requestsLogOff {
		runner.SetRequestsLogOn()
	}
	if httpStatOn {
		runner.SetHTTPStatOn()
	}
	if pluginLogOn {
		runner.SetPluginLogOn()
	}
	if venv != "" {
		runner.SetPython3Venv(venv)
	}
	if proxyUrl != "" {
		runner.SetProxyUrl(proxyUrl)
	}
	return runner
}
