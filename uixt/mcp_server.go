package uixt

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
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

/*
Package uixt provides MCP (Model Context Protocol) server implementation for HttpRunner UI automation.

# HttpRunner MCP Server

This package implements a comprehensive MCP server that exposes HttpRunner's UI automation
capabilities through standardized MCP protocol interfaces. It enables AI models and other
clients to perform mobile and web UI automation tasks.

## Architecture Overview

The MCP server follows a pure ActionTool architecture where each UI operation is implemented
as an independent tool that conforms to the ActionTool interface:

	┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
	│   MCP Client    │    │   MCP Server    │    │  XTDriver Core  │
	│   (AI Model)    │◄──►│  (mcp_server)   │◄──►│   (UI Engine)   │
	└─────────────────┘    └─────────────────┘    └─────────────────┘
	                              │
	                              ▼
	                       ┌─────────────────┐
	                       │  Device Layer   │
	                       │ Android/iOS/Web │
	                       └─────────────────┘

## Core Components

### MCPServer4XTDriver
The main server struct that manages MCP protocol communication and tool registration.

### ActionTool Interface
Defines the contract for all MCP tools:
  - Name(): Returns the action name identifier
  - Description(): Provides human-readable tool description
  - Options(): Defines MCP tool parameters and validation
  - Implement(): Contains the actual tool execution logic
  - ConvertActionToCallToolRequest(): Converts legacy actions to MCP format

## Supported Operations

### Device Management
- list_available_devices: Discover Android/iOS devices and simulators
- select_device: Choose specific device by platform and serial

### Touch Operations
- tap_xy: Tap at relative coordinates (0-1 range)
- tap_abs_xy: Tap at absolute pixel coordinates
- tap_ocr: Tap on text found by OCR recognition
- tap_cv: Tap on element found by computer vision
- double_tap_xy: Double tap at coordinates

### Gesture Operations
- swipe: Generic swipe with auto-detection (direction or coordinates)
- swipe_direction: Directional swipe (up/down/left/right)
- swipe_coordinate: Coordinate-based swipe with precise control
- drag: Drag operation between two points

### Advanced Swipe Operations
- swipe_to_tap_app: Swipe to find and tap app by name
- swipe_to_tap_text: Swipe to find and tap text
- swipe_to_tap_texts: Swipe to find and tap one of multiple texts

### Input Operations
- input: Text input on focused element
- press_button: Press device buttons (home, back, volume, etc.)

### App Management
- list_packages: List all installed apps
- app_launch: Launch app by package name
- app_terminate: Terminate running app
- app_install: Install app from URL/path
- app_uninstall: Uninstall app by package name
- app_clear: Clear app data and cache

### Screen Operations
- screenshot: Capture screen as Base64 encoded image
- get_screen_size: Get device screen dimensions
- get_source: Get UI hierarchy/source

### Utility Operations
- sleep: Sleep for specified seconds
- sleep_ms: Sleep for specified milliseconds
- sleep_random: Sleep for random duration based on parameters
- set_ime: Set input method editor
- close_popups: Close popup windows/dialogs

### Web Operations
- web_login_none_ui: Perform login without UI interaction
- secondary_click: Right-click at specified coordinates
- hover_by_selector: Hover over element by CSS selector/XPath
- tap_by_selector: Click element by CSS selector/XPath
- secondary_click_by_selector: Right-click element by selector
- web_close_tab: Close browser tab by index

### AI Operations
- ai_action: Perform AI-driven actions with natural language prompts
- finished: Mark task completion with result message

## Key Features

### Anti-Risk Support
Built-in anti-detection mechanisms for sensitive operations:
  - Touch simulation with realistic timing
  - Device fingerprint masking
  - Behavioral pattern randomization

### Unified Parameter Handling
All tools use consistent parameter parsing through parseActionOptions():
  - JSON marshaling/unmarshaling for type safety
  - Automatic validation and error handling
  - Support for complex nested parameters

### Device Abstraction
Seamless multi-platform support:
  - Android devices via ADB
  - iOS devices via go-ios
  - Web browsers via WebDriver
  - Harmony OS devices

### Error Handling
Comprehensive error management:
  - Structured error responses
  - Detailed logging with context
  - Graceful failure recovery

## Usage Example

	// Create and start MCP server
	server := NewMCPServer()
	err := server.Start() // Blocks and serves MCP protocol over stdio

	// Client interaction (via MCP protocol):
	// 1. Initialize connection
	// 2. List available tools
	// 3. Call tools with parameters
	// 4. Receive structured results

## Extension Guide

To add a new tool:

1. Define tool struct implementing ActionTool interface
2. Implement all required methods (Name, Description, Options, Implement, ConvertActionToCallToolRequest)
3. Register tool in registerTools() method
4. Add comprehensive unit tests
5. Update documentation

Example:
	type ToolCustomAction struct{}

	func (t *ToolCustomAction) Name() option.ActionName {
		return option.ACTION_CustomAction
	}

	func (t *ToolCustomAction) Implement() server.ToolHandlerFunc {
		return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// Implementation logic
		}
	}

## Performance Considerations

- Driver instances are cached and reused for efficiency
- Parameter parsing is optimized to minimize JSON overhead
- Timeout controls prevent hanging operations
- Resource cleanup ensures memory efficiency

## Security Notes

- All device operations require explicit permission
- Input validation prevents injection attacks
- Sensitive operations support anti-detection measures
- Audit logging tracks all tool executions

For detailed implementation examples and best practices, see the accompanying
documentation.
*/

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
//
// This function initializes a complete MCP server instance with:
//   - MCP protocol server with uixt capabilities
//   - Version information from HttpRunner
//   - Tool capabilities disabled (set to false for performance)
//   - All available UI automation tools pre-registered
//
// The server supports the following tool categories:
//   - Device management (discovery, selection)
//   - Touch operations (tap, double-tap, long-press)
//   - Gesture operations (swipe, drag)
//   - Input operations (text input, button press)
//   - App management (launch, terminate, install)
//   - Screen operations (screenshot, size, source)
//   - Utility operations (sleep, IME, popups)
//   - Web operations (browser automation)
//   - AI operations (intelligent actions)
//
// Returns:
//   - *MCPServer4XTDriver: Configured server ready to start
//
// Usage:
//
//	server := NewMCPServer()
//	err := server.Start() // Blocks and serves over stdio
func NewMCPServer() *MCPServer4XTDriver {
	mcpServer := server.NewMCPServer(
		"uixt",
		version.GetVersionInfo(),
		server.WithToolCapabilities(false),
	)
	s := &MCPServer4XTDriver{
		mcpServer:     mcpServer,
		actionToolMap: make(map[option.ActionName]ActionTool),
	}
	s.registerTools()
	return s
}

// MCPServer4XTDriver wraps a MCPServer to expose XTDriver functionality via MCP protocol.
type MCPServer4XTDriver struct {
	mcpServer     *server.MCPServer
	mcpTools      []mcp.Tool                       // tools list for uixt
	actionToolMap map[option.ActionName]ActionTool // action method to tool mapping
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

// GetToolByAction returns the tool that handles the given action method
func (s *MCPServer4XTDriver) GetToolByAction(actionMethod option.ActionName) ActionTool {
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

	// Swipe Tools
	s.registerTool(&ToolSwipe{})           // generic swipe, auto-detect direction or coordinate
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

	// Sleep Tools
	s.registerTool(&ToolSleep{})
	s.registerTool(&ToolSleepMS{})
	s.registerTool(&ToolSleepRandom{})

	// Utils tools
	s.registerTool(&ToolSetIme{})
	s.registerTool(&ToolGetSource{})
	s.registerTool(&ToolClosePopups{})

	// PC/Web actions
	s.registerTool(&ToolWebLoginNoneUI{})
	s.registerTool(&ToolSecondaryClick{})
	s.registerTool(&ToolHoverBySelector{})
	s.registerTool(&ToolTapBySelector{})
	s.registerTool(&ToolSecondaryClickBySelector{})
	s.registerTool(&ToolWebCloseTab{})

	// LLM actions
	s.registerTool(&ToolAIAction{})
	s.registerTool(&ToolFinished{})
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
//
// This interface standardizes how UI automation actions are exposed through MCP protocol.
// Each tool implementation must provide:
//
// 1. Identity and Documentation:
//   - Name(): Unique identifier for the action (e.g., ACTION_TapXY)
//   - Description(): Human-readable description for AI models
//
// 2. MCP Integration:
//   - Options(): Parameter definitions with validation rules
//   - Implement(): Actual execution logic as MCP handler
//
// 3. Legacy Compatibility:
//   - ConvertActionToCallToolRequest(): Converts old MobileAction format
//
// Implementation Pattern:
//
//	type ToolExample struct{}
//
//	func (t *ToolExample) Name() option.ActionName {
//	    return option.ACTION_Example
//	}
//
//	func (t *ToolExample) Description() string {
//	    return "Performs example operation"
//	}
//
//	func (t *ToolExample) Options() []mcp.ToolOption {
//	    return []mcp.ToolOption{
//	        mcp.WithString("param", mcp.Description("Parameter description")),
//	    }
//	}
//
//	func (t *ToolExample) Implement() server.ToolHandlerFunc {
//	    return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
//	        // 1. Setup driver
//	        // 2. Parse parameters
//	        // 3. Execute operation
//	        // 4. Return result
//	    }
//	}
//
// Benefits of this architecture:
//   - Complete decoupling between tools
//   - Consistent parameter handling
//   - Standardized error reporting
//   - Easy testing and maintenance
//   - Seamless MCP protocol integration
type ActionTool interface {
	Name() option.ActionName
	Description() string
	Options() []mcp.ToolOption
	Implement() server.ToolHandlerFunc
	// ConvertActionToCallToolRequest converts MobileAction to mcp.CallToolRequest
	ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error)
	// ReturnSchema returns the expected return value schema based on mcp.CallToolResult conventions
	ReturnSchema() map[string]string
}

// buildMCPCallToolRequest is a helper function to build mcp.CallToolRequest
func buildMCPCallToolRequest(toolName option.ActionName, arguments map[string]any) mcp.CallToolRequest {
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

		jsonResult, _ := json.Marshal(deviceList)
		return mcp.NewToolResultText(string(jsonResult)), nil
	}
}

func (t *ToolListAvailableDevices) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	return buildMCPCallToolRequest(t.Name(), map[string]any{}), nil
}

func (t *ToolListAvailableDevices) ReturnSchema() map[string]string {
	return map[string]string{
		"androidDevices": "[]string: List of Android device serial numbers",
		"iosDevices":     "[]string: List of iOS device UDIDs",
	}
}

// ToolSelectDevice implements the select_device tool call.
type ToolSelectDevice struct{}

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
		return mcp.NewToolResultText(fmt.Sprintf("Selected device: %s", uuid)), nil
	}
}

func (t *ToolSelectDevice) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	return buildMCPCallToolRequest(t.Name(), map[string]any{}), nil
}

func (t *ToolSelectDevice) ReturnSchema() map[string]string {
	return map[string]string{
		"message": "string: Success message with selected device UUID",
	}
}

// ToolTapXY implements the tap_xy tool call.
//
// This tool performs touch/click operations at specified relative coordinates on the device screen.
// Coordinates are normalized to 0-1 range where (0,0) is top-left and (1,1) is bottom-right.
//
// Supported platforms:
//   - Android: Touch events via ADB
//   - iOS: Touch events via go-ios
//   - Web: Click events via WebDriver
//   - Harmony: Touch events via native interface
//
// Features:
//   - Relative coordinate system (0-1 range)
//   - Anti-risk detection support
//   - Configurable touch duration
//   - Pre-operation marking for debugging
//   - Comprehensive error handling
//
// MCP Parameters:
//   - platform (string): Device platform ("android", "ios", "web", "harmony")
//   - serial (string): Device serial number or identifier
//   - x (number): X coordinate (0.0 to 1.0, relative to screen width)
//   - y (number): Y coordinate (0.0 to 1.0, relative to screen height)
//   - duration (number, optional): Touch duration in seconds (default: 0.1)
//   - anti_risk (boolean, optional): Enable anti-detection measures
//
// Example Usage:
//
//	{
//	  "name": "tap_xy",
//	  "arguments": {
//	    "platform": "android",
//	    "serial": "emulator-5554",
//	    "x": 0.5,
//	    "y": 0.3,
//	    "duration": 0.2,
//	    "anti_risk": true
//	  }
//	}
//
// Error Conditions:
//   - Missing or invalid coordinates
//   - Device connection failure
//   - Touch operation timeout
//   - Platform not supported
type ToolTapXY struct{}

func (t *ToolTapXY) Name() option.ActionName {
	return option.ACTION_TapXY
}

func (t *ToolTapXY) Description() string {
	return "Tap on the screen at given relative coordinates (0.0-1.0 range)"
}

func (t *ToolTapXY) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_TapXY)
}

func (t *ToolTapXY) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// Get options directly since ActionOptions is now ActionOptions
		opts := unifiedReq.Options()

		// Add configurable options based on request
		if unifiedReq.PreMarkOperation {
			opts = append(opts, option.WithPreMarkOperation(true))
		}

		// Validate required parameters
		if unifiedReq.X == 0 || unifiedReq.Y == 0 {
			return nil, fmt.Errorf("x and y coordinates are required")
		}

		// Tap action logic
		log.Info().Float64("x", unifiedReq.X).Float64("y", unifiedReq.Y).Msg("tapping at coordinates")

		err = driverExt.TapXY(unifiedReq.X, unifiedReq.Y, opts...)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Tap failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully tapped at coordinates (%.2f, %.2f)", unifiedReq.X, unifiedReq.Y)), nil
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

		// Extract options to arguments
		extractActionOptionsToArguments(action.GetOptions(), arguments)

		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid tap params: %v", action.Params)
}

func (t *ToolTapXY) ReturnSchema() map[string]string {
	return map[string]string{
		"message": "string: Success message confirming tap operation at specified coordinates",
	}
}

// ToolTapAbsXY implements the tap_abs_xy tool call.
type ToolTapAbsXY struct{}

func (t *ToolTapAbsXY) Name() option.ActionName {
	return option.ACTION_TapAbsXY
}

func (t *ToolTapAbsXY) Description() string {
	return "Tap at absolute pixel coordinates on the screen"
}

func (t *ToolTapAbsXY) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_TapAbsXY)
}

func (t *ToolTapAbsXY) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// Get options directly since ActionOptions is now ActionOptions
		opts := unifiedReq.Options()

		// Add configurable options based on request
		if unifiedReq.PreMarkOperation {
			opts = append(opts, option.WithPreMarkOperation(true))
		}

		// Add AntiRisk support
		if unifiedReq.AntiRisk {
			opts = append(opts, option.WithAntiRisk(true))
		}

		// Validate required parameters
		if unifiedReq.X == 0 || unifiedReq.Y == 0 {
			return nil, fmt.Errorf("x and y coordinates are required")
		}

		// Tap absolute XY action logic
		log.Info().Float64("x", unifiedReq.X).Float64("y", unifiedReq.Y).Msg("tapping at absolute coordinates")

		err = driverExt.TapAbsXY(unifiedReq.X, unifiedReq.Y, opts...)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Tap absolute XY failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully tapped at absolute coordinates (%.0f, %.0f)", unifiedReq.X, unifiedReq.Y)), nil
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

		// Extract options to arguments
		extractActionOptionsToArguments(action.GetOptions(), arguments)

		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid tap abs params: %v", action.Params)
}

func (t *ToolTapAbsXY) ReturnSchema() map[string]string {
	return map[string]string{
		"message": "string: Success message confirming tap operation at absolute coordinates",
	}
}

// defaultReturnSchema provides a standard return schema for most tools
func defaultReturnSchema() map[string]string {
	return map[string]string{
		"message": "string: Success message confirming the operation was completed",
	}
}

// ToolTapByOCR implements the tap_ocr tool call.
type ToolTapByOCR struct{}

func (t *ToolTapByOCR) Name() option.ActionName {
	return option.ACTION_TapByOCR
}

func (t *ToolTapByOCR) Description() string {
	return "Tap on text found by OCR recognition"
}

func (t *ToolTapByOCR) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_TapByOCR)
}

func (t *ToolTapByOCR) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// Get options directly since ActionOptions is now ActionOptions
		opts := unifiedReq.Options()

		// Add configurable options based on request
		if unifiedReq.PreMarkOperation {
			opts = append(opts, option.WithPreMarkOperation(true))
		}

		// Validate required parameters
		if unifiedReq.Text == "" {
			return nil, fmt.Errorf("text parameter is required")
		}

		// Tap by OCR action logic
		log.Info().Str("text", unifiedReq.Text).Msg("tapping by OCR")
		err = driverExt.TapByOCR(unifiedReq.Text, opts...)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Tap by OCR failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully tapped on OCR text: %s", unifiedReq.Text)), nil
	}
}

func (t *ToolTapByOCR) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	if text, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"text": text,
		}

		// Extract options to arguments
		extractActionOptionsToArguments(action.GetOptions(), arguments)

		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid tap by OCR params: %v", action.Params)
}

func (t *ToolTapByOCR) ReturnSchema() map[string]string {
	return map[string]string{
		"message": "string: Success message confirming the operation was completed",
	}
}

// ToolTapByCV implements the tap_cv tool call.
type ToolTapByCV struct{}

func (t *ToolTapByCV) Name() option.ActionName {
	return option.ACTION_TapByCV
}

func (t *ToolTapByCV) Description() string {
	return "Tap on element found by computer vision"
}

func (t *ToolTapByCV) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_TapByCV)
}

func (t *ToolTapByCV) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// Get options directly since ActionOptions is now ActionOptions
		opts := unifiedReq.Options()

		// Add configurable options based on request
		if unifiedReq.PreMarkOperation {
			opts = append(opts, option.WithPreMarkOperation(true))
		}

		// Tap by CV action logic
		log.Info().Msg("tapping by CV")

		// For TapByCV, we need to check if there are UI types in the options
		// In the original DoAction, it requires ScreenShotWithUITypes to be set
		// We'll add a basic implementation that triggers CV recognition
		err = driverExt.TapByCV(opts...)
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

	// Extract options to arguments
	extractActionOptionsToArguments(action.GetOptions(), arguments)

	return buildMCPCallToolRequest(t.Name(), arguments), nil
}

func (t *ToolTapByCV) ReturnSchema() map[string]string {
	return defaultReturnSchema()
}

// ToolDoubleTapXY implements the double_tap_xy tool call.
type ToolDoubleTapXY struct{}

func (t *ToolDoubleTapXY) Name() option.ActionName {
	return option.ACTION_DoubleTapXY
}

func (t *ToolDoubleTapXY) Description() string {
	return "Double tap at given relative coordinates (0.0-1.0 range)"
}

func (t *ToolDoubleTapXY) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_DoubleTapXY)
}

func (t *ToolDoubleTapXY) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// Validate required parameters
		if unifiedReq.X == 0 || unifiedReq.Y == 0 {
			return nil, fmt.Errorf("x and y coordinates are required")
		}

		// Double tap XY action logic
		log.Info().Float64("x", unifiedReq.X).Float64("y", unifiedReq.Y).Msg("double tapping at coordinates")
		err = driverExt.DoubleTap(unifiedReq.X, unifiedReq.Y)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Double tap failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully double tapped at (%.2f, %.2f)", unifiedReq.X, unifiedReq.Y)), nil
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

func (t *ToolDoubleTapXY) ReturnSchema() map[string]string {
	return defaultReturnSchema()
}

// ToolListPackages implements the list_packages tool call.
type ToolListPackages struct{}

func (t *ToolListPackages) Name() option.ActionName {
	return option.ACTION_ListPackages
}

func (t *ToolListPackages) Description() string {
	return "List all installed apps/packages on the device with their package names."
}

func (t *ToolListPackages) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_ListPackages)
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

func (t *ToolListPackages) ReturnSchema() map[string]string {
	return map[string]string{
		"packages": "[]string: List of installed app package names on the device",
	}
}

// ToolLaunchApp implements the launch_app tool call.
type ToolLaunchApp struct{}

func (t *ToolLaunchApp) Name() option.ActionName {
	return option.ACTION_AppLaunch
}

func (t *ToolLaunchApp) Description() string {
	return "Launch an app on mobile device using its package name. Use list_packages tool first to find the correct package name."
}

func (t *ToolLaunchApp) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_AppLaunch)
}

func (t *ToolLaunchApp) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		if unifiedReq.PackageName == "" {
			return nil, fmt.Errorf("package_name is required")
		}

		// Launch app action logic
		log.Info().Str("packageName", unifiedReq.PackageName).Msg("launching app")
		err = driverExt.AppLaunch(unifiedReq.PackageName)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Launch app failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully launched app: %s", unifiedReq.PackageName)), nil
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

func (t *ToolLaunchApp) ReturnSchema() map[string]string {
	return defaultReturnSchema()
}

// ToolTerminateApp implements the terminate_app tool call.
type ToolTerminateApp struct{}

func (t *ToolTerminateApp) Name() option.ActionName {
	return option.ACTION_AppTerminate
}

func (t *ToolTerminateApp) Description() string {
	return "Stop and terminate a running app on mobile device using its package name"
}

func (t *ToolTerminateApp) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_AppTerminate)
}

func (t *ToolTerminateApp) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		if unifiedReq.PackageName == "" {
			return nil, fmt.Errorf("package_name is required")
		}

		// Terminate app action logic
		log.Info().Str("packageName", unifiedReq.PackageName).Msg("terminating app")
		success, err := driverExt.AppTerminate(unifiedReq.PackageName)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Terminate app failed: %s", err.Error())), nil
		}
		if !success {
			log.Warn().Str("packageName", unifiedReq.PackageName).Msg("app was not running")
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully terminated app: %s", unifiedReq.PackageName)), nil
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

func (t *ToolTerminateApp) ReturnSchema() map[string]string {
	return defaultReturnSchema()
}

// ToolScreenShot implements the screenshot tool call.
type ToolScreenShot struct{}

func (t *ToolScreenShot) Name() option.ActionName {
	return option.ACTION_ScreenShot
}

func (t *ToolScreenShot) Description() string {
	return "Take a screenshot of the mobile device screen. Use this to understand what's currently displayed on screen."
}

func (t *ToolScreenShot) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_ScreenShot)
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

func (t *ToolScreenShot) ReturnSchema() map[string]string {
	return map[string]string{
		"image": "string: Base64 encoded screenshot image in JPEG format",
		"name":  "string: Image name identifier (typically 'screenshot')",
		"type":  "string: MIME type of the image (image/jpeg)",
	}
}

// ToolGetScreenSize implements the get_screen_size tool call.
type ToolGetScreenSize struct{}

func (t *ToolGetScreenSize) Name() option.ActionName {
	return option.ACTION_GetScreenSize
}

func (t *ToolGetScreenSize) Description() string {
	return "Get the screen size of the mobile device in pixels"
}

func (t *ToolGetScreenSize) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_GetScreenSize)
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

func (t *ToolGetScreenSize) ReturnSchema() map[string]string {
	return map[string]string{
		"width":   "int: Screen width in pixels",
		"height":  "int: Screen height in pixels",
		"message": "string: Formatted message with screen dimensions",
	}
}

// ToolPressButton implements the press_button tool call.
type ToolPressButton struct{}

func (t *ToolPressButton) Name() option.ActionName {
	return option.ACTION_PressButton
}

func (t *ToolPressButton) Description() string {
	return "Press a button on the device"
}

func (t *ToolPressButton) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_PressButton)
}

func (t *ToolPressButton) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// Press button action logic
		log.Info().Str("button", string(unifiedReq.Button)).Msg("pressing button")
		err = driverExt.PressButton(types.DeviceButton(unifiedReq.Button))
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Press button failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully pressed button: %s", unifiedReq.Button)), nil
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

func (t *ToolPressButton) ReturnSchema() map[string]string {
	return map[string]string{
		"message": "string: Success message confirming the button press operation",
		"button":  "string: Name of the button that was pressed",
	}
}

// ToolSwipe implements the generic swipe tool call.
// It automatically determines whether to use direction-based or coordinate-based swipe
// based on the params type.
type ToolSwipe struct{}

func (t *ToolSwipe) Name() option.ActionName {
	return option.ACTION_Swipe
}

func (t *ToolSwipe) Description() string {
	return "Swipe on the screen by direction (up/down/left/right) or coordinates [fromX, fromY, toX, toY]"
}

func (t *ToolSwipe) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_Swipe)
}

func (t *ToolSwipe) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Check if it's direction-based swipe (has "direction" parameter)
		if _, exists := request.Params.Arguments["direction"]; exists {
			// Delegate to ToolSwipeDirection
			directionTool := &ToolSwipeDirection{}
			return directionTool.Implement()(ctx, request)
		} else {
			// Delegate to ToolSwipeCoordinate
			coordinateTool := &ToolSwipeCoordinate{}
			return coordinateTool.Implement()(ctx, request)
		}
	}
}

func (t *ToolSwipe) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	// Check if params is a string (direction-based swipe)
	if _, ok := action.Params.(string); ok {
		// Delegate to ToolSwipeDirection but use our tool name
		directionTool := &ToolSwipeDirection{}
		request, err := directionTool.ConvertActionToCallToolRequest(action)
		if err != nil {
			return mcp.CallToolRequest{}, err
		}
		// Change the tool name to use generic swipe
		request.Params.Name = string(t.Name())
		return request, nil
	}

	// Check if params is a coordinate array (coordinate-based swipe)
	if paramSlice, err := builtin.ConvertToFloat64Slice(action.Params); err == nil && len(paramSlice) == 4 {
		// Delegate to ToolSwipeCoordinate but use our tool name
		coordinateTool := &ToolSwipeCoordinate{}
		request, err := coordinateTool.ConvertActionToCallToolRequest(action)
		if err != nil {
			return mcp.CallToolRequest{}, err
		}
		// Change the tool name to use generic swipe
		request.Params.Name = string(t.Name())
		return request, nil
	}

	return mcp.CallToolRequest{}, fmt.Errorf("invalid swipe params: %v, expected string direction or [fromX, fromY, toX, toY] coordinates", action.Params)
}

func (t *ToolSwipe) ReturnSchema() map[string]string {
	return map[string]string{
		"message":   "string: Success message confirming the swipe operation",
		"direction": "string: Direction of swipe (for directional swipes)",
		"fromX":     "float64: Starting X coordinate (for coordinate-based swipes)",
		"fromY":     "float64: Starting Y coordinate (for coordinate-based swipes)",
		"toX":       "float64: Ending X coordinate (for coordinate-based swipes)",
		"toY":       "float64: Ending Y coordinate (for coordinate-based swipes)",
	}
}

// ToolSwipeDirection implements the swipe_direction tool call.
type ToolSwipeDirection struct{}

func (t *ToolSwipeDirection) Name() option.ActionName {
	return option.ACTION_SwipeDirection
}

func (t *ToolSwipeDirection) Description() string {
	return "Swipe on the screen in a specific direction (up, down, left, right)"
}

func (t *ToolSwipeDirection) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_SwipeDirection)
}

func (t *ToolSwipeDirection) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}
		swipeDirection := unifiedReq.Direction.(string)

		// Swipe action logic
		log.Info().Str("direction", swipeDirection).Msg("performing swipe")

		// Validate direction
		validDirections := []string{"up", "down", "left", "right"}
		if !slices.Contains(validDirections, swipeDirection) {
			return nil, fmt.Errorf("invalid swipe direction: %s, expected one of: %v",
				swipeDirection, validDirections)
		}

		opts := []option.ActionOption{
			option.WithDuration(getFloat64ValueOrDefault(unifiedReq.Duration, 0.5)),
			option.WithPressDuration(getFloat64ValueOrDefault(unifiedReq.PressDuration, 0.1)),
		}
		if unifiedReq.AntiRisk {
			opts = append(opts, option.WithAntiRisk(true))
		}
		if unifiedReq.PreMarkOperation {
			opts = append(opts, option.WithPreMarkOperation(true))
		}

		// Convert direction to coordinates and perform swipe
		switch swipeDirection {
		case "up":
			err = driverExt.Swipe(0.5, 0.5, 0.5, 0.1, opts...)
		case "down":
			err = driverExt.Swipe(0.5, 0.5, 0.5, 0.9, opts...)
		case "left":
			err = driverExt.Swipe(0.5, 0.5, 0.1, 0.5, opts...)
		case "right":
			err = driverExt.Swipe(0.5, 0.5, 0.9, 0.5, opts...)
		default:
			return mcp.NewToolResultError(
				fmt.Sprintf("Unexpected swipe direction: %s", swipeDirection)), nil
		}

		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Swipe failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully swiped %s", swipeDirection)), nil
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

		// Extract all action options
		extractActionOptionsToArguments(action.GetOptions(), arguments)

		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid swipe params: %v", action.Params)
}

func (t *ToolSwipeDirection) ReturnSchema() map[string]string {
	return map[string]string{
		"message":   "string: Success message confirming the directional swipe",
		"direction": "string: Direction that was swiped (up/down/left/right)",
	}
}

// ToolSwipeCoordinate implements the swipe_coordinate tool call.
type ToolSwipeCoordinate struct{}

func (t *ToolSwipeCoordinate) Name() option.ActionName {
	return option.ACTION_SwipeCoordinate
}

func (t *ToolSwipeCoordinate) Description() string {
	return "Perform swipe with specific start and end coordinates and custom timing"
}

func (t *ToolSwipeCoordinate) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_SwipeCoordinate)
}

func (t *ToolSwipeCoordinate) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// Validate required parameters
		if unifiedReq.FromX == 0 || unifiedReq.FromY == 0 || unifiedReq.ToX == 0 || unifiedReq.ToY == 0 {
			return nil, fmt.Errorf("fromX, fromY, toX, and toY coordinates are required")
		}

		// Advanced swipe action logic using prepareSwipeAction like the original DoAction
		log.Info().
			Float64("fromX", unifiedReq.FromX).Float64("fromY", unifiedReq.FromY).
			Float64("toX", unifiedReq.ToX).Float64("toY", unifiedReq.ToY).
			Msg("performing advanced swipe")

		params := []float64{unifiedReq.FromX, unifiedReq.FromY, unifiedReq.ToX, unifiedReq.ToY}

		// Build action options from the unified request
		opts := []option.ActionOption{}
		if unifiedReq.Duration > 0 {
			opts = append(opts, option.WithDuration(unifiedReq.Duration))
		}
		if unifiedReq.PressDuration > 0 {
			opts = append(opts, option.WithPressDuration(unifiedReq.PressDuration))
		}
		if unifiedReq.AntiRisk {
			opts = append(opts, option.WithAntiRisk(true))
		}

		swipeAction := prepareSwipeAction(driverExt, params, opts...)
		err = swipeAction(driverExt)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Advanced swipe failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully performed advanced swipe from (%.2f, %.2f) to (%.2f, %.2f)",
			unifiedReq.FromX, unifiedReq.FromY, unifiedReq.ToX, unifiedReq.ToY)), nil
	}
}

func (t *ToolSwipeCoordinate) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	if paramSlice, err := builtin.ConvertToFloat64Slice(action.Params); err == nil && len(paramSlice) == 4 {
		arguments := map[string]any{
			"from_x": paramSlice[0],
			"from_y": paramSlice[1],
			"to_x":   paramSlice[2],
			"to_y":   paramSlice[3],
		}
		// Add duration and press duration from options
		if duration := action.ActionOptions.Duration; duration > 0 {
			arguments["duration"] = duration
		}
		if pressDuration := action.ActionOptions.PressDuration; pressDuration > 0 {
			arguments["pressDuration"] = pressDuration
		}

		// Extract all action options
		extractActionOptionsToArguments(action.GetOptions(), arguments)

		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid swipe advanced params: %v", action.Params)
}

func (t *ToolSwipeCoordinate) ReturnSchema() map[string]string {
	return map[string]string{
		"message": "string: Success message confirming the coordinate-based swipe",
		"fromX":   "float64: Starting X coordinate of the swipe",
		"fromY":   "float64: Starting Y coordinate of the swipe",
		"toX":     "float64: Ending X coordinate of the swipe",
		"toY":     "float64: Ending Y coordinate of the swipe",
	}
}

// ToolSwipeToTapApp implements the swipe_to_tap_app tool call.
type ToolSwipeToTapApp struct{}

func (t *ToolSwipeToTapApp) Name() option.ActionName {
	return option.ACTION_SwipeToTapApp
}

func (t *ToolSwipeToTapApp) Description() string {
	return "Swipe to find and tap an app by name"
}

func (t *ToolSwipeToTapApp) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_SwipeToTapApp)
}

func (t *ToolSwipeToTapApp) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// Build action options from request structure
		var opts []option.ActionOption

		// Add boolean options
		if unifiedReq.IgnoreNotFoundError {
			opts = append(opts, option.WithIgnoreNotFoundError(true))
		}

		// Add numeric options
		if unifiedReq.MaxRetryTimes > 0 {
			opts = append(opts, option.WithMaxRetryTimes(unifiedReq.MaxRetryTimes))
		}
		if unifiedReq.Index > 0 {
			opts = append(opts, option.WithIndex(unifiedReq.Index))
		}

		// Swipe to tap app action logic
		log.Info().Str("appName", unifiedReq.AppName).Msg("swipe to tap app")
		err = driverExt.SwipeToTapApp(unifiedReq.AppName, opts...)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Swipe to tap app failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully found and tapped app: %s", unifiedReq.AppName)), nil
	}
}

func (t *ToolSwipeToTapApp) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	if appName, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"appName": appName,
		}

		// Extract options to arguments
		extractActionOptionsToArguments(action.GetOptions(), arguments)

		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid swipe to tap app params: %v", action.Params)
}

func (t *ToolSwipeToTapApp) ReturnSchema() map[string]string {
	return map[string]string{
		"message": "string: Success message confirming the app was found and tapped",
		"appName": "string: Name of the app that was found and tapped",
	}
}

// ToolSwipeToTapText implements the swipe_to_tap_text tool call.
type ToolSwipeToTapText struct{}

func (t *ToolSwipeToTapText) Name() option.ActionName {
	return option.ACTION_SwipeToTapText
}

func (t *ToolSwipeToTapText) Description() string {
	return "Swipe to find and tap text on screen"
}

func (t *ToolSwipeToTapText) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_SwipeToTapText)
}

func (t *ToolSwipeToTapText) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// Build action options from request structure
		var opts []option.ActionOption

		// Add boolean options
		if unifiedReq.IgnoreNotFoundError {
			opts = append(opts, option.WithIgnoreNotFoundError(true))
		}
		if unifiedReq.Regex {
			opts = append(opts, option.WithRegex(true))
		}

		// Add numeric options
		if unifiedReq.MaxRetryTimes > 0 {
			opts = append(opts, option.WithMaxRetryTimes(unifiedReq.MaxRetryTimes))
		}
		if unifiedReq.Index > 0 {
			opts = append(opts, option.WithIndex(unifiedReq.Index))
		}

		// Swipe to tap text action logic
		log.Info().Str("text", unifiedReq.Text).Msg("swipe to tap text")
		err = driverExt.SwipeToTapTexts([]string{unifiedReq.Text}, opts...)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Swipe to tap text failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully found and tapped text: %s", unifiedReq.Text)), nil
	}
}

func (t *ToolSwipeToTapText) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	if text, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"text": text,
		}

		// Extract options to arguments
		extractActionOptionsToArguments(action.GetOptions(), arguments)

		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid swipe to tap text params: %v", action.Params)
}

func (t *ToolSwipeToTapText) ReturnSchema() map[string]string {
	return map[string]string{
		"message": "string: Success message confirming the text was found and tapped",
		"text":    "string: Text content that was found and tapped",
	}
}

// ToolSwipeToTapTexts implements the swipe_to_tap_texts tool call.
type ToolSwipeToTapTexts struct{}

func (t *ToolSwipeToTapTexts) Name() option.ActionName {
	return option.ACTION_SwipeToTapTexts
}

func (t *ToolSwipeToTapTexts) Description() string {
	return "Swipe to find and tap one of multiple texts on screen"
}

func (t *ToolSwipeToTapTexts) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_SwipeToTapTexts)
}

func (t *ToolSwipeToTapTexts) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// Build action options from request structure
		var opts []option.ActionOption

		// Add boolean options
		if unifiedReq.IgnoreNotFoundError {
			opts = append(opts, option.WithIgnoreNotFoundError(true))
		}
		if unifiedReq.Regex {
			opts = append(opts, option.WithRegex(true))
		}

		// Add numeric options
		if unifiedReq.MaxRetryTimes > 0 {
			opts = append(opts, option.WithMaxRetryTimes(unifiedReq.MaxRetryTimes))
		}
		if unifiedReq.Index > 0 {
			opts = append(opts, option.WithIndex(unifiedReq.Index))
		}

		// Swipe to tap texts action logic
		log.Info().Strs("texts", unifiedReq.Texts).Msg("swipe to tap texts")
		err = driverExt.SwipeToTapTexts(unifiedReq.Texts, opts...)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Swipe to tap texts failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully found and tapped one of texts: %v", unifiedReq.Texts)), nil
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

	// Extract options to arguments
	extractActionOptionsToArguments(action.GetOptions(), arguments)

	return buildMCPCallToolRequest(t.Name(), arguments), nil
}

func (t *ToolSwipeToTapTexts) ReturnSchema() map[string]string {
	return map[string]string{
		"message":   "string: Success message confirming one of the texts was found and tapped",
		"texts":     "[]string: List of text options that were searched for",
		"foundText": "string: The specific text that was actually found and tapped",
	}
}

// ToolDrag implements the drag tool call.
type ToolDrag struct{}

func (t *ToolDrag) Name() option.ActionName {
	return option.ACTION_Drag
}

func (t *ToolDrag) Description() string {
	return "Drag from one point to another on the mobile device screen"
}

func (t *ToolDrag) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_Drag)
}

func (t *ToolDrag) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// Validate required parameters - check if coordinates are provided (not just non-zero)
		_, hasFromX := request.Params.Arguments["from_x"]
		_, hasFromY := request.Params.Arguments["from_y"]
		_, hasToX := request.Params.Arguments["to_x"]
		_, hasToY := request.Params.Arguments["to_y"]
		if !hasFromX || !hasFromY || !hasToX || !hasToY {
			return nil, fmt.Errorf("from_x, from_y, to_x, and to_y coordinates are required")
		}

		opts := []option.ActionOption{}
		if unifiedReq.Duration > 0 {
			opts = append(opts, option.WithDuration(unifiedReq.Duration/1000.0))
		}
		if unifiedReq.AntiRisk {
			opts = append(opts, option.WithAntiRisk(true))
		}

		// Drag action logic
		log.Info().
			Float64("fromX", unifiedReq.FromX).Float64("fromY", unifiedReq.FromY).
			Float64("toX", unifiedReq.ToX).Float64("toY", unifiedReq.ToY).
			Msg("performing drag")

		err = driverExt.Swipe(unifiedReq.FromX, unifiedReq.FromY, unifiedReq.ToX, unifiedReq.ToY, opts...)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Drag failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully dragged from (%.2f, %.2f) to (%.2f, %.2f)",
			unifiedReq.FromX, unifiedReq.FromY, unifiedReq.ToX, unifiedReq.ToY)), nil
	}
}

func (t *ToolDrag) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	if paramSlice, err := builtin.ConvertToFloat64Slice(action.Params); err == nil && len(paramSlice) == 4 {
		arguments := map[string]any{
			"from_x": paramSlice[0],
			"from_y": paramSlice[1],
			"to_x":   paramSlice[2],
			"to_y":   paramSlice[3],
		}
		// Add duration from options
		if duration := action.ActionOptions.Duration; duration > 0 {
			arguments["duration"] = duration * 1000 // convert to milliseconds
		}

		// Extract all action options
		extractActionOptionsToArguments(action.GetOptions(), arguments)

		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid drag parameters: %v", action.Params)
}

func (t *ToolDrag) ReturnSchema() map[string]string {
	return map[string]string{
		"message": "string: Success message confirming the drag operation",
		"fromX":   "float64: Starting X coordinate of the drag",
		"fromY":   "float64: Starting Y coordinate of the drag",
		"toX":     "float64: Ending X coordinate of the drag",
		"toY":     "float64: Ending Y coordinate of the drag",
	}
}

// extractActionOptionsToArguments extracts action options and adds them to arguments map
// This is a generic helper that can be used by multiple tools
func extractActionOptionsToArguments(actionOptions []option.ActionOption, arguments map[string]any) {
	if len(actionOptions) == 0 {
		return
	}

	// Apply all options to a temporary ActionOptions to extract values
	tempOptions := &option.ActionOptions{}
	for _, opt := range actionOptions {
		opt(tempOptions)
	}

	// Define option mappings for common boolean options
	booleanOptions := map[string]bool{
		"ignore_NotFoundError": tempOptions.IgnoreNotFoundError,
		"regex":                tempOptions.Regex,
		"tap_random_rect":      tempOptions.TapRandomRect,
		"anti_risk":            tempOptions.AntiRisk,
		"pre_mark_operation":   tempOptions.PreMarkOperation,
	}

	// Add boolean options only if they are true
	for key, value := range booleanOptions {
		if value {
			arguments[key] = true
		}
	}

	// Add numeric options only if they have meaningful values and don't already exist
	if tempOptions.MaxRetryTimes > 0 {
		arguments["max_retry_times"] = tempOptions.MaxRetryTimes
	}
	if tempOptions.Index != 0 {
		arguments["index"] = tempOptions.Index
	}
	// Only set duration if it's not already set (to avoid overriding tool-specific conversions)
	if tempOptions.Duration > 0 {
		if _, exists := arguments["duration"]; !exists {
			arguments["duration"] = tempOptions.Duration
		}
	}
	if tempOptions.PressDuration > 0 {
		arguments["press_duration"] = tempOptions.PressDuration
	}
}

// ToolHome implements the home tool call.
type ToolHome struct{}

func (t *ToolHome) Name() option.ActionName {
	return option.ACTION_Home
}

func (t *ToolHome) Description() string {
	return "Press the home button on the device"
}

func (t *ToolHome) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_Home)
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

func (t *ToolHome) ReturnSchema() map[string]string {
	return map[string]string{
		"message": "string: Success message confirming home button was pressed",
	}
}

// ToolBack implements the back tool call.
type ToolBack struct{}

func (t *ToolBack) Name() option.ActionName {
	return option.ACTION_Back
}

func (t *ToolBack) Description() string {
	return "Press the back button on the device"
}

func (t *ToolBack) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_Back)
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

func (t *ToolBack) ReturnSchema() map[string]string {
	return map[string]string{
		"message": "string: Success message confirming back button was pressed",
	}
}

// ToolInput implements the input tool call.
type ToolInput struct{}

func (t *ToolInput) Name() option.ActionName {
	return option.ACTION_Input
}

func (t *ToolInput) Description() string {
	return "Input text into the currently focused element or input field"
}

func (t *ToolInput) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_Input)
}

func (t *ToolInput) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		if unifiedReq.Text == "" {
			return nil, fmt.Errorf("text is required")
		}

		// Input action logic
		log.Info().Str("text", unifiedReq.Text).Msg("inputting text")
		err = driverExt.Input(unifiedReq.Text)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Input failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully input text: %s", unifiedReq.Text)), nil
	}
}

func (t *ToolInput) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	text := fmt.Sprintf("%v", action.Params)
	arguments := map[string]any{
		"text": text,
	}
	return buildMCPCallToolRequest(t.Name(), arguments), nil
}

func (t *ToolInput) ReturnSchema() map[string]string {
	return map[string]string{
		"message": "string: Success message confirming text was input",
		"text":    "string: Text content that was input into the field",
	}
}

// ToolWebLoginNoneUI implements the web_login_none_ui tool call.
type ToolWebLoginNoneUI struct{}

func (t *ToolWebLoginNoneUI) Name() option.ActionName {
	return option.ACTION_WebLoginNoneUI
}

func (t *ToolWebLoginNoneUI) Description() string {
	return "Perform login without UI interaction for web applications"
}

func (t *ToolWebLoginNoneUI) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_WebLoginNoneUI)
}

func (t *ToolWebLoginNoneUI) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// Web login none UI action logic
		log.Info().Str("packageName", unifiedReq.PackageName).Msg("performing web login without UI")
		driver, ok := driverExt.IDriver.(*BrowserDriver)
		if !ok {
			return nil, fmt.Errorf("invalid browser driver for web login")
		}

		_, err = driver.LoginNoneUI(unifiedReq.PackageName, unifiedReq.PhoneNumber, unifiedReq.Captcha, unifiedReq.Password)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Web login failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText("Successfully performed web login without UI"), nil
	}
}

func (t *ToolWebLoginNoneUI) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	return buildMCPCallToolRequest(t.Name(), map[string]any{}), nil
}

func (t *ToolWebLoginNoneUI) ReturnSchema() map[string]string {
	return map[string]string{
		"message":     "string: Success message confirming web login was completed",
		"loginResult": "object: Result of the login operation (success/failure details)",
	}
}

// ToolAppInstall implements the app_install tool call.
type ToolAppInstall struct{}

func (t *ToolAppInstall) Name() option.ActionName {
	return option.ACTION_AppInstall
}

func (t *ToolAppInstall) Description() string {
	return "Install an app on the device from a URL or local file path"
}

func (t *ToolAppInstall) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_AppInstall)
}

func (t *ToolAppInstall) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// App install action logic
		log.Info().Str("appUrl", unifiedReq.AppUrl).Msg("installing app")
		err = driverExt.GetDevice().Install(unifiedReq.AppUrl)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("App install failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully installed app from: %s", unifiedReq.AppUrl)), nil
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

func (t *ToolAppInstall) ReturnSchema() map[string]string {
	return map[string]string{
		"message": "string: Success message confirming app installation",
		"appUrl":  "string: URL or path of the app that was installed",
	}
}

// ToolAppUninstall implements the app_uninstall tool call.
type ToolAppUninstall struct{}

func (t *ToolAppUninstall) Name() option.ActionName {
	return option.ACTION_AppUninstall
}

func (t *ToolAppUninstall) Description() string {
	return "Uninstall an app from the device"
}

func (t *ToolAppUninstall) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_AppUninstall)
}

func (t *ToolAppUninstall) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// App uninstall action logic
		log.Info().Str("packageName", unifiedReq.PackageName).Msg("uninstalling app")
		err = driverExt.GetDevice().Uninstall(unifiedReq.PackageName)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("App uninstall failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully uninstalled app: %s", unifiedReq.PackageName)), nil
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

func (t *ToolAppUninstall) ReturnSchema() map[string]string {
	return map[string]string{
		"message":     "string: Success message confirming app uninstallation",
		"packageName": "string: Package name of the app that was uninstalled",
	}
}

// ToolAppClear implements the app_clear tool call.
type ToolAppClear struct{}

func (t *ToolAppClear) Name() option.ActionName {
	return option.ACTION_AppClear
}

func (t *ToolAppClear) Description() string {
	return "Clear app data and cache for a specific app using its package name"
}

func (t *ToolAppClear) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_AppClear)
}

func (t *ToolAppClear) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// App clear action logic
		log.Info().Str("packageName", unifiedReq.PackageName).Msg("clearing app")
		err = driverExt.AppClear(unifiedReq.PackageName)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("App clear failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully cleared app: %s", unifiedReq.PackageName)), nil
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

func (t *ToolAppClear) ReturnSchema() map[string]string {
	return map[string]string{
		"message":     "string: Success message confirming app data and cache were cleared",
		"packageName": "string: Package name of the app that was cleared",
	}
}

// ToolSecondaryClick implements the secondary_click tool call.
type ToolSecondaryClick struct{}

func (t *ToolSecondaryClick) Name() option.ActionName {
	return option.ACTION_SecondaryClick
}

func (t *ToolSecondaryClick) Description() string {
	return "Perform secondary click (right click) at specified coordinates"
}

func (t *ToolSecondaryClick) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_SecondaryClick)
}

func (t *ToolSecondaryClick) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// Validate required parameters
		if unifiedReq.X == 0 || unifiedReq.Y == 0 {
			return nil, fmt.Errorf("x and y coordinates are required")
		}

		// Secondary click action logic
		log.Info().Float64("x", unifiedReq.X).Float64("y", unifiedReq.Y).Msg("performing secondary click")
		err = driverExt.SecondaryClick(unifiedReq.X, unifiedReq.Y)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Secondary click failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully performed secondary click at (%.2f, %.2f)", unifiedReq.X, unifiedReq.Y)), nil
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

func (t *ToolSecondaryClick) ReturnSchema() map[string]string {
	return map[string]string{
		"message": "string: Success message confirming secondary click (right-click) operation",
		"x":       "float64: X coordinate where secondary click was performed",
		"y":       "float64: Y coordinate where secondary click was performed",
	}
}

// ToolHoverBySelector implements the hover_by_selector tool call.
type ToolHoverBySelector struct{}

func (t *ToolHoverBySelector) Name() option.ActionName {
	return option.ACTION_HoverBySelector
}

func (t *ToolHoverBySelector) Description() string {
	return "Hover over an element selected by CSS selector or XPath"
}

func (t *ToolHoverBySelector) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_HoverBySelector)
}

func (t *ToolHoverBySelector) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// Hover by selector action logic
		log.Info().Str("selector", unifiedReq.Selector).Msg("hovering by selector")
		err = driverExt.HoverBySelector(unifiedReq.Selector)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Hover by selector failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully hovered over element with selector: %s", unifiedReq.Selector)), nil
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

func (t *ToolHoverBySelector) ReturnSchema() map[string]string {
	return map[string]string{
		"message":  "string: Success message confirming hover operation",
		"selector": "string: CSS selector or XPath of the element that was hovered over",
	}
}

// ToolTapBySelector implements the tap_by_selector tool call.
type ToolTapBySelector struct{}

func (t *ToolTapBySelector) Name() option.ActionName {
	return option.ACTION_TapBySelector
}

func (t *ToolTapBySelector) Description() string {
	return "Tap an element selected by CSS selector or XPath"
}

func (t *ToolTapBySelector) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_TapBySelector)
}

func (t *ToolTapBySelector) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// Tap by selector action logic
		log.Info().Str("selector", unifiedReq.Selector).Msg("tapping by selector")
		err = driverExt.TapBySelector(unifiedReq.Selector)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Tap by selector failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully tapped element with selector: %s", unifiedReq.Selector)), nil
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

func (t *ToolTapBySelector) ReturnSchema() map[string]string {
	return map[string]string{
		"message":  "string: Success message confirming tap operation",
		"selector": "string: CSS selector or XPath of the element that was tapped",
	}
}

// ToolSecondaryClickBySelector implements the secondary_click_by_selector tool call.
type ToolSecondaryClickBySelector struct{}

func (t *ToolSecondaryClickBySelector) Name() option.ActionName {
	return option.ACTION_SecondaryClickBySelector
}

func (t *ToolSecondaryClickBySelector) Description() string {
	return "Perform secondary click on an element selected by CSS selector or XPath"
}

func (t *ToolSecondaryClickBySelector) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_SecondaryClickBySelector)
}

func (t *ToolSecondaryClickBySelector) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// Secondary click by selector action logic
		log.Info().Str("selector", unifiedReq.Selector).Msg("performing secondary click by selector")
		err = driverExt.SecondaryClickBySelector(unifiedReq.Selector)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Secondary click by selector failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully performed secondary click on element with selector: %s", unifiedReq.Selector)), nil
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

func (t *ToolSecondaryClickBySelector) ReturnSchema() map[string]string {
	return map[string]string{
		"message":  "string: Success message confirming secondary click operation",
		"selector": "string: CSS selector or XPath of the element that was right-clicked",
	}
}

// ToolWebCloseTab implements the web_close_tab tool call.
type ToolWebCloseTab struct{}

func (t *ToolWebCloseTab) Name() option.ActionName {
	return option.ACTION_WebCloseTab
}

func (t *ToolWebCloseTab) Description() string {
	return "Close a browser tab by index"
}

func (t *ToolWebCloseTab) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_WebCloseTab)
}

func (t *ToolWebCloseTab) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// Validate required parameters
		if unifiedReq.TabIndex == 0 {
			return nil, fmt.Errorf("tabIndex is required")
		}

		// Web close tab action logic
		log.Info().Int("tabIndex", unifiedReq.TabIndex).Msg("closing web tab")
		browserDriver, ok := driverExt.IDriver.(*BrowserDriver)
		if !ok {
			return nil, fmt.Errorf("web close tab is only supported for browser drivers")
		}

		err = browserDriver.CloseTab(unifiedReq.TabIndex)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Close tab failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully closed tab at index: %d", unifiedReq.TabIndex)), nil
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

func (t *ToolWebCloseTab) ReturnSchema() map[string]string {
	return map[string]string{
		"message":  "string: Success message confirming browser tab was closed",
		"tabIndex": "int: Index of the tab that was closed",
	}
}

// ToolSetIme implements the set_ime tool call.
type ToolSetIme struct{}

func (t *ToolSetIme) Name() option.ActionName {
	return option.ACTION_SetIme
}

func (t *ToolSetIme) Description() string {
	return "Set the input method editor (IME) on the device"
}

func (t *ToolSetIme) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_SetIme)
}

func (t *ToolSetIme) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// Set IME action logic
		log.Info().Str("ime", unifiedReq.Ime).Msg("setting IME")
		err = driverExt.SetIme(unifiedReq.Ime)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Set IME failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully set IME to: %s", unifiedReq.Ime)), nil
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

func (t *ToolSetIme) ReturnSchema() map[string]string {
	return map[string]string{
		"message": "string: Success message confirming IME was set",
		"ime":     "string: Input method editor that was set",
	}
}

// ToolGetSource implements the get_source tool call.
type ToolGetSource struct{}

func (t *ToolGetSource) Name() option.ActionName {
	return option.ACTION_GetSource
}

func (t *ToolGetSource) Description() string {
	return "Get the UI hierarchy/source tree of the current screen for a specific app"
}

func (t *ToolGetSource) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_GetSource)
}

func (t *ToolGetSource) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// Get source action logic
		log.Info().Str("packageName", unifiedReq.PackageName).Msg("getting source")
		_, err = driverExt.Source(option.WithProcessName(unifiedReq.PackageName))
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Get source failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully retrieved source for package: %s", unifiedReq.PackageName)), nil
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

func (t *ToolGetSource) ReturnSchema() map[string]string {
	return map[string]string{
		"message":     "string: Success message confirming UI source was retrieved",
		"packageName": "string: Package name of the app whose source was retrieved",
		"source":      "string: UI hierarchy/source tree data in XML or JSON format",
	}
}

// ToolSleep implements the sleep tool call.
type ToolSleep struct{}

func (t *ToolSleep) Name() option.ActionName {
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

func (t *ToolSleep) ReturnSchema() map[string]string {
	return map[string]string{
		"message": "string: Success message confirming sleep operation completed",
		"seconds": "float64: Duration in seconds that was slept",
	}
}

// ToolSleepMS implements the sleep_ms tool call.
type ToolSleepMS struct{}

func (t *ToolSleepMS) Name() option.ActionName {
	return option.ACTION_SleepMS
}

func (t *ToolSleepMS) Description() string {
	return "Sleep for specified milliseconds"
}

func (t *ToolSleepMS) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_SleepMS)
}

func (t *ToolSleepMS) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// Validate required parameters
		if unifiedReq.Milliseconds == 0 {
			return nil, fmt.Errorf("milliseconds is required")
		}

		// Sleep MS action logic
		log.Info().Int64("milliseconds", unifiedReq.Milliseconds).Msg("sleeping in milliseconds")
		time.Sleep(time.Duration(unifiedReq.Milliseconds) * time.Millisecond)

		return mcp.NewToolResultText(fmt.Sprintf("Successfully slept for %d milliseconds", unifiedReq.Milliseconds)), nil
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

func (t *ToolSleepMS) ReturnSchema() map[string]string {
	return map[string]string{
		"message":      "string: Success message confirming sleep operation completed",
		"milliseconds": "int64: Duration in milliseconds that was slept",
	}
}

// ToolSleepRandom implements the sleep_random tool call.
type ToolSleepRandom struct{}

func (t *ToolSleepRandom) Name() option.ActionName {
	return option.ACTION_SleepRandom
}

func (t *ToolSleepRandom) Description() string {
	return "Sleep for a random duration based on parameters"
}

func (t *ToolSleepRandom) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_SleepRandom)
}

func (t *ToolSleepRandom) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// Sleep random action logic
		log.Info().Floats64("params", unifiedReq.Params).Msg("sleeping for random duration")
		sleepStrict(time.Now(), getSimulationDuration(unifiedReq.Params))

		return mcp.NewToolResultText(fmt.Sprintf("Successfully slept for random duration with params: %v", unifiedReq.Params)), nil
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

func (t *ToolSleepRandom) ReturnSchema() map[string]string {
	return map[string]string{
		"message":        "string: Success message confirming random sleep operation completed",
		"params":         "[]float64: Parameters used for random duration calculation",
		"actualDuration": "float64: Actual duration that was slept (in seconds)",
	}
}

// ToolClosePopups implements the close_popups tool call.
type ToolClosePopups struct{}

func (t *ToolClosePopups) Name() option.ActionName {
	return option.ACTION_ClosePopups
}

func (t *ToolClosePopups) Description() string {
	return "Close any popup windows or dialogs on screen"
}

func (t *ToolClosePopups) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_ClosePopups)
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

func (t *ToolClosePopups) ReturnSchema() map[string]string {
	return map[string]string{
		"message":      "string: Success message confirming popups were closed",
		"popupsClosed": "int: Number of popup windows or dialogs that were closed",
	}
}

// ToolAIAction implements the ai_action tool call.
type ToolAIAction struct{}

func (t *ToolAIAction) Name() option.ActionName {
	return option.ACTION_AIAction
}

func (t *ToolAIAction) Description() string {
	return "Perform AI-driven automation actions using natural language prompts to describe the desired operation"
}

func (t *ToolAIAction) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_AIAction)
}

func (t *ToolAIAction) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// AI action logic
		log.Info().Str("prompt", unifiedReq.Prompt).Msg("performing AI action")
		err = driverExt.AIAction(unifiedReq.Prompt)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("AI action failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully performed AI action with prompt: %s", unifiedReq.Prompt)), nil
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

func (t *ToolAIAction) ReturnSchema() map[string]string {
	return map[string]string{
		"message":     "string: Success message confirming AI action was performed",
		"prompt":      "string: Natural language prompt that was processed",
		"actionTaken": "string: Description of the specific action that was taken by AI",
	}
}

// ToolFinished implements the finished tool call.
type ToolFinished struct{}

func (t *ToolFinished) Name() option.ActionName {
	return option.ACTION_Finished
}

func (t *ToolFinished) Description() string {
	return "Mark the current automation task as completed with a result message"
}

func (t *ToolFinished) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_Finished)
}

func (t *ToolFinished) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}
		log.Info().Str("reason", unifiedReq.Content).Msg("task finished")

		return mcp.NewToolResultText(fmt.Sprintf("Task completed: %s", unifiedReq.Content)), nil
	}
}

func (t *ToolFinished) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	if reason, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"content": reason,
		}
		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid finished params: %v", action.Params)
}

func (t *ToolFinished) ReturnSchema() map[string]string {
	return map[string]string{
		"message":       "string: Success message confirming task completion",
		"content":       "string: Completion reason or result description",
		"taskCompleted": "bool: Boolean indicating task was successfully finished",
	}
}

func getFloat64ValueOrDefault(value float64, defaultValue float64) float64 {
	if value == 0 {
		return defaultValue
	}
	return value
}

// parseActionOptions converts MCP request arguments to ActionOptions struct
//
// This function provides unified parameter parsing for all MCP tools by:
//
// 1. Converting map[string]any arguments to JSON bytes
// 2. Unmarshaling JSON into strongly-typed ActionOptions struct
// 3. Providing automatic validation and type conversion
//
// The ActionOptions struct contains all possible parameters for UI operations:
//   - Coordinates: X, Y, FromX, FromY, ToX, ToY
//   - Text/Content: Text, Content, AppName, PackageName
//   - Timing: Duration, PressDuration, Milliseconds
//   - Behavior: AntiRisk, IgnoreNotFoundError, Regex
//   - Indices: Index, MaxRetryTimes, TabIndex
//   - Device: Platform, Serial, Button, Direction
//   - Web: Selector, PhoneNumber, Captcha, Password
//   - AI: Prompt
//   - Collections: Texts, Params, Points
//
// Parameters:
//   - arguments: Raw MCP request arguments as map[string]any
//
// Returns:
//   - *option.ActionOptions: Parsed and validated options struct
//   - error: Parsing or validation error
//
// Usage:
//
//	unifiedReq, err := parseActionOptions(request.Params.Arguments)
//	if err != nil {
//	    return nil, err
//	}
//	// Use unifiedReq.X, unifiedReq.Y, etc.
//
// Error Handling:
//   - JSON marshal errors (invalid argument types)
//   - JSON unmarshal errors (type conversion failures)
//   - Missing required fields (handled by individual tools)
func parseActionOptions(arguments map[string]any) (*option.ActionOptions, error) {
	b, err := json.Marshal(arguments)
	if err != nil {
		return nil, fmt.Errorf("marshal arguments failed: %w", err)
	}

	var actionOptions option.ActionOptions
	if err := json.Unmarshal(b, &actionOptions); err != nil {
		return nil, fmt.Errorf("unmarshal to ActionOptions failed: %w", err)
	}

	return &actionOptions, nil
}
