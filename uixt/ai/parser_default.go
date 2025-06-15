package ai

import (
	"fmt"
	"strings"

	"github.com/cloudwego/eino/schema"
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
	if option.IS_UI_TARS(modelType) {
		return &UITARSContentParser{
			modelType:     modelType,
			systemPrompt:  doubao_1_5_ui_tars_planning_prompt,
			actionMapping: doubao_1_5_ui_tars_action_mapping,
		}
	} else {
		return &JSONContentParser{
			modelType:     modelType,
			systemPrompt:  doubao_1_5_thinking_vision_pro_planning_prompt,
			actionMapping: doubao_1_5_thinking_vision_pro_action_mapping,
		}
	}
}

// JSONContentParser parses the response as JSON string format
type JSONContentParser struct {
	modelType     option.LLMServiceType
	systemPrompt  string
	actionMapping map[string]option.ActionName
}

func (p *JSONContentParser) SystemPrompt() string {
	return p.systemPrompt
}

func (p *JSONContentParser) Parse(content string, size types.Size) (*PlanningResult, error) {
	content = strings.TrimSpace(content)

	// Extract JSON content from markdown code blocks
	jsonContent := extractJSONFromContent(content)
	if jsonContent == "" {
		return nil, fmt.Errorf("no valid JSON content found in response")
	}

	// Define a temporary struct to parse the expected JSON format
	var jsonResponse struct {
		Actions []Action `json:"actions"`
		Thought string   `json:"thought"`
		Error   string   `json:"error"`
	}

	if err := json.Unmarshal([]byte(jsonContent), &jsonResponse); err != nil {
		return nil, fmt.Errorf("failed to parse VLM response: %v", err)
	}

	if jsonResponse.Error != "" {
		return nil, errors.New(jsonResponse.Error)
	}

	// Handle cases where no actions are returned
	if len(jsonResponse.Actions) == 0 {
		// If there's a valid thought but no actions, this might be an informational response
		// rather than an actionable UI task. Return the result with empty tool calls.
		if jsonResponse.Thought != "" {
			return &PlanningResult{
				ToolCalls: []schema.ToolCall{}, // Empty tool calls for informational responses
				Thought:   jsonResponse.Thought,
				Content:   content, // Include the full response content
				ModelName: string(p.modelType),
			}, nil
		}
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

		// Convert processedArgs based on action type and coordinate parameters
		finalArgs, err := convertProcessedArgs(processedArgs, action.ActionType)
		if err != nil {
			return nil, err
		}

		action.ActionInputs = finalArgs
		normalizedActions = append(normalizedActions, action)
	}

	// Convert actions to tool calls using function from parser_ui_tars.go
	toolCalls := convertActionsToToolCalls(normalizedActions, p.actionMapping)

	return &PlanningResult{
		ToolCalls: toolCalls,
		Thought:   jsonResponse.Thought,
		Content:   content,
		ModelName: string(p.modelType),
	}, nil
}
