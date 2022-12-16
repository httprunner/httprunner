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

var pcapCmd = &cobra.Command{
	Use:   "pcap",
	Short: "capture ios network packets",
	RunE: func(cmd *cobra.Command, args []string) error {
		device, err := uixt.NewIOSDevice(
			uixt.WithUDID(udid),
		)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to init ios device")
		}

		err = builtin.EnsureFolderExists(env.ResultsPath)
		if err != nil {
			return err
		}

		if err = device.StartPcap(); err != nil {
			return err
		}
		defer device.StopPcap()

		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
		timer := time.NewTimer(time.Duration(timeDuration) * time.Second)
		for {
			select {
			case <-timer.C:
				return nil
			case <-c:
				log.Warn().Msg("received signal, stop pcap")
				return nil
			}
		}
	},
}

var timeDuration int

func init() {
	pcapCmd.Flags().StringVarP(&udid, "udid", "u", "", "specify device by udid")
	pcapCmd.Flags().IntVarP(&timeDuration, "duration", "t", 10, "specify time duraion in seconds")
	iosRootCmd.AddCommand(pcapCmd)
}
