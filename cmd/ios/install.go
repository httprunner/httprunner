package ios

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v5/internal/sdk"
	"github.com/httprunner/httprunner/v5/pkg/uixt"
)

var installCmd = &cobra.Command{
	Use:   "install [flags] PACKAGE",
	Short: "push package to the device and install them automatically",
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

		err = device.Install(args[0])
		if err != nil {
			fmt.Println(err)
			return err
		}
		fmt.Println("success")
		return nil
	},
}

func init() {
	installCmd.Flags().StringVarP(&udid, "udid", "u", "", "filter by device's serial")

	iosRootCmd.AddCommand(installCmd)
}
