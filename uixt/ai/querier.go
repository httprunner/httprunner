package ai

import (
	"context"
	"fmt"
	"reflect"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/getkin/kin-openapi/openapi3gen"
	"github.com/pkg/errors"
	arkModel "github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/json"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/httprunner/httprunner/v5/uixt/types"
)

// IQuerier interface defines the contract for query operations
type IQuerier interface {
	Query(ctx context.Context, opts *QueryOptions) (*QueryResult, error)
}

// QueryOptions represents the input options for query
type QueryOptions struct {
	Query        string      `json:"query"`                  // The query text to extract information
	Screenshot   string      `json:"screenshot"`             // Base64 encoded screenshot
	Size         types.Size  `json:"size"`                   // Screen dimensions
	OutputSchema interface{} `json:"outputSchema,omitempty"` // Custom output schema for structured response
}

// QueryResult represents the response from an AI query
type QueryResult struct {
	Content   string             `json:"content"`         // The extracted content/information
	Thought   string             `json:"thought"`         // The reasoning process
	Data      interface{}        `json:"data,omitempty"`  // Structured data when OutputSchema is provided
	ModelName string             `json:"model_name"`      // model name used for query
	Usage     *schema.TokenUsage `json:"usage,omitempty"` // token usage statistics
}

// Querier handles query operations using different AI models
type Querier struct {
	modelConfig  *ModelConfig
	model        model.ToolCallingChatModel
	systemPrompt string
	history      ConversationHistory
}

// NewQuerier creates a new Querier instance
func NewQuerier(ctx context.Context, modelConfig *ModelConfig) (*Querier, error) {
	querier := &Querier{
		modelConfig:  modelConfig,
		systemPrompt: defaultQueryPrompt,
	}

	if option.IS_UI_TARS(modelConfig.ModelType) {
		querier.systemPrompt += "\n" + uiTarsQueryResponseFormat
	} else {
		// define default output format
		type OutputFormat struct {
			Content string `json:"content"`
			Thought string `json:"thought"`
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
				Name:        "query_result",
				Description: "data that describes query result",
				Schema:      outputFormatSchema.Value,
				Strict:      false,
			},
		}
	}

	var err error
	querier.model, err = ark.NewChatModel(ctx, modelConfig.ChatModelConfig)
	if err != nil {
		return nil, errors.Wrap(code.LLMPrepareRequestError, err.Error())
	}

	return querier, nil
}

// callModelWithLogging calls the model with automatic logging and timing

// Query performs the information extraction from the screenshot
func (q *Querier) Query(ctx context.Context, opts *QueryOptions) (result *QueryResult, err error) {
	// Validate input parameters
	if err := validateQueryInput(opts); err != nil {
		return nil, errors.Wrap(err, "validate query parameters failed")
	}

	// Handle custom output schema if provided
	if opts.OutputSchema != nil {
		return q.queryWithCustomSchema(ctx, opts)
	}

	// Reset history for each new query
	q.history = ConversationHistory{
		{
			Role:    schema.System,
			Content: q.systemPrompt,
		},
	}

	// Create user message with screenshot and query
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
Here is the query. Please extract the requested information from the screenshot.
=====================================
%s
=====================================
  `, opts.Query),
			},
		},
	}

	// Append user message to history
	q.history.Append(userMsg)

	// Call model service with logging
	message, err := callModelWithLogging(ctx, q.model, q.history,
		q.modelConfig.ModelType, "query")
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
	result, err = parseQueryResult(message.Content, q.modelConfig.ModelType)
	if err != nil {
		return nil, errors.Wrap(code.LLMParseQueryResponseError, err.Error())
	}

	// Append assistant message to history
	q.history.Append(&schema.Message{
		Role:    schema.Assistant,
		Content: message.Content,
	})

	return result, nil
}

// validateQueryInput validates the input parameters for query
func validateQueryInput(opts *QueryOptions) error {
	if opts.Query == "" {
		return errors.Wrap(code.LLMPrepareRequestError, "query text is required")
	}
	if opts.Screenshot == "" {
		return errors.Wrap(code.LLMPrepareRequestError, "screenshot is required")
	}
	return nil
}

// parseQueryResult parses the model response into QueryResult
func parseQueryResult(content string, modelType option.LLMServiceType) (*QueryResult, error) {
	var result QueryResult

	// Use the generic structured response parser with enhanced error recovery
	if err := parseStructuredResponse(content, &result); err != nil {
		// If parseStructuredResponse fails completely, treat content as plain text
		return &QueryResult{
			Content:   content,
			Thought:   "Failed to parse response, returning raw content",
			ModelName: string(modelType),
		}, nil
	}

	result.ModelName = string(modelType)
	return &result, nil
}

// queryWithCustomSchema performs query with custom output schema
func (q *Querier) queryWithCustomSchema(ctx context.Context, opts *QueryOptions) (*QueryResult, error) {
	// Create a new model config with custom schema
	modelConfig := *q.modelConfig

	if !option.IS_UI_TARS(modelConfig.ModelType) {
		// Generate schema from the provided output schema
		outputFormatSchema, err := openapi3gen.NewSchemaRefForValue(opts.OutputSchema, nil)
		if err != nil {
			return nil, errors.Wrap(code.LLMPrepareRequestError, err.Error())
		}

		// Create custom response format with the provided schema
		modelConfig.ChatModelConfig.ResponseFormat = &ark.ResponseFormat{
			Type: arkModel.ResponseFormatJSONSchema,
			JSONSchema: &arkModel.ResponseFormatJSONSchemaJSONSchemaParam{
				Name:        "custom_query_result",
				Description: "custom structured data response",
				Schema:      outputFormatSchema.Value,
				Strict:      false,
			},
		}
	}

	// Create a new model instance with custom schema
	model, err := ark.NewChatModel(ctx, modelConfig.ChatModelConfig)
	if err != nil {
		return nil, errors.Wrap(code.LLMPrepareRequestError, err.Error())
	}

	// Reset history for each new query
	systemPrompt := q.systemPrompt
	if option.IS_UI_TARS(modelConfig.ModelType) {
		systemPrompt += "\n" + uiTarsQueryResponseFormat
	} else {
		// Add instruction for structured output
		systemPrompt += "\n\nPlease respond with structured data according to the specified schema. Include both the structured data and your reasoning process."
	}

	history := ConversationHistory{
		{
			Role:    schema.System,
			Content: systemPrompt,
		},
	}

	// Create user message with screenshot and query
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
Here is the query. Please extract the requested information from the screenshot and return it in the specified structured format.
=====================================
%s
=====================================
  `, opts.Query),
			},
		},
	}

	// Append user message to history
	history.Append(userMsg)

	// Call model service with logging
	message, err := callModelWithLogging(ctx, model, history, modelConfig.ModelType, "custom schema query")
	if err != nil {
		return nil, errors.Wrap(code.LLMRequestServiceError, err.Error())
	}

	// Parse result with custom schema
	result, err := parseCustomSchemaResult(message.Content, opts.OutputSchema)
	if err != nil {
		return nil, errors.Wrap(code.LLMParseQueryResponseError, err.Error())
	}

	// Append assistant message to history
	q.history.Append(&schema.Message{
		Role:    schema.Assistant,
		Content: message.Content,
	})

	return result, nil
}

// setDefaultFieldValue sets a default value for a field in the structured data using reflection
func setDefaultFieldValue(structValue reflect.Value, fieldName, defaultValue string) {
	if field := structValue.FieldByName(fieldName); field.IsValid() && field.CanSet() && field.Kind() == reflect.String {
		field.SetString(defaultValue)
	}
}

// ensureDefaultValues ensures that Content and Thought fields have default values if empty
func ensureDefaultValues(result *QueryResult, structuredData interface{}) {
	const (
		defaultContent = "Structured data extracted successfully"
		defaultThought = "Parsed structured response according to custom schema"
	)

	// Set defaults for QueryResult
	if result.Content == "" {
		result.Content = defaultContent
	}
	if result.Thought == "" {
		result.Thought = defaultThought
	}

	// Set defaults in structured data if it's a pointer to struct
	if structuredData != nil {
		if structValue := reflect.ValueOf(structuredData); structValue.Kind() == reflect.Ptr {
			if elem := structValue.Elem(); elem.IsValid() && elem.Kind() == reflect.Struct {
				if result.Content == defaultContent {
					setDefaultFieldValue(elem, "Content", defaultContent)
				}
				if result.Thought == defaultThought {
					setDefaultFieldValue(elem, "Thought", defaultThought)
				}
			}
		}
	}
}

// parseCustomSchemaResult parses the model response with custom schema
func parseCustomSchemaResult(content string, outputSchema interface{}) (*QueryResult, error) {
	// Extract JSON content from response
	jsonContent := extractJSONFromContent(content)
	if jsonContent == "" {
		// If no JSON found, treat the entire content as the result
		return &QueryResult{
			Content: content,
			Thought: "Direct response from model",
		}, nil
	}

	// Handle OpenAI structured output properties wrapper
	actualJSONContent := unwrapPropertiesIfNeeded(jsonContent)

	// Try direct unmarshaling first (most efficient)
	if result, err := tryDirectUnmarshal(actualJSONContent, outputSchema); err == nil {
		return result, nil
	}

	// Fallback: try generic parsing and conversion
	if result, err := tryGenericParsingAndConversion(actualJSONContent, outputSchema); err == nil {
		return result, nil
	}

	// Final fallback: treat as plain text
	return &QueryResult{
		Content: content,
		Thought: "Failed to parse as structured data, returning raw content",
	}, nil
}

// unwrapPropertiesIfNeeded handles OpenAI structured output properties wrapper
func unwrapPropertiesIfNeeded(jsonContent string) string {
	var tempMap map[string]interface{}
	if err := json.Unmarshal([]byte(jsonContent), &tempMap); err == nil {
		if properties, exists := tempMap["properties"]; exists {
			if propertiesBytes, err := json.Marshal(properties); err == nil {
				return string(propertiesBytes)
			}
		}
	}
	return jsonContent
}

// tryDirectUnmarshal attempts to unmarshal directly into the schema type
func tryDirectUnmarshal(jsonContent string, outputSchema interface{}) (*QueryResult, error) {
	// Create a new instance of the schema type
	newInstance := createSchemaInstance(outputSchema)

	// Try to unmarshal directly into the schema type
	if err := json.Unmarshal([]byte(jsonContent), newInstance); err != nil {
		return nil, err
	}

	// Create result with the typed data
	result := &QueryResult{Data: newInstance}

	// Extract content and thought fields
	extractContentAndThoughtFromStruct(result, newInstance)
	if result.Content == "" && result.Thought == "" {
		extractContentAndThoughtFromJSON(result, jsonContent)
	}

	// Ensure default values are set
	ensureDefaultValues(result, newInstance)
	return result, nil
}

// tryGenericParsingAndConversion attempts generic parsing and type conversion
func tryGenericParsingAndConversion(jsonContent string, outputSchema interface{}) (*QueryResult, error) {
	var structuredData interface{}
	if err := json.Unmarshal([]byte(jsonContent), &structuredData); err != nil {
		return nil, err
	}

	// Try to convert to the expected schema type
	if convertedData, err := convertToSchemaType(structuredData, outputSchema); err == nil {
		result := &QueryResult{Data: convertedData}
		extractContentAndThoughtFromMap(result, structuredData)
		ensureDefaultValues(result, convertedData)
		return result, nil
	}

	// If conversion failed, store the generic data
	if dataMap, ok := structuredData.(map[string]interface{}); ok {
		result := &QueryResult{Data: structuredData}
		extractContentAndThoughtFromMap(result, dataMap)
		ensureDefaultValues(result, nil)
		return result, nil
	}

	return nil, errors.New("failed to parse structured data")
}

// createSchemaInstance creates a new instance of the schema type
func createSchemaInstance(outputSchema interface{}) interface{} {
	schemaType := reflect.TypeOf(outputSchema)
	if schemaType.Kind() == reflect.Ptr {
		schemaType = schemaType.Elem()
	}
	return reflect.New(schemaType).Interface()
}

// extractContentAndThoughtFromStruct extracts content and thought from struct fields using reflection
func extractContentAndThoughtFromStruct(result *QueryResult, structData interface{}) {
	schemaValue := reflect.ValueOf(structData).Elem()

	if contentField := schemaValue.FieldByName("Content"); contentField.IsValid() && contentField.Kind() == reflect.String {
		result.Content = contentField.String()
	}

	if thoughtField := schemaValue.FieldByName("Thought"); thoughtField.IsValid() && thoughtField.Kind() == reflect.String {
		result.Thought = thoughtField.String()
	}
}

// extractContentAndThoughtFromJSON extracts content and thought from JSON map
func extractContentAndThoughtFromJSON(result *QueryResult, jsonContent string) {
	var dataMap map[string]interface{}
	if err := json.Unmarshal([]byte(jsonContent), &dataMap); err == nil {
		extractContentAndThoughtFromMap(result, dataMap)
	}
}

// extractContentAndThoughtFromMap extracts content and thought from a map
func extractContentAndThoughtFromMap(result *QueryResult, dataMap interface{}) {
	if mapData, ok := dataMap.(map[string]interface{}); ok {
		if content, exists := mapData["content"]; exists {
			if contentStr, ok := content.(string); ok {
				result.Content = contentStr
			}
		}
		if thought, exists := mapData["thought"]; exists {
			if thoughtStr, ok := thought.(string); ok {
				result.Thought = thoughtStr
			}
		}
	}
}

// convertToSchemaType converts generic data to the specified schema type
func convertToSchemaType(data interface{}, outputSchema interface{}) (interface{}, error) {
	// Get the type of the output schema
	schemaType := reflect.TypeOf(outputSchema)
	if schemaType.Kind() == reflect.Ptr {
		schemaType = schemaType.Elem()
	}

	// Create a new instance of the schema type
	newInstance := reflect.New(schemaType).Interface()

	// Convert via JSON marshaling/unmarshaling
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal data to JSON")
	}

	if err := json.Unmarshal(jsonData, newInstance); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal data to target type")
	}

	return newInstance, nil
}

// ConvertQueryResultData converts QueryResult.Data to the specified type T
// This is a helper function for type-safe conversion of the structured data
//
// Note: When using QueryOptions.OutputSchema, the Data field is automatically
// converted to the correct type, so this function is typically not needed.
// This function is mainly useful for:
// 1. Converting data when OutputSchema was not used
// 2. Converting to a different type than the original OutputSchema
// 3. Handling legacy code or edge cases
func ConvertQueryResultData[T any](result *QueryResult) (*T, error) {
	if result.Data == nil {
		return nil, errors.New("no structured data available")
	}

	// If Data is already of the correct type, return it directly
	if typedData, ok := result.Data.(*T); ok {
		return typedData, nil
	}

	// If Data is a pointer to the correct type, dereference and return
	if reflect.TypeOf(result.Data).Kind() == reflect.Ptr {
		if typedData, ok := result.Data.(*T); ok {
			return typedData, nil
		}
		// Try to get the value that the pointer points to
		dataValue := reflect.ValueOf(result.Data)
		if dataValue.Kind() == reflect.Ptr && !dataValue.IsNil() {
			elem := dataValue.Elem()
			if elem.Type() == reflect.TypeOf((*T)(nil)).Elem() {
				typedData := elem.Interface().(T)
				return &typedData, nil
			}
		}
	}

	// Fallback: try to convert via JSON marshaling/unmarshaling
	jsonData, err := json.Marshal(result.Data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal data to JSON")
	}

	var converted T
	if err := json.Unmarshal(jsonData, &converted); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal data to target type")
	}

	return &converted, nil
}
