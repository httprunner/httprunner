package adb

import (
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
)

var screencapAndroidDevicesCmd = &cobra.Command{
	Use:   "screencap",
	Short: "Start android screen capture",
	RunE: func(cmd *cobra.Command, args []string) error {
		device, err := getDevice(serial)
		if err != nil {
			return err
		}

		res, err := device.ScreenCap()
		if err != nil {
			return err
		}

		filepath := fmt.Sprintf("%s.png", builtin.GenNameWithTimestamp("screencap_"))
		if err = ioutil.WriteFile(filepath, res, 0o644); err != nil {
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
