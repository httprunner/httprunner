package uixt

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/httprunner/httprunner/v5/uixt/ai"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func NewXTDriver(driver IDriver, opts ...option.AIServiceOption) (*XTDriver, error) {
	driverExt := &XTDriver{
		IDriver: driver,
		client: &MCPClient4XTDriver{
			Server: NewMCPServer(),
		},
		loadedMCPClients: make(map[string]client.MCPClient),
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
			return nil, errors.Wrap(err, "init llm service failed")
		}

		// Register uixt MCP tools to LLM service
		mcpTools := driverExt.client.Server.ListTools()
		einoTools := ai.ConvertMCPToolsToEinoToolInfos(mcpTools, "uixt")
		if err := driverExt.LLMService.RegisterTools(einoTools); err != nil {
			log.Warn().Err(err).Msg("failed to register uixt tools")
		}
	}

	return driverExt, nil
}

// XTDriver = IDriver + AI
type XTDriver struct {
	IDriver
	CVService  ai.ICVService  // OCR/CV
	LLMService ai.ILLMService // LLM

	client           *MCPClient4XTDriver         // MCP Client for built-in uixt server
	loadedMCPClients map[string]client.MCPClient // External MCP clients
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

func (dExt *XTDriver) ExecuteAction(ctx context.Context, action option.MobileAction) ([]*SubActionResult, error) {
	subActionStartTime := time.Now()

	// Find the corresponding tool for this action method
	tool := dExt.client.Server.GetToolByAction(action.Method)
	if tool == nil {
		return nil, fmt.Errorf("no tool found for action method: %s", action.Method)
	}

	// Use the tool's own conversion method
	req, err := tool.ConvertActionToCallToolRequest(action)
	if err != nil {
		return nil, fmt.Errorf("failed to convert action to MCP tool call: %w", err)
	}

	// Create sub-action result
	subActionResult := &SubActionResult{
		ActionName: string(action.Method),
		Arguments:  action.Params,
		StartTime:  subActionStartTime.Unix(),
	}

	// Execute via MCP tool
	result, err := dExt.client.CallTool(ctx, req)
	subActionResult.Elapsed = time.Since(subActionStartTime).Milliseconds()
	if err != nil {
		subActionResult.Error = err
		return []*SubActionResult{subActionResult}, fmt.Errorf("MCP tool call failed: %w", err)
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
		subActionResult.Error = err
		return []*SubActionResult{subActionResult}, err
	}

	// For regular actions, collect session data and return single sub-action result
	subActionResult.SessionData = dExt.GetSession().GetData(true) // reset after getting data

	log.Debug().Str("tool", string(tool.Name())).
		Msg("execute action via MCP tool")
	return []*SubActionResult{subActionResult}, nil
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
	serverName, toolName string, arguments map[string]any) (result *mcp.CallToolResult, err error) {
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
		Params: struct {
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments,omitempty"`
			Meta      *struct {
				ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
			} `json:"_meta,omitempty"`
		}{
			Name:      toolName,
			Arguments: arguments,
		},
	}

	result, err = mcpClient.CallTool(ctx, req)
	if err != nil {
		log.Debug().Err(err).
			Str("server", serverName).
			Str("tool", toolName).
			Msg("MCP hook call failed")
		return nil, err
	}

	if result.IsError {
		log.Debug().
			Str("server", serverName).
			Str("tool", toolName).
			Interface("content", result.Content).
			Msg("MCP hook returned error")
		return nil, fmt.Errorf("MCP hook returned error")
	}

	log.Debug().
		Str("server", serverName).
		Str("tool", toolName).
		Msg("MCP hook called successfully")
	return result, nil
}
