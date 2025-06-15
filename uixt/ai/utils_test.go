package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractJSONFromContent(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name: "simple JSON",
			content: `{
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
		{
			name: "JSON with Chinese characters in strings",
			content: `{
  "actions": [
    {
      "action_type": "type",
      "action_inputs": {
        "content": "2048经典"
      }
    }
  ],
  "thought": "搜索框已经清空了，现在我要输入\"2048经典\"这个关键词。看到键盘已经弹出来了，正好可以直接开始输入。这样一来，就能找到我们想要玩的那个小游戏了。",
  "error": null
}`,
			expected: `{
  "actions": [
    {
      "action_type": "type",
      "action_inputs": {
        "content": "2048经典"
      }
    }
  ],
  "thought": "搜索框已经清空了，现在我要输入\"2048经典\"这个关键词。看到键盘已经弹出来了，正好可以直接开始输入。这样一来，就能找到我们想要玩的那个小游戏了。",
  "error": null
}`,
		},
		{
			name: "JSON with markdown wrapper",
			content: "```json\n" + `{
  "actions": [
    {
      "action_type": "click",
      "action_inputs": {
        "start_box": [100, 200, 150, 250]
      }
    }
  ],
  "thought": "点击按钮",
  "error": null
}` + "\n```",
			expected: `{
  "actions": [
    {
      "action_type": "click",
      "action_inputs": {
        "start_box": [100, 200, 150, 250]
      }
    }
  ],
  "thought": "点击按钮",
  "error": null
}`,
		},
		{
			name: "JSON embedded in text with Chinese",
			content: `这是一个包含中文的响应：{
  "actions": [
    {
      "action_type": "type",
      "action_inputs": {
        "content": "测试内容"
      }
    }
  ],
  "thought": "这是一个测试思路",
  "error": null
} 后面还有一些文本`,
			expected: `{
  "actions": [
    {
      "action_type": "type",
      "action_inputs": {
        "content": "测试内容"
      }
    }
  ],
  "thought": "这是一个测试思路",
  "error": null
}`,
		},
		{
			name: "JSON with escaped quotes and Chinese",
			content: `{
  "actions": [
    {
      "action_type": "type",
      "action_inputs": {
        "content": "他说：\"你好，世界！\""
      }
    }
  ],
  "thought": "输入包含引号的中文文本",
  "error": null
}`,
			expected: `{
  "actions": [
    {
      "action_type": "type",
      "action_inputs": {
        "content": "他说：\"你好，世界！\""
      }
    }
  ],
  "thought": "输入包含引号的中文文本",
  "error": null
}`,
		},
		{
			name:     "no JSON content",
			content:  "这只是一些普通的文本，没有JSON内容",
			expected: "",
		},
		{
			name: "nested JSON objects with Chinese",
			content: `{
  "actions": [
    {
      "action_type": "click",
      "action_inputs": {
        "start_box": [100, 200, 150, 250],
        "metadata": {
          "description": "点击操作",
          "target": "按钮"
        }
      }
    }
  ],
  "thought": "执行嵌套对象的点击操作",
  "error": null
}`,
			expected: `{
  "actions": [
    {
      "action_type": "click",
      "action_inputs": {
        "start_box": [100, 200, 150, 250],
        "metadata": {
          "description": "点击操作",
          "target": "按钮"
        }
      }
    }
  ],
  "thought": "执行嵌套对象的点击操作",
  "error": null
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractJSONFromContent(tt.content)
			assert.Equal(t, tt.expected, result)
		})
	}
}
