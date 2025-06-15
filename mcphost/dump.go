package mcphost

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/uixt"
	"github.com/httprunner/httprunner/v5/uixt/option"
)

// MCPToolRecord represents a single tool record in the database
// Each record contains detailed information about a tool and its server
type MCPToolRecord struct {
	ToolID          string    `json:"tool_id"`          // Unique identifier for the tool record
	BizID           string    `json:"biz_id"`           // Business ID of the tool
	VisibleRange    int       `json:"visible_range"`    // Visible range of the tool, 0: visible to biz, 1: visible to all
	ToolType        string    `json:"tool_type"`        // Type of the tool
	ServerName      string    `json:"mcp_server"`       // Name of the MCP server
	ToolName        string    `json:"tool_name"`        // Name of the tool
	Description     string    `json:"description"`      // Tool description
	Parameters      string    `json:"parameters"`       // Tool input parameters in JSON format
	Returns         string    `json:"return_desc"`      // Tool return value format in JSON format
	TeardownPair    string    `json:"teardown_pair"`    // Teardown pair of the tool
	Examples        string    `json:"examples"`         // Examples of the tool
	SupportPatterns string    `json:"support_patterns"` // Support pattern of the tool
	CreatedAt       time.Time `json:"created_at"`       // Record creation time
	LastUpdatedAt   time.Time `json:"last_updated_at"`  // Record last update time
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

// ActionToolProvider defines the interface for MCP servers that provide ActionTool implementations
type ActionToolProvider interface {
	GetToolByAction(actionName option.ActionName) uixt.ActionTool
}

// ConvertToolsToRecords converts []MCPTools to a list of database records
func (host *MCPHost) ConvertToolsToRecords(tools []MCPTools) []MCPToolRecord {
	var records []MCPToolRecord
	now := time.Now()

	for _, mcpTools := range tools {
		if mcpTools.Err != nil {
			log.Error().Str("server", mcpTools.ServerName).Err(mcpTools.Err).Msg("skip tools conversion due to error")
			continue
		}

		for _, tool := range mcpTools.Tools {
			record := host.convertSingleToolToRecord(mcpTools.ServerName, tool, now)
			records = append(records, record)
		}
	}

	return records
}

// convertSingleToolToRecord converts a single MCP tool to a database record
func (host *MCPHost) convertSingleToolToRecord(serverName string, tool mcp.Tool, timestamp time.Time) MCPToolRecord {
	// Generate unique ID
	id := fmt.Sprintf("%s__%s", serverName, tool.Name)

	// Extract description from docstring
	info := extractDocStringInfo(tool.Description)

	// Extract parameters
	paramsJSON := host.extractParameters(tool, info)

	// Extract returns
	returnsJSON := host.extractReturns(serverName, tool.Name, info)

	return MCPToolRecord{
		ToolID:        id,
		VisibleRange:  1,
		ToolType:      "Hrp",
		ServerName:    serverName,
		ToolName:      tool.Name,
		Description:   info.Description,
		Parameters:    paramsJSON,
		Returns:       returnsJSON,
		CreatedAt:     timestamp,
		LastUpdatedAt: timestamp,
	}
}

// extractParameters extracts parameter information from tool schema or docstring
func (host *MCPHost) extractParameters(tool mcp.Tool, info DocStringInfo) string {
	// Priority 1: Extract from InputSchema.Properties
	if len(tool.InputSchema.Properties) > 0 {
		return host.extractParametersFromSchema(tool.InputSchema.Properties)
	}

	// Priority 2: Extract from docstring
	if len(info.Parameters) > 0 {
		return host.marshalToJSON(info.Parameters, "docstring parameters")
	}

	return "{}"
}

// extractParametersFromSchema extracts parameters from MCP tool input schema
func (host *MCPHost) extractParametersFromSchema(properties map[string]interface{}) string {
	schemaParams := make(map[string]string)

	for propName, propValue := range properties {
		propMap, ok := propValue.(map[string]interface{})
		if !ok {
			continue
		}

		description := host.getPropertyDescription(propMap)
		schemaParams[propName] = description
	}

	return host.marshalToJSON(schemaParams, "schema parameters")
}

// getPropertyDescription extracts description from property map
func (host *MCPHost) getPropertyDescription(propMap map[string]interface{}) string {
	if desc, exists := propMap["description"]; exists {
		if descStr, ok := desc.(string); ok {
			return descStr
		}
	}

	// Fallback to type information
	if propType, exists := propMap["type"]; exists {
		if typeStr, ok := propType.(string); ok {
			return fmt.Sprintf("Parameter of type %s", typeStr)
		}
	}

	return "Parameter"
}

// extractReturns extracts return value information from ActionTool or docstring
func (host *MCPHost) extractReturns(serverName, toolName string, info DocStringInfo) string {
	// Priority 1: Get from ActionTool interface if available
	if actionToolProvider := host.getActionToolProvider(serverName); actionToolProvider != nil {
		if actionTool := actionToolProvider.GetToolByAction(option.ActionName(toolName)); actionTool != nil {
			returnSchema := uixt.GenerateReturnSchema(actionTool)
			if len(returnSchema) > 0 {
				return host.marshalToJSON(returnSchema, "return schema")
			}
		}
	}

	// Priority 2: Use docstring returns as fallback
	if len(info.Returns) > 0 {
		return host.marshalToJSON(info.Returns, "docstring returns")
	}

	return "{}"
}

// marshalToJSON marshals data to JSON string with error handling
func (host *MCPHost) marshalToJSON(data interface{}, dataType string) string {
	jsonBytes, err := sonic.MarshalString(data)
	if err != nil {
		log.Warn().Interface("data", data).Err(err).
			Msgf("failed to marshal %s to JSON", dataType)
		return "{}"
	}
	return jsonBytes
}

// ExportToolsToJSON dumps MCP tools to JSON file
func (h *MCPHost) ExportToolsToJSON(ctx context.Context, dumpPath string) error {
	// get all tools
	tools := h.GetTools(ctx)
	// convert to records
	records := h.ConvertToolsToRecords(tools)
	// convert to JSON
	recordsJSON, err := sonic.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal records to JSON: %w", err)
	}
	// create output directory
	outputDir := filepath.Dir(dumpPath)
	if outputDir != "." {
		if err := os.MkdirAll(outputDir, 0o754); err != nil {
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
