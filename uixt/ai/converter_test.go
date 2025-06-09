package ai

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
)

func TestConvertMCPToolToEinoToolInfo(t *testing.T) {
	// Create a mock MCP tool
	mcpTool := mcp.Tool{
		Name:        "test_tool",
		Description: "Test tool description",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"param1": map[string]interface{}{
					"type":        "string",
					"description": "Test parameter 1",
				},
				"param2": map[string]interface{}{
					"type":        "number",
					"description": "Test parameter 2",
				},
			},
			Required: []string{"param1"},
		},
	}

	// Convert to eino ToolInfo using shared converter
	einoTool := ConvertMCPToolToEinoToolInfo(mcpTool, "uixt")

	// Verify the conversion
	assert.NotNil(t, einoTool)
	assert.Equal(t, "uixt__test_tool", einoTool.Name)
	assert.Equal(t, "Test tool description", einoTool.Desc)
	assert.NotNil(t, einoTool.ParamsOneOf)
}

func TestConvertMCPToolWithoutParams(t *testing.T) {
	// Create a mock MCP tool without parameters
	mcpTool := mcp.Tool{
		Name:        "simple_tool",
		Description: "Simple tool without parameters",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
		},
	}

	// Convert to eino ToolInfo using shared converter
	einoTool := ConvertMCPToolToEinoToolInfo(mcpTool, "uixt")

	// Verify the conversion
	assert.NotNil(t, einoTool)
	assert.Equal(t, "uixt__simple_tool", einoTool.Name)
	assert.Equal(t, "Simple tool without parameters", einoTool.Desc)
}

func TestConvertMCPToolsToEinoToolInfos(t *testing.T) {
	// Create multiple mock MCP tools
	mcpTools := []mcp.Tool{
		{
			Name:        "tool1",
			Description: "First tool",
			InputSchema: mcp.ToolInputSchema{Type: "object"},
		},
		{
			Name:        "tool2",
			Description: "Second tool",
			InputSchema: mcp.ToolInputSchema{Type: "object"},
		},
	}

	// Convert to eino ToolInfos using shared converter
	einoTools := ConvertMCPToolsToEinoToolInfos(mcpTools, "test_server")

	// Verify the conversion
	assert.Len(t, einoTools, 2)
	assert.Equal(t, "test_server__tool1", einoTools[0].Name)
	assert.Equal(t, "test_server__tool2", einoTools[1].Name)
	assert.Equal(t, "First tool", einoTools[0].Desc)
	assert.Equal(t, "Second tool", einoTools[1].Desc)
}
