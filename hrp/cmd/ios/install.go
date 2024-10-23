package ios

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v4/hrp/internal/sdk"
	"github.com/httprunner/httprunner/v4/hrp/pkg/uixt"
)

var installCmd = &cobra.Command{
	Use:   "install [flags] PACKAGE",
	Short: "Push package to the device and install them atomically",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		startTime := time.Now()
		defer func() {
			sdk.SendGA4Event("hrp_adb_devices", map[string]interface{}{
				"args":                 strings.Join(args, "-"),
				"success":              err == nil,
				"engagement_time_msec": time.Since(startTime).Milliseconds(),
			})
		}()
		_, err = getDevice(udid)
		if err != nil {
			return err
		}

		device, err := uixt.NewIOSDevice(uixt.WithUDID(udid))
		if err != nil {
			fmt.Println(err)
			return err
		}
		driverExt, err := device.NewDriver()
		if err != nil {
			fmt.Println(err)
			return err
		}

		err = driverExt.Install(args[0])
		if err != nil {
			fmt.Println(err)
			return err
		}
		fmt.Println("success")
		return nil
	},
}

func init() {
	installCmd.Flags().StringVarP(&udid, "serial", "s", "", "filter by device's serial")

	iosRootCmd.AddCommand(installCmd)
}
