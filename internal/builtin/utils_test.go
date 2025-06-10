package builtin

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInterface2Float64(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    float64
		wantErr bool
	}{
		{
			name:    "convert int",
			input:   42,
			want:    42.0,
			wantErr: false,
		},
		{
			name:    "convert int32",
			input:   int32(42),
			want:    42.0,
			wantErr: false,
		},
		{
			name:    "convert int64",
			input:   int64(42),
			want:    42.0,
			wantErr: false,
		},
		{
			name:    "convert float32",
			input:   float32(42.5),
			want:    42.5,
			wantErr: false,
		},
		{
			name:    "convert float64",
			input:   42.5,
			want:    42.5,
			wantErr: false,
		},
		{
			name:    "convert string valid number",
			input:   "42.5",
			want:    42.5,
			wantErr: false,
		},
		{
			name:    "convert string valid number",
			input:   "425",
			want:    425.0,
			wantErr: false,
		},
		{
			name:    "convert string invalid number",
			input:   "invalid",
			want:    0,
			wantErr: true,
		},
		{
			name:    "convert json.Number valid",
			input:   json.Number("42.5"),
			want:    42.5,
			wantErr: false,
		},
		{
			name:    "convert json.Number invalid",
			input:   json.Number("invalid"),
			want:    0,
			wantErr: true,
		},
		{
			name:    "convert unsupported type",
			input:   []int{1, 2, 3},
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Interface2Float64(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

// TestUTF8Encoding tests that Chinese characters are properly encoded in JSON files
func TestUTF8Encoding(t *testing.T) {
	// Create test data with Chinese characters
	testData := map[string]interface{}{
		"name":        "连连看小游戏自动化测试",
		"description": "这是一个包含中文字符的测试用例",
		"steps": []map[string]interface{}{
			{
				"name":   "启动抖音「连了又连」小游戏",
				"action": "启动应用程序",
				"result": "成功启动游戏",
			},
			{
				"name":   "开始游戏",
				"action": "点击开始按钮",
				"result": "游戏开始运行",
			},
		},
		"platform": map[string]string{
			"os":      "安卓系统",
			"version": "版本 12",
			"device":  "测试设备",
		},
	}

	// Create temporary file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_utf8.json")

	// Test the fixed Dump2JSON function
	err := Dump2JSON(testData, testFile)
	if err != nil {
		t.Fatalf("Failed to dump JSON: %v", err)
	}

	// Read the file back and verify content
	fileContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read JSON file: %v", err)
	}

	// Parse the JSON to ensure it's valid
	var parsedData map[string]interface{}
	err = json.Unmarshal(fileContent, &parsedData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Verify Chinese characters are preserved
	if parsedData["name"] != "连连看小游戏自动化测试" {
		t.Errorf("Chinese characters not preserved in name field")
	}

	if parsedData["description"] != "这是一个包含中文字符的测试用例" {
		t.Errorf("Chinese characters not preserved in description field")
	}

	// Verify nested Chinese characters
	steps, ok := parsedData["steps"].([]interface{})
	if !ok {
		t.Fatalf("Steps field is not an array")
	}

	firstStep, ok := steps[0].(map[string]interface{})
	if !ok {
		t.Fatalf("First step is not a map")
	}

	if firstStep["name"] != "启动抖音「连了又连」小游戏" {
		t.Errorf("Chinese characters not preserved in step name")
	}

	t.Logf("UTF-8 encoding test passed. File content length: %d bytes", len(fileContent))
}
