package uixt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/httprunner/httprunner/v5/internal/json"
	"github.com/httprunner/httprunner/v5/uixt/option"
)

func TestNewMCPServer(t *testing.T) {
	server := NewMCPServer()
	assert.NotNil(t, server)
	assert.NotEmpty(t, server.ListTools())

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
		"swipe",
		"swipe_direction",
		"swipe_coordinate",
		"swipe_to_tap_app",
		"swipe_to_tap_text",
		"swipe_to_tap_texts",
		"drag",
		"input",
		"screenshot",
		"screenrecord",
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
		&ToolSwipe{},
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
		&ToolAIQuery{},
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
	action := option.MobileAction{
		Method:        option.ACTION_TapByOCR,
		Params:        "test_text",
		ActionOptions: *actionOptions,
	}

	// Convert action to MCP call tool request
	request, err := tapOCRTool.ConvertActionToCallToolRequest(action)
	assert.NoError(t, err, "Should convert action to request without error")

	// Verify that ignore_NotFoundError option is included in arguments
	args := request.GetArguments()
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

// TestToolListAvailableDevices tests the ToolListAvailableDevices implementation
func TestToolListAvailableDevices(t *testing.T) {
	tool := &ToolListAvailableDevices{}

	// Test Name
	assert.Equal(t, option.ACTION_ListAvailableDevices, tool.Name())

	// Test Description
	assert.NotEmpty(t, tool.Description())

	// Test Options
	options := tool.Options()
	assert.NotNil(t, options)

	// Test ConvertActionToCallToolRequest
	action := option.MobileAction{
		Method: option.ACTION_ListAvailableDevices,
		Params: nil,
	}
	request, err := tool.ConvertActionToCallToolRequest(action)
	assert.NoError(t, err)
	assert.Equal(t, string(option.ACTION_ListAvailableDevices), request.Params.Name)
	assert.Empty(t, request.GetArguments())
}

// TestToolSelectDevice tests the ToolSelectDevice implementation
func TestToolSelectDevice(t *testing.T) {
	tool := &ToolSelectDevice{}

	// Test Name
	assert.Equal(t, option.ACTION_SelectDevice, tool.Name())

	// Test Description
	assert.NotEmpty(t, tool.Description())

	// Test Options
	options := tool.Options()
	assert.NotNil(t, options)
	assert.Len(t, options, 2) // platform and serial

	// Test ConvertActionToCallToolRequest
	action := option.MobileAction{
		Method: option.ACTION_SelectDevice,
		Params: nil,
	}
	request, err := tool.ConvertActionToCallToolRequest(action)
	assert.NoError(t, err)
	assert.Equal(t, string(option.ACTION_SelectDevice), request.Params.Name)
}

// TestToolTapXY tests the ToolTapXY implementation
func TestToolTapXY(t *testing.T) {
	tool := &ToolTapXY{}

	// Test Name
	assert.Equal(t, option.ACTION_TapXY, tool.Name())

	// Test Description
	assert.NotEmpty(t, tool.Description())

	// Test Options
	options := tool.Options()
	assert.NotNil(t, options)

	// Test ConvertActionToCallToolRequest with valid params
	action := option.MobileAction{
		Method: option.ACTION_TapXY,
		Params: []float64{0.5, 0.6},
		ActionOptions: option.ActionOptions{
			Duration: 1.5,
		},
	}
	request, err := tool.ConvertActionToCallToolRequest(action)
	assert.NoError(t, err)
	assert.Equal(t, string(option.ACTION_TapXY), request.Params.Name)
	args := request.GetArguments()
	assert.Equal(t, 0.5, args["x"])
	assert.Equal(t, 0.6, args["y"])
	assert.Equal(t, 1.5, args["duration"])

	// Test ConvertActionToCallToolRequest with invalid params
	invalidAction := option.MobileAction{
		Method: option.ACTION_TapXY,
		Params: "invalid",
	}
	_, err = tool.ConvertActionToCallToolRequest(invalidAction)
	assert.Error(t, err)
}

// TestToolTapAbsXY tests the ToolTapAbsXY implementation
func TestToolTapAbsXY(t *testing.T) {
	tool := &ToolTapAbsXY{}

	// Test Name
	assert.Equal(t, option.ACTION_TapAbsXY, tool.Name())

	// Test Description
	assert.NotEmpty(t, tool.Description())

	// Test Options
	options := tool.Options()
	assert.NotNil(t, options)

	// Test ConvertActionToCallToolRequest with valid params
	action := option.MobileAction{
		Method: option.ACTION_TapAbsXY,
		Params: []float64{100.0, 200.0},
		ActionOptions: option.ActionOptions{
			Duration: 2.0,
		},
	}
	request, err := tool.ConvertActionToCallToolRequest(action)
	assert.NoError(t, err)
	assert.Equal(t, string(option.ACTION_TapAbsXY), request.Params.Name)
	args := request.GetArguments()
	assert.Equal(t, 100.0, args["x"])
	assert.Equal(t, 200.0, args["y"])
	assert.Equal(t, 2.0, args["duration"])

	// Test ConvertActionToCallToolRequest with invalid params
	invalidAction := option.MobileAction{
		Method: option.ACTION_TapAbsXY,
		Params: []float64{100.0}, // missing y coordinate
	}
	_, err = tool.ConvertActionToCallToolRequest(invalidAction)
	assert.Error(t, err)
}

// TestToolTapByOCR tests the ToolTapByOCR implementation
func TestToolTapByOCR(t *testing.T) {
	tool := &ToolTapByOCR{}

	// Test Name
	assert.Equal(t, option.ACTION_TapByOCR, tool.Name())

	// Test Description
	assert.NotEmpty(t, tool.Description())

	// Test Options
	options := tool.Options()
	assert.NotNil(t, options)

	// Test ConvertActionToCallToolRequest with valid params
	actionOptions := option.NewActionOptions(
		option.WithIgnoreNotFoundError(true),
		option.WithMaxRetryTimes(3),
		option.WithIndex(1),
		option.WithRegex(true),
		option.WithTapRandomRect(true),
	)
	action := option.MobileAction{
		Method:        option.ACTION_TapByOCR,
		Params:        "test_text",
		ActionOptions: *actionOptions,
	}
	request, err := tool.ConvertActionToCallToolRequest(action)
	assert.NoError(t, err)
	assert.Equal(t, string(option.ACTION_TapByOCR), request.Params.Name)
	args := request.GetArguments()
	assert.Equal(t, "test_text", args["text"])
	assert.Equal(t, true, args["ignore_NotFoundError"])
	assert.Equal(t, 3, args["max_retry_times"])
	assert.Equal(t, 1, args["index"])
	assert.Equal(t, true, args["regex"])
	assert.Equal(t, true, args["tap_random_rect"])

	// Test ConvertActionToCallToolRequest with invalid params
	invalidAction := option.MobileAction{
		Method: option.ACTION_TapByOCR,
		Params: 123, // should be string
	}
	_, err = tool.ConvertActionToCallToolRequest(invalidAction)
	assert.Error(t, err)
}

// TestToolTapByCV tests the ToolTapByCV implementation
func TestToolTapByCV(t *testing.T) {
	tool := &ToolTapByCV{}

	// Test Name
	assert.Equal(t, option.ACTION_TapByCV, tool.Name())

	// Test Description
	assert.NotEmpty(t, tool.Description())

	// Test Options
	options := tool.Options()
	assert.NotNil(t, options)

	// Test ConvertActionToCallToolRequest
	actionOptions := option.NewActionOptions(
		option.WithIgnoreNotFoundError(true),
		option.WithMaxRetryTimes(2),
		option.WithTapRandomRect(true),
	)
	action := option.MobileAction{
		Method:        option.ACTION_TapByCV,
		Params:        nil,
		ActionOptions: *actionOptions,
	}
	request, err := tool.ConvertActionToCallToolRequest(action)
	assert.NoError(t, err)
	assert.Equal(t, string(option.ACTION_TapByCV), request.Params.Name)
	args := request.GetArguments()
	assert.Equal(t, "", args["imagePath"])
	assert.Equal(t, true, args["ignore_NotFoundError"])
	assert.Equal(t, 2, args["max_retry_times"])
	assert.Equal(t, true, args["tap_random_rect"])
}

// TestToolDoubleTapXY tests the ToolDoubleTapXY implementation
func TestToolDoubleTapXY(t *testing.T) {
	tool := &ToolDoubleTapXY{}

	// Test Name
	assert.Equal(t, option.ACTION_DoubleTapXY, tool.Name())

	// Test Description
	assert.NotEmpty(t, tool.Description())

	// Test Options
	options := tool.Options()
	assert.NotNil(t, options)

	// Test ConvertActionToCallToolRequest with valid params
	action := option.MobileAction{
		Method: option.ACTION_DoubleTapXY,
		Params: []float64{0.3, 0.7},
	}
	request, err := tool.ConvertActionToCallToolRequest(action)
	assert.NoError(t, err)
	assert.Equal(t, string(option.ACTION_DoubleTapXY), request.Params.Name)
	args := request.GetArguments()
	assert.Equal(t, 0.3, args["x"])
	assert.Equal(t, 0.7, args["y"])

	// Test ConvertActionToCallToolRequest with invalid params
	invalidAction := option.MobileAction{
		Method: option.ACTION_DoubleTapXY,
		Params: "invalid",
	}
	_, err = tool.ConvertActionToCallToolRequest(invalidAction)
	assert.Error(t, err)
}

// TestToolSwipe tests the ToolSwipe implementation
func TestToolSwipe(t *testing.T) {
	tool := &ToolSwipe{}

	// Test Name
	assert.Equal(t, option.ACTION_Swipe, tool.Name())

	// Test Description
	assert.NotEmpty(t, tool.Description())

	// Test Options
	options := tool.Options()
	assert.NotNil(t, options)

	// Test ConvertActionToCallToolRequest with direction params (string)
	directionAction := option.MobileAction{
		Method: option.ACTION_Swipe,
		Params: "up",
		ActionOptions: option.ActionOptions{
			Duration:      1.5,
			PressDuration: 0.5,
		},
	}
	request, err := tool.ConvertActionToCallToolRequest(directionAction)
	assert.NoError(t, err)
	assert.Equal(t, string(option.ACTION_Swipe), request.Params.Name)
	args := request.GetArguments()
	assert.Equal(t, "up", args["direction"])
	assert.Equal(t, 1.5, args["duration"])
	assert.Equal(t, 0.5, args["pressDuration"])

	// Test ConvertActionToCallToolRequest with coordinate params
	coordinateAction := option.MobileAction{
		Method: option.ACTION_Swipe,
		Params: []float64{0.1, 0.2, 0.8, 0.9},
		ActionOptions: option.ActionOptions{
			Duration:      2.0,
			PressDuration: 1.0,
		},
	}
	request, err = tool.ConvertActionToCallToolRequest(coordinateAction)
	assert.NoError(t, err)
	assert.Equal(t, string(option.ACTION_Swipe), request.Params.Name)
	args = request.GetArguments()
	assert.Equal(t, 0.1, args["from_x"])
	assert.Equal(t, 0.2, args["from_y"])
	assert.Equal(t, 0.8, args["to_x"])
	assert.Equal(t, 0.9, args["to_y"])
	assert.Equal(t, 2.0, args["duration"])
	assert.Equal(t, 1.0, args["pressDuration"])

	// Test ConvertActionToCallToolRequest with invalid params
	invalidAction := option.MobileAction{
		Method: option.ACTION_Swipe,
		Params: 123, // should be string or []float64
	}
	_, err = tool.ConvertActionToCallToolRequest(invalidAction)
	assert.Error(t, err)

	// Test ConvertActionToCallToolRequest with incomplete coordinate params
	incompleteAction := option.MobileAction{
		Method: option.ACTION_Swipe,
		Params: []float64{0.1, 0.2}, // missing toX and toY
	}
	_, err = tool.ConvertActionToCallToolRequest(incompleteAction)
	assert.Error(t, err)
}

// TestToolSwipeDirection tests the ToolSwipeDirection implementation
func TestToolSwipeDirection(t *testing.T) {
	tool := &ToolSwipeDirection{}

	// Test Name
	assert.Equal(t, option.ACTION_SwipeDirection, tool.Name())

	// Test Description
	assert.NotEmpty(t, tool.Description())

	// Test Options
	options := tool.Options()
	assert.NotNil(t, options)

	// Test ConvertActionToCallToolRequest with valid params
	action := option.MobileAction{
		Method: option.ACTION_SwipeDirection,
		Params: "up",
		ActionOptions: option.ActionOptions{
			Duration:      1.0,
			PressDuration: 0.5,
		},
	}
	request, err := tool.ConvertActionToCallToolRequest(action)
	assert.NoError(t, err)
	assert.Equal(t, string(option.ACTION_SwipeDirection), request.Params.Name)
	args := request.GetArguments()
	assert.Equal(t, "up", args["direction"])
	assert.Equal(t, 1.0, args["duration"])
	assert.Equal(t, 0.5, args["pressDuration"])

	// Test ConvertActionToCallToolRequest with invalid params
	invalidAction := option.MobileAction{
		Method: option.ACTION_SwipeDirection,
		Params: 123, // should be string
	}
	_, err = tool.ConvertActionToCallToolRequest(invalidAction)
	assert.Error(t, err)
}

// TestToolSwipeCoordinate tests the ToolSwipeCoordinate implementation
func TestToolSwipeCoordinate(t *testing.T) {
	tool := &ToolSwipeCoordinate{}

	// Test Name
	assert.Equal(t, option.ACTION_SwipeCoordinate, tool.Name())

	// Test Description
	assert.NotEmpty(t, tool.Description())

	// Test Options
	options := tool.Options()
	assert.NotNil(t, options)

	// Test ConvertActionToCallToolRequest with valid params
	action := option.MobileAction{
		Method: option.ACTION_SwipeCoordinate,
		Params: []float64{0.1, 0.2, 0.8, 0.9},
		ActionOptions: option.ActionOptions{
			Duration:      2.0,
			PressDuration: 1.0,
		},
	}
	request, err := tool.ConvertActionToCallToolRequest(action)
	assert.NoError(t, err)
	assert.Equal(t, string(option.ACTION_SwipeCoordinate), request.Params.Name)
	args := request.GetArguments()
	assert.Equal(t, 0.1, args["from_x"])
	assert.Equal(t, 0.2, args["from_y"])
	assert.Equal(t, 0.8, args["to_x"])
	assert.Equal(t, 0.9, args["to_y"])
	assert.Equal(t, 2.0, args["duration"])
	assert.Equal(t, 1.0, args["pressDuration"])

	// Test ConvertActionToCallToolRequest with invalid params
	invalidAction := option.MobileAction{
		Method: option.ACTION_SwipeCoordinate,
		Params: []float64{0.1, 0.2}, // missing toX and toY
	}
	_, err = tool.ConvertActionToCallToolRequest(invalidAction)
	assert.Error(t, err)
}

// TestToolSwipeToTapApp tests the ToolSwipeToTapApp implementation
func TestToolSwipeToTapApp(t *testing.T) {
	tool := &ToolSwipeToTapApp{}

	// Test Name
	assert.Equal(t, option.ACTION_SwipeToTapApp, tool.Name())

	// Test Description
	assert.NotEmpty(t, tool.Description())

	// Test Options
	options := tool.Options()
	assert.NotNil(t, options)

	// Test ConvertActionToCallToolRequest with valid params
	actionOptions := option.NewActionOptions(
		option.WithIgnoreNotFoundError(true),
		option.WithMaxRetryTimes(3),
		option.WithIndex(1),
	)
	action := option.MobileAction{
		Method:        option.ACTION_SwipeToTapApp,
		Params:        "WeChat",
		ActionOptions: *actionOptions,
	}
	request, err := tool.ConvertActionToCallToolRequest(action)
	assert.NoError(t, err)
	assert.Equal(t, string(option.ACTION_SwipeToTapApp), request.Params.Name)
	args := request.GetArguments()
	assert.Equal(t, "WeChat", args["appName"])
	assert.Equal(t, true, args["ignore_NotFoundError"])
	assert.Equal(t, 3, args["max_retry_times"])
	assert.Equal(t, 1, args["index"])

	// Test ConvertActionToCallToolRequest with invalid params
	invalidAction := option.MobileAction{
		Method: option.ACTION_SwipeToTapApp,
		Params: 123, // should be string
	}
	_, err = tool.ConvertActionToCallToolRequest(invalidAction)
	assert.Error(t, err)
}

// TestToolSwipeToTapText tests the ToolSwipeToTapText implementation
func TestToolSwipeToTapText(t *testing.T) {
	tool := &ToolSwipeToTapText{}

	// Test Name
	assert.Equal(t, option.ACTION_SwipeToTapText, tool.Name())

	// Test Description
	assert.NotEmpty(t, tool.Description())

	// Test Options
	options := tool.Options()
	assert.NotNil(t, options)

	// Test ConvertActionToCallToolRequest with valid params
	actionOptions := option.NewActionOptions(
		option.WithIgnoreNotFoundError(true),
		option.WithMaxRetryTimes(2),
		option.WithRegex(true),
	)
	action := option.MobileAction{
		Method:        option.ACTION_SwipeToTapText,
		Params:        "Submit",
		ActionOptions: *actionOptions,
	}
	request, err := tool.ConvertActionToCallToolRequest(action)
	assert.NoError(t, err)
	assert.Equal(t, string(option.ACTION_SwipeToTapText), request.Params.Name)
	args := request.GetArguments()
	assert.Equal(t, "Submit", args["text"])
	assert.Equal(t, true, args["ignore_NotFoundError"])
	assert.Equal(t, 2, args["max_retry_times"])
	assert.Equal(t, true, args["regex"])

	// Test ConvertActionToCallToolRequest with invalid params
	invalidAction := option.MobileAction{
		Method: option.ACTION_SwipeToTapText,
		Params: []int{1, 2, 3}, // should be string
	}
	_, err = tool.ConvertActionToCallToolRequest(invalidAction)
	assert.Error(t, err)
}

// TestToolSwipeToTapTexts tests the ToolSwipeToTapTexts implementation
func TestToolSwipeToTapTexts(t *testing.T) {
	tool := &ToolSwipeToTapTexts{}

	// Test Name
	assert.Equal(t, option.ACTION_SwipeToTapTexts, tool.Name())

	// Test Description
	assert.NotEmpty(t, tool.Description())

	// Test Options
	options := tool.Options()
	assert.NotNil(t, options)

	// Test ConvertActionToCallToolRequest with valid params
	actionOptions := option.NewActionOptions(
		option.WithIgnoreNotFoundError(true),
		option.WithRegex(true),
	)
	action := option.MobileAction{
		Method:        option.ACTION_SwipeToTapTexts,
		Params:        []string{"OK", "确定", "Submit"},
		ActionOptions: *actionOptions,
	}
	request, err := tool.ConvertActionToCallToolRequest(action)
	assert.NoError(t, err)
	assert.Equal(t, string(option.ACTION_SwipeToTapTexts), request.Params.Name)

	args := request.GetArguments()
	texts, ok := args["texts"].([]string)
	require.True(t, ok)
	assert.Equal(t, []string{"OK", "确定", "Submit"}, texts)
	assert.Equal(t, true, args["ignore_NotFoundError"])
	assert.Equal(t, true, args["regex"])

	// Test ConvertActionToCallToolRequest with invalid params
	invalidAction := option.MobileAction{
		Method: option.ACTION_SwipeToTapTexts,
		Params: "single_string", // should be []string
	}
	_, err = tool.ConvertActionToCallToolRequest(invalidAction)
	assert.Error(t, err)
}

// TestToolDrag tests the ToolDrag implementation
func TestToolDrag(t *testing.T) {
	tool := &ToolDrag{}

	// Test Name
	assert.Equal(t, option.ACTION_Drag, tool.Name())

	// Test Description
	assert.NotEmpty(t, tool.Description())

	// Test Options
	options := tool.Options()
	assert.NotNil(t, options)

	// Test ConvertActionToCallToolRequest with valid params
	action := option.MobileAction{
		Method: option.ACTION_Drag,
		Params: []float64{0.1, 0.2, 0.8, 0.9},
		ActionOptions: option.ActionOptions{
			Duration: 2.5,
		},
	}
	request, err := tool.ConvertActionToCallToolRequest(action)
	assert.NoError(t, err)
	assert.Equal(t, string(option.ACTION_Drag), request.Params.Name)
	args := request.GetArguments()
	assert.Equal(t, 0.1, args["from_x"])
	assert.Equal(t, 0.2, args["from_y"])
	assert.Equal(t, 0.8, args["to_x"])
	assert.Equal(t, 0.9, args["to_y"])
	assert.Equal(t, 2500.0, args["duration"]) // converted to milliseconds

	// Test ConvertActionToCallToolRequest with invalid params
	invalidAction := option.MobileAction{
		Method: option.ACTION_Drag,
		Params: []float64{0.1, 0.2}, // missing toX and toY
	}
	_, err = tool.ConvertActionToCallToolRequest(invalidAction)
	assert.Error(t, err)
}

// TestToolInput tests the ToolInput implementation
func TestToolInput(t *testing.T) {
	tool := &ToolInput{}

	// Test Name
	assert.Equal(t, option.ACTION_Input, tool.Name())

	// Test Description
	assert.NotEmpty(t, tool.Description())

	// Test Options
	options := tool.Options()
	assert.NotNil(t, options)

	// Test ConvertActionToCallToolRequest with valid params
	action := option.MobileAction{
		Method: option.ACTION_Input,
		Params: "Hello World",
	}
	request, err := tool.ConvertActionToCallToolRequest(action)
	assert.NoError(t, err)
	assert.Equal(t, string(option.ACTION_Input), request.Params.Name)
	assert.Equal(t, "Hello World", request.GetArguments()["text"])
}

// TestToolScreenShot tests the ToolScreenShot implementation
func TestToolScreenShot(t *testing.T) {
	tool := &ToolScreenShot{}

	// Test Name
	assert.Equal(t, option.ACTION_ScreenShot, tool.Name())

	// Test Description
	assert.NotEmpty(t, tool.Description())

	// Test Options
	options := tool.Options()
	assert.NotNil(t, options)

	// Test ConvertActionToCallToolRequest
	action := option.MobileAction{
		Method: option.ACTION_ScreenShot,
		Params: nil,
	}
	request, err := tool.ConvertActionToCallToolRequest(action)
	assert.NoError(t, err)
	assert.Equal(t, string(option.ACTION_ScreenShot), request.Params.Name)
	assert.Empty(t, request.GetArguments())
}

// TestToolGetScreenSize tests the ToolGetScreenSize implementation
func TestToolGetScreenSize(t *testing.T) {
	tool := &ToolGetScreenSize{}

	// Test Name
	assert.Equal(t, option.ACTION_GetScreenSize, tool.Name())

	// Test Description
	assert.NotEmpty(t, tool.Description())

	// Test Options
	options := tool.Options()
	assert.NotNil(t, options)

	// Test ConvertActionToCallToolRequest
	action := option.MobileAction{
		Method: option.ACTION_GetScreenSize,
		Params: nil,
	}
	request, err := tool.ConvertActionToCallToolRequest(action)
	assert.NoError(t, err)
	assert.Equal(t, string(option.ACTION_GetScreenSize), request.Params.Name)
	assert.Empty(t, request.GetArguments())
}

// TestToolPressButton tests the ToolPressButton implementation
func TestToolPressButton(t *testing.T) {
	tool := &ToolPressButton{}

	// Test Name
	assert.Equal(t, option.ACTION_PressButton, tool.Name())

	// Test Description
	assert.NotEmpty(t, tool.Description())

	// Test Options
	options := tool.Options()
	assert.NotNil(t, options)

	// Test ConvertActionToCallToolRequest with valid params
	action := option.MobileAction{
		Method: option.ACTION_PressButton,
		Params: "HOME",
	}
	request, err := tool.ConvertActionToCallToolRequest(action)
	assert.NoError(t, err)
	assert.Equal(t, string(option.ACTION_PressButton), request.Params.Name)
	assert.Equal(t, "HOME", request.GetArguments()["button"])

	// Test ConvertActionToCallToolRequest with invalid params
	invalidAction := option.MobileAction{
		Method: option.ACTION_PressButton,
		Params: 123, // should be string
	}
	_, err = tool.ConvertActionToCallToolRequest(invalidAction)
	assert.Error(t, err)
}

// TestToolHome tests the ToolHome implementation
func TestToolHome(t *testing.T) {
	tool := &ToolHome{}

	// Test Name
	assert.Equal(t, option.ACTION_Home, tool.Name())

	// Test Description
	assert.NotEmpty(t, tool.Description())

	// Test Options
	options := tool.Options()
	assert.NotNil(t, options)

	// Test ConvertActionToCallToolRequest
	action := option.MobileAction{
		Method: option.ACTION_Home,
		Params: nil,
	}
	request, err := tool.ConvertActionToCallToolRequest(action)
	assert.NoError(t, err)
	assert.Equal(t, string(option.ACTION_Home), request.Params.Name)
	assert.Empty(t, request.GetArguments())
}

// TestToolBack tests the ToolBack implementation
func TestToolBack(t *testing.T) {
	tool := &ToolBack{}

	// Test Name
	assert.Equal(t, option.ACTION_Back, tool.Name())

	// Test Description
	assert.NotEmpty(t, tool.Description())

	// Test Options
	options := tool.Options()
	assert.NotNil(t, options)

	// Test ConvertActionToCallToolRequest
	action := option.MobileAction{
		Method: option.ACTION_Back,
		Params: nil,
	}
	request, err := tool.ConvertActionToCallToolRequest(action)
	assert.NoError(t, err)
	assert.Equal(t, string(option.ACTION_Back), request.Params.Name)
	assert.Empty(t, request.GetArguments())
}

// TestToolListPackages tests the ToolListPackages implementation
func TestToolListPackages(t *testing.T) {
	tool := &ToolListPackages{}

	// Test Name
	assert.Equal(t, option.ACTION_ListPackages, tool.Name())

	// Test Description
	assert.NotEmpty(t, tool.Description())

	// Test Options
	options := tool.Options()
	assert.NotNil(t, options)

	// Test ConvertActionToCallToolRequest
	action := option.MobileAction{
		Method: option.ACTION_ListPackages,
		Params: nil,
	}
	request, err := tool.ConvertActionToCallToolRequest(action)
	assert.NoError(t, err)
	assert.Equal(t, string(option.ACTION_ListPackages), request.Params.Name)
	assert.Empty(t, request.GetArguments())
}

// TestToolLaunchApp tests the ToolLaunchApp implementation
func TestToolLaunchApp(t *testing.T) {
	tool := &ToolLaunchApp{}

	// Test Name
	assert.Equal(t, option.ACTION_AppLaunch, tool.Name())

	// Test Description
	assert.NotEmpty(t, tool.Description())

	// Test Options
	options := tool.Options()
	assert.NotNil(t, options)

	// Test ConvertActionToCallToolRequest with valid params
	action := option.MobileAction{
		Method: option.ACTION_AppLaunch,
		Params: "com.example.app",
	}
	request, err := tool.ConvertActionToCallToolRequest(action)
	assert.NoError(t, err)
	assert.Equal(t, string(option.ACTION_AppLaunch), request.Params.Name)
	assert.Equal(t, "com.example.app", request.GetArguments()["packageName"])

	// Test ConvertActionToCallToolRequest with invalid params
	invalidAction := option.MobileAction{
		Method: option.ACTION_AppLaunch,
		Params: 123, // should be string
	}
	_, err = tool.ConvertActionToCallToolRequest(invalidAction)
	assert.Error(t, err)
}

// TestToolTerminateApp tests the ToolTerminateApp implementation
func TestToolTerminateApp(t *testing.T) {
	tool := &ToolTerminateApp{}

	// Test Name
	assert.Equal(t, option.ACTION_AppTerminate, tool.Name())

	// Test Description
	assert.NotEmpty(t, tool.Description())

	// Test Options
	options := tool.Options()
	assert.NotNil(t, options)

	// Test ConvertActionToCallToolRequest with valid params
	action := option.MobileAction{
		Method: option.ACTION_AppTerminate,
		Params: "com.example.app",
	}
	request, err := tool.ConvertActionToCallToolRequest(action)
	assert.NoError(t, err)
	assert.Equal(t, string(option.ACTION_AppTerminate), request.Params.Name)
	assert.Equal(t, "com.example.app", request.GetArguments()["packageName"])

	// Test ConvertActionToCallToolRequest with invalid params
	invalidAction := option.MobileAction{
		Method: option.ACTION_AppTerminate,
		Params: []int{1, 2, 3}, // should be string
	}
	_, err = tool.ConvertActionToCallToolRequest(invalidAction)
	assert.Error(t, err)
}

// TestToolAppInstall tests the ToolAppInstall implementation
func TestToolAppInstall(t *testing.T) {
	tool := &ToolAppInstall{}

	// Test Name
	assert.Equal(t, option.ACTION_AppInstall, tool.Name())

	// Test Description
	assert.NotEmpty(t, tool.Description())

	// Test Options
	options := tool.Options()
	assert.NotNil(t, options)

	// Test ConvertActionToCallToolRequest with valid params
	action := option.MobileAction{
		Method: option.ACTION_AppInstall,
		Params: "https://example.com/app.apk",
	}
	request, err := tool.ConvertActionToCallToolRequest(action)
	assert.NoError(t, err)
	assert.Equal(t, string(option.ACTION_AppInstall), request.Params.Name)
	assert.Equal(t, "https://example.com/app.apk", request.GetArguments()["appUrl"])

	// Test ConvertActionToCallToolRequest with invalid params
	invalidAction := option.MobileAction{
		Method: option.ACTION_AppInstall,
		Params: 123, // should be string
	}
	_, err = tool.ConvertActionToCallToolRequest(invalidAction)
	assert.Error(t, err)
}

// TestToolAppUninstall tests the ToolAppUninstall implementation
func TestToolAppUninstall(t *testing.T) {
	tool := &ToolAppUninstall{}

	// Test Name
	assert.Equal(t, option.ACTION_AppUninstall, tool.Name())

	// Test Description
	assert.NotEmpty(t, tool.Description())

	// Test Options
	options := tool.Options()
	assert.NotNil(t, options)

	// Test ConvertActionToCallToolRequest with valid params
	action := option.MobileAction{
		Method: option.ACTION_AppUninstall,
		Params: "com.example.app",
	}
	request, err := tool.ConvertActionToCallToolRequest(action)
	assert.NoError(t, err)
	assert.Equal(t, string(option.ACTION_AppUninstall), request.Params.Name)
	assert.Equal(t, "com.example.app", request.GetArguments()["packageName"])

	// Test ConvertActionToCallToolRequest with invalid params
	invalidAction := option.MobileAction{
		Method: option.ACTION_AppUninstall,
		Params: 123, // should be string
	}
	_, err = tool.ConvertActionToCallToolRequest(invalidAction)
	assert.Error(t, err)
}

// TestToolAppClear tests the ToolAppClear implementation
func TestToolAppClear(t *testing.T) {
	tool := &ToolAppClear{}

	// Test Name
	assert.Equal(t, option.ACTION_AppClear, tool.Name())

	// Test Description
	assert.NotEmpty(t, tool.Description())

	// Test Options
	options := tool.Options()
	assert.NotNil(t, options)

	// Test ConvertActionToCallToolRequest with valid params
	action := option.MobileAction{
		Method: option.ACTION_AppClear,
		Params: "com.example.app",
	}
	request, err := tool.ConvertActionToCallToolRequest(action)
	assert.NoError(t, err)
	assert.Equal(t, string(option.ACTION_AppClear), request.Params.Name)
	assert.Equal(t, "com.example.app", request.GetArguments()["packageName"])

	// Test ConvertActionToCallToolRequest with invalid params
	invalidAction := option.MobileAction{
		Method: option.ACTION_AppClear,
		Params: 123, // should be string
	}
	_, err = tool.ConvertActionToCallToolRequest(invalidAction)
	assert.Error(t, err)
}

// TestToolSleep tests the ToolSleep implementation
func TestToolSleep(t *testing.T) {
	tool := &ToolSleep{}

	// Test Name
	assert.Equal(t, option.ACTION_Sleep, tool.Name())

	// Test Description
	assert.NotEmpty(t, tool.Description())

	// Test Options
	options := tool.Options()
	assert.NotNil(t, options)

	// Test ConvertActionToCallToolRequest with valid params
	action := option.MobileAction{
		Method: option.ACTION_Sleep,
		Params: 2.5,
	}
	request, err := tool.ConvertActionToCallToolRequest(action)
	assert.NoError(t, err)
	assert.Equal(t, string(option.ACTION_Sleep), request.Params.Name)
	assert.Equal(t, 2.5, request.GetArguments()["seconds"])
}

// TestToolSleepMS tests the ToolSleepMS implementation
func TestToolSleepMS(t *testing.T) {
	tool := &ToolSleepMS{}

	// Test Name
	assert.Equal(t, option.ACTION_SleepMS, tool.Name())

	// Test Description
	assert.NotEmpty(t, tool.Description())

	// Test Options
	options := tool.Options()
	assert.NotNil(t, options)

	// Test ConvertActionToCallToolRequest with valid params
	action := option.MobileAction{
		Method: option.ACTION_SleepMS,
		Params: int64(1500),
	}
	request, err := tool.ConvertActionToCallToolRequest(action)
	assert.NoError(t, err)
	assert.Equal(t, string(option.ACTION_SleepMS), request.Params.Name)
	assert.Equal(t, int64(1500), request.GetArguments()["milliseconds"])

	// Test ConvertActionToCallToolRequest with invalid params
	invalidAction := option.MobileAction{
		Method: option.ACTION_SleepMS,
		Params: "invalid", // should be int64
	}
	_, err = tool.ConvertActionToCallToolRequest(invalidAction)
	assert.Error(t, err)
}

// TestToolSleepRandom tests the ToolSleepRandom implementation
func TestToolSleepRandom(t *testing.T) {
	tool := &ToolSleepRandom{}

	// Test Name
	assert.Equal(t, option.ACTION_SleepRandom, tool.Name())

	// Test Description
	assert.NotEmpty(t, tool.Description())

	// Test Options
	options := tool.Options()
	assert.NotNil(t, options)

	// Test ConvertActionToCallToolRequest with valid params
	action := option.MobileAction{
		Method: option.ACTION_SleepRandom,
		Params: []float64{1.0, 3.0},
	}
	request, err := tool.ConvertActionToCallToolRequest(action)
	assert.NoError(t, err)
	assert.Equal(t, string(option.ACTION_SleepRandom), request.Params.Name)

	params, ok := request.GetArguments()["params"].([]float64)
	require.True(t, ok)
	assert.Equal(t, []float64{1.0, 3.0}, params)

	// Test ConvertActionToCallToolRequest with invalid params
	invalidAction := option.MobileAction{
		Method: option.ACTION_SleepRandom,
		Params: "invalid", // should be []float64
	}
	_, err = tool.ConvertActionToCallToolRequest(invalidAction)
	assert.Error(t, err)
}

// TestToolSetIme tests the ToolSetIme implementation
func TestToolSetIme(t *testing.T) {
	tool := &ToolSetIme{}

	// Test Name
	assert.Equal(t, option.ACTION_SetIme, tool.Name())

	// Test Description
	assert.NotEmpty(t, tool.Description())

	// Test Options
	options := tool.Options()
	assert.NotNil(t, options)

	// Test ConvertActionToCallToolRequest with valid params
	action := option.MobileAction{
		Method: option.ACTION_SetIme,
		Params: "com.google.android.inputmethod.latin",
	}
	request, err := tool.ConvertActionToCallToolRequest(action)
	assert.NoError(t, err)
	assert.Equal(t, string(option.ACTION_SetIme), request.Params.Name)
	assert.Equal(t, "com.google.android.inputmethod.latin", request.GetArguments()["ime"])

	// Test ConvertActionToCallToolRequest with invalid params
	invalidAction := option.MobileAction{
		Method: option.ACTION_SetIme,
		Params: 123, // should be string
	}
	_, err = tool.ConvertActionToCallToolRequest(invalidAction)
	assert.Error(t, err)
}

// TestToolGetSource tests the ToolGetSource implementation
func TestToolGetSource(t *testing.T) {
	tool := &ToolGetSource{}

	// Test Name
	assert.Equal(t, option.ACTION_GetSource, tool.Name())

	// Test Description
	assert.NotEmpty(t, tool.Description())

	// Test Options
	options := tool.Options()
	assert.NotNil(t, options)

	// Test ConvertActionToCallToolRequest with valid params
	action := option.MobileAction{
		Method: option.ACTION_GetSource,
		Params: "com.example.app",
	}
	request, err := tool.ConvertActionToCallToolRequest(action)
	assert.NoError(t, err)
	assert.Equal(t, string(option.ACTION_GetSource), request.Params.Name)
	assert.Equal(t, "com.example.app", request.GetArguments()["packageName"])

	// Test ConvertActionToCallToolRequest with invalid params
	invalidAction := option.MobileAction{
		Method: option.ACTION_GetSource,
		Params: 123, // should be string
	}
	_, err = tool.ConvertActionToCallToolRequest(invalidAction)
	assert.Error(t, err)
}

// TestToolClosePopups tests the ToolClosePopups implementation
func TestToolClosePopups(t *testing.T) {
	tool := &ToolClosePopups{}

	// Test Name
	assert.Equal(t, option.ACTION_ClosePopups, tool.Name())

	// Test Description
	assert.NotEmpty(t, tool.Description())

	// Test Options
	options := tool.Options()
	assert.NotNil(t, options)

	// Test ConvertActionToCallToolRequest
	action := option.MobileAction{
		Method: option.ACTION_ClosePopups,
		Params: nil,
	}
	request, err := tool.ConvertActionToCallToolRequest(action)
	assert.NoError(t, err)
	assert.Equal(t, string(option.ACTION_ClosePopups), request.Params.Name)
	assert.Empty(t, request.GetArguments())
}

// TestToolAIAction tests the ToolAIAction implementation
func TestToolAIAction(t *testing.T) {
	tool := &ToolAIAction{}

	// Test Name
	assert.Equal(t, option.ACTION_AIAction, tool.Name())

	// Test Description
	assert.NotEmpty(t, tool.Description())

	// Test Options
	options := tool.Options()
	assert.NotNil(t, options)

	// Test ConvertActionToCallToolRequest with valid params
	action := option.MobileAction{
		Method: option.ACTION_AIAction,
		Params: "Click on the login button",
	}
	request, err := tool.ConvertActionToCallToolRequest(action)
	assert.NoError(t, err)
	assert.Equal(t, string(option.ACTION_AIAction), request.Params.Name)
	assert.Equal(t, "Click on the login button", request.GetArguments()["prompt"])

	// Test ConvertActionToCallToolRequest with invalid params
	invalidAction := option.MobileAction{
		Method: option.ACTION_AIAction,
		Params: 123, // should be string
	}
	_, err = tool.ConvertActionToCallToolRequest(invalidAction)
	assert.Error(t, err)
}

// TestToolAIQuery tests the ToolAIQuery implementation
func TestToolAIQuery(t *testing.T) {
	tool := &ToolAIQuery{}

	// Test Name
	assert.Equal(t, option.ACTION_Query, tool.Name())

	// Test Description
	assert.NotEmpty(t, tool.Description())

	// Test Options
	options := tool.Options()
	assert.NotNil(t, options)

	// Test ConvertActionToCallToolRequest with valid params
	action := option.MobileAction{
		Method: option.ACTION_Query,
		Params: "What is displayed on the screen?",
	}
	request, err := tool.ConvertActionToCallToolRequest(action)
	assert.NoError(t, err)
	assert.Equal(t, string(option.ACTION_Query), request.Params.Name)
	assert.Equal(t, "What is displayed on the screen?", request.GetArguments()["prompt"])

	// Test ConvertActionToCallToolRequest with invalid params
	invalidAction := option.MobileAction{
		Method: option.ACTION_Query,
		Params: 123, // should be string
	}
	_, err = tool.ConvertActionToCallToolRequest(invalidAction)
	assert.Error(t, err)
}

// TestToolFinished tests the ToolFinished implementation
func TestToolFinished(t *testing.T) {
	tool := &ToolFinished{}

	// Test Name
	assert.Equal(t, option.ACTION_Finished, tool.Name())

	// Test Description
	assert.NotEmpty(t, tool.Description())

	// Test Options
	options := tool.Options()
	assert.NotNil(t, options)

	// Test ConvertActionToCallToolRequest with valid params
	action := option.MobileAction{
		Method: option.ACTION_Finished,
		Params: "Task completed successfully",
	}
	request, err := tool.ConvertActionToCallToolRequest(action)
	assert.NoError(t, err)
	assert.Equal(t, string(option.ACTION_Finished), request.Params.Name)
	assert.Equal(t, "Task completed successfully", request.GetArguments()["content"])

	// Test ConvertActionToCallToolRequest with invalid params
	invalidAction := option.MobileAction{
		Method: option.ACTION_Finished,
		Params: 123, // should be string
	}
	_, err = tool.ConvertActionToCallToolRequest(invalidAction)
	assert.Error(t, err)
}

// TestToolWebLoginNoneUI tests the ToolWebLoginNoneUI implementation
func TestToolWebLoginNoneUI(t *testing.T) {
	tool := &ToolWebLoginNoneUI{}

	// Test Name
	assert.Equal(t, option.ACTION_WebLoginNoneUI, tool.Name())

	// Test Description
	assert.NotEmpty(t, tool.Description())

	// Test Options
	options := tool.Options()
	assert.NotNil(t, options)

	// Test ConvertActionToCallToolRequest
	action := option.MobileAction{
		Method: option.ACTION_WebLoginNoneUI,
		Params: nil,
	}
	request, err := tool.ConvertActionToCallToolRequest(action)
	assert.NoError(t, err)
	assert.Equal(t, string(option.ACTION_WebLoginNoneUI), request.Params.Name)
	assert.Empty(t, request.GetArguments())
}

// TestToolSecondaryClick tests the ToolSecondaryClick implementation
func TestToolSecondaryClick(t *testing.T) {
	tool := &ToolSecondaryClick{}

	// Test Name
	assert.Equal(t, option.ACTION_SecondaryClick, tool.Name())

	// Test Description
	assert.NotEmpty(t, tool.Description())

	// Test Options
	options := tool.Options()
	assert.NotNil(t, options)

	// Test ConvertActionToCallToolRequest with valid params
	action := option.MobileAction{
		Method: option.ACTION_SecondaryClick,
		Params: []float64{0.5, 0.6},
	}
	request, err := tool.ConvertActionToCallToolRequest(action)
	assert.NoError(t, err)
	assert.Equal(t, string(option.ACTION_SecondaryClick), request.Params.Name)
	args := request.GetArguments()
	assert.Equal(t, 0.5, args["x"])
	assert.Equal(t, 0.6, args["y"])

	// Test ConvertActionToCallToolRequest with invalid params
	invalidAction := option.MobileAction{
		Method: option.ACTION_SecondaryClick,
		Params: "invalid", // should be []float64
	}
	_, err = tool.ConvertActionToCallToolRequest(invalidAction)
	assert.Error(t, err)
}

// TestToolHoverBySelector tests the ToolHoverBySelector implementation
func TestToolHoverBySelector(t *testing.T) {
	tool := &ToolHoverBySelector{}

	// Test Name
	assert.Equal(t, option.ACTION_HoverBySelector, tool.Name())

	// Test Description
	assert.NotEmpty(t, tool.Description())

	// Test Options
	options := tool.Options()
	assert.NotNil(t, options)

	// Test ConvertActionToCallToolRequest with valid params
	action := option.MobileAction{
		Method: option.ACTION_HoverBySelector,
		Params: "#login-button",
	}
	request, err := tool.ConvertActionToCallToolRequest(action)
	assert.NoError(t, err)
	assert.Equal(t, string(option.ACTION_HoverBySelector), request.Params.Name)
	assert.Equal(t, "#login-button", request.GetArguments()["selector"])

	// Test ConvertActionToCallToolRequest with invalid params
	invalidAction := option.MobileAction{
		Method: option.ACTION_HoverBySelector,
		Params: 123, // should be string
	}
	_, err = tool.ConvertActionToCallToolRequest(invalidAction)
	assert.Error(t, err)
}

// TestToolTapBySelector tests the ToolTapBySelector implementation
func TestToolTapBySelector(t *testing.T) {
	tool := &ToolTapBySelector{}

	// Test Name
	assert.Equal(t, option.ACTION_TapBySelector, tool.Name())

	// Test Description
	assert.NotEmpty(t, tool.Description())

	// Test Options
	options := tool.Options()
	assert.NotNil(t, options)

	// Test ConvertActionToCallToolRequest with valid params
	action := option.MobileAction{
		Method: option.ACTION_TapBySelector,
		Params: "//button[@id='submit']",
	}
	request, err := tool.ConvertActionToCallToolRequest(action)
	assert.NoError(t, err)
	assert.Equal(t, string(option.ACTION_TapBySelector), request.Params.Name)
	assert.Equal(t, "//button[@id='submit']", request.GetArguments()["selector"])

	// Test ConvertActionToCallToolRequest with invalid params
	invalidAction := option.MobileAction{
		Method: option.ACTION_TapBySelector,
		Params: 123, // should be string
	}
	_, err = tool.ConvertActionToCallToolRequest(invalidAction)
	assert.Error(t, err)
}

// TestToolSecondaryClickBySelector tests the ToolSecondaryClickBySelector implementation
func TestToolSecondaryClickBySelector(t *testing.T) {
	tool := &ToolSecondaryClickBySelector{}

	// Test Name
	assert.Equal(t, option.ACTION_SecondaryClickBySelector, tool.Name())

	// Test Description
	assert.NotEmpty(t, tool.Description())

	// Test Options
	options := tool.Options()
	assert.NotNil(t, options)

	// Test ConvertActionToCallToolRequest with valid params
	action := option.MobileAction{
		Method: option.ACTION_SecondaryClickBySelector,
		Params: ".context-menu-trigger",
	}
	request, err := tool.ConvertActionToCallToolRequest(action)
	assert.NoError(t, err)
	assert.Equal(t, string(option.ACTION_SecondaryClickBySelector), request.Params.Name)
	assert.Equal(t, ".context-menu-trigger", request.GetArguments()["selector"])

	// Test ConvertActionToCallToolRequest with invalid params
	invalidAction := option.MobileAction{
		Method: option.ACTION_SecondaryClickBySelector,
		Params: 123, // should be string
	}
	_, err = tool.ConvertActionToCallToolRequest(invalidAction)
	assert.Error(t, err)
}

// TestToolWebCloseTab tests the ToolWebCloseTab implementation
func TestToolWebCloseTab(t *testing.T) {
	tool := &ToolWebCloseTab{}

	// Test Name
	assert.Equal(t, option.ACTION_WebCloseTab, tool.Name())

	// Test Description
	assert.NotEmpty(t, tool.Description())

	// Test Options
	options := tool.Options()
	assert.NotNil(t, options)

	// Test ConvertActionToCallToolRequest with valid params
	action := option.MobileAction{
		Method: option.ACTION_WebCloseTab,
		Params: 1,
	}
	request, err := tool.ConvertActionToCallToolRequest(action)
	assert.NoError(t, err)
	assert.Equal(t, string(option.ACTION_WebCloseTab), request.Params.Name)
	assert.Equal(t, 1, request.GetArguments()["tabIndex"])

	// Test ConvertActionToCallToolRequest with invalid params
	invalidAction := option.MobileAction{
		Method: option.ACTION_WebCloseTab,
		Params: "invalid", // should be int
	}
	_, err = tool.ConvertActionToCallToolRequest(invalidAction)
	assert.Error(t, err)
}

func TestPreMarkOperationConfiguration(t *testing.T) {
	// Test that pre_mark_operation is configurable and not hardcoded
	server := NewMCPServer()

	// Get the tap_xy tool
	tapTool := server.GetToolByAction(option.ACTION_TapXY)
	assert.NotNil(t, tapTool)

	// Test conversion with pre_mark_operation enabled
	actionWithPreMark := option.MobileAction{
		Method:        option.ACTION_TapXY,
		Params:        []float64{0.5, 0.5},
		ActionOptions: *option.NewActionOptions(option.WithPreMarkOperation(true)),
	}

	request, err := tapTool.ConvertActionToCallToolRequest(actionWithPreMark)
	assert.NoError(t, err)
	assert.Equal(t, true, request.GetArguments()["pre_mark_operation"])

	// Test conversion without pre_mark_operation
	actionWithoutPreMark := option.MobileAction{
		Method:        option.ACTION_TapXY,
		Params:        []float64{0.5, 0.5},
		ActionOptions: *option.NewActionOptions(option.WithPreMarkOperation(false)),
	}

	request2, err := tapTool.ConvertActionToCallToolRequest(actionWithoutPreMark)
	assert.NoError(t, err)
	// Should not have pre_mark_operation in arguments when false
	_, exists := request2.GetArguments()["pre_mark_operation"]
	assert.False(t, exists)
}

func TestGenerateReturnSchema(t *testing.T) {
	// Test with ToolListPackages
	tool := ToolListPackages{}
	schema := GenerateReturnSchema(tool)

	// Check that standard MCPResponse fields are included
	assert.Contains(t, schema, "action")
	assert.Contains(t, schema, "success")
	assert.Contains(t, schema, "message")
	assert.Equal(t, "string: Action performed", schema["action"])
	assert.Equal(t, "boolean: Whether the operation was successful", schema["success"])
	assert.Equal(t, "string: Human-readable message describing the result", schema["message"])

	// Check that tool-specific fields are included at the same level
	assert.Contains(t, schema, "packages")
	assert.Contains(t, schema, "count")
	assert.Equal(t, "[]string: List of installed app package names on the device", schema["packages"])
	assert.Equal(t, "int: Number of installed packages", schema["count"])

	// Ensure "data" field is not present in the new flat structure
	assert.NotContains(t, schema, "data")
}

func TestMCPResponseInheritance(t *testing.T) {
	// Test creating a response with tool data
	returnData := ToolListPackages{
		Packages: []string{"com.example.app1", "com.example.app2"},
		Count:    2,
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(returnData)
	assert.NoError(t, err)

	// Parse back to verify structure
	var parsed map[string]interface{}
	err = json.Unmarshal(jsonData, &parsed)
	assert.NoError(t, err)

	// Check that tool-specific fields are present
	assert.Equal(t, float64(2), parsed["count"]) // JSON numbers are float64

	packages, ok := parsed["packages"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, packages, 2)
	assert.Equal(t, "com.example.app1", packages[0])
	assert.Equal(t, "com.example.app2", packages[1])
}

func TestNewMCPSuccessResponse(t *testing.T) {
	// Test the simplified NewMCPSuccessResponse function
	message := "Successfully slept for 5 seconds"
	returnData := ToolSleep{
		Seconds:  5.0,
		Duration: "5s",
	}

	// Test JSON marshaling directly first
	jsonData, err := json.Marshal(returnData)
	assert.NoError(t, err)

	// Parse the JSON to verify structure
	var parsed map[string]interface{}
	err = json.Unmarshal(jsonData, &parsed)
	assert.NoError(t, err)

	assert.Equal(t, float64(5.0), parsed["seconds"])
	assert.Equal(t, "5s", parsed["duration"])

	// Test the MCP response function with actual tool instance
	tool := &ToolSleep{}
	result := NewMCPSuccessResponse(message, tool)
	assert.NotNil(t, result)
}

func TestNewMCPErrorResponse(t *testing.T) {
	// Test error response creation
	result := NewMCPErrorResponse("Test error message")
	assert.NotNil(t, result)
}
