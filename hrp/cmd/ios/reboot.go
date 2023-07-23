package ios

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v4/hrp/internal/sdk"
)

var rebootCmd = &cobra.Command{
	Use:              "reboot",
	Short:            "reboot or shutdown ios device",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		startTime := time.Now()
		defer func() {
			sdk.SendGA4Event("hrp_ios_reboot", map[string]interface{}{
				"args":                 strings.Join(args, "-"),
				"success":              err == nil,
				"engagement_time_msec": time.Since(startTime).Milliseconds(),
			})
		}()

		device, err := getDevice(udid)
		if err != nil {
			return err
		}

		if isShutdown {
			err = device.Shutdown()
		} else {
			err = device.Reboot()
		}
		if err != nil {
			return err
		}
		fmt.Printf("reboot %s success\n", device.Properties().UDID)
		return nil
	},
}

var isShutdown bool

func init() {
	rebootCmd.Flags().StringVarP(&udid, "udid", "u", "", "specify device by udid")
	rebootCmd.Flags().BoolVarP(&isShutdown, "shutdown", "s", false, "shutdown ios device")
	iosRootCmd.AddCommand(rebootCmd)
}
