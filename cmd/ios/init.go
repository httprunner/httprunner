package ios

import (
	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v5/pkg/uixt"
	"github.com/httprunner/httprunner/v5/pkg/uixt/options"
)

var iosRootCmd = &cobra.Command{
	Use:   "ios",
	Short: "simple utils for ios device management",
}

func getDevice(udid string) (*uixt.IOSDevice, error) {
	device, err := uixt.NewIOSDevice(options.WithUDID(udid))
	if err != nil {
		return nil, err
	}
	return device, nil
}

func Init(rootCmd *cobra.Command) {
	rootCmd.AddCommand(iosRootCmd)
}
