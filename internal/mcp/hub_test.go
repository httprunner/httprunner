package mcp

import (
	"context"
	"testing"

	"github.com/cloudwego/eino/components/tool"
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
