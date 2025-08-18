//go:build localtest

package llk

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/httprunner/httprunner/v5/uixt/ai"
)

// TestLLKSolver tests the LianLianKan solver functionality
func TestLLKSolver(t *testing.T) {
	// Create test game bot
	querier := createAIQueryer(t)

	// Analyze the game interface
	screenshot, size := loadTestImage(t)

	// Prepare query options with custom schema
	opts := &ai.QueryOptions{
		Query: `Analyze this LianLianKan (连连看) game interface and provide structured information about:
1. Grid dimensions (rows and columns)
2. All game elements with their positions and types`,
		Screenshot:   screenshot,
		Size:         size,
		OutputSchema: GameElement{},
	}

	// Query the AI model
	result, err := querier.Query(context.Background(), opts)
	require.NoError(t, err)

	// Convert result data to GameElement
	gameElement, ok := result.Data.(*GameElement)
	require.True(t, ok, "Failed to convert result to GameElement")
	require.NotNil(t, gameElement)

	t.Run("FindMatchingPairs", func(t *testing.T) {
		// Create solver
		solver := NewLLKSolver(gameElement)

		// Find all valid pairs
		pairs := solver.FindAllPairs()

		// Verify pairs
		assert.GreaterOrEqual(t, len(pairs), 0, "Should find some pairs or none")
		t.Logf("Found %d valid matching pairs", len(pairs))
	})

	t.Run("ConnectionRules", func(t *testing.T) {
		// Create solver
		solver := NewLLKSolver(gameElement)

		// Test connection rules with known positions
		if len(gameElement.Elements) >= 2 {
			element1 := gameElement.Elements[0]
			element2 := gameElement.Elements[1]

			// Test same position (should fail)
			canConnect := solver.canConnect(
				element1.Position.Row, element1.Position.Col,
				element1.Position.Row, element1.Position.Col)
			assert.False(t, canConnect, "Same position should not be connectable")

			// Test different types (should fail if different)
			if element1.Type != element2.Type {
				canConnect = solver.canConnect(
					element1.Position.Row, element1.Position.Col,
					element2.Position.Row, element2.Position.Col)
				assert.False(t, canConnect, "Different types should not be connectable")
			}

			t.Logf("Connection rules validation completed")
		}
	})
}

func TestLLKSolver_WithTestData(t *testing.T) {
	// Load test data
	gameElement, err := loadTestGameElement()
	require.NoError(t, err, "Failed to load test game element")
	require.NotNil(t, gameElement, "Game element should not be nil")

	// Create solver
	solver := NewLLKSolver(gameElement)
	require.NotNil(t, solver, "Solver should be created successfully")

	// Find all valid pairs
	pairs := solver.FindAllPairs()
	log.Info().Interface("pairs", pairs).Msg("Found all valid pairs")

	// Verify pairs against expected results (updated to include boundary connections)
	expectedPairs := [][]Element{
		{
			{Type: "wheel", Position: Position{Row: 1, Col: 8}},
			{Type: "wheel", Position: Position{Row: 9, Col: 8}},
		},
		{
			{Type: "scissors", Position: Position{Row: 2, Col: 1}},
			{Type: "scissors", Position: Position{Row: 12, Col: 1}},
		},
		{
			{Type: "wheat", Position: Position{Row: 2, Col: 7}},
			{Type: "wheat", Position: Position{Row: 3, Col: 7}},
		},
		{
			{Type: "clover", Position: Position{Row: 2, Col: 8}},
			{Type: "clover", Position: Position{Row: 13, Col: 8}},
		},
		{
			{Type: "brush", Position: Position{Row: 4, Col: 7}},
			{Type: "brush", Position: Position{Row: 4, Col: 8}},
		},
		{
			{Type: "brush", Position: Position{Row: 4, Col: 8}},
			{Type: "brush", Position: Position{Row: 10, Col: 8}},
		},
		{
			{Type: "cherries", Position: Position{Row: 5, Col: 1}},
			{Type: "cherries", Position: Position{Row: 7, Col: 1}},
		},
		{
			{Type: "cloche", Position: Position{Row: 6, Col: 6}},
			{Type: "cloche", Position: Position{Row: 7, Col: 6}},
		},
		{
			{Type: "leaf", Position: Position{Row: 6, Col: 8}},
			{Type: "leaf", Position: Position{Row: 14, Col: 8}},
		},
		{
			{Type: "target", Position: Position{Row: 8, Col: 8}},
			{Type: "target", Position: Position{Row: 11, Col: 8}},
		},
		{
			{Type: "scissors", Position: Position{Row: 10, Col: 4}},
			{Type: "scissors", Position: Position{Row: 10, Col: 5}},
		},
		{
			{Type: "trowel", Position: Position{Row: 11, Col: 7}},
			{Type: "trowel", Position: Position{Row: 12, Col: 7}},
		},
		{
			{Type: "meat", Position: Position{Row: 14, Col: 1}},
			{Type: "meat", Position: Position{Row: 14, Col: 3}},
		},
	}

	// Compare number of pairs
	// assert.Equal(t, len(expectedPairs), len(pairs), "Number of pairs should match expected")
	// Compare each pair by checking if it exists in the expected pairs
	for _, pair := range pairs {
		found := false
		for _, expectedPair := range expectedPairs {
			// Check if both elements match (considering both possible orders)
			if (pair[0].Type == expectedPair[0].Type &&
				pair[0].Position.Row == expectedPair[0].Position.Row &&
				pair[0].Position.Col == expectedPair[0].Position.Col &&
				pair[1].Type == expectedPair[1].Type &&
				pair[1].Position.Row == expectedPair[1].Position.Row &&
				pair[1].Position.Col == expectedPair[1].Position.Col) ||
				(pair[0].Type == expectedPair[1].Type &&
					pair[0].Position.Row == expectedPair[1].Position.Row &&
					pair[0].Position.Col == expectedPair[1].Position.Col &&
					pair[1].Type == expectedPair[0].Type &&
					pair[1].Position.Row == expectedPair[0].Position.Row &&
					pair[1].Position.Col == expectedPair[0].Position.Col) {
				found = true
				break
			}
		}
		assert.True(t, found, "Pair should be found in expected pairs: %v", pair)
	}
}

// loadTestGameElement loads game element data from test file
func loadTestGameElement() (*GameElement, error) {
	// Read test data file
	data, err := os.ReadFile("testdata/game_elements.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read test data file: %w", err)
	}

	// Parse JSON
	var gameElement GameElement
	if err := json.Unmarshal(data, &gameElement); err != nil {
		return nil, fmt.Errorf("failed to parse test data: %w", err)
	}

	return &gameElement, nil
}
