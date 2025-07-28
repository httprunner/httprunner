package ai

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/getkin/kin-openapi/openapi3gen"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	arkModel "github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/httprunner/httprunner/v5/uixt/types"
)

// IAsserter interface defines the contract for assertion operations
type IAsserter interface {
	Assert(ctx context.Context, opts *AssertOptions) (*AssertionResult, error)
}

// AssertOptions represents the input options for assertion
type AssertOptions struct {
	Assertion    string     `json:"assertion"`     // The assertion text to verify
	Screenshot   string     `json:"screenshot"`    // Base64 encoded screenshot
	Size         types.Size `json:"size"`          // Screen dimensions
	ResetHistory bool       `json:"reset_history"` // whether to reset conversation history before assertion
}

// AssertionResult represents the response from an AI assertion
type AssertionResult struct {
	Pass      bool               `json:"pass"`
	Thought   string             `json:"thought"`
	Content   string             `json:"content,omitempty"` // raw response content
	ModelName string             `json:"model_name"`        // model name used for assertion
	Usage     *schema.TokenUsage `json:"usage,omitempty"`   // token usage statistics
}

// Asserter handles assertion using different AI models
type Asserter struct {
	modelConfig  *ModelConfig
	model        model.ToolCallingChatModel
	systemPrompt string
	history      ConversationHistory
}

// NewAsserter creates a new Asserter instance
func NewAsserter(ctx context.Context, modelConfig *ModelConfig) (*Asserter, error) {
	asserter := &Asserter{
		modelConfig:  modelConfig,
		systemPrompt: defaultAssertionPrompt,
	}

	if option.IS_UI_TARS(modelConfig.ModelType) {
		asserter.systemPrompt += "\n" + uiTarsAssertionResponseFormat
	} else {
		// define output format
		type OutputFormat struct {
			Thought string `json:"thought"`
			Pass    bool   `json:"pass"`
			Error   string `json:"error,omitempty"`
		}
		outputFormatSchema, err := openapi3gen.NewSchemaRefForValue(&OutputFormat{}, nil)
		if err != nil {
			return nil, errors.Wrap(code.LLMPrepareRequestError, err.Error())
		}
		// set structured response format
		modelConfig.ChatModelConfig.ResponseFormat = &ark.ResponseFormat{
			Type: arkModel.ResponseFormatJSONSchema,
			JSONSchema: &arkModel.ResponseFormatJSONSchemaJSONSchemaParam{
				Name:        "assertion_result",
				Description: "data that describes assertion result",
				Schema:      outputFormatSchema.Value,
				Strict:      false,
			},
		}
	}

	var err error
	asserter.model, err = ark.NewChatModel(ctx, modelConfig.ChatModelConfig)
	if err != nil {
		return nil, errors.Wrap(code.LLMPrepareRequestError, err.Error())
	}

	return asserter, nil
}

// Assert performs the assertion check on the screenshot
func (a *Asserter) Assert(ctx context.Context, opts *AssertOptions) (result *AssertionResult, err error) {
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
	message, err := callModelWithLogging(ctx, a.model, a.history,
		a.modelConfig.ModelType, "assertion")
	if err != nil {
		return nil, errors.Wrap(code.LLMRequestServiceError, err.Error())
	}

	defer func() {
		// Extract usage information if available
		if message.ResponseMeta != nil && message.ResponseMeta.Usage != nil {
			result.Usage = message.ResponseMeta.Usage
		}
	}()

	// Parse result
	result, err = parseAssertionResult(message.Content, a.modelConfig.ModelType)
	if err != nil {
		return nil, errors.Wrap(code.LLMParseAssertionResponseError, err.Error())
	}

	// Append assistant message to history
	a.history.Append(&schema.Message{
		Role:    schema.Assistant,
		Content: message.Content,
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
func parseAssertionResult(content string, modelType option.LLMServiceType) (*AssertionResult, error) {
	var result AssertionResult

	// Use the generic structured response parser
	if err := parseStructuredResponse(content, &result); err != nil {
		log.Warn().
			Interface("original_content", content).
			Msg("parse assertion result failed")
		return nil, errors.Wrap(code.LLMParseAssertionResponseError, err.Error())
	}

	result.ModelName = string(modelType)
	result.Content = content // Store the original response content
	return &result, nil
}
