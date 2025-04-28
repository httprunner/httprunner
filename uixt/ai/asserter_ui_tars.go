package ai

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/schema"
	"github.com/httprunner/httprunner/v5/internal/json"
	"github.com/httprunner/httprunner/v5/uixt/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type IAsserter interface {
	Assert(opts *AssertOptions) (*AssertionResponse, error)
}

// UI-TARS assertion system prompt
const uiTarsAssertionPrompt = `You are a senior testing engineer. User will give an assertion and a screenshot of a page. By carefully viewing the screenshot, please tell whether the assertion is truthy.

## Output Json String Format
` + "```" + `
"{
  "pass": <<is a boolean value from the enum [true, false], true means the assertion is truthy>>,
  "thought": "<<is a string, give the reason why the assertion is falsy or truthy. Otherwise.>>"
}"
` + "```" + `

## Rules **MUST** follow
- Make sure to return **only** the JSON, with **no additional** text or explanations.
- Use Chinese in 'thought' part.
- You **MUST** strictly follow up the **Output Json String Format**.`

// AssertionResponse represents the response from an AI assertion
type AssertionResponse struct {
	Pass    bool   `json:"pass"`
	Thought string `json:"thought"`
}

// UITarsAsserter handles assertion using UI-TARS VLM
type UITarsAsserter struct {
	ctx          context.Context
	model        *ark.ChatModel
	config       *ark.ChatModelConfig
	systemPrompt string
	history      ConversationHistory
}

// NewUITarsAsserter creates a new UITarsAsserter instance
func NewUITarsAsserter(ctx context.Context) (*UITarsAsserter, error) {
	config, err := GetArkModelConfig()
	if err != nil {
		return nil, err
	}
	chatModel, err := ark.NewChatModel(ctx, config)
	if err != nil {
		return nil, err
	}

	return &UITarsAsserter{
		ctx:          ctx,
		config:       config,
		model:        chatModel,
		systemPrompt: uiTarsAssertionPrompt,
	}, nil
}

// AssertOptions represents the input options for assertion
type AssertOptions struct {
	Assertion  string     `json:"assertion"`  // The assertion text to verify
	Screenshot string     `json:"screenshot"` // Base64 encoded screenshot
	Size       types.Size `json:"size"`       // Screen dimensions
}

// Assert performs the assertion check on the screenshot
func (a *UITarsAsserter) Assert(opts *AssertOptions) (*AssertionResponse, error) {
	// Validate input parameters
	if opts.Assertion == "" {
		return nil, errors.New("assertion text is required")
	}
	if opts.Screenshot == "" {
		return nil, errors.New("screenshot is required")
	}

	// Reset history for each new assertion
	a.history = ConversationHistory{
		{
			Role:    schema.System,
			Content: a.systemPrompt,
		},
	}

	// Create user message with screenshot and assertion
	userMsg := &schema.Message{
		Role: schema.User,
		MultiContent: []schema.ChatMessagePart{
			{
				Type: schema.ChatMessagePartTypeImageURL,
				ImageURL: &schema.ChatMessageImageURL{
					URL:    opts.Screenshot,
					Detail: schema.ImageURLDetailAuto,
				},
			},
			{
				Type: schema.ChatMessagePartTypeText,
				Text: fmt.Sprintf(`
Here is the assertion. Please tell whether it is truthy according to the screenshot.
=====================================
%s
=====================================
  `, opts.Assertion),
			},
		},
	}

	// Append user message to history
	a.history.Append(userMsg)

	// Call model service, generate response
	logRequest(a.history)
	startTime := time.Now()
	resp, err := a.model.Generate(a.ctx, a.history)
	log.Info().Float64("elapsed(s)", time.Since(startTime).Seconds()).
		Str("model", a.config.Model).Msg("call model service for assertion")
	if err != nil {
		return nil, fmt.Errorf("request model service failed: %w", err)
	}
	logResponse(resp)

	// Parse result
	result, err := parseAssertionResult(resp.Content)
	if err != nil {
		return nil, errors.Wrap(err, "parse assertion result failed")
	}

	// Append assistant message to history
	a.history.Append(&schema.Message{
		Role:    schema.Assistant,
		Content: resp.Content,
	})

	return result, nil
}

// parseAssertionResult 解析模型返回的JSON响应
func parseAssertionResult(content string) (*AssertionResponse, error) {
	// 1. 从响应中提取JSON内容
	jsonContent := extractJSON(content)
	if jsonContent == "" {
		return nil, errors.New("could not extract JSON from response")
	}

	// 2. 预处理和标准解析尝试
	jsonContent = prepareJSON(jsonContent)
	var result AssertionResponse
	if err := json.Unmarshal([]byte(jsonContent), &result); err == nil {
		return &result, nil
	}

	// 3. 备用：正则表达式解析
	if pass, thought := extractWithRegex(jsonContent); thought != "" {
		return &AssertionResponse{Pass: pass, Thought: thought}, nil
	}

	return nil, errors.New("failed to parse assertion result")
}

// prepareJSON 预处理JSON字符串，修复常见问题
func prepareJSON(jsonStr string) string {
	// 1. 去除可能的外层引号
	jsonStr = strings.TrimSpace(jsonStr)
	if strings.HasPrefix(jsonStr, "\"") && strings.HasSuffix(jsonStr, "\"") {
		jsonStr = jsonStr[1 : len(jsonStr)-1]
	}

	// 2. 转义thought内容中的引号
	thoughtRegex := regexp.MustCompile(`"thought":\s*"([^"]*)"`)
	matches := thoughtRegex.FindStringSubmatch(jsonStr)
	if len(matches) > 1 {
		thoughtValue := matches[1]
		fixedThought := strings.ReplaceAll(thoughtValue, "\"", "\\\"")
		jsonStr = strings.Replace(jsonStr, matches[0], fmt.Sprintf(`"thought": "%s"`, fixedThought), 1)
	}

	// 3. 处理换行和特殊字符
	jsonStr = strings.ReplaceAll(jsonStr, "\n", "\\n")
	jsonStr = strings.ReplaceAll(jsonStr, "\r", "\\r")
	jsonStr = strings.ReplaceAll(jsonStr, "\t", "\\t")

	return jsonStr
}

// extractWithRegex 使用正则表达式提取pass和thought值
func extractWithRegex(jsonStr string) (pass bool, thought string) {
	// 提取pass值
	passRegex := regexp.MustCompile(`"pass":\s*(true|false)`)
	passMatches := passRegex.FindStringSubmatch(jsonStr)

	// 提取thought值
	thoughtRegex := regexp.MustCompile(`"thought":\s*"([^"]*(?:"[^"]*)*)"`)
	thoughtMatches := thoughtRegex.FindStringSubmatch(jsonStr)

	if len(passMatches) > 1 && len(thoughtMatches) > 1 {
		// 处理提取的值
		pass = passMatches[1] == "true"
		thought = strings.ReplaceAll(thoughtMatches[1], "\\\"", "\"")
		thought = strings.ReplaceAll(thought, "\\\\", "\\")
		return pass, thought
	}

	return false, ""
}

// extractJSON extracts JSON content from a string that might contain markdown or other formatting
func extractJSON(content string) string {
	// Try to extract JSON directly
	content = strings.TrimSpace(content)

	// If the content is already a valid JSON, return it
	if strings.HasPrefix(content, "{") && strings.HasSuffix(content, "}") {
		return content
	}

	// Check for markdown code blocks with more flexible pattern
	jsonRegex := regexp.MustCompile(`(?:json)?\s*({[\s\S]*?})\s*`)
	matches := jsonRegex.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// Try a more robust approach for JSON with Chinese characters
	// First look for the outermost pair of curly braces
	startIdx := strings.Index(content, "{")
	if startIdx >= 0 {
		depth := 1
		for i := startIdx + 1; i < len(content); i++ {
			if content[i] == '{' {
				depth++
			} else if content[i] == '}' {
				depth--
				if depth == 0 {
					// Found the closing brace
					return content[startIdx : i+1]
				}
			}
		}
	}

	// Fallback to regex approach
	braceRegex := regexp.MustCompile(`{[\s\S]*?}`)
	matches = braceRegex.FindStringSubmatch(content)
	if len(matches) > 0 {
		return strings.TrimSpace(matches[0])
	}

	return ""
}
