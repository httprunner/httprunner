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
	modelConfig, err := GetModelConfig(option.DOUBAO_1_5_UI_TARS_250428)
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
