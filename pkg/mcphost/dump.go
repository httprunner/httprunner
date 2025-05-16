package mcphost

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	mcpp "github.com/cloudwego/eino-ext/components/tool/mcp"
	"github.com/cloudwego/eino/components/tool"
	"github.com/rs/zerolog/log"
)

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

// ExportToolsToJSON dumps MCP tools to JSON file
func (h *MCPHost) ExportToolsToJSON(ctx context.Context, dumpPath string) error {
	// get all tools
	tools := h.GetTools(ctx)
	// convert to records
	records := ConvertToolsToRecords(tools)
	// convert to JSON
	recordsJSON, err := sonic.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal records to JSON: %w", err)
	}
	// create output directory
	outputDir := filepath.Dir(dumpPath)
	if outputDir != "." {
		if err := os.MkdirAll(outputDir, 0o755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}
	// write to file
	if err := os.WriteFile(dumpPath, []byte(recordsJSON), 0o644); err != nil {
		return fmt.Errorf("failed to write records to file: %w", err)
	}
	log.Info().Str("path", dumpPath).Msg("Tools records exported successfully")
	return nil
}

// GetEinoTool returns an eino tool from the MCP server
func (h *MCPHost) GetEinoTool(ctx context.Context, serverName, toolName string) (tool.BaseTool, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// filter MCP server by serverName
	conn, exists := h.connections[serverName]
	if !exists {
		return nil, fmt.Errorf("no connection found for server %s", serverName)
	}

	if conn.Config.IsDisabled() {
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
