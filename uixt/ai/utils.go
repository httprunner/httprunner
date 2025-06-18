package ai

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/internal/json"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/pkg/errors"
)

// PlanningJSONResponse represents the JSON response structure for planning
type PlanningJSONResponse struct {
	Actions []Action `json:"actions"`
	Thought string   `json:"thought"`
	Error   string   `json:"error"`
}

// extractJSONFromContent extracts JSON content from various formats in the response
// This function handles multiple formats:
// 1. ```json ... ``` markdown code blocks
// 2. ``` ... ``` generic code blocks
// 3. JSON objects embedded in text
// 4. Plain JSON content
func extractJSONFromContent(content string) string {
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

	// Case 3: Look for JSON object in the content using rune-based brace counting (most reliable method)
	start := strings.Index(content, "{")
	if start != -1 {
		// Find the matching closing brace using rune-based iteration to handle UTF-8 properly
		braceCount := 0
		inString := false
		escaped := false

		// Use byte-based iteration but track string state properly
		for i := start; i < len(content); {
			r, size := utf8.DecodeRuneInString(content[i:])

			if escaped {
				escaped = false
				i += size
				continue
			}

			if r == '\\' && inString {
				escaped = true
				i += size
				continue
			}

			if r == '"' {
				inString = !inString
				i += size
				continue
			}

			if !inString {
				if r == '{' {
					braceCount++
				} else if r == '}' {
					braceCount--
					if braceCount == 0 {
						jsonContent := strings.TrimSpace(content[start : i+size])
						return jsonContent
					}
				}
			}
			i += size
		}
	}

	// Case 4: Try regex approach for markdown-like formats (fallback)
	jsonRegex := regexp.MustCompile(`(?:json)?\s*({[\s\S]*?})\s*`)
	matches := jsonRegex.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// Case 5: If content itself looks like JSON
	if strings.HasPrefix(content, "{") && strings.HasSuffix(content, "}") {
		return content
	}

	return ""
}

// sanitizeUTF8Content cleans invalid UTF-8 characters from content
func sanitizeUTF8Content(content string) string {
	if utf8.ValidString(content) {
		return content
	}

	// Convert to bytes and filter out invalid UTF-8 sequences
	bytes := []byte(content)
	var validBytes []byte

	for len(bytes) > 0 {
		r, size := utf8.DecodeRune(bytes)
		if r != utf8.RuneError {
			// Valid rune, keep it
			validBytes = append(validBytes, bytes[:size]...)
		}
		// Skip invalid bytes (including RuneError)
		bytes = bytes[size:]
	}

	return string(validBytes)
}

// parseJSONWithFallback attempts to parse JSON with multiple strategies for any struct type
func parseJSONWithFallback(jsonContent string, result interface{}) error {
	// Strategy 1: Direct JSON unmarshaling
	if err := json.Unmarshal([]byte(jsonContent), result); err == nil {
		// For specific types, ensure required fields have default values even after successful parsing
		switch v := result.(type) {
		case *QueryResult:
			// Ensure QueryResult has meaningful defaults for empty fields
			if v.Content == "" && v.Thought == "" {
				v.Content = "Empty response content"
				v.Thought = "No content extracted from response"
			} else if v.Content == "" {
				v.Content = "No content extracted"
			} else if v.Thought == "" {
				v.Thought = "Successfully parsed structured response"
			}
		case *AssertionResult:
			// Ensure AssertionResult has meaningful defaults
			if v.Thought == "" {
				v.Thought = "Successfully parsed assertion response"
			}
		}
		return nil
	}

	// Strategy 2: Try cleaning JSON content and parse again
	cleanedJSON := cleanJSONContent(jsonContent)
	if err := json.Unmarshal([]byte(cleanedJSON), result); err == nil {
		// Apply the same default value logic for cleaned JSON
		switch v := result.(type) {
		case *QueryResult:
			if v.Content == "" && v.Thought == "" {
				v.Content = "Empty response content"
				v.Thought = "No content extracted from response"
			} else if v.Content == "" {
				v.Content = "No content extracted"
			} else if v.Thought == "" {
				v.Thought = "Successfully parsed structured response"
			}
		case *AssertionResult:
			if v.Thought == "" {
				v.Thought = "Successfully parsed assertion response"
			}
		}
		return nil
	}

	// Strategy 3: For specific types, try manual extraction or content analysis
	switch v := result.(type) {
	case *AssertionResult:
		if fallbackResult, err := extractAssertionFieldsManually(jsonContent); err == nil {
			*v = *fallbackResult
			return nil
		}
		// Final fallback for assertions: content analysis
		*v = *analyzeContentForAssertion(jsonContent)
		return nil

	case *QueryResult:
		// For QueryResult, try basic field extraction
		if fallbackResult, err := extractQueryFieldsManually(jsonContent); err == nil {
			*v = *fallbackResult
			return nil
		}
		// Fallback to treating content as plain text
		*v = QueryResult{
			Content: jsonContent,
			Thought: "Failed to parse as JSON, returning raw content",
		}
		return nil

	case *PlanningJSONResponse:
		// For PlanningJSONResponse, try basic field extraction
		if fallbackResult, err := extractPlanningFieldsManually(jsonContent); err == nil {
			*v = *fallbackResult
			return nil
		}
		// Fallback with empty actions but preserve any recognizable thought content
		*v = PlanningJSONResponse{
			Actions: []Action{},
			Thought: "Failed to parse structured response",
			Error:   "JSON parsing failed, returning minimal structure",
		}
		return nil
	}

	return errors.New("failed to parse JSON with all strategies")
}

// extractAssertionFieldsManually extracts pass and thought fields from text
func extractAssertionFieldsManually(content string) (*AssertionResult, error) {
	result := &AssertionResult{}

	// Try to extract "pass" field
	if strings.Contains(strings.ToLower(content), `"pass":true`) ||
		strings.Contains(strings.ToLower(content), `"pass": true`) {
		result.Pass = true
	} else if strings.Contains(strings.ToLower(content), `"pass":false`) ||
		strings.Contains(strings.ToLower(content), `"pass": false`) {
		result.Pass = false
	} else {
		return nil, errors.New("cannot extract pass field")
	}

	// Try to extract "thought" field
	thoughtStart := strings.Index(content, `"thought"`)
	if thoughtStart != -1 {
		thoughtSection := content[thoughtStart:]
		colonIndex := strings.Index(thoughtSection, ":")
		if colonIndex != -1 {
			afterColon := strings.TrimSpace(thoughtSection[colonIndex+1:])
			if strings.HasPrefix(afterColon, `"`) {
				// Find the matching closing quote, handling escaped quotes
				thoughtContent := extractQuotedString(afterColon)
				result.Thought = thoughtContent
			}
		}
	}

	return result, nil
}

// extractQuotedString extracts content from a quoted string, handling escaped quotes
func extractQuotedString(s string) string {
	if !strings.HasPrefix(s, `"`) {
		return ""
	}

	s = s[1:] // Remove opening quote
	var result strings.Builder
	escaped := false

	for _, r := range s {
		if escaped {
			result.WriteRune(r)
			escaped = false
			continue
		}

		if r == '\\' {
			escaped = true
			continue
		}

		if r == '"' {
			// Found closing quote
			return result.String()
		}

		result.WriteRune(r)
	}

	return result.String()
}

// cleanJSONContent removes common JSON formatting issues
func cleanJSONContent(content string) string {
	// Remove any non-printable characters
	cleaned := strings.Map(func(r rune) rune {
		if r >= 32 && r < 127 || r > 127 { // Keep printable ASCII and Unicode
			return r
		}
		return -1 // Remove non-printable characters
	}, content)

	// Remove any trailing commas before closing braces/brackets
	cleaned = strings.ReplaceAll(cleaned, ",}", "}")
	cleaned = strings.ReplaceAll(cleaned, ",]", "]")

	return cleaned
}

// analyzeContentForAssertion creates a fallback result by analyzing content
func analyzeContentForAssertion(content string) *AssertionResult {
	content = strings.ToLower(content)

	// Simple heuristic: look for positive/negative indicators
	positiveIndicators := []string{"true", "pass", "success", "correct", "valid", "match"}
	negativeIndicators := []string{"false", "fail", "error", "incorrect", "invalid", "mismatch"}

	positiveCount := 0
	negativeCount := 0

	for _, indicator := range positiveIndicators {
		if strings.Contains(content, indicator) {
			positiveCount++
		}
	}

	for _, indicator := range negativeIndicators {
		if strings.Contains(content, indicator) {
			negativeCount++
		}
	}

	pass := positiveCount > negativeCount
	thought := fmt.Sprintf("Fallback analysis of malformed response (positive: %d, negative: %d)",
		positiveCount, negativeCount)

	return &AssertionResult{
		Pass:    pass,
		Thought: thought,
	}
}

// extractQueryFieldsManually extracts content and thought fields for QueryResult
func extractQueryFieldsManually(content string) (*QueryResult, error) {
	result := &QueryResult{}

	// Try to extract "content" field
	if contentStart := strings.Index(content, `"content"`); contentStart != -1 {
		contentSection := content[contentStart:]
		if colonIndex := strings.Index(contentSection, ":"); colonIndex != -1 {
			afterColon := strings.TrimSpace(contentSection[colonIndex+1:])
			if strings.HasPrefix(afterColon, `"`) {
				result.Content = extractQuotedString(afterColon)
			}
		}
	}

	// Try to extract "thought" field
	if thoughtStart := strings.Index(content, `"thought"`); thoughtStart != -1 {
		thoughtSection := content[thoughtStart:]
		if colonIndex := strings.Index(thoughtSection, ":"); colonIndex != -1 {
			afterColon := strings.TrimSpace(thoughtSection[colonIndex+1:])
			if strings.HasPrefix(afterColon, `"`) {
				result.Thought = extractQuotedString(afterColon)
			}
		}
	}

	// If we couldn't extract any fields, return error
	if result.Content == "" && result.Thought == "" {
		return nil, errors.New("cannot extract content or thought fields")
	}

	// Set defaults for missing fields (ALWAYS set defaults if any field was extracted)
	if result.Content == "" {
		result.Content = "Extracted partial information"
	}
	if result.Thought == "" {
		result.Thought = "Partial extraction from malformed response"
	}

	return result, nil
}

// extractPlanningFieldsManually extracts thought and error fields for PlanningJSONResponse
func extractPlanningFieldsManually(content string) (*PlanningJSONResponse, error) {
	result := &PlanningJSONResponse{
		Actions: []Action{}, // Default to empty actions
	}

	// Try to extract "thought" field
	if thoughtStart := strings.Index(content, `"thought"`); thoughtStart != -1 {
		thoughtSection := content[thoughtStart:]
		if colonIndex := strings.Index(thoughtSection, ":"); colonIndex != -1 {
			afterColon := strings.TrimSpace(thoughtSection[colonIndex+1:])
			if strings.HasPrefix(afterColon, `"`) {
				result.Thought = extractQuotedString(afterColon)
			}
		}
	}

	// Try to extract "error" field
	if errorStart := strings.Index(content, `"error"`); errorStart != -1 {
		errorSection := content[errorStart:]
		if colonIndex := strings.Index(errorSection, ":"); colonIndex != -1 {
			afterColon := strings.TrimSpace(errorSection[colonIndex+1:])
			if strings.HasPrefix(afterColon, `"`) {
				result.Error = extractQuotedString(afterColon)
			}
		}
	}

	// If we couldn't extract any meaningful fields, return error
	if result.Thought == "" && result.Error == "" {
		return nil, errors.New("cannot extract thought or error fields")
	}

	// Set defaults for missing fields
	if result.Thought == "" {
		result.Thought = "Partial extraction from malformed response"
	}

	return result, nil
}

// parseStructuredResponse parses model response into structured format with error recovery
func parseStructuredResponse(content string, result interface{}) error {
	// Clean and validate UTF-8 content first
	cleanContent := sanitizeUTF8Content(content)

	// Extract JSON content from response
	jsonContent := extractJSONFromContent(cleanContent)
	if jsonContent == "" {
		// If JSON extraction failed, try parsing the content directly as a fallback
		jsonContent = cleanContent
	}

	// Parse JSON response with error recovery
	return parseJSONWithFallback(jsonContent, result)
}

// callModelWithLogging is a common function to call model with logging and timing
// It handles the common pattern of:
// 1. Log request
// 2. Start timing
// 3. Call model.Generate
// 4. Log timing and model info
// 5. Log response
func callModelWithLogging(ctx context.Context, model model.ToolCallingChatModel, history ConversationHistory, modelType option.LLMServiceType, operation string) (*schema.Message, error) {
	logRequest(history)

	startTime := time.Now()
	defer func() {
		log.Debug().Float64("elapsed(s)", time.Since(startTime).Seconds()).
			Str("model", string(modelType)).
			Msgf("call model service for %s", operation)
	}()

	message, err := model.Generate(ctx, history)
	if err != nil {
		return nil, err
	}

	logResponse(message)
	return message, nil
}
