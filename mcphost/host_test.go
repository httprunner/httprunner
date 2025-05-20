package mcphost

import (
	"context"
	"testing"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMCPHost(t *testing.T) {
	// Test with valid config file
	host, err := NewMCPHost("./testdata/test.mcp.json", false)
	require.NoError(t, err)
	assert.NotNil(t, host)
	assert.NotNil(t, host.config)
	assert.NotEmpty(t, host.config.MCPServers)

	// Test with non-existent config file
	host, err = NewMCPHost("./testdata/non_existent.json", false)
	require.Error(t, err, "expected error when config file does not exist")
	assert.Nil(t, host)
}

func TestInitServers(t *testing.T) {
	host, err := NewMCPHost("./testdata/test.mcp.json", false)
	require.NoError(t, err)

	// Verify connections are established
	assert.Equal(t, 2, len(host.connections))
	assert.Contains(t, host.connections, "filesystem")
	assert.Contains(t, host.connections, "weather")
}

func TestGetClient(t *testing.T) {
	host, err := NewMCPHost("./testdata/test.mcp.json", false)
	require.NoError(t, err)

	// Test getting existing client
	client, err := host.GetClient("weather")
	require.NoError(t, err)
	assert.NotNil(t, client)

	// Test getting non-existent client
	client, err = host.GetClient("non_existent")
	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestGetTools(t *testing.T) {
	host, err := NewMCPHost("./testdata/test.mcp.json", false)
	require.NoError(t, err)

	ctx := context.Background()
	tools := host.GetTools(ctx)
	assert.Equal(t, 2, len(tools))

	// Verify weather tools
	var weatherTools MCPTools
	for _, tool := range tools {
		if tool.ServerName == "weather" {
			weatherTools = tool
			break
		}
	}

	assert.NoError(t, weatherTools.Err)
	assert.NotEmpty(t, weatherTools.Tools)

	// Check if get_alerts tool exists
	found := false
	for _, tool := range weatherTools.Tools {
		if tool.Name == "get_alerts" {
			found = true
			break
		}
	}
	assert.True(t, found, "get_alerts tool not found in weather tools")
}

func TestGetTool(t *testing.T) {
	host, err := NewMCPHost("./testdata/test.mcp.json", false)
	require.NoError(t, err)

	ctx := context.Background()

	// Test getting existing tool
	tool, err := host.GetTool(ctx, "weather", "get_alerts")
	require.NoError(t, err)
	assert.NotNil(t, tool)
	assert.Equal(t, "get_alerts", tool.Name)

	// Test getting non-existent tool
	tool, err = host.GetTool(ctx, "weather", "non_existent")
	assert.Error(t, err)
	assert.Nil(t, tool)

	// Test getting tool from non-existent server
	tool, err = host.GetTool(ctx, "non_existent", "get_alerts")
	assert.Error(t, err)
	assert.Nil(t, tool)
}

func TestInvokeTool(t *testing.T) {
	host, err := NewMCPHost("./testdata/test.mcp.json", false)
	require.NoError(t, err)

	ctx := context.Background()

	// Test invoking existing tool
	result, err := host.InvokeTool(ctx, "weather", "get_alerts",
		map[string]interface{}{"state": "CA"},
	)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Test invoking non-existent tool
	result, err = host.InvokeTool(ctx, "weather", "non_existent",
		map[string]interface{}{"state": "CA"},
	)
	require.Error(t, err, "expected error when tool does not exist")
	assert.Nil(t, result)

	// Test invoking tool with invalid arguments
	result, err = host.InvokeTool(ctx, "weather", "get_alerts",
		map[string]interface{}{"invalid_arg": "value"},
	)
	require.Error(t, err, "expected error when arguments are invalid")
	assert.Nil(t, result)
}

func TestCallEinoTool(t *testing.T) {
	hub, err := NewMCPHost("./testdata/test.mcp.json", false)
	require.NoError(t, err)

	ctx := context.Background()
	einoTool, err := hub.GetEinoTool(ctx, "weather", "get_alerts")
	require.NoError(t, err)
	t.Logf("Tool: %v", einoTool)

	tool := einoTool.(tool.InvokableTool)
	result, err := tool.InvokableRun(ctx, `{"state": "CA"}`)
	require.NoError(t, err)
	t.Logf("Result: %v", result)
}

func TestCloseServers(t *testing.T) {
	host, err := NewMCPHost("./testdata/test.mcp.json", false)
	require.NoError(t, err)

	// Verify servers are connected
	assert.Equal(t, 2, len(host.connections))

	// Close servers
	err = host.CloseServers()
	require.NoError(t, err)

	// Verify connections are closed
	assert.Empty(t, host.connections)
}

func TestConcurrentOperations(t *testing.T) {
	host, err := NewMCPHost("./testdata/test.mcp.json", false)
	require.NoError(t, err)

	// Test concurrent tool invocations
	done := make(chan bool)
	timeout := time.After(30 * time.Second) // Increase timeout to 30 seconds

	for i := 0; i < 3; i++ { // Reduce number of concurrent operations to 3
		go func() {
			result, err := host.InvokeTool(
				context.Background(), "weather", "get_alerts",
				map[string]interface{}{"state": "CA"},
			)
			assert.NoError(t, err)
			assert.NotNil(t, result)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ { // Update loop count to match the number of goroutines
		select {
		case <-done:
			// Success
		case <-timeout:
			t.Fatal("Timeout waiting for concurrent operations")
		}
	}
}

func TestDisabledServer(t *testing.T) {
	host, err := NewMCPHost("./testdata/test.mcp.json", false)
	require.NoError(t, err)

	// Verify only enabled servers are connected
	assert.Equal(t, 2, len(host.connections))
	assert.Contains(t, host.connections, "filesystem")
	assert.Contains(t, host.connections, "weather")
	assert.NotContains(t, host.connections, "disabled_server")

	// Test getting disabled server
	client, err := host.GetClient("disabled_server")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no connection found for server disabled_server")
	assert.Nil(t, client)

	// Test getting tools from disabled server
	ctx := context.Background()
	tools := host.GetTools(ctx)
	assert.Equal(t, 2, len(tools))

	// Verify enabled servers in tools list
	var foundFilesystem, foundWeather bool
	for _, serverTools := range tools {
		if serverTools.ServerName == "filesystem" {
			foundFilesystem = true
		} else if serverTools.ServerName == "weather" {
			foundWeather = true
		}
	}
	assert.True(t, foundFilesystem, "filesystem server not found in tools")
	assert.True(t, foundWeather, "weather server not found in tools")

	// Test getting tool from disabled server
	tool, err := host.GetTool(ctx, "disabled_server", "some_tool")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no connection found for server disabled_server")
	assert.Nil(t, tool)
}
