package ai

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/httprunner/httprunner/v5/uixt/types"
	"github.com/rs/zerolog/log"
)

const (
	DefaultFactor = 1000
)

// UITARSContentParser parses the Thought/Action format response
type UITARSContentParser struct {
	systemPrompt  string
	actionMapping map[string]option.ActionName
}

func (p *UITARSContentParser) SystemPrompt() string {
	return p.systemPrompt
}

// ParseActionToStructureOutput parses the model output text into structured actions.
func (p *UITARSContentParser) Parse(content string, size types.Size) (*PlanningResult, error) {
	content = strings.TrimSpace(content)

	// Extract thought string
	thought := p.extractThought(content)

	// Extract action string
	actionStr, err := p.extractActionString(content)
	if err != nil {
		return nil, err
	}

	// Parse and process actions
	actions, err := p.parseActionString(actionStr, size)
	if err != nil {
		return nil, err
	}

	// Convert actions to tool calls
	toolCalls := convertActionsToToolCalls(actions, p.actionMapping)

	return &PlanningResult{
		ToolCalls:     toolCalls,
		ActionSummary: thought,
		Thought:       thought,
		Content:       content,
	}, nil
}

// extractThought extracts thought from the text
func (p *UITARSContentParser) extractThought(text string) string {
	re := regexp.MustCompile(`Thought:(.*?)\nAction:`)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

// extractActionString extracts the action string from the text
func (p *UITARSContentParser) extractActionString(text string) (string, error) {
	// Extract Action part using regex
	re := regexp.MustCompile(`Action:(.*?)(?:\n|$)`)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1]), nil
	}
	return "", fmt.Errorf("no Action: found")
}

// parseActionString parse and process actions
func (p *UITARSContentParser) parseActionString(actionStr string, size types.Size) ([]Action, error) {
	// Parse action type and raw arguments
	actionType, rawArgs, err := parseActionTypeAndArguments(actionStr)
	if err != nil {
		return nil, err
	}

	// Process and normalize arguments
	processedArgs, err := processActionArguments(rawArgs, size)
	if err != nil {
		return nil, err
	}

	// Convert processedArgs based on action type and coordinate parameters
	finalArgs, err := convertProcessedArgs(processedArgs, actionType)
	if err != nil {
		return nil, err
	}

	action := Action{
		ActionType:   actionType,
		ActionInputs: finalArgs,
	}

	return []Action{action}, nil
}

// normalizeCoordinatesFormat standardizes coordinate format in text (without pixel conversion)
func normalizeCoordinatesFormat(text string) string {
	// Convert point tags to coordinate format
	if strings.Contains(text, "<point>") {
		// support <point>x1 y1 x2 y2</point> or <point>x y</point>
		re := regexp.MustCompile(`<point>(\d+)\s+(\d+)(?:\s+(\d+)\s+(\d+))?</point>`)
		text = re.ReplaceAllStringFunc(text, func(match string) string {
			submatches := re.FindStringSubmatch(match)
			if submatches[3] != "" && submatches[4] != "" {
				// 4 numbers
				return fmt.Sprintf("(%s,%s,%s,%s)",
					submatches[1], submatches[2], submatches[3], submatches[4])
			}
			// 2 numbers
			return fmt.Sprintf("(%s,%s)", submatches[1], submatches[2])
		})
	}

	// Convert bbox tags to coordinate format
	if strings.Contains(text, "<bbox>") {
		// support <bbox>x1 y1 x2 y2</bbox>
		re := regexp.MustCompile(`<bbox>(\d+)\s+(\d+)\s+(\d+)\s+(\d+)</bbox>`)
		text = re.ReplaceAllStringFunc(text, func(match string) string {
			submatches := re.FindStringSubmatch(match)
			// 4 numbers for bbox
			return fmt.Sprintf("(%s,%s,%s,%s)",
				submatches[1], submatches[2], submatches[3], submatches[4])
		})
	}

	// Convert bracket format [x1, y1, x2, y2] to coordinate format
	if strings.Contains(text, "[") && strings.Contains(text, "]") {
		// support [x1, y1, x2, y2] format
		re := regexp.MustCompile(`\[(\d+),\s*(\d+),\s*(\d+),\s*(\d+)\]`)
		text = re.ReplaceAllStringFunc(text, func(match string) string {
			submatches := re.FindStringSubmatch(match)
			// 4 numbers for bracket format
			return fmt.Sprintf("(%s,%s,%s,%s)",
				submatches[1], submatches[2], submatches[3], submatches[4])
		})
	}

	return text
}

// convertRelativeToAbsolute converts relative coordinates to absolute pixel coordinates
// The coordinate system uses a 1000x1000 relative coordinate system as the base.
// This function maps relative coordinates to actual screen resolution coordinates.
//
// Conversion formula:
//   - For X coordinates: absolute_x = (relative_x / 1000) * screen_width
//   - For Y coordinates: absolute_y = (relative_y / 1000) * screen_height
//
// Example:
//   - Screen size: 1920x1080
//   - Relative coordinate: 500 (in 1000x1000 system)
//   - X conversion: 500/1000 * 1920 = 960 pixels
//   - Y conversion: 500/1000 * 1080 = 540 pixels
func convertRelativeToAbsolute(relativeCoord float64, isXCoord bool, size types.Size) float64 {
	if isXCoord {
		return math.Round((relativeCoord/DefaultFactor*float64(size.Width))*10) / 10
	}
	return math.Round((relativeCoord/DefaultFactor*float64(size.Height))*10) / 10
}

// parseActionTypeAndArguments extracts function name and raw parameter map from action string
// Input: "click(start_box='100,200,150,250')" or "click(start_point='100,200,150,250')"
// Output: actionType="click", rawArgs={"start_box": "100,200,150,250"}
func parseActionTypeAndArguments(actionStr string) (actionType string, rawArgs map[string]interface{}, err error) {
	// Parse action type and parameters
	actionParts := strings.SplitN(actionStr, "(", 2)
	if len(actionParts) < 2 {
		return "", nil, fmt.Errorf("not a function call")
	}

	actionType = strings.TrimSpace(actionParts[0])
	paramsText := strings.TrimSuffix(strings.TrimSpace(actionParts[1]), ")")

	// Parse string parameters to map
	rawArgs = make(map[string]interface{})
	if paramsText != "" {
		// Use regex to extract key=value pairs, handling quoted values properly
		re := regexp.MustCompile(`(\w+)\s*=\s*['"]([^'"]*?)['"]`)
		matches := re.FindAllStringSubmatch(paramsText, -1)
		for _, match := range matches {
			if len(match) >= 3 {
				key := strings.TrimSpace(match[1])
				value := strings.TrimSpace(match[2])

				// Apply parameter name mapping (legacy compatibility)
				key = normalizeParameterName(key)
				rawArgs[key] = value
			}
		}
	}

	return actionType, rawArgs, nil
}

// normalizeParameterName applies legacy parameter name mappings
func normalizeParameterName(paramName string) string {
	switch paramName {
	case "start_point":
		return "start_box"
	case "end_point":
		return "end_box"
	case "point":
		return "start_box"
	default:
		return paramName
	}
}

// processActionArguments processes raw arguments based on action type and parameter types
// Input: rawArgs={"start_box": "100,200,150,250"}
// Output: processedArgs={"start_box": [125.0, 225.0]} (converted to center point coordinates)
// For drag: rawArgs={"start_box": "100,200,150,250", "end_box": "300,400,350,450"}
// Output: processedArgs={"start_box": [125.0, 225.0], "end_box": [325.0, 425.0]} (both converted to center points)
func processActionArguments(rawArgs map[string]interface{}, size types.Size) (map[string]interface{}, error) {
	processedArgs := make(map[string]interface{})

	// Process each argument based on its type and context
	for paramName, paramValue := range rawArgs {
		processed, err := processArgument(paramName, paramValue, size)
		if err != nil {
			return nil, fmt.Errorf("failed to process argument %s: %w", paramName, err)
		}
		processedArgs[paramName] = processed
	}

	return processedArgs, nil
}

// Process a single argument based on its name and value
func processArgument(paramName string, paramValue interface{}, size types.Size) (interface{}, error) {
	// Handle coordinate parameters - convert bounding box to center point
	if isCoordinateParameter(paramName) {
		return normalizeActionCoordinatesToCenterPoint(paramValue, size)
	}

	// Handle other parameter types (content, key, direction, etc.)
	return normalizeStringParam(paramName, paramValue), nil
}

// Check if a parameter is a coordinate parameter
func isCoordinateParameter(paramName string) bool {
	return strings.Contains(paramName, "box") || strings.Contains(paramName, "point")
}

// convertProcessedArgs converts processed arguments based on action type and coordinate parameters
// For single start_box: {"start_box": [125.0, 225.0]} -> {"start_box": [125.0, 225.0]}
// For drag with start_box and end_box: {"start_box": [125.0, 225.0], "end_box": [325.0, 425.0]} -> {"start_box": [125.0, 225.0, 325.0, 425.0]}
func convertProcessedArgs(processedArgs map[string]interface{}, actionType string) (map[string]interface{}, error) {
	// Handle coordinate parameters based on action type
	startBox, hasStartBox := processedArgs["start_box"]
	endBox, hasEndBox := processedArgs["end_box"]

	// Check if this is a drag operation that should merge coordinates
	if hasStartBox && hasEndBox {
		// Drag operation: merge start_box and end_box into a single coordinate array
		startCoords, ok1 := startBox.([]float64)
		endCoords, ok2 := endBox.([]float64)

		if !ok1 || !ok2 {
			return nil, fmt.Errorf("invalid coordinate format for drag operation")
		}

		if len(startCoords) != 2 || len(endCoords) != 2 {
			return nil, fmt.Errorf("drag operation requires 2-element coordinate arrays, got start: %d, end: %d", len(startCoords), len(endCoords))
		}

		options := option.ActionOptions{
			FromX: startCoords[0],
			FromY: startCoords[1],
			ToX:   endCoords[0],
			ToY:   endCoords[1],
		}
		return options.ToMap(), nil
	}

	// For single coordinate operations, return the coordinate array directly
	if hasStartBox {
		startCoords, ok := startBox.([]float64)
		if !ok {
			return nil, fmt.Errorf("invalid coordinate format for single operation")
		}
		options := option.ActionOptions{
			X: startCoords[0],
			Y: startCoords[1],
		}
		return options.ToMap(), nil
	}

	// For non-coordinate operations, return the original arguments map
	// TODO
	finalArgs := make(map[string]interface{})
	for key, value := range processedArgs {
		finalArgs[key] = value
	}
	return finalArgs, nil
}

// normalizeActionCoordinates normalizes coordinates from various formats to actual pixel coordinates
func normalizeActionCoordinates(coordData interface{}, size types.Size) ([]float64, error) {
	switch v := coordData.(type) {
	case []interface{}:
		// Handle JSON array format: [x1, y1, x2, y2] or [x1, y1]
		if len(v) < 2 {
			return nil, fmt.Errorf("coordinate array must have at least 2 elements, got %d", len(v))
		}

		coords := make([]float64, len(v))
		for i, val := range v {
			switch num := val.(type) {
			case float64:
				// Convert relative coordinates to absolute coordinates using DefaultFactor
				if i%2 == 0 { // x coordinates
					coords[i] = convertRelativeToAbsolute(num, true, size)
				} else { // y coordinates
					coords[i] = convertRelativeToAbsolute(num, false, size)
				}
			case int:
				numFloat := float64(num)
				// Convert relative coordinates to absolute coordinates using DefaultFactor
				if i%2 == 0 { // x coordinates
					coords[i] = convertRelativeToAbsolute(numFloat, true, size)
				} else { // y coordinates
					coords[i] = convertRelativeToAbsolute(numFloat, false, size)
				}
			default:
				return nil, fmt.Errorf("coordinate value must be a number, got %T", val)
			}
		}
		return coords, nil

	case []float64:
		// Handle already parsed float64 slice
		coords := make([]float64, len(v))
		for i, val := range v {
			if i%2 == 0 { // x coordinates
				coords[i] = convertRelativeToAbsolute(val, true, size)
			} else { // y coordinates
				coords[i] = convertRelativeToAbsolute(val, false, size)
			}
		}
		return coords, nil

	case string:
		// Handle string format (from UI-TARS or string coordinates)
		return normalizeStringCoordinates(v, size)

	default:
		return nil, fmt.Errorf("unsupported coordinate format: %T", coordData)
	}
}

// normalizeStringParam normalizes string parameters, handling escape characters for content
func normalizeStringParam(paramName string, paramValue interface{}) interface{} {
	if paramValue == nil {
		return paramValue
	}

	// Convert to string if possible
	param, ok := paramValue.(string)
	if !ok {
		return paramValue // Return as-is if not a string
	}

	param = strings.TrimSpace(param)
	if param == "" {
		return param
	}

	// Handle escape characters for content parameter
	if paramName == "content" {
		param = strings.ReplaceAll(param, "\\n", "\n")
		param = strings.ReplaceAll(param, "\\\"", "\"")
		param = strings.ReplaceAll(param, "\\'", "'")
	}

	return param
}

// normalizeStringCoordinates normalizes coordinates from string format
func normalizeStringCoordinates(coordStr string, size types.Size) ([]float64, error) {
	// check empty string
	if coordStr == "" {
		return nil, fmt.Errorf("empty coordinate string")
	}

	// Apply coordinate format normalization using the shared function
	normalizedStr := normalizeCoordinatesFormat(coordStr)

	// Extract numbers from the normalized string using regex
	re := regexp.MustCompile(`\d+`)
	numbers := re.FindAllString(normalizedStr, -1)
	if len(numbers) >= 2 {
		coords := make([]float64, len(numbers))
		for i, numStr := range numbers {
			num, err := strconv.ParseFloat(numStr, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid coordinate: %s", numStr)
			}
			// Convert relative coordinates to absolute coordinates
			if i%2 == 0 { // x coordinates
				coords[i] = convertRelativeToAbsolute(num, true, size)
			} else { // y coordinates
				coords[i] = convertRelativeToAbsolute(num, false, size)
			}
		}
		return coords, nil
	}

	return nil, fmt.Errorf("invalid coordinate string format: %s", coordStr)
}

// normalizeActionCoordinatesToCenterPoint converts bounding box coordinates to center point coordinates
// Input: "100,200,150,250" (x1,y1,x2,y2) -> Output: [125.0, 225.0] (center point in absolute pixels)
// Input: "100,200" (x,y) -> Output: [100.0, 200.0] (point in absolute pixels)
func normalizeActionCoordinatesToCenterPoint(coordData interface{}, size types.Size) ([]float64, error) {
	// First normalize coordinates to get absolute pixel coordinates
	coords, err := normalizeActionCoordinates(coordData, size)
	if err != nil {
		return nil, err
	}

	// Convert bounding box to center point
	if len(coords) == 4 {
		// [x1, y1, x2, y2] -> [center_x, center_y]
		centerX := (coords[0] + coords[2]) / 2
		centerY := (coords[1] + coords[3]) / 2
		return []float64{centerX, centerY}, nil
	} else if len(coords) == 2 {
		// Already a point [x, y], return as-is
		return coords, nil
	} else {
		return nil, fmt.Errorf("invalid coordinate format: expected 2 or 4 coordinates, got %d", len(coords))
	}
}

// Action represents a parsed action with its context.
type Action struct {
	ActionType   string         `json:"action_type"` // map to option.ActionName
	ActionInputs map[string]any `json:"action_inputs"`
}

// convertActionsToToolCalls converts actions to tool calls
// This is a shared function used by both JSONContentParser and UITARSContentParser
func convertActionsToToolCalls(actions []Action, actionMapping map[string]option.ActionName) []schema.ToolCall {
	toolCalls := make([]schema.ToolCall, 0, len(actions))
	for _, action := range actions {
		jsonArgs, err := json.Marshal(action.ActionInputs)
		if err != nil {
			log.Error().Interface("action", action).Msg("failed to marshal action inputs")
			continue
		}
		actionName := string(actionMapping[action.ActionType])
		if actionName == "" {
			actionName = action.ActionType
		}
		toolCalls = append(toolCalls, schema.ToolCall{
			ID:   actionName + "_" + strconv.FormatInt(time.Now().Unix(), 10),
			Type: "function",
			Function: schema.FunctionCall{
				Name:      "uixt__" + actionName,
				Arguments: string(jsonArgs),
			},
		})
	}
	return toolCalls
}
