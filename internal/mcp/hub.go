package mcp

import (
	"context"
	"fmt"
	"sync"

	mcpp "github.com/cloudwego/eino-ext/components/tool/mcp"
	"github.com/cloudwego/eino/components/tool"
	"github.com/httprunner/httprunner/v5/internal/version"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type MCPTools struct {
	Name  string
	Tools []mcp.Tool
	Err   error
}

type MCPHub struct {
	mu          sync.RWMutex
	connections map[string]*Connection
	config      *MCPSettings
}

type Connection struct {
	Client client.MCPClient
	Config ServerConfig
}

func NewMCPHub(configPath string) (*MCPHub, error) {
	settings, err := LoadSettings(configPath)
	if err != nil {
		return nil, err
	}
	return &MCPHub{
		connections: make(map[string]*Connection),
		config:      settings,
	}, nil
}

// InitServers initializes all enabled MCP servers
func (h *MCPHub) InitServers(ctx context.Context) error {
	for name, config := range h.config.MCPServers {
		if config.Disabled {
			continue
		}

		if err := h.connectToServer(ctx, name, config); err != nil {
			return fmt.Errorf("failed to connect to server %s: %w", name, err)
		}
	}

	return nil
}

// GetClient returns the client for the specified server
func (h *MCPHub) GetClient(serverName string) (client.MCPClient, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	conn, exists := h.connections[serverName]
	if !exists {
		return nil, fmt.Errorf("no connection found for server %s", serverName)
	}

	if conn.Config.Disabled {
		return nil, fmt.Errorf("server %s is disabled", serverName)
	}

	return conn.Client, nil
}

// connectToServer establishes connection to a single MCP server
func (h *MCPHub) connectToServer(ctx context.Context, serverName string, config ServerConfig) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Close existing connection if any
	if existing, exists := h.connections[serverName]; exists {
		if err := existing.Client.Close(); err != nil {
			return fmt.Errorf("failed to close existing connection: %w", err)
		}
		delete(h.connections, serverName)
	}

	var mcpClient client.MCPClient
	var err error

	// create client
	switch config.TransportType {
	case "sse":
		mcpClient, err = client.NewSSEMCPClient(config.URL)
	case "stdio", "": // default to stdio
		var env []string
		for k, v := range config.Env {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
		mcpClient, err = client.NewStdioMCPClient(config.Command,
			env, config.Args...)
	default:
		return fmt.Errorf("unsupported transport type: %s", config.TransportType)
	}
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// prepare client init request
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.Capabilities = mcp.ClientCapabilities{}
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "HttpRunner",
		Version: version.VERSION,
	}

	// initialize client
	_, err = mcpClient.Initialize(ctx, initRequest)
	if err != nil {
		mcpClient.Close()
		return errors.Wrapf(err, "initialize MCP client for %s failed", serverName)
	}

	log.Info().Str("server", serverName).Msg("connected to MCP server")
	h.connections[serverName] = &Connection{
		Client: mcpClient,
		Config: config,
	}
	return nil
}

// GetTools fetches available tools from all connected MCP servers
func (h *MCPHub) GetTools(ctx context.Context) map[string]MCPTools {
	h.mu.RLock()
	defer h.mu.RUnlock()

	results := make(map[string]MCPTools)

	for serverName, conn := range h.connections {
		if conn.Config.Disabled {
			continue
		}

		// get tools from MCP server tools
		listResults, err := conn.Client.ListTools(ctx, mcp.ListToolsRequest{})
		if err != nil {
			results[serverName] = MCPTools{
				Name:  serverName,
				Tools: nil,
				Err:   fmt.Errorf("failed to get tools: %w", err),
			}
			continue
		}

		results[serverName] = MCPTools{
			Name:  serverName,
			Tools: listResults.Tools,
			Err:   nil,
		}
	}

	return results
}

func (h *MCPHub) GetTool(ctx context.Context, serverName, toolName string) (*mcp.Tool, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// filter MCP server by serverName
	mcpTools, exists := h.GetTools(ctx)[serverName]
	if !exists {
		return nil, fmt.Errorf("no connection found for server %s", serverName)
	} else if mcpTools.Err != nil {
		return nil, mcpTools.Err
	}

	// filter tool by toolName
	for _, tool := range mcpTools.Tools {
		if tool.Name == toolName {
			return &tool, nil
		}
	}

	return nil, fmt.Errorf("tool %s not found", toolName)
}

// InvokeTool calls a tool with the given arguments
func (h *MCPHub) InvokeTool(ctx context.Context,
	serverName, toolName string, arguments map[string]interface{},
) (*mcp.CallToolResult, error) {
	conn, err := h.GetClient(serverName)
	if err != nil {
		return nil, errors.Wrapf(err,
			"get mcp client for server %s failed", serverName)
	}

	mcpTool, err := h.GetTool(ctx, serverName, toolName)
	if err != nil {
		return nil, errors.Wrapf(err,
			"get mcp tool %s/%s failed", serverName, toolName)
	}

	req := mcp.CallToolRequest{}
	req.Params.Name = mcpTool.Name
	req.Params.Arguments = arguments
	callToolResult, err := conn.CallTool(ctx, req)
	if err != nil {
		return nil, errors.Wrapf(err,
			"call tool %s/%s failed", serverName, toolName)
	}

	return callToolResult, nil
}

// GetEinoTool returns an eino tool from the MCP server
func (h *MCPHub) GetEinoTool(ctx context.Context, serverName, toolName string) (tool.BaseTool, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// filter MCP server by serverName
	conn, exists := h.connections[serverName]
	if !exists {
		return nil, fmt.Errorf("no connection found for server %s", serverName)
	}

	if conn.Config.Disabled {
		return nil, fmt.Errorf("server %s is disabled", serverName)
	}

	// get tools from MCP server and convert to eino tools
	tools, err := mcpp.GetTools(ctx, &mcpp.Config{
		Cli:          conn.Client,
		ToolNameList: []string{toolName},
	})
	if err != nil || len(tools) == 0 {
		log.Error().Err(err).
			Str("server", serverName).Str("tool", toolName).
			Msg("get MCP tool failed")
		return nil, err
	}

	return tools[0], nil
}

// CloseServers closes all connected MCP servers
func (h *MCPHub) CloseServers() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	log.Info().Msg("Shutting down MCP servers...")
	for name, client := range h.connections {
		if err := client.Client.Close(); err != nil {
			log.Error().Str("name", name).Err(err).Msg("Failed to close server")
		} else {
			delete(h.connections, name)
			log.Info().Str("name", name).Msg("Server closed")
		}
	}

	return nil
}
