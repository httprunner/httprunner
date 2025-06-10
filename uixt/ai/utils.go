package ai

import (
	"regexp"
	"strings"
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

	// Case 3: Try regex approach for markdown-like formats
	jsonRegex := regexp.MustCompile(`(?:json)?\s*({[\s\S]*?})\s*`)
	matches := jsonRegex.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// Case 4: Look for JSON object in the content using brace counting
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

	// Case 5: If content itself looks like JSON
	if strings.HasPrefix(content, "{") && strings.HasSuffix(content, "}") {
		return content
	}

	return ""
}
