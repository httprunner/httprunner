package uixt

import (
	"testing"

	"github.com/httprunner/httprunner/v5/uixt/option"
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
		"tap_xy",
		"tap_abs_xy",
		"tap_ocr",
		"tap_cv",
		"double_tap_xy",
		"swipe_direction",
		"swipe_coordinate",
		"swipe_to_tap_app",
		"swipe_to_tap_text",
		"swipe_to_tap_texts",
		"drag",
		"input",
		"screenshot",
		"get_screen_size",
		"press_button",
		"home",
		"back",
		"list_packages",
		"app_launch",
		"app_terminate",
		"app_install",
		"app_uninstall",
		"app_clear",
		"sleep",
		"sleep_ms",
		"sleep_random",
		"set_ime",
		"get_source",
		"close_popups",
		"web_login_none_ui",
		"secondary_click",
		"hover_by_selector",
		"tap_by_selector",
		"secondary_click_by_selector",
		"web_close_tab",
		"ai_action",
		"finished",
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
		&ToolTapXY{},
		&ToolTapAbsXY{},
		&ToolTapByOCR{},
		&ToolTapByCV{},
		&ToolDoubleTapXY{},
		&ToolSwipeDirection{},
		&ToolSwipeCoordinate{},
		&ToolSwipeToTapApp{},
		&ToolSwipeToTapText{},
		&ToolSwipeToTapTexts{},
		&ToolDrag{},
		&ToolInput{},
		&ToolScreenShot{},
		&ToolGetScreenSize{},
		&ToolPressButton{},
		&ToolHome{},
		&ToolBack{},
		&ToolListPackages{},
		&ToolLaunchApp{},
		&ToolTerminateApp{},
		&ToolAppInstall{},
		&ToolAppUninstall{},
		&ToolAppClear{},
		&ToolSleep{},
		&ToolSleepMS{},
		&ToolSleepRandom{},
		&ToolSetIme{},
		&ToolGetSource{},
		&ToolClosePopups{},
		&ToolWebLoginNoneUI{},
		&ToolSecondaryClick{},
		&ToolHoverBySelector{},
		&ToolTapBySelector{},
		&ToolSecondaryClickBySelector{},
		&ToolWebCloseTab{},
		&ToolAIAction{},
		&ToolFinished{},
	}

	for _, tool := range tools {
		assert.NotEmpty(t, string(tool.Name()), "Tool name should not be empty")
		assert.NotEmpty(t, tool.Description(), "Tool description should not be empty")
		assert.NotNil(t, tool.Options(), "Tool options should not be nil")
		assert.NotNil(t, tool.Implement(), "Tool implementation should not be nil")
	}
}

func TestIgnoreNotFoundErrorOption(t *testing.T) {
	// Test that ignore_NotFoundError option is properly extracted and applied
	server := NewMCPServer()

	// Test TapByOCR tool
	tapOCRTool := server.GetToolByAction(option.ACTION_TapByOCR)
	assert.NotNil(t, tapOCRTool, "TapByOCR tool should be available")

	// Create a mock action with ignore_NotFoundError option
	actionOptions := option.NewActionOptions(
		option.WithIgnoreNotFoundError(true),
		option.WithMaxRetryTimes(2),
		option.WithIndex(1),
		option.WithRegex(true),
		option.WithTapRandomRect(true),
	)
	action := MobileAction{
		Method:        option.ACTION_TapByOCR,
		Params:        "test_text",
		ActionOptions: *actionOptions,
	}

	// Convert action to MCP call tool request
	request, err := tapOCRTool.ConvertActionToCallToolRequest(action)
	assert.NoError(t, err, "Should convert action to request without error")

	// Verify that ignore_NotFoundError option is included in arguments
	args := request.Params.Arguments
	assert.Equal(t, true, args["ignore_NotFoundError"], "ignore_NotFoundError should be true")
	assert.Equal(t, 2, args["max_retry_times"], "max_retry_times should be 2")
	assert.Equal(t, 1, args["index"], "index should be 1")
	assert.Equal(t, true, args["regex"], "regex should be true")
	assert.Equal(t, true, args["tap_random_rect"], "tap_random_rect should be true")
	assert.Equal(t, "test_text", args["text"], "text should be test_text")
}

func TestExtractActionOptionsToArguments(t *testing.T) {
	// Test the extractActionOptionsToArguments helper function
	actionOptions := []option.ActionOption{
		option.WithIgnoreNotFoundError(true),
		option.WithMaxRetryTimes(3),
		option.WithIndex(2),
		option.WithRegex(true),
		option.WithTapRandomRect(false), // false should not be included
		option.WithDuration(1.5),
	}

	arguments := make(map[string]any)
	extractActionOptionsToArguments(actionOptions, arguments)

	// Verify extracted options
	assert.Equal(t, true, arguments["ignore_NotFoundError"], "ignore_NotFoundError should be extracted")
	assert.Equal(t, 3, arguments["max_retry_times"], "max_retry_times should be extracted")
	assert.Equal(t, 2, arguments["index"], "index should be extracted")
	assert.Equal(t, true, arguments["regex"], "regex should be extracted")
	assert.Equal(t, 1.5, arguments["duration"], "duration should be extracted")

	// tap_random_rect should not be included since it's false
	_, exists := arguments["tap_random_rect"]
	assert.False(t, exists, "tap_random_rect should not be included when false")
}
