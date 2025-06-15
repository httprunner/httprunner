package llk

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

// LLKSolver represents a LianLianKan puzzle solver
type LLKSolver struct {
	board    [][]string  // Simplified board matrix with element types (immutable)
	elements [][]Element // Original elements with coordinates
	rows     int
	cols     int
	allPairs [][]Element // All possible pairs found in initial state
}

// NewLLKSolver creates a new LianLianKan solver
func NewLLKSolver(gameElement *GameElement) *LLKSolver {
	solver := &LLKSolver{
		rows: gameElement.Dimensions.Rows,
		cols: gameElement.Dimensions.Cols,
	}

	// Initialize board matrix and elements grid
	solver.board = make([][]string, solver.rows)
	solver.elements = make([][]Element, solver.rows)
	for i := range solver.board {
		solver.board[i] = make([]string, solver.cols)
		solver.elements[i] = make([]Element, solver.cols)
	}

	// Populate board and elements from gameElement
	// Check if data uses 1-based indexing by looking for any position >= dimensions
	// or by checking if position (1,1) exists (common indicator of 1-based indexing)
	uses1BasedIndexing := false
	for _, element := range gameElement.Elements {
		if element.Position.Row > solver.rows || element.Position.Col > solver.cols {
			uses1BasedIndexing = true
			break
		}
		// Also check if we have position (1,1) which is common in 1-based systems
		if element.Position.Row == 1 && element.Position.Col == 1 {
			uses1BasedIndexing = true
			break
		}
	}

	for _, element := range gameElement.Elements {
		row, col := element.Position.Row, element.Position.Col

		// Convert from 1-based to 0-based indexing if data uses 1-based
		if uses1BasedIndexing {
			row = row - 1
			col = col - 1
		}

		if solver.isValidPosition(row, col) {
			solver.board[row][col] = element.Type
			// Store original element (keep original 1-based coordinates)
			solver.elements[row][col] = element
		}
	}

	return solver
}

// findAllPairs finds all possible pairs that can be connected in the initial state (private method)
func (solver *LLKSolver) FindAllPairs() [][]Element {
	var pairs [][]Element
	used := make(map[string]bool) // Track used positions

	for row1 := 0; row1 < solver.rows; row1++ {
		for col1 := 0; col1 < solver.cols; col1++ {
			if solver.board[row1][col1] == "" {
				continue
			}

			// Skip if this position is already used
			pos1Key := fmt.Sprintf("%d,%d", row1, col1)
			if used[pos1Key] {
				continue
			}

			for row2 := 0; row2 < solver.rows; row2++ {
				for col2 := 0; col2 < solver.cols; col2++ {
					if solver.board[row2][col2] == "" {
						continue
					}

					// Avoid duplicate pairs by ensuring (row1,col1) < (row2,col2)
					if row1 > row2 || (row1 == row2 && col1 >= col2) {
						continue
					}

					// Skip if this position is already used
					pos2Key := fmt.Sprintf("%d,%d", row2, col2)
					if used[pos2Key] {
						continue
					}

					// Validate and add pair only if it passes all checks
					if solver.isValidPair(row1, col1, row2, col2) {
						element1 := solver.elements[row1][col1]
						element2 := solver.elements[row2][col2]
						pairs = append(pairs, []Element{element1, element2})

						// Mark both positions as used
						used[pos1Key] = true
						used[pos2Key] = true

						// Break out of inner loops since we found a pair for this element
						goto nextElement
					}
				}
			}
		nextElement:
		}
	}

	solver.allPairs = pairs
	return pairs
}

// isValidPosition checks if position is within board boundaries
func (solver *LLKSolver) isValidPosition(row, col int) bool {
	return row >= 0 && row < solver.rows && col >= 0 && col < solver.cols
}

// isEmpty checks if position is empty (already eliminated)
func (solver *LLKSolver) isEmpty(row, col int) bool {
	return solver.board[row][col] == ""
}

// canConnect checks if two positions can be connected according to LianLianKan rules
func (solver *LLKSolver) canConnect(row1, col1, row2, col2 int) bool {
	// Check if positions are valid and contain the same item
	if !solver.isValidPosition(row1, col1) ||
		!solver.isValidPosition(row2, col2) ||
		solver.isEmpty(row1, col1) ||
		solver.isEmpty(row2, col2) ||
		solver.board[row1][col1] != solver.board[row2][col2] {
		return false
	}

	// Same position
	if row1 == row2 && col1 == col2 {
		return false
	}

	// Try direct connection (0 turns)
	if solver.canConnectDirect(row1, col1, row2, col2) {
		return true
	}

	// Try one turn connection
	if solver.canConnectWithOneTurn(row1, col1, row2, col2) {
		return true
	}

	// Try two turns connection
	if solver.canConnectWithTwoTurns(row1, col1, row2, col2) {
		return true
	}

	return false
}

// canConnectHorizontal checks if two points can be connected horizontally
func (solver *LLKSolver) canConnectHorizontal(row, col1, col2 int) bool {
	startCol := col1
	endCol := col2
	if col1 > col2 {
		startCol = col2
		endCol = col1
	}

	// Check all positions between start and end (exclusive)
	for col := startCol + 1; col < endCol; col++ {
		if !solver.isEmpty(row, col) {
			return false
		}
	}
	return true
}

// canConnectVertical checks if two points can be connected vertically
func (solver *LLKSolver) canConnectVertical(col, row1, row2 int) bool {
	startRow := row1
	endRow := row2
	if row1 > row2 {
		startRow = row2
		endRow = row1
	}

	// Check all positions between start and end (exclusive)
	for row := startRow + 1; row < endRow; row++ {
		if !solver.isEmpty(row, col) {
			return false
		}
	}
	return true
}

// canConnectDirect checks if two points can be connected directly (straight line)
func (solver *LLKSolver) canConnectDirect(row1, col1, row2, col2 int) bool {
	// Same row - horizontal connection
	if row1 == row2 {
		return solver.canConnectHorizontal(row1, col1, col2)
	}

	// Same column - vertical connection
	if col1 == col2 {
		return solver.canConnectVertical(col1, row1, row2)
	}

	return false
}

// canConnectWithOneTurn checks if two points can be connected with one turn (L-shape)
func (solver *LLKSolver) canConnectWithOneTurn(row1, col1, row2, col2 int) bool {
	// Try corner at (row1, col2)
	corner1Row, corner1Col := row1, col2
	if solver.isEmpty(corner1Row, corner1Col) || (corner1Row == row2 && corner1Col == col2) {
		if solver.canConnectHorizontal(row1, col1, corner1Col) &&
			solver.canConnectVertical(corner1Col, corner1Row, row2) {
			return true
		}
	}

	// Try corner at (row2, col1)
	corner2Row, corner2Col := row2, col1
	if solver.isEmpty(corner2Row, corner2Col) || (corner2Row == row1 && corner2Col == col1) {
		if solver.canConnectVertical(col1, row1, corner2Row) &&
			solver.canConnectHorizontal(corner2Row, corner2Col, col2) {
			return true
		}
	}

	return false
}

// canConnectWithTwoTurns checks if two points can be connected with two turns (Z-shape)
func (solver *LLKSolver) canConnectWithTwoTurns(row1, col1, row2, col2 int) bool {
	// Try horizontal first, then vertical, then horizontal (internal paths)
	for col := 0; col < solver.cols; col++ {
		if col == col1 || col == col2 {
			continue
		}
		if solver.isEmpty(row1, col) && solver.isEmpty(row2, col) &&
			solver.canConnectHorizontal(row1, col1, col) &&
			solver.canConnectHorizontal(row2, col, col2) &&
			solver.canConnectVertical(col, row1, row2) {
			return true
		}
	}

	// Try vertical first, then horizontal, then vertical (internal paths)
	for row := 0; row < solver.rows; row++ {
		if row == row1 || row == row2 {
			continue
		}
		if solver.isEmpty(row, col1) && solver.isEmpty(row, col2) &&
			solver.canConnectVertical(col1, row1, row) &&
			solver.canConnectVertical(col2, row, row2) &&
			solver.canConnectHorizontal(row, col1, col2) {
			return true
		}
	}

	// Try boundary connections
	// Left boundary connection: go left -> down/up -> right
	if solver.canConnectToBoundary(row1, col1, "left") &&
		solver.canConnectToBoundary(row2, col2, "left") {
		return true
	}

	// Right boundary connection: go right -> down/up -> left
	if solver.canConnectToBoundary(row1, col1, "right") &&
		solver.canConnectToBoundary(row2, col2, "right") {
		return true
	}

	// Top boundary connection: go up -> left/right -> down
	if solver.canConnectToBoundary(row1, col1, "top") &&
		solver.canConnectToBoundary(row2, col2, "top") {
		return true
	}

	// Bottom boundary connection: go down -> left/right -> up
	if solver.canConnectToBoundary(row1, col1, "bottom") &&
		solver.canConnectToBoundary(row2, col2, "bottom") {
		return true
	}

	return false
}

// canConnectToBoundary checks if a position can connect to a boundary
func (solver *LLKSolver) canConnectToBoundary(row, col int, boundary string) bool {
	switch boundary {
	case "left":
		// Check if we can go horizontally left to column -1 (boundary)
		for c := col - 1; c >= 0; c-- {
			if !solver.isEmpty(row, c) {
				return false
			}
		}
		return true
	case "right":
		// Check if we can go horizontally right to column solver.cols (boundary)
		for c := col + 1; c < solver.cols; c++ {
			if !solver.isEmpty(row, c) {
				return false
			}
		}
		return true
	case "top":
		// Check if we can go vertically up to row -1 (boundary)
		for r := row - 1; r >= 0; r-- {
			if !solver.isEmpty(r, col) {
				return false
			}
		}
		return true
	case "bottom":
		// Check if we can go vertically down to row solver.rows (boundary)
		for r := row + 1; r < solver.rows; r++ {
			if !solver.isEmpty(r, col) {
				return false
			}
		}
		return true
	}
	return false
}

// isValidPair checks if two positions form a valid pair according to LianLianKan rules
func (solver *LLKSolver) isValidPair(row1, col1, row2, col2 int) bool {
	// Check positions are valid
	if !solver.isValidPosition(row1, col1) || !solver.isValidPosition(row2, col2) {
		return false
	}

	// Check positions are different
	if row1 == row2 && col1 == col2 {
		return false
	}

	// Check board cells are not empty
	if solver.board[row1][col1] == "" || solver.board[row2][col2] == "" {
		return false
	}

	// Check element types match and are not empty
	if solver.board[row1][col1] != solver.board[row2][col2] || solver.board[row1][col1] == "" {
		return false
	}

	// Check connectivity according to LianLianKan game rules
	return solver.canConnect(row1, col1, row2, col2)
}

// printSolution prints all available pairs for debugging
func (solver *LLKSolver) printSolution() {
	log.Info().Int("totalPairs", len(solver.allPairs)).
		Msg("All pairs validated and ready")

	for i, pair := range solver.allPairs {
		element1, element2 := pair[0], pair[1]
		log.Info().
			Int("pair", i+1).
			Str("elementType", element1.Type).
			Interface("pos1", element1.Position).
			Interface("pos2", element2.Position).
			Msg("Valid pair")
	}
}
