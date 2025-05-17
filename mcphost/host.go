package mcphost

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	mcpp "github.com/cloudwego/eino-ext/components/tool/mcp"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/httprunner/httprunner/v5/internal/version"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// MCPHost manages MCP server connections and tools
type MCPHost struct {
	mu          sync.RWMutex
	connections map[string]*Connection
	config      *MCPConfig
}

// Connection represents a connection to an MCP server
type Connection struct {
	Client client.MCPClient
	Config ServerConfig
}

// MCPTools represents tools from a single MCP server
type MCPTools struct {
	ServerName string
	Tools      []mcp.Tool
	Err        error
}

// NewMCPHost creates a new MCPHost instance
func NewMCPHost(configPath string) (*MCPHost, error) {
	config, err := LoadMCPConfig(configPath)
	if err != nil {
		return nil, err
	}

	host := &MCPHost{
		connections: make(map[string]*Connection),
		config:      config,
	}

	// Initialize MCP servers
	if err := host.InitServers(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to initialize MCP servers: %w", err)
	}

	return host, nil
}

// InitServers initializes all MCP servers
func (h *MCPHost) InitServers(ctx context.Context) error {
	for name, server := range h.config.MCPServers {
		if server.Config.IsDisabled() {
			continue
		}

		if err := h.connectToServer(ctx, name, server.Config); err != nil {
			return fmt.Errorf("failed to connect to server %s: %w", name, err)
		}
	}
	return nil
}

// connectToServer establishes connection to a single MCP server
func (h *MCPHost) connectToServer(ctx context.Context, serverName string, config ServerConfig) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	log.Debug().Str("server", serverName).Msg("connecting to MCP server")

	// Close existing connection if any
	if existing, exists := h.connections[serverName]; exists {
		if err := existing.Client.Close(); err != nil {
			return fmt.Errorf("failed to close existing connection: %w", err)
		}
		delete(h.connections, serverName)
	}

	var mcpClient client.MCPClient
	var err error

	// create client based on server type
	switch cfg := config.(type) {
	case SSEServerConfig:
		mcpClient, err = client.NewSSEMCPClient(cfg.Url,
			client.WithHeaders(parseHeaders(cfg.Headers)))
	case STDIOServerConfig:
		env := make([]string, 0, len(cfg.Env))
		for k, v := range cfg.Env {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
		mcpClient, err = client.NewStdioMCPClient(cfg.Command, env, cfg.Args...)
		if stdioClient, ok := mcpClient.(*client.Client); ok {
			stderr, _ := client.GetStderr(stdioClient)
			startStdioLog(stderr, serverName)
		}
	default:
		return fmt.Errorf("unsupported transport type: %s", config.GetType())
	}

	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// initialize client
	_, err = mcpClient.Initialize(ctx, prepareClientInitRequest())
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

// CloseServers closes all connected MCP servers
func (h *MCPHost) CloseServers() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	log.Info().Msg("Shutting down MCP servers...")
	for name, conn := range h.connections {
		if err := conn.Client.Close(); err != nil {
			log.Error().Str("name", name).Err(err).Msg("Failed to close server")
		} else {
			delete(h.connections, name)
			log.Info().Str("name", name).Msg("Server closed")
		}
	}

	return nil
}

// GetClient returns the client for the specified server
func (h *MCPHost) GetClient(serverName string) (client.MCPClient, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	conn, exists := h.connections[serverName]
	if !exists {
		return nil, fmt.Errorf("no connection found for server %s", serverName)
	}

	return conn.Client, nil
}

// GetTools returns all tools from all MCP servers
func (h *MCPHost) GetTools(ctx context.Context) []MCPTools {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var results []MCPTools

	for serverName, conn := range h.connections {
		listResults, err := conn.Client.ListTools(ctx, mcp.ListToolsRequest{})
		if err != nil {
			log.Error().Err(err).Str("server", serverName).Msg("failed to get tools")
			continue
		}

		results = append(results, MCPTools{
			ServerName: serverName,
			Tools:      listResults.Tools,
			Err:        nil,
		})
	}

	return results
}

// GetTool returns a specific tool from a server
func (h *MCPHost) GetTool(ctx context.Context, serverName, toolName string) (*mcp.Tool, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Get all tools
	results := h.GetTools(ctx)

	// Find the server's tools
	var serverTools MCPTools
	found := false
	for _, tools := range results {
		if tools.ServerName == serverName {
			serverTools = tools
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("no connection found for server %s", serverName)
	}
	if serverTools.Err != nil {
		return nil, serverTools.Err
	}

	// Find the specific tool
	for _, tool := range serverTools.Tools {
		if tool.Name == toolName {
			return &tool, nil
		}
	}

	return nil, fmt.Errorf("tool %s not found", toolName)
}

// InvokeTool calls a tool with the given arguments
func (h *MCPHost) InvokeTool(ctx context.Context,
	serverName, toolName string, arguments map[string]any,
) (*mcp.CallToolResult, error) {
	log.Info().Str("tool", toolName).Interface("args", arguments).
		Str("server", serverName).Msg("invoke tool")

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

	req := mcp.CallToolRequest{
		Params: struct {
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments,omitempty"`
			Meta      *struct {
				ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
			} `json:"_meta,omitempty"`
		}{
			Name:      mcpTool.Name,
			Arguments: arguments,
		},
	}

	result, err := conn.CallTool(ctx, req)
	if err != nil {
		return nil, errors.Wrapf(err,
			"call tool %s/%s failed", serverName, toolName)
	}

	if err := handleToolError(result); err != nil {
		return nil, err
	}

	return result, nil
}

// GetEinoTool returns an eino tool for the given server and tool name
func (h *MCPHost) GetEinoTool(ctx context.Context, serverName, toolName string) (tool.BaseTool, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	conn, ok := h.connections[serverName]
	if !ok {
		return nil, fmt.Errorf("server not found: %s", serverName)
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

// GetEinoToolInfos convert MCP tools to eino tool infos
func (h *MCPHost) GetEinoToolInfos(ctx context.Context) ([]*schema.ToolInfo, error) {
	results := h.GetTools(ctx)
	if len(results) == 0 {
		return nil, fmt.Errorf("no MCP servers loaded")
	}

	var tools []*schema.ToolInfo
	for _, serverTools := range results {
		if serverTools.Err != nil {
			log.Error().Err(serverTools.Err).Str("server", serverTools.ServerName).Msg("failed to get tools")
			continue
		}

		for _, tool := range serverTools.Tools {
			einoTool, err := h.GetEinoTool(ctx, serverTools.ServerName, tool.Name)
			if err != nil {
				log.Error().Err(err).Str("server", serverTools.ServerName).Str("tool", tool.Name).Msg("failed to get eino tool")
				continue
			}
			einoToolInfo, err := einoTool.Info(ctx)
			if err != nil {
				log.Error().Err(err).Str("server", serverTools.ServerName).Str("tool", tool.Name).Msg("failed to get eino tool info")
				continue
			}
			einoToolInfo.Name = fmt.Sprintf("%s__%s", serverTools.ServerName, tool.Name)
			tools = append(tools, einoToolInfo)
		}
	}

	log.Info().Int("count", len(tools)).Msg("eino tool infos loaded")
	return tools, nil
}

// parseHeaders parses header strings into a map
func parseHeaders(headerList []string) map[string]string {
	headers := make(map[string]string)
	for _, header := range headerList {
		parts := strings.SplitN(header, ":", 2)
		if len(parts) == 2 {
			headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return headers
}

// startStdioLog starts a goroutine to print stdio logs
func startStdioLog(stderr io.Reader, serverName string) {
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			fmt.Fprintf(os.Stderr, "MCP Server %s: %s\n", serverName, scanner.Text())
		}
	}()
}

// prepareClientInitRequest creates a standard initialization request
func prepareClientInitRequest() mcp.InitializeRequest {
	return mcp.InitializeRequest{
		Params: struct {
			ProtocolVersion string                 `json:"protocolVersion"`
			Capabilities    mcp.ClientCapabilities `json:"capabilities"`
			ClientInfo      mcp.Implementation     `json:"clientInfo"`
		}{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			Capabilities:    mcp.ClientCapabilities{},
			ClientInfo: mcp.Implementation{
				Name:    "hrp-mcphost",
				Version: version.GetVersionInfo(),
			},
		},
	}
}

// handleToolError handles tool execution errors
func handleToolError(result *mcp.CallToolResult) error {
	if !result.IsError {
		return nil
	}
	if len(result.Content) > 0 {
		return fmt.Errorf("tool error: %v", result.Content[0])
	}
	return fmt.Errorf("tool error: unknown error")
}
