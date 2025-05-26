package uixt

import (
	"context"
	"fmt"
	"strings"

	"github.com/httprunner/httprunner/v5/uixt/ai"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rs/zerolog/log"
)

func NewXTDriver(driver IDriver, opts ...option.AIServiceOption) (*XTDriver, error) {
	driverExt := &XTDriver{
		IDriver: driver,
		Client: &MCPClient4XTDriver{
			Server: NewMCPServer(),
		},
	}

	services := option.NewAIServiceOptions(opts...)

	var err error
	if services.CVService != "" {
		driverExt.CVService, err = ai.NewCVService(services.CVService)
		if err != nil {
			log.Error().Err(err).Msg("init vedem image service failed")
			return nil, err
		}
	}
	if services.LLMService != "" {
		driverExt.LLMService, err = ai.NewLLMService(services.LLMService)
		if err != nil {
			log.Error().Err(err).Msg("init llm service failed")
			return nil, err
		}
	}

	return driverExt, nil
}

// XTDriver = IDriver + AI
type XTDriver struct {
	IDriver
	CVService  ai.ICVService  // OCR/CV
	LLMService ai.ILLMService // LLM

	Client *MCPClient4XTDriver // MCP Client
}

// MCPClient4XTDriver is a mock MCP client that only implements the methods used by the host
type MCPClient4XTDriver struct {
	client.MCPClient
	Server *MCPServer4XTDriver
}

func (c *MCPClient4XTDriver) ListTools(ctx context.Context, req mcp.ListToolsRequest) (*mcp.ListToolsResult, error) {
	tools := c.Server.ListTools()
	return &mcp.ListToolsResult{Tools: tools}, nil
}

func (c *MCPClient4XTDriver) CallTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	actionTool := c.Server.GetToolByAction(option.ActionMethod(req.Params.Name))
	if actionTool == nil {
		return mcp.NewToolResultError(fmt.Sprintf("action %s for tool not found", req.Params.Name)), nil
	}
	handler := actionTool.Implement()
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

func (dExt *XTDriver) ExecuteAction(action MobileAction) (err error) {
	// Find the corresponding tool for this action method
	tool := dExt.Client.Server.GetToolByAction(action.Method)
	if tool == nil {
		return fmt.Errorf("no tool found for action method: %s", action.Method)
	}

	// Use the tool's own conversion method
	req, err := tool.ConvertActionToCallToolRequest(action)
	if err != nil {
		return fmt.Errorf("failed to convert action to MCP tool call: %w", err)
	}

	// Execute via MCP tool
	result, err := dExt.Client.CallTool(context.Background(), req)
	if err != nil {
		return fmt.Errorf("MCP tool call failed: %w", err)
	}

	// Check if the tool execution had business logic errors
	if result.IsError {
		if len(result.Content) > 0 {
			return fmt.Errorf("invoke tool %s failed: %v",
				tool.Name(), result.Content)
		}
		return fmt.Errorf("invoke tool %s failed", tool.Name())
	}

	log.Debug().Str("tool", string(tool.Name())).
		Msg("execute action via MCP tool")
	return nil
}

// NewXTDriverWithDefault is a helper function to create a XTDriver with default options
func NewXTDriverWithDefault(platform, serial string) (*XTDriver, error) {
	device, err := NewDeviceWithDefault(platform, serial)
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

// NewDeviceWithDefault is a helper function to create a device with default options
func NewDeviceWithDefault(platform, serial string) (device IDevice, err error) {
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

	return device, nil
}
