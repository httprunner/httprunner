package ios

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v5/internal/sdk"
)

var rebootCmd = &cobra.Command{
	Use:              "reboot",
	Short:            "reboot ios device",
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

		err = device.Reboot()
		if err != nil {
			return err
		}
		fmt.Printf("reboot %s success\n", device.Options.UDID)
		return nil
	},
}

func init() {
	rebootCmd.Flags().StringVarP(&udid, "udid", "u", "", "specify device by udid")
	CmdIOSRoot.AddCommand(rebootCmd)
}
