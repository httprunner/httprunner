package ios

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v4/hrp/internal/sdk"
	"github.com/httprunner/httprunner/v4/hrp/pkg/uixt"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall [flags] PACKAGE",
	Short: "uninstall Package atomically",
	Args:  cobra.MinimumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		startTime := time.Now()
		defer func() {
			sdk.SendGA4Event("hrp_adb_devices", map[string]interface{}{
				"args":                 strings.Join(args, "-"),
				"success":              err == nil,
				"engagement_time_msec": time.Since(startTime).Milliseconds(),
			})
		}()
		if len(bundleId) == 0 {
			return fmt.Errorf("bundleId is empty")
		}

		_, err = getDevice(udid)
		if err != nil {
			return err
		}

		device, err := uixt.NewIOSDevice(uixt.WithUDID(udid))
		if err != nil {
			fmt.Println(err)
			return err
		}

		err = device.Uninstall(bundleId)
		if err != nil {
			fmt.Println(err)
			return err
		}
		fmt.Println("success")
		return nil
	},
}

var bundleId string

func init() {
	uninstallCmd.Flags().StringVarP(&udid, "udid", "u", "", "filter by device's serial")
	uninstallCmd.Flags().StringVarP(&bundleId, "bundleId", "b", "", "bundleId to uninstall")

	iosRootCmd.AddCommand(uninstallCmd)
}
