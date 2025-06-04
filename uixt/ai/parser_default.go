package ai

import (
	"fmt"
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
			systemPrompt:  doubao_1_5_ui_tars_planning_prompt,
			actionMapping: doubao_1_5_ui_tars_action_mapping,
		}
	default:
		return &JSONContentParser{
			systemPrompt:  defaultPlanningResponseJsonFormat,
			actionMapping: map[string]option.ActionName{},
		}
	}
}

// JSONContentParser parses the response as JSON string format
type JSONContentParser struct {
	systemPrompt  string
	actionMapping map[string]option.ActionName
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

	// Define a temporary struct to parse the expected JSON format
	var jsonResponse struct {
		Actions []Action `json:"actions"`
		Summary string   `json:"summary"`
		Error   string   `json:"error"`
	}

	if err := json.Unmarshal([]byte(content), &jsonResponse); err != nil {
		return nil, fmt.Errorf("failed to parse VLM response: %v", err)
	}

	if jsonResponse.Error != "" {
		return nil, errors.New(jsonResponse.Error)
	}

	if len(jsonResponse.Actions) == 0 {
		return nil, errors.New("no actions returned from VLM")
	}

	// normalize actions using unified function from ui-tars parser
	var normalizedActions []Action
	for i := range jsonResponse.Actions {
		// create a new variable, avoid implicit memory aliasing in for loop.
		action := jsonResponse.Actions[i]

		// Process and normalize arguments (from JSON parser)
		processedArgs, err := processActionArguments(action.ActionInputs, size)
		if err != nil {
			return nil, errors.Wrap(err, "failed to process action arguments")
		}
		action.ActionInputs = processedArgs

		normalizedActions = append(normalizedActions, action)
	}

	// Convert actions to tool calls using function from parser_ui_tars.go
	toolCalls := convertActionsToToolCalls(normalizedActions, p.actionMapping)

	return &PlanningResult{
		ToolCalls:     toolCalls,
		ActionSummary: jsonResponse.Summary,
		Thought:       jsonResponse.Summary,
		Content:       content,
	}, nil
}
