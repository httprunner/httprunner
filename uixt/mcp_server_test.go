package uixt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMCPServer(t *testing.T) {
	server := NewMCPServer()
	assert.NotNil(t, server)

	// Check that tools are registered
	tools := server.ListTools()
	assert.Greater(t, len(tools), 0, "Should have at least one tool registered")

	// Check specific tools exist
	expectedTools := []string{
		"list_available_devices",
		"select_device",
		"list_packages",
		"launch_app",
		"terminate_app",
		"get_screen_size",
		"press_button",
		"tap_xy",
		"swipe",
		"drag",
		"screenshot",
		"home",
		"back",
		"input",
		"sleep",
	}

	registeredToolNames := make(map[string]bool)
	for _, tool := range tools {
		registeredToolNames[tool.Name] = true
	}

	for _, expectedTool := range expectedTools {
		assert.True(t, registeredToolNames[expectedTool], "Tool %s should be registered", expectedTool)
	}
}

func TestToolInterfaces(t *testing.T) {
	// Test that all tools implement the ActionTool interface correctly
	tools := []ActionTool{
		&ToolListAvailableDevices{},
		&ToolSelectDevice{},
		&ToolListPackages{},
		&ToolLaunchApp{},
		&ToolTerminateApp{},
		&ToolGetScreenSize{},
		&ToolPressButton{},
		&ToolTapXY{},
		&ToolSwipe{},
		&ToolDrag{},
		&ToolScreenShot{},
		&ToolHome{},
		&ToolBack{},
		&ToolInput{},
		&ToolSleep{},
	}

	for _, tool := range tools {
		assert.NotEmpty(t, tool.Name(), "Tool name should not be empty")
		assert.NotEmpty(t, tool.Description(), "Tool description should not be empty")
		assert.NotNil(t, tool.Options(), "Tool options should not be nil")
		assert.NotNil(t, tool.Implement(), "Tool implementation should not be nil")
	}
}
