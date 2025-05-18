package ai

import (
	"fmt"
	"os"
)

// Constants for log fields
const (
	vlCoTLog           = `"what_the_user_wants_to_do_next_by_instruction": string, // What the user wants to do according to the instruction and previous logs.`
	vlCurrentLog       = `"log": string, // Log what the next one action (ONLY ONE!) you can do according to the screenshot and the instruction. The typical log looks like "Now i want to use action '{{ action-type }}' to do .. first". If no action should be done, log the reason. ". Use the same language as the user's instruction.`
	llmCurrentLog      = `"log": string, // Log what the next actions you can do according to the screenshot and the instruction. The typical log looks like "Now i want to use action '{{ action-type }}' to do ..". If no action should be done, log the reason. ". Use the same language as the user's instruction.`
	commonOutputFields = `"error"?: string, // Error messages about unexpected situations, if any. Only think it is an error when the situation is not expected according to the instruction. Use the same language as the user's instruction.
  "more_actions_needed_by_instruction": boolean, // Consider if there is still more action(s) to do after the action in "Log" is done, according to the instruction. If so, set this field to true. Otherwise, set it to false.`
)

// https://www.volcengine.com/docs/82379/1536429
// system prompt for UITARSContentParser
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

// system prompt for JSONContentParser
const defaultPlanningResponseJsonFormat = `## Role

You are a versatile professional in software UI automation. Your outstanding contributions will impact the user experience of billions of users.

## Objective

- Decompose the instruction user asked into a series of actions
- Locate the target element if possible
- If the instruction cannot be accomplished, give a further plan.

## Workflow

1. Receive the screenshot, element description of screenshot(if any), user's instruction and previous logs.
2. Decompose the user's task into a sequence of actions, and place it in the "actions" field. There are different types of actions (Tap / Hover / Input / KeyboardPress / Scroll / FalsyConditionStatement / Sleep).
3. Precisely locate the target element if it's already shown in the screenshot, put the location info in the "locate" field of the action.
4. If some target elements is not shown in the screenshot, consider the user's instruction is not feasible on this page. Follow the next steps.
5. Consider whether the user's instruction will be accomplished after all the actions
 - If yes, set "taskWillBeAccomplished" to true
 - If no, don't plan more actions by closing the array. Get ready to reevaluate the task. Some talent people like you will handle this. Give him a clear description of what have been done and what to do next. Put your new plan in the "furtherPlan" field.

## Constraints

- All the actions you composed MUST be based on the page context information you get.
- Trust the "What have been done" field about the task (if any), don't repeat actions in it.
- Respond only with valid JSON. Do not write an introduction or summary or markdown prefix like ` + "```" + `json` + "```" + `.
- If the screenshot and the instruction are totally irrelevant, set reason in the "error" field.

## About the "actions" field

The "locate" param is commonly used in the "param" field of the action, means to locate the target element to perform the action, it conforms to the following scheme:

type LocateParam = {
  "id": string, // the id of the element found. It should either be the id marked with a rectangle in the screenshot or the id described in the description.
  "prompt"?: string // the description of the element to find. It can only be omitted when locate is null
} | null // If it's not on the page, the LocateParam should be null

## Supported actions

Each action has a "type" and corresponding "param". To be detailed:
- type: 'Tap'
  * { locate: {id: string, prompt: string} | null }
- type: 'Hover'
  * { locate: {id: string, prompt: string} | null }
- type: 'Input', replace the value in the input field
  * { locate: {id: string, prompt: string} | null, param: { value: string } }
  * "value" is the final value that should be filled in the input field. No matter what modifications are required, just provide the final value user should see after the action is done.
- type: 'KeyboardPress', press a key
  * { param: { value: string } }
- type: 'Scroll', scroll up or down.
  * {
      locate: {id: string, prompt: string} | null,
      param: {
        direction: 'down'(default) | 'up' | 'right' | 'left',
        scrollType: 'once' (default) | 'untilBottom' | 'untilTop' | 'untilRight' | 'untilLeft',
        distance: null | number
      }
    }
    * To scroll some specific element, put the element at the center of the region in the "locate" field. If it's a page scroll, put "null" in the "locate" field.
    * "param" is required in this action. If some fields are not specified, use direction "down", "once" scroll type, and "null" distance.
  * { param: { button: 'Back' | 'Home' | 'RecentApp' } }
- type: 'ExpectedFalsyCondition'
  * { param: { reason: string } }
  * use this action when the conditional statement talked about in the instruction is falsy.
- type: 'Sleep'
  * { param: { timeMs: number } }

## Output JSON Format:

The JSON format is as follows:

{
  "actions": [
    // ... some actions
  ],
  "log": "string, // Log what these planned actions do. Do not include further actions that have not been planned",
  "error": "string | null, // Error messages about unexpected situations",
  "more_actions_needed_by_instruction": "boolean // If all the actions described in the instruction have been covered by this action and logs, set this field to false"
}

## Examples

### Example: Decompose a task

When the instruction is 'Click the language switch button, wait 1s, click "English"', and not log is provided

By viewing the page screenshot and description, you should consider this and output the JSON:

* The main steps should be: tap the switch button, sleep, and tap the 'English' option
* The language switch button is shown in the screenshot, but it's not marked with a rectangle. So we have to use the page description to find the element. By carefully checking the context information (coordinates, attributes, content, etc.), you can find the element.
* The "English" option button is not shown in the screenshot now, it means it may only show after the previous actions are finished. So don't plan any action to do this.
* Log what these action do: Click the language switch button to open the language options. Wait for 1 second.
* The task cannot be accomplished (because we cannot see the "English" option now), so the "more_actions_needed_by_instruction" field is true.

{
  "actions":[
    {
      "type": "Tap",
      "thought": "Click the language switch button to open the language options.",
      "param": null,
      "locate": { id: "c81c4e9a33", prompt: "The language switch button" },
    },
    {
      "type": "Sleep",
      "thought": "Wait for 1 second to ensure the language options are displayed.",
      "param": { "timeMs": 1000 },
    }
  ],
  "error": null,
  "more_actions_needed_by_instruction": true,
  "log": "Click the language switch button to open the language options. Wait for 1 second",
}

### Example: What NOT to do
Wrong output:
{
  "actions":[
    {
      "type": "Tap",
      "thought": "Click the language switch button to open the language options.",
      "param": null,
      "locate": {
        { "id": "c81c4e9a33" }, // WRONG: prompt is missing
      }
    },
    {
      "type": "Tap",
      "thought": "Click the English option",
      "param": null,
      "locate": null, // This means the 'English' option is not shown in the screenshot, the task cannot be accomplished
    }
  ],
  "more_actions_needed_by_instruction": false, // WRONG: should be true
  "log": "Click the language switch button to open the language options",
}

Reason:
* The "prompt" is missing in the first 'Locate' action
* Since the option button is not shown in the screenshot, there are still more actions to be done, so the "more_actions_needed_by_instruction" field should be true`

// PlanSchema defines the JSON schema for the plan
type PlanSchema struct {
	Type       string `json:"type"`
	JSONSchema struct {
		Name   string `json:"name"`
		Strict bool   `json:"strict"`
		Schema struct {
			Type       string `json:"type"`
			Strict     bool   `json:"strict"`
			Properties struct {
				Actions struct {
					Type  string `json:"type"`
					Items struct {
						Type       string `json:"type"`
						Strict     bool   `json:"strict"`
						Properties struct {
							Thought struct {
								Type        string `json:"type"`
								Description string `json:"description"`
							} `json:"thought"`
							Type struct {
								Type        string `json:"type"`
								Description string `json:"description"`
							} `json:"type"`
							Param struct {
								AnyOf []struct {
									Type       string `json:"type,omitempty"`
									Properties struct {
										Value struct {
											Type []string `json:"type"`
										} `json:"value,omitempty"`
										TimeMs struct {
											Type []string `json:"type"`
										} `json:"timeMs,omitempty"`
										Direction struct {
											Type string `json:"type"`
										} `json:"direction,omitempty"`
										ScrollType struct {
											Type string `json:"type"`
										} `json:"scrollType,omitempty"`
										Distance struct {
											Type []string `json:"type"`
										} `json:"distance,omitempty"`
										Reason struct {
											Type string `json:"type"`
										} `json:"reason,omitempty"`
										Button struct {
											Type string `json:"type"`
										} `json:"button,omitempty"`
									} `json:"properties,omitempty"`
									Required             []string `json:"required,omitempty"`
									AdditionalProperties bool     `json:"additionalProperties,omitempty"`
								} `json:"anyOf"`
								Description string `json:"description"`
							} `json:"param"`
							Locate struct {
								Type       []string `json:"type"`
								Properties struct {
									ID struct {
										Type string `json:"type"`
									} `json:"id"`
									Prompt struct {
										Type string `json:"type"`
									} `json:"prompt"`
								} `json:"properties"`
								Required             []string `json:"required"`
								AdditionalProperties bool     `json:"additionalProperties"`
								Description          string   `json:"description"`
							} `json:"locate"`
						} `json:"properties"`
						Required             []string `json:"required"`
						AdditionalProperties bool     `json:"additionalProperties"`
					} `json:"items"`
					Description string `json:"description"`
				} `json:"actions"`
				MoreActionsNeededByInstruction struct {
					Type        string `json:"type"`
					Description string `json:"description"`
				} `json:"more_actions_needed_by_instruction"`
				Log struct {
					Type        string `json:"type"`
					Description string `json:"description"`
				} `json:"log"`
				Error struct {
					Type        []string `json:"type"`
					Description string   `json:"description"`
				} `json:"error"`
			} `json:"properties"`
			Required             []string `json:"required"`
			AdditionalProperties bool     `json:"additionalProperties"`
		} `json:"schema"`
	} `json:"json_schema"`
}

// GetPlanningResponseJsonFormat returns the planning response format based on page type
func GetPlanningResponseJsonFormat(pageType string) string {
	if pageType == "android" {
		return defaultPlanningResponseJsonFormat + `
- type: 'AndroidBackButton', trigger the system "back" operation on Android devices
  * { param: {} }
- type: 'AndroidHomeButton', trigger the system "home" operation on Android devices
  * { param: {} }
- type: 'AndroidRecentAppsButton', trigger the system "recent apps" operation on Android devices
  * { param: {} }`
	}
	return defaultPlanningResponseJsonFormat
}

// GenerateTaskBackgroundContext generates the task background context
func GenerateTaskBackgroundContext(userInstruction string, log string, userActionContext string) string {
	if log != "" {
		return fmt.Sprintf(`
Here is the user's instruction:

<instruction>
  <high_priority_knowledge>
    %s
  </high_priority_knowledge>

  %s
</instruction>

These are the logs from previous executions, which indicate what was done in the previous actions.
Do NOT repeat these actions.
<previous_logs>
%s
</previous_logs>
`, userActionContext, userInstruction, log)
	}

	return fmt.Sprintf(`
Here is the user's instruction:
<instruction>
  <high_priority_knowledge>
    %s
  </high_priority_knowledge>

  %s
</instruction>
`, userActionContext, userInstruction)
}

// AutomationUserPrompt generates the automation user prompt
func AutomationUserPrompt(vlMode bool, pageDescription string, taskBackgroundContext string) string {
	if vlMode {
		return taskBackgroundContext
	}

	return fmt.Sprintf(`
pageDescription:
=====================================
%s
=====================================

%s`, pageDescription, taskBackgroundContext)
}

const defaultPlanningResponseStringFormat = `
You are a helpful assistant.
`

// loadSystemPrompt loads the system prompt from a JSON file
func loadSystemPrompt(filePath string) (string, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("system prompt file does not exist: %s", filePath)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("error reading prompt file: %v", err)
	}

	// Read file content directly as prompt
	return string(data), nil
}
