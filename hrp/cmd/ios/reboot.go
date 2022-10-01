package ios

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rebootCmd = &cobra.Command{
	Use:   "reboot",
	Short: "reboot or shutdown ios device",
	RunE: func(cmd *cobra.Command, args []string) error {
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
	rebootCmd.Flags().StringVarP(&udid, "udid", "u", "", "filter by device's udid")
	rebootCmd.Flags().BoolVarP(&isShutdown, "shutdown", "s", false, "shutdown ios device")
	iosRootCmd.AddCommand(rebootCmd)
}
