package ios

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v4/hrp/internal/sdk"
	"github.com/httprunner/httprunner/v4/hrp/pkg/gidevice"
	"github.com/httprunner/httprunner/v4/hrp/pkg/uixt"
)

type Device struct {
	d               gidevice.Device
	UDID            string        `json:"UDID"`
	Status          string        `json:"status"`
	ConnectionType  string        `json:"connectionType"`
	ConnectionSpeed int           `json:"connectionSpeed"`
	DeviceDetail    *DeviceDetail `json:"deviceDetail,omitempty"`
}

type DeviceDetail struct {
	DeviceName        string `json:"deviceName,omitempty"`
	DeviceClass       string `json:"deviceClass,omitempty"`
	ProductVersion    string `json:"productVersion,omitempty"`
	ProductType       string `json:"productType,omitempty"`
	ProductName       string `json:"productName,omitempty"`
	PasswordProtected bool   `json:"passwordProtected,omitempty"`
	ModelNumber       string `json:"modelNumber,omitempty"`
	SerialNumber      string `json:"serialNumber,omitempty"`
	SIMStatus         string `json:"simStatus,omitempty"`
	PhoneNumber       string `json:"phoneNumber,omitempty"`
	CPUArchitecture   string `json:"cpuArchitecture,omitempty"`
	ProtocolVersion   string `json:"protocolVersion,omitempty"`
	RegionInfo        string `json:"regionInfo,omitempty"`
	TimeZone          string `json:"timeZone,omitempty"`
	UniqueDeviceID    string `json:"uniqueDeviceID,omitempty"`
	WiFiAddress       string `json:"wifiAddress,omitempty"`
	BuildVersion      string `json:"buildVersion,omitempty"`
}

func (device *Device) GetDetail() (*DeviceDetail, error) {
	value, err := device.d.GetValue("", "")
	if err != nil {
		return nil, errors.Wrap(err, "get device detail failed")
	}
	detailByte, _ := json.Marshal(value)
	detail := &DeviceDetail{}
	json.Unmarshal(detailByte, detail)
	return detail, nil
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

		devices, err := uixt.GetIOSDevices(udid)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}

		for _, d := range devices {
			deviceProperties := d.Properties()
			device := &Device{
				d:               d,
				UDID:            deviceProperties.SerialNumber,
				ConnectionType:  deviceProperties.ConnectionType,
				ConnectionSpeed: deviceProperties.ConnectionSpeed,
			}
			device.Status = device.GetStatus()

			if isDetail {
				device.DeviceDetail, err = device.GetDetail()
				if err != nil {
					return err
				}
				fmt.Println(device.ToFormat())
			} else {
				fmt.Println(device.UDID, device.ConnectionType, device.Status)
			}
		}
		return nil
	},
}

var (
	udid     string
	isDetail bool
)

func init() {
	listDevicesCmd.Flags().StringVarP(&udid, "udid", "u", "", "filter by device's udid")
	listDevicesCmd.Flags().BoolVarP(&isDetail, "detail", "d", false, "print device's detail")
	iosRootCmd.AddCommand(listDevicesCmd)
}
