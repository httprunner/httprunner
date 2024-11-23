package ios

import (
	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v4/hrp/pkg/uixt"
)

var iosRootCmd = &cobra.Command{
	Use:   "ios",
	Short: "simple utils for ios device management",
}

func getDevice(udid string) (*uixt.IOSDevice, error) {
	device, err := uixt.NewIOSDevice(uixt.WithUDID(udid))
	if err != nil {
		return nil, err
	}
	return device, nil
}

func Init(rootCmd *cobra.Command) {
	rootCmd.AddCommand(iosRootCmd)
}
