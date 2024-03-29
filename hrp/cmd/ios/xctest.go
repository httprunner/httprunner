package ios

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v4/hrp/internal/sdk"
)

var xctestCmd = &cobra.Command{
	Use:   "xctest",
	Short: "run xctest",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		startTime := time.Now()
		defer func() {
			sdk.SendGA4Event("hrp_ios_xctest", map[string]interface{}{
				"args":                 strings.Join(args, "-"),
				"success":              err == nil,
				"engagement_time_msec": time.Since(startTime).Milliseconds(),
			})
		}()

		if bundleID == "" {
			return fmt.Errorf("bundleID is required")
		}
		device, err := getDevice(udid)
		if err != nil {
			return err
		}

		log.Info().Str("bundleID", bundleID).Msg("run xctest")
		out, cancel, err := device.XCTest(bundleID)
		if err != nil {
			return errors.Wrap(err, "run xctest failed")
		}

		done := make(chan os.Signal, 1)
		signal.Notify(done, syscall.SIGTERM, syscall.SIGINT)

		// print xctest running logs
		go func() {
			for s := range out {
				fmt.Print(s)
			}
			done <- os.Interrupt
		}()

		<-done
		cancel()

		return nil
	},
}

var bundleID string

func init() {
	xctestCmd.Flags().StringVarP(&udid, "udid", "u", "", "filter by device's udid")
	xctestCmd.Flags().StringVarP(&bundleID, "bundleID", "b", "", "specify ios bundleID")
	iosRootCmd.AddCommand(xctestCmd)
}
