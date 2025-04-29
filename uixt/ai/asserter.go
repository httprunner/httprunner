package ai

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/json"
	"github.com/httprunner/httprunner/v5/uixt/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// IAsserter interface defines the contract for assertion operations
type IAsserter interface {
	Assert(opts *AssertOptions) (*AssertionResponse, error)
}

// AssertOptions represents the input options for assertion
type AssertOptions struct {
	Assertion  string     `json:"assertion"`  // The assertion text to verify
	Screenshot string     `json:"screenshot"` // Base64 encoded screenshot
	Size       types.Size `json:"size"`       // Screen dimensions
}

// AssertionResponse represents the response from an AI assertion
type AssertionResponse struct {
	Pass    bool   `json:"pass"`
	Thought string `json:"thought"`
}

// Asserter handles assertion using different AI models
type Asserter struct {
	ctx          context.Context
	model        model.ToolCallingChatModel
	systemPrompt string
	history      ConversationHistory
	modelType    LLMServiceType
}

// NewAsserter creates a new Asserter instance
func NewAsserter(ctx context.Context, modelType LLMServiceType) (*Asserter, error) {
	asserter := &Asserter{
		ctx:          ctx,
		modelType:    modelType,
		systemPrompt: getAssertionSystemPrompt(modelType),
	}

	switch modelType {
	case LLMServiceTypeUITARS:
		config, err := GetArkModelConfig()
		if err != nil {
			return nil, err
		}
		asserter.model, err = ark.NewChatModel(ctx, config)
		if err != nil {
			return nil, err
		}
	case LLMServiceTypeGPT4Vision, LLMServiceTypeGPT4o:
		config, err := GetOpenAIModelConfig()
		if err != nil {
			return nil, err
		}
		asserter.model, err = openai.NewChatModel(ctx, config)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("not supported model type for asserter")
	}

	return asserter, nil
}

// getAssertionSystemPrompt returns the appropriate system prompt for the given model type
func getAssertionSystemPrompt(modelType LLMServiceType) string {
	if modelType == LLMServiceTypeUITARS {
		return defaultAssertionPrompt + "\n\n" + uiTarsAssertionResponseFormat
	}
	return defaultAssertionPrompt + "\n\n" + defaultAssertionResponseJsonFormat
}

// Assert performs the assertion check on the screenshot
func (a *Asserter) Assert(opts *AssertOptions) (*AssertionResponse, error) {
	// Validate input parameters
	if err := validateAssertionInput(opts); err != nil {
		return nil, errors.Wrap(err, "validate assertion parameters failed")
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
		Str("model", string(a.modelType)).Msg("call model service for assertion")
	if err != nil {
		return nil, errors.Wrap(code.LLMRequestServiceError, err.Error())
	}
	logResponse(resp)

	// Parse result
	result, err := parseAssertionResult(resp.Content)
	if err != nil {
		return nil, errors.Wrap(code.LLMParseAssertionResponseError, err.Error())
	}

	// Append assistant message to history
	a.history.Append(&schema.Message{
		Role:    schema.Assistant,
		Content: resp.Content,
	})

	return result, nil
}

// validateAssertionInput validates the input parameters for assertion
func validateAssertionInput(opts *AssertOptions) error {
	if opts.Assertion == "" {
		return errors.Wrap(code.LLMPrepareRequestError, "assertion text is required")
	}
	if opts.Screenshot == "" {
		return errors.Wrap(code.LLMPrepareRequestError, "screenshot is required")
	}
	return nil
}

// parseAssertionResult parses the model response into AssertionResponse
func parseAssertionResult(content string) (*AssertionResponse, error) {
	// Extract JSON content from response
	jsonContent := extractJSON(content)
	if jsonContent == "" {
		return nil, errors.New("could not extract JSON from response")
	}

	// Parse JSON response
	var result AssertionResponse
	if err := json.Unmarshal([]byte(jsonContent), &result); err != nil {
		return nil, errors.Wrap(code.LLMParseAssertionResponseError, err.Error())
	}

	return &result, nil
}

// extractJSON extracts JSON content from a string that might contain markdown or other formatting
func extractJSON(content string) string {
	content = strings.TrimSpace(content)

	// If the content is already a valid JSON, return it
	if strings.HasPrefix(content, "{") && strings.HasSuffix(content, "}") {
		return content
	}

	// Try to extract JSON from markdown code blocks
	jsonRegex := regexp.MustCompile(`(?:json)?\s*({[\s\S]*?})\s*`)
	matches := jsonRegex.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// Try a more robust approach for JSON with Chinese characters
	startIdx := strings.Index(content, "{")
	if startIdx >= 0 {
		depth := 1
		for i := startIdx + 1; i < len(content); i++ {
			if content[i] == '{' {
				depth++
			} else if content[i] == '}' {
				depth--
				if depth == 0 {
					return content[startIdx : i+1]
				}
			}
		}
	}

	return content
}
