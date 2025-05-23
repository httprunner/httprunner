package ai

import (
	"context"
	"testing"

	"github.com/cloudwego/eino/schema"
	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVLMPlanning(t *testing.T) {
	imageBase64, size, err := builtin.LoadImage("testdata/llk_1.png")
	require.NoError(t, err)

	userInstruction := `连连看是一款经典的益智消除类小游戏，通常以图案或图标为主要元素。以下是连连看的基本规则说明：
	1. 游戏目标: 玩家需要在规定时间内，通过连接相同的图案或图标，将它们从游戏界面中消除。
	2. 连接规则:
	- 两个相同的图案可以通过不超过三条直线连接。
	- 连接线可以水平或垂直，但不能穿过其他图案。
	- 连接线的转折次数不能超过两次。
	3. 游戏界面: 游戏界面通常是一个矩形区域，内含多个图案或图标，排列成行和列。
	4. 时间限制: 游戏通常设有时间限制，玩家需要在时间耗尽前完成所有图案的消除。
	5. 得分机制: 每成功连接并消除一对图案，玩家会获得相应的分数。完成游戏后，根据剩余时间和消除效率计算总分。
	6. 关卡设计: 游戏可能包含多个关卡，随着关卡的推进，图案的复杂度和数量会增加。`

	userInstruction += "\n\n请基于以上游戏规则，给出下一步可点击的两个图标坐标"

	modelConfig, err := GetModelConfig(option.LLMServiceTypeUITARS)
	require.NoError(t, err)

	planner, err := NewPlanner(context.Background(), modelConfig)
	require.NoError(t, err)

	opts := &PlanningOptions{
		UserInstruction: userInstruction,
		Message: &schema.Message{
			Role: schema.User,
			MultiContent: []schema.ChatMessagePart{
				{
					Type: schema.ChatMessagePartTypeImageURL,
					ImageURL: &schema.ChatMessageImageURL{
						URL: imageBase64,
					},
				},
			},
		},
		Size: size,
	}

	// 执行规划
	result, err := planner.Call(context.Background(), opts)

	// 验证结果
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotEmpty(t, result.ToolCalls)

	// 验证动作
	toolCall := result.ToolCalls[0]
	assert.NotEmpty(t, toolCall.Function.Name)
	assert.NotEmpty(t, result.Thought)
	assert.NotEmpty(t, result.Text)
}

func TestXHSPlanning(t *testing.T) {
	imageBase64, size, err := builtin.LoadImage("testdata/xhs-feed.jpeg")
	require.NoError(t, err)

	userInstruction := "点击第二个帖子的作者头像"

	modelConfig, err := GetModelConfig(option.LLMServiceTypeUITARS)
	require.NoError(t, err)

	planner, err := NewPlanner(context.Background(), modelConfig)
	require.NoError(t, err)

	opts := &PlanningOptions{
		UserInstruction: userInstruction,
		Message: &schema.Message{
			Role: schema.User,
			MultiContent: []schema.ChatMessagePart{
				{
					Type: schema.ChatMessagePartTypeImageURL,
					ImageURL: &schema.ChatMessageImageURL{
						URL: imageBase64,
					},
				},
			},
		},
		Size: size,
	}

	// 执行规划
	result, err := planner.Call(context.Background(), opts)

	// 验证结果
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotEmpty(t, result.Actions)

	// 验证动作
	action := result.Actions[0]
	assert.NotEmpty(t, action.ActionType)
	assert.NotEmpty(t, result.Thought)
	assert.NotEmpty(t, result.Text)
}

func TestChatList(t *testing.T) {
	imageBase64, size, err := builtin.LoadImage("testdata/chat_list.jpeg")
	require.NoError(t, err)

	userInstruction := "请结合图片的文字信息，请告诉我一共有多少个群聊，哪些群聊右下角有绿点"

	modelConfig, err := GetModelConfig(option.LLMServiceTypeUITARS)
	require.NoError(t, err)

	planner, err := NewPlanner(context.Background(), modelConfig)
	require.NoError(t, err)

	opts := &PlanningOptions{
		UserInstruction: userInstruction,
		Message: &schema.Message{
			Role: schema.User,
			MultiContent: []schema.ChatMessagePart{
				{
					Type: schema.ChatMessagePartTypeImageURL,
					ImageURL: &schema.ChatMessageImageURL{
						URL: imageBase64,
					},
				},
			},
		},
		Size: size,
	}

	// 执行规划
	result, err := planner.Call(context.Background(), opts)

	// 验证结果
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestHandleSwitch(t *testing.T) {
	userInstruction := "发送框下方的联网搜索开关是开启状态" // 点击开启联网搜索开关
	// 检查发送框下方的联网搜索开关，蓝色为开启状态，灰色为关闭状态；若开关处于关闭状态，则点击进行开启

	modelConfig, err := GetModelConfig(option.LLMServiceTypeUITARS)
	require.NoError(t, err)

	planner, err := NewPlanner(context.Background(), modelConfig)
	require.NoError(t, err)

	testCases := []struct {
		imageFile  string
		actionType string
	}{
		{"testdata/deepseek_think_off.png", "finished"},
		{"testdata/deepseek_think_on.png", "finished"},
		{"testdata/deepseek_network_on.png", "finished"},
	}

	for _, tc := range testCases {
		imageBase64, size, err := builtin.LoadImage(tc.imageFile)
		require.NoError(t, err)

		opts := &PlanningOptions{
			UserInstruction: userInstruction,
			Message: &schema.Message{
				Role: schema.User,
				MultiContent: []schema.ChatMessagePart{
					{
						Type: schema.ChatMessagePartTypeImageURL,
						ImageURL: &schema.ChatMessageImageURL{
							URL: imageBase64,
						},
					},
				},
			},
			Size: size,
		}

		// Execute planning
		result, err := planner.Call(context.Background(), opts)

		// Validate results
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, result.Actions[0].ActionType, tc.actionType,
			"Unexpected action type for image file: %s", tc.imageFile)
	}
}

func TestValidateInput(t *testing.T) {
	imageBase64, size, err := builtin.LoadImage("testdata/popup_risk_warning.png")
	require.NoError(t, err)

	tests := []struct {
		name    string
		opts    *PlanningOptions
		wantErr error
	}{
		{
			name: "valid input",
			opts: &PlanningOptions{
				UserInstruction: "点击继续使用按钮",
				Message: &schema.Message{
					Role: schema.User,
					MultiContent: []schema.ChatMessagePart{
						{
							Type: schema.ChatMessagePartTypeImageURL,
							ImageURL: &schema.ChatMessageImageURL{
								URL: imageBase64,
							},
						},
					},
				},
				Size: size,
			},
			wantErr: nil,
		},
		{
			name: "empty instruction",
			opts: &PlanningOptions{
				UserInstruction: "",
				Message: &schema.Message{
					Role:         schema.User,
					MultiContent: []schema.ChatMessagePart{},
				},
				Size: size,
			},
			wantErr: code.LLMPrepareRequestError,
		},
		{
			name: "empty conversation history",
			opts: &PlanningOptions{
				UserInstruction: "点击立即卸载按钮",
				Message:         &schema.Message{},
				Size:            size,
			},
			wantErr: code.LLMPrepareRequestError,
		},
		{
			name: "invalid image data",
			opts: &PlanningOptions{
				UserInstruction: "点击继续使用按钮",
				Message: &schema.Message{
					Role: schema.User,
					MultiContent: []schema.ChatMessagePart{
						{
							Type: schema.ChatMessagePartTypeImageURL,
							Text: "no image",
						},
					},
				},
				Size: size,
			},
			wantErr: code.LLMPrepareRequestError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePlanningInput(tt.opts)
			if tt.wantErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLoadImage(t *testing.T) {
	// Test PNG image
	pngBase64, pngSize, err := builtin.LoadImage("testdata/llk_1.png")
	require.NoError(t, err)
	assert.NotEmpty(t, pngBase64)
	assert.Greater(t, pngSize.Width, 0)
	assert.Greater(t, pngSize.Height, 0)

	// Test JPEG image
	jpegBase64, jpegSize, err := builtin.LoadImage("testdata/xhs-feed.jpeg")
	require.NoError(t, err)
	assert.NotEmpty(t, jpegBase64)
	assert.Greater(t, jpegSize.Width, 0)
	assert.Greater(t, jpegSize.Height, 0)
}
