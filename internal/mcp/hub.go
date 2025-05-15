package mcp

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/bytedance/sonic"
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

	log.Debug().Str("server", serverName).Msg("connecting to MCP server")

	// Close existing connection if any
	if existing, exists := h.connections[serverName]; exists {
		if err := existing.Client.Close(); err != nil {
			return fmt.Errorf("failed to close existing connection: %w", err)
		}
		delete(h.connections, serverName)
	}

	var mcpClient *client.Client
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

		// print MCP Server logs for stdio transport
		stderr, _ := client.GetStderr(mcpClient)
		go func() {
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				fmt.Fprintf(os.Stderr, "MCP Server %s: %s\n",
					serverName, scanner.Text())
			}
		}()

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

// MCPToolRecord represents a single tool record in the database
// Each record contains detailed information about a tool and its server
type MCPToolRecord struct {
	ToolID        string    `json:"tool_id"`         // Unique identifier for the tool record
	ServerName    string    `json:"mcp_server"`      // Name of the MCP server
	ToolName      string    `json:"tool_name"`       // Name of the tool
	Description   string    `json:"description"`     // Tool description
	Parameters    string    `json:"parameters"`      // Tool input parameters in JSON format
	Returns       string    `json:"returns"`         // Tool return value format in JSON format
	CreatedAt     time.Time `json:"created_at"`      // Record creation time
	LastUpdatedAt time.Time `json:"last_updated_at"` // Record last update time
}

// DocStringInfo contains the parsed information from a Python docstring
type DocStringInfo struct {
	Description string
	Parameters  map[string]string
	Returns     map[string]string
}

// extractDocStringInfo extracts information from a Python docstring
// Example input:
// """Get weather alerts for a US state.
//
//	Args:
//	    state: Two-letter US state code (e.g. CA, NY)
//
//	Returns:
//	    alerts: List of active weather alerts for the specified state
//	    error: Error message if the request fails
//	"""
func extractDocStringInfo(docstring string) DocStringInfo {
	info := DocStringInfo{
		Parameters: make(map[string]string),
		Returns:    make(map[string]string),
	}

	// Find the Args and Returns sections
	argsIndex := strings.Index(docstring, "Args:")
	returnsIndex := strings.Index(docstring, "Returns:")

	// Extract description (everything before Args)
	if argsIndex != -1 {
		info.Description = strings.TrimSpace(docstring[:argsIndex])
	} else if returnsIndex != -1 {
		info.Description = strings.TrimSpace(docstring[:returnsIndex])
	} else {
		info.Description = strings.TrimSpace(docstring)
		return info
	}

	// Helper function to extract key-value pairs from a section
	extractSection := func(content string) map[string]string {
		result := make(map[string]string)
		lines := strings.Split(content, "\n")

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			if key != "" && value != "" {
				result[key] = value
			}
		}

		return result
	}

	// Extract Args section
	if argsIndex != -1 {
		endIndex := returnsIndex
		if endIndex == -1 {
			endIndex = len(docstring)
		}
		argsContent := docstring[argsIndex+len("Args:") : endIndex]
		info.Parameters = extractSection(argsContent)
	}

	// Extract Returns section
	if returnsIndex != -1 {
		returnsContent := docstring[returnsIndex+len("Returns:"):]
		info.Returns = extractSection(returnsContent)
	}

	return info
}

// ConvertToolsToRecords converts map[string]MCPTools to a list of database records
func ConvertToolsToRecords(toolsMap map[string]MCPTools) []MCPToolRecord {
	var records []MCPToolRecord
	now := time.Now()

	for serverName, mcpTools := range toolsMap {
		if mcpTools.Err != nil {
			log.Error().Str("server", serverName).Err(mcpTools.Err).Msg("skip tools conversion due to error")
			continue
		}

		for _, tool := range mcpTools.Tools {
			// Generate unique ID by combining server name and tool name
			id := fmt.Sprintf("%s_%s", serverName, tool.Name)

			// Extract docstring information
			info := extractDocStringInfo(tool.Description)

			// Convert parameters and returns to JSON
			paramsJSON, err := sonic.MarshalString(info.Parameters)
			if err != nil {
				log.Warn().Interface("params", info.Parameters).Err(err).Msg("failed to marshal parameters to JSON")
				paramsJSON = "{}"
			}

			returnsJSON, err := sonic.MarshalString(info.Returns)
			if err != nil {
				log.Warn().Interface("returns", info.Returns).Err(err).Msg("failed to marshal returns to JSON")
				returnsJSON = "{}"
			}

			record := MCPToolRecord{
				ToolID:        id,
				ServerName:    serverName,
				ToolName:      tool.Name,
				Description:   info.Description,
				Parameters:    paramsJSON,
				Returns:       returnsJSON,
				CreatedAt:     now,
				LastUpdatedAt: now,
			}

			records = append(records, record)
		}
	}

	return records
}
