package mcphost

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	mcpp "github.com/cloudwego/eino-ext/components/tool/mcp"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/httprunner/httprunner/v5/internal/version"
	"github.com/httprunner/httprunner/v5/uixt"
	"github.com/httprunner/httprunner/v5/uixt/ai"
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
	withUIXT    bool
	ctx         context.Context
	cancel      context.CancelFunc
	shutdownCh  chan struct{}
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
func NewMCPHost(configPath string, withUIXT bool) (*MCPHost, error) {
	config, err := LoadMCPConfig(configPath)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	host := &MCPHost{
		connections: make(map[string]*Connection),
		config:      config,
		withUIXT:    withUIXT,
		ctx:         ctx,
		cancel:      cancel,
		shutdownCh:  make(chan struct{}),
	}

	// Set up signal handling
	go host.handleSignals()

	// Initialize MCP servers
	if err := host.InitServers(ctx); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize MCP servers: %w", err)
	}

	return host, nil
}

// InitServers initializes all MCP servers
func (h *MCPHost) InitServers(ctx context.Context) error {
	// initialize uixt MCP server
	if h.withUIXT {
		h.connections["uixt"] = &Connection{
			Client: &uixt.MCPClient4XTDriver{
				Server: uixt.NewMCPServer(),
			},
			Config: nil,
		}
	}

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

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

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
		// Start with current process environment variables
		env := os.Environ()

		// Add or override with config-specific environment variables
		for k, v := range cfg.Env {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}

		mcpClient, err = client.NewStdioMCPClient(cfg.Command, env, cfg.Args...)
		if err == nil {
			if stdioClient, ok := mcpClient.(*client.Client); ok {
				stderr, _ := client.GetStderr(stdioClient)
				startStdioLog(stderr, serverName, h.ctx)
				log.Debug().Str("server", serverName).Msg("STDIO MCP server started")
			}
		}
	default:
		return fmt.Errorf("unsupported transport type: %s", config.GetType())
	}

	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// initialize client with timeout
	initCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	_, err = mcpClient.Initialize(initCtx, prepareClientInitRequest())
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

	// Use a longer timeout for graceful shutdown
	timeout := 5 * time.Second

	for name, conn := range h.connections {
		// Create a timeout context for each server
		ctx, cancel := context.WithTimeout(context.Background(), timeout)

		// Close server in a goroutine with timeout
		done := make(chan error, 1)
		go func(serverName string, client client.MCPClient) {
			done <- client.Close()
		}(name, conn.Client)

		select {
		case err := <-done:
			if err != nil {
				// Check if it's a signal-related error (expected during CTRL+C)
				if isSignalError(err) {
					log.Debug().Str("name", name).Err(err).
						Msg("Server terminated by signal (expected during shutdown)")
				} else {
					log.Error().Str("name", name).Err(err).Msg("Failed to close server")
				}
			} else {
				log.Info().Str("name", name).Msg("Server closed gracefully")
			}
		case <-ctx.Done():
			log.Warn().Str("name", name).Msg("Server close timeout, forcing termination")
		}

		cancel()
		delete(h.connections, name)
	}

	return nil
}

// isSignalError checks if the error is caused by signal interruption
func isSignalError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	// Common signal-related error patterns
	return strings.Contains(errStr, "signal: interrupt") ||
		strings.Contains(errStr, "signal: terminated") ||
		strings.Contains(errStr, "exit status 120") ||
		strings.Contains(errStr, "exit status 130") ||
		strings.Contains(errStr, "exit status 143") || // SIGTERM (15)
		strings.Contains(errStr, "broken pipe") ||
		strings.Contains(errStr, "connection reset")
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

// GetAllClients returns all MCP clients
func (h *MCPHost) GetAllClients() map[string]client.MCPClient {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients := make(map[string]client.MCPClient)
	for name, conn := range h.connections {
		clients[name] = conn.Client
	}
	return clients
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

	// Get connection for the server
	conn, exists := h.connections[serverName]
	if !exists {
		return nil, fmt.Errorf("no connection found for MCP server %s", serverName)
	}

	// Get tools from the specific server
	listResults, err := conn.Client.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to get tools from server %s: %w", serverName, err)
	}

	// Find the specific tool
	for _, tool := range listResults.Tools {
		if tool.Name == toolName {
			return &tool, nil
		}
	}

	return nil, fmt.Errorf("MCP tool %s/%s not found", serverName, toolName)
}

// InvokeTool calls a tool with the given arguments
func (h *MCPHost) InvokeTool(ctx context.Context,
	serverName, toolName string, arguments map[string]any,
) (*mcp.CallToolResult, error) {
	// Check if host is shutting down or context is cancelled
	select {
	case <-h.shutdownCh:
		return nil, fmt.Errorf("MCP host is shutting down")
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

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

	// Add shorter timeout for tool invocation
	toolCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	// Call tool and wait for result or cancellation
	result, err := conn.CallTool(toolCtx, req)
	if err != nil {
		// Check if it's a timeout or cancellation
		select {
		case <-h.shutdownCh:
			return nil, fmt.Errorf("MCP host is shutting down")
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-toolCtx.Done():
			return nil, fmt.Errorf("tool call timeout: %s/%s", serverName, toolName)
		default:
			return nil, errors.Wrapf(err, "call tool %s/%s failed", serverName, toolName)
		}
	}

	if result.IsError {
		if len(result.Content) > 0 {
			return nil, fmt.Errorf("invoke tool %s/%s failed: %v",
				serverName, toolName, result.Content)
		}
		return nil, fmt.Errorf("invoke tool %s/%s failed", serverName, toolName)
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

	var allTools []*schema.ToolInfo
	for _, serverTools := range results {
		if serverTools.Err != nil {
			log.Error().Err(serverTools.Err).
				Str("server", serverTools.ServerName).Msg("failed to get tools")
			continue
		}

		// convert MCP tools to eino tools
		einoTools := ai.ConvertMCPToolsToEinoToolInfos(
			serverTools.Tools, serverTools.ServerName)
		allTools = append(allTools, einoTools...)
	}

	return allTools, nil
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
func startStdioLog(stderr io.Reader, serverName string, ctx context.Context) {
	go func() {
		scanner := bufio.NewScanner(stderr)
		for {
			select {
			case <-ctx.Done():
				log.Debug().Str("server", serverName).Msg("stopping stdio log due to context cancellation")
				return
			default:
				if scanner.Scan() {
					fmt.Fprintf(os.Stderr, "MCP Server %s: %s\n", serverName, scanner.Text())
				} else {
					// Scanner finished or encountered error
					if err := scanner.Err(); err != nil {
						// Check if it's a normal shutdown error (pipe closed)
						if isNormalShutdownError(err) {
							log.Debug().Str("server", serverName).Msg("stdio log stopped due to normal shutdown")
						} else {
							log.Debug().Str("server", serverName).Err(err).Msg("stdio log scanner error")
						}
					}
					return
				}
			}
		}
	}()
}

// isNormalShutdownError checks if the error is caused by normal shutdown (pipe closed)
func isNormalShutdownError(err error) bool {
	errStr := err.Error()
	// Common pipe closed error patterns during normal shutdown
	return strings.Contains(errStr, "file already closed") ||
		strings.Contains(errStr, "broken pipe") ||
		strings.Contains(errStr, "use of closed file") ||
		strings.Contains(errStr, "read/write on closed pipe")
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

// handleSignals handles OS signals for graceful shutdown
func (h *MCPHost) handleSignals() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		log.Info().Str("signal", sig.String()).Msg("received signal, shutting down MCP servers")
		h.Shutdown()
	case <-h.ctx.Done():
		return
	}
}

// Shutdown gracefully shuts down all MCP servers
func (h *MCPHost) Shutdown() {
	log.Debug().Msg("Starting MCP host shutdown")
	h.cancel()

	// Close shutdown channel to signal shutdown
	select {
	case <-h.shutdownCh:
		// Already shutting down
		log.Debug().Msg("MCP host already shutting down")
		return
	default:
		close(h.shutdownCh)
	}

	// Close all servers with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		h.CloseServers()
	}()

	select {
	case <-done:
		log.Info().Msg("MCP servers shut down gracefully")
	case <-ctx.Done():
		log.Warn().Msg("MCP servers shutdown timeout, forcing exit")
		// Force close any remaining connections
		h.forceCloseAll()
	}
}

// forceCloseAll forcefully closes all remaining connections
func (h *MCPHost) forceCloseAll() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for name := range h.connections {
		log.Warn().Str("name", name).Msg("Force closing server")
		delete(h.connections, name)
	}
}

// getActionToolProvider returns an ActionToolProvider for the given server name if available
// This method checks if the MCP server implements the ActionToolProvider interface
func (h *MCPHost) getActionToolProvider(serverName string) ActionToolProvider {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if conn, exists := h.connections[serverName]; exists {
		// Check if the client directly implements ActionToolProvider interface
		if actionToolProvider, ok := conn.Client.(ActionToolProvider); ok {
			return actionToolProvider
		}
	}
	return nil
}
