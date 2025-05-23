package ai

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/httprunner/httprunner/v5/internal/json"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/httprunner/httprunner/v5/uixt/types"
	"github.com/pkg/errors"
)

// LLMContentParser parses the content from the LLM response
// parser is corresponding to the model type and system prompt
type LLMContentParser interface {
	SystemPrompt() string
	Parse(content string, size types.Size) (*PlanningResult, error)
}

func NewLLMContentParser(modelType option.LLMServiceType) LLMContentParser {
	switch modelType {
	case option.LLMServiceTypeUITARS:
		return &UITARSContentParser{
			systemPrompt: doubao_1_5_ui_tars_planning_prompt,
		}
	default:
		return &JSONContentParser{
			systemPrompt: defaultPlanningResponseJsonFormat,
		}
	}
}

// JSONContentParser parses the response as JSON string format
type JSONContentParser struct {
	systemPrompt string
}

func (p *JSONContentParser) SystemPrompt() string {
	return p.systemPrompt
}

func (p *JSONContentParser) Parse(content string, size types.Size) (*PlanningResult, error) {
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "```json") && strings.HasSuffix(content, "```") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimSuffix(content, "```")
	}
	content = strings.TrimSpace(content)

	var response PlanningResult
	if err := json.Unmarshal([]byte(content), &response); err != nil {
		return nil, fmt.Errorf("failed to parse VLM response: %v", err)
	}

	if response.Error != "" {
		return nil, errors.New(response.Error)
	}

	if len(response.Actions) == 0 {
		return nil, errors.New("no actions returned from VLM")
	}

	// normalize actions
	var normalizedActions []Action
	for i := range response.Actions {
		// create a new variable, avoid implicit memory aliasing in for loop.
		action := response.Actions[i]
		if err := normalizeAction(&action); err != nil {
			return nil, errors.Wrap(err, "failed to normalize action")
		}
		normalizedActions = append(normalizedActions, action)
	}

	return &PlanningResult{
		Actions:       normalizedActions,
		ActionSummary: response.ActionSummary,
	}, nil
}

// normalizeAction normalizes the coordinates in the action
func normalizeAction(action *Action) error {
	switch action.ActionType {
	case "click", "drag":
		// handle click and drag action coordinates
		if startBox, ok := action.ActionInputs["startBox"].(string); ok {
			normalized, err := normalizeCoordinates(startBox)
			if err != nil {
				return fmt.Errorf("failed to normalize startBox: %w", err)
			}
			action.ActionInputs["startBox"] = normalized
		}

		if endBox, ok := action.ActionInputs["endBox"].(string); ok {
			normalized, err := normalizeCoordinates(endBox)
			if err != nil {
				return fmt.Errorf("failed to normalize endBox: %w", err)
			}
			action.ActionInputs["endBox"] = normalized
		}
	}

	return nil
}

// normalizeCoordinates normalizes the coordinates based on the factor
func normalizeCoordinates(coordStr string) (coords []float64, err error) {
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
