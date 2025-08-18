//go:build localtest

package llk

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hrp "github.com/httprunner/httprunner/v5"
	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/uixt/ai"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/httprunner/httprunner/v5/uixt/types"
)

func TestGameLianliankan(t *testing.T) {
	userInstruction := `连连看是一款经典的益智消除类小游戏，通常以图案或图标为主要元素。以下是连连看的基本规则说明：
1. 游戏目标: 玩家需要通过连接相同的图案或图标，将它们从游戏界面中消除。
2. 连接规则:
- 两个相同的图案可以通过不超过三条直线连接。
- 连接线可以水平或垂直，但不能斜线，也不能跨过其他图案。
- 连接线的转折次数不能超过两次。
3. 游戏界面:
- 游戏界面是一个矩形区域，内含多个图案或图标，排列成行和列；图案或图标在未选中状态下背景为白色，选中状态下背景为绿色。
- 游戏界面下方是道具区域，共有 3 种道具，从左到右分别是：「高亮显示」、「随机打乱」、「减少种类」。
4、游戏攻略：建议多次使用道具，可以降低游戏难度
- 优先使用「减少种类」道具，可以将图案种类随机减少一种
- 遇到困难时，推荐使用「随机打乱」道具，可以获得很多新的消除机会
- 观看广告视频，待屏幕右上角出现「领取成功」后，点击其右侧的 X 即可关闭广告，继续游戏

请严格按照以上游戏规则，开始游戏
`

	testCase := &hrp.TestCase{
		Config: hrp.NewConfig("连连看小游戏自动化测试").
			SetLLMService(option.DOUBAO_1_5_THINKING_VISION_PRO_250428),
		TestSteps: []hrp.IStep{
			hrp.NewStep("启动抖音「连了又连」小游戏").
				Android().
				StartToGoal("启动抖音，搜索「连了又连」小游戏，并启动游戏").
				Validate().
				AssertAI("当前位于抖音「连了又连」小游戏页面"),
			hrp.NewStep("开始游戏").
				Android().
				StartToGoal(userInstruction, option.WithTimeLimit(300)),
		},
	}
	err := testCase.Dump2JSON("game_llk.json")
	require.Nil(t, err)

	// err = hrp.NewRunner(t).Run(testCase)
	// assert.Nil(t, err)
}

// convertToGameElementFromQueryResult converts AI query result to GameElement for testing
func convertToGameElementFromQueryResult(result *ai.QueryResult) (*GameElement, error) {
	if result == nil {
		return nil, fmt.Errorf("query result is nil")
	}

	// Try direct conversion first
	if gameElement, ok := result.Data.(*GameElement); ok {
		return gameElement, nil
	}

	// Convert to JSON and back for flexible parsing
	var gameElement GameElement
	var sourceData interface{}

	// Use Data if available, otherwise try Content
	if result.Data != nil {
		sourceData = result.Data
	} else if result.Content != "" {
		var contentData map[string]interface{}
		if err := json.Unmarshal([]byte(result.Content), &contentData); err != nil {
			return nil, fmt.Errorf("failed to parse JSON from Content: %w", err)
		}
		sourceData = contentData
	} else {
		return nil, fmt.Errorf("no data available in query result")
	}

	// Convert via JSON marshaling/unmarshaling
	jsonBytes, err := json.Marshal(sourceData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result data: %w", err)
	}

	if err := json.Unmarshal(jsonBytes, &gameElement); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to GameElement: %w", err)
	}

	return &gameElement, nil
}

// loadTestImage loads the test image from testdata
func loadTestImage(t *testing.T) (string, types.Size) {
	screenshot, size, err := builtin.LoadImage("../../../uixt/ai/testdata/llk_1.png")
	require.NoError(t, err)
	return screenshot, size
}

// createAIQueryer creates a AI queryer with AI analysis capability
func createAIQueryer(t *testing.T) *ai.Querier {
	ctx := context.Background()
	modelConfig, err := ai.GetModelConfig(option.DOUBAO_SEED_1_6_250615)
	require.NoError(t, err)
	querier, err := ai.NewQuerier(ctx, modelConfig)
	require.NoError(t, err)
	return querier
}

// TestLLKGameBot_AnalyzeGameInterface comprehensive test for game interface analysis
func TestLLKGameBot_AnalyzeGameInterface(t *testing.T) {
	t.Run("AnalyzeWithTestImage", func(t *testing.T) {
		// Create test bot and load test image
		querier := createAIQueryer(t)
		screenshot, size := loadTestImage(t)
		t.Logf("Loaded test image with size: %dx%d", size.Width, size.Height)

		// Prepare query options for AI analysis
		opts := &ai.QueryOptions{
			Query: `Analyze this LianLianKan (连连看) game interface and provide CONCISE structured information:

1. Game type: "LianLianKan"
2. Grid dimensions (rows x columns) - CRITICAL: rows are horizontal lines, columns are vertical lines
3. Game elements with positions and types - LIMIT to essential info only
4. Bounding boxes - use approximate coordinates

REQUIREMENTS:
- Count ROWS as horizontal lines (top to bottom)
- Count COLUMNS as vertical lines (left to right)
- Position: row=0 is top, col=0 is left
- Keep response SHORT to avoid truncation
- Use simple element type names (max 10 chars)
- Omit detailed descriptions

Return JSON with: content, dimensions{rows,cols}, elements[{type,position{row,col},boundBox{x,y,width,height}}], statistics{totalElements,uniqueTypes}.`,
			Screenshot:   screenshot,
			Size:         size,
			OutputSchema: GameElement{},
		}

		// Query AI model and convert result
		result, err := querier.Query(context.Background(), opts)
		require.NoError(t, err, "Failed to query AI model")

		// Convert result using enhanced compatibility logic
		gameElement, err := convertToGameElementFromQueryResult(result)
		require.NoError(t, err, "Failed to convert query result to GameElement")
		require.NotNil(t, gameElement, "GameElement should not be nil")

		// Log analysis results
		t.Logf("\n=== Game Interface Analysis Results ===")
		t.Logf("Dimensions: %dx%d", gameElement.Dimensions.Rows, gameElement.Dimensions.Cols)

		// Basic validations
		assert.NotEmpty(t, gameElement.Content, "Content should not be empty")
		assert.Greater(t, gameElement.Dimensions.Rows, 0, "Rows should be greater than 0")
		assert.Greater(t, gameElement.Dimensions.Cols, 0, "Cols should be greater than 0")
		assert.Greater(t, len(gameElement.Elements), 0, "Should have detected elements")

		// Test solver integration
		t.Logf("\n=== Solver Integration Test ===")
		solver := NewLLKSolver(gameElement)
		require.NotNil(t, solver, "Solver should be created successfully")

		pairs := solver.FindAllPairs()
		t.Logf("Solver found %d valid matching pairs", len(pairs))

		// Log sample element details
		t.Logf("\n=== Sample Elements ===")
		for i, element := range gameElement.Elements {
			if i < 5 { // Show first 5 elements
				t.Logf("Element %d: %s at grid(%d,%d)",
					i+1, element.Type,
					element.Position.Row, element.Position.Col)
			}
		}
		if len(gameElement.Elements) > 5 {
			t.Logf("... and %d more elements", len(gameElement.Elements)-5)
		}

		t.Logf("\n=== Analysis Test Completed Successfully ===")
	})
}

// TestLLKGameBot_RealDevice test with real Android device
func TestLLKGameBot_RealDevice(t *testing.T) {
	t.Run("CreateAndAnalyze", func(t *testing.T) {
		// Create game bot with real device
		bot, err := NewLLKGameBot("android", "")
		require.NoError(t, err, "Failed to create LLKGameBot")
		defer bot.Close()

		// err = bot.EnterGame(context.Background())
		// require.NoError(t, err, "Failed to enter game")

		err = bot.Play()
		require.NoError(t, err, "Failed to play game")
	})
}
