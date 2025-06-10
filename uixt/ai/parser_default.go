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
	switch modelType {
	case option.DOUBAO_1_5_UI_TARS_250428:
		return &UITARSContentParser{
			modelType:     modelType,
			systemPrompt:  doubao_1_5_ui_tars_planning_prompt,
			actionMapping: doubao_1_5_ui_tars_action_mapping,
		}
	default:
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

// extractJSONContent extracts JSON content from various formats in the response
func (p *JSONContentParser) extractJSONContent(content string) string {
	content = strings.TrimSpace(content)

	// Case 1: Content wrapped in ```json ... ```
	if strings.Contains(content, "```json") {
		start := strings.Index(content, "```json")
		if start != -1 {
			start += 7 // length of "```json"
			end := strings.Index(content[start:], "```")
			if end != -1 {
				jsonContent := strings.TrimSpace(content[start : start+end])
				return jsonContent
			}
		}
	}

	// Case 2: Content wrapped in ``` ... ``` (without json specifier)
	if strings.HasPrefix(content, "```") && strings.HasSuffix(content, "```") {
		lines := strings.Split(content, "\n")
		if len(lines) >= 3 {
			// Remove first and last lines (the ``` markers)
			jsonLines := lines[1 : len(lines)-1]
			jsonContent := strings.Join(jsonLines, "\n")
			jsonContent = strings.TrimSpace(jsonContent)
			// Check if it looks like JSON
			if strings.HasPrefix(jsonContent, "{") && strings.HasSuffix(jsonContent, "}") {
				return jsonContent
			}
		}
	}

	// Case 3: Look for JSON object in the content
	start := strings.Index(content, "{")
	if start != -1 {
		// Find the matching closing brace
		braceCount := 0
		for i := start; i < len(content); i++ {
			if content[i] == '{' {
				braceCount++
			} else if content[i] == '}' {
				braceCount--
				if braceCount == 0 {
					jsonContent := strings.TrimSpace(content[start : i+1])
					return jsonContent
				}
			}
		}
	}

	// Case 4: If content itself looks like JSON
	if strings.HasPrefix(content, "{") && strings.HasSuffix(content, "}") {
		return content
	}

	return ""
}

func (p *JSONContentParser) Parse(content string, size types.Size) (*PlanningResult, error) {
	content = strings.TrimSpace(content)

	// Extract JSON content from markdown code blocks
	jsonContent := p.extractJSONContent(content)
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
