package uixt

import (
	"context"
	"fmt"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/httprunner/httprunner/v5/pkg/gadb"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rs/zerolog/log"
)

// ToolListAvailableDevices implements the list_available_devices tool call.
type ToolListAvailableDevices struct {
	// Return data fields - these define the structure of data returned by this tool
	AndroidDevices []string `json:"androidDevices" desc:"List of Android device serial numbers"`
	IosDevices     []string `json:"iosDevices" desc:"List of iOS device UDIDs"`
	TotalCount     int      `json:"totalCount" desc:"Total number of available devices"`
	AndroidCount   int      `json:"androidCount" desc:"Number of Android devices"`
	IosCount       int      `json:"iosCount" desc:"Number of iOS devices"`
}

func (t *ToolListAvailableDevices) Name() option.ActionName {
	return option.ACTION_ListAvailableDevices
}

func (t *ToolListAvailableDevices) Description() string {
	return "List all available devices including Android devices and iOS devices. If there are multiple devices returned, you need to let the user select one of them."
}

func (t *ToolListAvailableDevices) Options() []mcp.ToolOption {
	return []mcp.ToolOption{}
}

func (t *ToolListAvailableDevices) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		deviceList := make(map[string][]string)
		if client, err := gadb.NewClient(); err == nil {
			if androidDevices, err := client.DeviceList(); err == nil {
				serialList := make([]string, 0, len(androidDevices))
				for _, device := range androidDevices {
					serialList = append(serialList, device.Serial())
				}
				deviceList["androidDevices"] = serialList
			}
		}
		if iosDevices, err := ios.ListDevices(); err == nil {
			serialList := make([]string, 0, len(iosDevices.DeviceList))
			for _, dev := range iosDevices.DeviceList {
				device, err := NewIOSDevice(
					option.WithUDID(dev.Properties.SerialNumber))
				if err != nil {
					continue
				}
				properties := device.Properties
				err = ios.Pair(dev)
				if err != nil {
					log.Error().Err(err).Msg("failed to pair device")
					continue
				}
				serialList = append(serialList, properties.SerialNumber)
			}
			deviceList["iosDevices"] = serialList
		}

		// Create structured response
		totalDevices := len(deviceList["androidDevices"]) + len(deviceList["iosDevices"])
		message := fmt.Sprintf("Found %d available devices (%d Android, %d iOS)",
			totalDevices, len(deviceList["androidDevices"]), len(deviceList["iosDevices"]))
		returnData := ToolListAvailableDevices{
			AndroidDevices: deviceList["androidDevices"],
			IosDevices:     deviceList["iosDevices"],
			TotalCount:     totalDevices,
			AndroidCount:   len(deviceList["androidDevices"]),
			IosCount:       len(deviceList["iosDevices"]),
		}

		return NewMCPSuccessResponse(message, &returnData), nil
	}
}

func (t *ToolListAvailableDevices) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	return buildMCPCallToolRequest(t.Name(), map[string]any{}), nil
}

// ToolSelectDevice implements the select_device tool call.
type ToolSelectDevice struct {
	// Return data fields - these define the structure of data returned by this tool
	DeviceUUID string `json:"deviceUUID" desc:"UUID of the selected device"`
}

func (t *ToolSelectDevice) Name() option.ActionName {
	return option.ACTION_SelectDevice
}

func (t *ToolSelectDevice) Description() string {
	return "Select a device to use from the list of available devices. Use the list_available_devices tool first to get a list of available devices."
}

func (t *ToolSelectDevice) Options() []mcp.ToolOption {
	return []mcp.ToolOption{
		mcp.WithString("platform", mcp.Enum("android", "ios"), mcp.Description("The platform type of device to select")),
		mcp.WithString("serial", mcp.Description("The device serial number or UDID to select")),
	}
}

func (t *ToolSelectDevice) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		uuid := driverExt.IDriver.GetDevice().UUID()
		message := fmt.Sprintf("Selected device: %s", uuid)
		returnData := ToolSelectDevice{DeviceUUID: uuid}

		return NewMCPSuccessResponse(message, &returnData), nil
	}
}

func (t *ToolSelectDevice) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	return buildMCPCallToolRequest(t.Name(), map[string]any{}), nil
}
