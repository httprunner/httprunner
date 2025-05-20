package mcphost

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/httprunner/httprunner/v5/internal/version"
	"github.com/httprunner/httprunner/v5/uixt"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/httprunner/httprunner/v5/uixt/types"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rs/zerolog/log"
)

// MCPServer4XTDriver wraps a MCPServer to expose XTDriver functionality via MCP protocol.
type MCPServer4XTDriver struct {
	mcpServer   *server.MCPServer
	driverCache sync.Map   // key is serial, value is *XTDriver
	tools       []mcp.Tool // 本地维护的工具列表
}

// NewMCPServer creates a new MCP server for XTDriver and registers all tools.
func NewMCPServer() *MCPServer4XTDriver {
	mcpServer := server.NewMCPServer(
		"uixt",
		version.GetVersionInfo(),
		server.WithToolCapabilities(false),
	)
	s := &MCPServer4XTDriver{
		mcpServer: mcpServer,
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
	// TapXY Tool
	tapParams := append(
		[]mcp.ToolOption{mcp.WithDescription("Taps on the device screen at the given coordinates.")},
		commonToolOptions...,
	)
	tapParams = append(tapParams, generateMCPOptions(types.TapRequest{})...)
	tapXYTool := mcp.NewTool("tap_xy", tapParams...)
	ums.mcpServer.AddTool(tapXYTool, ums.handleTapXY)
	ums.tools = append(ums.tools, tapXYTool)
	log.Info().Str("name", tapXYTool.Name).Msg("Register tool")

	// Swipe Tool
	swipeParams := append(
		[]mcp.ToolOption{mcp.WithDescription("Swipes on the device screen from one point to another.")},
		commonToolOptions...,
	)
	swipeParams = append(swipeParams, generateMCPOptions(types.DragRequest{})...)
	swipeTool := mcp.NewTool("swipe", swipeParams...)
	ums.mcpServer.AddTool(swipeTool, ums.handleSwipe)
	ums.tools = append(ums.tools, swipeTool)
	log.Info().Str("name", swipeTool.Name).Msg("Register tool")

	// ScreenShot Tool
	screenShotTool := mcp.NewTool("screenshot",
		mcp.WithDescription("Takes a screenshot of the device screen and returns it as a base64 encoded string."),
	)
	ums.mcpServer.AddTool(screenShotTool, ums.handleScreenShot)
	ums.tools = append(ums.tools, screenShotTool)
	log.Info().Str("name", screenShotTool.Name).Msg("Register tool")
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
		err := driverExt.Drag(tapReq.X, tapReq.Y, tapReq.X, tapReq.Y, option.WithDuration(tapReq.Duration))
		if err != nil {
			return mcp.NewToolResultError("Tap failed: " + err.Error()), nil
		}
	} else {
		err := driverExt.TapXY(tapReq.X, tapReq.Y)
		if err != nil {
			return mcp.NewToolResultError("Tap failed: " + err.Error()), nil
		}
	}
	return mcp.NewToolResultText("Tap successful."), nil
}

// handleSwipe handles the swipe tool call.
func (ums *MCPServer4XTDriver) handleSwipe(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	driverExt, err := ums.setupXTDriver(ctx, request.Params.Arguments)
	if err != nil {
		return nil, err
	}
	var swipeReq types.DragRequest
	if err := mapToStruct(request.Params.Arguments, &swipeReq); err != nil {
		return mcp.NewToolResultError("parse parameters error: " + err.Error()), nil
	}
	actionOptions := []option.ActionOption{}
	if swipeReq.Duration > 0 {
		actionOptions = append(actionOptions, option.WithDuration(swipeReq.Duration/1000.0))
	}
	err = driverExt.Swipe(swipeReq.FromX, swipeReq.FromY, swipeReq.ToX, swipeReq.ToY, actionOptions...)
	if err != nil {
		return mcp.NewToolResultError("Swipe failed: " + err.Error()), nil
	}
	return mcp.NewToolResultText("Swipe successful."), nil
}

// handleScreenShot handles the screenshot tool call.
func (ums *MCPServer4XTDriver) handleScreenShot(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Info().Msg("Executing ScreenShot")
	driverExt, err := ums.setupXTDriver(ctx, request.Params.Arguments)
	if err != nil {
		return nil, err
	}
	buffer, err := driverExt.ScreenShot()
	if err != nil {
		log.Error().Err(err).Msg("ScreenShot failed")
		return mcp.NewToolResultError(fmt.Sprintf("Failed to take screenshot: %v", err)), nil
	}
	if buffer == nil || buffer.Len() == 0 {
		log.Error().Msg("Screenshot buffer is nil or empty")
		return mcp.NewToolResultError("Screenshot returned empty buffer"), nil
	}
	encodedString := base64.StdEncoding.EncodeToString(buffer.Bytes())
	log.Info().Int("image_size_bytes", len(buffer.Bytes())).Int("base64_len", len(encodedString)).Msg("Screenshot successful")
	return mcp.NewToolResultText(encodedString), nil
}

// setupXTDriver initializes an XTDriver based on the platform and serial.
func (ums *MCPServer4XTDriver) setupXTDriver(_ context.Context, args map[string]interface{}) (*uixt.XTDriver, error) {
	platform, _ := args["platform"].(string)
	serial, _ := args["serial"].(string)
	if platform == "" || serial == "" {
		return nil, fmt.Errorf("platform and serial are required")
	}

	// Check if driver exists in cache
	cacheKey := fmt.Sprintf("%s_%s", platform, serial)
	if cachedDriver, ok := ums.driverCache.Load(cacheKey); ok {
		if driverExt, ok := cachedDriver.(*uixt.XTDriver); ok {
			log.Info().Str("platform", platform).Str("serial", serial).Msg("Using cached driver")
			return driverExt, nil
		}
	}

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

// ListTools 返回所有注册的 mcp.Tool
func (s *MCPServer4XTDriver) ListTools() []mcp.Tool {
	return s.tools
}

// GetTool 根据名称返回 mcp.Tool 指针
func (s *MCPServer4XTDriver) GetTool(name string) *mcp.Tool {
	for i := range s.tools {
		if s.tools[i].Name == name {
			return &s.tools[i]
		}
	}
	return nil
}
