package ai

import (
	"context"
	"testing"

	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/httprunner/httprunner/v5/uixt/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createAsserter(t *testing.T) *Asserter {
	modelConfig, err := GetModelConfig(option.DOUBAO_1_5_UI_TARS_250328)
	require.NoError(t, err)
	asserter, err := NewAsserter(context.Background(), modelConfig)
	require.NoError(t, err)
	return asserter
}

// 测试有效断言
func TestValidAssertions(t *testing.T) {
	asserter := createAsserter(t)

	testCases := []struct {
		name       string
		assertion  string
		imagePath  string
		expectPass bool
	}{
		{
			name:       "深度思考功能已开启",
			assertion:  "输入框下方的「深度思考」文字是蓝色的",
			imagePath:  "testdata/deepseek_think_on.png",
			expectPass: true,
		},
		{
			name:       "深度思考功能未开启",
			assertion:  "输入框下方的「深度思考」文字不是蓝色的",
			imagePath:  "testdata/deepseek_think_off.png",
			expectPass: true,
		},
		{
			name:       "联网搜索功能已开启",
			assertion:  "输入框下方的「联网搜索」文字是蓝色的",
			imagePath:  "testdata/deepseek_network_on.png",
			expectPass: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			imageBase64, size, err := builtin.LoadImage(tc.imagePath)
			require.NoError(t, err)

			result, err := asserter.Assert(context.Background(), &AssertOptions{
				Assertion:  tc.assertion,
				Screenshot: imageBase64,
				Size:       size,
			})
			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tc.expectPass, result.Pass)
		})
	}
}

// 测试无效参数
func TestInvalidParameters(t *testing.T) {
	asserter := createAsserter(t)
	testCases := []struct {
		name          string
		assertion     string
		screenshot    string
		size          types.Size
		expectedError string
	}{
		{
			name:          "缺少截图",
			assertion:     "测试断言",
			screenshot:    "",
			size:          types.Size{},
			expectedError: "screenshot is required",
		},
		{
			name:          "缺少断言",
			assertion:     "",
			screenshot:    "some-base64-data",
			size:          types.Size{},
			expectedError: "assertion text is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := asserter.Assert(context.Background(), &AssertOptions{
				Assertion:  tc.assertion,
				Screenshot: tc.screenshot,
				Size:       tc.size,
			})
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedError)
		})
	}
}

// Test the main parseAssertionResult function with problematic input
func TestParseAssertionResult(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		shouldSucceed bool
	}{
		{
			name:          "valid JSON response",
			input:         `{"pass": true, "thought": "Assertion passed"}`,
			shouldSucceed: true,
		},
		{
			name:          "response with UTF-8 replacement characters",
			input:         "浅蓝色的搜索框，里面显示着输入的\"ma\"，而\ufffd\ufffd且在搜索框的右上角有一个喇叭 {\"pass\": true, \"thought\": \"found search box\"}",
			shouldSucceed: true,
		},
		{
			name:          "malformed JSON with extraction",
			input:         `malformed start {"pass": true, "thought": "extracted successfully"} malformed end`,
			shouldSucceed: true,
		},
		{
			name:          "completely malformed but analyzable",
			input:         "This assertion test passed and was successful",
			shouldSucceed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseAssertionResult(tt.input, option.DOUBAO_1_5_UI_TARS_250328)
			if tt.shouldSucceed {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotEmpty(t, result.Thought)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
