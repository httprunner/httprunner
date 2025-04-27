package ios

import (
	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v5/uixt"
	"github.com/httprunner/httprunner/v5/uixt/option"
)

var CmdIOSRoot = &cobra.Command{
	Use:   "ios",
	Short: "simple utils for ios device management",
}

func getDevice(udid string) (*uixt.IOSDevice, error) {
	device, err := uixt.NewIOSDevice(option.WithUDID(udid))
	if err != nil {
		return nil, err
	}
	return device, nil
}
