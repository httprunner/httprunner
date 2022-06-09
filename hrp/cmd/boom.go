package cmd

import (
	"os"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v4/hrp"
	"github.com/httprunner/httprunner/v4/hrp/internal/boomer"
	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
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
			path := hrp.TestCasePath(arg)
			paths = append(paths, &path)
		}
		// if set profile, the priority is higher than the other commands
		if boomArgs.profile != "" {
			err := builtin.LoadFile(boomArgs.profile, &boomArgs)
			if err != nil {
				log.Error().Err(err).Msg("failed to load profile")
				os.Exit(1)
			}
		}

		hrpBoomer := hrp.NewBoomer(boomArgs.SpawnCount, boomArgs.SpawnRate)
		hrpBoomer.SetRateLimiter(boomArgs.MaxRPS, boomArgs.RequestIncreaseRate)
		if boomArgs.LoopCount > 0 {
			hrpBoomer.SetLoopCount(boomArgs.LoopCount)
		}
		if !boomArgs.DisableConsoleOutput {
			hrpBoomer.AddOutput(boomer.NewConsoleOutput())
		}
		if boomArgs.PrometheusPushgatewayURL != "" {
			hrpBoomer.AddOutput(boomer.NewPrometheusPusherOutput(boomArgs.PrometheusPushgatewayURL, "hrp", hrpBoomer.GetMode()))
		}
		hrpBoomer.SetDisableKeepAlive(boomArgs.DisableKeepalive)
		hrpBoomer.SetDisableCompression(boomArgs.DisableCompression)
		hrpBoomer.SetClientTransport()
		hrpBoomer.EnableCPUProfile(boomArgs.CPUProfile, boomArgs.CPUProfileDuration)
		hrpBoomer.EnableMemoryProfile(boomArgs.MemoryProfile, boomArgs.MemoryProfileDuration)
		hrpBoomer.EnableGracefulQuit()
		hrpBoomer.Run(paths...)
	},
}

type BoomArgs struct {
	SpawnCount               int           `json:"spawn-count,omitempty" yaml:"spawn-count,omitempty"`
	SpawnRate                float64       `json:"spawn-rate,omitempty" yaml:"spawn-rate,omitempty"`
	MaxRPS                   int64         `json:"max-rps,omitempty" yaml:"max-rps,omitempty"`
	LoopCount                int64         `json:"loop-count,omitempty" yaml:"loop-count,omitempty"`
	RequestIncreaseRate      string        `json:"request-increase-rate,omitempty" yaml:"request-increase-rate,omitempty"`
	MemoryProfile            string        `json:"memory-profile,omitempty" yaml:"memory-profile,omitempty"`
	MemoryProfileDuration    time.Duration `json:"memory-profile-duration" yaml:"memory-profile-duration"`
	CPUProfile               string        `json:"cpu-profile,omitempty" yaml:"cpu-profile,omitempty"`
	CPUProfileDuration       time.Duration `json:"cpu-profile-duration,omitempty" yaml:"cpu-profile-duration,omitempty"`
	PrometheusPushgatewayURL string        `json:"prometheus-gateway,omitempty" yaml:"prometheus-gateway,omitempty"`
	DisableConsoleOutput     bool          `json:"disable-console-output,omitempty" yaml:"disable-console-output,omitempty"`
	DisableCompression       bool          `json:"disable-compression,omitempty" yaml:"disable-compression,omitempty"`
	DisableKeepalive         bool          `json:"disable-keepalive,omitempty" yaml:"disable-keepalive,omitempty"`
	profile                  string
}

var boomArgs BoomArgs

func init() {
	rootCmd.AddCommand(boomCmd)

	boomCmd.Flags().Int64Var(&boomArgs.MaxRPS, "max-rps", 0, "Max RPS that boomer can generate, disabled by default.")
	boomCmd.Flags().StringVar(&boomArgs.RequestIncreaseRate, "request-increase-rate", "-1", "Request increase rate, disabled by default.")
	boomCmd.Flags().IntVar(&boomArgs.SpawnCount, "spawn-count", 1, "The number of users to spawn for load testing")
	boomCmd.Flags().Float64Var(&boomArgs.SpawnRate, "spawn-rate", 1, "The rate for spawning users")
	boomCmd.Flags().Int64Var(&boomArgs.LoopCount, "loop-count", -1, "The specify running cycles for load testing")
	boomCmd.Flags().StringVar(&boomArgs.MemoryProfile, "mem-profile", "", "Enable memory profiling.")
	boomCmd.Flags().DurationVar(&boomArgs.MemoryProfileDuration, "mem-profile-duration", 30*time.Second, "Memory profile duration.")
	boomCmd.Flags().StringVar(&boomArgs.CPUProfile, "cpu-profile", "", "Enable CPU profiling.")
	boomCmd.Flags().DurationVar(&boomArgs.CPUProfileDuration, "cpu-profile-duration", 30*time.Second, "CPU profile duration.")
	boomCmd.Flags().StringVar(&boomArgs.PrometheusPushgatewayURL, "prometheus-gateway", "", "Prometheus Pushgateway url.")
	boomCmd.Flags().BoolVar(&boomArgs.DisableConsoleOutput, "disable-console-output", false, "Disable console output.")
	boomCmd.Flags().BoolVar(&boomArgs.DisableCompression, "disable-compression", false, "Disable compression")
	boomCmd.Flags().BoolVar(&boomArgs.DisableKeepalive, "disable-keepalive", false, "Disable keepalive")
	boomCmd.Flags().StringVar(&boomArgs.profile, "profile", "", "profile for load testing")
}
