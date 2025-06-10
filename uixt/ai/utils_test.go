package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractJSONFromContent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid JSON",
			input:    `{"content": "test", "thought": "test"}`,
			expected: `{"content": "test", "thought": "test"}`,
		},
		{
			name:  "JSON in markdown",
			input: "```json\n{\n  \"content\": \"test\"\n}\n```",
			expected: `{
  "content": "test"
}`,
		},
		{
			name:     "incomplete JSON without closing brace",
			input:    `{"content": "incomplete json"`,
			expected: "",
		},
		{
			name:     "incomplete JSON with missing closing brace",
			input:    `{"content": "incomplete json", "missing_closing_brace": true`,
			expected: "",
		},
		{
			name:     "plain text",
			input:    "This is just plain text",
			expected: "",
		},
		{
			name: "complex nested JSON with arrays",
			input: `{
  "actions": [
    {
      "action_type": "click",
      "action_inputs": {
        "start_box": [371, 235, 425, 270]
      }
    }
  ],
  "thought": "点击桌面上的抖音应用图标以启动抖音",
  "error": null
}`,
			expected: `{
  "actions": [
    {
      "action_type": "click",
      "action_inputs": {
        "start_box": [371, 235, 425, 270]
      }
    }
  ],
  "thought": "点击桌面上的抖音应用图标以启动抖音",
  "error": null
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractJSONFromContent(tt.input)
			t.Logf("Input: %s", tt.input)
			t.Logf("Output: %s", result)
			assert.Equal(t, tt.expected, result)
		})
	}
}
