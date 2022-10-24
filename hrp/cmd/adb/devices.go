package adb

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v4/hrp/pkg/gadb"
	"github.com/httprunner/httprunner/v4/hrp/pkg/uixt"
)

func format(data map[string]string) string {
	result, _ := json.MarshalIndent(data, "", "\t")
	return string(result)
}

var listAndroidDevicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "List all Android devices",
	RunE: func(cmd *cobra.Command, args []string) error {
		devices, err := uixt.DeviceList()
		if err != nil {
			return errors.Wrap(err, "list android devices failed")
		}

		var deviceList []gadb.Device
		// filter by serial
		for _, d := range devices {
			if serial != "" && serial != d.Serial() {
				continue
			}
			deviceList = append(deviceList, d)
		}
		if serial != "" && len(deviceList) == 0 {
			fmt.Printf("no android device found for serial: %s\n", serial)
			os.Exit(1)
		}

		for _, d := range deviceList {
			if isDetail {
				fmt.Println(format(d.DeviceInfo()))
			} else {
				fmt.Println(d.Serial(), d.Usb())
			}
		}
		return nil
	},
}

var (
	serial   string
	isDetail bool
)

func init() {
	listAndroidDevicesCmd.Flags().StringVarP(&serial, "serial", "s", "", "filter by device's serial")
	listAndroidDevicesCmd.Flags().BoolVarP(&isDetail, "detail", "d", false, "print device's detail")
	androidRootCmd.AddCommand(listAndroidDevicesCmd)
}
