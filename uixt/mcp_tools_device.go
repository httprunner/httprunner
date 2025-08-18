package uixt

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
		driverExt, err := setupXTDriver(ctx, request.GetArguments())
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
		arguments := request.GetArguments()
		driverExt, err := setupXTDriver(ctx, arguments)
		if err != nil {
			return nil, err
		}

		// Parse options from arguments
		var opts []option.ActionOption

		if duration, ok := arguments["duration"].(float64); ok && duration > 0 {
			opts = append(opts, option.WithDuration(duration))
		}

		if path, ok := arguments["screenRecordPath"].(string); ok && path != "" {
			opts = append(opts, option.WithScreenRecordPath(path))
		}

		if audio, ok := arguments["screenRecordWithAudio"].(bool); ok && audio {
			opts = append(opts, option.WithScreenRecordAudio(true))
		}

		if scrcpy, ok := arguments["screenRecordWithScrcpy"].(bool); ok && scrcpy {
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

// ToolPushAlbums implements the push_albums tool call.
type ToolPushAlbums struct {
	// Return data fields - these define the structure of data returned by this tool
	FilePath string `json:"filePath" desc:"Path of the file that was pushed"`
	FileUrl  string `json:"fileUrl,omitempty" desc:"URL of the file that was downloaded and pushed (if applicable)"`
	FileType string `json:"fileType" desc:"Type of the file that was pushed (image or video)"`
	Cleared  bool   `json:"cleared,omitempty" desc:"Whether albums were cleared before pushing (if applicable)"`
}

func (t *ToolPushAlbums) Name() option.ActionName {
	return option.ACTION_PushAlbums
}

func (t *ToolPushAlbums) Description() string {
	return "Push a media file (image or video) to the device's gallery. For Android, this will push the file to the DCIM/Camera directory. For iOS, this will add the file to the photo album."
}

func (t *ToolPushAlbums) Options() []mcp.ToolOption {
	return []mcp.ToolOption{
		mcp.WithString("platform", mcp.Enum("android", "ios"), mcp.Description("The platform type of device to push media to")),
		mcp.WithString("serial", mcp.Description("The device serial number or UDID")),
		mcp.WithString("filePath", mcp.Description("Path to the local media file to push to the device")),
		mcp.WithString("fileUrl", mcp.Description("URL of the media file to download and push to the device")),
		mcp.WithBoolean("cleanup", mcp.Description("Whether to delete the downloaded file after pushing it to the device")),
		mcp.WithBoolean("clearBefore", mcp.Description("Whether to clear albums before pushing (if applicable)")),
	}
}

func (t *ToolPushAlbums) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	arguments := map[string]any{}

	// Handle string param as fileUrl
	if fileUrl, ok := action.Params.(string); ok && fileUrl != "" {
		arguments["fileUrl"] = fileUrl
	}

	// Handle map params with fileUrl or filePath
	if params, ok := action.Params.(map[string]interface{}); ok {
		if fileUrl, ok := params["fileUrl"].(string); ok && fileUrl != "" {
			arguments["fileUrl"] = fileUrl
		}
		if filePath, ok := params["filePath"].(string); ok && filePath != "" {
			arguments["filePath"] = filePath
		}
		if cleanup, ok := params["cleanup"].(bool); ok {
			arguments["cleanup"] = cleanup
		}
		if clearBefore, ok := params["clearBefore"].(bool); ok {
			arguments["clearBefore"] = clearBefore
		}
	}

	return BuildMCPCallToolRequest(t.Name(), arguments, action), nil
}

func (t *ToolPushAlbums) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.GetArguments())
		if err != nil {
			return nil, err
		}

		// Get file path or URL
		filePath, hasPath := request.GetArguments()["filePath"].(string)
		fileUrl, hasUrl := request.GetArguments()["fileUrl"].(string)
		cleanup, _ := request.GetArguments()["cleanup"].(bool)
		clearBefore, _ := request.GetArguments()["clearBefore"].(bool)

		// Check if we have either path or URL
		if (!hasPath || filePath == "") && (!hasUrl || fileUrl == "") {
			return nil, fmt.Errorf("either filePath or fileUrl is required")
		}

		// If we have a URL, download it
		downloadedFile := false
		fileType := "image" // Default file type
		if hasUrl && fileUrl != "" {
			log.Info().Str("fileUrl", fileUrl).Msg("Downloading media file from URL")
			downloadedPath, err := DownloadFileByUrl(fileUrl)
			if err != nil {
				return nil, fmt.Errorf("failed to download media file from URL: %v", err)
			}

			// Detect file type and rename with proper extension
			renamedPath, err := DetectAndRenameMediaFile(downloadedPath)
			if err != nil {
				log.Warn().Err(err).Str("path", downloadedPath).Msg("Failed to detect file type or rename file, using original file")
				filePath = downloadedPath
			} else {
				filePath = renamedPath
				// Determine if it's a video based on extension
				ext := strings.ToLower(filepath.Ext(renamedPath))
				if ext == ".mp4" || ext == ".mov" || ext == ".avi" || ext == ".wmv" || ext == ".flv" || ext == ".webm" || ext == ".mkv" {
					fileType = "video"
				}
			}
			downloadedFile = true
		}

		// Clear albums before pushing if requested
		cleared := false
		if clearBefore {
			log.Info().Msg("Clearing albums before pushing new media file")
			err := driverExt.IDriver.ClearImages()
			if err != nil {
				log.Warn().Err(err).Msg("Failed to clear albums before pushing, continuing anyway")
			} else {
				cleared = true
			}
		}

		// Push the file to the device
		err = driverExt.IDriver.PushImage(filePath)
		if err != nil {
			// If we downloaded the file and failed to push it, clean up
			if downloadedFile && cleanup {
				_ = os.Remove(filePath)
			}
			return nil, err
		}

		// Clean up downloaded file if requested
		if downloadedFile && cleanup {
			log.Info().Str("filePath", filePath).Msg("Cleaning up downloaded media file")
			_ = os.Remove(filePath)
		}

		message := fmt.Sprintf("Successfully pushed %s to device", fileType)
		returnData := ToolPushAlbums{
			FilePath: filePath,
			FileType: fileType,
			Cleared:  cleared,
		}

		// Include URL in response if it was used
		if hasUrl && fileUrl != "" {
			returnData.FileUrl = fileUrl
			message = fmt.Sprintf("Successfully downloaded and pushed %s from %s to device", fileType, fileUrl)
		}

		// Add cleared info to message if applicable
		if cleared {
			message = fmt.Sprintf("%s (albums cleared before pushing)", message)
		}

		return NewMCPSuccessResponse(message, &returnData), nil
	}
}

// Old ToolPushImage implementation has been removed as part of the refactoring to ToolPushAlbums

// ToolClearAlbums implements the clear_albums tool call.
type ToolClearAlbums struct {
	// Return data fields - these define the structure of data returned by this tool
	Cleared bool `json:"cleared" desc:"Whether albums were cleared successfully"`
}

func (t *ToolClearAlbums) Name() option.ActionName {
	return option.ACTION_ClearAlbums
}

func (t *ToolClearAlbums) Description() string {
	return "Clear media files (images and videos) from the device's gallery. For Android, this will clear media from the DCIM/Camera directory. For iOS, this will clear media from the device's photo album."
}

func (t *ToolClearAlbums) Options() []mcp.ToolOption {
	return []mcp.ToolOption{
		mcp.WithString("platform", mcp.Enum("android", "ios"), mcp.Description("The platform type of device to clear media from")),
		mcp.WithString("serial", mcp.Description("The device serial number or UDID")),
	}
}

func (t *ToolClearAlbums) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.GetArguments())
		if err != nil {
			return nil, err
		}

		err = driverExt.IDriver.ClearImages()
		if err != nil {
			return nil, err
		}

		message := "Successfully cleared media files from device"
		returnData := ToolClearAlbums{Cleared: true}

		return NewMCPSuccessResponse(message, &returnData), nil
	}
}

func (t *ToolClearAlbums) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	return BuildMCPCallToolRequest(t.Name(), map[string]any{}, action), nil
}
