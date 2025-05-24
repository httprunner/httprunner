package uixt

import (
	"context"
	"fmt"

	"github.com/httprunner/httprunner/v5/uixt/ai"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rs/zerolog/log"
)

func NewXTDriver(driver IDriver, opts ...option.AIServiceOption) (*XTDriver, error) {
	driverExt := &XTDriver{
		IDriver: driver,
		client: &MCPClient4XTDriver{
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

	client *MCPClient4XTDriver // MCP Client
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
	handler := c.Server.GetHandler(req.Params.Name)
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

func convertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	// req := mcp.CallToolRequest{
	// 	Params: struct {
	// 		Name      string         `json:"name"`
	// 		Arguments map[string]any `json:"arguments,omitempty"`
	// 		Meta      *struct {
	// 			ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
	// 		} `json:"_meta,omitempty"`
	// 	}{
	// 		Name:      action.Method,
	// 		Arguments: action.Params,
	// 	},
	// }
	return mcp.CallToolRequest{}, nil
}

func (dExt *XTDriver) ExecuteAction(action MobileAction) (err error) {
	// convert action to call tool request
	req, err := convertActionToCallToolRequest(action)
	if err != nil {
		return err
	}
	_, err = dExt.client.CallTool(context.Background(), req)
	return err
}
