package adb

import (
	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v5/uixt"
	"github.com/httprunner/httprunner/v5/uixt/option"
)

var serial string

var CmdAndroidRoot = &cobra.Command{
	Use:              "adb",
	Short:            "simple utils for android device management",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {},
}

func getDevice(serial string) (*uixt.AndroidDevice, error) {
	device, err := uixt.NewAndroidDevice(option.WithSerialNumber(serial))
	if err != nil {
		return nil, err
	}
	return device, nil
}
