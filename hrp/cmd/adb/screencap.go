package adb

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
	"github.com/httprunner/httprunner/v4/hrp/internal/sdk"
)

var screencapAndroidDevicesCmd = &cobra.Command{
	Use:   "screencap",
	Short: "Start android screen capture",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		startTime := time.Now()
		defer func() {
			sdk.SendGA4Event("hrp_adb_screencap", map[string]interface{}{
				"args":                 strings.Join(args, "-"),
				"success":              err == nil,
				"engagement_time_msec": time.Since(startTime).Milliseconds(),
			})
		}()

		device, err := getDevice(serial)
		if err != nil {
			return err
		}

		res, err := device.ScreenCap()
		if err != nil {
			return err
		}

		filepath := fmt.Sprintf("%s.png", builtin.GenNameWithTimestamp("screencap_%d"))
		if err = os.WriteFile(filepath, res, 0o644); err != nil {
			return err
		}
		fmt.Println("screencap saved to", filepath)
		return nil
	},
}

func init() {
	screencapAndroidDevicesCmd.Flags().StringVarP(&serial, "serial", "s", "", "filter by device's serial")
	androidRootCmd.AddCommand(screencapAndroidDevicesCmd)
}
