package mcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadSettings(t *testing.T) {
	// Load settings from test.mcp.json
	settings, err := LoadSettings("testdata/test.mcp.json")
	if err != nil {
		t.Fatalf("Failed to load settings: %v", err)
	}

	// Verify settings are loaded correctly
	assert.NotNil(t, settings)
	assert.Contains(t, settings.MCPServers, "filesystem")
	assert.Contains(t, settings.MCPServers, "weather")

	// Verify specific server configurations
	filesystemConfig := settings.MCPServers["filesystem"]
	assert.Equal(t, "npx", filesystemConfig.Command)
	assert.Equal(t, []string{"-y", "@modelcontextprotocol/server-filesystem", "/tmp"}, filesystemConfig.Args)

	weatherConfig := settings.MCPServers["weather"]
	assert.Equal(t, "uv", weatherConfig.Command)
	assert.Equal(t, []string{"--directory", "/Users/debugtalk/MyProjects/HttpRunner-dev/httprunner/internal/mcp/testdata", "run", "demo_weather.py"}, weatherConfig.Args)
	assert.Equal(t, []string{"get_forecast"}, weatherConfig.AutoApprove)
	assert.Equal(t, map[string]string{"ABC": "123"}, weatherConfig.Env)
}
