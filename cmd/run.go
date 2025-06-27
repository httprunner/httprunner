package cmd

import (
	"github.com/spf13/cobra"

	hrp "github.com/httprunner/httprunner/v5"
)

// runCmd represents the run command
var CmdRun = &cobra.Command{
	Use:   "run $path...",
	Short: "Run API test with go engine",
	Long:  `Run yaml/json testcase files for API test`,
	Example: `  $ hrp run demo.json	# run specified json testcase file
  $ hrp run demo.yaml	# run specified yaml testcase file
  $ hrp run examples/	# run testcases in specified folder`,
	Args: cobra.MinimumNArgs(1),
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
	caseTimeout       float32
	runMCPConfigPath  string // MCP config path for run command
	autoPopupHandler  bool   // enable auto popup handler for all steps
)

func init() {
	CmdRun.Flags().BoolVarP(&continueOnFailure, "continue-on-failure", "c", false, "continue running next step when failure occurs")
	CmdRun.Flags().BoolVar(&requestsLogOff, "log-requests-off", false, "turn off request & response details logging")
	CmdRun.Flags().BoolVar(&httpStatOn, "http-stat", false, "turn on HTTP latency stat (DNSLookup, TCP Connection, etc.)")
	CmdRun.Flags().BoolVar(&pluginLogOn, "log-plugin", false, "turn on plugin logging")
	CmdRun.Flags().StringVarP(&proxyUrl, "proxy-url", "p", "", "set proxy url")
	CmdRun.Flags().BoolVarP(&saveTests, "save-tests", "s", false, "save tests summary")
	CmdRun.Flags().BoolVarP(&genHTMLReport, "gen-html-report", "g", false, "generate html report")
	CmdRun.Flags().Float32Var(&caseTimeout, "case-timeout", 3600, "set testcase timeout (seconds)")
	CmdRun.Flags().StringVar(&runMCPConfigPath, "mcp-config", "", "path to the MCP config file")
	CmdRun.Flags().BoolVar(&autoPopupHandler, "enable-auto-popup-handler", false, "enable auto popup handler for all UI steps")
}

func makeHRPRunner() *hrp.HRPRunner {
	runner := hrp.NewRunner(nil).
		SetFailfast(!continueOnFailure).
		SetSaveTests(saveTests).
		SetCaseTimeout(caseTimeout)
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
	if runMCPConfigPath != "" {
		runner.SetMCPConfigPath(runMCPConfigPath)
	}
	if autoPopupHandler {
		runner.EnableAutoPopupHandler(autoPopupHandler)
	}
	return runner
}
