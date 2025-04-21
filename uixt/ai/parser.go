package ai

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// NewActionParser creates a new ActionParser instance
func NewActionParser(factor float64) *ActionParser {
	return &ActionParser{
		Factor: factor,
	}
}

// ActionParser parses VLM responses and converts them to structured actions
type ActionParser struct {
	Factor float64 // TODO
}

// Parse parses the prediction text and extracts actions
func (p *ActionParser) Parse(predictionText string) ([]ParsedAction, error) {
	// try parsing JSON format, from VLM like openai/gpt-4o
	jsonActions, jsonErr := p.parseJSON(predictionText)
	if jsonErr == nil {
		return jsonActions, nil
	}

	// json parsing failed, try parsing Thought/Action format, from VLM like UI-TARS
	thoughtActions, thoughtErr := p.parseThoughtAction(predictionText)
	if thoughtErr == nil {
		return thoughtActions, nil
	}

	return nil, errors.Wrap(thoughtErr, "parse planner response failed")
}

// parseJSON tries to parse the response as JSON format
func (p *ActionParser) parseJSON(predictionText string) ([]ParsedAction, error) {
	predictionText = strings.TrimSpace(predictionText)
	if strings.HasPrefix(predictionText, "```json") && strings.HasSuffix(predictionText, "```") {
		predictionText = strings.TrimPrefix(predictionText, "```json")
		predictionText = strings.TrimSuffix(predictionText, "```")
	}
	predictionText = strings.TrimSpace(predictionText)

	var response VLMResponse
	if err := json.Unmarshal([]byte(predictionText), &response); err != nil {
		return nil, fmt.Errorf("failed to parse VLM response: %v", err)
	}

	if response.Error != "" {
		return nil, errors.New(response.Error)
	}

	if len(response.Actions) == 0 {
		return nil, errors.New("no actions returned from VLM")
	}

	// normalize actions
	var normalizedActions []ParsedAction
	for i := range response.Actions {
		// create a new variable, avoid implicit memory aliasing in for loop.
		action := response.Actions[i]
		if err := p.normalizeAction(&action); err != nil {
			return nil, errors.Wrap(err, "failed to normalize action")
		}
		normalizedActions = append(normalizedActions, action)
	}

	return normalizedActions, nil
}

// parseThoughtAction parses the Thought/Action format response
func (p *ActionParser) parseThoughtAction(predictionText string) ([]ParsedAction, error) {
	thoughtRegex := regexp.MustCompile(`(?is)Thought:(.+?)Action:`)
	actionRegex := regexp.MustCompile(`(?is)Action:(.+)`)

	// extract Thought part
	thoughtMatch := thoughtRegex.FindStringSubmatch(predictionText)
	var thought string
	if len(thoughtMatch) > 1 {
		thought = strings.TrimSpace(thoughtMatch[1])
	}

	// extract Action part, e.g. "click(start_box='(552,454)')"
	actionMatch := actionRegex.FindStringSubmatch(predictionText)
	if len(actionMatch) < 2 {
		return nil, errors.New("no action found in the response")
	}

	actionsText := strings.TrimSpace(actionMatch[1])

	// parse action type and parameters
	return p.parseActionText(actionsText, thought)
}

// parseActionText parses the action text to extract the action type and parameters
func (p *ActionParser) parseActionText(actionsText, thought string) ([]ParsedAction, error) {
	// remove trailing comments
	if idx := strings.Index(actionsText, "#"); idx > 0 {
		actionsText = strings.TrimSpace(actionsText[:idx])
	}

	// supported action types and regexes
	actionRegexes := map[ActionType]*regexp.Regexp{
		"click":        regexp.MustCompile(`click\(start_box='([^']+)'\)`),
		"left_double":  regexp.MustCompile(`left_double\(start_box='([^']+)'\)`),
		"right_single": regexp.MustCompile(`right_single\(start_box='([^']+)'\)`),
		"drag":         regexp.MustCompile(`drag\(start_box='([^']+)', end_box='([^']+)'\)`),
		"type":         regexp.MustCompile(`type\(content='([^']+)'\)`),
		"scroll":       regexp.MustCompile(`scroll\(start_box='([^']+)', direction='([^']+)'\)`),
		"wait":         regexp.MustCompile(`wait\(\)`),
		"finished":     regexp.MustCompile(`finished\(content='([^']+)'\)`),
		"call_user":    regexp.MustCompile(`call_user\(\)`),
	}

	// one or multiple actions, separated by newline
	// "click(start_box='<bbox>229 379 229 379</bbox>')
	// "click(start_box='<bbox>229 379 229 379</bbox>')\n\nclick(start_box='<bbox>769 519 769 519</bbox>')"
	parsedActions := make([]ParsedAction, 0)
	for _, actionText := range strings.Split(actionsText, "\n") {
		actionText = strings.TrimSpace(actionText)
		for actionType, regex := range actionRegexes {
			matches := regex.FindStringSubmatch(actionText)
			if len(matches) == 0 {
				continue
			}

			var action ParsedAction
			action.ActionType = actionType
			action.ActionInputs = make(map[string]interface{})
			action.Thought = thought

			// parse parameters based on action type
			switch actionType {
			case ActionTypeClick:
				if len(matches) > 1 {
					coord, err := p.normalizeCoordinates(matches[1])
					if err != nil {
						return nil, errors.Wrapf(err, "normalize point failed: %s", matches[1])
					}
					action.ActionInputs["startBox"] = coord
				}
			case ActionTypeDrag:
				if len(matches) > 2 {
					// handle start point
					startBox, err := p.normalizeCoordinates(matches[1])
					if err != nil {
						return nil, errors.Wrapf(err, "normalize startBox failed: %s", matches[1])
					}
					action.ActionInputs["startBox"] = startBox

					// handle end point
					endBox, err := p.normalizeCoordinates(matches[2])
					if err != nil {
						return nil, errors.Wrapf(err, "normalize endBox failed: %s", matches[2])
					}
					action.ActionInputs["endBox"] = endBox
				}
			case ActionTypeType:
				if len(matches) > 1 {
					action.ActionInputs["content"] = matches[1]
				}
			case ActionTypeScroll:
				if len(matches) > 2 {
					startBox, err := p.normalizeCoordinates(matches[1])
					if err != nil {
						return nil, errors.Wrapf(err, "normalize startBox failed: %s", matches[1])
					}
					action.ActionInputs["startBox"] = startBox
					action.ActionInputs["direction"] = matches[2]
				}
			case ActionTypeWait, ActionTypeFinished, ActionTypeCallUser:
				// 这些动作没有额外参数
			}

			parsedActions = append(parsedActions, action)
		}
	}

	if len(parsedActions) == 0 {
		return nil, fmt.Errorf("no valid actions returned from VLM")
	}
	return parsedActions, nil
}

// normalizeAction normalizes the coordinates in the action
func (p *ActionParser) normalizeAction(action *ParsedAction) error {
	switch action.ActionType {
	case "click", "drag":
		// handle click and drag action coordinates
		if startBox, ok := action.ActionInputs["startBox"].(string); ok {
			normalized, err := p.normalizeCoordinates(startBox)
			if err != nil {
				return fmt.Errorf("failed to normalize startBox: %w", err)
			}
			action.ActionInputs["startBox"] = normalized
		}

		if endBox, ok := action.ActionInputs["endBox"].(string); ok {
			normalized, err := p.normalizeCoordinates(endBox)
			if err != nil {
				return fmt.Errorf("failed to normalize endBox: %w", err)
			}
			action.ActionInputs["endBox"] = normalized
		}
	}

	return nil
}

// normalizeCoordinates normalizes the coordinates based on the factor
func (p *ActionParser) normalizeCoordinates(coordStr string) (coords []float64, err error) {
	// check empty string
	if coordStr == "" {
		return nil, fmt.Errorf("empty coordinate string")
	}

	// handle BBox format: <bbox>x1 y1 x2 y2</bbox>
	bboxRegex := regexp.MustCompile(`<bbox>(\d+\s+\d+\s+\d+\s+\d+)</bbox>`)
	bboxMatches := bboxRegex.FindStringSubmatch(coordStr)
	if len(bboxMatches) > 1 {
		// Extract space-separated values from inside the bbox tags
		bboxContent := bboxMatches[1]
		// Split by whitespace
		parts := strings.Fields(bboxContent)
		if len(parts) == 4 {
			coords = make([]float64, 4)
			for i, part := range parts {
				val, e := strconv.ParseFloat(part, 64)
				if e != nil {
					return nil, fmt.Errorf("failed to parse coordinate value '%s': %w", part, e)
				}
				coords[i] = val
			}
			// 将 val 转换为 [x,y] 坐标
			x := (coords[0] + coords[2]) / 2
			y := (coords[1] + coords[3]) / 2
			return []float64{x, y}, nil
		}
	}

	// handle coordinate string, e.g. "[100, 200]", "(100, 200)"
	if strings.Contains(coordStr, ",") {
		// remove possible brackets and split coordinates
		coordStr = strings.Trim(coordStr, "[]() \t")

		// try parsing JSON array
		jsonStr := coordStr
		if !strings.HasPrefix(jsonStr, "[") {
			jsonStr = "[" + coordStr + "]"
		}

		err = json.Unmarshal([]byte(jsonStr), &coords)
		if err != nil {
			return nil, fmt.Errorf("failed to parse coordinate string: %w", err)
		}
		return coords, nil
	}

	return nil, fmt.Errorf("invalid coordinate string format: %s", coordStr)
}
