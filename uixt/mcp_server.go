package uixt

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/internal/version"
	"github.com/httprunner/httprunner/v5/pkg/gadb"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/httprunner/httprunner/v5/uixt/types"
)

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
	s.registerTools()
	return s
}

type toolCall = func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error)

// MCPServer4XTDriver wraps a MCPServer to expose XTDriver functionality via MCP protocol.
type MCPServer4XTDriver struct {
	mcpServer  *server.MCPServer
	tools      []mcp.Tool          // tools list for uixt
	handlerMap map[string]toolCall // tool name to handler
}

// Start runs the MCP server (blocking).
func (s *MCPServer4XTDriver) Start() error {
	log.Info().Msg("Starting HttpRunner UIXT MCP Server...")
	return server.ServeStdio(s.mcpServer)
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

// registerTools registers all MCP tools.
func (ums *MCPServer4XTDriver) registerTools() {
	// ListAvailableDevices Tool
	ums.registerTool(&ToolListAvailableDevices{})

	// SelectDevice Tool
	ums.registerTool(&ToolSelectDevice{})

	// ListPackages Tool
	ums.registerTool(&ToolListPackages{})

	// LaunchApp Tool
	ums.registerTool(&ToolLaunchApp{})

	// TerminateApp Tool
	ums.registerTool(&ToolTerminateApp{})

	// GetScreenSize Tool
	ums.registerTool(&ToolGetScreenSize{})

	// PressButton Tool
	ums.registerTool(&ToolPressButton{})

	// TapXY Tool
	ums.registerTool(&ToolTapXY{})

	// Swipe Tool
	ums.registerTool(&ToolSwipe{})

	// Drag Tool
	ums.registerTool(&ToolDrag{})

	// ScreenShot Tool
	ums.registerTool(&ToolScreenShot{})
}

func (ums *MCPServer4XTDriver) registerTool(tool ActionTool) {
	options := []mcp.ToolOption{
		mcp.WithDescription(tool.Description()),
	}
	options = append(options, tool.Options()...)
	mcpTool := mcp.NewTool(tool.Name(), options...)
	ums.mcpServer.AddTool(mcpTool, tool.Implement())
	ums.tools = append(ums.tools, mcpTool)
	ums.handlerMap[tool.Name()] = tool.Implement()
	log.Debug().Str("name", tool.Name()).Msg("register tool")
}

type ActionTool interface {
	Name() string
	Description() string
	Options() []mcp.ToolOption
	Implement() toolCall
}

// ToolListAvailableDevices implements the list_available_devices tool call.
type ToolListAvailableDevices struct{}

func (t *ToolListAvailableDevices) Name() string {
	return "list_available_devices"
}

func (t *ToolListAvailableDevices) Description() string {
	return "List all available devices. If there are more than one device returned, you need to let the user select one of them."
}

func (t *ToolListAvailableDevices) Options() []mcp.ToolOption {
	return []mcp.ToolOption{}
}

func (t *ToolListAvailableDevices) Implement() toolCall {
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

// ToolSelectDevice implements the select_device tool call.
type ToolSelectDevice struct{}

func (t *ToolSelectDevice) Name() string {
	return "select_device"
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

func (t *ToolSelectDevice) Implement() toolCall {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		uuid := driverExt.IDriver.GetDevice().UUID()
		return mcp.NewToolResultText(fmt.Sprintf("Selected device: %s", uuid)), nil
	}
}

// ToolListPackages implements the list_packages tool call.
type ToolListPackages struct{}

func (t *ToolListPackages) Name() string {
	return "list_packages"
}

func (t *ToolListPackages) Description() string {
	return "List all the apps/packages on the device."
}

func (t *ToolListPackages) Options() []mcp.ToolOption {
	return generateMCPOptions(&types.TargetDeviceRequest{})
}

func (t *ToolListPackages) Implement() toolCall {
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

// ToolLaunchApp implements the launch_app tool call.
type ToolLaunchApp struct{}

func (t *ToolLaunchApp) Name() string {
	return "launch_app"
}

func (t *ToolLaunchApp) Description() string {
	return "Launch an app on mobile device. Use this to open a specific app. You can find the package name of the app by calling list_packages."
}

func (t *ToolLaunchApp) Options() []mcp.ToolOption {
	return generateMCPOptions(&types.AppLaunchRequest{})
}

func (t *ToolLaunchApp) Implement() toolCall {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
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
}

// ToolTerminateApp implements the terminate_app tool call.
type ToolTerminateApp struct{}

func (t *ToolTerminateApp) Name() string {
	return "terminate_app"
}

func (t *ToolTerminateApp) Description() string {
	return "Stop and terminate an app on mobile device"
}

func (t *ToolTerminateApp) Options() []mcp.ToolOption {
	return generateMCPOptions(&types.AppTerminateRequest{})
}

func (t *ToolTerminateApp) Implement() toolCall {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
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
}

// ToolGetScreenSize implements the get_screen_size tool call.
type ToolGetScreenSize struct{}

func (t *ToolGetScreenSize) Name() string {
	return "get_screen_size"
}

func (t *ToolGetScreenSize) Description() string {
	return "Get the screen size of the mobile device in pixels"
}

func (t *ToolGetScreenSize) Options() []mcp.ToolOption {
	return generateMCPOptions(&types.TargetDeviceRequest{})
}

func (t *ToolGetScreenSize) Implement() toolCall {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
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
}

// ToolPressButton implements the press_button tool call.
type ToolPressButton struct{}

func (t *ToolPressButton) Name() string {
	return "press_button"
}

func (t *ToolPressButton) Description() string {
	return "Press a button on the device"
}

func (t *ToolPressButton) Options() []mcp.ToolOption {
	return generateMCPOptions(&types.PressButtonRequest{})
}

func (t *ToolPressButton) Implement() toolCall {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
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
}

// ToolTapXY implements the tap_xy tool call.
type ToolTapXY struct{}

func (t *ToolTapXY) Name() string {
	return "tap_xy"
}

func (t *ToolTapXY) Description() string {
	return "Click on the screen at given x,y coordinates"
}

func (t *ToolTapXY) Options() []mcp.ToolOption {
	return generateMCPOptions(&types.TapRequest{})
}

func (t *ToolTapXY) Implement() toolCall {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return mcp.NewToolResultError("Tap failed: " + err.Error()), nil
		}
		var tapReq types.TapRequest
		if err := mapToStruct(request.Params.Arguments, &tapReq); err != nil {
			return mcp.NewToolResultError("parse parameters error: " + err.Error()), nil
		}
		err = driverExt.TapXY(tapReq.X, tapReq.Y,
			option.WithDuration(tapReq.Duration),
			option.WithPreMarkOperation(true))
		if err != nil {
			return mcp.NewToolResultError("Tap failed: " + err.Error()), nil
		}
		return mcp.NewToolResultText(
			fmt.Sprintf("tap (%f,%f) success", tapReq.X, tapReq.Y),
		), nil
	}
}

// ToolSwipe implements the swipe tool call.
type ToolSwipe struct{}

func (t *ToolSwipe) Name() string {
	return "swipe"
}

func (t *ToolSwipe) Description() string {
	return "Swipe on the screen"
}

func (t *ToolSwipe) Options() []mcp.ToolOption {
	return generateMCPOptions(&types.SwipeRequest{})
}

func (t *ToolSwipe) Implement() toolCall {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return mcp.NewToolResultError("Swipe failed: " + err.Error()), nil
		}
		var swipeReq types.SwipeRequest
		if err := mapToStruct(request.Params.Arguments, &swipeReq); err != nil {
			return mcp.NewToolResultError("parse parameters error: " + err.Error()), nil
		}

		options := []option.ActionOption{
			option.WithPreMarkOperation(true),
			option.WithDuration(swipeReq.Duration),
			option.WithPressDuration(swipeReq.PressDuration),
		}

		// enum direction: up, down, left, right
		switch swipeReq.Direction {
		case "up":
			err = driverExt.Swipe(0.5, 0.5, 0.5, 0.1, options...)
		case "down":
			err = driverExt.Swipe(0.5, 0.5, 0.5, 0.9, options...)
		case "left":
			err = driverExt.Swipe(0.5, 0.5, 0.1, 0.5, options...)
		case "right":
			err = driverExt.Swipe(0.5, 0.5, 0.9, 0.5, options...)
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
}

// ToolDrag implements the drag tool call.
type ToolDrag struct{}

func (t *ToolDrag) Name() string {
	return "drag"
}

func (t *ToolDrag) Description() string {
	return "Drag on the mobile device"
}

func (t *ToolDrag) Options() []mcp.ToolOption {
	return generateMCPOptions(&types.DragRequest{})
}

func (t *ToolDrag) Implement() toolCall {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
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
}

// ToolScreenShot implements the screenshot tool call.
type ToolScreenShot struct{}

func (t *ToolScreenShot) Name() string {
	return "screenshot"
}

func (t *ToolScreenShot) Description() string {
	return "Take a screenshot of the mobile device. Use this to understand what's on screen. Do not cache this result."
}

func (t *ToolScreenShot) Options() []mcp.ToolOption {
	return generateMCPOptions(&types.TargetDeviceRequest{})
}

func (t *ToolScreenShot) Implement() toolCall {
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

var driverCache sync.Map // key is serial, value is *XTDriver

// setupXTDriver initializes an XTDriver based on the platform and serial.
func setupXTDriver(_ context.Context, args map[string]interface{}) (*XTDriver, error) {
	platform, _ := args["platform"].(string)
	serial, _ := args["serial"].(string)
	if platform == "" {
		log.Warn().Msg("platform is not set, using android as default")
		platform = "android"
	}

	// Check if driver exists in cache
	cacheKey := fmt.Sprintf("%s_%s", platform, serial)
	if cachedDriver, ok := driverCache.Load(cacheKey); ok {
		if driverExt, ok := cachedDriver.(*XTDriver); ok {
			log.Info().Str("platform", platform).Str("serial", serial).Msg("Using cached driver")
			return driverExt, nil
		}
	}

	driverExt, err := NewDriverExt(platform, serial)
	if err != nil {
		return nil, err
	}
	// store driver in cache
	driverCache.Store(cacheKey, driverExt)
	return driverExt, nil
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
