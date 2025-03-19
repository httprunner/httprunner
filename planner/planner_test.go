package planner

import (
	"context"
	"os"
	"testing"

	"github.com/cloudwego/eino/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVLMPlanning(t *testing.T) {
	err := loadEnv("testdata/.env")
	require.NoError(t, err)

	imageBase64, err := loadImage("testdata/llk_1.png")
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

	userInstruction += "\n\n请基于以上游戏规则，请先点击第一个图标"

	planner, err := NewPlanner(context.Background())
	require.NoError(t, err)

	opts := &PlanningOptions{
		UserInstruction: userInstruction,
		ConversationHistory: []*schema.Message{
			{
				Role: schema.User,
				MultiContent: []schema.ChatMessagePart{
					{
						Type: "image_url",
						ImageURL: &schema.ChatMessageImageURL{
							URL: imageBase64,
						},
					},
				},
			},
		},
	}

	// 执行规划
	result, err := planner.Start(opts)

	// 验证结果
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotEmpty(t, result.Actions)

	// 验证动作
	action := result.Actions[0]
	assert.NotEmpty(t, action.ActionType)
	assert.NotEmpty(t, action.Thought)

	// 根据动作类型验证参数
	switch action.ActionType {
	case "click", "drag", "left_double", "right_single", "scroll":
		// 这些动作需要验证坐标
		assert.NotEmpty(t, action.ActionInputs["startBox"])

		// 验证坐标格式
		coords, ok := action.ActionInputs["startBox"].([]float64)
		require.True(t, ok)
		require.True(t, len(coords) >= 2) // 至少有 x, y 坐标

		// 验证坐标范围
		for _, coord := range coords {
			assert.GreaterOrEqual(t, coord, float64(0))
		}

	case "type":
		// 验证文本内容
		assert.NotEmpty(t, action.ActionInputs["content"])

	case "hotkey":
		// 验证按键
		assert.NotEmpty(t, action.ActionInputs["key"])

	case "wait", "finished", "call_user":
		// 这些动作不需要额外参数

	default:
		t.Fatalf("未知的动作类型: %s", action.ActionType)
	}
}

func TestXHSPlanning(t *testing.T) {
	err := loadEnv("testdata/.env")
	require.NoError(t, err)

	imageBase64, err := loadImage("testdata/xhs-feed.jpeg")
	require.NoError(t, err)

	userInstruction := `点击第二个帖子的作者头像`

	planner, err := NewPlanner(context.Background())
	require.NoError(t, err)

	opts := &PlanningOptions{
		UserInstruction: userInstruction,
		ConversationHistory: []*schema.Message{
			{
				Role: schema.User,
				MultiContent: []schema.ChatMessagePart{
					{
						Type: "image_url",
						ImageURL: &schema.ChatMessageImageURL{
							URL: imageBase64,
						},
					},
				},
			},
		},
	}

	// 执行规划
	result, err := planner.Start(opts)

	// 验证结果
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotEmpty(t, result.Actions)

	// 验证动作
	action := result.Actions[0]
	assert.NotEmpty(t, action.ActionType)
	assert.NotEmpty(t, action.Thought)
}

func TestValidateInput(t *testing.T) {
	imageBase64, err := loadImage("testdata/popup_risk_warning.png")
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
				ConversationHistory: []*schema.Message{
					{
						Role: schema.User,
						MultiContent: []schema.ChatMessagePart{
							{
								Type: "image_url",
								ImageURL: &schema.ChatMessageImageURL{
									URL: imageBase64,
								},
							},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "empty instruction",
			opts: &PlanningOptions{
				UserInstruction: "",
				ConversationHistory: []*schema.Message{
					{
						Role:    schema.User,
						Content: "",
					},
				},
			},
			wantErr: ErrEmptyInstruction,
		},
		{
			name: "empty conversation history",
			opts: &PlanningOptions{
				UserInstruction:     "点击立即卸载按钮",
				ConversationHistory: []*schema.Message{},
			},
			wantErr: ErrNoConversationHistory,
		},
		{
			name: "invalid image data",
			opts: &PlanningOptions{
				UserInstruction: "点击继续使用按钮",
				ConversationHistory: []*schema.Message{
					{
						Role:    schema.User,
						Content: "no image",
					},
				},
			},
			wantErr: ErrInvalidImageData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateInput(tt.opts)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProcessVLMResponse(t *testing.T) {
	tests := []struct {
		name    string
		actions []ParsedAction
		wantErr bool
	}{
		{
			name: "valid response",
			actions: []ParsedAction{
				{
					ActionType: "click",
					ActionInputs: map[string]interface{}{
						"startBox": []float64{0.5, 0.5},
					},
					Thought: "点击中心位置",
				},
			},
			wantErr: false,
		},
		{
			name:    "empty actions",
			actions: []ParsedAction{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := processVLMResponse(tt.actions)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.actions, result.Actions)
		})
	}
}

func TestSavePositionImg(t *testing.T) {
	imageBase64, err := loadImage("testdata/popup_risk_warning.png")
	require.NoError(t, err)

	params := struct {
		InputImgBase64 string
		Rect           struct {
			X float64
			Y float64
		}
		OutputPath string
	}{
		InputImgBase64: imageBase64,
		Rect: struct {
			X float64
			Y float64
		}{
			X: 100,
			Y: 100,
		},
		OutputPath: "testdata/output.png",
	}

	err = SavePositionImg(params)
	require.NoError(t, err)

	// cleanup
	defer os.Remove(params.OutputPath)
}
