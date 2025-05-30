package mcphost

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertToolsToRecordsFromFile(t *testing.T) {
	hub, err := NewMCPHost("./testdata/test.mcp.json", true)
	require.NoError(t, err)

	// use ExportToolsToJSON to dump tools to JSON file
	err = hub.ExportToolsToJSON(context.Background(), "./tools_records.json")
	require.NoError(t, err)

	// read the exported JSON file
	data, err := os.ReadFile("./tools_records.json")
	require.NoError(t, err)

	// parse the exported JSON data
	var records []MCPToolRecord
	err = json.Unmarshal(data, &records)
	require.NoError(t, err)

	// verify the number of records
	assert.NotEmpty(t, records, "Exported records should not be empty")

	t.Logf("Tools records written to ./tools_records.json")
}

func TestExtractDocStringInfo(t *testing.T) {
	tests := []struct {
		name      string
		docstring string
		want      DocStringInfo
	}{
		{
			name: "complete docstring with args and returns",
			docstring: `Get weather alerts for a US state.

    Args:
        state: Two-letter US state code (e.g. CA, NY)

    Returns:
        alerts: List of active weather alerts for the specified state
        error: Error message if the request fails
    `,
			want: DocStringInfo{
				Description: "Get weather alerts for a US state.",
				Parameters: map[string]string{
					"state": "Two-letter US state code (e.g. CA, NY)",
				},
				Returns: map[string]string{
					"alerts": "List of active weather alerts for the specified state",
					"error":  "Error message if the request fails",
				},
			},
		},
		{
			name: "docstring with only args",
			docstring: `Do screen swipe action.

    Args:
        direction: swipe direction (up, down)
    `,
			want: DocStringInfo{
				Description: "Do screen swipe action.",
				Parameters: map[string]string{
					"direction": "swipe direction (up, down)",
				},
				Returns: map[string]string{},
			},
		},
		{
			name:      "docstring with only description",
			docstring: "Simple tool with no parameters.",
			want: DocStringInfo{
				Description: "Simple tool with no parameters.",
				Parameters:  map[string]string{},
				Returns:     map[string]string{},
			},
		},
		{
			name: "docstring with multiple parameters",
			docstring: `Perform complex operation.

    Args:
        param1: first parameter description
        param2: second parameter description
        param3: third parameter description

    Returns:
        result: operation result
    `,
			want: DocStringInfo{
				Description: "Perform complex operation.",
				Parameters: map[string]string{
					"param1": "first parameter description",
					"param2": "second parameter description",
					"param3": "third parameter description",
				},
				Returns: map[string]string{
					"result": "operation result",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractDocStringInfo(tt.docstring)
			assert.Equal(t, tt.want.Description, got.Description)
			assert.Equal(t, tt.want.Parameters, got.Parameters)
			assert.Equal(t, tt.want.Returns, got.Returns)
		})
	}
}

func TestConvertToolsToRecords(t *testing.T) {
	// Create a mock MCPHost for testing
	host := &MCPHost{
		connections: make(map[string]*Connection),
	}

	tests := []struct {
		name  string
		tools []MCPTools
		want  []MCPToolRecord
	}{
		{
			name: "convert weather tool",
			tools: []MCPTools{
				{
					ServerName: "weather",
					Tools: []mcp.Tool{
						{
							Name: "get_alerts",
							Description: `Get weather alerts for a US state.

    Args:
        state: Two-letter US state code (e.g. CA, NY)

    Returns:
        alerts: List of active weather alerts for the specified state
        error: Error message if the request fails
    `,
						},
					},
				},
			},
			want: []MCPToolRecord{
				{
					ToolID:      "weather__get_alerts",
					ServerName:  "weather",
					ToolName:    "get_alerts",
					Description: "Get weather alerts for a US state.",
					Parameters:  `{"state":"Two-letter US state code (e.g. CA, NY)"}`,
					Returns:     `{"alerts":"List of active weather alerts for the specified state","error":"Error message if the request fails"}`,
				},
			},
		},
		{
			name: "convert multiple tools",
			tools: []MCPTools{
				{
					ServerName: "ui",
					Tools: []mcp.Tool{
						{
							Name: "swipe",
							Description: `Do screen swipe action.

    Args:
        direction: swipe direction (up, down)
    `,
						},
						{
							Name:        "tap",
							Description: "Tap on screen at specified position.",
						},
					},
				},
			},
			want: []MCPToolRecord{
				{
					ToolID:      "ui__swipe",
					ServerName:  "ui",
					ToolName:    "swipe",
					Description: "Do screen swipe action.",
					Parameters:  `{"direction":"swipe direction (up, down)"}`,
					Returns:     "{}",
				},
				{
					ToolID:      "ui__tap",
					ServerName:  "ui",
					ToolName:    "tap",
					Description: "Tap on screen at specified position.",
					Parameters:  "{}",
					Returns:     "{}",
				},
			},
		},
		{
			name: "convert tool with InputSchema",
			tools: []MCPTools{
				{
					ServerName: "test",
					Tools: []mcp.Tool{
						{
							Name:        "test_tool",
							Description: "Test tool with input schema",
							InputSchema: mcp.ToolInputSchema{
								Type: "object",
								Properties: map[string]interface{}{
									"param1": map[string]interface{}{
										"type":        "string",
										"description": "First parameter",
									},
									"param2": map[string]interface{}{
										"type": "number",
									},
								},
							},
						},
					},
				},
			},
			want: []MCPToolRecord{
				{
					ToolID:      "test__test_tool",
					ServerName:  "test",
					ToolName:    "test_tool",
					Description: "Test tool with input schema",
					Parameters:  `{"param1":"First parameter","param2":"Parameter of type number"}`,
					Returns:     "{}",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := host.ConvertToolsToRecords(tt.tools)

			// Compare each record
			require.Equal(t, len(tt.want), len(got))
			for i := range tt.want {
				assert.Equal(t, tt.want[i].ToolID, got[i].ToolID)
				assert.Equal(t, tt.want[i].ServerName, got[i].ServerName)
				assert.Equal(t, tt.want[i].ToolName, got[i].ToolName)
				assert.Equal(t, tt.want[i].Description, got[i].Description)

				// Compare JSON content (ignoring whitespace differences)
				var wantParams, gotParams, wantReturns, gotReturns map[string]string
				require.NoError(t, json.Unmarshal([]byte(tt.want[i].Parameters), &wantParams))
				require.NoError(t, json.Unmarshal([]byte(got[i].Parameters), &gotParams))
				require.NoError(t, json.Unmarshal([]byte(tt.want[i].Returns), &wantReturns))
				require.NoError(t, json.Unmarshal([]byte(got[i].Returns), &gotReturns))

				assert.Equal(t, wantParams, gotParams)
				assert.Equal(t, wantReturns, gotReturns)

				// Verify timestamps are recent (within last 5 seconds)
				now := time.Now()
				assert.True(t, now.Sub(got[i].CreatedAt) < 5*time.Second, "CreatedAt should be recent")
				assert.True(t, now.Sub(got[i].LastUpdatedAt) < 5*time.Second, "LastUpdatedAt should be recent")
				// CreatedAt and LastUpdatedAt should be the same
				assert.Equal(t, got[i].CreatedAt, got[i].LastUpdatedAt)
			}
		})
	}
}

// TestExtractParameters tests the extractParameters method
func TestExtractParameters(t *testing.T) {
	host := &MCPHost{}

	tests := []struct {
		name     string
		tool     mcp.Tool
		info     DocStringInfo
		expected string
	}{
		{
			name: "extract from InputSchema",
			tool: mcp.Tool{
				InputSchema: mcp.ToolInputSchema{
					Properties: map[string]interface{}{
						"param1": map[string]interface{}{
							"type":        "string",
							"description": "First parameter",
						},
						"param2": map[string]interface{}{
							"type": "number",
						},
					},
				},
			},
			info:     DocStringInfo{Parameters: map[string]string{"old": "old param"}},
			expected: `{"param1":"First parameter","param2":"Parameter of type number"}`,
		},
		{
			name: "fallback to docstring",
			tool: mcp.Tool{},
			info: DocStringInfo{
				Parameters: map[string]string{
					"param": "parameter description",
				},
			},
			expected: `{"param":"parameter description"}`,
		},
		{
			name:     "empty parameters",
			tool:     mcp.Tool{},
			info:     DocStringInfo{},
			expected: "{}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := host.extractParameters(tt.tool, tt.info)
			assert.Equal(t, tt.expected, got)
		})
	}
}

// TestExtractReturns tests the extractReturns method
func TestExtractReturns(t *testing.T) {
	host := &MCPHost{
		connections: make(map[string]*Connection),
	}

	tests := []struct {
		name       string
		serverName string
		toolName   string
		info       DocStringInfo
		expected   string
	}{
		{
			name:       "fallback to docstring returns",
			serverName: "unknown_server",
			toolName:   "unknown_tool",
			info: DocStringInfo{
				Returns: map[string]string{
					"result": "operation result",
					"error":  "error message",
				},
			},
			expected: `{"error":"error message","result":"operation result"}`,
		},
		{
			name:       "empty returns",
			serverName: "unknown_server",
			toolName:   "unknown_tool",
			info:       DocStringInfo{},
			expected:   "{}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := host.extractReturns(tt.serverName, tt.toolName, tt.info)
			assert.Equal(t, tt.expected, got)
		})
	}
}

// TestGetPropertyDescription tests the getPropertyDescription method
func TestGetPropertyDescription(t *testing.T) {
	host := &MCPHost{}

	tests := []struct {
		name     string
		propMap  map[string]interface{}
		expected string
	}{
		{
			name: "with description",
			propMap: map[string]interface{}{
				"type":        "string",
				"description": "Parameter description",
			},
			expected: "Parameter description",
		},
		{
			name: "without description, with type",
			propMap: map[string]interface{}{
				"type": "number",
			},
			expected: "Parameter of type number",
		},
		{
			name:     "without description and type",
			propMap:  map[string]interface{}{},
			expected: "Parameter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := host.getPropertyDescription(tt.propMap)
			assert.Equal(t, tt.expected, got)
		})
	}
}

// TestMarshalToJSON tests the marshalToJSON method
func TestMarshalToJSON(t *testing.T) {
	host := &MCPHost{}

	tests := []struct {
		name     string
		data     interface{}
		dataType string
		expected string
	}{
		{
			name: "valid map",
			data: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			dataType: "test data",
			expected: `{"key1":"value1","key2":"value2"}`,
		},
		{
			name:     "empty map",
			data:     map[string]string{},
			dataType: "test data",
			expected: "{}",
		},
		{
			name:     "invalid data (channel)",
			data:     make(chan int),
			dataType: "test data",
			expected: "{}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := host.marshalToJSON(tt.data, tt.dataType)
			assert.Equal(t, tt.expected, got)
		})
	}
}
