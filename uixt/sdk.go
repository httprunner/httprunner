package uixt

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/uixt/ai"
	"github.com/httprunner/httprunner/v5/uixt/option"
)

func NewXTDriver(driver IDriver, opts ...option.AIServiceOption) (*XTDriver, error) {
	services := option.NewAIServiceOptions(opts...)
	driverExt := &XTDriver{
		IDriver: driver,
		client: &MCPClient4XTDriver{
			Server: NewMCPServer(),
		},
		services:         services,
		loadedMCPClients: make(map[string]client.MCPClient),
	}
	log.Info().Interface("services", services).Msg("init XTDriver with AI services")

	var err error

	// Handle LLM service initialization
	if services.LLMConfig != nil {
		// Use advanced LLM service configuration if provided
		driverExt.LLMService, err = ai.NewLLMServiceWithOptionConfig(services.LLMConfig)
		if err != nil {
			log.Warn().Err(err).Interface("service", services.LLMConfig).
				Msg("init llm service with advanced config failed")
		} else {
			log.Info().Interface("service", services.LLMConfig).
				Msg("LLM service initialized with advanced config")
		}
	} else if services.LLMService != "" {
		// Use simple LLM service configuration if provided
		driverExt.LLMService, err = ai.NewLLMService(services.LLMService)
		if err != nil {
			log.Warn().Err(err).Str("service", string(services.LLMService)).
				Msg("init llm service with simple config failed")
		} else {
			log.Info().Str("service", string(services.LLMService)).
				Msg("LLM service initialized with simple config")
		}
	} else {
		// Use Wings service as fallback
		driverExt.LLMService, err = ai.NewWingsService()
		if err != nil {
			log.Warn().Err(err).Msg("init Wings service failed")
		} else {
			log.Info().Msg("Wings service initialized as fallback")
		}
	}

	// Register uixt MCP tools to LLM service if it exists
	if driverExt.LLMService != nil {
		mcpTools := driverExt.client.Server.ListTools()
		einoTools := ai.ConvertMCPToolsToEinoToolInfos(mcpTools, "uixt")
		if err = driverExt.LLMService.RegisterTools(einoTools); err != nil {
			log.Warn().Err(err).Msg("failed to register uixt tools to LLM service")
		}
	}

	return driverExt, nil
}

// XTDriver = IDriver + AI
type XTDriver struct {
	IDriver
	CVService  ai.ICVService  // OCR/CV
	LLMService ai.ILLMService // LLM

	services         *option.AIServiceOptions    // AI services options
	client           *MCPClient4XTDriver         // MCP Client for built-in uixt server
	loadedMCPClients map[string]client.MCPClient // External MCP clients
}

func (dExt *XTDriver) initCVService() error {
	if dExt.CVService != nil {
		return nil
	}
	cvServiceType := dExt.services.CVService
	if cvServiceType == "" {
		log.Warn().Msg("no CV service config provided, use default vedem")
		cvServiceType = option.CVServiceTypeVEDEM
	}
	cvService, err := ai.NewCVService(cvServiceType)
	if err != nil {
		log.Error().Err(err).Str("type", string(cvServiceType)).
			Msg("init cv service failed")
		return errors.Wrap(err, "init cv service failed")
	}
	dExt.CVService = cvService
	return nil
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
	actionName := strings.TrimPrefix(req.Params.Name, "uixt__")
	actionTool := c.Server.GetToolByAction(option.ActionName(actionName))
	if actionTool == nil {
		return mcp.NewToolResultError(fmt.Sprintf("action %s for tool not found", actionName)), nil
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

// GetToolByAction implements ActionToolProvider interface
func (c *MCPClient4XTDriver) GetToolByAction(actionName option.ActionName) ActionTool {
	return c.Server.GetToolByAction(actionName)
}

func (dExt *XTDriver) ExecuteAction(ctx context.Context, action option.MobileAction) (SessionData, error) {
	// Find the corresponding tool for this action method
	tool := dExt.client.Server.GetToolByAction(action.Method)
	if tool == nil {
		return SessionData{}, fmt.Errorf("no tool found for action method: %s", action.Method)
	}

	// Use the tool's own conversion method
	req, err := tool.ConvertActionToCallToolRequest(action)
	if err != nil {
		return SessionData{}, fmt.Errorf("failed to convert action to MCP tool call: %w", err)
	}

	// Execute via MCP tool
	result, err := dExt.client.CallTool(ctx, req)
	if err != nil {
		// Notice: preserve the original error code
		return SessionData{}, errors.Wrap(err, "call MCP tool failed")
	}

	// Check if the tool execution had business logic errors
	if result.IsError {
		var errMsg string
		if len(result.Content) > 0 {
			errMsg = fmt.Sprintf("invoke tool %s failed: %v", tool.Name(), result.Content)
		} else {
			errMsg = fmt.Sprintf("invoke tool %s failed", tool.Name())
		}
		err := errors.New(errMsg)
		return SessionData{}, err
	}

	// For regular actions, collect session data and return it directly
	sessionData := dExt.GetSession().GetData(true) // reset after getting data

	// Log execution result, but avoid printing base64 data for screenshot tools
	logger := log.Debug().Str("tool", string(tool.Name()))
	if tool.Name() != option.ACTION_ScreenShot {
		logger.Interface("result", result.Content)
	}
	logger.Msg("executed action via MCP tool")

	return sessionData, nil
}

// NewDeviceWithDefault is a helper function to create a device with default options
func NewDeviceWithDefault(platform, serial string) (device IDevice, err error) {
	if serial == "" {
		return nil, fmt.Errorf("serial is empty")
	}

	switch strings.ToLower(platform) {
	case "android":
		device, err = NewAndroidDevice(option.WithSerialNumber(serial))
	case "ios":
		device, err = NewIOSDevice(
			option.WithUDID(serial),
			option.WithWDAPort(8700),
			option.WithWDAMjpegPort(8800),
			option.WithResetHomeOnStartup(false),
		)
	case "browser":
		device, err = NewBrowserDevice(option.WithBrowserID(serial))
	case "harmony":
		device, err = NewHarmonyDevice(option.WithConnectKey(serial))
	default:
		return nil, fmt.Errorf("unsupported platform: %s", platform)
	}

	return device, err
}

// SetMCPClients sets the external MCP clients for the driver
func (dExt *XTDriver) SetMCPClients(clients map[string]client.MCPClient) {
	if dExt.loadedMCPClients == nil {
		dExt.loadedMCPClients = make(map[string]client.MCPClient)
	}
	for name, client := range clients {
		dExt.loadedMCPClients[name] = client
	}
}

// GetMCPClient returns the MCP client for the specified server name
func (dExt *XTDriver) GetMCPClient(serverName string) (client.MCPClient, bool) {
	if dExt.loadedMCPClients == nil {
		return nil, false
	}
	client, exists := dExt.loadedMCPClients[serverName]
	return client, exists
}

// CallMCPTool calls the specified MCP tool
func (dExt *XTDriver) CallMCPTool(ctx context.Context,
	serverName, toolName string, arguments map[string]any,
) (result *mcp.CallToolResult, err error) {
	// Get MCP client

	mcpClient, exists := dExt.GetMCPClient(serverName)
	if !exists {
		log.Warn().Str("server", serverName).Msg("MCP server not found")
		return nil, fmt.Errorf("MCP server %s not found", serverName)
	}

	// Prepare arguments
	if arguments == nil {
		arguments = make(map[string]any)
	}

	log.Debug().Str("server", serverName).Str("tool", toolName).
		Interface("arguments", arguments).Msg("call MCP tool")

	// Call MCP tool
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      toolName,
			Arguments: arguments,
		},
	}

	result, err = mcpClient.CallTool(ctx, req)
	if err != nil {
		log.Debug().Err(err).
			Str("server", serverName).
			Str("tool", toolName).
			Msg("call MCP tool failed")
		return nil, errors.Wrap(err, "call MCP tool failed")
	}

	if result.IsError {
		logger := log.Debug().
			Str("server", serverName).
			Str("tool", toolName)

		// Avoid printing base64 data for screenshot tools
		if toolName != string(option.ACTION_ScreenShot) {
			logger.Interface("content", result.Content)
		}
		logger.Msg("call MCP tool failed")

		return nil, fmt.Errorf("call MCP tool %s failed", toolName)
	}

	log.Debug().
		Str("server", serverName).
		Str("tool", toolName).
		Msg("call MCP tool successfully")
	return result, nil
}
