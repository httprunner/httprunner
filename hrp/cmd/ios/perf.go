package ios

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
	"github.com/httprunner/httprunner/v4/hrp/internal/env"
	"github.com/httprunner/httprunner/v4/hrp/pkg/uixt"
)

var perfCmd = &cobra.Command{
	Use:   "perf",
	Short: "capture ios performance data (cpu,mem,disk,net,fps,etc.)",
	RunE: func(cmd *cobra.Command, args []string) error {
		perfOptions := []uixt.IOSPerfOption{}
		for _, p := range indicators {
			switch p {
			case "sys_cpu":
				perfOptions = append(perfOptions, uixt.WithIOSPerfSystemCPU(true))
			case "sys_mem":
				perfOptions = append(perfOptions, uixt.WithIOSPerfSystemMem(true))
			case "sys_net":
				perfOptions = append(perfOptions, uixt.WithIOSPerfSystemNetwork(true))
			case "sys_disk":
				perfOptions = append(perfOptions, uixt.WithIOSPerfSystemDisk(true))
			case "network":
				perfOptions = append(perfOptions, uixt.WithIOSPerfNetwork(true))
			case "fps":
				perfOptions = append(perfOptions, uixt.WithIOSPerfFPS(true))
			case "gpu":
				perfOptions = append(perfOptions, uixt.WithIOSPerfGPU(true))
			}
		}
		perfOptions = append(perfOptions, uixt.WithIOSPerfOutputInterval(interval*1000))

		device, err := uixt.NewIOSDevice(
			uixt.WithUDID(udid),
			uixt.WithIOSPerfOptions(perfOptions...),
		)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to init ios device")
		}

		err = builtin.EnsureFolderExists(env.ResultsPath)
		if err != nil {
			return err
		}

		if err = device.StartPerf(); err != nil {
			return err
		}
		defer device.StopPerf()

		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
		timer := time.NewTimer(time.Duration(timeDuration) * time.Second)
		for {
			select {
			case <-timer.C:
				return nil
			case <-c:
				log.Warn().Msg("received signal, stop perf")
				return nil
			}
		}
	},
}

var (
	interval   int
	indicators []string
)

func init() {
	perfCmd.Flags().StringVarP(&udid, "udid", "u", "", "specify device by udid")
	perfCmd.Flags().StringSliceVarP(&indicators, "indicators", "p", []string{"sys_cpu", "sys_mem"},
		"specify performance monitor, e.g. sys_cpu,sys_mem,sys_net,sys_disk,fps,network,gpu")
	perfCmd.Flags().IntVarP(&timeDuration, "duration", "t", 10, "specify time duraion in seconds")
	perfCmd.Flags().IntVarP(&interval, "interval", "i", 3, "set interval in seconds")
	iosRootCmd.AddCommand(perfCmd)
}
