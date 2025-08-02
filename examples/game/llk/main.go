package llk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	hrp "github.com/httprunner/httprunner/v5"
	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/internal/config"
	"github.com/httprunner/httprunner/v5/uixt"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/rs/zerolog/log"
)

// GameElement represents a game element detected in the interface
type GameElement struct {
	Content    string     `json:"content"`    // Human-readable description
	Thought    string     `json:"thought"`    // AI reasoning process
	Dimensions Dimensions `json:"dimensions"` // Grid dimensions
	Elements   []Element  `json:"elements"`   // Game elements detected
}

// Dimensions represents grid dimensions
type Dimensions struct {
	Rows int `json:"rows"` // Number of rows
	Cols int `json:"cols"` // Number of columns
}

// Element represents a single game element
type Element struct {
	Type     string   `json:"type"`     // Element type/name
	Position Position `json:"position"` // Position in grid
	BoundBox BoundBox `json:"boundBox"` // Bounding box coordinates
}

// BoundBox represents bounding box coordinates
type BoundBox struct {
	X      float64 `json:"x"`      // X coordinate
	Y      float64 `json:"y"`      // Y coordinate
	Width  float64 `json:"width"`  // Box width
	Height float64 `json:"height"` // Box height
}

// Position represents grid position
type Position struct {
	Row int `json:"row"` // Row index (0-based)
	Col int `json:"col"` // Column index (0-based)
}

// LLKGameBot represents the main bot for playing LianLianKan game
type LLKGameBot struct {
	*hrp.UIXTRunner

	analyzeIndex int
}

// NewLLKGameBot creates a new LianLianKan game bot
func NewLLKGameBot(platform string, serial string) (*LLKGameBot, error) {
	// Create driver cache config
	config := hrp.UIXTConfig{
		DriverCacheConfig: uixt.DriverCacheConfig{
			Platform: platform,
			Serial:   serial,
			AIOptions: []option.AIServiceOption{
				option.WithCVService(option.CVServiceTypeVEDEM),
				option.WithLLMConfig(
					option.NewLLMServiceConfig(option.DOUBAO_1_5_UI_TARS_250328).
						WithQuerierModel(option.DOUBAO_SEED_1_6_250615),
				),
			},
		},
	}
	uixtRunner, err := hrp.NewUIXTRunner(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to create session runner: %w", err)
	}
	bot := &LLKGameBot{
		UIXTRunner:   uixtRunner,
		analyzeIndex: 0,
	}

	log.Info().Msg("LianLianKan game bot initialized successfully")
	return bot, nil
}

func (bot *LLKGameBot) EnterGame(ctx context.Context) error {
	_, err := bot.Session.RunStep(
		hrp.NewStep("进入游戏").
			Android().StartToGoal(
			"启动抖音，搜索「连了又连」小游戏，并启动游戏",
		),
	)
	if err != nil {
		return fmt.Errorf("failed to enter game: %w", err)
	}
	return nil
}

// AnalyzeGameInterface analyzes the game interface and extracts element information
func (bot *LLKGameBot) AnalyzeGameInterface() (*GameElement, error) {
	bot.analyzeIndex++
	query := `Analyze this LianLianKan (连连看) game interface and provide structured information about:
1. Grid dimensions (rows and columns)
2. All game elements with their positions and types`

	// Query the AI model
	result, err := bot.DriverExt.AIQuery(query,
		option.WithOutputSchema(GameElement{}))
	if err != nil {
		return nil, fmt.Errorf("failed to query AI model: %w", err)
	}

	// Convert result to GameElement
	gameElement, err := convertToGameElement(result)
	if err != nil {
		return nil, fmt.Errorf("failed to convert query result to GameElement: %w", err)
	}

	// Save debug data
	gameElementsPath := filepath.Join(config.GetConfig().ResultsPath(),
		fmt.Sprintf("game_elements_%d.json", bot.analyzeIndex))
	if err := builtin.Dump2JSON(gameElement, gameElementsPath); err != nil {
		log.Error().Err(err).Msg("failed to dump game elements data")
	} else {
		log.Info().Str("gameElementsPath", gameElementsPath).Msg("dumped game elements data")
	}

	return gameElement, nil
}

// convertToGameElement converts AI execution result to GameElement
func convertToGameElement(result *uixt.AIExecutionResult) (*GameElement, error) {
	if result == nil {
		return nil, fmt.Errorf("AI execution result is nil")
	}

	if result.QueryResult == nil {
		return nil, fmt.Errorf("query result is nil in AI execution result")
	}
	queryResult := result.QueryResult

	// Try direct conversion first
	if gameElement, ok := queryResult.Data.(*GameElement); ok {
		return gameElement, nil
	}

	// Convert to JSON and back for flexible parsing
	var gameElement GameElement
	var sourceData interface{}

	// Use Data if available, otherwise try Content
	if queryResult.Data != nil {
		sourceData = queryResult.Data
	} else if queryResult.Content != "" {
		var contentData map[string]interface{}
		if err := json.Unmarshal([]byte(queryResult.Content), &contentData); err != nil {
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

// SolveGame finds all possible pairs in the initial game state
func (bot *LLKGameBot) SolveGame(gameElement *GameElement) ([][]Element, error) {
	// Create solver instance
	solver := NewLLKSolver(gameElement)
	// Get all possible pairs from initial state (already validated)
	allPairs := solver.FindAllPairs()

	log.Info().Int("pairs", len(allPairs)).Msg("Found all valid pairs (passed game rules validation)")

	// Print solution details
	solver.printSolution()

	return allPairs, nil
}

// Play analyze game interface and solve game, then execute all clicks in sequence
func (bot *LLKGameBot) Play() error {
	// Analyze current screen
	gameElement, err := bot.AnalyzeGameInterface()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to analyze game interface")
	}

	// Solve game
	clickSequence, err := bot.SolveGame(gameElement)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to solve game")
	}

	systemPrompt := `连连看是一款经典的益智消除类小游戏，通常以图案或图标为主要元素。以下是连连看的基本规则说明：
1. 游戏目标: 玩家需要通过连接相同的图案或图标，将它们从游戏界面中消除。
2. 连接规则:
- 两个相同的图案可以通过不超过三条直线连接。
- 连接线可以水平或垂直，但不能斜线，也不能跨过其他图案。
- 连接线的转折次数不能超过两次。
3. 游戏界面:
- 游戏界面是一个矩形区域，内含多个图案或图标，排列成行和列；图案或图标在未选中状态下背景为白色，选中状态下背景为绿色。
- 游戏界面下方是道具区域，共有 3 种道具，从左到右分别是：「高亮显示」、「随机打乱」、「减少种类」。
4、游戏攻略：
- 游戏失败后，可观看广告视频，待屏幕右上角出现「领取成功」后，点击其右侧的 X 即可关闭广告，继续游戏

请严格按照以上游戏规则，仅完成如下2个相同图标的点击，完成后即结束，等待下一次任务：
`

	// Execute all clicks in sequence
	for _, pair := range clickSequence {
		prompt := fmt.Sprintf("点击连连看游戏界面上的 2 个相同图标 %s，坐标序列分别为 %+v, %+v",
			pair[0].Type, pair[0].Position, pair[1].Position)
		log.Info().Msg(prompt)

		_, err := bot.Session.RunStep(
			hrp.NewStep("").
				Android().StartToGoal(
				systemPrompt+prompt, option.WithMaxRetryTimes(2),
			),
		)
		if err != nil && !errors.Is(err, code.MaxRetryError) {
			log.Error().Err(err).Msg("Failed to click game interface")
			return err
		}
	}

	return nil
}

func (bot *LLKGameBot) GenerateReport() error {
	return bot.Session.GenerateReport()
}

// Close cleans up resources
func (bot *LLKGameBot) Close() error {
	if bot.DriverExt != nil {
		if err := bot.DriverExt.DeleteSession(); err != nil {
			log.Warn().Err(err).Msg("Warning: failed to delete driver session")
		}
		// Release driver from cache
		serial := bot.DriverExt.GetDevice().UUID()
		if err := uixt.ReleaseXTDriver(serial); err != nil {
			log.Warn().Err(err).Msg("Warning: failed to release driver")
		}
	}
	log.Info().Msg("LianLianKan game bot closed")
	return nil
}
