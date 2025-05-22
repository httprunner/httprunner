package ai

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

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
	if strings.Contains(text, "<point>") {
		text = convertPointToCoordinates(text)
	}
	text = strings.ReplaceAll(text, "start_point=", "start_box=")
	text = strings.ReplaceAll(text, "end_point=", "end_box=")
	text = strings.ReplaceAll(text, "point=", "start_box=")

	// Extract context (thought/reflection)
	var thought, reflection string
	actionIdx := strings.Index(text, "Action:")
	prefix := ""
	if actionIdx != -1 {
		prefix = text[:actionIdx]
	}
	if strings.HasPrefix(prefix, "Thought:") {
		thought = strings.TrimSpace(strings.TrimPrefix(prefix, "Thought:"))
	} else if strings.HasPrefix(prefix, "Reflection:") {
		refIdx := strings.Index(prefix, "Action_Summary:")
		if refIdx != -1 {
			reflection = strings.TrimSpace(strings.TrimPrefix(prefix[:refIdx], "Reflection:"))
			thought = strings.TrimSpace(strings.TrimPrefix(prefix[refIdx:], "Action_Summary:"))
		}
	} else if strings.HasPrefix(prefix, "Action_Summary:") {
		thought = strings.TrimSpace(strings.TrimPrefix(prefix, "Action_Summary:"))
	}
	if !strings.Contains(text, "Action:") {
		return nil, fmt.Errorf("no Action: found")
	}
	actionStr := strings.SplitN(text, "Action: ", 2)[1]

	rawActions := strings.Split(actionStr, ")\n\n")
	normalizedActions := make([]string, 0, len(rawActions))
	for _, act := range rawActions {
		actionStr := act
		if strings.Contains(actionStr, "type(content") {
			if !strings.HasSuffix(strings.TrimSpace(actionStr), ")") {
				actionStr = strings.TrimSpace(actionStr) + ")"
			}
			pattern := regexp.MustCompile(`type\(content='(.*?)'\)`)
			m := pattern.FindStringSubmatch(actionStr)
			if len(m) > 1 {
				content := m[1]
				actionStr = "type(content='" + escapeSingleQuotes(content) + "')"
			} else {
				return nil, fmt.Errorf("pattern not found in the input string")
			}
		}
		if !strings.HasSuffix(strings.TrimSpace(actionStr), ")") {
			actionStr = strings.TrimSpace(actionStr) + ")"
		}
		normalizedActions = append(normalizedActions, actionStr)
	}

	actions := make([]Action, 0, len(normalizedActions))
	for _, action := range normalizedActions {
		parsed, err := ParseAction(strings.ReplaceAll(action, "\n", "\\n"))
		if err != nil {
			return nil, fmt.Errorf("Action can't parse: %s", action)
		}
		actionType := parsed.Function
		params := parsed.Args
		actionInputs := make(map[string]any)
		imageWidth := size.Width
		imageHeight := size.Height
		for paramName, param := range params {
			if param == "" {
				continue
			}
			param = strings.TrimLeft(param, " ")
			actionInputs[paramName] = param
			if strings.Contains(paramName, "start_box") || strings.Contains(paramName, "end_box") {
				oriBox := param
				parameters := strings.Split(strings.ReplaceAll(strings.ReplaceAll(oriBox, "(", ""), ")", ""), ",")
				floatNumbers := make([]float64, 0, len(parameters))
				for _, numStr := range parameters {
					num, err := strconv.ParseFloat(strings.TrimSpace(numStr), 64)
					if err != nil {
						log.Error().Interface("parameters", parameters).Msg("invalid float action parameters")
						return nil, fmt.Errorf("invalid action parameters")
					}
					floatNumbers = append(floatNumbers, num)
				}
				// The model generates a 2D coordinate output that represents relative positions.
				// To convert these values to image-relative coordinates, divide each component by 1000 to obtain values in the range [0,1].
				// The absolute coordinates required by the Action can be calculated by:
				// - X absolute = X relative × image width / 1000
				// - Y absolute = Y relative × image height / 1000
				if len(floatNumbers) == 2 {
					floatNumbers[0] = math.Round((floatNumbers[0]/DefaultFactor*float64(imageWidth))*10) / 10
					floatNumbers[1] = math.Round((floatNumbers[1]/DefaultFactor*float64(imageHeight))*10) / 10
				} else if len(floatNumbers) == 4 {
					floatNumbers[0] = math.Round((floatNumbers[0]/DefaultFactor*float64(imageWidth))*10) / 10
					floatNumbers[1] = math.Round((floatNumbers[1]/DefaultFactor*float64(imageHeight))*10) / 10
					floatNumbers[2] = math.Round((floatNumbers[2]/DefaultFactor*float64(imageWidth))*10) / 10
					floatNumbers[3] = math.Round((floatNumbers[3]/DefaultFactor*float64(imageHeight))*10) / 10
				} else {
					log.Error().Interface("parameters", floatNumbers).Msg("invalid float action parameters")
					return nil, fmt.Errorf("invalid action parameters")
				}
				actionInputs[paramName] = floatNumbers
			}
		}
		actions = append(actions, Action{
			Reflection:   reflection,
			Thought:      thought,
			ActionType:   actionType,
			ActionInputs: actionInputs,
			Text:         text,
		})
	}
	return &PlanningResult{
		Actions: actions,
	}, nil
}

// Action represents a parsed action with its context.
type Action struct {
	Reflection   string         `json:"reflection"`
	Thought      string         `json:"thought"`
	ActionType   string         `json:"action_type"`
	ActionInputs map[string]any `json:"action_inputs"`
	Text         string         `json:"text"`
}

// ParsedActionArgs represents the result of parsing an action string.
type ParsedActionArgs struct {
	Function string
	Args     map[string]string
}

// convertPointToCoordinates replaces <point>x y</point> with (x,y)
func convertPointToCoordinates(text string) string {
	// 支持 <point>x1 y1 x2 y2</point> 或 <point>x y</point>
	re := regexp.MustCompile(`<point>(\d+)\s+(\d+)(?:\s+(\d+)\s+(\d+))?</point>`)
	return re.ReplaceAllStringFunc(text, func(match string) string {
		submatches := re.FindStringSubmatch(match)
		if submatches[3] != "" && submatches[4] != "" {
			// 4 个数字
			return fmt.Sprintf("(%s,%s,%s,%s)", submatches[1], submatches[2], submatches[3], submatches[4])
		}
		// 2 个数字
		return fmt.Sprintf("(%s,%s)", submatches[1], submatches[2])
	})
}

// escapeSingleQuotes escapes unescaped single quotes in a string.
func escapeSingleQuotes(text string) string {
	var b strings.Builder
	n := len(text)
	for i := 0; i < n; i++ {
		if text[i] == '\'' && (i == 0 || text[i-1] != '\\') {
			b.WriteString("\\'")
		} else {
			b.WriteByte(text[i])
		}
	}
	return b.String()
}

// ParseAction parses an action string into function name and arguments.
func ParseAction(actionStr string) (*ParsedActionArgs, error) {
	re := regexp.MustCompile(`^(\w+)\((.*)\)$`)
	matches := re.FindStringSubmatch(actionStr)
	if len(matches) < 3 {
		return nil, fmt.Errorf("not a function call")
	}
	funcName := matches[1]
	argsStr := matches[2]
	args := make(map[string]string)
	argRe := regexp.MustCompile(`(\w+)\s*=\s*'([^']*)'`)
	for _, m := range argRe.FindAllStringSubmatch(argsStr, -1) {
		args[m[1]] = m[2]
	}
	return &ParsedActionArgs{Function: funcName, Args: args}, nil
}
