package mcphost

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/httprunner/httprunner/v5/internal/version"
	"github.com/httprunner/httprunner/v5/pkg/gadb"
	"github.com/httprunner/httprunner/v5/uixt"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/httprunner/httprunner/v5/uixt/types"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rs/zerolog/log"
)

// MCPClient4XTDriver is a minimal MCP client that only implements the methods used by the host
type MCPClient4XTDriver struct {
	client.MCPClient
	server *MCPServer4XTDriver
}

func (c *MCPClient4XTDriver) ListTools(ctx context.Context, req mcp.ListToolsRequest) (*mcp.ListToolsResult, error) {
	tools := c.server.ListTools()
	return &mcp.ListToolsResult{Tools: tools}, nil
}

func (c *MCPClient4XTDriver) CallTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	handler := c.server.GetHandler(req.Params.Name)
	if handler == nil {
		return mcp.NewToolResultError(fmt.Sprintf("handler for tool %s not found", req.Params.Name)), nil
	}
	return handler(ctx, req)
}

func (c *MCPClient4XTDriver) Initialize(ctx context.Context, req mcp.InitializeRequest) (*mcp.InitializeResult, error) {
	// no need to initialize for local server
	return &mcp.InitializeResult{}, nil
}

func (c *MCPClient4XTDriver) Close() error {
	// no need to close for local server
	return nil
}

type toolCall = func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error)

// MCPServer4XTDriver wraps a MCPServer to expose XTDriver functionality via MCP protocol.
type MCPServer4XTDriver struct {
	mcpServer   *server.MCPServer
	driverCache sync.Map            // key is serial, value is *XTDriver
	tools       []mcp.Tool          // tools list for uixt
	handlerMap  map[string]toolCall // tool name to handler
}

// NewMCPServer creates a new MCP server for XTDriver and registers all tools.
func NewMCPServer() *MCPServer4XTDriver {
	mcpServer := server.NewMCPServer(
		"uixt",
		version.GetVersionInfo(),
		server.WithToolCapabilities(false),
	)
	s := &MCPServer4XTDriver{
		mcpServer:  mcpServer,
		handlerMap: make(map[string]toolCall),
	}
	s.addTools()
	return s
}

// Start runs the MCP server (blocking).
func (s *MCPServer4XTDriver) Start() error {
	log.Info().Msg("Starting HttpRunner UIXT MCP Server...")
	return server.ServeStdio(s.mcpServer)
}

// addTools registers all MCP tools.
func (ums *MCPServer4XTDriver) addTools() {
	// ListAvailableDevices Tool
	listDevicesTool := mcp.NewTool("list_available_devices",
		mcp.WithDescription("List all available devices. If there are more than one device returned, you need to let the user select one of them."),
	)
	ums.mcpServer.AddTool(listDevicesTool, ums.handleListAvailableDevices)
	ums.tools = append(ums.tools, listDevicesTool)
	ums.handlerMap[listDevicesTool.Name] = ums.handleListAvailableDevices

	// SelectDevice Tool
	selectDeviceTool := mcp.NewTool("select_device",
		mcp.WithDescription("Select a device to use from the list of available devices. Use the list_available_devices tool to get a list of available devices."),
		mcp.WithString("platform", mcp.Enum("android", "ios"), mcp.Description("The type of device to select")),
		mcp.WithString("serial", mcp.Description("The device serial/udid to select")),
	)
	ums.mcpServer.AddTool(selectDeviceTool, ums.handleSelectDevice)
	ums.tools = append(ums.tools, selectDeviceTool)
	ums.handlerMap[selectDeviceTool.Name] = ums.handleSelectDevice

	// ListPackages Tool
	listPackagesParams := append(
		[]mcp.ToolOption{mcp.WithDescription("List all the apps/packages on the device.")},
		commonToolOptions...,
	)
	listPackagesTool := mcp.NewTool("list_packages", listPackagesParams...)
	ums.mcpServer.AddTool(listPackagesTool, ums.handleListPackages)
	ums.tools = append(ums.tools, listPackagesTool)
	ums.handlerMap[listPackagesTool.Name] = ums.handleListPackages

	// LaunchApp Tool
	launchAppParams := append(
		[]mcp.ToolOption{mcp.WithDescription("Launch an app on mobile device. Use this to open a specific app. You can find the package name of the app by calling list_packages.")},
		commonToolOptions...,
	)
	launchAppParams = append(launchAppParams, generateMCPOptions(types.AppLaunchRequest{})...)
	launchAppTool := mcp.NewTool("launch_app", launchAppParams...)
	ums.mcpServer.AddTool(launchAppTool, ums.handleLaunchApp)
	ums.tools = append(ums.tools, launchAppTool)
	ums.handlerMap[launchAppTool.Name] = ums.handleLaunchApp

	// TerminateApp Tool
	terminateAppParams := append(
		[]mcp.ToolOption{mcp.WithDescription("Stop and terminate an app on mobile device")},
		commonToolOptions...,
	)
	terminateAppParams = append(terminateAppParams, generateMCPOptions(types.AppTerminateRequest{})...)
	terminateAppTool := mcp.NewTool("terminate_app", terminateAppParams...)
	ums.mcpServer.AddTool(terminateAppTool, ums.handleTerminateApp)
	ums.tools = append(ums.tools, terminateAppTool)
	ums.handlerMap[terminateAppTool.Name] = ums.handleTerminateApp

	// GetScreenSize Tool
	getScreenSizeParams := append(
		[]mcp.ToolOption{mcp.WithDescription("Get the screen size of the mobile device in pixels")},
		commonToolOptions...,
	)
	getScreenSizeTool := mcp.NewTool("get_screen_size", getScreenSizeParams...)
	ums.mcpServer.AddTool(getScreenSizeTool, ums.handleGetScreenSize)
	ums.tools = append(ums.tools, getScreenSizeTool)
	ums.handlerMap[getScreenSizeTool.Name] = ums.handleGetScreenSize

	// PressButton Tool
	pressButtonParams := append(
		[]mcp.ToolOption{mcp.WithDescription("Press a button on device")},
		commonToolOptions...,
	)
	pressButtonTool := mcp.NewTool("press_button", pressButtonParams...)
	ums.mcpServer.AddTool(pressButtonTool, ums.handlePressButton)
	ums.tools = append(ums.tools, pressButtonTool)
	ums.handlerMap[pressButtonTool.Name] = ums.handlePressButton

	// TapXY Tool
	tapParams := append(
		[]mcp.ToolOption{mcp.WithDescription("Click on the screen at given x,y coordinates")},
		commonToolOptions...,
	)
	tapParams = append(tapParams, generateMCPOptions(types.TapRequest{})...)
	tapXYTool := mcp.NewTool("tap_xy", tapParams...)
	ums.mcpServer.AddTool(tapXYTool, ums.handleTapXY)
	ums.tools = append(ums.tools, tapXYTool)
	ums.handlerMap[tapXYTool.Name] = ums.handleTapXY
	log.Info().Str("name", tapXYTool.Name).Msg("Register tool")

	// Swipe Tool
	swipeParams := append(
		[]mcp.ToolOption{mcp.WithDescription("Swipe on the screen")},
		commonToolOptions...,
	)
	swipeParams = append(swipeParams, generateMCPOptions(types.SwipeRequest{})...)
	swipeTool := mcp.NewTool("swipe", swipeParams...)
	ums.mcpServer.AddTool(swipeTool, ums.handleSwipe)
	ums.tools = append(ums.tools, swipeTool)
	ums.handlerMap[swipeTool.Name] = ums.handleSwipe
	log.Info().Str("name", swipeTool.Name).Msg("Register tool")

	// ScreenShot Tool
	takeScreenShotParams := append(
		[]mcp.ToolOption{mcp.WithDescription("Take a screenshot of the mobile device. Use this to understand what's on screen. Do not cache this result.")},
		commonToolOptions...,
	)
	screenShotTool := mcp.NewTool("take_screenshot", takeScreenShotParams...)
	ums.mcpServer.AddTool(screenShotTool, ums.handleScreenShot)
	ums.tools = append(ums.tools, screenShotTool)
	ums.handlerMap[screenShotTool.Name] = ums.handleScreenShot
	log.Info().Str("name", screenShotTool.Name).Msg("Register tool")
}

// handleListAvailableDevices handles the list_available_devices tool call.
func (ums *MCPServer4XTDriver) handleListAvailableDevices(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
			device, err := uixt.NewIOSDevice(
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

// handleSelectDevice handles the select_device tool call.
func (ums *MCPServer4XTDriver) handleSelectDevice(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	driverExt, err := ums.setupXTDriver(ctx, request.Params.Arguments)
	if err != nil {
		return nil, err
	}

	uuid := driverExt.IDriver.GetDevice().UUID()
	return mcp.NewToolResultText(fmt.Sprintf("Selected device: %s", uuid)), nil
}

// handleListPackages handles the list_packages tool call.
func (ums *MCPServer4XTDriver) handleListPackages(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	driverExt, err := ums.setupXTDriver(ctx, request.Params.Arguments)
	if err != nil {
		return nil, err
	}

	apps, err := driverExt.IDriver.GetDevice().ListPackages()
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(fmt.Sprintf("Device packages: %v", apps)), nil
}

// handleLaunchApp handles the launch_app tool call.
func (ums *MCPServer4XTDriver) handleLaunchApp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	driverExt, err := ums.setupXTDriver(ctx, request.Params.Arguments)
	if err != nil {
		return nil, err
	}
	var appLaunchReq types.AppLaunchRequest
	if err := mapToStruct(request.Params.Arguments, &appLaunchReq); err != nil {
		return mcp.NewToolResultError("parse parameters error: " + err.Error()), nil
	}
	packageName := appLaunchReq.PackageName
	if packageName == "" {
		return mcp.NewToolResultError("package_name is required"), nil
	}
	err = driverExt.AppLaunch(packageName)
	if err != nil {
		return mcp.NewToolResultError("Launch app failed: " + err.Error()), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("Launched app success: %s", packageName)), nil
}

// handleTerminateApp handles the terminate_app tool call.
func (ums *MCPServer4XTDriver) handleTerminateApp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	driverExt, err := ums.setupXTDriver(ctx, request.Params.Arguments)
	if err != nil {
		return nil, err
	}
	var appTerminateReq types.AppTerminateRequest
	if err := mapToStruct(request.Params.Arguments, &appTerminateReq); err != nil {
		return mcp.NewToolResultError("parse parameters error: " + err.Error()), nil
	}
	packageName := appTerminateReq.PackageName
	if packageName == "" {
		return mcp.NewToolResultError("package_name is required"), nil
	}
	_, err = driverExt.AppTerminate(packageName)
	if err != nil {
		return mcp.NewToolResultError("Terminate app failed: " + err.Error()), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("Terminated app success: %s", packageName)), nil
}

// handleGetScreenSize handles the get_screen_size tool call.
func (ums *MCPServer4XTDriver) handleGetScreenSize(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	driverExt, err := ums.setupXTDriver(ctx, request.Params.Arguments)
	if err != nil {
		return nil, err
	}
	screenSize, err := driverExt.IDriver.WindowSize()
	if err != nil {
		return mcp.NewToolResultError("Get screen size failed: " + err.Error()), nil
	}
	return mcp.NewToolResultText(
		fmt.Sprintf("Screen size: %d x %d pixels", screenSize.Width, screenSize.Height),
	), nil
}

// handlePressButton handles the press_button tool call.
func (ums *MCPServer4XTDriver) handlePressButton(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	driverExt, err := ums.setupXTDriver(ctx, request.Params.Arguments)
	if err != nil {
		return nil, err
	}
	var pressButtonReq types.PressButtonRequest
	if err := mapToStruct(request.Params.Arguments, &pressButtonReq); err != nil {
		return mcp.NewToolResultError("parse parameters error: " + err.Error()), nil
	}
	err = driverExt.PressButton(pressButtonReq.Button)
	if err != nil {
		return mcp.NewToolResultError("Press button failed: " + err.Error()), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("Pressed button: %s", pressButtonReq.Button)), nil
}

// handleTapXY handles the tap_xy tool call.
func (ums *MCPServer4XTDriver) handleTapXY(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	driverExt, err := ums.setupXTDriver(ctx, request.Params.Arguments)
	if err != nil {
		return nil, err
	}
	var tapReq types.TapRequest
	if err := mapToStruct(request.Params.Arguments, &tapReq); err != nil {
		return mcp.NewToolResultError("parse parameters error: " + err.Error()), nil
	}
	if tapReq.Duration > 0 {
		err := driverExt.Drag(tapReq.X, tapReq.Y, tapReq.X, tapReq.Y,
			option.WithDuration(tapReq.Duration))
		if err != nil {
			return mcp.NewToolResultError("Tap failed: " + err.Error()), nil
		}
	} else {
		err := driverExt.TapXY(tapReq.X, tapReq.Y)
		if err != nil {
			return mcp.NewToolResultError("Tap failed: " + err.Error()), nil
		}
	}
	return mcp.NewToolResultText(
		fmt.Sprintf("tap (%f,%f) success", tapReq.X, tapReq.Y),
	), nil
}

// handleSwipe handles the swipe tool call.
func (ums *MCPServer4XTDriver) handleSwipe(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	driverExt, err := ums.setupXTDriver(ctx, request.Params.Arguments)
	if err != nil {
		return nil, err
	}
	var swipeReq types.SwipeRequest
	if err := mapToStruct(request.Params.Arguments, &swipeReq); err != nil {
		return mcp.NewToolResultError("parse parameters error: " + err.Error()), nil
	}

	// enum direction: up, down, left, right
	switch swipeReq.Direction {
	case "up":
		err = driverExt.Swipe(0.5, 0.5, 0.5, 0.1)
	case "down":
		err = driverExt.Swipe(0.5, 0.5, 0.5, 0.9)
	case "left":
		err = driverExt.Swipe(0.5, 0.5, 0.1, 0.5)
	case "right":
		err = driverExt.Swipe(0.5, 0.5, 0.9, 0.5)
	default:
		return mcp.NewToolResultError(fmt.Sprintf("get unexpected swipe direction: %s", swipeReq.Direction)), nil
	}
	if err != nil {
		return mcp.NewToolResultError("Swipe failed: " + err.Error()), nil
	}
	return mcp.NewToolResultText(
		fmt.Sprintf("swipe %s success", swipeReq.Direction),
	), nil
}

// handleDrag handles the drag tool call.
func (ums *MCPServer4XTDriver) handleDrag(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	driverExt, err := ums.setupXTDriver(ctx, request.Params.Arguments)
	if err != nil {
		return nil, err
	}
	var dragReq types.DragRequest
	if err := mapToStruct(request.Params.Arguments, &dragReq); err != nil {
		return mcp.NewToolResultError("parse parameters error: " + err.Error()), nil
	}
	actionOptions := []option.ActionOption{}
	if dragReq.Duration > 0 {
		actionOptions = append(actionOptions, option.WithDuration(dragReq.Duration/1000.0))
	}
	err = driverExt.Swipe(dragReq.FromX, dragReq.FromY,
		dragReq.ToX, dragReq.ToY, actionOptions...)
	if err != nil {
		return mcp.NewToolResultError("Swipe failed: " + err.Error()), nil
	}
	return mcp.NewToolResultText(
		fmt.Sprintf("swipe (%f,%f)->(%f,%f) success",
			dragReq.FromX, dragReq.FromY, dragReq.ToX, dragReq.ToY),
	), nil
}

// handleScreenShot handles the screenshot tool call.
func (ums *MCPServer4XTDriver) handleScreenShot(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Info().Msg("take screenshot")
	driverExt, err := ums.setupXTDriver(ctx, request.Params.Arguments)
	if err != nil {
		return nil, err
	}

	bufferBase64, err := uixt.GetScreenShotBufferBase64(driverExt.IDriver)
	if err != nil {
		log.Error().Err(err).Msg("ScreenShot failed")
		return mcp.NewToolResultError(fmt.Sprintf("Failed to take screenshot: %v", err)), nil
	}
	log.Debug().Int("imageBytes", len(bufferBase64)).Msg("take screenshot success")

	return mcp.NewToolResultImage("screenshot", bufferBase64, "image/jpeg"), nil
}

// setupXTDriver initializes an XTDriver based on the platform and serial.
func (ums *MCPServer4XTDriver) setupXTDriver(_ context.Context, args map[string]interface{}) (*uixt.XTDriver, error) {
	platform, _ := args["platform"].(string)
	serial, _ := args["serial"].(string)
	if platform == "" {
		log.Warn().Msg("platform is not set, using android as default")
		platform = "android"
	}

	// Check if driver exists in cache
	cacheKey := fmt.Sprintf("%s_%s", platform, serial)
	if cachedDriver, ok := ums.driverCache.Load(cacheKey); ok {
		if driverExt, ok := cachedDriver.(*uixt.XTDriver); ok {
			log.Info().Str("platform", platform).Str("serial", serial).Msg("Using cached driver")
			return driverExt, nil
		}
	}

	driverExt, err := initDriverExt(platform, serial)
	if err != nil {
		return nil, err
	}
	// store driver in cache
	ums.driverCache.Store(cacheKey, driverExt)
	return driverExt, nil
}

func initDriverExt(platform, serial string) (*uixt.XTDriver, error) {
	// init device
	var device uixt.IDevice
	var err error
	switch strings.ToLower(platform) {
	case "android":
		device, err = uixt.NewAndroidDevice(option.WithSerialNumber(serial))
	case "ios":
		device, err = uixt.NewIOSDevice(
			option.WithUDID(serial),
			option.WithWDAPort(8700),
			option.WithWDAMjpegPort(8800),
			option.WithResetHomeOnStartup(false),
		)
	case "browser":
		device, err = uixt.NewBrowserDevice(option.WithBrowserID(serial))
	default:
		return nil, fmt.Errorf("invalid platform: %s", platform)
	}
	if err != nil {
		return nil, fmt.Errorf("init device failed: %w", err)
	}
	if err := device.Setup(); err != nil {
		return nil, fmt.Errorf("setup device failed: %w", err)
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
	driverExt, err := uixt.NewXTDriver(driver,
		option.WithCVService(option.CVServiceTypeVEDEM))
	if err != nil {
		return nil, fmt.Errorf("init XT driver failed: %w", err)
	}
	return driverExt, nil
}

// generateMCPOptions generates mcp.NewTool parameters from a struct type.
// It automatically generates mcp.NewTool parameters based on the struct fields and their desc tags.
func generateMCPOptions(t interface{}) (options []mcp.ToolOption) {
	tType := reflect.TypeOf(t)
	for i := 0; i < tType.NumField(); i++ {
		field := tType.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}
		name := strings.Split(jsonTag, ",")[0]
		binding := field.Tag.Get("binding")
		required := strings.Contains(binding, "required")
		desc := field.Tag.Get("desc")
		switch field.Type.Kind() {
		case reflect.Float64, reflect.Float32, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if required {
				options = append(options, mcp.WithNumber(name, mcp.Required(), mcp.Description(desc)))
			} else {
				options = append(options, mcp.WithNumber(name, mcp.Description(desc)))
			}
		case reflect.String:
			if required {
				options = append(options, mcp.WithString(name, mcp.Required(), mcp.Description(desc)))
			} else {
				options = append(options, mcp.WithString(name, mcp.Description(desc)))
			}
		case reflect.Bool:
			if required {
				options = append(options, mcp.WithBoolean(name, mcp.Required(), mcp.Description(desc)))
			} else {
				options = append(options, mcp.WithBoolean(name, mcp.Description(desc)))
			}
		default:
			log.Warn().Str("field_type", field.Type.String()).Msg("Unsupported field type")
		}
	}
	return options
}

// mapToStruct convert map[string]interface{} to target struct
func mapToStruct(m map[string]interface{}, out interface{}) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}

// commonToolOptions is the common tool options for all tools.
var commonToolOptions = []mcp.ToolOption{
	mcp.WithString("platform", mcp.Required(), mcp.Description("Device platform: android/ios/browser")),
	mcp.WithString("serial", mcp.Required(), mcp.Description("Device serial/udid/browser id")),
}

// ListTools returns all registered tools
func (s *MCPServer4XTDriver) ListTools() []mcp.Tool {
	return s.tools
}

// GetTool returns a pointer to the mcp.Tool with the given name
func (s *MCPServer4XTDriver) GetTool(name string) *mcp.Tool {
	for i := range s.tools {
		if s.tools[i].Name == name {
			return &s.tools[i]
		}
	}
	return nil
}

// GetHandler returns the tool handler for the given name
func (s *MCPServer4XTDriver) GetHandler(name string) toolCall {
	if s.handlerMap == nil {
		return nil
	}
	return s.handlerMap[name]
}
