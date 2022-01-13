package cmd

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/httprunner/hrp"
	"github.com/httprunner/hrp/internal/boomer"
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
		boomer.SetUlimit(10240) // ulimit -n 10240
		setLogLevel("WARN")     // disable info logs for load testing
	},
	Run: func(cmd *cobra.Command, args []string) {
		var paths []hrp.ITestCase
		for _, arg := range args {
			paths = append(paths, &hrp.TestCasePath{Path: arg})
		}
		hrpBoomer := hrp.NewBoomer(spawnCount, spawnRate)
		hrpBoomer.SetRateLimiter(maxRPS, requestIncreaseRate)
		if loopCount > 0 {
			hrpBoomer.SetLoopCount(loopCount)
		}
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
	maxRPS                   int64
	loopCount                int64
	requestIncreaseRate      string
	memoryProfile            string
	memoryProfileDuration    time.Duration
	cpuProfile               string
	cpuProfileDuration       time.Duration
	prometheusPushgatewayURL string
	disableConsoleOutput     bool
)

func init() {
	rootCmd.AddCommand(boomCmd)

	boomCmd.Flags().Int64Var(&maxRPS, "max-rps", 0, "Max RPS that boomer can generate, disabled by default.")
	boomCmd.Flags().StringVar(&requestIncreaseRate, "request-increase-rate", "-1", "Request increase rate, disabled by default.")
	boomCmd.Flags().IntVar(&spawnCount, "spawn-count", 1, "The number of users to spawn for load testing")
	boomCmd.Flags().Float64Var(&spawnRate, "spawn-rate", 1, "The rate for spawning users")
	boomCmd.Flags().Int64Var(&loopCount, "loop-count", -1, "The specify running cycles for load testing")
	boomCmd.Flags().StringVar(&memoryProfile, "mem-profile", "", "Enable memory profiling.")
	boomCmd.Flags().DurationVar(&memoryProfileDuration, "mem-profile-duration", 30*time.Second, "Memory profile duration.")
	boomCmd.Flags().StringVar(&cpuProfile, "cpu-profile", "", "Enable CPU profiling.")
	boomCmd.Flags().DurationVar(&cpuProfileDuration, "cpu-profile-duration", 30*time.Second, "CPU profile duration.")
	boomCmd.Flags().StringVar(&prometheusPushgatewayURL, "prometheus-gateway", "", "Prometheus Pushgateway url.")
	boomCmd.Flags().BoolVar(&disableConsoleOutput, "disable-console-output", false, "Disable console output.")
}
