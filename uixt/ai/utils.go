package ai

import (
	"context"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/uixt/option"
)

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
