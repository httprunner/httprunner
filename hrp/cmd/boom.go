package cmd

import (
	"time"

	"github.com/myzhan/boomer"
	"github.com/spf13/cobra"

	"github.com/httprunner/hrp"
)

// boomCmd represents the boom command
var boomCmd = &cobra.Command{
	Use:   "boom",
	Short: "run load test with boomer",
	Long:  `run yaml/json testcase files for load test`,
	Example: `  $ hrp boom demo.json	# run specified json testcase file
  $ hrp boom demo.yaml	# run specified yaml testcase file
  $ hrp boom examples/	# run testcases in specified folder`,
	Args: cobra.MinimumNArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		hrp.SetLogger("WARN", logJSON) // disable info logs for load testing
	},
	Run: func(cmd *cobra.Command, args []string) {
		var paths []hrp.ITestCase
		for _, arg := range args {
			paths = append(paths, &hrp.TestCasePath{Path: arg})
		}
		hrpBoomer := hrp.NewStandaloneBoomer(spawnCount, spawnRate)
		if !disableConsoleOutput {
			hrpBoomer.AddOutput(boomer.NewConsoleOutput())
		}
		if prometheusPushgatewayURL != "" {
			hrpBoomer.AddOutput(boomer.NewPrometheusPusherOutput(prometheusPushgatewayURL, "hrp"))
		}
		hrpBoomer.EnableCPUProfile(cpuProfile, cpuProfileDuration)
		hrpBoomer.EnableMemoryProfile(memoryProfile, memoryProfileDuration)
		hrpBoomer.Run(paths...)
	},
}

var (
	spawnCount               int
	spawnRate                float64
	maxRPS                   int64  // TODO: init boomer with this flag
	requestIncreaseRate      string // TODO: init boomer with this flag
	runTasks                 string // TODO: init boomer with this flag
	memoryProfile            string
	memoryProfileDuration    time.Duration
	cpuProfile               string
	cpuProfileDuration       time.Duration
	prometheusPushgatewayURL string
	disableConsoleOutput     bool
)

func init() {
	RootCmd.AddCommand(boomCmd)

	boomCmd.Flags().Int64Var(&maxRPS, "max-rps", 0, "Max RPS that boomer can generate, disabled by default.")
	boomCmd.Flags().StringVar(&requestIncreaseRate, "request-increase-rate", "-1", "Request increase rate, disabled by default.")
	boomCmd.Flags().StringVar(&runTasks, "run-tasks", "", "Run tasks without connecting to the master, multiply tasks is separated by comma. Usually, it's for debug purpose.")
	boomCmd.Flags().IntVar(&spawnCount, "spawn-count", 1, "The number of users to spawn for load testing")
	boomCmd.Flags().Float64Var(&spawnRate, "spawn-rate", 1, "The rate for spawning users")
	boomCmd.Flags().StringVar(&memoryProfile, "mem-profile", "", "Enable memory profiling.")
	boomCmd.Flags().DurationVar(&memoryProfileDuration, "mem-profile-duration", 30*time.Second, "Memory profile duration.")
	boomCmd.Flags().StringVar(&cpuProfile, "cpu-profile", "", "Enable CPU profiling.")
	boomCmd.Flags().DurationVar(&cpuProfileDuration, "cpu-profile-duration", 30*time.Second, "CPU profile duration.")
	boomCmd.Flags().StringVar(&prometheusPushgatewayURL, "prometheus-gateway", "", "Prometheus Pushgateway url.")
	boomCmd.Flags().BoolVar(&disableConsoleOutput, "disable-console-output", false, "Disable console output.")
}
