package mcphost

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadSettings(t *testing.T) {
	// Load settings from test.mcp.json
	settings, err := LoadMCPConfig("testdata/test.mcp.json")
	if err != nil {
		t.Fatalf("Failed to load settings: %v", err)
	}

	// Verify settings are loaded correctly
	assert.NotNil(t, settings)
	assert.Contains(t, settings.MCPServers, "filesystem")
	assert.Contains(t, settings.MCPServers, "weather")

	// Verify specific server configurations
	filesystemConfig := settings.MCPServers["filesystem"].Config.(STDIOServerConfig)
	assert.Equal(t, "npx", filesystemConfig.Command)
	assert.Equal(t, []string{"-y", "@modelcontextprotocol/server-filesystem", "./"}, filesystemConfig.Args)

	weatherConfig := settings.MCPServers["weather"].Config.(STDIOServerConfig)
	assert.Equal(t, "uv", weatherConfig.Command)
	assert.Equal(t, []string{"--directory", "/Users/debugtalk/MyProjects/HttpRunner-dev/httprunner/mcphost/testdata", "run", "demo_weather.py"}, weatherConfig.Args)
	assert.Equal(t, map[string]string{"ABC": "123"}, weatherConfig.Env)
}
