package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractJSONFromContent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple JSON object",
			input:    `{"key": "value"}`,
			expected: `{"key": "value"}`,
		},
		{
			name:     "JSON in markdown code block",
			input:    "```json\n{\"key\": \"value\"}\n```",
			expected: `{"key": "value"}`,
		},
		{
			name:     "JSON in code block without language",
			input:    "```\n{\"key\": \"value\"}\n```",
			expected: `{"key": "value"}`,
		},
		{
			name:     "JSON with surrounding text",
			input:    `Here is the result: {"key": "value"} and some more text`,
			expected: `{"key": "value"}`,
		},
		{
			name:     "multiple JSON objects",
			input:    `{"first": "object"} and {"second": "object"}`,
			expected: `{"first": "object"}`,
		},
		{
			name:     "nested JSON in markdown",
			input:    "```json\n{\"data\": {\"nested\": \"value\"}}\n```",
			expected: `{"data": {"nested": "value"}}`,
		},
		{
			name:     "JSON array",
			input:    `[{"item": 1}, {"item": 2}]`,
			expected: `[{"item": 1}, {"item": 2}]`,
		},
		{
			name:     "JSON array in markdown",
			input:    "```json\n[{\"item\": 1}, {\"item\": 2}]\n```",
			expected: `[{"item": 1}, {"item": 2}]`,
		},
		{
			name:     "text without JSON",
			input:    "This is just plain text without any JSON",
			expected: "",
		},
		{
			name:     "malformed JSON",
			input:    `{"key": "value"`,
			expected: `{"key": "value"`,
		},
		{
			name:     "JSON with unicode",
			input:    `{"message": "测试消息"}`,
			expected: `{"message": "测试消息"}`,
		},
		{
			name:     "multiple code blocks, select first JSON",
			input:    "First block:\n```json\n{\"first\": true}\n```\nSecond block:\n```json\n{\"second\": true}\n```",
			expected: `{"first": true}`,
		},
		{
			name:     "mixed language code blocks",
			input:    "```python\nprint('hello')\n```\n```json\n{\"key\": \"value\"}\n```",
			expected: `{"key": "value"}`,
		},
		{
			name:     "JSON with special characters",
			input:    `{"special": "chars: @#$%^&*()"}`,
			expected: `{"special": "chars: @#$%^&*()"}`,
		},
		{
			name:     "empty JSON object",
			input:    `{}`,
			expected: `{}`,
		},
		{
			name:     "empty JSON array",
			input:    `[]`,
			expected: `[]`,
		},
		{
			name:     "JSON with line breaks",
			input:    "{\n  \"key\": \"value\",\n  \"number\": 123\n}",
			expected: "{\n  \"key\": \"value\",\n  \"number\": 123\n}",
		},
		{
			name:     "markdown with extra whitespace",
			input:    "  ```json  \n  {\"key\": \"value\"}  \n  ```  ",
			expected: `{"key": "value"}`,
		},
		{
			name:     "code block with tildes",
			input:    "~~~json\n{\"key\": \"value\"}\n~~~",
			expected: `{"key": "value"}`,
		},
		{
			name:     "JSON after other text patterns",
			input:    `The response should be formatted as: {"status": "success"}`,
			expected: `{"status": "success"}`,
		},
		{
			name:     "JSON in mixed content",
			input:    `Analysis complete. Result: {"analysis": "positive", "confidence": 0.95} - End of analysis.`,
			expected: `{"analysis": "positive", "confidence": 0.95}`,
		},
		{
			name:     "complex nested JSON",
			input:    `{"outer": {"inner": {"deep": "value", "numbers": [1, 2, 3]}}}`,
			expected: `{"outer": {"inner": {"deep": "value", "numbers": [1, 2, 3]}}}`,
		},
		{
			name:     "JSON with escaped quotes",
			input:    `{"message": "He said \"Hello\" to me"}`,
			expected: `{"message": "He said \"Hello\" to me"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractJSONFromContent(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeUTF8Content(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid UTF-8",
			input:    "Hello 世界",
			expected: "Hello 世界",
		},
		{
			name:     "invalid UTF-8 with replacement characters",
			input:    "Hello \ufffd\ufffd World",
			expected: "Hello  World",
		},
		{
			name:     "mixed valid and invalid",
			input:    "测试\ufffd消息\ufffd",
			expected: "测试消息",
		},
		{
			name:     "only replacement characters",
			input:    "\ufffd\ufffd\ufffd",
			expected: "",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "ASCII only",
			input:    "Hello World 123",
			expected: "Hello World 123",
		},
		{
			name:     "JSON with UTF-8 issues",
			input:    `{"message": "搜索框\ufffd\ufffd显示"}`,
			expected: `{"message": "搜索框显示"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeUTF8Content(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseJSONWithFallback(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedValid   bool
		expectedPass    bool
		expectedThought string
	}{
		{
			name:            "valid JSON",
			input:           `{"pass": true, "thought": "test passed"}`,
			expectedValid:   true,
			expectedPass:    true,
			expectedThought: "test passed",
		},
		{
			name:            "valid JSON with false",
			input:           `{"pass": false, "thought": "test failed"}`,
			expectedValid:   true,
			expectedPass:    false,
			expectedThought: "test failed",
		},
		{
			name:            "malformed JSON with extractable fields",
			input:           `malformed start {"pass": true, "thought": "extracted"} end`,
			expectedValid:   true,
			expectedPass:    true,
			expectedThought: "extracted",
		},
		{
			name:            "content analysis fallback - positive",
			input:           `The test was successful and passed with true result`,
			expectedValid:   true,
			expectedPass:    true,
			expectedThought: "Fallback analysis of malformed response (positive: 3, negative: 0)",
		},
		{
			name:            "content analysis fallback - negative",
			input:           `The test failed with false result and error occurred`,
			expectedValid:   true,
			expectedPass:    false,
			expectedThought: "Fallback analysis of malformed response (positive: 0, negative: 3)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result AssertionResult
			err := parseJSONWithFallback(tt.input, &result)

			if tt.expectedValid {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedPass, result.Pass)
				assert.Equal(t, tt.expectedThought, result.Thought)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestExtractAssertionFieldsManually(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedPass    bool
		expectedThought string
		shouldError     bool
	}{
		{
			name:            "pass true",
			input:           `{"pass": true, "thought": "manual test"}`,
			expectedPass:    true,
			expectedThought: "manual test",
			shouldError:     false,
		},
		{
			name:            "pass false",
			input:           `{"pass": false, "thought": "manual fail"}`,
			expectedPass:    false,
			expectedThought: "manual fail",
			shouldError:     false,
		},
		{
			name:        "no pass field",
			input:       `{"thought": "no pass field"}`,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractAssertionFieldsManually(tt.input)
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedPass, result.Pass)
				assert.Equal(t, tt.expectedThought, result.Thought)
			}
		})
	}
}

func TestExtractQuotedString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple quoted string",
			input:    `"hello world"`,
			expected: "hello world",
		},
		{
			name:     "quoted string with escaped quotes",
			input:    `"He said \"Hello\""`,
			expected: `He said "Hello"`,
		},
		{
			name:     "quoted string with escaped backslash",
			input:    `"path\\to\\file"`,
			expected: `path\to\file`,
		},
		{
			name:     "empty quoted string",
			input:    `""`,
			expected: "",
		},
		{
			name:     "quoted string with unicode",
			input:    `"测试消息"`,
			expected: "测试消息",
		},
		{
			name:     "not a quoted string",
			input:    "hello world",
			expected: "",
		},
		{
			name:     "unclosed quoted string",
			input:    `"unclosed string`,
			expected: "unclosed string",
		},
		{
			name:     "quoted string with extra content after",
			input:    `"content" and more`,
			expected: "content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractQuotedString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCleanJSONContent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "remove trailing comma in object",
			input:    `{"key": "value",}`,
			expected: `{"key": "value"}`,
		},
		{
			name:     "remove trailing comma in array",
			input:    `["item1", "item2",]`,
			expected: `["item1", "item2"]`,
		},
		{
			name:     "clean non-printable characters",
			input:    "{\n\"key\": \"value\"\u0000\u0001}",
			expected: "{\n\"key\": \"value\"}",
		},
		{
			name:     "preserve unicode characters",
			input:    `{"message": "测试消息"}`,
			expected: `{"message": "测试消息"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanJSONContent(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAnalyzeContentForAssertion(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedPass bool
	}{
		{
			name:         "positive indicators",
			input:        "The test was successful and passed",
			expectedPass: true,
		},
		{
			name:         "negative indicators",
			input:        "The test failed with error",
			expectedPass: false,
		},
		{
			name:         "mixed with more positive",
			input:        "Some errors occurred but overall test passed successfully",
			expectedPass: true,
		},
		{
			name:         "no clear indicators",
			input:        "This is just plain text",
			expectedPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzeContentForAssertion(tt.input)
			assert.Equal(t, tt.expectedPass, result.Pass)
			assert.NotEmpty(t, result.Thought)
		})
	}
}

func TestParseStructuredResponse(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		shouldSucceed bool
	}{
		{
			name:          "valid AssertionResult JSON",
			input:         `{"pass": true, "thought": "test passed"}`,
			shouldSucceed: true,
		},
		{
			name:          "malformed JSON with extractable fields",
			input:         `malformed start {"pass": false, "thought": "extracted thought"} end`,
			shouldSucceed: true,
		},
		{
			name:          "UTF-8 issues with JSON",
			input:         "测试结果：\ufffd\ufffd {\"pass\": true, \"thought\": \"处理完成\"}",
			shouldSucceed: true,
		},
		{
			name:          "content analysis fallback",
			input:         "The assertion was successful and passed correctly",
			shouldSucceed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result AssertionResult
			err := parseStructuredResponse(tt.input, &result)
			if tt.shouldSucceed {
				require.NoError(t, err)
				assert.NotEmpty(t, result.Thought)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// Add more test cases for different struct types
func TestParseJSONWithFallback_QueryResult(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedValid   bool
		expectedContent string
		expectedThought string
	}{
		{
			name:            "valid QueryResult JSON",
			input:           `{"content": "extracted info", "thought": "analysis complete"}`,
			expectedValid:   true,
			expectedContent: "extracted info",
			expectedThought: "analysis complete",
		},
		{
			name:            "malformed QueryResult with extractable fields",
			input:           `malformed { "content": "partial info", "thought": "partial analysis" } more text`,
			expectedValid:   true,
			expectedContent: "partial info",
			expectedThought: "partial analysis",
		},
		{
			name:            "completely malformed QueryResult",
			input:           `This is just plain text with no structure`,
			expectedValid:   true,
			expectedContent: "This is just plain text with no structure",
			expectedThought: "Failed to parse as JSON, returning raw content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result QueryResult
			err := parseJSONWithFallback(tt.input, &result)

			if tt.expectedValid {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedContent, result.Content)
				assert.Equal(t, tt.expectedThought, result.Thought)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestParseJSONWithFallback_PlanningResponse(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedValid   bool
		expectedThought string
		expectedError   string
		expectedActions int
	}{
		{
			name:            "valid PlanningJSONResponse",
			input:           `{"actions": [{"action_type": "click"}], "thought": "planning complete", "error": ""}`,
			expectedValid:   true,
			expectedThought: "planning complete",
			expectedError:   "",
			expectedActions: 1,
		},
		{
			name:            "malformed PlanningResponse with extractable thought",
			input:           `malformed { "thought": "partial planning" } more text`,
			expectedValid:   true,
			expectedThought: "partial planning",
			expectedActions: 0,
		},
		{
			name:            "completely malformed PlanningResponse",
			input:           `This is just plain text with no structure`,
			expectedValid:   true,
			expectedThought: "Failed to parse structured response",
			expectedError:   "JSON parsing failed, returning minimal structure",
			expectedActions: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result PlanningJSONResponse
			err := parseJSONWithFallback(tt.input, &result)

			if tt.expectedValid {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedThought, result.Thought)
				assert.Equal(t, tt.expectedError, result.Error)
				assert.Len(t, result.Actions, tt.expectedActions)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestExtractQueryFieldsManually(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedContent string
		expectedThought string
		shouldError     bool
	}{
		{
			name:            "both content and thought",
			input:           `{"content": "test content", "thought": "test thought"}`,
			expectedContent: "test content",
			expectedThought: "test thought",
			shouldError:     false,
		},
		{
			name:            "only content",
			input:           `{"content": "only content"}`,
			expectedContent: "only content",
			expectedThought: "Partial extraction from malformed response",
			shouldError:     false,
		},
		{
			name:            "only thought",
			input:           `{"thought": "only thought"}`,
			expectedContent: "Extracted partial information",
			expectedThought: "only thought",
			shouldError:     false,
		},
		{
			name:        "no extractable fields",
			input:       `{"other": "data"}`,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractQueryFieldsManually(tt.input)
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedContent, result.Content)
				assert.Equal(t, tt.expectedThought, result.Thought)
			}
		})
	}
}

func TestExtractPlanningFieldsManually(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedThought string
		expectedError   string
		shouldError     bool
	}{
		{
			name:            "both thought and error",
			input:           `{"thought": "test planning", "error": "test error"}`,
			expectedThought: "test planning",
			expectedError:   "test error",
			shouldError:     false,
		},
		{
			name:            "only thought",
			input:           `{"thought": "only planning"}`,
			expectedThought: "only planning",
			expectedError:   "",
			shouldError:     false,
		},
		{
			name:            "only error",
			input:           `{"error": "only error"}`,
			expectedThought: "Partial extraction from malformed response",
			expectedError:   "only error",
			shouldError:     false,
		},
		{
			name:        "no extractable fields",
			input:       `{"other": "data"}`,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractPlanningFieldsManually(tt.input)
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedThought, result.Thought)
				assert.Equal(t, tt.expectedError, result.Error)
				assert.NotNil(t, result.Actions) // Should always be initialized
			}
		})
	}
}

// Test the integrated parseStructuredResponse with QueryResult
func TestParseStructuredResponse_QueryResult(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		shouldSucceed bool
	}{
		{
			name:          "valid QueryResult JSON",
			input:         `{"content": "extracted data", "thought": "processing complete"}`,
			shouldSucceed: true,
		},
		{
			name:          "QueryResult with UTF-8 issues",
			input:         "extracted data: 搜索框，里面显示着\ufffd\ufffd {\"content\": \"search box found\", \"thought\": \"visual analysis\"}",
			shouldSucceed: true,
		},
		{
			name:          "malformed QueryResult",
			input:         `malformed start {"content": "partial info"} end`,
			shouldSucceed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result QueryResult
			err := parseStructuredResponse(tt.input, &result)
			if tt.shouldSucceed {
				require.NoError(t, err)
				assert.NotEmpty(t, result.Content, "Content should not be empty")
				assert.NotEmpty(t, result.Thought, "Thought should not be empty")
			} else {
				assert.Error(t, err)
			}
		})
	}
}
