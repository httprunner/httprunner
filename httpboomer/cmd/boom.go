package cmd

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/httprunner/httpboomer"
)

// boomCmd represents the boom command
var boomCmd = &cobra.Command{
	Use:   "boom",
	Short: "run load test with boomer",
	Long:  `run yaml/json testcase files for load test`,
	Example: `  $ httpboomer boom demo.json	# run specified json testcase file
  $ httpboomer boom demo.yaml	# run specified yaml testcase file
  $ httpboomer boom examples/	# run testcases in specified folder`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var paths []httpboomer.ITestCase
		for _, arg := range args {
			paths = append(paths, &httpboomer.TestCasePath{Path: arg})
		}
		boomer := httpboomer.NewBoomer(masterHost, masterPort)
		boomer.EnableCPUProfile(cpuProfile, cpuProfileDuration)
		boomer.EnableMemoryProfile(memoryProfile, memoryProfileDuration)
		boomer.Run(paths...)
	},
}

var masterHost string
var masterPort int
var maxRPS int64               // TODO: init boomer with this flag
var requestIncreaseRate string // TODO: init boomer with this flag
var runTasks string            // TODO: init boomer with this flag
var memoryProfile string
var memoryProfileDuration time.Duration
var cpuProfile string
var cpuProfileDuration time.Duration

func init() {
	RootCmd.AddCommand(boomCmd)

	boomCmd.Flags().Int64Var(&maxRPS, "max-rps", 0, "Max RPS that boomer can generate, disabled by default.")
	boomCmd.Flags().StringVar(&requestIncreaseRate, "request-increase-rate", "-1", "Request increase rate, disabled by default.")
	boomCmd.Flags().StringVar(&runTasks, "run-tasks", "", "Run tasks without connecting to the master, multiply tasks is separated by comma. Usually, it's for debug purpose.")
	boomCmd.Flags().StringVar(&masterHost, "master-host", "127.0.0.1", "Host or IP address of locust master for distributed load testing.")
	boomCmd.Flags().IntVar(&masterPort, "master-port", 5557, "The port to connect to that is used by the locust master for distributed load testing.")
	boomCmd.Flags().StringVar(&memoryProfile, "mem-profile", "", "Enable memory profiling.")
	boomCmd.Flags().DurationVar(&memoryProfileDuration, "mem-profile-duration", 30*time.Second, "Memory profile duration.")
	boomCmd.Flags().StringVar(&cpuProfile, "cpu-profile", "", "Enable CPU profiling.")
	boomCmd.Flags().DurationVar(&cpuProfileDuration, "cpu-profile-duration", 30*time.Second, "CPU profile duration.")
}
