package uixt

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/internal/version"
	"github.com/httprunner/httprunner/v5/pkg/gadb"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/httprunner/httprunner/v5/uixt/types"
)

// MCPServer4XTDriver provides MCP (Model Context Protocol) interface for XTDriver.
//
// This implementation adopts a pure ActionTool-style architecture where:
//   - Each MCP tool is implemented as a struct that implements the ActionTool interface
//   - Operation logic is directly embedded in each tool's Implement() method
//   - No intermediate action methods or coupling between tools
//   - Complete decoupling from the original large switch-case DoAction method
//
// Architecture:
//   MCP Request -> ActionTool.Implement() -> Direct Driver Method Call
//
// Benefits:
//   - True ActionTool interface consistency across all tools
//   - Complete decoupling with no method interdependencies
//   - Unified code organization in a single file
//   - Simplified error handling and logging per tool
//   - Easy extensibility for new features

// NewMCPServer creates a new MCP server for XTDriver and registers all tools.
func NewMCPServer() *MCPServer4XTDriver {
	mcpServer := server.NewMCPServer(
		"uixt",
		version.GetVersionInfo(),
		server.WithToolCapabilities(false),
	)
	s := &MCPServer4XTDriver{
		mcpServer:     mcpServer,
		actionToolMap: make(map[option.ActionMethod]ActionTool),
	}
	s.registerTools()
	return s
}

// MCPServer4XTDriver wraps a MCPServer to expose XTDriver functionality via MCP protocol.
type MCPServer4XTDriver struct {
	mcpServer     *server.MCPServer
	mcpTools      []mcp.Tool                         // tools list for uixt
	actionToolMap map[option.ActionMethod]ActionTool // action method to tool mapping
}

// Start runs the MCP server (blocking).
func (s *MCPServer4XTDriver) Start() error {
	log.Info().Msg("Starting HttpRunner UIXT MCP Server...")
	return server.ServeStdio(s.mcpServer)
}

// ListTools returns all registered tools
func (s *MCPServer4XTDriver) ListTools() []mcp.Tool {
	return s.mcpTools
}

// GetTool returns a pointer to the mcp.Tool with the given name
func (s *MCPServer4XTDriver) GetTool(name string) *mcp.Tool {
	for i := range s.mcpTools {
		if s.mcpTools[i].Name == name {
			return &s.mcpTools[i]
		}
	}
	return nil
}

// GetToolByActionMethod returns the tool that handles the given action method
func (s *MCPServer4XTDriver) GetToolByActionMethod(actionMethod option.ActionMethod) ActionTool {
	if s.actionToolMap == nil {
		return nil
	}
	return s.actionToolMap[actionMethod]
}

// registerTools registers all MCP tools.
func (s *MCPServer4XTDriver) registerTools() {
	// Device Tool
	s.registerTool(&ToolListAvailableDevices{}) // ListAvailableDevices
	s.registerTool(&ToolSelectDevice{})         // SelectDevice

	// Tap Tools
	s.registerTool(&ToolTapXY{})       // tap xy
	s.registerTool(&ToolTapAbsXY{})    // tap abs xy
	s.registerTool(&ToolTapByOCR{})    // tap by OCR
	s.registerTool(&ToolTapByCV{})     // tap by CV
	s.registerTool(&ToolDoubleTapXY{}) // double tap xy

	// Swipe Tool
	s.registerTool(&ToolSwipeDirection{})  // swipe direction, up/down/left/right
	s.registerTool(&ToolSwipeCoordinate{}) // swipe coordinate, [fromX, fromY, toX, toY]
	s.registerTool(&ToolSwipeToTapApp{})
	s.registerTool(&ToolSwipeToTapText{})
	s.registerTool(&ToolSwipeToTapTexts{})

	// Drag Tool
	s.registerTool(&ToolDrag{})

	// Input Tool
	s.registerTool(&ToolInput{})

	// ScreenShot Tool
	s.registerTool(&ToolScreenShot{})

	// GetScreenSize Tool
	s.registerTool(&ToolGetScreenSize{})

	// PressButton Tool
	s.registerTool(&ToolPressButton{})
	s.registerTool(&ToolHome{}) // Home
	s.registerTool(&ToolBack{}) // Back

	// App actions
	s.registerTool(&ToolListPackages{}) // ListPackages
	s.registerTool(&ToolLaunchApp{})    // LaunchApp
	s.registerTool(&ToolTerminateApp{}) // TerminateApp
	s.registerTool(&ToolAppInstall{})   // AppInstall
	s.registerTool(&ToolAppUninstall{}) // AppUninstall
	s.registerTool(&ToolAppClear{})     // AppClear

	// Sleep Tool
	s.registerTool(&ToolSleep{})
	s.registerTool(&ToolSleepMS{})
	s.registerTool(&ToolSleepRandom{})

	// Utils tools
	s.registerTool(&ToolSetIme{})
	s.registerTool(&ToolGetSource{})
	s.registerTool(&ToolClosePopups{})
	s.registerTool(&ToolAIAction{})

	// PC/Web actions
	s.registerTool(&ToolWebLoginNoneUI{})
	s.registerTool(&ToolSecondaryClick{})
	s.registerTool(&ToolHoverBySelector{})
	s.registerTool(&ToolTapBySelector{})
	s.registerTool(&ToolSecondaryClickBySelector{})
	s.registerTool(&ToolWebCloseTab{})
}

func (s *MCPServer4XTDriver) registerTool(tool ActionTool) {
	options := []mcp.ToolOption{
		mcp.WithDescription(tool.Description()),
	}
	options = append(options, tool.Options()...)

	toolName := string(tool.Name())
	mcpTool := mcp.NewTool(toolName, options...)
	s.mcpServer.AddTool(mcpTool, tool.Implement())

	s.mcpTools = append(s.mcpTools, mcpTool)
	s.actionToolMap[tool.Name()] = tool

	log.Debug().Str("name", toolName).Str("type", toolName).Msg("register tool")
}

// ActionTool interface defines the contract for MCP tools
type ActionTool interface {
	Name() option.ActionMethod
	Description() string
	Options() []mcp.ToolOption
	Implement() server.ToolHandlerFunc
	// ConvertActionToCallToolRequest converts MobileAction to mcp.CallToolRequest
	ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error)
}

// buildMCPCallToolRequest is a helper function to build mcp.CallToolRequest
func buildMCPCallToolRequest(toolName option.ActionMethod, arguments map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: struct {
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments,omitempty"`
			Meta      *struct {
				ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
			} `json:"_meta,omitempty"`
		}{
			Name:      string(toolName),
			Arguments: arguments,
		},
	}
}

// ToolListAvailableDevices implements the list_available_devices tool call.
type ToolListAvailableDevices struct{}

func (t *ToolListAvailableDevices) Name() option.ActionMethod {
	return option.ACTION_ListAvailableDevices
}

func (t *ToolListAvailableDevices) Description() string {
	return "List all available devices. If there are more than one device returned, you need to let the user select one of them."
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

		jsonResult, _ := json.Marshal(deviceList)
		return mcp.NewToolResultText(string(jsonResult)), nil
	}
}

func (t *ToolListAvailableDevices) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	return buildMCPCallToolRequest(t.Name(), map[string]any{}), nil
}

// ToolSelectDevice implements the select_device tool call.
type ToolSelectDevice struct{}

func (t *ToolSelectDevice) Name() option.ActionMethod {
	return option.ACTION_SelectDevice
}

func (t *ToolSelectDevice) Description() string {
	return "Select a device to use from the list of available devices. Use the list_available_devices tool to get a list of available devices."
}

func (t *ToolSelectDevice) Options() []mcp.ToolOption {
	return []mcp.ToolOption{
		mcp.WithString("platform", mcp.Enum("android", "ios"), mcp.Description("The type of device to select")),
		mcp.WithString("serial", mcp.Description("The device serial/udid to select")),
	}
}

func (t *ToolSelectDevice) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		uuid := driverExt.IDriver.GetDevice().UUID()
		return mcp.NewToolResultText(fmt.Sprintf("Selected device: %s", uuid)), nil
	}
}

func (t *ToolSelectDevice) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	return buildMCPCallToolRequest(t.Name(), map[string]any{}), nil
}

// ToolTapXY implements the tap_xy tool call.
type ToolTapXY struct{}

func (t *ToolTapXY) Name() option.ActionMethod {
	return option.ACTION_TapXY
}

func (t *ToolTapXY) Description() string {
	return "Click on the screen at given x,y coordinates"
}

func (t *ToolTapXY) Options() []mcp.ToolOption {
	return option.NewMCPOptions(option.TapRequest{})
}

func (t *ToolTapXY) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		var tapReq option.TapRequest
		if err := mapToStruct(request.Params.Arguments, &tapReq); err != nil {
			return nil, fmt.Errorf("parse parameters error: %w", err)
		}

		// Tap action logic
		log.Info().Float64("x", tapReq.X).Float64("y", tapReq.Y).Msg("tapping at coordinates")
		opts := []option.ActionOption{
			option.WithDuration(tapReq.Duration),
			option.WithPreMarkOperation(true),
		}

		err = driverExt.TapXY(tapReq.X, tapReq.Y, opts...)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Tap failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully tapped at coordinates (%.2f, %.2f)", tapReq.X, tapReq.Y)), nil
	}
}

func (t *ToolTapXY) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	if params, err := builtin.ConvertToFloat64Slice(action.Params); err == nil && len(params) == 2 {
		x, y := params[0], params[1]
		arguments := map[string]any{
			"x": x,
			"y": y,
		}
		// Add duration if available from action options
		if duration := action.ActionOptions.Duration; duration > 0 {
			arguments["duration"] = duration
		}
		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid tap params: %v", action.Params)
}

// ToolTapAbsXY implements the tap_abs_xy tool call.
type ToolTapAbsXY struct{}

func (t *ToolTapAbsXY) Name() option.ActionMethod {
	return option.ACTION_TapAbsXY
}

func (t *ToolTapAbsXY) Description() string {
	return "Tap at absolute pixel coordinates"
}

func (t *ToolTapAbsXY) Options() []mcp.ToolOption {
	return option.NewMCPOptions(option.TapAbsXYRequest{})
}

func (t *ToolTapAbsXY) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		var tapAbsReq option.TapAbsXYRequest
		if err := mapToStruct(request.Params.Arguments, &tapAbsReq); err != nil {
			return nil, fmt.Errorf("parse parameters error: %w", err)
		}

		// Tap absolute XY action logic
		log.Info().Float64("x", tapAbsReq.X).Float64("y", tapAbsReq.Y).Msg("tapping at absolute coordinates")
		opts := []option.ActionOption{}
		if tapAbsReq.Duration > 0 {
			opts = append(opts, option.WithDuration(tapAbsReq.Duration))
		}

		err = driverExt.TapAbsXY(tapAbsReq.X, tapAbsReq.Y, opts...)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Tap absolute XY failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully tapped at absolute coordinates (%.0f, %.0f)", tapAbsReq.X, tapAbsReq.Y)), nil
	}
}

func (t *ToolTapAbsXY) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	if params, err := builtin.ConvertToFloat64Slice(action.Params); err == nil && len(params) == 2 {
		x, y := params[0], params[1]
		arguments := map[string]any{
			"x": x,
			"y": y,
		}
		// Add duration if available
		if duration := action.ActionOptions.Duration; duration > 0 {
			arguments["duration"] = duration
		}
		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid tap abs params: %v", action.Params)
}

// ToolTapByOCR implements the tap_by_ocr tool call.
type ToolTapByOCR struct{}

func (t *ToolTapByOCR) Name() option.ActionMethod {
	return option.ACTION_TapByOCR
}

func (t *ToolTapByOCR) Description() string {
	return "Tap on text found by OCR recognition"
}

func (t *ToolTapByOCR) Options() []mcp.ToolOption {
	return option.NewMCPOptions(option.TapByOCRRequest{})
}

func (t *ToolTapByOCR) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		var ocrReq option.TapByOCRRequest
		if err := mapToStruct(request.Params.Arguments, &ocrReq); err != nil {
			return nil, fmt.Errorf("parse parameters error: %w", err)
		}

		// Tap by OCR action logic
		log.Info().Str("text", ocrReq.Text).Msg("tapping by OCR")
		err = driverExt.TapByOCR(ocrReq.Text)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Tap by OCR failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully tapped on OCR text: %s", ocrReq.Text)), nil
	}
}

func (t *ToolTapByOCR) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	if text, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"text": text,
		}
		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid tap by OCR params: %v", action.Params)
}

// ToolTapByCV implements the tap_by_cv tool call.
type ToolTapByCV struct{}

func (t *ToolTapByCV) Name() option.ActionMethod {
	return option.ACTION_TapByCV
}

func (t *ToolTapByCV) Description() string {
	return "Tap on element found by computer vision"
}

func (t *ToolTapByCV) Options() []mcp.ToolOption {
	return option.NewMCPOptions(option.TapByCVRequest{})
}

func (t *ToolTapByCV) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		var cvReq option.TapByCVRequest
		if err := mapToStruct(request.Params.Arguments, &cvReq); err != nil {
			return nil, fmt.Errorf("parse parameters error: %w", err)
		}

		// Tap by CV action logic
		log.Info().Str("imagePath", cvReq.ImagePath).Msg("tapping by CV")

		// For TapByCV, we need to check if there are UI types in the options
		// In the original DoAction, it requires ScreenShotWithUITypes to be set
		// We'll add a basic implementation that triggers CV recognition
		err = driverExt.TapByCV()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Tap by CV failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText("Successfully tapped by computer vision"), nil
	}
}

func (t *ToolTapByCV) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	// For TapByCV, the original action might not have params but relies on options
	arguments := map[string]any{
		"imagePath": "", // Will be handled by the tool based on UI types
	}
	return buildMCPCallToolRequest(t.Name(), arguments), nil
}

// ToolDoubleTapXY implements the double_tap_xy tool call.
type ToolDoubleTapXY struct{}

func (t *ToolDoubleTapXY) Name() option.ActionMethod {
	return option.ACTION_DoubleTapXY
}

func (t *ToolDoubleTapXY) Description() string {
	return "Double tap at given coordinates"
}

func (t *ToolDoubleTapXY) Options() []mcp.ToolOption {
	return option.NewMCPOptions(option.DoubleTapXYRequest{})
}

func (t *ToolDoubleTapXY) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		var doubleTapReq option.DoubleTapXYRequest
		if err := mapToStruct(request.Params.Arguments, &doubleTapReq); err != nil {
			return nil, fmt.Errorf("parse parameters error: %w", err)
		}

		// Double tap XY action logic
		log.Info().Float64("x", doubleTapReq.X).Float64("y", doubleTapReq.Y).Msg("double tapping at coordinates")
		err = driverExt.DoubleTap(doubleTapReq.X, doubleTapReq.Y)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Double tap failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully double tapped at (%.2f, %.2f)", doubleTapReq.X, doubleTapReq.Y)), nil
	}
}

func (t *ToolDoubleTapXY) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	if params, err := builtin.ConvertToFloat64Slice(action.Params); err == nil && len(params) == 2 {
		x, y := params[0], params[1]
		arguments := map[string]any{
			"x": x,
			"y": y,
		}
		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid double tap params: %v", action.Params)
}

// ToolListPackages implements the list_packages tool call.
type ToolListPackages struct{}

func (t *ToolListPackages) Name() option.ActionMethod {
	return option.ACTION_ListPackages
}

func (t *ToolListPackages) Description() string {
	return "List all the apps/packages on the device."
}

func (t *ToolListPackages) Options() []mcp.ToolOption {
	return option.NewMCPOptions(option.TargetDeviceRequest{})
}

func (t *ToolListPackages) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		apps, err := driverExt.IDriver.GetDevice().ListPackages()
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText(fmt.Sprintf("Device packages: %v", apps)), nil
	}
}

func (t *ToolListPackages) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	return buildMCPCallToolRequest(t.Name(), map[string]any{}), nil
}

// ToolLaunchApp implements the launch_app tool call.
type ToolLaunchApp struct{}

func (t *ToolLaunchApp) Name() option.ActionMethod {
	return option.ACTION_AppLaunch
}

func (t *ToolLaunchApp) Description() string {
	return "Launch an app on mobile device. Use this to open a specific app. You can find the package name of the app by calling list_packages."
}

func (t *ToolLaunchApp) Options() []mcp.ToolOption {
	return option.NewMCPOptions(option.AppLaunchRequest{})
}

func (t *ToolLaunchApp) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		var appLaunchReq option.AppLaunchRequest
		if err := mapToStruct(request.Params.Arguments, &appLaunchReq); err != nil {
			return nil, fmt.Errorf("parse parameters error: %w", err)
		}

		if appLaunchReq.PackageName == "" {
			return nil, fmt.Errorf("package_name is required")
		}

		// Launch app action logic
		log.Info().Str("packageName", appLaunchReq.PackageName).Msg("launching app")
		err = driverExt.AppLaunch(appLaunchReq.PackageName)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Launch app failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully launched app: %s", appLaunchReq.PackageName)), nil
	}
}

func (t *ToolLaunchApp) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	if packageName, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"packageName": packageName,
		}
		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid app launch params: %v", action.Params)
}

// ToolTerminateApp implements the terminate_app tool call.
type ToolTerminateApp struct{}

func (t *ToolTerminateApp) Name() option.ActionMethod {
	return option.ACTION_AppTerminate
}

func (t *ToolTerminateApp) Description() string {
	return "Stop and terminate an app on mobile device"
}

func (t *ToolTerminateApp) Options() []mcp.ToolOption {
	return option.NewMCPOptions(option.AppTerminateRequest{})
}

func (t *ToolTerminateApp) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		var appTerminateReq option.AppTerminateRequest
		if err := mapToStruct(request.Params.Arguments, &appTerminateReq); err != nil {
			return nil, fmt.Errorf("parse parameters error: %w", err)
		}

		if appTerminateReq.PackageName == "" {
			return nil, fmt.Errorf("package_name is required")
		}

		// Terminate app action logic
		log.Info().Str("packageName", appTerminateReq.PackageName).Msg("terminating app")
		success, err := driverExt.AppTerminate(appTerminateReq.PackageName)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Terminate app failed: %s", err.Error())), nil
		}
		if !success {
			log.Warn().Str("packageName", appTerminateReq.PackageName).Msg("app was not running")
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully terminated app: %s", appTerminateReq.PackageName)), nil
	}
}

func (t *ToolTerminateApp) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	if packageName, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"packageName": packageName,
		}
		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid app terminate params: %v", action.Params)
}

// ToolScreenShot implements the screenshot tool call.
type ToolScreenShot struct{}

func (t *ToolScreenShot) Name() option.ActionMethod {
	return option.ACTION_ScreenShot
}

func (t *ToolScreenShot) Description() string {
	return "Take a screenshot of the mobile device. Use this to understand what's on screen. Do not cache this result."
}

func (t *ToolScreenShot) Options() []mcp.ToolOption {
	return option.NewMCPOptions(option.TargetDeviceRequest{})
}

func (t *ToolScreenShot) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, err
		}
		bufferBase64, err := GetScreenShotBufferBase64(driverExt.IDriver)
		if err != nil {
			log.Error().Err(err).Msg("ScreenShot failed")
			return mcp.NewToolResultError(fmt.Sprintf("Failed to take screenshot: %v", err)), nil
		}
		log.Debug().Int("imageBytes", len(bufferBase64)).Msg("take screenshot success")

		return mcp.NewToolResultImage("screenshot", bufferBase64, "image/jpeg"), nil
	}
}

func (t *ToolScreenShot) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	return buildMCPCallToolRequest(t.Name(), map[string]any{}), nil
}

// ToolGetScreenSize implements the get_screen_size tool call.
type ToolGetScreenSize struct{}

func (t *ToolGetScreenSize) Name() option.ActionMethod {
	return option.ACTION_GetScreenSize
}

func (t *ToolGetScreenSize) Description() string {
	return "Get the screen size of the mobile device in pixels"
}

func (t *ToolGetScreenSize) Options() []mcp.ToolOption {
	return option.NewMCPOptions(option.TargetDeviceRequest{})
}

func (t *ToolGetScreenSize) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		screenSize, err := driverExt.IDriver.WindowSize()
		if err != nil {
			return mcp.NewToolResultError("Get screen size failed: " + err.Error()), nil
		}
		return mcp.NewToolResultText(
			fmt.Sprintf("Screen size: %d x %d pixels", screenSize.Width, screenSize.Height),
		), nil
	}
}

func (t *ToolGetScreenSize) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	return buildMCPCallToolRequest(t.Name(), map[string]any{}), nil
}

// ToolPressButton implements the press_button tool call.
type ToolPressButton struct{}

func (t *ToolPressButton) Name() option.ActionMethod {
	return option.ACTION_PressButton
}

func (t *ToolPressButton) Description() string {
	return "Press a button on the device"
}

func (t *ToolPressButton) Options() []mcp.ToolOption {
	return option.NewMCPOptions(option.PressButtonRequest{})
}

func (t *ToolPressButton) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		var pressButtonReq option.PressButtonRequest
		if err := mapToStruct(request.Params.Arguments, &pressButtonReq); err != nil {
			return nil, fmt.Errorf("parse parameters error: %w", err)
		}

		// Press button action logic
		log.Info().Str("button", string(pressButtonReq.Button)).Msg("pressing button")
		err = driverExt.PressButton(types.DeviceButton(pressButtonReq.Button))
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Press button failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully pressed button: %s", pressButtonReq.Button)), nil
	}
}

func (t *ToolPressButton) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	if button, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"button": button,
		}
		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid press button params: %v", action.Params)
}

// ToolSwipeDirection implements the swipe tool call.
type ToolSwipeDirection struct{}

func (t *ToolSwipeDirection) Name() option.ActionMethod {
	return option.ACTION_SwipeDirection
}

func (t *ToolSwipeDirection) Description() string {
	return "Swipe on the screen"
}

func (t *ToolSwipeDirection) Options() []mcp.ToolOption {
	return option.NewMCPOptions(option.SwipeRequest{})
}

func (t *ToolSwipeDirection) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		var swipeReq option.SwipeRequest
		if err := mapToStruct(request.Params.Arguments, &swipeReq); err != nil {
			return nil, fmt.Errorf("parse parameters error: %w", err)
		}

		// Swipe action logic
		log.Info().Str("direction", swipeReq.Direction).Msg("performing swipe")

		// Validate direction
		validDirections := []string{"up", "down", "left", "right"}
		isValid := false
		for _, validDir := range validDirections {
			if swipeReq.Direction == validDir {
				isValid = true
				break
			}
		}
		if !isValid {
			return nil, fmt.Errorf("invalid swipe direction: %s, expected one of: %v", swipeReq.Direction, validDirections)
		}

		opts := []option.ActionOption{
			option.WithPreMarkOperation(true),
			option.WithDuration(swipeReq.Duration),
			option.WithPressDuration(swipeReq.PressDuration),
		}

		// Convert direction to coordinates and perform swipe
		switch swipeReq.Direction {
		case "up":
			err = driverExt.Swipe(0.5, 0.5, 0.5, 0.1, opts...)
		case "down":
			err = driverExt.Swipe(0.5, 0.5, 0.5, 0.9, opts...)
		case "left":
			err = driverExt.Swipe(0.5, 0.5, 0.1, 0.5, opts...)
		case "right":
			err = driverExt.Swipe(0.5, 0.5, 0.9, 0.5, opts...)
		default:
			return mcp.NewToolResultError(fmt.Sprintf("Unexpected swipe direction: %s", swipeReq.Direction)), nil
		}

		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Swipe failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully swiped %s", swipeReq.Direction)), nil
	}
}

func (t *ToolSwipeDirection) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	// Handle direction swipe like "up", "down", "left", "right"
	if direction, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"direction": direction,
		}
		// Add duration and press duration from options
		if duration := action.ActionOptions.Duration; duration > 0 {
			arguments["duration"] = duration
		}
		if pressDuration := action.ActionOptions.PressDuration; pressDuration > 0 {
			arguments["pressDuration"] = pressDuration
		}
		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid swipe params: %v", action.Params)
}

// ToolSwipeCoordinate implements the swipe_advanced tool call.
type ToolSwipeCoordinate struct{}

func (t *ToolSwipeCoordinate) Name() option.ActionMethod {
	return option.ACTION_SwipeCoordinate
}

func (t *ToolSwipeCoordinate) Description() string {
	return "Perform advanced swipe with custom coordinates and timing"
}

func (t *ToolSwipeCoordinate) Options() []mcp.ToolOption {
	return option.NewMCPOptions(option.SwipeAdvancedRequest{})
}

func (t *ToolSwipeCoordinate) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		var swipeAdvReq option.SwipeAdvancedRequest
		if err := mapToStruct(request.Params.Arguments, &swipeAdvReq); err != nil {
			return nil, fmt.Errorf("parse parameters error: %w", err)
		}

		// Advanced swipe action logic using prepareSwipeAction like the original DoAction
		log.Info().
			Float64("fromX", swipeAdvReq.FromX).Float64("fromY", swipeAdvReq.FromY).
			Float64("toX", swipeAdvReq.ToX).Float64("toY", swipeAdvReq.ToY).
			Msg("performing advanced swipe")

		params := []float64{swipeAdvReq.FromX, swipeAdvReq.FromY, swipeAdvReq.ToX, swipeAdvReq.ToY}
		opts := []option.ActionOption{}
		if swipeAdvReq.Duration > 0 {
			opts = append(opts, option.WithDuration(swipeAdvReq.Duration))
		}
		if swipeAdvReq.PressDuration > 0 {
			opts = append(opts, option.WithPressDuration(swipeAdvReq.PressDuration))
		}

		swipeAction := prepareSwipeAction(driverExt, params, opts...)
		err = swipeAction(driverExt)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Advanced swipe failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully performed advanced swipe from (%.2f, %.2f) to (%.2f, %.2f)",
			swipeAdvReq.FromX, swipeAdvReq.FromY, swipeAdvReq.ToX, swipeAdvReq.ToY)), nil
	}
}

func (t *ToolSwipeCoordinate) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	if paramSlice, err := builtin.ConvertToFloat64Slice(action.Params); err == nil && len(paramSlice) == 4 {
		arguments := map[string]any{
			"fromX": paramSlice[0],
			"fromY": paramSlice[1],
			"toX":   paramSlice[2],
			"toY":   paramSlice[3],
		}
		// Add duration and press duration from options
		if duration := action.ActionOptions.Duration; duration > 0 {
			arguments["duration"] = duration
		}
		if pressDuration := action.ActionOptions.PressDuration; pressDuration > 0 {
			arguments["pressDuration"] = pressDuration
		}
		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid swipe advanced params: %v", action.Params)
}

// ToolSwipeToTapApp implements the swipe_to_tap_app tool call.
type ToolSwipeToTapApp struct{}

func (t *ToolSwipeToTapApp) Name() option.ActionMethod {
	return option.ACTION_SwipeToTapApp
}

func (t *ToolSwipeToTapApp) Description() string {
	return "Swipe to find and tap an app by name"
}

func (t *ToolSwipeToTapApp) Options() []mcp.ToolOption {
	return option.NewMCPOptions(option.SwipeToTapAppRequest{})
}

func (t *ToolSwipeToTapApp) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		var swipeAppReq option.SwipeToTapAppRequest
		if err := mapToStruct(request.Params.Arguments, &swipeAppReq); err != nil {
			return nil, fmt.Errorf("parse parameters error: %w", err)
		}

		// Swipe to tap app action logic
		log.Info().Str("appName", swipeAppReq.AppName).Msg("swipe to tap app")
		err = driverExt.SwipeToTapApp(swipeAppReq.AppName)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Swipe to tap app failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully found and tapped app: %s", swipeAppReq.AppName)), nil
	}
}

func (t *ToolSwipeToTapApp) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	if appName, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"appName": appName,
		}
		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid swipe to tap app params: %v", action.Params)
}

// ToolSwipeToTapText implements the swipe_to_tap_text tool call.
type ToolSwipeToTapText struct{}

func (t *ToolSwipeToTapText) Name() option.ActionMethod {
	return option.ACTION_SwipeToTapText
}

func (t *ToolSwipeToTapText) Description() string {
	return "Swipe to find and tap text on screen"
}

func (t *ToolSwipeToTapText) Options() []mcp.ToolOption {
	return option.NewMCPOptions(option.SwipeToTapTextRequest{})
}

func (t *ToolSwipeToTapText) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		var swipeTextReq option.SwipeToTapTextRequest
		if err := mapToStruct(request.Params.Arguments, &swipeTextReq); err != nil {
			return nil, fmt.Errorf("parse parameters error: %w", err)
		}

		// Swipe to tap text action logic
		log.Info().Str("text", swipeTextReq.Text).Msg("swipe to tap text")
		err = driverExt.SwipeToTapTexts([]string{swipeTextReq.Text})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Swipe to tap text failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully found and tapped text: %s", swipeTextReq.Text)), nil
	}
}

func (t *ToolSwipeToTapText) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	if text, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"text": text,
		}
		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid swipe to tap text params: %v", action.Params)
}

// ToolSwipeToTapTexts implements the swipe_to_tap_texts tool call.
type ToolSwipeToTapTexts struct{}

func (t *ToolSwipeToTapTexts) Name() option.ActionMethod {
	return option.ACTION_SwipeToTapTexts
}

func (t *ToolSwipeToTapTexts) Description() string {
	return "Swipe to find and tap one of multiple texts on screen"
}

func (t *ToolSwipeToTapTexts) Options() []mcp.ToolOption {
	return option.NewMCPOptions(option.SwipeToTapTextsRequest{})
}

func (t *ToolSwipeToTapTexts) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		var swipeTextsReq option.SwipeToTapTextsRequest
		if err := mapToStruct(request.Params.Arguments, &swipeTextsReq); err != nil {
			return nil, fmt.Errorf("parse parameters error: %w", err)
		}

		// Swipe to tap texts action logic
		log.Info().Strs("texts", swipeTextsReq.Texts).Msg("swipe to tap texts")
		err = driverExt.SwipeToTapTexts(swipeTextsReq.Texts)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Swipe to tap texts failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully found and tapped one of texts: %v", swipeTextsReq.Texts)), nil
	}
}

func (t *ToolSwipeToTapTexts) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	var texts []string
	if textsSlice, ok := action.Params.([]string); ok {
		texts = textsSlice
	} else if textsInterface, err := builtin.ConvertToStringSlice(action.Params); err == nil {
		texts = textsInterface
	} else {
		return mcp.CallToolRequest{}, fmt.Errorf("invalid swipe to tap texts params: %v", action.Params)
	}
	arguments := map[string]any{
		"texts": texts,
	}
	return buildMCPCallToolRequest(t.Name(), arguments), nil
}

// ToolDrag implements the drag tool call.
type ToolDrag struct{}

func (t *ToolDrag) Name() option.ActionMethod {
	return option.ACTION_Drag
}

func (t *ToolDrag) Description() string {
	return "Drag on the mobile device"
}

func (t *ToolDrag) Options() []mcp.ToolOption {
	return option.NewMCPOptions(option.DragRequest{})
}

func (t *ToolDrag) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		var dragReq option.DragRequest
		if err := mapToStruct(request.Params.Arguments, &dragReq); err != nil {
			return nil, fmt.Errorf("parse parameters error: %w", err)
		}

		opts := []option.ActionOption{}
		if dragReq.Duration > 0 {
			opts = append(opts, option.WithDuration(dragReq.Duration/1000.0))
		}

		// Drag action logic
		log.Info().
			Float64("fromX", dragReq.FromX).Float64("fromY", dragReq.FromY).
			Float64("toX", dragReq.ToX).Float64("toY", dragReq.ToY).
			Msg("performing drag")

		err = driverExt.Swipe(dragReq.FromX, dragReq.FromY, dragReq.ToX, dragReq.ToY, opts...)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Drag failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully dragged from (%.2f, %.2f) to (%.2f, %.2f)",
			dragReq.FromX, dragReq.FromY, dragReq.ToX, dragReq.ToY)), nil
	}
}

func (t *ToolDrag) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	if paramSlice, err := builtin.ConvertToFloat64Slice(action.Params); err == nil && len(paramSlice) == 4 {
		arguments := map[string]any{
			"fromX": paramSlice[0],
			"fromY": paramSlice[1],
			"toX":   paramSlice[2],
			"toY":   paramSlice[3],
		}
		// Add duration from options
		if duration := action.ActionOptions.Duration; duration > 0 {
			arguments["duration"] = duration * 1000 // convert to milliseconds
		}
		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid drag params: %v", action.Params)
}

func NewDriverExt(platform, serial string) (*XTDriver, error) {
	device, err := NewDevice(platform, serial)
	if err != nil {
		return nil, err
	}

	// init driver
	driver, err := device.NewDriver()
	if err != nil {
		return nil, fmt.Errorf("init driver failed: %w", err)
	}
	if err := driver.Setup(); err != nil {
		return nil, fmt.Errorf("setup driver failed: %w", err)
	}

	// init XTDriver
	driverExt, err := NewXTDriver(driver,
		option.WithCVService(option.CVServiceTypeVEDEM))
	if err != nil {
		return nil, fmt.Errorf("init XT driver failed: %w", err)
	}
	return driverExt, nil
}

func NewDevice(platform, serial string) (device IDevice, err error) {
	if serial == "" {
		return nil, fmt.Errorf("serial is empty")
	}
	switch strings.ToLower(platform) {
	case "android":
		device, err = NewAndroidDevice(
			option.WithSerialNumber(serial))
		if err != nil {
			return
		}
	case "ios":
		device, err = NewIOSDevice(
			option.WithUDID(serial),
			option.WithWDAPort(8700),
			option.WithWDAMjpegPort(8800),
			option.WithResetHomeOnStartup(false),
		)
		if err != nil {
			return
		}
	case "browser":
		device, err = NewBrowserDevice(option.WithBrowserID(serial))
		if err != nil {
			return
		}
	default:
		return nil, fmt.Errorf("invalid platform: %s", platform)
	}
	err = device.Setup()
	if err != nil {
		log.Error().Err(err).Msg("setup device failed")
	}
	return device, nil
}

// mapToStruct convert map[string]any to target struct
func mapToStruct(m map[string]any, out interface{}) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}

// ToolHome implements the home tool call.
type ToolHome struct{}

func (t *ToolHome) Name() option.ActionMethod {
	return option.ACTION_Home
}

func (t *ToolHome) Description() string {
	return "Press the home button on the device"
}

func (t *ToolHome) Options() []mcp.ToolOption {
	return option.NewMCPOptions(option.TargetDeviceRequest{})
}

func (t *ToolHome) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		// Home action logic
		log.Info().Msg("pressing home button")
		err = driverExt.Home()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Home button press failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText("Successfully pressed home button"), nil
	}
}

func (t *ToolHome) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	return buildMCPCallToolRequest(t.Name(), map[string]any{}), nil
}

// ToolBack implements the back tool call.
type ToolBack struct{}

func (t *ToolBack) Name() option.ActionMethod {
	return option.ACTION_Back
}

func (t *ToolBack) Description() string {
	return "Press the back button on the device"
}

func (t *ToolBack) Options() []mcp.ToolOption {
	return option.NewMCPOptions(option.TargetDeviceRequest{})
}

func (t *ToolBack) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		// Back action logic
		log.Info().Msg("pressing back button")
		err = driverExt.Back()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Back button press failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText("Successfully pressed back button"), nil
	}
}

func (t *ToolBack) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	return buildMCPCallToolRequest(t.Name(), map[string]any{}), nil
}

// ToolInput implements the input tool call.
type ToolInput struct{}

func (t *ToolInput) Name() option.ActionMethod {
	return option.ACTION_Input
}

func (t *ToolInput) Description() string {
	return "Input text on the current active element"
}

func (t *ToolInput) Options() []mcp.ToolOption {
	return option.NewMCPOptions(option.InputRequest{})
}

func (t *ToolInput) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		var inputReq option.InputRequest
		if err := mapToStruct(request.Params.Arguments, &inputReq); err != nil {
			return nil, fmt.Errorf("parse parameters error: %w", err)
		}

		if inputReq.Text == "" {
			return nil, fmt.Errorf("text is required")
		}

		// Input action logic
		log.Info().Str("text", inputReq.Text).Msg("inputting text")
		err = driverExt.Input(inputReq.Text)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Input failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully input text: %s", inputReq.Text)), nil
	}
}

func (t *ToolInput) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	text := fmt.Sprintf("%v", action.Params)
	arguments := map[string]any{
		"text": text,
	}
	return buildMCPCallToolRequest(t.Name(), arguments), nil
}

// ToolWebLoginNoneUI implements the web_login_none_ui tool call.
type ToolWebLoginNoneUI struct{}

func (t *ToolWebLoginNoneUI) Name() option.ActionMethod {
	return option.ACTION_WebLoginNoneUI
}

func (t *ToolWebLoginNoneUI) Description() string {
	return "Perform login without UI interaction for web applications"
}

func (t *ToolWebLoginNoneUI) Options() []mcp.ToolOption {
	return option.NewMCPOptions(option.WebLoginNoneUIRequest{})
}

func (t *ToolWebLoginNoneUI) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		var loginReq option.WebLoginNoneUIRequest
		if err := mapToStruct(request.Params.Arguments, &loginReq); err != nil {
			return nil, fmt.Errorf("parse parameters error: %w", err)
		}

		// Web login none UI action logic
		log.Info().Str("packageName", loginReq.PackageName).Msg("performing web login without UI")
		driver, ok := driverExt.IDriver.(*BrowserDriver)
		if !ok {
			return nil, fmt.Errorf("invalid browser driver for web login")
		}

		_, err = driver.LoginNoneUI(loginReq.PackageName, loginReq.PhoneNumber, loginReq.Captcha, loginReq.Password)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Web login failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText("Successfully performed web login without UI"), nil
	}
}

func (t *ToolWebLoginNoneUI) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	return buildMCPCallToolRequest(t.Name(), map[string]any{}), nil
}

// ToolAppInstall implements the app_install tool call.
type ToolAppInstall struct{}

func (t *ToolAppInstall) Name() option.ActionMethod {
	return option.ACTION_AppInstall
}

func (t *ToolAppInstall) Description() string {
	return "Install an app on the device"
}

func (t *ToolAppInstall) Options() []mcp.ToolOption {
	return option.NewMCPOptions(option.AppInstallRequest{})
}

func (t *ToolAppInstall) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		var installReq option.AppInstallRequest
		if err := mapToStruct(request.Params.Arguments, &installReq); err != nil {
			return nil, fmt.Errorf("parse parameters error: %w", err)
		}

		// App install action logic
		log.Info().Str("appUrl", installReq.AppUrl).Msg("installing app")
		err = driverExt.GetDevice().Install(installReq.AppUrl)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("App install failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully installed app from: %s", installReq.AppUrl)), nil
	}
}

func (t *ToolAppInstall) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	if appUrl, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"appUrl": appUrl,
		}
		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid app install params: %v", action.Params)
}

// ToolAppUninstall implements the app_uninstall tool call.
type ToolAppUninstall struct{}

func (t *ToolAppUninstall) Name() option.ActionMethod {
	return option.ACTION_AppUninstall
}

func (t *ToolAppUninstall) Description() string {
	return "Uninstall an app from the device"
}

func (t *ToolAppUninstall) Options() []mcp.ToolOption {
	return option.NewMCPOptions(option.AppUninstallRequest{})
}

func (t *ToolAppUninstall) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		var uninstallReq option.AppUninstallRequest
		if err := mapToStruct(request.Params.Arguments, &uninstallReq); err != nil {
			return nil, fmt.Errorf("parse parameters error: %w", err)
		}

		// App uninstall action logic
		log.Info().Str("packageName", uninstallReq.PackageName).Msg("uninstalling app")
		err = driverExt.GetDevice().Uninstall(uninstallReq.PackageName)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("App uninstall failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully uninstalled app: %s", uninstallReq.PackageName)), nil
	}
}

func (t *ToolAppUninstall) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	if packageName, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"packageName": packageName,
		}
		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid app uninstall params: %v", action.Params)
}

// ToolAppClear implements the app_clear tool call.
type ToolAppClear struct{}

func (t *ToolAppClear) Name() option.ActionMethod {
	return option.ACTION_AppClear
}

func (t *ToolAppClear) Description() string {
	return "Clear app data and cache"
}

func (t *ToolAppClear) Options() []mcp.ToolOption {
	return option.NewMCPOptions(option.AppClearRequest{})
}

func (t *ToolAppClear) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		var clearReq option.AppClearRequest
		if err := mapToStruct(request.Params.Arguments, &clearReq); err != nil {
			return nil, fmt.Errorf("parse parameters error: %w", err)
		}

		// App clear action logic
		log.Info().Str("packageName", clearReq.PackageName).Msg("clearing app")
		err = driverExt.AppClear(clearReq.PackageName)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("App clear failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully cleared app: %s", clearReq.PackageName)), nil
	}
}

func (t *ToolAppClear) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	if packageName, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"packageName": packageName,
		}
		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid app clear params: %v", action.Params)
}

// ToolSecondaryClick implements the secondary_click tool call.
type ToolSecondaryClick struct{}

func (t *ToolSecondaryClick) Name() option.ActionMethod {
	return option.ACTION_SecondaryClick
}

func (t *ToolSecondaryClick) Description() string {
	return "Perform secondary click (right click) at coordinates"
}

func (t *ToolSecondaryClick) Options() []mcp.ToolOption {
	return option.NewMCPOptions(option.SecondaryClickRequest{})
}

func (t *ToolSecondaryClick) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		var clickReq option.SecondaryClickRequest
		if err := mapToStruct(request.Params.Arguments, &clickReq); err != nil {
			return nil, fmt.Errorf("parse parameters error: %w", err)
		}

		// Secondary click action logic
		log.Info().Float64("x", clickReq.X).Float64("y", clickReq.Y).Msg("performing secondary click")
		err = driverExt.SecondaryClick(clickReq.X, clickReq.Y)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Secondary click failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully performed secondary click at (%.2f, %.2f)", clickReq.X, clickReq.Y)), nil
	}
}

func (t *ToolSecondaryClick) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	if params, err := builtin.ConvertToFloat64Slice(action.Params); err == nil && len(params) == 2 {
		arguments := map[string]any{
			"x": params[0],
			"y": params[1],
		}
		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid secondary click params: %v", action.Params)
}

// ToolHoverBySelector implements the hover_by_selector tool call.
type ToolHoverBySelector struct{}

func (t *ToolHoverBySelector) Name() option.ActionMethod {
	return option.ACTION_HoverBySelector
}

func (t *ToolHoverBySelector) Description() string {
	return "Hover over an element selected by CSS selector or XPath"
}

func (t *ToolHoverBySelector) Options() []mcp.ToolOption {
	return option.NewMCPOptions(option.SelectorRequest{})
}

func (t *ToolHoverBySelector) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		var selectorReq option.SelectorRequest
		if err := mapToStruct(request.Params.Arguments, &selectorReq); err != nil {
			return nil, fmt.Errorf("parse parameters error: %w", err)
		}

		// Hover by selector action logic
		log.Info().Str("selector", selectorReq.Selector).Msg("hovering by selector")
		err = driverExt.HoverBySelector(selectorReq.Selector)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Hover by selector failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully hovered over element with selector: %s", selectorReq.Selector)), nil
	}
}

func (t *ToolHoverBySelector) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	if selector, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"selector": selector,
		}
		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid hover by selector params: %v", action.Params)
}

// ToolTapBySelector implements the tap_by_selector tool call.
type ToolTapBySelector struct{}

func (t *ToolTapBySelector) Name() option.ActionMethod {
	return option.ACTION_TapBySelector
}

func (t *ToolTapBySelector) Description() string {
	return "Tap an element selected by CSS selector or XPath"
}

func (t *ToolTapBySelector) Options() []mcp.ToolOption {
	return option.NewMCPOptions(option.SelectorRequest{})
}

func (t *ToolTapBySelector) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		var selectorReq option.SelectorRequest
		if err := mapToStruct(request.Params.Arguments, &selectorReq); err != nil {
			return nil, fmt.Errorf("parse parameters error: %w", err)
		}

		// Tap by selector action logic
		log.Info().Str("selector", selectorReq.Selector).Msg("tapping by selector")
		err = driverExt.TapBySelector(selectorReq.Selector)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Tap by selector failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully tapped element with selector: %s", selectorReq.Selector)), nil
	}
}

func (t *ToolTapBySelector) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	if selector, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"selector": selector,
		}
		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid tap by selector params: %v", action.Params)
}

// ToolSecondaryClickBySelector implements the secondary_click_by_selector tool call.
type ToolSecondaryClickBySelector struct{}

func (t *ToolSecondaryClickBySelector) Name() option.ActionMethod {
	return option.ACTION_SecondaryClickBySelector
}

func (t *ToolSecondaryClickBySelector) Description() string {
	return "Perform secondary click on an element selected by CSS selector or XPath"
}

func (t *ToolSecondaryClickBySelector) Options() []mcp.ToolOption {
	return option.NewMCPOptions(option.SelectorRequest{})
}

func (t *ToolSecondaryClickBySelector) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		var selectorReq option.SelectorRequest
		if err := mapToStruct(request.Params.Arguments, &selectorReq); err != nil {
			return nil, fmt.Errorf("parse parameters error: %w", err)
		}

		// Secondary click by selector action logic
		log.Info().Str("selector", selectorReq.Selector).Msg("performing secondary click by selector")
		err = driverExt.SecondaryClickBySelector(selectorReq.Selector)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Secondary click by selector failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully performed secondary click on element with selector: %s", selectorReq.Selector)), nil
	}
}

func (t *ToolSecondaryClickBySelector) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	if selector, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"selector": selector,
		}
		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid secondary click by selector params: %v", action.Params)
}

// ToolWebCloseTab implements the web_close_tab tool call.
type ToolWebCloseTab struct{}

func (t *ToolWebCloseTab) Name() option.ActionMethod {
	return option.ACTION_WebCloseTab
}

func (t *ToolWebCloseTab) Description() string {
	return "Close a browser tab by index"
}

func (t *ToolWebCloseTab) Options() []mcp.ToolOption {
	return option.NewMCPOptions(option.WebCloseTabRequest{})
}

func (t *ToolWebCloseTab) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		var closeTabReq option.WebCloseTabRequest
		if err := mapToStruct(request.Params.Arguments, &closeTabReq); err != nil {
			return nil, fmt.Errorf("parse parameters error: %w", err)
		}

		// Web close tab action logic
		log.Info().Int("tabIndex", closeTabReq.TabIndex).Msg("closing web tab")
		browserDriver, ok := driverExt.IDriver.(*BrowserDriver)
		if !ok {
			return nil, fmt.Errorf("web close tab is only supported for browser drivers")
		}

		err = browserDriver.CloseTab(closeTabReq.TabIndex)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Close tab failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully closed tab at index: %d", closeTabReq.TabIndex)), nil
	}
}

func (t *ToolWebCloseTab) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	var tabIndex int
	if param, ok := action.Params.(json.Number); ok {
		paramInt64, _ := param.Int64()
		tabIndex = int(paramInt64)
	} else if param, ok := action.Params.(int64); ok {
		tabIndex = int(param)
	} else if param, ok := action.Params.(int); ok {
		tabIndex = param
	} else {
		return mcp.CallToolRequest{}, fmt.Errorf("invalid web close tab params: %v", action.Params)
	}
	arguments := map[string]any{
		"tabIndex": tabIndex,
	}
	return buildMCPCallToolRequest(t.Name(), arguments), nil
}

// ToolSetIme implements the set_ime tool call.
type ToolSetIme struct{}

func (t *ToolSetIme) Name() option.ActionMethod {
	return option.ACTION_SetIme
}

func (t *ToolSetIme) Description() string {
	return "Set the input method editor (IME) on the device"
}

func (t *ToolSetIme) Options() []mcp.ToolOption {
	return option.NewMCPOptions(option.SetImeRequest{})
}

func (t *ToolSetIme) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		var imeReq option.SetImeRequest
		if err := mapToStruct(request.Params.Arguments, &imeReq); err != nil {
			return nil, fmt.Errorf("parse parameters error: %w", err)
		}

		// Set IME action logic
		log.Info().Str("ime", imeReq.Ime).Msg("setting IME")
		err = driverExt.SetIme(imeReq.Ime)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Set IME failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully set IME to: %s", imeReq.Ime)), nil
	}
}

func (t *ToolSetIme) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	if ime, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"ime": ime,
		}
		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid set ime params: %v", action.Params)
}

// ToolGetSource implements the get_source tool call.
type ToolGetSource struct{}

func (t *ToolGetSource) Name() option.ActionMethod {
	return option.ACTION_GetSource
}

func (t *ToolGetSource) Description() string {
	return "Get the source/hierarchy of the current screen"
}

func (t *ToolGetSource) Options() []mcp.ToolOption {
	return option.NewMCPOptions(option.GetSourceRequest{})
}

func (t *ToolGetSource) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		var sourceReq option.GetSourceRequest
		if err := mapToStruct(request.Params.Arguments, &sourceReq); err != nil {
			return nil, fmt.Errorf("parse parameters error: %w", err)
		}

		// Get source action logic
		log.Info().Str("packageName", sourceReq.PackageName).Msg("getting source")
		_, err = driverExt.Source(option.WithProcessName(sourceReq.PackageName))
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Get source failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully retrieved source for package: %s", sourceReq.PackageName)), nil
	}
}

func (t *ToolGetSource) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	if packageName, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"packageName": packageName,
		}
		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid get source params: %v", action.Params)
}

// ToolSleep implements the sleep tool call.
type ToolSleep struct{}

func (t *ToolSleep) Name() option.ActionMethod {
	return option.ACTION_Sleep
}

func (t *ToolSleep) Description() string {
	return "Sleep for a specified number of seconds"
}

func (t *ToolSleep) Options() []mcp.ToolOption {
	return []mcp.ToolOption{
		mcp.WithNumber("seconds", mcp.Description("Number of seconds to sleep")),
	}
}

func (t *ToolSleep) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		seconds, ok := request.Params.Arguments["seconds"]
		if !ok {
			return nil, fmt.Errorf("seconds parameter is required")
		}

		// Sleep action logic
		log.Info().Interface("seconds", seconds).Msg("sleeping")

		var duration time.Duration
		switch v := seconds.(type) {
		case float64:
			duration = time.Duration(v*1000) * time.Millisecond
		case int:
			duration = time.Duration(v) * time.Second
		case int64:
			duration = time.Duration(v) * time.Second
		case string:
			s, err := builtin.ConvertToFloat64(v)
			if err != nil {
				return nil, fmt.Errorf("invalid sleep duration: %v", v)
			}
			duration = time.Duration(s*1000) * time.Millisecond
		default:
			return nil, fmt.Errorf("unsupported sleep duration type: %T", v)
		}

		time.Sleep(duration)

		return mcp.NewToolResultText(fmt.Sprintf("Successfully slept for %v seconds", seconds)), nil
	}
}

func (t *ToolSleep) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	arguments := map[string]any{
		"seconds": action.Params,
	}
	return buildMCPCallToolRequest(t.Name(), arguments), nil
}

// ToolSleepMS implements the sleep_ms tool call.
type ToolSleepMS struct{}

func (t *ToolSleepMS) Name() option.ActionMethod {
	return option.ACTION_SleepMS
}

func (t *ToolSleepMS) Description() string {
	return "Sleep for specified milliseconds"
}

func (t *ToolSleepMS) Options() []mcp.ToolOption {
	return option.NewMCPOptions(option.SleepMSRequest{})
}

func (t *ToolSleepMS) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var sleepReq option.SleepMSRequest
		if err := mapToStruct(request.Params.Arguments, &sleepReq); err != nil {
			return nil, fmt.Errorf("parse parameters error: %w", err)
		}

		// Sleep MS action logic
		log.Info().Int64("milliseconds", sleepReq.Milliseconds).Msg("sleeping in milliseconds")
		time.Sleep(time.Duration(sleepReq.Milliseconds) * time.Millisecond)

		return mcp.NewToolResultText(fmt.Sprintf("Successfully slept for %d milliseconds", sleepReq.Milliseconds)), nil
	}
}

func (t *ToolSleepMS) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	var milliseconds int64
	if param, ok := action.Params.(json.Number); ok {
		milliseconds, _ = param.Int64()
	} else if param, ok := action.Params.(int64); ok {
		milliseconds = param
	} else {
		return mcp.CallToolRequest{}, fmt.Errorf("invalid sleep ms params: %v", action.Params)
	}
	arguments := map[string]any{
		"milliseconds": milliseconds,
	}
	return buildMCPCallToolRequest(t.Name(), arguments), nil
}

// ToolSleepRandom implements the sleep_random tool call.
type ToolSleepRandom struct{}

func (t *ToolSleepRandom) Name() option.ActionMethod {
	return option.ACTION_SleepRandom
}

func (t *ToolSleepRandom) Description() string {
	return "Sleep for a random duration based on parameters"
}

func (t *ToolSleepRandom) Options() []mcp.ToolOption {
	return option.NewMCPOptions(option.SleepRandomRequest{})
}

func (t *ToolSleepRandom) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var sleepRandomReq option.SleepRandomRequest
		if err := mapToStruct(request.Params.Arguments, &sleepRandomReq); err != nil {
			return nil, fmt.Errorf("parse parameters error: %w", err)
		}

		// Sleep random action logic
		log.Info().Floats64("params", sleepRandomReq.Params).Msg("sleeping for random duration")
		sleepStrict(time.Now(), getSimulationDuration(sleepRandomReq.Params))

		return mcp.NewToolResultText(fmt.Sprintf("Successfully slept for random duration with params: %v", sleepRandomReq.Params)), nil
	}
}

func (t *ToolSleepRandom) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	if params, err := builtin.ConvertToFloat64Slice(action.Params); err == nil {
		arguments := map[string]any{
			"params": params,
		}
		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid sleep random params: %v", action.Params)
}

// ToolClosePopups implements the close_popups tool call.
type ToolClosePopups struct{}

func (t *ToolClosePopups) Name() option.ActionMethod {
	return option.ACTION_ClosePopups
}

func (t *ToolClosePopups) Description() string {
	return "Close any popup windows or dialogs on screen"
}

func (t *ToolClosePopups) Options() []mcp.ToolOption {
	return option.NewMCPOptions(option.TargetDeviceRequest{})
}

func (t *ToolClosePopups) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		// Close popups action logic
		log.Info().Msg("closing popups")
		err = driverExt.ClosePopupsHandler()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Close popups failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText("Successfully closed popups"), nil
	}
}

func (t *ToolClosePopups) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	return buildMCPCallToolRequest(t.Name(), map[string]any{}), nil
}

// ToolAIAction implements the ai_action tool call.
type ToolAIAction struct{}

func (t *ToolAIAction) Name() option.ActionMethod {
	return option.ACTION_AIAction
}

func (t *ToolAIAction) Description() string {
	return "Perform actions using AI with a given prompt"
}

func (t *ToolAIAction) Options() []mcp.ToolOption {
	return option.NewMCPOptions(option.AIActionRequest{})
}

func (t *ToolAIAction) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		var aiReq option.AIActionRequest
		if err := mapToStruct(request.Params.Arguments, &aiReq); err != nil {
			return nil, fmt.Errorf("parse parameters error: %w", err)
		}

		// AI action logic
		log.Info().Str("prompt", aiReq.Prompt).Msg("performing AI action")
		err = driverExt.AIAction(aiReq.Prompt)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("AI action failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully performed AI action with prompt: %s", aiReq.Prompt)), nil
	}
}

func (t *ToolAIAction) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	if prompt, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"prompt": prompt,
		}
		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid AI action params: %v", action.Params)
}
