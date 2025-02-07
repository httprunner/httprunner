package adb

import (
	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v5/pkg/uixt"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
)

var serial string

var androidRootCmd = &cobra.Command{
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

func Init(rootCmd *cobra.Command) {
	rootCmd.AddCommand(androidRootCmd)
}
