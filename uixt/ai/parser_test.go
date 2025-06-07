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
	assert.Equal(t, function.Name, "uixt__click")

	var arguments map[string]interface{}
	err = json.Unmarshal([]byte(function.Arguments), &arguments)
	assert.Nil(t, err)
	assert.Contains(t, arguments, "x")
	assert.Contains(t, arguments, "y")

	text = "Thought: 我看到页面上有几个帖子，第二个帖子的标题是\"字节四年，头发白了\"。要完成任务，我需要点击这个帖子下方的作者头像，这样就能进入作者的个人主页了。\nAction: click(start_point='<point>550 450 550 450</point>')"
	result, err = parser.Parse(text, types.Size{Height: 2341, Width: 1024})
	assert.Nil(t, err)
	function = result.ToolCalls[0].Function
	assert.Equal(t, function.Name, "uixt__click")

	err = json.Unmarshal([]byte(function.Arguments), &arguments)
	assert.Nil(t, err)
	assert.Contains(t, arguments, "x")
	assert.Contains(t, arguments, "y")

	// Test new bracket format - should convert bounding box to center point
	text = "Thought: 我需要点击这个按钮\nAction: click(start_box='[100, 200, 150, 250]')"
	result, err = parser.Parse(text, types.Size{Height: 1000, Width: 1000})
	assert.Nil(t, err)
	function = result.ToolCalls[0].Function
	assert.Equal(t, function.Name, "uixt__click")

	err = json.Unmarshal([]byte(function.Arguments), &arguments)
	assert.Nil(t, err)
	// Should be converted to center point x=125, y=225 from bounding box [100, 200, 150, 250]
	assert.Equal(t, 125.0, arguments["x"]) // (100 + 150) / 2 = 125
	assert.Equal(t, 225.0, arguments["y"]) // (200 + 250) / 2 = 225

	// Test drag operation with both start_box and end_box - should use from_x,from_y,to_x,to_y format
	text = "Thought: 我需要拖拽元素\nAction: drag(start_box='[100, 200, 150, 250]', end_box='[300, 400, 350, 450]')"
	result, err = parser.Parse(text, types.Size{Height: 1000, Width: 1000})
	assert.Nil(t, err)
	function = result.ToolCalls[0].Function
	assert.Equal(t, function.Name, "uixt__drag")
	// ActionInputs is now in from_x,from_y,to_x,to_y format for drag operations
	err = json.Unmarshal([]byte(function.Arguments), &arguments)
	assert.Nil(t, err)
	// Should be converted to from_x,from_y,to_x,to_y format
	assert.Equal(t, 125.0, arguments["from_x"]) // start center x: (100 + 150) / 2 = 125
	assert.Equal(t, 225.0, arguments["from_y"]) // start center y: (200 + 250) / 2 = 225
	assert.Equal(t, 325.0, arguments["to_x"])   // end center x: (300 + 350) / 2 = 325
	assert.Equal(t, 425.0, arguments["to_y"])   // end center y: (400 + 450) / 2 = 425
}

// Test normalizeCoordinatesFormat function
func TestNormalizeCoordinatesFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Basic format conversions
		{
			name:     "point_tag_2_numbers",
			input:    "<point>100 200</point>",
			expected: "(100,200)",
		},
		{
			name:     "point_tag_4_numbers",
			input:    "<point>100 200 150 250</point>",
			expected: "(100,200,150,250)",
		},
		{
			name:     "bbox_tag",
			input:    "<bbox>100 200 150 250</bbox>",
			expected: "(100,200,150,250)",
		},
		{
			name:     "bracket_format_4_coords",
			input:    "[100, 200, 150, 250]",
			expected: "(100,200,150,250)",
		},
		// Edge cases
		{
			name:     "zero_coordinates",
			input:    "<point>0 0</point>",
			expected: "(0,0)",
		},
		{
			name:     "large_coordinates",
			input:    "<point>1920 1080</point>",
			expected: "(1920,1080)",
		},
		// Multiple formats in one string
		{
			name:     "mixed_formats",
			input:    "<point>100 200</point> and [300, 400, 350, 450]",
			expected: "(100,200) and (300,400,350,450)",
		},
		// Unsupported formats (should remain unchanged)
		{
			name:     "bracket_2_coords_not_converted",
			input:    "[100, 200]",
			expected: "[100, 200]",
		},
		{
			name:     "decimals_not_converted",
			input:    "<point>100.5 200.7</point>",
			expected: "<point>100.5 200.7</point>",
		},
		{
			name:     "no_coordinates",
			input:    "click on button",
			expected: "click on button",
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
		// Basic conversion tests
		{
			name:           "standard_x_coordinate",
			size:           types.Size{Width: 1920, Height: 1080},
			relativeCoord:  500, // 500/1000 * 1920 = 960
			isXCoord:       true,
			expectedResult: 960.0,
			description:    "Standard X coordinate conversion",
		},
		{
			name:           "standard_y_coordinate",
			size:           types.Size{Width: 1920, Height: 1080},
			relativeCoord:  500, // 500/1000 * 1080 = 540
			isXCoord:       false,
			expectedResult: 540.0,
			description:    "Standard Y coordinate conversion",
		},
		// Mobile device tests
		{
			name:           "mobile_x_coordinate",
			size:           types.Size{Width: 375, Height: 812},
			relativeCoord:  200, // 200/1000 * 375 = 75
			isXCoord:       true,
			expectedResult: 75.0,
			description:    "Mobile device X coordinate",
		},
		{
			name:           "mobile_y_coordinate",
			size:           types.Size{Width: 375, Height: 812},
			relativeCoord:  600, // 600/1000 * 812 = 487.2
			isXCoord:       false,
			expectedResult: 487.2,
			description:    "Mobile device Y coordinate",
		},
		// Edge cases
		{
			name:           "zero_coordinate",
			size:           types.Size{Width: 1920, Height: 1080},
			relativeCoord:  0,
			isXCoord:       true,
			expectedResult: 0.0,
			description:    "Zero coordinate",
		},
		{
			name:           "maximum_coordinate",
			size:           types.Size{Width: 1920, Height: 1080},
			relativeCoord:  1000, // 1000/1000 * 1920 = 1920
			isXCoord:       true,
			expectedResult: 1920.0,
			description:    "Maximum coordinate (1000 -> full width)",
		},
		// Coordinates > 1000 (normalization scenarios)
		{
			name:           "coordinate_greater_than_1000",
			size:           types.Size{Width: 1920, Height: 1080},
			relativeCoord:  1500, // 1500/1000 * 1920 = 2880
			isXCoord:       true,
			expectedResult: 2880.0,
			description:    "Coordinate > 1000: normalization test",
		},
		{
			name:           "very_large_coordinate",
			size:           types.Size{Width: 1920, Height: 1080},
			relativeCoord:  2000, // 2000/1000 * 1080 = 2160
			isXCoord:       false,
			expectedResult: 2160.0,
			description:    "Very large coordinate test",
		},
		// High resolution test
		{
			name:           "4k_resolution_large_coordinate",
			size:           types.Size{Width: 3840, Height: 2160},
			relativeCoord:  1500, // 1500/1000 * 3840 = 5760
			isXCoord:       true,
			expectedResult: 5760.0,
			description:    "4K resolution with large coordinate",
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
		// Basic coordinate formats
		{
			name:        "simple_coordinate_string",
			coordStr:    "100,200,150,250",
			size:        types.Size{Width: 1000, Height: 1000},
			expected:    []float64{100.0, 200.0, 150.0, 250.0},
			description: "Simple comma-separated coordinates",
		},
		{
			name:        "point_tag_format",
			coordStr:    "<point>235 512</point>",
			size:        types.Size{Width: 1920, Height: 1080},
			expected:    []float64{451.2, 553.0}, // 235/1000*1920=451.2, 512/1000*1080=553.0
			description: "Point tag format with screen scaling",
		},
		{
			name:        "bbox_tag_format",
			coordStr:    "<bbox>100 200 150 250</bbox>",
			size:        types.Size{Width: 1920, Height: 1080},
			expected:    []float64{192.0, 216.0, 288.0, 270.0}, // All scaled to 1920x1080
			description: "Bbox tag format with screen scaling",
		},
		{
			name:        "bracket_format",
			coordStr:    "[100, 200, 150, 250]",
			size:        types.Size{Width: 1000, Height: 1000},
			expected:    []float64{100.0, 200.0, 150.0, 250.0},
			description: "Bracket format coordinates",
		},
		// Mobile device test
		{
			name:        "mobile_device_coordinates",
			coordStr:    "<point>200 600</point>",
			size:        types.Size{Width: 375, Height: 812},
			expected:    []float64{75.0, 487.2}, // 200/1000*375=75, 600/1000*812=487.2
			description: "Mobile device coordinate conversion",
		},
		// Edge cases
		{
			name:        "zero_coordinates",
			coordStr:    "0,0,0,0",
			size:        types.Size{Width: 1920, Height: 1080},
			expected:    []float64{0.0, 0.0, 0.0, 0.0},
			description: "Zero coordinates",
		},
		{
			name:        "maximum_coordinates",
			coordStr:    "1000,1000,1000,1000",
			size:        types.Size{Width: 1920, Height: 1080},
			expected:    []float64{1920.0, 1080.0, 1920.0, 1080.0}, // Maximum -> screen edges
			description: "Maximum coordinates (1000 -> screen edges)",
		},
		// Coordinates > 1000 (normalization scenarios)
		{
			name:        "coordinates_greater_than_1000",
			coordStr:    "1200,1500,1400,1800",
			size:        types.Size{Width: 1920, Height: 1080},
			expected:    []float64{2304.0, 1620.0, 2688.0, 1944.0}, // Scaled up for larger screen
			description: "Coordinates > 1000: scaling to larger screen",
		},
		{
			name:        "very_large_coordinates",
			coordStr:    "[2000, 3000, 2500, 3500]",
			size:        types.Size{Width: 1920, Height: 1080},
			expected:    []float64{3840.0, 3240.0, 4800.0, 3780.0}, // Very large coordinates
			description: "Very large coordinates > 2000",
		},
		{
			name:        "mixed_coordinates_boundary",
			coordStr:    "800,1200,1000,1500",
			size:        types.Size{Width: 1920, Height: 1080},
			expected:    []float64{1536.0, 1296.0, 1920.0, 1620.0}, // Mixed coordinates
			description: "Mixed coordinates around 1000 boundary",
		},
		// Error cases
		{
			name:        "empty_string",
			coordStr:    "",
			size:        types.Size{Width: 1000, Height: 1000},
			expectError: true,
			description: "Empty string should cause error",
		},
		{
			name:        "invalid_coordinate_string",
			coordStr:    "abc,def",
			size:        types.Size{Width: 1000, Height: 1000},
			expectError: true,
			description: "Invalid coordinate string should cause error",
		},
		{
			name:        "insufficient_coordinates",
			coordStr:    "100",
			size:        types.Size{Width: 1000, Height: 1000},
			expectError: true,
			description: "Insufficient coordinates should cause error",
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
	size := types.Size{Width: 1920, Height: 800} // Width>1000, Height<1000 for testing coordinate normalization

	tests := []struct {
		name        string
		coordData   interface{}
		expected    []float64
		expectError bool
	}{
		{
			name:      "JSON array format - []interface{}",
			coordData: []interface{}{100.0, 200.0, 150.0, 250.0},
			expected:  []float64{192.0, 160.0, 288.0, 200.0}, // Scaled: 100/1000*1920=192, 200/1000*800=160, etc.
		},
		{
			name:      "JSON array format with int values",
			coordData: []interface{}{100, 200, 150, 250},
			expected:  []float64{192.0, 160.0, 288.0, 200.0}, // Scaled: 100/1000*1920=192, 200/1000*800=160, etc.
		},
		{
			name:      "float64 slice format",
			coordData: []float64{100.0, 200.0, 150.0, 250.0},
			expected:  []float64{192.0, 160.0, 288.0, 200.0}, // Scaled: 100/1000*1920=192, 200/1000*800=160, etc.
		},
		{
			name:      "string format",
			coordData: "100,200,150,250",
			expected:  []float64{192.0, 160.0, 288.0, 200.0}, // Scaled: 100/1000*1920=192, 200/1000*800=160, etc.
		},
		{
			name:      "two-element coordinate",
			coordData: []interface{}{100.0, 200.0},
			expected:  []float64{192.0, 160.0}, // Scaled: 100/1000*1920=192, 200/1000*800=160
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
	size := types.Size{Width: 1920, Height: 800} // Width>1000, Height<1000 for testing coordinate normalization

	tests := []struct {
		name        string
		rawArgs     map[string]interface{}
		expected    map[string]interface{}
		expectError bool
	}{
		{
			name: "basic_coordinate_and_text_parameters",
			rawArgs: map[string]interface{}{
				"start_box": "100,200,150,250",
				"content":   "Hello\\nWorld",
			},
			expected: map[string]interface{}{
				"start_box": []float64{240.0, 180.0}, // Center point: [100,200,150,250] -> scaled coords [192,160,288,200] -> center (192+288)/2=240, (160+200)/2=180
				"content":   "Hello\nWorld",
			},
		},
		{
			name: "drag_operation_dual_coordinates",
			rawArgs: map[string]interface{}{
				"start_box": "100,200,150,250",
				"end_box":   "300,400,350,450",
			},
			expected: map[string]interface{}{
				"start_box": []float64{240.0, 180.0}, // Center point: [100,200,150,250] -> scaled coords [192,160,288,200] -> center (192+288)/2=240, (160+200)/2=180
				"end_box":   []float64{624.0, 340.0}, // Center point: [300,400,350,450] -> scaled coords [576,320,672,360] -> center (576+672)/2=624, (320+360)/2=340
			},
		},
		{
			name: "coordinates_greater_than_1000",
			rawArgs: map[string]interface{}{
				"start_box": "1200,1500,1400,1800",
			},
			expected: map[string]interface{}{
				"start_box": []float64{2496.0, 1320.0}, // Center point: [1200,1500,1400,1800] -> scaled coords [2304,1200,2688,1440] -> center (2304+2688)/2=2496, (1200+1440)/2=1320
			},
		},
		{
			name: "mixed_large_and_small_coordinates",
			rawArgs: map[string]interface{}{
				"start_box": "800,1200,1000,1500",
				"end_box":   "1500,500,2000,800",
			},
			expected: map[string]interface{}{
				"start_box": []float64{1728.0, 1080.0}, // Center point: [800,1200,1000,1500] -> scaled coords [1536,960,1920,1200] -> center (1536+1920)/2=1728, (960+1200)/2=1080
				"end_box":   []float64{3360.0, 520.0},  // Center point: [1500,500,2000,800] -> scaled coords [2880,400,3840,640] -> center (2880+3840)/2=3360, (400+640)/2=520
			},
		},
		{
			name: "non_coordinate_parameters_only",
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
			name:     "empty_arguments",
			rawArgs:  map[string]interface{}{},
			expected: map[string]interface{}{},
		},
		{
			name: "invalid_coordinate_parameter",
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

// Test new coordinate conversion logic
func TestNewCoordinateConversion(t *testing.T) {
	parser := &UITARSContentParser{}

	// Test single start_box conversion to center point
	text := "Thought: 我需要点击这个按钮\nAction: click(start_box='100,200,150,250')"
	result, err := parser.Parse(text, types.Size{Height: 1000, Width: 1000})
	assert.Nil(t, err)
	function := result.ToolCalls[0].Function
	assert.Equal(t, function.Name, "uixt__click")

	var arguments map[string]interface{}
	err = json.Unmarshal([]byte(function.Arguments), &arguments)
	assert.Nil(t, err)

	// Should convert bounding box [100,200,150,250] to center point x=125.0, y=225.0
	assert.Equal(t, 125.0, arguments["x"]) // (100 + 150) / 2 = 125
	assert.Equal(t, 225.0, arguments["y"]) // (200 + 250) / 2 = 225

	// Test drag operation conversion to from_x,from_y,to_x,to_y format
	text = "Thought: 我需要拖拽元素\nAction: drag(start_box='100,200,150,250', end_box='300,400,350,450')"
	result, err = parser.Parse(text, types.Size{Height: 1000, Width: 1000})
	assert.Nil(t, err)
	function = result.ToolCalls[0].Function
	assert.Equal(t, function.Name, "uixt__drag")

	// ActionInputs is now in from_x,from_y,to_x,to_y format for drag operations
	err = json.Unmarshal([]byte(function.Arguments), &arguments)
	assert.Nil(t, err)

	// Should convert to from_x,from_y,to_x,to_y format
	assert.Equal(t, 125.0, arguments["from_x"]) // start center x: (100 + 150) / 2 = 125
	assert.Equal(t, 225.0, arguments["from_y"]) // start center y: (200 + 250) / 2 = 225
	assert.Equal(t, 325.0, arguments["to_x"])   // end center x: (300 + 350) / 2 = 325
	assert.Equal(t, 425.0, arguments["to_y"])   // end center y: (400 + 450) / 2 = 425

	// Test non-coordinate operation (type action)
	text = "Thought: 我需要输入文本\nAction: type(content='Hello World')"
	result, err = parser.Parse(text, types.Size{Height: 1000, Width: 1000})
	assert.Nil(t, err)
	function = result.ToolCalls[0].Function
	assert.Equal(t, function.Name, "uixt__type")

	// ActionInputs should be a map for non-coordinate operations with parameter mapping
	err = json.Unmarshal([]byte(function.Arguments), &arguments)
	assert.Nil(t, err)
	assert.Equal(t, "Hello World", arguments["text"]) // content should be mapped to text
}

// Test convertProcessedArgs function
func TestConvertProcessedArgs(t *testing.T) {
	tests := []struct {
		name          string
		processedArgs map[string]interface{}
		actionType    string
		expected      map[string]interface{}
		expectError   bool
		description   string
	}{
		// Single coordinate operation tests
		{
			name: "single_coordinate_operation",
			processedArgs: map[string]interface{}{
				"start_box": []float64{125.0, 225.0},
			},
			actionType: "click",
			expected: map[string]interface{}{
				"x": 125.0,
				"y": 225.0,
			},
			description: "Single coordinate operation should convert to x,y format",
		},
		{
			name: "single_coordinate_with_rounding",
			processedArgs: map[string]interface{}{
				"start_box": []float64{125.123456, 225.987654},
			},
			actionType: "click",
			expected: map[string]interface{}{
				"x": 125.1,
				"y": 226.0,
			},
			description: "Coordinates should be rounded to one decimal place",
		},
		// Drag operation tests
		{
			name: "drag_operation_dual_coordinates",
			processedArgs: map[string]interface{}{
				"start_box": []float64{125.0, 225.0},
				"end_box":   []float64{325.0, 425.0},
			},
			actionType: "drag",
			expected: map[string]interface{}{
				"from_x": 125.0,
				"from_y": 225.0,
				"to_x":   325.0,
				"to_y":   425.0,
			},
			description: "Drag operation should convert to from_x,from_y,to_x,to_y format",
		},
		{
			name: "drag_operation_with_rounding",
			processedArgs: map[string]interface{}{
				"start_box": []float64{125.123456, 225.987654},
				"end_box":   []float64{325.555555, 425.444444},
			},
			actionType: "drag",
			expected: map[string]interface{}{
				"from_x": 125.1,
				"from_y": 226.0,
				"to_x":   325.6,
				"to_y":   425.4,
			},
			description: "Drag coordinates should be rounded to one decimal place",
		},
		// Non-coordinate operation tests
		{
			name: "non_coordinate_operation_with_parameter_mapping",
			processedArgs: map[string]interface{}{
				"content":   "Hello World",
				"direction": "down",
			},
			actionType: "type",
			expected: map[string]interface{}{
				"text":      "Hello World", // content should be mapped to text
				"direction": "down",
			},
			description: "Non-coordinate operation should apply parameter name mapping",
		},
		{
			name: "non_coordinate_operation_key_mapping",
			processedArgs: map[string]interface{}{
				"key": "enter",
			},
			actionType: "hotkey",
			expected: map[string]interface{}{
				"keycode": "enter", // key should be mapped to keycode
			},
			description: "Key parameter should be mapped to keycode",
		},
		{
			name: "non_coordinate_operation_mixed_parameters",
			processedArgs: map[string]interface{}{
				"content":   "Test input",
				"key":       "ctrl+c",
				"direction": "up",
				"timeout":   5,
			},
			actionType: "mixed",
			expected: map[string]interface{}{
				"text":      "Test input", // content -> text
				"keycode":   "ctrl+c",     // key -> keycode
				"direction": "up",         // unchanged
				"timeout":   5,            // unchanged
			},
			description: "Mixed parameters should apply correct mappings",
		},
		{
			name:          "empty_arguments",
			processedArgs: map[string]interface{}{},
			actionType:    "empty",
			expected:      map[string]interface{}{},
			description:   "Empty arguments should return empty map",
		},
		// Error cases
		{
			name: "invalid_single_coordinate_format",
			processedArgs: map[string]interface{}{
				"start_box": "invalid",
			},
			actionType:  "click",
			expectError: true,
			description: "Invalid coordinate format should cause error",
		},
		{
			name: "invalid_drag_start_coordinate",
			processedArgs: map[string]interface{}{
				"start_box": "invalid",
				"end_box":   []float64{325.0, 425.0},
			},
			actionType:  "drag",
			expectError: true,
			description: "Invalid start coordinate in drag should cause error",
		},
		{
			name: "invalid_drag_end_coordinate",
			processedArgs: map[string]interface{}{
				"start_box": []float64{125.0, 225.0},
				"end_box":   "invalid",
			},
			actionType:  "drag",
			expectError: true,
			description: "Invalid end coordinate in drag should cause error",
		},
		{
			name: "drag_insufficient_start_coordinates",
			processedArgs: map[string]interface{}{
				"start_box": []float64{125.0}, // Only one coordinate
				"end_box":   []float64{325.0, 425.0},
			},
			actionType:  "drag",
			expectError: true,
			description: "Insufficient start coordinates in drag should cause error",
		},
		{
			name: "drag_insufficient_end_coordinates",
			processedArgs: map[string]interface{}{
				"start_box": []float64{125.0, 225.0},
				"end_box":   []float64{325.0}, // Only one coordinate
			},
			actionType:  "drag",
			expectError: true,
			description: "Insufficient end coordinates in drag should cause error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertProcessedArgs(tt.processedArgs, tt.actionType)

			if tt.expectError {
				assert.Error(t, err, "Test case: %s", tt.description)
				return
			}

			assert.NoError(t, err, "Test case: %s", tt.description)
			assert.Equal(t, len(tt.expected), len(result), "Test case: %s", tt.description)

			for key, expectedValue := range tt.expected {
				actualValue, exists := result[key]
				assert.True(t, exists, "Key %s should exist in result for test: %s", key, tt.description)
				assert.Equal(t, expectedValue, actualValue, "Value for key %s should match for test: %s", key, tt.description)
			}
		})
	}
}

// Test mapParameterName function
func TestMapParameterName(t *testing.T) {
	tests := []struct {
		name        string
		paramName   string
		expected    string
		description string
	}{
		{
			name:        "content_to_text",
			paramName:   "content",
			expected:    "text",
			description: "content parameter should be mapped to text",
		},
		{
			name:        "key_to_keycode",
			paramName:   "key",
			expected:    "keycode",
			description: "key parameter should be mapped to keycode",
		},
		{
			name:        "unchanged_parameter_direction",
			paramName:   "direction",
			expected:    "direction",
			description: "direction parameter should remain unchanged",
		},
		{
			name:        "unchanged_parameter_start_box",
			paramName:   "start_box",
			expected:    "start_box",
			description: "start_box parameter should remain unchanged",
		},
		{
			name:        "unchanged_parameter_end_box",
			paramName:   "end_box",
			expected:    "end_box",
			description: "end_box parameter should remain unchanged",
		},
		{
			name:        "unchanged_parameter_timeout",
			paramName:   "timeout",
			expected:    "timeout",
			description: "timeout parameter should remain unchanged",
		},
		{
			name:        "unchanged_parameter_custom",
			paramName:   "custom_param",
			expected:    "custom_param",
			description: "custom parameter should remain unchanged",
		},
		{
			name:        "empty_parameter_name",
			paramName:   "",
			expected:    "",
			description: "empty parameter name should remain empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapParameterName(tt.paramName)
			assert.Equal(t, tt.expected, result, "Test case: %s", tt.description)
		})
	}
}
