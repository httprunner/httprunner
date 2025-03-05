package ios

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/sdk"
	"github.com/httprunner/httprunner/v5/uixt"
)

type Device struct {
	d               ios.DeviceEntry
	UDID            string             `json:"UDID"`
	Status          string             `json:"status"`
	ConnectionType  string             `json:"connectionType"`
	ConnectionSpeed int                `json:"connectionSpeed"`
	DeviceDetail    *uixt.DeviceDetail `json:"deviceDetail,omitempty"`
}

func (device *Device) GetStatus() string {
	if device.ConnectionType != "" {
		return "online"
	} else {
		return "offline"
	}
}

func (device *Device) ToFormat() string {
	result, _ := json.MarshalIndent(device, "", "\t")
	return string(result)
}

var listDevicesCmd = &cobra.Command{
	Use:              "devices",
	Short:            "List all iOS devices",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		startTime := time.Now()
		defer func() {
			sdk.SendGA4Event("hrp_ios_devices", map[string]interface{}{
				"args":                 strings.Join(args, "-"),
				"success":              err == nil,
				"engagement_time_msec": time.Since(startTime).Milliseconds(),
			})
		}()

		devices, err := ios.ListDevices()
		if err != nil {
			return errors.Wrap(code.DeviceConnectionError,
				fmt.Sprintf("list ios devices failed: %v", err))
		}

		for _, d := range devices.DeviceList {
			deviceProperties := d.Properties
			device := &Device{
				d:               d,
				UDID:            deviceProperties.SerialNumber,
				ConnectionType:  deviceProperties.ConnectionType,
				ConnectionSpeed: deviceProperties.ConnectionSpeed,
			}
			fmt.Println(device.UDID, device.ConnectionType, device.GetStatus())
		}
		return nil
	},
}

var udid string

func init() {
	iosRootCmd.AddCommand(listDevicesCmd)
}
