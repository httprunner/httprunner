package uixt

import (
	"context"
	"fmt"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/pkg/gadb"
	"github.com/httprunner/httprunner/v5/uixt/option"
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
	return BuildMCPCallToolRequest(t.Name(), map[string]any{}, action), nil
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
	return BuildMCPCallToolRequest(t.Name(), map[string]any{}, action), nil
}

// ToolScreenRecord implements the screenrecord tool call.
type ToolScreenRecord struct {
	// Return data fields - these define the structure of data returned by this tool
	VideoPath string  `json:"videoPath" desc:"Path to the recorded video file"`
	Duration  float64 `json:"duration" desc:"Duration of the recording in seconds"`
	Method    string  `json:"method" desc:"Recording method used (adb or scrcpy)"`
}

func (t *ToolScreenRecord) Name() option.ActionName {
	return option.ACTION_ScreenRecord
}

func (t *ToolScreenRecord) Description() string {
	return "Record the screen of the mobile device. Supports both ADB screenrecord and scrcpy recording methods. ADB recording is limited to 180 seconds, while scrcpy supports longer recordings and audio capture on Android 11+."
}

func (t *ToolScreenRecord) Options() []mcp.ToolOption {
	return []mcp.ToolOption{
		mcp.WithString("platform", mcp.Enum("android", "ios"), mcp.Description("The platform type of device to record")),
		mcp.WithString("serial", mcp.Description("The device serial number or UDID")),
		mcp.WithNumber("duration", mcp.Description("Recording duration in seconds. If not specified, recording will continue until manually stopped. ADB recording is limited to 180 seconds.")),
		mcp.WithString("screenRecordPath", mcp.Description("Custom path for the output video file. If not specified, a timestamped filename will be generated.")),
		mcp.WithBoolean("screenRecordWithAudio", mcp.Description("Enable audio recording (requires scrcpy and Android 11+). Default: false")),
		mcp.WithBoolean("screenRecordWithScrcpy", mcp.Description("Force use of scrcpy for recording instead of ADB. Default: false (auto-detect based on audio requirement)")),
	}
}

func (t *ToolScreenRecord) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// Parse options from arguments
		var opts []option.ActionOption

		if duration, ok := request.Params.Arguments["duration"].(float64); ok && duration > 0 {
			opts = append(opts, option.WithDuration(duration))
		}

		if path, ok := request.Params.Arguments["screenRecordPath"].(string); ok && path != "" {
			opts = append(opts, option.WithScreenRecordPath(path))
		}

		if audio, ok := request.Params.Arguments["screenRecordWithAudio"].(bool); ok && audio {
			opts = append(opts, option.WithScreenRecordAudio(true))
		}

		if scrcpy, ok := request.Params.Arguments["screenRecordWithScrcpy"].(bool); ok && scrcpy {
			opts = append(opts, option.WithScreenRecordScrcpy(true))
		}

		// Add context to options for proper cancellation handling
		opts = append(opts, option.WithContext(ctx))

		// Start screen recording
		videoPath, err := driverExt.IDriver.ScreenRecord(opts...)
		if err != nil {
			log.Error().Err(err).Msg("ScreenRecord failed")
			return NewMCPErrorResponse("Failed to record screen: " + err.Error()), nil
		}

		// Determine recording method and duration
		options := option.NewActionOptions(opts...)
		method := "adb"
		duration := options.Duration
		if options.ScreenRecordDuration > 0 {
			duration = options.ScreenRecordDuration
		}

		if options.ScreenRecordWithScrcpy || options.ScreenRecordWithAudio {
			method = "scrcpy"
		}

		message := fmt.Sprintf("Screen recording completed successfully. Video saved to: %s", videoPath)
		returnData := ToolScreenRecord{
			VideoPath: videoPath,
			Duration:  duration,
			Method:    method,
		}

		return NewMCPSuccessResponse(message, &returnData), nil
	}
}

func (t *ToolScreenRecord) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	return BuildMCPCallToolRequest(t.Name(), map[string]any{}, action), nil
}
