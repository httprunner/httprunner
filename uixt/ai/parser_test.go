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
			name:     "bracket format",
			input:    "[100, 200, 150, 250]",
			expected: "(100,200,150,250)",
		},
		{
			name:     "bracket format with spaces",
			input:    "[100,  200,  150,  250]",
			expected: "(100,200,150,250)",
		},
		{
			name:     "multiple point tags",
			input:    "<point>100 200</point> and <point>300 400</point>",
			expected: "(100,200) and (300,400)",
		},
		{
			name:     "no coordinates",
			input:    "click on button",
			expected: "click on button",
		},
		{
			name:     "mixed formats",
			input:    "<point>100 200</point> and [300, 400, 350, 450]",
			expected: "(100,200) and (300,400,350,450)",
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
	size := types.Size{Width: 1000, Height: 2000}

	tests := []struct {
		name           string
		relativeCoord  float64
		isXCoord       bool
		expectedResult float64
	}{
		{
			name:           "x coordinate conversion",
			relativeCoord:  500, // 500/1000 * 1000 = 500
			isXCoord:       true,
			expectedResult: 500.0,
		},
		{
			name:           "y coordinate conversion",
			relativeCoord:  500, // 500/1000 * 2000 = 1000
			isXCoord:       false,
			expectedResult: 1000.0,
		},
		{
			name:           "x coordinate with rounding",
			relativeCoord:  333, // 333/1000 * 1000 = 333
			isXCoord:       true,
			expectedResult: 333.0,
		},
		{
			name:           "y coordinate with rounding",
			relativeCoord:  750, // 750/1000 * 2000 = 1500
			isXCoord:       false,
			expectedResult: 1500.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertRelativeToAbsolute(tt.relativeCoord, tt.isXCoord, size)
			assert.Equal(t, tt.expectedResult, result)
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
	size := types.Size{Width: 1000, Height: 1000}

	tests := []struct {
		name        string
		coordStr    string
		expected    []float64
		expectError bool
	}{
		{
			name:     "simple coordinate string",
			coordStr: "100,200,150,250",
			expected: []float64{100.0, 200.0, 150.0, 250.0},
		},
		{
			name:     "coordinate string with spaces",
			coordStr: " 100 , 200 , 150 , 250 ",
			expected: []float64{100.0, 200.0, 150.0, 250.0},
		},
		{
			name:     "point tag format",
			coordStr: "<point>100 200</point>",
			expected: []float64{100.0, 200.0},
		},
		{
			name:     "bbox tag format",
			coordStr: "<bbox>100 200 150 250</bbox>",
			expected: []float64{100.0, 200.0, 150.0, 250.0},
		},
		{
			name:     "bracket format",
			coordStr: "[100, 200, 150, 250]",
			expected: []float64{100.0, 200.0, 150.0, 250.0},
		},
		{
			name:        "empty string",
			coordStr:    "",
			expectError: true,
		},
		{
			name:        "invalid coordinate string",
			coordStr:    "abc,def",
			expectError: true,
		},
		{
			name:        "insufficient coordinates",
			coordStr:    "100",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := normalizeStringCoordinates(tt.coordStr, size)

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
