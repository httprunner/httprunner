package llk

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/internal/config"
	"github.com/httprunner/httprunner/v5/uixt"
	"github.com/httprunner/httprunner/v5/uixt/ai"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/httprunner/httprunner/v5/uixt/types"
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
}

// Position represents grid position
type Position struct {
	Row int `json:"row"` // Row index (0-based)
	Col int `json:"col"` // Column index (0-based)
}

// LLKGameBot represents the main bot for playing LianLianKan game
type LLKGameBot struct {
	Driver       *uixt.XTDriver
	ctx          context.Context
	analyzeIndex int
}

// NewLLKGameBot creates a new LianLianKan game bot
func NewLLKGameBot(platform string, serial string) (*LLKGameBot, error) {
	ctx := context.Background()

	// Create driver cache config
	config := uixt.DriverCacheConfig{
		Platform: platform,
		Serial:   serial,
		AIOptions: []option.AIServiceOption{
			option.WithCVService(option.CVServiceTypeVEDEM),
			option.WithLLMConfig(
				option.NewLLMServiceConfig(option.DOUBAO_1_5_UI_TARS_250328).
					WithQuerierModel(option.DOUBAO_SEED_1_6_250615),
			),
		},
	}

	// Get or create XTDriver
	driver, err := uixt.GetOrCreateXTDriver(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create XTDriver: %w", err)
	}

	// Initialize driver session
	if err := driver.InitSession(nil); err != nil {
		return nil, fmt.Errorf("failed to initialize driver session: %w", err)
	}

	bot := &LLKGameBot{
		ctx:    ctx,
		Driver: driver,
	}

	log.Info().Msg("LianLianKan game bot initialized successfully")
	log.Info().Str("platform", platform).Str("serial", driver.GetDevice().UUID()).Msg("Bot configuration")

	return bot, nil
}

func (bot *LLKGameBot) EnterGame(ctx context.Context) error {
	_, err := bot.Driver.StartToGoal(ctx, "启动抖音，搜索「连了又连」小游戏，并启动游戏")
	if err != nil {
		return fmt.Errorf("failed to enter game: %w", err)
	}
	return nil
}

// TakeScreenshot captures a screenshot and returns base64 encoded image with size
func (bot *LLKGameBot) TakeScreenshot() (string, types.Size, error) {
	// Take screenshot
	screenshotBuffer, err := bot.Driver.ScreenShot()
	if err != nil {
		return "", types.Size{}, fmt.Errorf("failed to take screenshot: %w", err)
	}

	// Get screen size
	size, err := bot.Driver.WindowSize()
	if err != nil {
		return "", types.Size{}, fmt.Errorf("failed to get window size: %w", err)
	}

	// Convert to base64
	screenshot := base64.StdEncoding.EncodeToString(screenshotBuffer.Bytes())
	screenshot = "data:image/png;base64," + screenshot

	log.Info().Int("width", size.Width).Int("height", size.Height).Msg("Screenshot captured successfully")
	return screenshot, size, nil
}

// AnalyzeGameInterface analyzes the game interface and extracts element information
func (bot *LLKGameBot) AnalyzeGameInterface() (*GameElement, error) {
	// Take screenshot
	screenshot, size, err := bot.TakeScreenshot()
	if err != nil {
		return nil, fmt.Errorf("failed to take screenshot: %w", err)
	}

	// Prepare query options with custom schema
	opts := &ai.QueryOptions{
		Query: `Analyze this LianLianKan (连连看) game interface and provide structured information about:
1. Grid dimensions (rows and columns)
2. All game elements with their positions and types`,
		Screenshot:   screenshot,
		Size:         size,
		OutputSchema: GameElement{},
	}
	bot.analyzeIndex++

	// Query the AI model
	result, err := bot.Driver.LLMService.Query(bot.ctx, opts)
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

// convertToGameElement converts AI query result to GameElement
func convertToGameElement(result *ai.QueryResult) (*GameElement, error) {
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

	// Execute all clicks in sequence
	for _, pair := range clickSequence {
		prompt := fmt.Sprintf("请点击连连看游戏界面上的 2 个相同图标 %s，坐标序列分别为 %+v, %+v",
			pair[0].Type, pair[0].Position, pair[1].Position)
		log.Info().Msg(prompt)
		_, err := bot.Driver.StartToGoal(context.Background(),
			prompt, option.WithMaxRetryTimes(2))
		if err != nil {
			log.Error().Err(err).Msg("Failed to click game interface")
		}

		time.Sleep(1 * time.Second)
	}

	return nil
}

// Close cleans up resources
func (bot *LLKGameBot) Close() error {
	if bot.Driver != nil {
		if err := bot.Driver.DeleteSession(); err != nil {
			log.Warn().Err(err).Msg("Warning: failed to delete driver session")
		}
		// Release driver from cache
		serial := bot.Driver.GetDevice().UUID()
		if err := uixt.ReleaseXTDriver(serial); err != nil {
			log.Warn().Err(err).Msg("Warning: failed to release driver")
		}
	}
	log.Info().Msg("LianLianKan game bot closed")
	return nil
}
