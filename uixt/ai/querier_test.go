package ai

import (
	"context"
	"fmt"
	"testing"

	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Data structures for testing custom output schemas

// GameIcon represents a single icon in the game grid
type GameIcon struct {
	Name string `json:"name"` // Icon name (e.g., "beach_ball", "glove")
	Row  int    `json:"row"`  // Row position (0-based)
	Col  int    `json:"col"`  // Column position (0-based)
}

// GameGrid represents the complete game grid
type GameGrid struct {
	Grid  [][]GameIcon `json:"grid"`  // 2D array of game icons
	Rows  int          `json:"rows"`  // Number of rows
	Cols  int          `json:"cols"`  // Number of columns
	Icons []string     `json:"icons"` // List of unique icon names
}

// LianliankanResponse represents the structured response for lianliankan game analysis
type LianliankanResponse struct {
	Content string   `json:"content"` // Description of the analysis
	Thought string   `json:"thought"` // Reasoning process
	Data    GameGrid `json:"data"`    // Structured game grid data
}

// SimpleGameInfo represents basic game information
type SimpleGameInfo struct {
	Content    string   `json:"content"`    // Description
	Thought    string   `json:"thought"`    // Reasoning
	Rows       int      `json:"rows"`       // Number of rows
	Cols       int      `json:"cols"`       // Number of columns
	IconTypes  []string `json:"iconTypes"`  // List of icon types
	TotalIcons int      `json:"totalIcons"` // Total number of icons
}

// Additional data structures for comprehensive testing

// GameAnalysisResult represents structured analysis of a game interface
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

// UIElementsResult represents structured analysis of UI elements
type UIElementsResult struct {
	Content    string      `json:"content"`    // Description
	Thought    string      `json:"thought"`    // Reasoning
	Elements   []UIElement `json:"elements"`   // UI elements found
	Categories []string    `json:"categories"` // Categories of elements
}

type UIElement struct {
	Type        string      `json:"type"`        // Element type (button, text, image, etc.)
	Text        string      `json:"text"`        // Text content if any
	Description string      `json:"description"` // Element description
	BoundBox    BoundingBox `json:"boundBox"`    // Pixel coordinates
	Clickable   bool        `json:"clickable"`   // Whether element is clickable
	Visible     bool        `json:"visible"`     // Whether element is visible
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
		{
			name:    "malformed JSON that can be extracted but not parsed",
			content: `{"content": "test", "invalid": }`,
			expected: &QueryResult{
				Content: `{"content": "test", "invalid": }`,
				Thought: "Failed to parse as JSON, returning raw content",
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

func setupTestQuerier(t *testing.T) *Querier {
	ctx := context.Background()
	modelConfig, err := GetModelConfig(option.OPENAI_GPT_4O)
	require.NoError(t, err)
	querier, err := NewQuerier(ctx, modelConfig)
	require.NoError(t, err)
	return querier
}

// TestQueryBasicUsage demonstrates basic query functionality without custom schema
func TestQueryBasicUsage(t *testing.T) {
	querier := setupTestQuerier(t)

	// Load screenshot
	screenshot, size, err := builtin.LoadImage("testdata/llk_1.png")
	require.NoError(t, err)

	// Prepare query options
	opts := &QueryOptions{
		Query:      "这是一张连连看小游戏的界面，请将其转换为一个二维数组，数组中的每个元素包含图案名称及其坐标",
		Screenshot: screenshot,
		Size:       size,
	}

	// Perform query
	result, err := querier.Query(context.Background(), opts)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Content)
	assert.NotEmpty(t, result.Thought)
	assert.Nil(t, result.Data) // Should be nil for standard query

	t.Logf("Query Result:")
	t.Logf("Content: %s", result.Content)
	t.Logf("Thought: %s", result.Thought)
}

// TestQueryWithCustomSchema tests the query functionality with custom output schema
func TestQueryWithCustomSchema(t *testing.T) {
	querier := setupTestQuerier(t)

	// Load test image
	screenshot, size, err := builtin.LoadImage("testdata/llk_1.png")
	require.NoError(t, err)

	// Define custom output schema for lianliankan game
	outputSchema := LianliankanResponse{}

	// Prepare query options with custom schema
	opts := &QueryOptions{
		Query: `这是一张连连看小游戏的界面，请分析游戏界面并返回结构化数据：
1. 游戏网格的行数和列数
2. 每个位置的图案名称和坐标
3. 所有不同类型的图案列表
请将结果组织成二维数组格式，每个元素包含图案名称及其坐标位置。`,
		Screenshot:   screenshot,
		Size:         size,
		OutputSchema: outputSchema,
	}

	// Perform query
	result, err := querier.Query(context.Background(), opts)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Content)
	assert.NotEmpty(t, result.Thought)
	assert.NotNil(t, result.Data)

	t.Logf("Query result content: %s", result.Content)
	t.Logf("Query result thought: %s", result.Thought)
	t.Logf("Structured data: %+v", result.Data)

	// Try to parse the structured data
	if dataMap, ok := result.Data.(map[string]interface{}); ok {
		if gridData, exists := dataMap["data"]; exists {
			t.Logf("Game grid data: %+v", gridData)
		}
		if rows, exists := dataMap["rows"]; exists {
			t.Logf("Rows: %v", rows)
		}
		if cols, exists := dataMap["cols"]; exists {
			t.Logf("Cols: %v", cols)
		}
		if icons, exists := dataMap["icons"]; exists {
			t.Logf("Icon Types: %v", icons)
		}
	}
}

// TestQueryWithSimpleSchema tests with a simpler custom schema
func TestQueryWithSimpleSchema(t *testing.T) {
	querier := setupTestQuerier(t)

	// Load test image
	screenshot, size, err := builtin.LoadImage("testdata/llk_1.png")
	require.NoError(t, err)

	outputSchema := SimpleGameInfo{}

	// Prepare query options
	opts := &QueryOptions{
		Query:        "请分析这个连连看游戏界面，告诉我有多少行多少列，有哪些不同类型的图案，总共有多少个图标",
		Screenshot:   screenshot,
		Size:         size,
		OutputSchema: outputSchema,
	}

	// Perform query
	result, err := querier.Query(context.Background(), opts)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Content)
	assert.NotEmpty(t, result.Thought)
	assert.NotNil(t, result.Data)

	t.Logf("Simple schema result: %+v", result)
}

// TestQueryWithGameAnalysisSchema tests with comprehensive game analysis schema
func TestQueryWithGameAnalysisSchema(t *testing.T) {
	querier := setupTestQuerier(t)

	// Load test image
	screenshot, size, err := builtin.LoadImage("testdata/llk_1.png")
	require.NoError(t, err)

	outputSchema := GameAnalysisResult{}

	// Prepare query options
	opts := &QueryOptions{
		Query: `Analyze this game interface and provide structured information about:
1. The type of game
2. Grid dimensions (rows and columns)
3. All game elements with their positions and types
4. Statistics about element distribution`,
		Screenshot:   screenshot,
		Size:         size,
		OutputSchema: outputSchema,
	}

	// Perform query
	result, err := querier.Query(context.Background(), opts)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Content)
	assert.NotEmpty(t, result.Thought)
	assert.NotNil(t, result.Data)

	t.Logf("Game analysis result: %+v", result)
}

// TestQueryWithUIElementsSchema tests UI elements analysis
func TestQueryWithUIElementsSchema(t *testing.T) {
	querier := setupTestQuerier(t)

	// Load test image
	screenshot, size, err := builtin.LoadImage("testdata/llk_1.png")
	require.NoError(t, err)

	outputSchema := UIElementsResult{}

	// Prepare query options
	opts := &QueryOptions{
		Query: `Analyze this interface and identify all UI elements including:
1. Buttons and their text
2. Text labels and content
3. Images and icons
4. Interactive elements
5. Their positions and properties`,
		Screenshot:   screenshot,
		Size:         size,
		OutputSchema: outputSchema,
	}

	// Perform query
	result, err := querier.Query(context.Background(), opts)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Content)
	assert.NotEmpty(t, result.Thought)
	assert.NotNil(t, result.Data)

	t.Logf("UI elements analysis result: %+v", result)
}

// TestQuerySchemaComparison compares standard vs custom schema queries
func TestQuerySchemaComparison(t *testing.T) {
	querier := setupTestQuerier(t)

	screenshot, size, err := builtin.LoadImage("testdata/llk_1.png")
	require.NoError(t, err)

	query := "请分析这个连连看游戏界面的基本信息"

	// Standard query (without custom schema)
	t.Run("StandardQuery", func(t *testing.T) {
		standardOpts := &QueryOptions{
			Query:      query,
			Screenshot: screenshot,
			Size:       size,
			// No OutputSchema specified
		}

		standardResult, err := querier.Query(context.Background(), standardOpts)
		assert.NoError(t, err)
		assert.NotNil(t, standardResult)
		assert.NotEmpty(t, standardResult.Content)
		assert.NotEmpty(t, standardResult.Thought)
		assert.Nil(t, standardResult.Data) // Should be nil for standard query

		t.Logf("Standard Query Result:")
		t.Logf("Content: %s", standardResult.Content)
		t.Logf("Thought: %s", standardResult.Thought)
		t.Logf("Data: %+v", standardResult.Data)
	})

	// Custom schema query
	t.Run("CustomSchemaQuery", func(t *testing.T) {
		type GameInfo struct {
			Content string   `json:"content"`
			Thought string   `json:"thought"`
			Rows    int      `json:"rows"`
			Cols    int      `json:"cols"`
			Icons   []string `json:"icons"`
		}

		customOpts := &QueryOptions{
			Query:        query,
			Screenshot:   screenshot,
			Size:         size,
			OutputSchema: GameInfo{},
		}

		customResult, err := querier.Query(context.Background(), customOpts)
		assert.NoError(t, err)
		assert.NotNil(t, customResult)
		assert.NotEmpty(t, customResult.Content)
		assert.NotEmpty(t, customResult.Thought)
		assert.NotNil(t, customResult.Data) // Should contain structured data

		t.Logf("Custom Schema Query Result:")
		t.Logf("Content: %s", customResult.Content)
		t.Logf("Thought: %s", customResult.Thought)
		t.Logf("Structured Data: %+v", customResult.Data)
	})
}

// TestQueryWithDifferentPrompts tests various types of queries on the same screenshot
func TestQueryWithDifferentPrompts(t *testing.T) {
	querier := setupTestQuerier(t)

	// Load screenshot
	screenshot, size, err := builtin.LoadImage("testdata/llk_1.png")
	require.NoError(t, err)

	// Example queries
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
			t.Logf("Reasoning: %s", result.Thought)
		})
	}
}

// TestConvertQueryResultData tests the type conversion functionality
func TestConvertQueryResultData(t *testing.T) {
	// Test data structure
	type TestSchema struct {
		Content string   `json:"content"`
		Thought string   `json:"thought"`
		Count   int      `json:"count"`
		Items   []string `json:"items"`
	}

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

	t.Logf("Successfully converted data: %+v", converted)
}

// TestQueryResultDataConsistency tests that QueryResult.Data matches OutputSchema
func TestQueryResultDataConsistency(t *testing.T) {
	querier := setupTestQuerier(t)

	// Load test image
	screenshot, size, err := builtin.LoadImage("testdata/llk_1.png")
	require.NoError(t, err)

	// Define a simple test schema
	type TestGameInfo struct {
		Content string   `json:"content"`
		Thought string   `json:"thought"`
		Rows    int      `json:"rows"`
		Cols    int      `json:"cols"`
		Icons   []string `json:"icons"`
	}

	outputSchema := TestGameInfo{}

	// Prepare query options
	opts := &QueryOptions{
		Query:        "请分析这个连连看游戏界面，告诉我有多少行多少列，有哪些不同类型的图案",
		Screenshot:   screenshot,
		Size:         size,
		OutputSchema: outputSchema,
	}

	// Perform query
	result, err := querier.Query(context.Background(), opts)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Data)
	gameInfo, ok := result.Data.(*TestGameInfo)
	assert.True(t, ok)
	assert.NotNil(t, gameInfo)

	// Verify that the converted data has the expected structure
	assert.NotEmpty(t, gameInfo.Content)
	assert.NotEmpty(t, gameInfo.Thought)
	assert.NotEmpty(t, gameInfo.Rows)
	assert.NotEmpty(t, gameInfo.Cols)
	assert.NotEmpty(t, gameInfo.Icons)
}

// TestAutoTypeConversion tests that QueryResult.Data is automatically converted to the correct type
func TestAutoTypeConversion(t *testing.T) {
	// Test data structure
	type TestSchema struct {
		Content string   `json:"content"`
		Thought string   `json:"thought"`
		Count   int      `json:"count"`
		Items   []string `json:"items"`
	}

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

	t.Logf("Auto-converted data: %+v", typedData)
}

// TestDirectTypeAssertion tests that users can directly use type assertion on QueryResult.Data
func TestDirectTypeAssertion(t *testing.T) {
	// Test data structure
	type GameInfo struct {
		Content string   `json:"content"`
		Thought string   `json:"thought"`
		Rows    int      `json:"rows"`
		Cols    int      `json:"cols"`
		Icons   []string `json:"icons"`
	}

	// Simulate a JSON response
	jsonResponse := `{
		"content": "Game analysis complete",
		"thought": "Analyzed the game grid structure",
		"rows": 8,
		"cols": 10,
		"icons": ["apple", "banana", "cherry", "grape"]
	}`

	// Test the parseCustomSchemaResult function
	result, err := parseCustomSchemaResult(jsonResponse, GameInfo{})
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Data)

	// Users can now directly use type assertion
	if gameInfo, ok := result.Data.(*GameInfo); ok {
		assert.Equal(t, "Game analysis complete", gameInfo.Content)
		assert.Equal(t, "Analyzed the game grid structure", gameInfo.Thought)
		assert.Equal(t, 8, gameInfo.Rows)
		assert.Equal(t, 10, gameInfo.Cols)
		assert.Equal(t, []string{"apple", "banana", "cherry", "grape"}, gameInfo.Icons)
		t.Logf("Direct type assertion successful: %+v", gameInfo)
	} else {
		t.Fatalf("Type assertion failed, Data type: %T", result.Data)
	}
}
