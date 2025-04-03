package mcp

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetTools(t *testing.T) {
	hub, err := NewMCPHub("./test.mcp.json")
	require.NoError(t, err)

	ctx := context.Background()
	err = hub.InitServers(ctx)
	require.NoError(t, err)

	tools := hub.GetTools(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(tools))
}

func TestCallTool(t *testing.T) {
	hub, err := NewMCPHub("./test.mcp.json")
	require.NoError(t, err)

	ctx := context.Background()
	err = hub.InitServers(ctx)
	require.NoError(t, err)

	result, err := hub.InvokeTool(ctx, "weather", "get_alerts", map[string]interface{}{
		"state": "CA",
	})
	require.NoError(t, err)
	t.Logf("Result: %v", result)
}

func TestCallEinoTool(t *testing.T) {
	hub, err := NewMCPHub("./test.mcp.json")
	require.NoError(t, err)

	ctx := context.Background()
	err = hub.InitServers(ctx)
	require.NoError(t, err)

	einoTool, err := hub.GetEinoTool(ctx, "weather", "get_alerts")
	require.NoError(t, err)
	t.Logf("Tool: %v", einoTool)

	tool := einoTool.(tool.InvokableTool)
	result, err := tool.InvokableRun(ctx, `{"state": "CA"}`)
	require.NoError(t, err)
	t.Logf("Result: %v", result)
}

func TestConvertToolsToRecordsFromFile(t *testing.T) {
	hub, err := NewMCPHub("./test.mcp.json")
	require.NoError(t, err)

	ctx := context.Background()
	err = hub.InitServers(ctx)
	require.NoError(t, err)

	tools := hub.GetTools(ctx)
	require.NoError(t, err)

	records := ConvertToolsToRecords(tools)

	// Convert records to JSON
	recordsJSON, err := json.Marshal(records)
	require.NoError(t, err)

	// Write JSON to file
	err = os.WriteFile("./tools_records.json", recordsJSON, 0o644)
	require.NoError(t, err)

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
	tests := []struct {
		name     string
		toolsMap map[string]MCPTools
		want     []MCPToolRecord
	}{
		{
			name: "convert weather tool",
			toolsMap: map[string]MCPTools{
				"weather": {
					Name: "weather",
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
					ToolID:      "weather_get_alerts",
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
			toolsMap: map[string]MCPTools{
				"ui": {
					Name: "ui",
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
					ToolID:      "ui_swipe",
					ServerName:  "ui",
					ToolName:    "swipe",
					Description: "Do screen swipe action.",
					Parameters:  `{"direction":"swipe direction (up, down)"}`,
					Returns:     "{}",
				},
				{
					ToolID:      "ui_tap",
					ServerName:  "ui",
					ToolName:    "tap",
					Description: "Tap on screen at specified position.",
					Parameters:  "{}",
					Returns:     "{}",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertToolsToRecords(tt.toolsMap)

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
