package llk

import (
	"context"
	"os"
	"testing"

	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/uixt/ai"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/httprunner/httprunner/v5/uixt/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// hasRequiredEnvVars checks if the required environment variables are set for testing
func hasRequiredEnvVars() bool {
	// Check for OpenAI environment variables
	if os.Getenv("OPENAI_BASE_URL") != "" && os.Getenv("OPENAI_API_KEY") != "" {
		return true
	}
	// Check for GPT-4O specific environment variables
	if os.Getenv("OPENAI_GPT_4O_BASE_URL") != "" && os.Getenv("OPENAI_GPT_4O_API_KEY") != "" {
		return true
	}
	return false
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
	if !hasRequiredEnvVars() {
		t.Skip("Skipping test: required environment variables not set")
	}

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
		gameElement, err := convertToGameElement(result)
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
