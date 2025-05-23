package ai

import (
	"testing"

	"github.com/httprunner/httprunner/v5/internal/json"
	"github.com/httprunner/httprunner/v5/uixt/types"
	"github.com/stretchr/testify/assert"
)

func TestParseActionToStructureOutput(t *testing.T) {
	text := "Thought: test\nAction: click(point='<point>200 300</point>')"
	parser := &UITARSContentParser{}
	result, err := parser.Parse(text, types.Size{Height: 224, Width: 224})
	assert.Nil(t, err)
	function := result.ToolCalls[0].Function
	assert.Equal(t, function.Name, "click")
	assert.Contains(t, function.Arguments, "start_box")

	text = "Thought: 我看到页面上有几个帖子，第二个帖子的标题是\"字节四年，头发白了\"。要完成任务，我需要点击这个帖子下方的作者头像，这样就能进入作者的个人主页了。\nAction: click(start_point='<point>550 450 550 450</point>')"
	result, err = parser.Parse(text, types.Size{Height: 2341, Width: 1024})
	assert.Nil(t, err)
	function = result.ToolCalls[0].Function
	assert.Equal(t, function.Name, "click")
	assert.Contains(t, function.Arguments, "start_box")

	// Test new bracket format
	text = "Thought: 我需要点击这个按钮\nAction: click(start_box='[100, 200, 150, 250]')"
	result, err = parser.Parse(text, types.Size{Height: 1000, Width: 1000})
	assert.Nil(t, err)
	function = result.ToolCalls[0].Function
	assert.Equal(t, function.Name, "click")
	assert.Contains(t, function.Arguments, "start_box")
	arguments := make(map[string]interface{})
	err = json.Unmarshal([]byte(function.Arguments), &arguments)
	assert.Nil(t, err)
	coordsInterface := arguments["start_box"].([]interface{})
	coords := make([]float64, len(coordsInterface))
	for i, v := range coordsInterface {
		coords[i] = v.(float64)
	}
	assert.Equal(t, 4, len(coords))
	assert.Equal(t, 100.0, coords[0])
	assert.Equal(t, 200.0, coords[1])
	assert.Equal(t, 150.0, coords[2])
	assert.Equal(t, 250.0, coords[3])

	// Test drag operation with both start_box and end_box
	text = "Thought: 我需要拖拽元素\nAction: drag(start_box='[100, 200, 150, 250]', end_box='[300, 400, 350, 450]')"
	result, err = parser.Parse(text, types.Size{Height: 1000, Width: 1000})
	assert.Nil(t, err)
	function = result.ToolCalls[0].Function
	assert.Equal(t, function.Name, "drag")
	assert.Contains(t, function.Arguments, "start_box")
	assert.Contains(t, function.Arguments, "end_box")
	arguments = make(map[string]interface{})
	err = json.Unmarshal([]byte(function.Arguments), &arguments)
	assert.Nil(t, err)
	startCoordsInterface := arguments["start_box"].([]interface{})
	endCoordsInterface := arguments["end_box"].([]interface{})
	startCoords := make([]float64, len(startCoordsInterface))
	endCoords := make([]float64, len(endCoordsInterface))
	for i, v := range startCoordsInterface {
		startCoords[i] = v.(float64)
	}
	for i, v := range endCoordsInterface {
		endCoords[i] = v.(float64)
	}
	assert.Equal(t, 4, len(startCoords))
	assert.Equal(t, 4, len(endCoords))
	assert.Equal(t, 100.0, startCoords[0])
	assert.Equal(t, 300.0, endCoords[0])
}

// Test normalizeCoordinatesFormat function
func TestNormalizeCoordinatesFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "point tag with 2 numbers",
			input:    "<point>100 200</point>",
			expected: "(100,200)",
		},
		{
			name:     "point tag with 4 numbers",
			input:    "<point>100 200 150 250</point>",
			expected: "(100,200,150,250)",
		},
		{
			name:     "bbox tag",
			input:    "<bbox>100 200 150 250</bbox>",
			expected: "(100,200,150,250)",
		},
		{
			name:     "bracket format with spaces",
			input:    "[100, 200, 150, 250]",
			expected: "(100,200,150,250)",
		},
		{
			name:     "bracket format without spaces",
			input:    "[100,200,150,250]",
			expected: "(100,200,150,250)",
		},
		{
			name:     "bracket format with irregular spaces",
			input:    "[100,  200,  150,  250]",
			expected: "(100,200,150,250)",
		},
		{
			name:     "multiple point tags",
			input:    "<point>100 200</point> and <point>300 400</point>",
			expected: "(100,200) and (300,400)",
		},
		{
			name:     "mixed formats",
			input:    "<point>100 200</point> and [300, 400, 350, 450]",
			expected: "(100,200) and (300,400,350,450)",
		},
		{
			name:     "documentation_example_coordinates",
			input:    "<point>235 512</point>",
			expected: "(235,512)",
		},
		{
			name:     "documentation_example_bbox",
			input:    "<bbox>235 512 451 553</bbox>",
			expected: "(235,512,451,553)",
		},
		{
			name:     "mobile_coordinates_point",
			input:    "<point>200 600</point>",
			expected: "(200,600)",
		},
		{
			name:     "tablet_coordinates_bbox",
			input:    "<bbox>750 400 800 450</bbox>",
			expected: "(750,400,800,450)",
		},
		// Note: Bracket format with 2 coordinates is NOT supported by the function
		// Only 4-coordinate bracket format is supported
		{
			name:     "bracket_format_two_coordinates_not_converted",
			input:    "[100, 200]",
			expected: "[100, 200]", // Function doesn't convert this format
		},
		// Note: Decimal coordinates are NOT supported by the regex (only \d+ is matched)
		{
			name:     "point_tag_with_decimals_not_converted",
			input:    "<point>100.5 200.7</point>",
			expected: "<point>100.5 200.7</point>", // Function doesn't convert decimals
		},
		{
			name:     "bbox_tag_with_decimals_not_converted",
			input:    "<bbox>100.5 200.7 150.3 250.9</bbox>",
			expected: "<bbox>100.5 200.7 150.3 250.9</bbox>", // Function doesn't convert decimals
		},
		{
			name:     "bracket_format_with_decimals_not_converted",
			input:    "[100.5, 200.7, 150.3, 250.9]",
			expected: "[100.5, 200.7, 150.3, 250.9]", // Function doesn't convert decimals
		},
		{
			name:     "multiple_bracket_formats",
			input:    "[100, 200] and [300, 400, 350, 450]",
			expected: "[100, 200] and (300,400,350,450)", // Only 4-coord format converted
		},
		{
			name:     "multiple_bbox_tags",
			input:    "<bbox>100 200 150 250</bbox> then <bbox>300 400 350 450</bbox>",
			expected: "(100,200,150,250) then (300,400,350,450)",
		},
		{
			name:     "edge_case_zero_coordinates",
			input:    "<point>0 0</point>",
			expected: "(0,0)",
		},
		{
			name:     "edge_case_maximum_coordinates",
			input:    "<point>1000 1000</point>",
			expected: "(1000,1000)",
		},
		{
			name:     "complex_mixed_formats",
			input:    "click <point>100 200</point> then drag [300, 400, 350, 450] to <bbox>500 600 550 650</bbox>",
			expected: "click (100,200) then drag (300,400,350,450) to (500,600,550,650)",
		},
		{
			name:     "no_coordinates",
			input:    "click on button",
			expected: "click on button",
		},
		{
			name:     "empty_string",
			input:    "",
			expected: "",
		},
		{
			name:     "only_text_no_tags",
			input:    "some random text without coordinates",
			expected: "some random text without coordinates",
		},
		// Note: Extra spaces in brackets with 4 coords are NOT handled properly by the regex
		{
			name:     "bracket_format_with_extra_spaces_not_converted",
			input:    "[  100  ,  200  ,  150  ,  250  ]",
			expected: "[  100  ,  200  ,  150  ,  250  ]", // Function regex doesn't handle extra spaces
		},
		{
			name:     "large_coordinates",
			input:    "<point>1920 1080</point>",
			expected: "(1920,1080)",
		},
		{
			name:     "ultrawide_coordinates",
			input:    "<bbox>0 0 3440 1440</bbox>",
			expected: "(0,0,3440,1440)",
		},
		{
			name:     "real_world_action_example",
			input:    "Action: click(start_box='<point>235 512</point>')",
			expected: "Action: click(start_box='(235,512)')",
		},
		{
			name:     "real_world_drag_example",
			input:    "Action: drag(start_box='[100, 200, 150, 250]', end_box='<bbox>300 400 350 450</bbox>')",
			expected: "Action: drag(start_box='(100,200,150,250)', end_box='(300,400,350,450)')",
		},
		{
			name:     "real_world_example_1",
			input:    "<point>235 512</point>",
			expected: "(235,512)", // Should be string format for normalizeCoordinatesFormat
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeCoordinatesFormat(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test convertRelativeToAbsolute function
func TestConvertRelativeToAbsolute(t *testing.T) {
	tests := []struct {
		name           string
		size           types.Size
		relativeCoord  float64
		isXCoord       bool
		expectedResult float64
		description    string
	}{
		{
			name:           "standard_1000x2000_x_coordinate",
			size:           types.Size{Width: 1000, Height: 2000},
			relativeCoord:  500, // 500/1000 * 1000 = 500
			isXCoord:       true,
			expectedResult: 500.0,
			description:    "Standard case: X coordinate conversion",
		},
		{
			name:           "standard_1000x2000_y_coordinate",
			size:           types.Size{Width: 1000, Height: 2000},
			relativeCoord:  500, // 500/1000 * 2000 = 1000
			isXCoord:       false,
			expectedResult: 1000.0,
			description:    "Standard case: Y coordinate conversion",
		},
		{
			name:           "example_from_documentation_x",
			size:           types.Size{Width: 1920, Height: 1080},
			relativeCoord:  235, // round(1920*235/1000) = 451
			isXCoord:       true,
			expectedResult: 451.2, // 实际计算值为451.2，测试精确值
			description:    "Documentation example: X coordinate (235, 512) on 1920x1080",
		},
		{
			name:           "example_from_documentation_y",
			size:           types.Size{Width: 1920, Height: 1080},
			relativeCoord:  512, // round(1080*512/1000) = 553
			isXCoord:       false,
			expectedResult: 553.0, // 实际计算值为553.0
			description:    "Documentation example: Y coordinate (235, 512) on 1920x1080",
		},
		{
			name:           "mobile_device_x_coordinate",
			size:           types.Size{Width: 375, Height: 812},
			relativeCoord:  200, // 200/1000 * 375 = 75
			isXCoord:       true,
			expectedResult: 75.0,
			description:    "Mobile device: iPhone X size X coordinate",
		},
		{
			name:           "mobile_device_y_coordinate",
			size:           types.Size{Width: 375, Height: 812},
			relativeCoord:  600, // 600/1000 * 812 = 487.2
			isXCoord:       false,
			expectedResult: 487.2,
			description:    "Mobile device: iPhone X size Y coordinate",
		},
		{
			name:           "tablet_device_x_coordinate",
			size:           types.Size{Width: 1024, Height: 768},
			relativeCoord:  750, // 750/1000 * 1024 = 768
			isXCoord:       true,
			expectedResult: 768.0,
			description:    "Tablet device: iPad size X coordinate",
		},
		{
			name:           "tablet_device_y_coordinate",
			size:           types.Size{Width: 1024, Height: 768},
			relativeCoord:  400, // 400/1000 * 768 = 307.2
			isXCoord:       false,
			expectedResult: 307.2,
			description:    "Tablet device: iPad size Y coordinate",
		},
		{
			name:           "edge_case_zero_coordinate",
			size:           types.Size{Width: 1920, Height: 1080},
			relativeCoord:  0, // 0/1000 * width/height = 0
			isXCoord:       true,
			expectedResult: 0.0,
			description:    "Edge case: Zero coordinate",
		},
		{
			name:           "edge_case_maximum_coordinate_x",
			size:           types.Size{Width: 1920, Height: 1080},
			relativeCoord:  1000, // 1000/1000 * 1920 = 1920
			isXCoord:       true,
			expectedResult: 1920.0,
			description:    "Edge case: Maximum X coordinate (1000 -> full width)",
		},
		{
			name:           "edge_case_maximum_coordinate_y",
			size:           types.Size{Width: 1920, Height: 1080},
			relativeCoord:  1000, // 1000/1000 * 1080 = 1080
			isXCoord:       false,
			expectedResult: 1080.0,
			description:    "Edge case: Maximum Y coordinate (1000 -> full height)",
		},
		{
			name:           "rounding_precision_test_x",
			size:           types.Size{Width: 1000, Height: 1000},
			relativeCoord:  333, // 333/1000 * 1000 = 333
			isXCoord:       true,
			expectedResult: 333.0,
			description:    "Precision test: X coordinate with rounding",
		},
		{
			name:           "rounding_precision_test_y",
			size:           types.Size{Width: 1000, Height: 2000},
			relativeCoord:  750, // 750/1000 * 2000 = 1500
			isXCoord:       false,
			expectedResult: 1500.0,
			description:    "Precision test: Y coordinate with rounding",
		},
		{
			name:           "small_screen_x_coordinate",
			size:           types.Size{Width: 480, Height: 800},
			relativeCoord:  125, // 125/1000 * 480 = 60
			isXCoord:       true,
			expectedResult: 60.0,
			description:    "Small screen: X coordinate conversion",
		},
		{
			name:           "small_screen_y_coordinate",
			size:           types.Size{Width: 480, Height: 800},
			relativeCoord:  875, // 875/1000 * 800 = 700
			isXCoord:       false,
			expectedResult: 700.0,
			description:    "Small screen: Y coordinate conversion",
		},
		{
			name:           "ultrawide_monitor_x_coordinate",
			size:           types.Size{Width: 3440, Height: 1440},
			relativeCoord:  450, // 450/1000 * 3440 = 1548
			isXCoord:       true,
			expectedResult: 1548.0,
			description:    "Ultrawide monitor: X coordinate conversion",
		},
		{
			name:           "ultrawide_monitor_y_coordinate",
			size:           types.Size{Width: 3440, Height: 1440},
			relativeCoord:  720, // 720/1000 * 1440 = 1036.8
			isXCoord:       false,
			expectedResult: 1036.8,
			description:    "Ultrawide monitor: Y coordinate conversion",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertRelativeToAbsolute(tt.relativeCoord, tt.isXCoord, tt.size)
			assert.Equal(t, tt.expectedResult, result, "Test case: %s", tt.description)
		})
	}
}

// Test parseActionTypeAndArguments function
func TestParseActionTypeAndArguments(t *testing.T) {
	tests := []struct {
		name         string
		actionStr    string
		expectedType string
		expectedArgs map[string]interface{}
		expectError  bool
	}{
		{
			name:         "simple click action",
			actionStr:    "click(start_box='100,200,150,250')",
			expectedType: "click",
			expectedArgs: map[string]interface{}{
				"start_box": "100,200,150,250",
			},
			expectError: false,
		},
		{
			name:         "drag action with two parameters",
			actionStr:    "drag(start_box='100,200,150,250', end_box='300,400,350,450')",
			expectedType: "drag",
			expectedArgs: map[string]interface{}{
				"start_box": "100,200,150,250",
				"end_box":   "300,400,350,450",
			},
			expectError: false,
		},
		{
			name:         "parameter name mapping - start_point to start_box",
			actionStr:    "click(start_point='100,200,150,250')",
			expectedType: "click",
			expectedArgs: map[string]interface{}{
				"start_box": "100,200,150,250", // should be mapped from start_point
			},
			expectError: false,
		},
		{
			name:         "parameter name mapping - point to start_box",
			actionStr:    "click(point='100,200')",
			expectedType: "click",
			expectedArgs: map[string]interface{}{
				"start_box": "100,200", // should be mapped from point
			},
			expectError: false,
		},
		{
			name:         "type action with content",
			actionStr:    "type(content='Hello World')",
			expectedType: "type",
			expectedArgs: map[string]interface{}{
				"content": "Hello World",
			},
			expectError: false,
		},
		{
			name:         "action without parameters",
			actionStr:    "press_home()",
			expectedType: "press_home",
			expectedArgs: map[string]interface{}{},
			expectError:  false,
		},
		{
			name:        "invalid format - no parentheses",
			actionStr:   "click",
			expectError: true,
		},
		{
			name:         "invalid format - missing closing parenthesis",
			actionStr:    "click(start_box='100,200'",
			expectedType: "click",
			expectedArgs: map[string]interface{}{
				"start_box": "100,200", // 正则表达式能够匹配到这个参数
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actionType, rawArgs, err := parseActionTypeAndArguments(tt.actionStr)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedType, actionType)
			assert.Equal(t, tt.expectedArgs, rawArgs)
		})
	}
}

// Test normalizeParameterName function
func TestNormalizeParameterName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "start_point to start_box",
			input:    "start_point",
			expected: "start_box",
		},
		{
			name:     "end_point to end_box",
			input:    "end_point",
			expected: "end_box",
		},
		{
			name:     "point to start_box",
			input:    "point",
			expected: "start_box",
		},
		{
			name:     "unchanged parameter",
			input:    "content",
			expected: "content",
		},
		{
			name:     "unchanged parameter - direction",
			input:    "direction",
			expected: "direction",
		},
		{
			name:     "unchanged parameter - start_box",
			input:    "start_box",
			expected: "start_box",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeParameterName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test isCoordinateParameter function
func TestIsCoordinateParameter(t *testing.T) {
	tests := []struct {
		name      string
		paramName string
		expected  bool
	}{
		{
			name:      "start_box is coordinate",
			paramName: "start_box",
			expected:  true,
		},
		{
			name:      "end_box is coordinate",
			paramName: "end_box",
			expected:  true,
		},
		{
			name:      "start_point is coordinate",
			paramName: "start_point",
			expected:  true,
		},
		{
			name:      "end_point is coordinate",
			paramName: "end_point",
			expected:  true,
		},
		{
			name:      "content is not coordinate",
			paramName: "content",
			expected:  false,
		},
		{
			name:      "direction is not coordinate",
			paramName: "direction",
			expected:  false,
		},
		{
			name:      "key is not coordinate",
			paramName: "key",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isCoordinateParameter(tt.paramName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test normalizeStringParam function
func TestNormalizeStringParam(t *testing.T) {
	tests := []struct {
		name       string
		paramName  string
		paramValue interface{}
		expected   interface{}
	}{
		{
			name:       "content with escape characters",
			paramName:  "content",
			paramValue: "Hello\\nWorld\\\"Test\\'",
			expected:   "Hello\nWorld\"Test'",
		},
		{
			name:       "content without escape characters",
			paramName:  "content",
			paramValue: "Hello World",
			expected:   "Hello World",
		},
		{
			name:       "non-content parameter with escape characters",
			paramName:  "direction",
			paramValue: "down\\nup",
			expected:   "down\\nup", // should not process escape chars
		},
		{
			name:       "string with leading/trailing spaces",
			paramName:  "content",
			paramValue: "  Hello World  ",
			expected:   "Hello World",
		},
		{
			name:       "empty string",
			paramName:  "content",
			paramValue: "",
			expected:   "",
		},
		{
			name:       "nil value",
			paramName:  "content",
			paramValue: nil,
			expected:   nil,
		},
		{
			name:       "non-string value",
			paramName:  "content",
			paramValue: 123,
			expected:   123,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeStringParam(tt.paramName, tt.paramValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test normalizeStringCoordinates function
func TestNormalizeStringCoordinates(t *testing.T) {
	tests := []struct {
		name        string
		coordStr    string
		size        types.Size
		expected    []float64
		expectError bool
		description string
	}{
		{
			name:        "simple_coordinate_string",
			coordStr:    "100,200,150,250",
			size:        types.Size{Width: 1000, Height: 1000},
			expected:    []float64{100.0, 200.0, 150.0, 250.0},
			description: "Simple comma-separated coordinate string",
		},
		{
			name:        "coordinate_string_with_spaces",
			coordStr:    " 100 , 200 , 150 , 250 ",
			size:        types.Size{Width: 1000, Height: 1000},
			expected:    []float64{100.0, 200.0, 150.0, 250.0},
			description: "Coordinate string with spaces",
		},
		{
			name:        "documentation_example_point_tag",
			coordStr:    "<point>235 512</point>",
			size:        types.Size{Width: 1920, Height: 1080},
			expected:    []float64{451.2, 553.0}, // 235/1000*1920=451.2, 512/1000*1080=553.0
			description: "Documentation example: point tag on 1920x1080",
		},
		{
			name:        "documentation_example_bbox_tag",
			coordStr:    "<bbox>235 512 451 553</bbox>",
			size:        types.Size{Width: 1920, Height: 1080},
			expected:    []float64{451.2, 553.0, 865.9, 597.2}, // All converted from relative to absolute
			description: "Documentation example: bbox tag on 1920x1080",
		},
		{
			name:        "mobile_device_point",
			coordStr:    "<point>200 600</point>",
			size:        types.Size{Width: 375, Height: 812},
			expected:    []float64{75.0, 487.2}, // 200/1000*375=75, 600/1000*812=487.2
			description: "Mobile device: iPhone X point coordinate",
		},
		{
			name:        "mobile_device_bbox",
			coordStr:    "<bbox>200 600 400 800</bbox>",
			size:        types.Size{Width: 375, Height: 812},
			expected:    []float64{75.0, 487.2, 150.0, 649.6}, // Mobile device bbox
			description: "Mobile device: iPhone X bbox coordinate",
		},
		{
			name:        "tablet_device_coordinates",
			coordStr:    "[750, 400, 800, 450]",
			size:        types.Size{Width: 1024, Height: 768},
			expected:    []float64{768.0, 307.2, 819.2, 345.6}, // Tablet coordinates
			description: "Tablet device: iPad coordinate conversion",
		},
		{
			name:        "bracket_format_two_coords",
			coordStr:    "[100, 200]",
			size:        types.Size{Width: 1000, Height: 1000},
			expected:    []float64{100.0, 200.0},
			description: "Bracket format with two coordinates",
		},
		{
			name:        "bracket_format_four_coords",
			coordStr:    "[100, 200, 150, 250]",
			size:        types.Size{Width: 1000, Height: 1000},
			expected:    []float64{100.0, 200.0, 150.0, 250.0},
			description: "Bracket format with four coordinates",
		},
		{
			name:        "edge_case_zero_coordinates",
			coordStr:    "0,0,0,0",
			size:        types.Size{Width: 1920, Height: 1080},
			expected:    []float64{0.0, 0.0, 0.0, 0.0},
			description: "Edge case: all zero coordinates",
		},
		{
			name:        "edge_case_maximum_coordinates",
			coordStr:    "1000,1000,1000,1000",
			size:        types.Size{Width: 1920, Height: 1080},
			expected:    []float64{1920.0, 1080.0, 1920.0, 1080.0}, // Maximum relative coords -> screen edges
			description: "Edge case: maximum coordinates (1000 -> screen edges)",
		},
		{
			name:        "ultrawide_monitor_coords",
			coordStr:    "<point>450 720</point>",
			size:        types.Size{Width: 3440, Height: 1440},
			expected:    []float64{1548.0, 1036.8}, // 450/1000*3440=1548, 720/1000*1440=1036.8
			description: "Ultrawide monitor: coordinate conversion",
		},
		{
			name:        "small_screen_coordinates",
			coordStr:    "<bbox>125 875 250 950</bbox>",
			size:        types.Size{Width: 480, Height: 800},
			expected:    []float64{60.0, 700.0, 120.0, 760.0}, // Small screen bbox
			description: "Small screen: coordinate conversion",
		},
		{
			name:        "real_world_example_1",
			coordStr:    "<point>235 512</point>",
			size:        types.Size{Width: 1920, Height: 1080},
			expected:    []float64{451.2, 553.0}, // Real documentation example
			description: "Real world: documentation example coordinates",
		},
		{
			name:        "real_world_example_2",
			coordStr:    "[375, 600, 425, 650]",
			size:        types.Size{Width: 1080, Height: 1920},
			expected:    []float64{405.0, 1152.0, 459.0, 1248.0}, // Portrait mobile bbox
			description: "Real world: portrait mobile bbox",
		},
		// Error cases - decimal coordinates are not supported by the regex (\d+ only matches integers)
		{
			name:        "empty_string",
			coordStr:    "",
			size:        types.Size{Width: 1000, Height: 1000},
			expectError: true,
			description: "Error case: empty string",
		},
		{
			name:        "invalid_coordinate_string",
			coordStr:    "abc,def",
			size:        types.Size{Width: 1000, Height: 1000},
			expectError: true,
			description: "Error case: invalid coordinate string",
		},
		{
			name:        "insufficient_coordinates",
			coordStr:    "100",
			size:        types.Size{Width: 1000, Height: 1000},
			expectError: true,
			description: "Error case: insufficient coordinates",
		},
		{
			name:        "invalid_bracket_format",
			coordStr:    "[abc, def]",
			size:        types.Size{Width: 1000, Height: 1000},
			expectError: true,
			description: "Error case: invalid bracket format",
		},
		{
			name:        "invalid_point_tag",
			coordStr:    "<point>abc def</point>",
			size:        types.Size{Width: 1000, Height: 1000},
			expectError: true,
			description: "Error case: invalid point tag",
		},
		{
			name:        "invalid_bbox_tag",
			coordStr:    "<bbox>abc def ghi jkl</bbox>",
			size:        types.Size{Width: 1000, Height: 1000},
			expectError: true,
			description: "Error case: invalid bbox tag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := normalizeStringCoordinates(tt.coordStr, tt.size)

			if tt.expectError {
				assert.Error(t, err, "Test case: %s", tt.description)
				return
			}

			assert.NoError(t, err, "Test case: %s", tt.description)
			assert.Equal(t, len(tt.expected), len(result), "Test case: %s", tt.description)
			for i, expected := range tt.expected {
				assert.Equal(t, expected, result[i], "Test case: %s - coordinate %d", tt.description, i)
			}
		})
	}
}

// Test normalizeActionCoordinates function
func TestNormalizeActionCoordinates(t *testing.T) {
	size := types.Size{Width: 1000, Height: 1000}

	tests := []struct {
		name        string
		coordData   interface{}
		expected    []float64
		expectError bool
	}{
		{
			name:      "JSON array format - []interface{}",
			coordData: []interface{}{100.0, 200.0, 150.0, 250.0},
			expected:  []float64{100.0, 200.0, 150.0, 250.0},
		},
		{
			name:      "JSON array format with int values",
			coordData: []interface{}{100, 200, 150, 250},
			expected:  []float64{100.0, 200.0, 150.0, 250.0},
		},
		{
			name:      "float64 slice format",
			coordData: []float64{100.0, 200.0, 150.0, 250.0},
			expected:  []float64{100.0, 200.0, 150.0, 250.0},
		},
		{
			name:      "string format",
			coordData: "100,200,150,250",
			expected:  []float64{100.0, 200.0, 150.0, 250.0},
		},
		{
			name:      "two-element coordinate",
			coordData: []interface{}{100.0, 200.0},
			expected:  []float64{100.0, 200.0},
		},
		{
			name:        "insufficient elements in array",
			coordData:   []interface{}{100.0},
			expectError: true,
		},
		{
			name:        "invalid array element type",
			coordData:   []interface{}{"abc", 200.0},
			expectError: true,
		},
		{
			name:        "unsupported coordinate format",
			coordData:   map[string]interface{}{"x": 100, "y": 200},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := normalizeActionCoordinates(tt.coordData, size)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, len(tt.expected), len(result))
			for i, expected := range tt.expected {
				assert.Equal(t, expected, result[i])
			}
		})
	}
}

// Test processActionArguments function
func TestProcessActionArguments(t *testing.T) {
	size := types.Size{Width: 1000, Height: 1000}

	tests := []struct {
		name        string
		rawArgs     map[string]interface{}
		expected    map[string]interface{}
		expectError bool
	}{
		{
			name: "coordinate and non-coordinate parameters",
			rawArgs: map[string]interface{}{
				"start_box": "100,200,150,250",
				"content":   "Hello\\nWorld",
			},
			expected: map[string]interface{}{
				"start_box": []float64{100.0, 200.0, 150.0, 250.0},
				"content":   "Hello\nWorld",
			},
		},
		{
			name: "multiple coordinate parameters",
			rawArgs: map[string]interface{}{
				"start_box": "100,200,150,250",
				"end_box":   "300,400,350,450",
			},
			expected: map[string]interface{}{
				"start_box": []float64{100.0, 200.0, 150.0, 250.0},
				"end_box":   []float64{300.0, 400.0, 350.0, 450.0},
			},
		},
		{
			name: "only non-coordinate parameters",
			rawArgs: map[string]interface{}{
				"content":   "Hello World",
				"direction": "down",
			},
			expected: map[string]interface{}{
				"content":   "Hello World",
				"direction": "down",
			},
		},
		{
			name:     "empty arguments",
			rawArgs:  map[string]interface{}{},
			expected: map[string]interface{}{},
		},
		{
			name: "invalid coordinate parameter",
			rawArgs: map[string]interface{}{
				"start_box": "invalid",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := processActionArguments(tt.rawArgs, size)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, len(tt.expected), len(result))

			for key, expectedValue := range tt.expected {
				actualValue, exists := result[key]
				assert.True(t, exists, "Key %s should exist in result", key)

				// Handle slice comparison separately
				if expectedSlice, ok := expectedValue.([]float64); ok {
					actualSlice, ok := actualValue.([]float64)
					assert.True(t, ok, "Value for key %s should be []float64", key)
					assert.Equal(t, len(expectedSlice), len(actualSlice))
					for i, expected := range expectedSlice {
						assert.Equal(t, expected, actualSlice[i])
					}
				} else {
					assert.Equal(t, expectedValue, actualValue)
				}
			}
		})
	}
}
