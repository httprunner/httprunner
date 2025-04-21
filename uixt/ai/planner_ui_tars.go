package ai

import (
	"context"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/config"
	"github.com/httprunner/httprunner/v5/internal/json"
	"github.com/httprunner/httprunner/v5/uixt/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

const (
	EnvArkBaseURL = "ARK_BASE_URL"
	EnvArkAPIKey  = "ARK_API_KEY"
	EnvArkModelID = "ARK_MODEL_ID"
)

func GetArkModelConfig() (*ark.ChatModelConfig, error) {
	if err := config.LoadEnv(); err != nil {
		return nil, errors.Wrap(code.LoadEnvError, err.Error())
	}

	arkBaseURL := os.Getenv(EnvArkBaseURL)
	arkAPIKey := os.Getenv(EnvArkAPIKey)
	if arkAPIKey == "" {
		return nil, errors.Wrapf(code.LLMEnvMissedError,
			"env %s missed", EnvArkAPIKey)
	}
	modelName := os.Getenv(EnvArkModelID)
	if modelName == "" {
		return nil, errors.Wrapf(code.LLMEnvMissedError,
			"env %s missed", EnvArkModelID)
	}
	timeout := defaultTimeout
	modelConfig := &ark.ChatModelConfig{
		BaseURL: arkBaseURL,
		APIKey:  arkAPIKey,
		Model:   modelName,
		Timeout: &timeout,
	}

	// log config info
	log.Info().Str("model", modelConfig.Model).
		Str("baseURL", modelConfig.BaseURL).
		Str("apiKey", maskAPIKey(modelConfig.APIKey)).
		Str("timeout", defaultTimeout.String()).
		Msg("get model config")

	return modelConfig, nil
}

func NewUITarsPlanner(ctx context.Context) (*UITarsPlanner, error) {
	config, err := GetArkModelConfig()
	if err != nil {
		return nil, err
	}
	chatModel, err := ark.NewChatModel(ctx, config)
	if err != nil {
		return nil, err
	}

	return &UITarsPlanner{
		ctx:          ctx,
		config:       config,
		model:        chatModel,
		systemPrompt: uiTarsPlanningPrompt,
	}, nil
}

// https://www.volcengine.com/docs/82379/1536429
const uiTarsPlanningPrompt = `
You are a GUI agent. You are given a task and your action history, with screenshots. You need to perform the next action to complete the task.

## Output Format
` + "```" + `
Thought: ...
Action: ...
` + "```" + `

## Action Space
click(start_box='[x1, y1, x2, y2]')
left_double(start_box='[x1, y1, x2, y2]')
right_single(start_box='[x1, y1, x2, y2]')
drag(start_box='[x1, y1, x2, y2]', end_box='[x3, y3, x4, y4]')
hotkey(key='')
type(content='') #If you want to submit your input, use "\n" at the end of ` + "`content`" + `.
scroll(start_box='[x1, y1, x2, y2]', direction='down or up or right or left')
wait() #Sleep for 5s and take a screenshot to check for any changes.
finished(content='xxx') # Use escape characters \\', \\", and \\n in content part to ensure we can parse the content in normal python string format.

## Note
- Use Chinese in ` + "`Thought`" + ` part.
- Write a small plan and finally summarize your next action (with its target element) in one sentence in ` + "`Thought`" + ` part.

## User Instruction
`

type UITarsPlanner struct {
	ctx          context.Context
	model        model.ToolCallingChatModel
	config       *ark.ChatModelConfig
	systemPrompt string
	history      []*schema.Message // conversation history
}

// Call performs UI planning using Vision Language Model
func (p *UITarsPlanner) Call(opts *PlanningOptions) (*PlanningResult, error) {
	// validate input parameters
	if err := validateInput(opts); err != nil {
		return nil, errors.Wrap(err, "validate input parameters failed")
	}

	// prepare prompt
	if len(p.history) == 0 {
		// add system message
		systemPrompt := uiTarsPlanningPrompt + opts.UserInstruction
		p.history = []*schema.Message{
			{
				Role:    schema.System,
				Content: systemPrompt,
			},
		}
	}
	// append user image message
	appendConversationHistory(p.history, opts.Message)

	// call model service, generate response
	logRequest(p.history)
	startTime := time.Now()
	resp, err := p.model.Generate(p.ctx, p.history)
	log.Info().Float64("elapsed(s)", time.Since(startTime).Seconds()).
		Str("model", p.config.Model).Msg("call model service")
	if err != nil {
		return nil, fmt.Errorf("request model service failed: %w", err)
	}
	logResponse(resp)

	// parse result
	result, err := p.parseResult(resp, opts.Size)
	if err != nil {
		return nil, errors.Wrap(err, "parse result failed")
	}

	// append assistant message
	appendConversationHistory(p.history, &schema.Message{
		Role:    schema.Assistant,
		Content: result.ActionSummary,
	})

	return result, nil
}

func (p *UITarsPlanner) parseResult(msg *schema.Message, size types.Size) (*PlanningResult, error) {
	// parse Thought/Action format from UI-TARS
	parseActions, thoughtErr := parseThoughtAction(msg.Content)
	if thoughtErr != nil {
		return nil, thoughtErr
	}

	// process response
	result, err := processVLMResponse(parseActions, size)
	if err != nil {
		return nil, errors.Wrap(err, "process VLM response failed")
	}

	log.Info().
		Interface("summary", result.ActionSummary).
		Interface("actions", result.NextActions).
		Msg("get VLM planning result")
	return result, nil
}

// parseThoughtAction parses the Thought/Action format response
func parseThoughtAction(predictionText string) ([]ParsedAction, error) {
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
	return parseActionText(actionsText, thought)
}

// parseActionText parses the action text to extract the action type and parameters
func parseActionText(actionsText, thought string) ([]ParsedAction, error) {
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
					coord, err := normalizeCoordinates(matches[1])
					if err != nil {
						return nil, errors.Wrapf(err, "normalize point failed: %s", matches[1])
					}
					action.ActionInputs["startBox"] = coord
				}
			case ActionTypeDrag:
				if len(matches) > 2 {
					// handle start point
					startBox, err := normalizeCoordinates(matches[1])
					if err != nil {
						return nil, errors.Wrapf(err, "normalize startBox failed: %s", matches[1])
					}
					action.ActionInputs["startBox"] = startBox

					// handle end point
					endBox, err := normalizeCoordinates(matches[2])
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
					startBox, err := normalizeCoordinates(matches[1])
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

// processVLMResponse processes the VLM response and converts it to PlanningResult
func processVLMResponse(actions []ParsedAction, size types.Size) (*PlanningResult, error) {
	log.Info().Msg("processing VLM response...")

	if len(actions) == 0 {
		return nil, fmt.Errorf("no actions returned from VLM")
	}

	// validate and post-process each action
	for i := range actions {
		// validate action type
		switch actions[i].ActionType {
		case "click":
			if err := convertCoordinateAction(&actions[i], "startBox", size); err != nil {
				return nil, errors.Wrap(err, "convert coordinate action failed")
			}
		case "drag":
			if err := convertCoordinateAction(&actions[i], "startBox", size); err != nil {
				return nil, errors.Wrap(err, "convert coordinate action failed")
			}
			if err := convertCoordinateAction(&actions[i], "endBox", size); err != nil {
				return nil, errors.Wrap(err, "convert coordinate action failed")
			}
		case "type":
			validateTypeContent(&actions[i])
		case "wait", "finished", "call_user":
			// these actions do not need extra parameters
		default:
			log.Printf("warning: unknown action type: %s, will try to continue processing", actions[i].ActionType)
		}
	}

	// extract action summary
	actionSummary := extractActionSummary(actions)

	return &PlanningResult{
		NextActions:   actions,
		ActionSummary: actionSummary,
	}, nil
}

// extractActionSummary extracts the summary from the actions
func extractActionSummary(actions []ParsedAction) string {
	if len(actions) == 0 {
		return ""
	}

	// use the Thought of the first action as summary
	if actions[0].Thought != "" {
		return actions[0].Thought
	}

	// if no Thought, generate summary from action type
	action := actions[0]
	switch action.ActionType {
	case "click":
		return "点击操作"
	case "drag":
		return "拖拽操作"
	case "type":
		content, _ := action.ActionInputs["content"].(string)
		if len(content) > 20 {
			content = content[:20] + "..."
		}
		return fmt.Sprintf("输入文本: %s", content)
	case "wait":
		return "等待操作"
	case "finished":
		return "完成操作"
	case "call_user":
		return "请求用户协助"
	default:
		return fmt.Sprintf("执行 %s 操作", action.ActionType)
	}
}

func convertCoordinateAction(action *ParsedAction, boxField string, size types.Size) error {
	// The model generates a 2D coordinate output that represents relative positions.
	// To convert these values to image-relative coordinates, divide each component by 1000 to obtain values in the range [0,1].
	// The absolute coordinates required by the Action can be calculated by:
	// - X absolute = X relative × image width / 1000
	// - Y absolute = Y relative × image height / 1000

	// get image width and height
	imageWidth := size.Width
	imageHeight := size.Height

	box := action.ActionInputs[boxField]
	coords, ok := box.([]float64)
	if !ok {
		log.Error().Interface("inputs", action.ActionInputs).Msg("invalid action inputs")
		return fmt.Errorf("invalid action inputs")
	}

	if len(coords) == 2 {
		coords[0] = math.Round((coords[0]/1000*float64(imageWidth))*10) / 10
		coords[1] = math.Round((coords[1]/1000*float64(imageHeight))*10) / 10
	} else if len(coords) == 4 {
		coords[0] = math.Round((coords[0]/1000*float64(imageWidth))*10) / 10
		coords[1] = math.Round((coords[1]/1000*float64(imageHeight))*10) / 10
		coords[2] = math.Round((coords[2]/1000*float64(imageWidth))*10) / 10
		coords[3] = math.Round((coords[3]/1000*float64(imageHeight))*10) / 10
	} else {
		log.Error().Interface("inputs", action.ActionInputs).Msg("invalid action inputs")
		return fmt.Errorf("invalid action inputs")
	}

	return nil
}

// validateTypeContent 验证输入文本内容
func validateTypeContent(action *ParsedAction) {
	if content, ok := action.ActionInputs["content"]; !ok || content == "" {
		// default to empty string
		action.ActionInputs["content"] = ""
		log.Warn().Msg("type action missing content parameter, set to default")
	}
}
