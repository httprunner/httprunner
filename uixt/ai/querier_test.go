package ai

import (
	"context"
	"fmt"
	"testing"

	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/httprunner/httprunner/v5/uixt/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test data structures

// GameInfo represents basic game information for testing
type GameInfo struct {
	Content    string   `json:"content"`    // Description
	Thought    string   `json:"thought"`    // Reasoning
	Rows       int      `json:"rows"`       // Number of rows
	Cols       int      `json:"cols"`       // Number of columns
	Icons      []string `json:"icons"`      // List of icon types
	TotalIcons int      `json:"totalIcons"` // Total number of icons
}

// GameAnalysisResult represents comprehensive game analysis for testing
type GameAnalysisResult struct {
	Content    string     `json:"content"`    // Human-readable description
	Thought    string     `json:"thought"`    // AI reasoning process
	GameType   string     `json:"gameType"`   // Type of game detected
	Dimensions Dimensions `json:"dimensions"` // Grid dimensions
	Elements   []Element  `json:"elements"`   // Game elements detected
	Statistics Statistics `json:"statistics"` // Game statistics
}

type Dimensions struct {
	Rows int `json:"rows"` // Number of rows
	Cols int `json:"cols"` // Number of columns
}

type Element struct {
	Type     string      `json:"type"`     // Element type/name
	Position Position    `json:"position"` // Position in grid
	BoundBox BoundingBox `json:"boundBox"` // Pixel coordinates
}

type Position struct {
	Row int `json:"row"` // Row index (0-based)
	Col int `json:"col"` // Column index (0-based)
}

type BoundingBox struct {
	X      int `json:"x"`      // Left coordinate
	Y      int `json:"y"`      // Top coordinate
	Width  int `json:"width"`  // Width in pixels
	Height int `json:"height"` // Height in pixels
}

type Statistics struct {
	TotalElements int         `json:"totalElements"` // Total number of elements
	UniqueTypes   int         `json:"uniqueTypes"`   // Number of unique element types
	TypeCounts    []TypeCount `json:"typeCounts"`    // Count of each type
}

type TypeCount struct {
	Type  string `json:"type"`  // Element type
	Count int    `json:"count"` // Number of occurrences
}

// Test helper functions

func setupTestQuerier(t *testing.T) *Querier {
	ctx := context.Background()
	modelConfig, err := GetModelConfig(option.OPENAI_GPT_4O)
	require.NoError(t, err)
	querier, err := NewQuerier(ctx, modelConfig)
	require.NoError(t, err)
	return querier
}

func loadTestImage(t *testing.T) (string, types.Size) {
	screenshot, size, err := builtin.LoadImage("testdata/llk_1.png")
	require.NoError(t, err)
	return screenshot, size
}

// Test functions

func TestParseQueryResult(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected *QueryResult
	}{
		{
			name: "valid JSON response",
			content: `{
				"content": "这是一个14行8列的连连看游戏界面，包含25种不同的图案",
				"thought": "通过分析图片，我识别出了游戏界面的结构和图案类型"
			}`,
			expected: &QueryResult{
				Content: "这是一个14行8列的连连看游戏界面，包含25种不同的图案",
				Thought: "通过分析图片，我识别出了游戏界面的结构和图案类型",
			},
		},
		{
			name:    "JSON in markdown",
			content: "```json\n{\n  \"content\": \"游戏界面分析结果\",\n  \"thought\": \"分析过程\"\n}\n```",
			expected: &QueryResult{
				Content: "游戏界面分析结果",
				Thought: "分析过程",
			},
		},
		{
			name:    "plain text response",
			content: "这是一个连连看游戏界面，包含多种图案。",
			expected: &QueryResult{
				Content: "这是一个连连看游戏界面，包含多种图案。",
				Thought: "Direct response from model",
			},
		},
		{
			name:    "invalid JSON",
			content: `{"content": "incomplete json", "missing_closing_brace": true`,
			expected: &QueryResult{
				Content: `{"content": "incomplete json", "missing_closing_brace": true`,
				Thought: "Direct response from model",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseQueryResult(tt.content)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected.Content, result.Content)
			assert.Equal(t, tt.expected.Thought, result.Thought)
		})
	}
}

// TestQueryFunctionality tests both basic and custom schema query functionality
func TestQueryFunctionality(t *testing.T) {
	querier := setupTestQuerier(t)
	screenshot, size := loadTestImage(t)

	t.Run("BasicQuery", func(t *testing.T) {
		opts := &QueryOptions{
			Query:      "这是一张连连看小游戏的界面，请分析游戏界面的基本信息",
			Screenshot: screenshot,
			Size:       size,
		}

		result, err := querier.Query(context.Background(), opts)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.Content)
		assert.NotEmpty(t, result.Thought)
		assert.Nil(t, result.Data) // Should be nil for standard query

		t.Logf("Basic Query Result: %s", result.Content)
	})

	t.Run("CustomSchemaQuery", func(t *testing.T) {
		opts := &QueryOptions{
			Query:        "请分析这个连连看游戏界面，告诉我有多少行多少列，有哪些不同类型的图案",
			Screenshot:   screenshot,
			Size:         size,
			OutputSchema: GameInfo{},
		}

		result, err := querier.Query(context.Background(), opts)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.Content)
		assert.NotEmpty(t, result.Thought)
		assert.NotNil(t, result.Data) // Should contain structured data

		// Verify structured data
		gameInfo, ok := result.Data.(*GameInfo)
		assert.True(t, ok)
		assert.NotNil(t, gameInfo)
		assert.NotEmpty(t, gameInfo.Content)
		assert.NotEmpty(t, gameInfo.Thought)
		assert.Equal(t, 4, gameInfo.Rows)
		assert.Equal(t, 3, gameInfo.Cols)
		assert.Equal(t, 5, gameInfo.TotalIcons)

		t.Logf("Custom Schema Query Result: %+v", gameInfo)
	})

	t.Run("ComprehensiveAnalysis", func(t *testing.T) {
		opts := &QueryOptions{
			Query: `Analyze this game interface and provide structured information about:
1. The type of game
2. Grid dimensions (rows and columns)
3. All game elements with their positions and types
4. Statistics about element distribution`,
			Screenshot:   screenshot,
			Size:         size,
			OutputSchema: GameAnalysisResult{},
		}

		result, err := querier.Query(context.Background(), opts)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.Content)
		assert.NotEmpty(t, result.Thought)
		assert.NotNil(t, result.Data)

		gameAnalysisResult, ok := result.Data.(*GameAnalysisResult)
		assert.True(t, ok)
		assert.NotNil(t, gameAnalysisResult)
		assert.NotEmpty(t, gameAnalysisResult.Content)
		assert.NotEmpty(t, gameAnalysisResult.Thought)
		assert.NotEmpty(t, gameAnalysisResult.GameType)
		assert.Equal(t, 4, gameAnalysisResult.Dimensions.Rows)
		assert.Equal(t, 3, gameAnalysisResult.Dimensions.Cols)
		assert.Equal(t, 12, len(gameAnalysisResult.Elements))

		t.Logf("Comprehensive Analysis Result: %+v", result.Data)
	})
}

// TestQueryWithDifferentPrompts tests various types of queries on the same screenshot
func TestQueryWithDifferentPrompts(t *testing.T) {
	querier := setupTestQuerier(t)
	screenshot, size := loadTestImage(t)

	queries := []string{
		"请描述这张图片中的内容",
		"这个游戏界面有多少行多少列？",
		"请识别图片中所有不同类型的图案",
		"请找出可以消除的图案对",
	}

	for i, query := range queries {
		t.Run(fmt.Sprintf("Query_%d", i+1), func(t *testing.T) {
			opts := &QueryOptions{
				Query:      query,
				Screenshot: screenshot,
				Size:       size,
			}

			result, err := querier.Query(context.Background(), opts)
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.NotEmpty(t, result.Content)
			assert.NotEmpty(t, result.Thought)

			t.Logf("Query %d: %s", i+1, query)
			t.Logf("Answer: %s", result.Content)
		})
	}
}

// TestTypeConversionAndAssertion tests data type conversion and assertion functionality
func TestTypeConversionAndAssertion(t *testing.T) {
	// Test data structure
	type TestSchema struct {
		Content string   `json:"content"`
		Thought string   `json:"thought"`
		Count   int      `json:"count"`
		Items   []string `json:"items"`
	}

	t.Run("ConvertQueryResultData", func(t *testing.T) {
		// Create a QueryResult with structured data
		testData := &TestSchema{
			Content: "Test content",
			Thought: "Test thought",
			Count:   5,
			Items:   []string{"item1", "item2", "item3"},
		}

		result := &QueryResult{
			Content: "Test content",
			Thought: "Test thought",
			Data:    testData,
		}

		// Test type conversion
		converted, err := ConvertQueryResultData[TestSchema](result)
		assert.NoError(t, err)
		assert.NotNil(t, converted)
		assert.Equal(t, "Test content", converted.Content)
		assert.Equal(t, "Test thought", converted.Thought)
		assert.Equal(t, 5, converted.Count)
		assert.Equal(t, []string{"item1", "item2", "item3"}, converted.Items)
	})

	t.Run("AutoTypeConversion", func(t *testing.T) {
		// Simulate a JSON response from the model
		jsonResponse := `{
			"content": "Test content from model",
			"thought": "Test reasoning process",
			"count": 42,
			"items": ["apple", "banana", "cherry"]
		}`

		// Test the parseCustomSchemaResult function directly
		result, err := parseCustomSchemaResult(jsonResponse, TestSchema{})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Data)

		// Verify that Data is automatically converted to the correct type
		typedData, ok := result.Data.(*TestSchema)
		assert.True(t, ok, "Data should be automatically converted to *TestSchema")
		assert.NotNil(t, typedData)

		// Verify the content
		assert.Equal(t, "Test content from model", typedData.Content)
		assert.Equal(t, "Test reasoning process", typedData.Thought)
		assert.Equal(t, 42, typedData.Count)
		assert.Equal(t, []string{"apple", "banana", "cherry"}, typedData.Items)

		// Verify that QueryResult fields are also populated
		assert.Equal(t, "Test content from model", result.Content)
		assert.Equal(t, "Test reasoning process", result.Thought)
	})

	t.Run("DirectTypeAssertion", func(t *testing.T) {
		// Simulate a JSON response
		jsonResponse := `{
			"content": "Game analysis complete",
			"thought": "Analyzed the game grid structure",
			"count": 100,
			"items": ["apple", "banana", "cherry", "grape"]
		}`

		// Test the parseCustomSchemaResult function
		result, err := parseCustomSchemaResult(jsonResponse, TestSchema{})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Data)

		// Users can now directly use type assertion
		if testData, ok := result.Data.(*TestSchema); ok {
			assert.Equal(t, "Game analysis complete", testData.Content)
			assert.Equal(t, "Analyzed the game grid structure", testData.Thought)
			assert.Equal(t, 100, testData.Count)
			assert.Equal(t, []string{"apple", "banana", "cherry", "grape"}, testData.Items)
		} else {
			t.Fatalf("Type assertion failed, Data type: %T", result.Data)
		}
	})
}
