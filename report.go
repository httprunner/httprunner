package hrp

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/uixt"
	"github.com/httprunner/httprunner/v5/uixt/option"
)

// GenerateHTMLReportFromFiles is a convenience function to generate HTML report
func GenerateHTMLReportFromFiles(summaryFile, logFile, outputFile string) error {
	generator, err := NewHTMLReportGenerator(summaryFile, logFile)
	if err != nil {
		return errors.Wrap(err, "failed to create HTML report generator")
	}
	err = generator.GenerateReport(outputFile)
	if err != nil {
		return errors.Wrap(err, "failed to generate HTML report")
	}
	return nil
}

// HTMLReportGenerator generates comprehensive HTML test reports
type HTMLReportGenerator struct {
	SummaryFile    string
	LogFile        string
	SummaryData    *Summary
	LogData        []LogEntry
	ReportDir      string
	SummaryContent string // Raw summary.json content for download
	LogContent     string // Raw hrp.log content for download
	CaseContent    string // Raw case.json content for display
}

// LogEntry represents a single log entry
type LogEntry struct {
	Time     string         `json:"time"`
	Level    string         `json:"level"`
	Message  string         `json:"message"`
	Fields   map[string]any `json:"-"` // Store all other fields
	LogIndex int            `json:"-"` // Original index to maintain order for same timestamps
}

// NewHTMLReportGenerator creates a new HTML report generator
func NewHTMLReportGenerator(summaryFile, logFile string) (*HTMLReportGenerator, error) {
	generator := &HTMLReportGenerator{
		SummaryFile: summaryFile,
		LogFile:     logFile,
		ReportDir:   filepath.Dir(summaryFile),
	}

	// Load summary data
	if err := generator.loadSummaryData(); err != nil {
		return nil, fmt.Errorf("failed to load summary data: %w", err)
	}

	// Load log data if provided
	if logFile != "" {
		if err := generator.loadLogData(); err != nil {
			log.Warn().Err(err).Msg("failed to load log data, continuing without logs")
		}
	}

	// Load case.json data if exists
	if err := generator.loadCaseData(); err != nil {
		log.Warn().Err(err).Msg("failed to load case data, continuing without case display")
	}

	return generator, nil
}

// loadSummaryData loads test summary data from JSON file
func (g *HTMLReportGenerator) loadSummaryData() error {
	data, err := os.ReadFile(g.SummaryFile)
	if err != nil {
		return err
	}

	// Parse JSON data first
	g.SummaryData = &Summary{}
	err = json.Unmarshal(data, g.SummaryData)
	if err != nil {
		return err
	}

	// Initialize nil fields to prevent template execution errors
	g.initializeSummaryFields()

	// Re-encode the summary data to ensure proper UTF-8 encoding for download
	// This fixes Chinese character encoding issues in legacy summary.json files
	buffer := new(strings.Builder)
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "    ")

	err = encoder.Encode(g.SummaryData)
	if err != nil {
		// Fallback to original content if re-encoding fails
		g.SummaryContent = string(data)
		return nil
	}

	// Store the properly encoded content for download
	g.SummaryContent = strings.TrimSpace(buffer.String())

	return nil
}

// initializeSummaryFields initializes nil fields in SummaryData to prevent template execution errors
func (g *HTMLReportGenerator) initializeSummaryFields() {
	if g.SummaryData == nil {
		g.SummaryData = &Summary{}
	}

	// Initialize Stat if nil
	if g.SummaryData.Stat == nil {
		g.SummaryData.Stat = &Stat{}
		// Initialize TestSteps.Actions map if needed
		if g.SummaryData.Stat.TestSteps.Actions == nil {
			g.SummaryData.Stat.TestSteps.Actions = make(map[option.ActionName]int)
		}
	}

	// Initialize Platform if nil
	if g.SummaryData.Platform == nil {
		g.SummaryData.Platform = &Platform{}
	}

	// Initialize Time if nil
	if g.SummaryData.Time == nil {
		g.SummaryData.Time = &TestCaseTime{}
	}

	// Initialize Details if nil
	if g.SummaryData.Details == nil {
		g.SummaryData.Details = []*TestCaseSummary{}
	}
}

// loadLogData loads test log data from log file
func (g *HTMLReportGenerator) loadLogData() error {
	if g.LogFile == "" || !builtin.FileExists(g.LogFile) {
		return nil
	}

	// Read raw log content for download
	logData, err := os.ReadFile(g.LogFile)
	if err != nil {
		return err
	}
	g.LogContent = string(logData)

	file, err := os.Open(g.LogFile)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	logIndex := 0 // Track original order
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// First parse into a generic map to get all fields
		var rawEntry map[string]any
		if err := json.Unmarshal([]byte(line), &rawEntry); err != nil {
			// Skip invalid JSON lines
			continue
		}

		// Create LogEntry with basic fields
		logEntry := LogEntry{
			Fields:   make(map[string]any),
			LogIndex: logIndex, // Store original order
		}
		logIndex++

		// Extract standard fields
		if time, ok := rawEntry["time"].(string); ok {
			logEntry.Time = time
		}
		if level, ok := rawEntry["level"].(string); ok {
			logEntry.Level = level
		}
		if message, ok := rawEntry["message"].(string); ok {
			logEntry.Message = message
		}

		// Store all other fields in Fields map
		for key, value := range rawEntry {
			if key != "time" && key != "level" && key != "message" {
				logEntry.Fields[key] = value
			}
		}

		g.LogData = append(g.LogData, logEntry)
	}

	return scanner.Err()
}

// loadCaseData loads test case data from case.json file
func (g *HTMLReportGenerator) loadCaseData() error {
	caseFile := filepath.Join(g.ReportDir, "case.json")
	if !builtin.FileExists(caseFile) {
		return nil // case.json is optional
	}

	data, err := os.ReadFile(caseFile)
	if err != nil {
		return err
	}

	// Store the case content for display
	g.CaseContent = string(data)
	return nil
}

// getStepLogs filters log entries for a specific test step using prefix matching and time range filtering
func (g *HTMLReportGenerator) getStepLogs(stepName string, startTime int64, elapsed int64) []LogEntry {
	if len(g.LogData) == 0 {
		return nil
	}

	var stepLogs []LogEntry
	var inCurrentStep bool = false

	// Calculate step end time (startTime + elapsed, both in milliseconds)
	endTime := startTime + elapsed

	// Convert step times to time.Time for comparison
	// The startTime from step result is in milliseconds timestamp
	stepStartTime := time.UnixMilli(startTime)
	stepEndTime := time.UnixMilli(endTime)

	// Use step start/end markers with prefix matching for precise boundaries
	for _, logEntry := range g.LogData {
		// Parse log entry timestamp for time range validation
		logTime, timeParseErr := g.parseLogTime(logEntry.Time)

		// Check for step boundaries to control inclusion
		if logEntry.Message == RUN_STEP_START {
			if stepFieldValue, exists := logEntry.Fields["step"].(string); exists {
				// use prefix matching for parameterized steps
				if strings.HasPrefix(stepName, stepFieldValue) {
					inCurrentStep = true
					stepLogs = append(stepLogs, logEntry)
					continue
				} else if inCurrentStep {
					// This is a different step starting, we're done with current step
					break
				}
			}
		}

		if logEntry.Message == RUN_STEP_END {
			if stepFieldValue, exists := logEntry.Fields["step"].(string); exists {
				// use prefix matching for parameterized steps
				if strings.HasPrefix(stepName, stepFieldValue) {
					stepLogs = append(stepLogs, logEntry)
					inCurrentStep = false
					break // End of current step, stop processing
				}
			}
		}

		// Include logs when we're in the current step AND within the time range
		if inCurrentStep {
			// Apply time range filtering if time parsing succeeded
			if timeParseErr == nil {
				// Only include logs within the step time range
				if (logTime.Equal(stepStartTime) || logTime.After(stepStartTime)) &&
					(logTime.Equal(stepEndTime) || logTime.Before(stepEndTime)) {
					stepLogs = append(stepLogs, logEntry)
				}
			} else {
				// If time parsing failed, include all logs in the step boundary
				stepLogs = append(stepLogs, logEntry)
			}
		}
	}

	// Sort logs by original index to maintain chronological order
	sort.Slice(stepLogs, func(i, j int) bool {
		return stepLogs[i].LogIndex < stepLogs[j].LogIndex
	})

	return stepLogs
}

// parseLogTime parses various time formats from log entries
func (g *HTMLReportGenerator) parseLogTime(timeStr string) (time.Time, error) {
	// Handle different time formats that might appear in logs
	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.000Z07:00",
		"2006-01-02T15:04:05.000+08:00",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05+08:00",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05Z",
	}

	// Replace common timezone formats
	timeStr = strings.ReplaceAll(timeStr, "Z", "+00:00")
	timeStr = strings.ReplaceAll(timeStr, "+0800", "+08:00")

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time: %s", timeStr)
}

// encodeImageToBase64 encodes an image file to base64 string with compression
func (g *HTMLReportGenerator) encodeImageToBase64(imagePath string) string {
	// Convert relative path to absolute path
	if !filepath.IsAbs(imagePath) {
		imagePath = filepath.Join(g.ReportDir, imagePath)
	}

	if !builtin.FileExists(imagePath) {
		log.Warn().Str("path", imagePath).Msg("image file not found")
		return ""
	}

	// Read and compress the image using the unified compression function
	// Enable resize with max width 800px for HTML reports
	compressedData, err := uixt.CompressImageFile(imagePath, true, 800)
	if err != nil {
		log.Warn().Err(err).Str("path", imagePath).Msg("failed to compress image, using original")
		// Fallback to original image if compression fails
		data, readErr := os.ReadFile(imagePath)
		if readErr != nil {
			log.Warn().Err(readErr).Str("path", imagePath).Msg("failed to read image file")
			return ""
		}
		return base64.StdEncoding.EncodeToString(data)
	}

	return base64.StdEncoding.EncodeToString(compressedData)
}

// formatDuration formats duration from milliseconds to human readable format
func (g *HTMLReportGenerator) formatDuration(duration any) string {
	var durationMs float64

	switch v := duration.(type) {
	case int64:
		durationMs = float64(v)
	case float64:
		durationMs = v
	case int:
		durationMs = float64(v)
	default:
		return "0ms"
	}

	if durationMs < 1000 {
		return fmt.Sprintf("%.0fms", durationMs)
	} else if durationMs < 60000 {
		return fmt.Sprintf("%.1fs", durationMs/1000)
	} else {
		minutes := int(durationMs / 60000)
		seconds := (durationMs - float64(minutes*60000)) / 1000
		return fmt.Sprintf("%dm %.1fs", minutes, seconds)
	}
}

// getStepLogsForTemplate is a template function to get filtered logs for a step
func (g *HTMLReportGenerator) getStepLogsForTemplate(step *StepResult) []LogEntry {
	if step == nil {
		return nil
	}
	return g.getStepLogs(step.Name, step.StartTime, step.Elapsed)
}

// calculateTotalActions calculates the total number of actions across all test cases
func (g *HTMLReportGenerator) calculateTotalActions() int {
	return g.iterateTestData(func(action *ActionResult) int {
		return 1 // Count each action
	})
}

// calculateTotalSubActions calculates the total number of sub-actions across all test cases
func (g *HTMLReportGenerator) calculateTotalSubActions() int {
	return g.iterateTestData(func(action *ActionResult) int {
		total := 0
		// Count sub-actions from start_to_goal results
		if action.Plannings != nil {
			for _, planning := range action.Plannings {
				if planning.SubActions != nil {
					total += len(planning.SubActions)
				}
			}
		} else {
			// Count other actions
			total += 1
		}
		return total
	})
}

// calculateTotalPlannings calculates the total number of planning results across all test cases
func (g *HTMLReportGenerator) calculateTotalPlannings() int {
	return g.iterateTestData(func(action *ActionResult) int {
		if action.Plannings != nil {
			return len(action.Plannings)
		}
		return 0
	})
}

// calculateTotalUsage calculates the total token usage across all test cases
func (g *HTMLReportGenerator) calculateTotalUsage() map[string]interface{} {
	totalUsage := map[string]interface{}{
		"prompt_tokens":     0,
		"completion_tokens": 0,
		"total_tokens":      0,
	}

	if g.SummaryData == nil || g.SummaryData.Details == nil {
		return totalUsage
	}

	for _, testCase := range g.SummaryData.Details {
		if testCase.Records == nil {
			continue
		}
		for _, step := range testCase.Records {
			if step.Actions == nil {
				continue
			}
			for _, action := range step.Actions {
				// Calculate planning usage
				if action.Plannings != nil {
					for _, planning := range action.Plannings {
						if planning.Usage != nil {
							totalUsage["prompt_tokens"] = totalUsage["prompt_tokens"].(int) + planning.Usage.PromptTokens
							totalUsage["completion_tokens"] = totalUsage["completion_tokens"].(int) + planning.Usage.CompletionTokens
							totalUsage["total_tokens"] = totalUsage["total_tokens"].(int) + planning.Usage.TotalTokens
						}
					}
				}

				// Calculate AI operations usage (ai_query, ai_action, ai_assert)
				if action.AIResult != nil {
					var usage *map[string]interface{}

					switch action.AIResult.Type {
					case "query":
						if action.AIResult.QueryResult != nil && action.AIResult.QueryResult.Usage != nil {
							usage = &map[string]interface{}{
								"prompt_tokens":     action.AIResult.QueryResult.Usage.PromptTokens,
								"completion_tokens": action.AIResult.QueryResult.Usage.CompletionTokens,
								"total_tokens":      action.AIResult.QueryResult.Usage.TotalTokens,
							}
						}
					case "action":
						if action.AIResult.PlanningResult != nil && action.AIResult.PlanningResult.Usage != nil {
							usage = &map[string]interface{}{
								"prompt_tokens":     action.AIResult.PlanningResult.Usage.PromptTokens,
								"completion_tokens": action.AIResult.PlanningResult.Usage.CompletionTokens,
								"total_tokens":      action.AIResult.PlanningResult.Usage.TotalTokens,
							}
						}
					case "assert":
						if action.AIResult.AssertionResult != nil && action.AIResult.AssertionResult.Usage != nil {
							usage = &map[string]interface{}{
								"prompt_tokens":     action.AIResult.AssertionResult.Usage.PromptTokens,
								"completion_tokens": action.AIResult.AssertionResult.Usage.CompletionTokens,
								"total_tokens":      action.AIResult.AssertionResult.Usage.TotalTokens,
							}
						}
					}

					if usage != nil {
						totalUsage["prompt_tokens"] = totalUsage["prompt_tokens"].(int) + (*usage)["prompt_tokens"].(int)
						totalUsage["completion_tokens"] = totalUsage["completion_tokens"].(int) + (*usage)["completion_tokens"].(int)
						totalUsage["total_tokens"] = totalUsage["total_tokens"].(int) + (*usage)["total_tokens"].(int)
					}
				}
			}
		}
	}
	return totalUsage
}

// iterateTestData is a helper function that iterates through all actions and applies a counting function
func (g *HTMLReportGenerator) iterateTestData(countFunc func(*ActionResult) int) int {
	total := 0
	if g.SummaryData == nil || g.SummaryData.Details == nil {
		return total
	}

	for _, testCase := range g.SummaryData.Details {
		if testCase.Records == nil {
			continue
		}
		for _, step := range testCase.Records {
			if step.Actions != nil {
				for _, action := range step.Actions {
					total += countFunc(action)
				}
			}
		}
	}
	return total
}

// GenerateReport generates the complete HTML test report
func (g *HTMLReportGenerator) GenerateReport(outputFile string) error {
	if outputFile == "" {
		outputFile = filepath.Join(g.ReportDir, "report.html")
	}

	// Create template functions
	funcMap := template.FuncMap{
		"formatDuration":           g.formatDuration,
		"encodeImageBase64":        g.encodeImageToBase64,
		"getStepLogs":              g.getStepLogsForTemplate,
		"calculateTotalActions":    g.calculateTotalActions,
		"calculateTotalSubActions": g.calculateTotalSubActions,
		"calculateTotalPlannings":  g.calculateTotalPlannings,
		"calculateTotalUsage":      g.calculateTotalUsage,
		"getSummaryContentBase64": func() string {
			return base64.StdEncoding.EncodeToString([]byte(g.SummaryContent))
		},
		"getLogContentBase64": func() string {
			return base64.StdEncoding.EncodeToString([]byte(g.LogContent))
		},
		"getCaseContentBase64": func() string {
			return base64.StdEncoding.EncodeToString([]byte(g.CaseContent))
		},
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"toJSONFormatted": func(v any) string {
			var buf strings.Builder
			encoder := json.NewEncoder(&buf)
			encoder.SetEscapeHTML(false)
			encoder.SetIndent("", "    ")
			_ = encoder.Encode(v)
			result := strings.TrimSpace(buf.String())
			return result
		},
		"add":   func(a, b int) int { return a + b },
		"base":  filepath.Base,
		"index": func(m map[string]any, key string) any { return m[key] },
		"title": func(s string) string {
			if s == "" {
				return ""
			}
			return strings.ToUpper(s[:1]) + s[1:]
		},
		"extractThought": func(content string) string {
			if content == "" {
				return ""
			}
			// Try to parse as JSON to extract thought field
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(content), &data); err == nil {
				if thought, ok := data["thought"].(string); ok && thought != "" {
					return thought
				}
			}
			// If not JSON or no thought field, return original content
			return content
		},
		"formatBodyContent": func(content string) string {
			// Try to parse as JSON to format
			var data interface{}
			if err := json.Unmarshal([]byte(content), &data); err == nil {
				var buf strings.Builder
				encoder := json.NewEncoder(&buf)
				encoder.SetEscapeHTML(false)
				encoder.SetIndent("", "    ")
				_ = encoder.Encode(data)
				return strings.TrimSpace(buf.String())
			}
			// If not JSON, return original content
			return content
		},
	}

	// Parse template
	tmpl, err := template.New("report").Funcs(funcMap).Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Create output file with explicit UTF-8 handling
	file, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Execute template (Go's html/template ensures UTF-8 encoding)
	if err := tmpl.Execute(file, g.SummaryData); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Ensure data is flushed to disk
	if err := file.Sync(); err != nil {
		return fmt.Errorf("failed to sync HTML report file: %w", err)
	}

	log.Info().Str("path", outputFile).Msg("HTML report generated successfully")
	return nil
}

// htmlTemplate contains the complete HTML template for test reports
const htmlTemplate = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>HttpRunner Test Report</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            line-height: 1.6;
            color: #333;
            background-color: #f5f5f5;
        }

        .container {
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
        }

        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 30px;
            border-radius: 10px;
            margin-bottom: 30px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
        }

        .header-content {
            display: flex;
            justify-content: space-between;
            align-items: center;
        }

        .header-left h1 {
            font-size: 2.5em;
            margin-bottom: 10px;
        }

        .header-left .subtitle {
            font-size: 1.2em;
            opacity: 0.9;
        }

        .header-right {
            text-align: right;
            display: flex;
            flex-direction: column;
            align-items: flex-end;
            gap: 15px;
        }

        .download-section {
            background: rgba(255, 255, 255, 0.2);
            padding: 15px 20px;
            border-radius: 8px;
            backdrop-filter: blur(10px);
            min-width: 240px;
            text-align: center;
        }

        .download-title {
            font-size: 0.9em;
            font-weight: 600;
            margin-bottom: 10px;
            opacity: 0.9;
        }

        .download-buttons {
            display: flex;
            gap: 10px;
            width: 100%;
        }

        .download-btn {
            background: rgba(255, 255, 255, 0.2);
            color: white;
            border: 2px solid rgba(255, 255, 255, 0.3);
            padding: 8px 12px;
            border-radius: 6px;
            cursor: pointer;
            font-size: 0.85em;
            font-weight: 500;
            transition: all 0.3s ease;
            backdrop-filter: blur(10px);
            text-decoration: none;
            display: flex;
            align-items: center;
            justify-content: center;
            gap: 6px;
            flex: 1;
            text-align: center;
        }

        .download-btn:hover {
            background: rgba(255, 255, 255, 0.3);
            border-color: rgba(255, 255, 255, 0.5);
            transform: translateY(-2px);
            box-shadow: 0 4px 8px rgba(0,0,0,0.2);
        }

        .download-btn:active {
            transform: translateY(0);
        }

        .summary {
            background: white;
            padding: 25px;
            border-radius: 10px;
            margin-bottom: 30px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }

        .summary h2 {
            color: #2c3e50;
            margin: 0;
            padding: 0;
            border: none;
        }

        .summary-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-bottom: 20px;
        }

        .summary-title-bar {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 20px;
            border-bottom: 2px solid #3498db;
            padding-bottom: 10px;
        }

        .toggle-all-btn {
            background-color: #ffc107;
            color: #212529;
            border: none;
            padding: 8px 16px;
            border-radius: 5px;
            cursor: pointer;
            font-size: 0.9em;
            font-weight: 500;
            transition: background-color 0.2s ease;
            flex-shrink: 0;
        }

        .toggle-all-btn:hover {
            background-color: #e0a800;
        }

        .summary-item {
            text-align: center;
            padding: 15px;
            border-radius: 8px;
            background: #f8f9fa;
        }

        .summary-item.success {
            background: #d4edda;
            border: 1px solid #c3e6cb;
        }

        .summary-item.failure {
            background: #f8d7da;
            border: 1px solid #f5c6cb;
        }

        .summary-item .value {
            font-size: 2em;
            font-weight: bold;
            color: #2c3e50;
        }

        .summary-item .label {
            color: #6c757d;
            margin-top: 5px;
        }

        .platform-info {
            background: #e9ecef;
            padding: 20px;
            border-radius: 8px;
            margin-top: 20px;
        }

        .platform-info h3 {
            margin-bottom: 15px;
            color: #495057;
        }

        .platform-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 15px;
        }

        .platform-item {
            background: white;
            padding: 15px;
            border-radius: 8px;
            text-align: center;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            transition: transform 0.2s, box-shadow 0.2s;
        }

        .platform-item:hover {
            transform: translateY(-2px);
            box-shadow: 0 4px 8px rgba(0,0,0,0.15);
        }

        .platform-label {
            font-size: 1.0em;
            color: #6c757d;
            margin-bottom: 8px;
            font-weight: 500;
        }

        .platform-value {
            font-size: 0.9em;
            font-weight: bold;
            color: #2c3e50;
            word-break: break-all;
        }

        .test-cases {
            margin-top: 20px;
        }

        .test-case {
            background: white;
            margin-bottom: 40px;
            border-radius: 15px;
            box-shadow: 0 6px 12px rgba(0,0,0,0.1);
            overflow: hidden;
            border: 2px solid #e9ecef;
            padding-bottom: 8px;
        }

        .test-case h2 {
            background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%);
            color: #495057;
            margin: 0;
            padding: 20px 30px;
            font-size: 1.5em;
            font-weight: 600;
            border-bottom: 2px solid #dee2e6;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }

        .case-info {
            display: flex;
            align-items: center;
            gap: 15px;
        }

        .step-container {
            background: white;
            margin-bottom: 8px;
            border-radius: 12px;
            box-shadow: 0 4px 8px rgba(0,0,0,0.1);
            overflow: hidden;
            border: 1px solid #e9ecef;
        }

        .step-container:first-of-type {
            margin-top: 8px;
        }

        .step-header {
            background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%);
            padding: 25px 30px;
            cursor: pointer;
            border-bottom: 2px solid #dee2e6;
            transition: all 0.3s ease;
        }

        .step-header:hover {
            background: linear-gradient(135deg, #e9ecef 0%, #dee2e6 100%);
            transform: translateY(-1px);
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }

        .step-header h3 {
            display: flex;
            align-items: center;
            gap: 20px;
            margin: 0;
            font-size: 1.3em;
            font-weight: 500;
        }

        .step-info-group {
            margin-left: auto;
            display: flex;
            align-items: center;
            gap: 12px;
            min-width: 300px;
            justify-content: flex-end;
        }

        .step-status {
            min-width: 70px;
            text-align: center;
        }

        .step-duration {
            min-width: 80px;
            text-align: center;
        }

        .step-type-fixed {
            min-width: 120px;
            text-align: center;
        }

        .step-number {
            background: linear-gradient(135deg, #007bff 0%, #0056b3 100%);
            color: white;
            width: 36px;
            height: 36px;
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 1.0em;
            font-weight: bold;
            box-shadow: 0 2px 4px rgba(0, 123, 255, 0.3);
        }

        .status-badge {
            padding: 6px 14px;
            border-radius: 20px;
            font-size: 0.85em;
            font-weight: bold;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }

        .status-badge.success {
            background: linear-gradient(135deg, #28a745 0%, #20c997 100%);
            color: white;
        }

        .status-badge.failure {
            background: linear-gradient(135deg, #dc3545 0%, #c82333 100%);
            color: white;
        }

        .duration {
            background: linear-gradient(135deg, #6c757d 0%, #5a6268 100%);
            color: white;
            padding: 4px 10px;
            border-radius: 12px;
            font-size: 0.8em;
            box-shadow: 0 1px 3px rgba(0,0,0,0.2);
        }

        .toggle-icon {
            margin-left: auto;
            font-size: 0.8em;
            transition: transform 0.3s;
        }

        .toggle-icon.rotated {
            transform: rotate(-90deg);
        }

        .step-type {
            background: linear-gradient(135deg, #17a2b8 0%, #138496 100%);
            color: white;
            padding: 3px 10px;
            border-radius: 12px;
            font-size: 0.8em;
            box-shadow: 0 1px 3px rgba(0,0,0,0.2);
        }

        .step-content {
            padding: 30px;
            display: none;
            background: #fafbfc;
            border-top: 1px solid #e9ecef;
        }

        .step-content.show {
            display: block;
        }

        .actions-section, .validators-section, .screenshots-section, .logs-section {
            margin-bottom: 30px;
        }

        .actions-section h4, .validators-section h4, .screenshots-section h4, .logs-section h4 {
            color: #495057;
            margin-bottom: 20px;
            padding-bottom: 12px;
            border-bottom: 2px solid #dee2e6;
            font-size: 1.2em;
            font-weight: 600;
        }

        .action-item {
            background: white;
            border: 2px solid #e9ecef;
            border-radius: 12px;
            padding: 20px;
            margin-bottom: 20px;
            box-shadow: 0 2px 6px rgba(0,0,0,0.08);
            transition: all 0.3s ease;
        }

        .action-item:hover {
            border-color: #007bff;
            box-shadow: 0 4px 12px rgba(0, 123, 255, 0.15);
            transform: translateY(-1px);
        }

        .action-header {
            display: flex;
            align-items: center;
            gap: 18px;
            margin-bottom: 15px;
            padding: 12px 15px;
            border-radius: 8px;
            background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%);
            border: 1px solid #dee2e6;
        }

        .action-header strong {
            color: #007bff;
            font-size: 1.1em;
            font-weight: 600;
        }

        .action-description {
            color: #6c757d;
            font-style: italic;
            margin: 10px 0;
            padding: 10px 15px;
            white-space: pre-wrap;
            word-wrap: break-word;
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
            background: #f8f9fa;
            border: 1px solid #dee2e6;
            border-radius: 4px;
            font-size: 0.9em;
            line-height: 1.4;
        }

        .action-content {
            display: block;
        }

        .error {
            color: #dc3545;
            font-weight: bold;
        }

        .planning-results {
            margin-top: 15px;
            padding: 15px;
            background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%);
            border: 1px solid #dee2e6;
            border-radius: 12px;
            box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
        }

        .planning-item {
            background: white;
            border: 1px solid #dee2e6;
            border-radius: 8px;
            padding: 15px;
            margin-bottom: 15px;
        }

        .planning-item:last-child {
            margin-bottom: 0;
        }

        .planning-header {
            display: flex;
            align-items: center;
            gap: 15px;
            margin-bottom: 15px;
            padding-bottom: 10px;
            border-bottom: 1px solid #dee2e6;
        }

        .planning-label {
            background: #007bff;
            color: white;
            padding: 4px 12px;
            border-radius: 15px;
            font-size: 0.9em;
            font-weight: bold;
        }

        .planning-three-columns {
            display: flex;
            gap: 20px;
            margin: 15px 0;
        }

        .planning-column-screenshot {
            flex: 0.9;
            min-width: 250px;
            max-width: 35%;
        }

        .planning-column-right-container {
            flex: 1.6;
            min-width: 350px;
            display: flex;
            flex-direction: column;
            gap: 15px;
        }

        .planning-column-model, .planning-column-actions {
            flex: 1;
            min-width: 0;
        }

        .planning-step-compact {
            background: white;
            border: 1px solid #dee2e6;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            height: fit-content;
        }

        .planning-column-screenshot .planning-step-compact {
            height: auto;
        }

        .step-header-compact {
            background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%);
            padding: 10px 12px;
            border-bottom: 1px solid #dee2e6;
            border-radius: 8px 8px 0 0;
            display: flex;
            align-items: center;
            justify-content: space-between;
        }

        .step-name {
            font-weight: 600;
            color: #495057;
            font-size: 0.9em;
        }

        .screenshot-display {
            padding: 12px;
        }

        .screenshot-item-compact {
            text-align: center;
        }

        .screenshot-item-compact .screenshot-image {
            padding: 15px;
            background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%);
            border-radius: 6px;
            display: flex;
            justify-content: center;
            align-items: center;
            position: relative;
            overflow: hidden;
        }

        .screenshot-item-compact .screenshot-image img {
            width: 100%;
            height: auto;
            max-height: 500px;
            border-radius: 4px;
            cursor: pointer;
            transition: transform 0.2s;
            object-fit: contain;
            box-shadow: 0 2px 6px rgba(0,0,0,0.1);
            display: block;
        }

        /* Handle very tall screenshots */
        .screenshot-item-compact .screenshot-image img[style*="height"] {
            max-height: 400px;
            width: auto;
            max-width: 100%;
        }

        .screenshot-item-compact .screenshot-image img:hover {
            transform: scale(1.02);
        }

        /* Horizontal scrolling screenshot styles */
        .screenshot-horizontal-scroll {
            display: flex;
            gap: 0 !important;
            overflow-x: auto;
            overflow-y: hidden;
            padding: 8px;
            scroll-behavior: smooth;
            -webkit-overflow-scrolling: touch;
            background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%);
            border: 1px solid #dee2e6;
            border-radius: 6px;
            align-items: center;
            justify-content: center;
            line-height: 0;
            font-size: 0;
        }

        .screenshot-horizontal-scroll::-webkit-scrollbar {
            height: 8px;
        }

        .screenshot-horizontal-scroll::-webkit-scrollbar-track {
            background: #f1f1f1;
            border-radius: 4px;
        }

        .screenshot-horizontal-scroll::-webkit-scrollbar-thumb {
            background: #888;
            border-radius: 4px;
        }

        .screenshot-horizontal-scroll::-webkit-scrollbar-thumb:hover {
            background: #555;
        }

        .screenshot-item-horizontal {
            flex: 0 0 auto;
            min-width: 180px;
            max-width: 280px;
            text-align: center;
            margin: 0 !important;
            padding: 0 !important;
            border: none !important;
            outline: none;
            line-height: 0;
        }

        .screenshot-item-horizontal .screenshot-image {
            padding: 0;
            margin: 0;
            background: transparent;
            border-radius: 0;
            display: flex;
            justify-content: center;
            align-items: center;
            position: relative;
            overflow: hidden;
            height: 350px;
            border: none;
        }

        .screenshot-item-horizontal .screenshot-image img {
            max-width: 100%;
            max-height: 100%;
            border-radius: 0;
            cursor: pointer;
            transition: transform 0.2s;
            object-fit: contain;
            box-shadow: none;
            display: block;
            margin: 0 !important;
            padding: 0 !important;
            border: none !important;
            vertical-align: top;
            float: left;
            outline: none;
        }

        .screenshot-item-horizontal .screenshot-image img:hover {
            transform: scale(1.05);
        }

        /* Direct inline screenshot styles */
        .screenshot-inline {
            max-height: 350px;
            object-fit: contain;
            cursor: pointer;
            transition: transform 0.2s;
            display: inline-block;
            margin: 0 4px 0 0 !important;
            padding: 0 !important;
            border: none !important;
            border-radius: 0 !important;
            box-shadow: none !important;
            vertical-align: top;
            outline: none;
        }

        .screenshot-inline:last-child {
            margin-right: 0 !important;
        }

        .screenshot-inline:hover {
            transform: scale(1.05);
        }

        .actions-details {
            padding: 12px;
            max-height: 300px;
            overflow-y: auto;
        }

        .action-detail-item {
            background: #f8f9fa;
            border: 1px solid #e9ecef;
            border-radius: 6px;
            padding: 8px;
            margin: 6px 0;
            font-size: 0.85em;
        }

        .action-detail-header {
            display: flex;
            align-items: center;
            gap: 8px;
            margin-bottom: 6px;
        }

        .action-detail-header .action-name {
            background: #6f42c1;
            color: white;
            padding: 2px 6px;
            border-radius: 10px;
            font-size: 0.8em;
            font-weight: bold;
        }

        .action-detail-header .duration {
            background: #6c757d;
            color: white;
            padding: 1px 4px;
            border-radius: 8px;
            font-size: 0.75em;
        }

        .action-detail-header .success {
            color: #28a745;
            font-size: 0.9em;
        }

        .action-detail-header .error {
            color: #dc3545;
            font-size: 0.9em;
        }

        .action-arguments {
            background: #ffffff;
            border: 1px solid #dee2e6;
            border-radius: 4px;
            padding: 4px 6px;
            margin: 4px 0;
            font-family: monospace;
            font-size: 0.8em;
            color: #495057;
            word-break: break-all;
        }

        .action-requests {
            margin-top: 6px;
        }

        .requests-toggle-compact {
            background: #6c757d;
            color: white;
            border: none;
            padding: 4px 8px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 0.8em;
            margin-bottom: 6px;
            transition: background-color 0.3s;
            display: block;
            width: 100%;
            text-align: left;
        }

        .requests-toggle-compact:hover {
            background: #5a6268;
        }

        .requests-content-compact {
            display: none;
        }

        .requests-content-compact.show {
            display: block;
        }

        .request-item-compact {
            background: #ffffff;
            border: 1px solid #e9ecef;
            border-radius: 4px;
            padding: 6px;
            margin: 4px 0;
            font-size: 0.75em;
        }

        .request-header-compact {
            display: flex;
            align-items: center;
            gap: 6px;
            margin-bottom: 4px;
            flex-wrap: wrap;
        }

        .request-header-compact .method {
            background: #007bff;
            color: white;
            padding: 1px 4px;
            border-radius: 3px;
            font-size: 0.7em;
            font-weight: bold;
        }

        .url-compact {
            color: #495057;
            font-family: monospace;
            font-size: 0.7em;
            word-break: break-all;
            flex: 1;
            min-width: 0;
        }

        .request-header-compact .status.success {
            color: #28a745;
            font-weight: bold;
            font-size: 0.7em;
        }

        .request-header-compact .status.failure {
            color: #dc3545;
            font-weight: bold;
            font-size: 0.7em;
        }

        .request-header-compact .duration {
            background: #6c757d;
            color: white;
            padding: 1px 4px;
            border-radius: 8px;
            font-size: 0.7em;
            font-weight: 500;
        }

        .request-body-compact, .response-body-compact {
            background: #f8f9fa;
            border: 1px solid #e9ecef;
            border-radius: 3px;
            padding: 4px;
            margin: 2px 0;
            font-family: monospace;
            font-size: 0.7em;
            max-height: 80px;
            overflow-y: auto;
            word-break: break-all;
            white-space: nowrap;
            overflow-x: auto;
            line-height: 1.3;
        }

        .model-output-compact {
            padding: 12px;
        }

        .model-info, .tool-calls-info, .actions-info, .usage-info {
            background: #f8f9fa;
            border: 1px solid #e9ecef;
            border-radius: 4px;
            padding: 8px 10px;
            margin: 6px 0;
            font-size: 0.85em;
            color: #495057;
        }

        .structured-data {
            background: #f8f9fa;
            border: 1px solid #28a745;
            border-radius: 6px;
            padding: 10px 12px;
            margin: 8px 0;
            font-size: 0.85em;
            color: #495057;
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
            white-space: pre-wrap;
            word-wrap: break-word;
            max-height: 200px;
            overflow-y: auto;
        }

        @media screen and (max-width: 768px) {
            .planning-three-columns {
                flex-direction: column;
                gap: 15px;
            }

            .planning-column-screenshot {
                flex: none;
                min-width: auto;
                max-width: none;
            }

            .planning-column-right-container {
                flex: none;
                min-width: auto;
                gap: 10px;
            }

            .screenshot-item-compact .screenshot-image {
                padding: 10px;
            }

            .screenshot-item-compact .screenshot-image img {
                width: 100%;
                height: auto;
            }
        }

        .action-details {
            display: flex;
            align-items: center;
            gap: 10px;
        }

        .action-details .action-name {
            background: #6f42c1;
            color: white;
            padding: 2px 8px;
            border-radius: 12px;
            font-size: 0.8em;
            font-weight: bold;
        }

        .action-details .action-desc {
            color: #6c757d;
            font-style: italic;
            font-size: 0.9em;
        }

        .thought {
            background: linear-gradient(135deg, #e3f2fd 0%, #f3e5f5 100%);
            border: 2px solid #2196f3;
            border-radius: 12px;
            padding: 15px;
            margin: 10px 0;
            font-style: italic;
            color: #1565c0;
            font-size: 1.0em;
            font-weight: 500;
            box-shadow: 0 2px 8px rgba(33, 150, 243, 0.15);
            display: flex;
            align-items: flex-start;
            gap: 10px;
        }

        .thought::before {
            content: "💭";
            font-size: 1.2em;
            flex-shrink: 0;
            margin-top: 0px;
            line-height: 1;
        }

        .arguments {
            background: #f8f9fa;
            border: 1px solid #dee2e6;
            border-radius: 4px;
            padding: 6px;
            margin: 6px 0;
            font-family: monospace;
            font-size: 0.9em;
        }

        .screenshots-section {
            background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%);
            border: 2px solid #28a745;
            border-radius: 12px;
            padding: 12px;
            box-shadow: 0 4px 12px rgba(40, 167, 69, 0.15);
        }

        .screenshots-section h4 {
            color: #155724;
            margin-bottom: 10px;
            font-size: 1.0em;
            font-weight: 600;
        }

        .screenshots-horizontal {
            display: flex;
            gap: 15px;
            overflow-x: auto;
            padding: 10px 0;
        }

        .screenshots-horizontal .screenshot-item {
            flex: 0 0 auto;
            min-width: 200px;
            max-width: 300px;
            margin-bottom: 0;
        }

        .screenshots-horizontal .screenshot-image {
            min-height: 300px;
            padding: 10px 0;
        }

        .screenshots-horizontal .screenshot-image img {
            max-height: 400px;
            width: auto;
        }

        .screenshot-item {
            background: white;
            border: 1px solid #dee2e6;
            border-radius: 8px;
            padding: 10px;
            margin-bottom: 15px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            transition: transform 0.2s, box-shadow 0.2s;
        }

        .screenshot-item:hover {
            transform: translateY(-2px);
            box-shadow: 0 4px 8px rgba(0,0,0,0.15);
        }

        .screenshot-item.small {
            margin-bottom: 10px;
        }

        .screenshot-info {
            display: flex;
            align-items: center;
            gap: 10px;
            margin-bottom: 8px;
        }

        .filename {
            font-family: monospace;
            font-size: 0.9em;
            color: #495057;
        }

        .resolution {
            background: #6c757d;
            color: white;
            padding: 2px 6px;
            border-radius: 10px;
            font-size: 0.8em;
        }

        .screenshot-image {
            text-align: center;
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 400px;
            padding: 20px 0;
            background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%);
            border-radius: 8px;
            margin: 10px 0;
        }

        .screenshot-image img {
            max-width: 100%;
            max-height: 600px;
            border-radius: 6px;
            cursor: pointer;
            transition: transform 0.2s;
            object-fit: contain;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }

        .screenshot-image img:hover {
            transform: scale(1.02);
        }

        .screenshot-item.small .screenshot-image {
            min-height: 300px;
            padding: 15px 0;
        }

        .screenshot-item.small .screenshot-image img {
            max-height: 350px;
        }

        .validator-item {
            background: #f8f9fa;
            border: 1px solid #dee2e6;
            border-radius: 8px;
            padding: 12px;
            margin-bottom: 10px;
        }

        .validator-item.success {
            border-color: #28a745;
            background: #d4edda;
        }

        .validator-item.failure {
            border-color: #dc3545;
            background: #f8d7da;
        }

        .validator-header {
            display: flex;
            align-items: center;
            gap: 15px;
            margin-bottom: 15px;
            padding: 12px 15px;
            border-radius: 8px;
            background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%);
            border: 1px solid #dee2e6;
        }

        .validator-header strong {
            color: #007bff;
            font-size: 1.1em;
            font-weight: 600;
        }

        .check-type, .assert-type {
            background: #6c757d;
            color: white;
            padding: 2px 8px;
            border-radius: 10px;
            font-size: 0.8em;
        }

        .result {
            font-weight: bold;
        }

        .validator-expect, .validator-message {
            margin: 8px 0;
            font-size: 0.9em;
            padding: 8px 12px;
            background: #f8f9fa;
            border: 1px solid #dee2e6;
            border-radius: 4px;
        }

        .validator-ai-content {
            margin-top: 15px;
            padding: 15px;
            background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%);
            border: 1px solid #dee2e6;
            border-radius: 12px;
            box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
        }

        .validator-ai-layout {
            display: flex;
            gap: 20px;
            margin: 15px 0;
        }

        .validator-column-screenshot {
            flex: 0.9;
            min-width: 250px;
            max-width: 35%;
        }

        .validator-column-analysis {
            flex: 1.6;
            min-width: 350px;
        }

        .validator-step-compact {
            background: white;
            border: 1px solid #dee2e6;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            height: fit-content;
        }

        .validator-ai-details {
            padding: 12px;
        }

        .validator-thought {
            background: linear-gradient(135deg, #e3f2fd 0%, #f3e5f5 100%);
            border: 2px solid #2196f3;
            border-radius: 12px;
            padding: 15px;
            margin: 10px 0;
            font-style: italic;
            color: #1565c0;
            font-size: 1.0em;
            font-weight: 500;
            box-shadow: 0 2px 8px rgba(33, 150, 243, 0.15);
            white-space: pre-wrap;
            word-wrap: break-word;
        }

        @media screen and (max-width: 768px) {
            .validator-ai-layout {
                flex-direction: column;
                gap: 15px;
            }

            .validator-column-screenshot {
                flex: none;
                min-width: auto;
                max-width: none;
            }

            .validator-column-analysis {
                flex: none;
                min-width: auto;
            }
        }

        .logs-section {
            margin-top: 20px;
        }

        .logs-header {
            display: flex;
            align-items: center;
            gap: 10px;
            cursor: pointer;
            padding: 8px;
            background: #f8f9fa;
            border: 1px solid #dee2e6;
            border-radius: 6px;
            transition: background-color 0.3s;
            margin-bottom: 10px;
        }

        .logs-header:hover {
            background: #e9ecef;
        }

        .logs-header h4 {
            margin: 0;
            color: #495057;
        }

        .logs-toggle {
            margin-left: auto;
            font-size: 0.8em;
            color: #6c757d;
            transition: transform 0.3s;
        }

        .logs-toggle.collapsed {
            transform: rotate(-90deg);
        }

        .logs-container {
            max-height: 400px;
            overflow-y: auto;
            border: 1px solid #dee2e6;
            border-radius: 6px;
            background: #f8f9fa;
            display: none;
        }

        .logs-container.show {
            display: block;
        }

        .log-entry {
            border-bottom: 1px solid #e9ecef;
            padding: 4px 8px;
            font-family: monospace;
            font-size: 0.8em;
            line-height: 1.2;
        }

        .log-entry:last-child {
            border-bottom: none;
        }

        .log-entry.debug {
            background: #f8f9fa;
        }

        .log-entry.info {
            background: #d1ecf1;
        }

        .log-entry.warn {
            background: #fff3cd;
        }

        .log-entry.error {
            background: #f8d7da;
        }

        .log-header {
            display: flex;
            align-items: center;
            gap: 10px;
            margin-bottom: 2px;
            flex-wrap: nowrap;
            cursor: pointer;
            transition: background-color 0.2s;
        }

        .log-header:hover {
            background-color: rgba(0,0,0,0.05);
        }

        .log-time {
            color: #6c757d;
            font-size: 0.75em;
            white-space: nowrap;
            min-width: 180px;
        }

        .log-level {
            padding: 1px 4px;
            border-radius: 3px;
            font-size: 0.65em;
            font-weight: bold;
            text-transform: uppercase;
            min-width: 45px;
            text-align: center;
        }

        .log-level.debug {
            background: #6c757d;
            color: white;
        }

        .log-level.info {
            background: #17a2b8;
            color: white;
        }

        .log-level.warn {
            background: #ffc107;
            color: #212529;
        }

        .log-level.error {
            background: #dc3545;
            color: white;
        }

        .log-message {
            color: #495057;
            word-wrap: break-word;
            flex: 1;
            margin-left: 10px;
        }

        .log-toggle {
            color: #6c757d;
            font-size: 0.8em;
            margin-left: auto;
            transition: transform 0.3s;
        }

        .log-toggle.rotated {
            transform: rotate(-90deg);
        }

        .log-fields {
            background: #f8f9fa;
            border-left: 3px solid #dee2e6;
            padding: 2px 6px;
            margin: 2px 0;
            font-size: 0.75em;
            color: #6c757d;
            max-height: 80px;
            overflow-y: auto;
            word-break: break-all;
            transition: max-height 0.3s ease-out;
        }

        .log-fields.collapsed {
            max-height: 0;
            padding: 0 6px;
            margin: 0;
            overflow: hidden;
        }

        /* Modal styles */
        .modal {
            display: none;
            position: fixed;
            z-index: 1000;
            left: 0;
            top: 0;
            width: 100%;
            height: 100%;
            background-color: rgba(0,0,0,0.9);
        }

        .modal-content {
            margin: auto;
            display: block;
            width: 80%;
            max-width: 700px;
            max-height: 80%;
            object-fit: contain;
        }

        .json-modal {
            display: none;
            position: fixed;
            z-index: 1000;
            left: 0;
            top: 0;
            width: 100%;
            height: 100%;
            background-color: rgba(0,0,0,0.8);
            overflow: auto;
        }

        .json-modal-content {
            background-color: #fefefe;
            margin: 2% auto;
            padding: 0;
            border: none;
            border-radius: 12px;
            width: 90%;
            max-width: 1000px;
            max-height: 90%;
            box-shadow: 0 4px 20px rgba(0,0,0,0.3);
            display: flex;
            flex-direction: column;
        }

        .json-modal-header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 20px 30px;
            border-radius: 12px 12px 0 0;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }

        .json-modal-title {
            font-size: 1.5em;
            font-weight: 600;
            margin: 0;
        }

        .json-modal-body {
            padding: 0;
            flex: 1;
            overflow: hidden;
            display: flex;
            flex-direction: column;
        }

        .json-content {
            background: #f8f9fa;
            margin: 0;
            padding: 20px;
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', 'Consolas', monospace;
            font-size: 13px;
            line-height: 1.5;
            color: #333;
            overflow: auto;
            flex: 1;
            white-space: pre;
            border-radius: 0 0 12px 12px;
        }

        /* JSON Syntax Highlighting */
        .json-key {
            color: #0066cc;
            font-weight: bold;
        }

        .json-string {
            color: #22863a;
        }

        .json-number {
            color: #e36209;
        }

        .json-boolean {
            color: #d73a49;
            font-weight: bold;
        }

        .json-null {
            color: #6f42c1;
            font-weight: bold;
        }

        .json-punctuation {
            color: #24292e;
        }

        .json-brace {
            color: #586069;
            font-weight: bold;
        }

        .json-bracket {
            color: #586069;
            font-weight: bold;
        }

        /* Inline JSON highlighting for smaller displays */
        .json-inline .json-key {
            color: #0066cc;
            font-weight: 600;
        }

        .json-inline .json-string {
            color: #22863a;
        }

        .json-inline .json-number {
            color: #e36209;
        }

        .json-inline .json-boolean {
            color: #d73a49;
            font-weight: 600;
        }

        .json-inline .json-null {
            color: #6f42c1;
            font-weight: 600;
        }

        .json-toolbar {
            background: #e9ecef;
            padding: 10px 20px;
            border-top: 1px solid #dee2e6;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }

        .json-toolbar button {
            background: #007bff;
            color: white;
            border: none;
            padding: 6px 12px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 0.9em;
            transition: background-color 0.3s;
        }

        .json-toolbar button:hover {
            background: #0056b3;
        }

        .json-toolbar .copy-status {
            font-size: 0.9em;
            color: #28a745;
            opacity: 0;
            transition: opacity 0.3s;
        }

        .json-toolbar .copy-status.show {
            opacity: 1;
        }

        .close {
            position: absolute;
            top: 15px;
            right: 35px;
            color: #f1f1f1;
            font-size: 40px;
            font-weight: bold;
            transition: 0.3s;
            cursor: pointer;
        }

        .json-close {
            color: white;
            font-size: 30px;
            font-weight: bold;
            cursor: pointer;
            transition: color 0.3s;
            background: none;
            border: none;
            padding: 0;
            line-height: 1;
        }

        .json-close:hover {
            color: #ffcccc;
        }

        .close:hover,
        .close:focus {
            color: #bbb;
            text-decoration: none;
        }

        /* Responsive design */
        @media screen and (max-width: 768px) {
            .container {
                padding: 10px;
            }

            .header-content {
                flex-direction: column;
                align-items: flex-start;
                gap: 20px;
            }

            .header-left h1 {
                font-size: 2em;
            }

            .header-right {
                text-align: center;
                width: 100%;
                flex-direction: column;
                align-items: center;
                gap: 15px;
            }

            .download-section {
                width: 100%;
                text-align: center;
                min-width: auto;
            }

            .download-buttons {
                justify-content: center;
                width: 100%;
            }

            .download-btn {
                padding: 6px 10px;
                font-size: 0.75em;
            }

            .platform-grid {
                grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
                gap: 10px;
            }

            .platform-item {
                padding: 12px;
            }

            .platform-label {
                font-size: 0.8em;
            }

            .platform-value {
                font-size: 1em;
            }

            .summary-grid {
                grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
                gap: 10px;
            }

            .step-header h3 {
                font-size: 1.2em;
                gap: 15px;
                flex-direction: column;
                align-items: flex-start;
            }

            .step-info-group {
                min-width: auto;
                width: 100%;
                justify-content: space-between;
                margin-left: 0;
                margin-top: 8px;
            }

            .step-status {
                min-width: 60px;
            }

            .step-duration {
                min-width: 70px;
            }

            .step-type-fixed {
                min-width: 100px;
            }

            .step-number {
                width: 32px;
                height: 32px;
                font-size: 0.9em;
            }

            .test-case h2 {
                font-size: 1.3em;
                padding: 15px 20px;
                flex-direction: column;
                align-items: flex-start;
                gap: 10px;
            }

            .case-info {
                align-self: flex-end;
            }

            .step-header {
                padding: 20px 25px;
            }

            .step-content {
                padding: 25px 20px;
            }

            .action-header {
                flex-direction: column;
                align-items: flex-start;
                gap: 8px;
            }

            .logs-header {
                flex-direction: column;
                align-items: flex-start;
                gap: 5px;
            }

            .logs-header h4 {
                font-size: 0.9em;
            }

            .request-header {
                flex-direction: column;
                align-items: flex-start;
                gap: 6px;
            }

            .screenshots-grid {
                grid-template-columns: 1fr;
                gap: 10px;
            }

            .screenshots-horizontal {
                flex-direction: column;
                overflow-x: visible;
            }

            .screenshots-horizontal .screenshot-item {
                flex: none;
                min-width: auto;
                max-width: none;
                width: 100%;
            }

            .screenshot-image {
                min-height: 300px;
                padding: 15px 0;
            }

            .screenshot-image img {
                max-height: 400px;
            }

            .screenshot-item.small .screenshot-image {
                min-height: 250px;
                padding: 10px 0;
            }

            .screenshot-item.small .screenshot-image img {
                max-height: 300px;
            }

            .log-header {
                flex-direction: column;
                align-items: flex-start;
                gap: 4px;
            }

            .log-time {
                min-width: auto;
                font-size: 0.7em;
            }

            .log-level {
                min-width: 35px;
                font-size: 0.6em;
            }

            .log-message {
                margin-left: 0;
                font-size: 0.75em;
            }

            .log-fields {
                margin: 2px 0;
                font-size: 0.7em;
            }

            .log-fields.collapsed {
                margin: 0;
            }
        }

        .action-content {
            margin-top: 10px;
        }

        .action-content strong {
            color: #6f42c1;
            display: block;
            margin-bottom: 8px;
            font-size: 0.95em;
        }

        .action-session-data {
            margin-top: 15px;
            padding: 15px;
            background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%);
            border: 1px solid #dee2e6;
            border-radius: 12px;
            box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
        }

        .session-requests {
            margin-bottom: 15px;
        }

        .session-screenshots {
            margin-top: 15px;
        }

        .sub-action-item {
            margin-top: 15px;
            margin-bottom: 15px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="header-content">
                <div class="header-left">
                    <h1>🚀 HttpRunner Test Report</h1>
                    <div class="subtitle">Start Time: {{.Time.StartAt.Format "2006-01-02 15:04:05"}}</div>
                </div>
                <div class="header-right">
                        <div class="download-section">
                            <div class="download-title">📥 View & Download</div>
                            <div class="download-buttons">
                                {{if getCaseContentBase64}}
                                <button class="download-btn" onclick="showCaseJson()">
                                    <span>📋</span>
                                    <span>case.json</span>
                                </button>
                                {{end}}
                                <button class="download-btn" onclick="showSummaryJson()">
                                    <span>📄</span>
                                    <span>summary.json</span>
                                </button>
                                <button class="download-btn" onclick="showLogContent()">
                                    <span>📋</span>
                                    <span>hrp.log</span>
                                </button>
                            </div>
                        </div>
                </div>
            </div>
        </div>

        <div class="summary">
            <div class="summary-title-bar">
                <h2>📊 Test Summary</h2>
                <button id="toggle-all-steps-btn" class="toggle-all-btn" onclick="toggleAllSteps()">Collapse All Steps</button>
            </div>
            <div class="summary-grid">
                <div class="summary-item success">
                    <div class="value">{{.Stat.TestCases.Success}}</div>
                    <div class="label">Passed TestCases</div>
                </div>
                <div class="summary-item failure">
                    <div class="value">{{.Stat.TestCases.Fail}}</div>
                    <div class="label">Failed TestCases</div>
                </div>
                <div class="summary-item">
                    <div class="value">{{.Stat.TestSteps.Total}}</div>
                    <div class="label">Total Steps</div>
                </div>
                <div class="summary-item">
                    <div class="value">{{calculateTotalActions}}</div>
                    <div class="label">Total Actions</div>
                </div>
                <div class="summary-item">
                    <div class="value">{{calculateTotalSubActions}}</div>
                    <div class="label">Total Sub-Actions</div>
                </div>
                <div class="summary-item">
                    <div class="value">{{calculateTotalPlannings}}</div>
                    <div class="label">Total Plannings</div>
                </div>
                <div class="summary-item">
                    <div class="value">{{printf "%.1f" .Time.Duration}}s</div>
                    <div class="label">Duration</div>
                </div>
                {{$usage := calculateTotalUsage}}
                <div class="summary-item">
                    <div class="value">{{index $usage "prompt_tokens"}}</div>
                    <div class="label">Input Tokens</div>
                </div>
                <div class="summary-item">
                    <div class="value">{{index $usage "completion_tokens"}}</div>
                    <div class="label">Output Tokens</div>
                </div>
                <div class="summary-item">
                    <div class="value">{{index $usage "total_tokens"}}</div>
                    <div class="label">Total Tokens</div>
                </div>
            </div>

            <div class="platform-info">
                <h3>🔧 Platform Information</h3>
                <div class="platform-grid">
                    <div class="platform-item">
                        <div class="platform-label">HttpRunner Version</div>
                        <div class="platform-value">{{.Platform.HttprunnerVersion}}</div>
                    </div>
                    <div class="platform-item">
                        <div class="platform-label">Go Version</div>
                        <div class="platform-value">{{.Platform.GoVersion}}</div>
                    </div>
                    <div class="platform-item">
                        <div class="platform-label">Platform</div>
                        <div class="platform-value">{{.Platform.Platform}}</div>
                    </div>
                </div>
            </div>
        </div>

        <div class="test-cases">
            {{range $caseIndex, $testCase := .Details}}
            <div class="test-case">
                <h2>
                    <span>📋 {{$testCase.Name}}</span>
                    <div class="case-info">
                        <span class="status-badge {{if $testCase.Success}}success{{else}}failure{{end}}">
                            {{if $testCase.Success}}✓ PASS{{else}}✗ FAIL{{end}}
                        </span>
                        <span class="duration">{{printf "%.1f" $testCase.Time.Duration}}s</span>
                    </div>
                </h2>

                {{range $stepIndex, $step := $testCase.Records}}
                <div class="step-container">
                    <div class="step-header" onclick="toggleStep({{$stepIndex}})">
                        <h3>
                            <span class="step-number">{{add $stepIndex 1}}</span>
                            {{$step.Name}}
                            <div class="step-info-group">
                                <span class="status-badge step-status {{if $step.Success}}success{{else}}failure{{end}}">
                                    {{if $step.Success}}✓ PASS{{else}}✗ FAIL{{end}}
                                </span>
                                <span class="duration step-duration">{{formatDuration $step.Elapsed}}</span>
                                <span class="step-type step-type-fixed">{{$step.StepType}}</span>
                                <span class="toggle-icon" id="toggle-{{$stepIndex}}">▼</span>
                            </div>
                        </h3>
                    </div>

                    <div class="step-content" id="step-{{$stepIndex}}">
                        <!-- Actions -->
                        {{if $step.Actions}}
                        <div class="actions-section">
                            <h4>Actions</h4>
                            {{range $actionIndex, $action := $step.Actions}}
                            <div class="action-item">
                                <div class="action-header">
                                    <strong>{{$action.Method}}</strong>
                                    <span class="duration">{{formatDuration $action.Elapsed}}</span>
                                    {{if $action.Error}}<span class="error">Error: {{$action.Error}}</span>{{end}}
                                </div>
                                <div class="action-description">{{$action.Params}}</div>
                                <div class="action-content">

                                {{if $action.Plannings}}
                                    {{range $planningIndex, $planning := $action.Plannings}}
                                    <div class="planning-item">
                                        <div class="planning-header">
                                            <span class="planning-label">🧠 Planning & Execution {{add $planningIndex 1}}</span>
                                            <span class="duration">{{formatDuration $planning.Elapsed}}</span>
                                            {{if $planning.Error}}<span class="error">Error: {{$planning.Error}}</span>{{end}}
                                        </div>

                                        {{$extractedThought := extractThought $planning.Content}}
                                        {{if or $planning.Thought $extractedThought}}
                                            <div class="thought">
                                                {{if $planning.Thought}}
                                                    {{$planning.Thought}}
                                                {{else}}
                                                    {{$extractedThought}}
                                                {{end}}
                                            </div>
                                        {{end}}

                                        <!-- Three-column layout: screenshot left, model output and actions right -->
                                        <div class="planning-three-columns">
                                                <!-- Left column: Screenshot (larger) -->
                                                <div class="planning-column-screenshot">
                                                    <div class="planning-step-compact">
                                                        <div class="step-header-compact">
                                                            <span class="step-name">📸 ScreenShots</span>
                                                            <span class="duration">{{formatDuration $planning.ScreenshotElapsed}}</span>
                                                        </div>
                                                        <div class="screenshot-display screenshot-horizontal-scroll">
                                                            {{if $planning.ScreenResult}}
                                                                {{if $planning.ScreenResult.ImagePath}}
                                                                {{$base64Image := encodeImageBase64 $planning.ScreenResult.ImagePath}}
                                                                {{if $base64Image}}
                                                                    <img src="data:image/jpeg;base64,{{$base64Image}}" alt="Planning Screenshot" onclick="openImageModal(this.src)" class="screenshot-inline" />
                                                                {{end}}
                                                                {{end}}
                                                            {{end}}
                                                            {{if $planning.SubActions}}
                                                                {{range $subAction := $planning.SubActions}}
                                                                    {{if $subAction.ScreenResults}}
                                                                    {{range $subScreenshot := $subAction.ScreenResults}}
                                                                    {{if $subScreenshot.ImagePath}}
                                                                        {{$base64Image := encodeImageBase64 $subScreenshot.ImagePath}}
                                                                        {{if $base64Image}}
                                                                            <img src="data:image/jpeg;base64,{{$base64Image}}" alt="Sub-action Screenshot" onclick="openImageModal(this.src)" class="screenshot-inline" />
                                                                        {{end}}
                                                                        {{end}}
                                                                    {{end}}
                                                                    {{end}}
                                                                {{end}}
                                                            {{end}}
                                                        </div>
                                                    </div>
                                                </div>

                                                <!-- Right column: Model Output and Actions -->
                                                <div class="planning-column-right-container">
                                                    <!-- Top right: Model Output -->
                                                    <div class="planning-column-model">
                                                        <div class="planning-step-compact">
                                                            <div class="step-header-compact">
                                                                <span class="step-name">🤖 Call Model & Parse Result</span>
                                                                <span class="duration">{{formatDuration $planning.ModelCallElapsed}}</span>
                                                            </div>
                                                            <div class="model-output-compact">
                                                                {{if $planning.ModelName}}
                                                                <div class="model-info">🤖 Model: {{$planning.ModelName}}</div>
                                                                {{end}}
                                                                {{if $planning.Usage}}
                                                                <div class="usage-info">📊 Tokens: {{$planning.Usage.PromptTokens}} in / {{$planning.Usage.CompletionTokens}} out / {{$planning.Usage.TotalTokens}} total</div>
                                                                {{end}}
                                                                {{if $planning.ToolCallsCount}}
                                                                <div class="tool-calls-info">🔧 Tool Calls: {{$planning.ToolCallsCount}}</div>
                                                                {{end}}
                                                                {{if $planning.ActionNames}}
                                                                <div class="actions-info json-inline">🎯 Actions: {{toJSONFormatted $planning.ActionNames}}</div>
                                                                {{end}}
                                                            </div>
                                                        </div>
                                                    </div>

                                                    <!-- Bottom right: Actions Details -->
                                                    {{if $planning.SubActions}}
                                                    <div class="planning-column-actions">
                                                        <div class="planning-step-compact">
                                                            <div class="step-header-compact">
                                                                <span class="step-name">🎯 Actions ({{len $planning.SubActions}})</span>
                                                            </div>
                                                            <div class="actions-details">
                                                                {{range $subAction := $planning.SubActions}}
                                                                <div class="action-detail-item">
                                                                    <div class="action-detail-header">
                                                                        <span class="action-name">{{$subAction.ActionName}}</span>
                                                                        <span class="duration">{{formatDuration $subAction.Elapsed}}</span>
                                                                        {{if $subAction.Error}}<span class="error">❌</span>{{else}}<span class="success">✅</span>{{end}}
                                                                    </div>
                                                                    {{if $subAction.Arguments}}
                                                                        <div class="action-arguments json-inline">
                                                                            {{toJSONFormatted $subAction.Arguments}}
                                                                        </div>
                                                                    {{end}}
                                                                    {{if $subAction.Requests}}
                                                                    <div class="action-requests">
                                                                        <button class="requests-toggle-compact" onclick="toggleRequestsCompact(this)">
                                                                            📡 {{len $subAction.Requests}} request(s)
                                                                        </button>
                                                                        <div class="requests-content-compact">
                                                                            {{range $request := $subAction.Requests}}
                                                                            <div class="request-item-compact">
                                                                                <div class="request-header-compact">
                                                                                    <span class="method">{{$request.RequestMethod}}</span>
                                                                                    <span class="url-compact">{{$request.RequestUrl}}</span>
                                                                                    <span class="status {{if $request.Success}}success{{else}}failure{{end}}">{{$request.ResponseStatus}}</span>
                                                                                    <span class="duration">{{formatDuration $request.ResponseDuration}}</span>
                                                                                </div>
                                                                                {{if $request.RequestBody}}
                                                                                <div class="request-body-compact">Request: {{$request.RequestBody}}</div>
                                                                                {{end}}
                                                                                {{if $request.ResponseBody}}
                                                                                <div class="response-body-compact">Response: {{$request.ResponseBody}}</div>
                                                                                {{end}}
                                                                            </div>
                                                                            {{end}}
                                                                        </div>
                                                                    </div>
                                                                    {{end}}
                                                                </div>
                                                                {{end}}
                                                            </div>
                                                        </div>
                                                    </div>
                                                    {{end}}
                                                </div>
                                            </div>


                                    </div>
                                    {{end}}
                                {{end}}

                                {{/* Enhanced AI Operations Display - using unified AIResult data structure */}}
                                {{if or (eq $action.Method "ai_query") (eq $action.Method "ai_action") (eq $action.Method "ai_assert")}}
                                {{if $action.AIResult}}
                                <div class="sub-action-item">
                                    <div class="validator-ai-content">
                                        <!-- Display AI Thought from specific result types -->
                                        {{if eq $action.AIResult.Type "query"}}
                                            {{if $action.AIResult.QueryResult.Thought}}
                                                <div class="thought">{{$action.AIResult.QueryResult.Thought}}</div>
                                            {{end}}
                                        {{else if eq $action.AIResult.Type "action"}}
                                            {{if $action.AIResult.PlanningResult.Thought}}
                                                <div class="thought">{{$action.AIResult.PlanningResult.Thought}}</div>
                                            {{end}}
                                        {{else if eq $action.AIResult.Type "assert"}}
                                            {{if $action.AIResult.AssertionResult.Thought}}
                                                <div class="thought">{{$action.AIResult.AssertionResult.Thought}}</div>
                                            {{end}}
                                        {{end}}

                                        <!-- AI Operation Layout: Screenshot left, Analysis right -->
                                        <div class="validator-ai-layout">
                                            <!-- Left column: Screenshot -->
                                            {{if $action.AIResult.ImagePath}}
                                            <div class="validator-column-screenshot">
                                                <div class="validator-step-compact">
                                                    <div class="step-header-compact">
                                                        <span class="step-name">📸 {{title $action.AIResult.Type}} Screenshot</span>
                                                        {{if $action.AIResult.ScreenshotElapsed}}
                                                        <span class="duration">{{formatDuration $action.AIResult.ScreenshotElapsed}}</span>
                                                        {{end}}
                                                    </div>
                                                    <div class="screenshot-display">
                                                        {{$base64Image := encodeImageBase64 $action.AIResult.ImagePath}}
                                                        {{if $base64Image}}
                                                        <div class="screenshot-item-compact">
                                                            <div class="screenshot-image">
                                                                <img src="data:image/jpeg;base64,{{$base64Image}}" alt="AI {{title $action.AIResult.Type}} Screenshot" onclick="openImageModal(this.src)" />
                                                            </div>
                                                        </div>
                                                        {{end}}
                                                    </div>
                                                </div>
                                            </div>
                                            {{end}}

                                            <!-- Right column: AI Analysis -->
                                            <div class="validator-column-analysis">
                                                <div class="validator-step-compact">
                                                    <div class="step-header-compact">
                                                        <span class="step-name">🤖 AI {{title $action.AIResult.Type}} Analysis</span>
                                                        {{if $action.AIResult.ModelCallElapsed}}
                                                        <span class="duration">{{formatDuration $action.AIResult.ModelCallElapsed}}</span>
                                                        {{end}}
                                                    </div>
                                                    <div class="validator-ai-details">
                                                        {{/* Model name and usage from specific result types */}}
                                                        {{if eq $action.AIResult.Type "query"}}
                                                            {{if $action.AIResult.QueryResult.ModelName}}
                                                                <div class="model-info">🤖 Model: {{$action.AIResult.QueryResult.ModelName}}</div>
                                                            {{end}}
                                                            {{if $action.AIResult.QueryResult.Usage}}
                                                                <div class="usage-info">📊 Tokens: {{$action.AIResult.QueryResult.Usage.PromptTokens}} in / {{$action.AIResult.QueryResult.Usage.CompletionTokens}} out / {{$action.AIResult.QueryResult.Usage.TotalTokens}} total</div>
                                                            {{end}}
                                                            {{/* Display structured data for query results */}}
                                                            {{if $action.AIResult.QueryResult.Data}}
                                                                <div class="model-info">📥 Structured Data:</div>
                                                                <div class="structured-data json-inline">{{toJSONFormatted $action.AIResult.QueryResult.Data}}</div>
                                                            {{end}}
                                                        {{else if eq $action.AIResult.Type "action"}}
                                                            {{if $action.AIResult.PlanningResult.ModelName}}
                                                                <div class="model-info">🤖 Model: {{$action.AIResult.PlanningResult.ModelName}}</div>
                                                            {{end}}
                                                            {{if $action.AIResult.PlanningResult.Usage}}
                                                                <div class="usage-info">📊 Tokens: {{$action.AIResult.PlanningResult.Usage.PromptTokens}} in / {{$action.AIResult.PlanningResult.Usage.CompletionTokens}} out / {{$action.AIResult.PlanningResult.Usage.TotalTokens}} total</div>
                                                            {{end}}
                                                        {{else if eq $action.AIResult.Type "assert"}}
                                                            {{if $action.AIResult.AssertionResult.ModelName}}
                                                                <div class="model-info">🤖 Model: {{$action.AIResult.AssertionResult.ModelName}}</div>
                                                            {{end}}
                                                            {{if $action.AIResult.AssertionResult.Usage}}
                                                                <div class="usage-info">📊 Tokens: {{$action.AIResult.AssertionResult.Usage.PromptTokens}} in / {{$action.AIResult.AssertionResult.Usage.CompletionTokens}} out / {{$action.AIResult.AssertionResult.Usage.TotalTokens}} total</div>
                                                            {{end}}
                                                        {{end}}
                                                        {{if $action.AIResult.Resolution}}
                                                        <div class="model-info">📐 Resolution: {{$action.AIResult.Resolution.Width}}x{{$action.AIResult.Resolution.Height}}</div>
                                                        {{end}}
                                                        {{/* Display Content from specific result types */}}
                                                        {{if eq $action.AIResult.Type "query"}}
                                                            {{if $action.AIResult.QueryResult.Content}}
                                                            <div class="model-info">💬 {{title $action.AIResult.Type}} Result: {{$action.AIResult.QueryResult.Content}}</div>
                                                            {{end}}
                                                        {{else if eq $action.AIResult.Type "action"}}
                                                            {{if $action.AIResult.PlanningResult.Content}}
                                                            <div class="model-info">💬 {{title $action.AIResult.Type}} Result: {{$action.AIResult.PlanningResult.Content}}</div>
                                                            {{end}}
                                                        {{else if eq $action.AIResult.Type "assert"}}
                                                            {{if $action.AIResult.AssertionResult.Content}}
                                                            <div class="model-info">💬 {{title $action.AIResult.Type}} Result: {{$action.AIResult.AssertionResult.Content}}</div>
                                                            {{end}}
                                                        {{end}}
                                                    </div>
                                                </div>
                                            </div>
                                        </div>
                                    </div>
                                </div>
                                {{end}}
                                {{end}}

                                {{/* Handle SessionData: display requests and screen results for non-planning actions */}}
                                {{if not $action.Plannings}}
                                    {{if or $action.Requests $action.ScreenResults}}
                                    <div class="action-session-data">
                                        <!-- Display requests if present -->
                                        {{if $action.Requests}}
                                        <div class="session-requests">
                                            <button class="requests-toggle-compact" onclick="toggleRequestsCompact(this)">
                                                📡 {{len $action.Requests}} request(s)
                                            </button>
                                            <div class="requests-content-compact">
                                                {{range $request := $action.Requests}}
                                                <div class="request-item-compact">
                                                    <div class="request-header-compact">
                                                        <span class="method">{{$request.RequestMethod}}</span>
                                                        <span class="url-compact">{{$request.RequestUrl}}</span>
                                                        <span class="status {{if $request.Success}}success{{else}}failure{{end}}">{{$request.ResponseStatus}}</span>
                                                        <span class="duration">{{formatDuration $request.ResponseDuration}}</span>
                                                    </div>
                                                    {{if $request.RequestBody}}
                                                    <div class="request-body-compact">Request: {{$request.RequestBody}}</div>
                                                    {{end}}
                                                    {{if $request.ResponseBody}}
                                                    <div class="response-body-compact">Response: {{$request.ResponseBody}}</div>
                                                    {{end}}
                                                </div>
                                                {{end}}
                                            </div>
                                        </div>
                                        {{end}}

                                        <!-- Display screen results if present -->
                                        {{if $action.ScreenResults}}
                                        <div class="session-screenshots">
                                            <h5 style="margin: 10px 0; color: #495057;">📸 Screen Results ({{len $action.ScreenResults}})</h5>
                                            <div class="screenshots-horizontal">
                                                {{range $screenshot := $action.ScreenResults}}
                                                {{if $screenshot.ImagePath}}
                                                {{$base64Image := encodeImageBase64 $screenshot.ImagePath}}
                                                {{if $base64Image}}
                                                <div class="screenshot-item small">
                                                    <div class="screenshot-info">
                                                        <span class="filename">{{base $screenshot.ImagePath}}</span>
                                                    </div>
                                                    <div class="screenshot-image">
                                                        <img src="data:image/jpeg;base64,{{$base64Image}}" alt="Screenshot" onclick="openImageModal(this.src)" />
                                                    </div>
                                                </div>
                                                {{end}}
                                                {{end}}
                                                {{end}}
                                            </div>
                                        </div>
                                        {{end}}
                                    </div>
                                    {{end}}
                                {{end}}
                                </div>
                            </div>
                            {{end}}
                        </div>
                        {{end}}

                        <!-- Validators -->
                        {{if and $step.Data $step.Data.validators}}
                        <div class="validators-section">
                            <h4>🔍 Validators</h4>
                            {{range $validatorIndex, $validator := $step.Data.validators}}
                            <div class="validator-item {{if eq $validator.check_result "pass"}}success{{else}}failure{{end}}">
                                <div class="validator-header">
                                    <strong>{{$validator.check}} - {{$validator.assert}}</strong>
                                    <span class="status-badge {{if eq $validator.check_result "pass"}}success{{else}}failure{{end}}">
                                        {{if eq $validator.check_result "pass"}}✓ PASS{{else}}✗ FAIL{{end}}
                                    </span>
                                </div>
                                <div class="validator-expect">Expected: {{$validator.expect}}</div>
                                {{if and $validator.msg (ne $validator.check_result "pass")}}
                                    <div class="validator-message">{{$validator.msg}}</div>
                                {{end}}
                            </div>
                            {{end}}
                        </div>
                        {{end}}

                        <!-- ScreenShots -->
                        {{if $step.Attachments}}
                        {{$attachments := $step.Attachments}}
                        {{if eq (printf "%T" $attachments) "map[string]interface {}"}}
                        {{if index $attachments "screen_results"}}
                        <div class="screenshots-section">
                            <h4>Attachment ScreenShots</h4>
                            <div class="screenshots-horizontal">
                                {{range $screenshot := index $attachments "screen_results"}}
                                {{$imagePath := ""}}
                                {{if $screenshot.ImagePath}}
                                    {{$imagePath = $screenshot.ImagePath}}
                                {{else if index $screenshot "image_path"}}
                                    {{$imagePath = index $screenshot "image_path"}}
                                {{end}}
                                {{if $imagePath}}
                                {{$base64Image := encodeImageBase64 $imagePath}}
                                {{if $base64Image}}
                                <div class="screenshot-item">
                                    <div class="screenshot-info">
                                        <span class="filename">{{base $imagePath}}</span>
                                    </div>
                                    <div class="screenshot-image">
                                        <img src="data:image/jpeg;base64,{{$base64Image}}" alt="Screenshot" onclick="openImageModal(this.src)" />
                                    </div>
                                </div>
                                {{end}}
                                {{end}}
                                {{end}}
                            </div>
                        </div>
                        {{end}}
                        {{end}}
                        {{end}}

                        <!-- Step Logs -->
                        {{$stepLogs := getStepLogs $step}}
                        {{if $stepLogs}}
                        <div class="logs-section">
                            <div class="logs-header" onclick="toggleStepLogs({{$stepIndex}})">
                                <h4>📋 Step Logs ({{len $stepLogs}})</h4>
                                <span class="logs-toggle collapsed" id="logs-toggle-{{$stepIndex}}">▶</span>
                            </div>
                            <div class="logs-container" id="logs-container-{{$stepIndex}}">
                                {{range $logEntry := $stepLogs}}
                                 <div class="log-entry {{$logEntry.Level}}">
                                     <div class="log-header" {{if $logEntry.Fields}}onclick="toggleLogFields(this)"{{end}}>
                                         <span class="log-time">{{$logEntry.Time}}</span>
                                         <span class="log-level {{$logEntry.Level}}">{{$logEntry.Level}}</span>
                                         <span class="log-message">{{$logEntry.Message}}</span>
                                         {{if $logEntry.Fields}}
                                         <span class="log-toggle">▼</span>
                                         {{end}}
                                     </div>
                                     {{if $logEntry.Fields}}
                                        <div class="log-fields collapsed json-inline">
                                            {{toJSONFormatted $logEntry.Fields}}
                                        </div>
                                     {{end}}
                                 </div>
                                {{end}}
                            </div>
                        </div>
                        {{end}}
                    </div>
                </div>
                {{end}}
            </div>
            {{end}}
        </div>
    </div>

    <!-- Image Modal -->
    <div id="imageModal" class="modal">
        <span class="close" onclick="closeModal()">&times;</span>
        <img class="modal-content" id="modalImage">
    </div>

    <!-- JSON Case Modal -->
    <div id="jsonModal" class="json-modal">
        <div class="json-modal-content">
            <div class="json-modal-header">
                <h2 class="json-modal-title">📋 Test Case JSON</h2>
                <button class="json-close" onclick="closeJsonModal()">&times;</button>
            </div>
            <div class="json-modal-body">
                <div class="json-toolbar">
                    <div>
                        <button onclick="copyJsonContent()">📋 Copy to Clipboard</button>
                        <button onclick="downloadCaseJson()">📥 Download case.json</button>
                    </div>
                    <span class="copy-status" id="copyStatus">✅ Copied!</span>
                </div>
                <pre class="json-content" id="jsonContent"></pre>
            </div>
        </div>
    </div>

    <!-- Summary JSON Modal -->
    <div id="summaryModal" class="json-modal">
        <div class="json-modal-content">
            <div class="json-modal-header">
                <h2 class="json-modal-title">📄 Summary JSON</h2>
                <button class="json-close" onclick="closeSummaryModal()">&times;</button>
            </div>
            <div class="json-modal-body">
                <div class="json-toolbar">
                    <div>
                        <button onclick="copySummaryContent()">📋 Copy to Clipboard</button>
                        <button onclick="downloadSummary()">📥 Download summary.json</button>
                    </div>
                    <span class="copy-status" id="summaryStatus">✅ Copied!</span>
                </div>
                <pre class="json-content" id="summaryContent"></pre>
            </div>
        </div>
    </div>

    <!-- Log Content Modal -->
    <div id="logModal" class="json-modal">
        <div class="json-modal-content">
            <div class="json-modal-header">
                <h2 class="json-modal-title">📋 Log Content</h2>
                <button class="json-close" onclick="closeLogModal()">&times;</button>
            </div>
            <div class="json-modal-body">
                <div class="json-toolbar">
                    <div>
                        <button onclick="copyLogContent()">📋 Copy to Clipboard</button>
                        <button onclick="downloadLog()">📥 Download hrp.log</button>
                    </div>
                    <span class="copy-status" id="logStatus">✅ Copied!</span>
                </div>
                <pre class="json-content" id="logContentDisplay"></pre>
            </div>
        </div>
    </div>

    <script>
        // Embedded file contents for download (Base64 encoded)
        const summaryContentBase64 = "{{getSummaryContentBase64}}";
        const logContentBase64 = "{{getLogContentBase64}}";
        const caseContentBase64 = "{{getCaseContentBase64}}";

        // Decode Base64 content with proper UTF-8 handling
        function decodeBase64UTF8(base64) {
            if (!base64) return "";
            try {
                // Use TextDecoder for proper UTF-8 decoding
                const binaryString = atob(base64);
                const bytes = new Uint8Array(binaryString.length);
                for (let i = 0; i < binaryString.length; i++) {
                    bytes[i] = binaryString.charCodeAt(i);
                }
                return new TextDecoder('utf-8').decode(bytes);
            } catch (e) {
                console.error('Failed to decode Base64 content:', e);
                return "";
            }
        }

        const summaryContent = decodeBase64UTF8(summaryContentBase64);
        const logContent = decodeBase64UTF8(logContentBase64);
        const caseContent = decodeBase64UTF8(caseContentBase64);

        // Enhanced JSON highlighting with better parsing
        function highlightJSONAdvanced(jsonString) {
            if (!jsonString || typeof jsonString !== 'string') {
                return jsonString;
            }

            let result = '';
            let i = 0;
            let inString = false;
            let inKey = false;
            let escaped = false;

            while (i < jsonString.length) {
                const char = jsonString[i];
                const nextChar = jsonString[i + 1];

                if (escaped) {
                    result += char;
                    escaped = false;
                    i++;
                    continue;
                }

                if (char === '\\' && inString) {
                    escaped = true;
                    result += char;
                    i++;
                    continue;
                }

                if (char === '"') {
                    if (!inString) {
                        // Starting a string
                        inString = true;
                        // Check if this is a key (followed by colon)
                        let j = i + 1;
                        let tempStr = '';
                        let tempEscaped = false;
                        while (j < jsonString.length) {
                            const c = jsonString[j];
                            if (tempEscaped) {
                                tempEscaped = false;
                                j++;
                                continue;
                            }
                            if (c === '\\') {
                                tempEscaped = true;
                                j++;
                                continue;
                            }
                            if (c === '"') {
                                // End of string, check what follows
                                j++;
                                while (j < jsonString.length && /\s/.test(jsonString[j])) j++;
                                if (j < jsonString.length && jsonString[j] === ':') {
                                    inKey = true;
                                }
                                break;
                            }
                            j++;
                        }

                        if (inKey) {
                            result += '<span class="json-key">"';
                        } else {
                            result += '<span class="json-string">"';
                        }
                    } else {
                        // Ending a string
                        inString = false;
                        result += '"</span>';
                        inKey = false;
                    }
                } else if (!inString) {
                    // Handle non-string content
                    if (char === ':') {
                        result += '<span class="json-punctuation">:</span>';
                    } else if (char === ',') {
                        result += '<span class="json-punctuation">,</span>';
                    } else if (char === '{' || char === '}') {
                        result += '<span class="json-brace">' + char + '</span>';
                    } else if (char === '[' || char === ']') {
                        result += '<span class="json-bracket">' + char + '</span>';
                    } else if (/\d/.test(char) || (char === '-' && /\d/.test(nextChar))) {
                        // Handle numbers
                        let numStr = '';
                        while (i < jsonString.length && /[\d\.\-\+e]/i.test(jsonString[i])) {
                            numStr += jsonString[i];
                            i++;
                        }
                        result += '<span class="json-number">' + numStr + '</span>';
                        i--; // Adjust for the loop increment
                    } else if (char === 't' && jsonString.substr(i, 4) === 'true') {
                        result += '<span class="json-boolean">true</span>';
                        i += 3; // Skip the rest of 'true'
                    } else if (char === 'f' && jsonString.substr(i, 5) === 'false') {
                        result += '<span class="json-boolean">false</span>';
                        i += 4; // Skip the rest of 'false'
                    } else if (char === 'n' && jsonString.substr(i, 4) === 'null') {
                        result += '<span class="json-null">null</span>';
                        i += 3; // Skip the rest of 'null'
                    } else {
                        result += char;
                    }
                } else {
                    // Inside string, just add character
                    result += char;
                }

                i++;
            }

            return result;
        }

        // Download functions
        function downloadSummary() {
            if (!summaryContent) {
                alert('Summary content not available');
                return;
            }
            downloadFile(summaryContent, 'summary.json', 'application/json');
        }

        function downloadLog() {
            if (!logContent) {
                alert('Log content not available');
                return;
            }
            downloadFile(logContent, 'hrp.log', 'text/plain');
        }

        function downloadFile(content, filename, mimeType) {
            const blob = new Blob([content], { type: mimeType + ';charset=utf-8' });
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = filename;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            window.URL.revokeObjectURL(url);
        }

        // JSON Case Modal functions
        function showCaseJson() {
            if (!caseContent) {
                alert('Case JSON content not available');
                return;
            }

            try {
                // Parse and format JSON for beautiful display
                const jsonObj = JSON.parse(caseContent);
                const formattedJson = JSON.stringify(jsonObj, null, 4);

                // Apply syntax highlighting
                const highlightedJson = highlightJSONAdvanced(formattedJson);
                document.getElementById('jsonContent').innerHTML = highlightedJson;
                document.getElementById('jsonModal').style.display = 'block';
            } catch (e) {
                console.error('Failed to parse JSON:', e);
                // Fallback to raw content if parsing fails
                document.getElementById('jsonContent').textContent = caseContent;
                document.getElementById('jsonModal').style.display = 'block';
            }
        }

        function closeJsonModal() {
            document.getElementById('jsonModal').style.display = 'none';
        }

        function copyJsonContent() {
            // Copy the original formatted JSON content instead of highlighted HTML
            if (!caseContent) {
                alert('No content to copy');
                return;
            }

            try {
                const jsonObj = JSON.parse(caseContent);
                const formattedJson = JSON.stringify(jsonObj, null, 4);

                navigator.clipboard.writeText(formattedJson).then(function() {
                    const copyStatus = document.getElementById('copyStatus');
                    copyStatus.classList.add('show');
                    setTimeout(function() {
                        copyStatus.classList.remove('show');
                    }, 2000);
                }).catch(function(err) {
                    console.error('Failed to copy to clipboard:', err);
                    alert('Failed to copy to clipboard. Please select and copy manually.');
                });
            } catch (e) {
                // Fallback to original content
                navigator.clipboard.writeText(caseContent).then(function() {
                    const copyStatus = document.getElementById('copyStatus');
                    copyStatus.classList.add('show');
                    setTimeout(function() {
                        copyStatus.classList.remove('show');
                    }, 2000);
                }).catch(function(err) {
                    console.error('Failed to copy to clipboard:', err);
                    alert('Failed to copy to clipboard. Please select and copy manually.');
                });
            }
        }

        function downloadCaseJson() {
            if (!caseContent) {
                alert('Case JSON content not available');
                return;
            }
            downloadFile(caseContent, 'case.json', 'application/json');
        }

        // Summary JSON Modal functions
        function showSummaryJson() {
            if (!summaryContent) {
                alert('Summary JSON content not available');
                return;
            }

            try {
                // Parse and format JSON for beautiful display
                const jsonObj = JSON.parse(summaryContent);
                const formattedJson = JSON.stringify(jsonObj, null, 4);

                // Apply syntax highlighting
                const highlightedJson = highlightJSONAdvanced(formattedJson);
                document.getElementById('summaryContent').innerHTML = highlightedJson;
                document.getElementById('summaryModal').style.display = 'block';
            } catch (e) {
                console.error('Failed to parse JSON:', e);
                // Fallback to raw content if parsing fails
                document.getElementById('summaryContent').textContent = summaryContent;
                document.getElementById('summaryModal').style.display = 'block';
            }
        }

        function closeSummaryModal() {
            document.getElementById('summaryModal').style.display = 'none';
        }

        function copySummaryContent() {
            // Copy the original formatted JSON content instead of highlighted HTML
            if (!summaryContent) {
                alert('No content to copy');
                return;
            }

            try {
                const jsonObj = JSON.parse(summaryContent);
                const formattedJson = JSON.stringify(jsonObj, null, 4);

                navigator.clipboard.writeText(formattedJson).then(function() {
                    const copyStatus = document.getElementById('summaryStatus');
                    copyStatus.classList.add('show');
                    setTimeout(function() {
                        copyStatus.classList.remove('show');
                    }, 2000);
                }).catch(function(err) {
                    console.error('Failed to copy to clipboard:', err);
                    alert('Failed to copy to clipboard. Please select and copy manually.');
                });
            } catch (e) {
                // Fallback to original content
                navigator.clipboard.writeText(summaryContent).then(function() {
                    const copyStatus = document.getElementById('summaryStatus');
                    copyStatus.classList.add('show');
                    setTimeout(function() {
                        copyStatus.classList.remove('show');
                    }, 2000);
                }).catch(function(err) {
                    console.error('Failed to copy to clipboard:', err);
                    alert('Failed to copy to clipboard. Please select and copy manually.');
                });
            }
        }

        // Log Content Modal functions
        function showLogContent() {
            if (!logContent) {
                alert('Log content not available');
                return;
            }

            document.getElementById('logContentDisplay').textContent = logContent;
            document.getElementById('logModal').style.display = 'block';
        }

        function closeLogModal() {
            document.getElementById('logModal').style.display = 'none';
        }

        function copyLogContent() {
            const content = document.getElementById('logContentDisplay').textContent;
            if (!content) {
                alert('No content to copy');
                return;
            }

            navigator.clipboard.writeText(content).then(function() {
                const copyStatus = document.getElementById('logStatus');
                copyStatus.classList.add('show');
                setTimeout(function() {
                    copyStatus.classList.remove('show');
                }, 2000);
            }).catch(function(err) {
                console.error('Failed to copy to clipboard:', err);
                alert('Failed to copy to clipboard. Please select and copy manually.');
            });
        }

        function toggleStep(stepIndex) {
            const content = document.getElementById('step-' + stepIndex);
            const icon = document.getElementById('toggle-' + stepIndex);

            if (content.classList.contains('show')) {
                content.classList.remove('show');
                icon.classList.remove('rotated');
            } else {
                content.classList.add('show');
                icon.classList.add('rotated');
            }
        }

        function toggleLogFields(headerElement) {
            const logEntry = headerElement.parentElement;
            const fieldsElement = logEntry.querySelector('.log-fields');
            const toggleIcon = headerElement.querySelector('.log-toggle');

            if (fieldsElement && toggleIcon) {
                if (fieldsElement.classList.contains('collapsed')) {
                    fieldsElement.classList.remove('collapsed');
                    toggleIcon.classList.add('rotated');
                    // Apply JSON highlighting when expanding
                    if (fieldsElement.classList.contains('json-inline')) {
                        const text = fieldsElement.textContent;
                        if (text && text.trim()) {
                            try {
                                JSON.parse(text);
                                const highlighted = highlightJSONAdvanced(text);
                                fieldsElement.innerHTML = highlighted;
                            } catch (e) {
                                // If not valid JSON, leave as is
                            }
                        }
                    }
                } else {
                    fieldsElement.classList.add('collapsed');
                    toggleIcon.classList.remove('rotated');
                }
            }
        }

        function toggleStepLogs(stepIndex) {
            const container = document.getElementById('logs-container-' + stepIndex);
            const toggle = document.getElementById('logs-toggle-' + stepIndex);

            if (container.classList.contains('show')) {
                container.classList.remove('show');
                toggle.classList.add('collapsed');
                toggle.textContent = '▶';
            } else {
                container.classList.add('show');
                toggle.classList.remove('collapsed');
                toggle.textContent = '▼';
            }
        }

        function toggleRequestsCompact(buttonElement) {
            const requestsDiv = buttonElement.parentElement;
            const requestsContent = requestsDiv.querySelector('.requests-content-compact');

            if (requestsContent.classList.contains('show')) {
                requestsContent.classList.remove('show');
                buttonElement.textContent = buttonElement.textContent.replace('Hide', 'Show');
            } else {
                requestsContent.classList.add('show');
                buttonElement.textContent = buttonElement.textContent.replace('Show', 'Hide');

                // Apply JSON highlighting to request/response bodies when expanding
                setTimeout(() => {
                    applyRequestResponseHighlighting(requestsContent);
                }, 10);
            }
        }

        // Apply JSON highlighting to request/response content
        function applyRequestResponseHighlighting(container) {
            // Find all request-body-compact and response-body-compact elements
            const requestBodies = container.querySelectorAll('.request-body-compact, .response-body-compact');

            requestBodies.forEach(function(element) {
                // Skip if already processed
                if (element.querySelector('.json-key, .json-string, .json-number')) {
                    return;
                }

                const text = element.textContent;
                if (text && text.trim()) {
                    // Extract the content after "Request:" or "Response:"
                    const match = text.match(/^(Request|Response):\s*(.+)$/s);
                    if (match) {
                        const label = match[1];
                        const content = match[2].trim();
                        try {
                            // Validate JSON by parsing it
                            const parsedJson = JSON.parse(content);
                            // Re-stringify to get a compact, normalized string, which removes extra spaces
                            const compactJson = JSON.stringify(parsedJson);
                            // Apply highlighting on the compact string
                            const highlighted = highlightJSONAdvanced(compactJson);
                            element.innerHTML = label + ': ' + highlighted;
                        } catch (e) {
                            // If not valid JSON, leave as is
                            console.log('Not valid JSON for ' + label + ':', content);
                        }
                    } else {
                        // Try to find JSON-like content even without exact format
                        const jsonMatch = text.match(/(\{.*\}|\[.*\])/s);
                        if (jsonMatch) {
                            const jsonContent = jsonMatch[1].trim();
                            try {
                                JSON.parse(jsonContent);
                                const beforeJson = text.substring(0, text.indexOf(jsonContent));
                                const afterJson = text.substring(text.indexOf(jsonContent) + jsonContent.length);
                                const highlighted = highlightJSONAdvanced(jsonContent);
                                element.innerHTML = beforeJson + highlighted + afterJson;
                            } catch (e) {
                                // Not valid JSON, leave as is
                            }
                        }
                    }
                }
            });
        }

        function openImageModal(src) {
            const modal = document.getElementById('imageModal');
            const modalImg = document.getElementById('modalImage');
            modal.style.display = 'block';
            modalImg.src = src;
        }

        function closeModal() {
            document.getElementById('imageModal').style.display = 'none';
        }

        // Close modal when clicking outside the image or JSON modals
        window.onclick = function(event) {
            const imageModal = document.getElementById('imageModal');
            const jsonModal = document.getElementById('jsonModal');
            const summaryModal = document.getElementById('summaryModal');
            const logModal = document.getElementById('logModal');

            if (event.target == imageModal) {
                imageModal.style.display = 'none';
            }
            if (event.target == jsonModal) {
                jsonModal.style.display = 'none';
            }
            if (event.target == summaryModal) {
                summaryModal.style.display = 'none';
            }
            if (event.target == logModal) {
                logModal.style.display = 'none';
            }
        }

        // Apply syntax highlighting to inline JSON content
        function applyInlineJSONHighlighting() {
            document.querySelectorAll('.json-inline').forEach(function(element) {
                const text = element.textContent;
                if (text && text.trim()) {
                    try {
                        // Validate and parse JSON
                        JSON.parse(text);
                        // Apply highlighting if valid JSON
                        const highlighted = highlightJSONAdvanced(text);
                        element.innerHTML = highlighted;
                    } catch (e) {
                        // If not valid JSON, leave as is
                        // This handles cases where content might not be pure JSON
                    }
                }
            });
        }

        // Auto-expand all steps on load to show actions
        document.addEventListener('DOMContentLoaded', function() {
            // Expand all steps to show the actions list
            const contents = document.querySelectorAll('.step-content');
            const icons = document.querySelectorAll('.toggle-icon');

            contents.forEach(content => content.classList.add('show'));
            icons.forEach(icon => icon.classList.add('rotated'));

            // Apply syntax highlighting to inline JSON content
            applyInlineJSONHighlighting();

            // Apply JSON highlighting to all visible request/response content
            document.querySelectorAll('.requests-content-compact').forEach(function(container) {
                applyRequestResponseHighlighting(container);
            });
        });

        function toggleAllSteps() {
            const contents = document.querySelectorAll('.step-content');
            const icons = document.querySelectorAll('.toggle-icon');
            const btn = document.getElementById('toggle-all-steps-btn');

            if (!contents || contents.length === 0) {
                return;
            }

            // If any step is expanded, collapse all. Otherwise expand all.
            let isAnyExpanded = false;
            for (let i = 0; i < contents.length; i++) {
                if (contents[i].classList.contains('show')) {
                    isAnyExpanded = true;
                    break;
                }
            }

            if (isAnyExpanded) {
                contents.forEach(content => content.classList.remove('show'));
                icons.forEach(icon => icon.classList.remove('rotated'));
                btn.textContent = 'Expand All Steps';
            } else {
                contents.forEach(content => content.classList.add('show'));
                icons.forEach(icon => icon.classList.add('rotated'));
                btn.textContent = 'Collapse All Steps';
            }
        }
    </script>
</body>
</html>`
