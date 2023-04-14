package adb

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v4/hrp/pkg/gadb"
	"github.com/httprunner/httprunner/v4/hrp/pkg/uixt"
)

var androidRootCmd = &cobra.Command{
	Use:              "adb",
	Short:            "simple utils for android device management",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {},
}

func getDevice(serial string) (*gadb.Device, error) {
	devices, err := uixt.GetAndroidDevices(serial)
	if err != nil {
		return nil, err
	}
	if len(devices) > 1 {
		return nil, fmt.Errorf("found multiple attached devices, please specify android serial")
	}
	return devices[0], nil
}

func Init(rootCmd *cobra.Command) {
	rootCmd.AddCommand(androidRootCmd)
}
