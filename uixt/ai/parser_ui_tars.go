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
	"github.com/httprunner/httprunner/v5/uixt/types"
	"github.com/rs/zerolog/log"
)

// reference:
// https://github.com/bytedance/UI-TARS/blob/main/codes/ui_tars/action_parser.py

const (
	DefaultFactor = 1000
)

// UITARSContentParser parses the Thought/Action format response
type UITARSContentParser struct {
	systemPrompt string
}

func (p *UITARSContentParser) SystemPrompt() string {
	return p.systemPrompt
}

// ParseActionToStructureOutput parses the model output text into structured actions.
func (p *UITARSContentParser) Parse(content string, size types.Size) (*PlanningResult, error) {
	text := strings.TrimSpace(content)

	// Extract thought/reflection
	thought := p.extractThought(text)

	// Normalize text first
	normalizedText := p.normalizeCoordinates(text)

	// Get action string from normalized text
	actionStr, err := p.extractActionString(normalizedText)
	if err != nil {
		return nil, err
	}

	// Parse actions directly
	actions, err := p.parseActionString(actionStr, size)
	if err != nil {
		return nil, err
	}

	// Convert actions to tool calls
	toolCalls := p.convertActionsToToolCalls(actions)

	return &PlanningResult{
		ToolCalls:     toolCalls,
		Actions:       actions,
		ActionSummary: thought,
		Thought:       thought,
		Text:          normalizedText,
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

// normalizeCoordinates normalizes the text by converting points to coordinates and replacing keywords
func (p *UITARSContentParser) normalizeCoordinates(text string) string {
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

	// Legacy parameter name replacements (keep for backward compatibility)
	text = strings.ReplaceAll(text, "start_point=", "start_box=")
	text = strings.ReplaceAll(text, "end_point=", "end_box=")
	text = strings.ReplaceAll(text, "point=", "start_box=")

	return text
}

// parseActionString parses the action string directly
func (p *UITARSContentParser) parseActionString(actionStr string, size types.Size) ([]Action, error) {
	actions := make([]Action, 0, 1)

	// Parse action type and parameters
	actionParts := strings.SplitN(actionStr, "(", 2)
	if len(actionParts) < 2 {
		return nil, fmt.Errorf("not a function call")
	}

	funcName := strings.TrimSpace(actionParts[0])
	paramsText := strings.TrimSuffix(strings.TrimSpace(actionParts[1]), ")")

	args := make(map[string]string)
	if paramsText != "" {
		// Use regex to extract key=value pairs, handling quoted values properly
		re := regexp.MustCompile(`(\w+)\s*=\s*['"]([^'"]*?)['"]`)
		matches := re.FindAllStringSubmatch(paramsText, -1)
		for _, match := range matches {
			if len(match) >= 3 {
				key := strings.TrimSpace(match[1])
				value := strings.TrimSpace(match[2])
				args[key] = value
			}
		}
	}

	actionInputs, err := p.parseActionInputs(args, size)
	if err != nil {
		return nil, err
	}

	actions = append(actions, Action{
		ActionType:   funcName,
		ActionInputs: actionInputs,
	})

	return actions, nil
}

// parseActionInputs parses action parameters and converts coordinates
func (p *UITARSContentParser) parseActionInputs(args map[string]string, size types.Size) (map[string]any, error) {
	actionInputs := make(map[string]any)
	imageWidth := size.Width
	imageHeight := size.Height

	for paramName, param := range args {
		if param == "" {
			continue
		}
		param = strings.TrimSpace(param)

		// Convert box coordinates
		if strings.Contains(paramName, "box") || strings.Contains(paramName, "point") {
			// Extract numbers from the parameter value using regex
			re := regexp.MustCompile(`\d+`)
			numbers := re.FindAllString(param, -1)
			if len(numbers) >= 2 {
				coords := make([]float64, len(numbers))
				for i, numStr := range numbers {
					num, err := strconv.ParseFloat(numStr, 64)
					if err != nil {
						return nil, fmt.Errorf("invalid coordinate: %s", numStr)
					}
					// Convert relative coordinates to absolute coordinates
					if i%2 == 0 { // x coordinates
						coords[i] = math.Round((num/DefaultFactor*float64(imageWidth))*10) / 10
					} else { // y coordinates
						coords[i] = math.Round((num/DefaultFactor*float64(imageHeight))*10) / 10
					}
				}
				actionInputs[paramName] = coords
			} else {
				actionInputs[paramName] = param
			}
		} else {
			// Handle other parameter types (content, key, direction, etc.)
			if paramName == "content" {
				// Handle escape characters
				param = strings.ReplaceAll(param, "\\n", "\n")
				param = strings.ReplaceAll(param, "\\\"", "\"")
				param = strings.ReplaceAll(param, "\\'", "'")
			}
			actionInputs[paramName] = param
		}
	}

	return actionInputs, nil
}

// convertActionsToToolCalls converts actions to tool calls
func (p *UITARSContentParser) convertActionsToToolCalls(actions []Action) []schema.ToolCall {
	toolCalls := make([]schema.ToolCall, 0, len(actions))
	for _, action := range actions {
		jsonArgs, err := json.Marshal(action.ActionInputs)
		if err != nil {
			log.Error().Interface("action", action).Msg("failed to marshal action inputs")
			continue
		}
		toolCalls = append(toolCalls, schema.ToolCall{
			ID:   action.ActionType + "_" + strconv.FormatInt(time.Now().Unix(), 10),
			Type: "function",
			Function: schema.FunctionCall{
				Name:      action.ActionType,
				Arguments: string(jsonArgs),
			},
		})
	}
	return toolCalls
}

// Action represents a parsed action with its context.
type Action struct {
	ActionType   string         `json:"action_type"`
	ActionInputs map[string]any `json:"action_inputs"`
}

// ParseAction parses an action string into function name and arguments.
func ParseAction(actionStr string) (*ParsedAction, error) {
	// Parse action type and parameters
	actionParts := strings.SplitN(actionStr, "(", 2)
	if len(actionParts) < 2 {
		return nil, fmt.Errorf("not a function call")
	}

	funcName := strings.TrimSpace(actionParts[0])
	paramsText := strings.TrimSuffix(strings.TrimSpace(actionParts[1]), ")")

	args := make(map[string]string)
	if paramsText != "" {
		// Split parameters by comma and parse key=value pairs
		for _, param := range strings.Split(paramsText, ",") {
			param = strings.TrimSpace(param)
			if strings.Contains(param, "=") {
				parts := strings.SplitN(param, "=", 2)
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				// Remove surrounding quotes
				value = strings.Trim(value, "'\"")
				args[key] = value
			}
		}
	}

	return &ParsedAction{Function: funcName, Args: args}, nil
}

// ParsedAction represents the result of parsing an action string.
type ParsedAction struct {
	Function string
	Args     map[string]string
}
