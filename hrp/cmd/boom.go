package cmd

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/hrp"
	"github.com/httprunner/httprunner/hrp/internal/boomer"
)

// boomCmd represents the boom command
var boomCmd = &cobra.Command{
	Use:   "boom",
	Short: "run load test with boomer",
	Long:  `run yaml/json testcase files for load test`,
	Example: `  $ hrp boom demo.json	# run specified json testcase file
  $ hrp boom demo.yaml	# run specified yaml testcase file
  $ hrp boom examples/	# run testcases in specified folder`,
	Args: cobra.MinimumNArgs(0),
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
		var hrpBoomer *hrp.HRPBoomer
		if master {
			hrpBoomer = hrp.NewMasterBoomer(masterBindHost, masterBindPort)
			if autoStart {
				hrpBoomer.SetAutoStart()
				hrpBoomer.SetExpectWorkers(expectWorkers, expectWorkersMaxWait)
				hrpBoomer.SetSpawn(spawnCount, spawnRate)
			}
			hrpBoomer.EnableGracefulQuit()
			hrpBoomer.RunMaster()
			return
		} else if worker {
			hrpBoomer = hrp.NewWorkerBoomer(masterHost, masterPort)
			go hrpBoomer.RunWorker()
			hrpBoomer.Wait()
		} else {
			hrpBoomer = hrp.NewStandaloneBoomer(spawnCount, spawnRate)
			if loopCount > 0 {
				hrpBoomer.SetLoopCount(loopCount)
			}
		}
		hrpBoomer.SetRateLimiter(maxRPS, requestIncreaseRate)
		if !disableConsoleOutput {
			hrpBoomer.AddOutput(boomer.NewConsoleOutput())
		}
		if prometheusPushgatewayURL != "" {
			hrpBoomer.AddOutput(boomer.NewPrometheusPusherOutput(prometheusPushgatewayURL, "hrp", hrpBoomer.GetMode()))
		}
		hrpBoomer.SetDisableKeepAlive(disableKeepalive)
		hrpBoomer.SetDisableCompression(disableCompression)
		hrpBoomer.EnableCPUProfile(cpuProfile, cpuProfileDuration)
		hrpBoomer.EnableMemoryProfile(memoryProfile, memoryProfileDuration)
		hrpBoomer.EnableGracefulQuit()
		hrpBoomer.Run(paths...)
	},
}

var (
	master                   bool
	worker                   bool
	masterHost               string
	masterPort               int
	masterBindHost           string
	masterBindPort           int
	autoStart                bool
	expectWorkers            int
	expectWorkersMaxWait     int
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
	disableCompression       bool
	disableKeepalive         bool
)

func init() {
	rootCmd.AddCommand(boomCmd)

	boomCmd.Flags().BoolVar(&master, "master", false, "master of distributed testing")
	boomCmd.Flags().StringVar(&masterBindHost, "master-bind-host", "127.0.0.1", "Interfaces (hostname, ip) that hrp master should bind to. Only used when running with --master. Defaults to * (all available interfaces).")
	boomCmd.Flags().IntVar(&masterBindPort, "master-bind-port", 5557, "Port that hrp master should bind to. Only used when running with --master. Defaults to 5557.")
	boomCmd.Flags().BoolVar(&worker, "worker", false, "worker of distributed testing")
	boomCmd.Flags().StringVar(&masterHost, "master-host", "127.0.0.1", "Host or IP address of hrp master for distributed load testing.")
	boomCmd.Flags().IntVar(&masterPort, "master-port", 5557, "The port to connect to that is used by the hrp master for distributed load testing.")
	boomCmd.Flags().BoolVar(&autoStart, "autostart", false, "Starts the test immediately (without disabling the web UI). Use --spawn-count and --spawn-rate to control user count and run time")
	boomCmd.Flags().IntVar(&expectWorkers, "expect-workers", 1, "How many workers master should expect to connect before starting the test (only when --autostart is used")
	boomCmd.Flags().IntVar(&expectWorkersMaxWait, "expect-workers-max-wait", 0, "How many workers master should expect to connect before starting the test (only when --autostart is used")
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
	boomCmd.Flags().BoolVar(&disableCompression, "disable-compression", false, "Disable compression")
	boomCmd.Flags().BoolVar(&disableKeepalive, "disable-keepalive", false, "Disable keepalive")
}
