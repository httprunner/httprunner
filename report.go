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

	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/uixt"
	"github.com/httprunner/httprunner/v5/uixt/ai"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
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

// getStepLogs filters log entries for a specific test step based on time range
func (g *HTMLReportGenerator) getStepLogs(stepName string, startTime int64, elapsed int64) []LogEntry {
	if len(g.LogData) == 0 {
		return nil
	}

	var stepLogs []LogEntry

	// Handle different time units properly
	// Check if startTime is already in milliseconds (> year 2100 in seconds)
	var startTimeMs, endTimeMs int64
	if startTime > 4000000000 { // If > year 2100, assume it's already in milliseconds
		startTimeMs = startTime
	} else {
		// startTime is in seconds (backward compatibility), elapsed is in milliseconds
		startTimeMs = startTime * 1000
	}
	endTimeMs = startTimeMs + elapsed

	// Find the actual time range of this step's logs by scanning for step boundaries
	var actualStartTime, actualEndTime int64
	actualStartTime = startTimeMs
	actualEndTime = endTimeMs

	// Look for step start/end markers to determine precise boundaries
	for _, logEntry := range g.LogData {
		logTime, err := g.parseLogTimeMs(logEntry.Time)
		if err != nil {
			continue
		}

		// Check for step start marker - more precise matching
		if logEntry.Message == "run step start" {
			if stepFieldValue, exists := logEntry.Fields["step"]; exists {
				if stepFieldValue == stepName {
					actualStartTime = logTime
				}
			}
		}
		// Check for step end marker - more precise matching
		if logEntry.Message == "run step end" {
			if stepFieldValue, exists := logEntry.Fields["step"]; exists {
				if stepFieldValue == stepName {
					actualEndTime = logTime
				}
			}
		}
	}

	// Find the next step's start time to avoid overlapping logs
	var nextStepStartTime int64 = 0
	for _, logEntry := range g.LogData {
		logTime, err := g.parseLogTimeMs(logEntry.Time)
		if err != nil {
			continue
		}

		// Find the next step start that occurs after our step's actual end time
		if logEntry.Message == "run step start" && logTime > actualEndTime {
			if stepFieldValue, exists := logEntry.Fields["step"]; exists {
				if stepFieldValue != stepName { // Make sure it's a different step
					nextStepStartTime = logTime
					break
				}
			}
		}
	}

	// Add intelligent buffer for AI conversations, but respect step boundaries
	var maxAIEndTime int64 = actualEndTime
	var inAIConversation bool

	for _, logEntry := range g.LogData {
		logTime, err := g.parseLogTimeMs(logEntry.Time)
		if err != nil {
			continue
		}

		// Skip logs that are clearly before our step
		if logTime < actualStartTime {
			continue
		}

		// Skip logs that are after the next step starts
		if nextStepStartTime > 0 && logTime >= nextStepStartTime {
			break
		}

		// Track AI conversation patterns
		if strings.Contains(logEntry.Message, ai.LOG_REQUEST_MESSAGES) {
			if logTime >= actualStartTime && logTime <= actualEndTime+30000 { // 30s buffer for AI requests
				inAIConversation = true
			}
		} else if strings.Contains(logEntry.Message, ai.LOG_RESPONSE_MESSAGE) && inAIConversation {
			// Extend end time to include complete AI conversation, but respect next step boundary
			extendedEndTime := logTime
			if nextStepStartTime > 0 && extendedEndTime >= nextStepStartTime {
				extendedEndTime = nextStepStartTime - 1 // Stop just before next step
			}
			if extendedEndTime > maxAIEndTime {
				maxAIEndTime = extendedEndTime
			}
			inAIConversation = false
		}
	}

	// Use the extended end time if AI conversation was detected
	if maxAIEndTime > actualEndTime {
		actualEndTime = maxAIEndTime
	}

	// Final boundary check: ensure we don't go beyond the next step
	if nextStepStartTime > 0 && actualEndTime >= nextStepStartTime {
		actualEndTime = nextStepStartTime - 1
	}

	// Final filtering: collect logs strictly within step boundaries
	var inCurrentStep bool = false

	for _, logEntry := range g.LogData {

		// Check for step boundaries to control inclusion
		if logEntry.Message == "run step start" {
			if stepFieldValue, exists := logEntry.Fields["step"]; exists {
				if stepFieldValue == stepName {
					inCurrentStep = true
					stepLogs = append(stepLogs, logEntry)
					continue
				} else {
					// This is a different step starting
					if inCurrentStep {
						// We were in our step, now we're not
						break
					}
					continue
				}
			}
		}

		if logEntry.Message == "run step end" {
			if stepFieldValue, exists := logEntry.Fields["step"]; exists {
				if stepFieldValue == stepName {
					stepLogs = append(stepLogs, logEntry)
					inCurrentStep = false
					// Continue to check for AI conversations that might extend beyond
					continue
				}
			}
		}

		// Only include logs when we're in the current step
		if inCurrentStep {
			stepLogs = append(stepLogs, logEntry)
		}
	}

	// Sort logs by time, then by original index for stable ordering
	sort.Slice(stepLogs, func(i, j int) bool {
		timeI, errI := g.parseLogTime(stepLogs[i].Time)
		timeJ, errJ := g.parseLogTime(stepLogs[j].Time)

		if errI != nil || errJ != nil {
			return stepLogs[i].LogIndex < stepLogs[j].LogIndex
		}

		if timeI.Equal(timeJ) {
			// For same timestamps, use original log index to maintain order
			return stepLogs[i].LogIndex < stepLogs[j].LogIndex
		}
		return timeI.Before(timeJ)
	})

	return stepLogs
}

// containsLogEntry checks if a log entry already exists in the slice
func containsLogEntry(logs []LogEntry, entry LogEntry) bool {
	for _, log := range logs {
		if log.LogIndex == entry.LogIndex {
			return true
		}
	}
	return false
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

// parseLogTimeMs parses time string to milliseconds since Unix epoch
func (g *HTMLReportGenerator) parseLogTimeMs(timeStr string) (int64, error) {
	t, err := g.parseLogTime(timeStr)
	if err != nil {
		return 0, err
	}
	return t.UnixNano() / 1000000, nil // Convert nanoseconds to milliseconds
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
				total += len(step.Actions)
			}
		}
	}
	return total
}

// calculateTotalSubActions calculates the total number of sub-actions across all test cases
func (g *HTMLReportGenerator) calculateTotalSubActions() int {
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
					// Count sub-actions from regular actions
					if action.SubActions != nil {
						total += len(action.SubActions)
					}
					// Count sub-actions from planning results
					if action.Plannings != nil {
						for _, planning := range action.Plannings {
							if planning.SubActions != nil {
								total += len(planning.SubActions)
							}
						}
					}
				}
			}
		}
	}
	return total
}

// calculateTotalPlannings calculates the total number of planning results across all test cases
func (g *HTMLReportGenerator) calculateTotalPlannings() int {
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
					if action.Plannings != nil {
						total += len(action.Plannings)
					}
				}
			}
		}
	}
	return total
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
				if action.Plannings == nil {
					continue
				}
				for _, planning := range action.Plannings {
					if planning.Usage == nil {
						continue
					}
					totalUsage["prompt_tokens"] = totalUsage["prompt_tokens"].(int) + planning.Usage.PromptTokens
					totalUsage["completion_tokens"] = totalUsage["completion_tokens"].(int) + planning.Usage.CompletionTokens
					totalUsage["total_tokens"] = totalUsage["total_tokens"].(int) + planning.Usage.TotalTokens
				}
			}
		}
	}
	return totalUsage
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
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"toJSON": func(v any) string {
			var buf strings.Builder
			encoder := json.NewEncoder(&buf)
			encoder.SetEscapeHTML(false)
			_ = encoder.Encode(v)
			result := buf.String()
			return strings.TrimSpace(result)
		},
		"mul":   func(a, b float64) float64 { return a * b },
		"add":   func(a, b int) int { return a + b },
		"sub":   func(a, b int) int { return a - b },
		"lt":    func(a, b int) bool { return a < b },
		"gt":    func(a, b int) bool { return a > b },
		"base":  filepath.Base,
		"index": func(m map[string]any, key string) any { return m[key] },
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
            margin-bottom: 20px;
            border-bottom: 2px solid #3498db;
            padding-bottom: 10px;
        }

        .summary-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-bottom: 20px;
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
            max-height: 400px;
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
            max-height: 60px;
            overflow-y: auto;
            word-break: break-all;
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

        .raw-content {
            margin-top: 10px;
        }

        .raw-content pre {
            background: #f1f3f4;
            border: 1px solid #dadce0;
            border-radius: 4px;
            padding: 8px;
            font-size: 0.8em;
            max-height: 150px;
            overflow-y: auto;
            white-space: pre-wrap;
            word-wrap: break-word;
        }

        .step-screenshots {
            margin-top: 10px;
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

        .model-name-container {
            background: #f8f9fa;
            border: 1px solid #e9ecef;
            border-radius: 6px;
            padding: 8px 12px;
            margin: 8px 0;
            font-size: 0.9em;
            display: flex;
            align-items: center;
            gap: 8px;
        }

        .model-label {
            font-weight: 600;
            color: #495057;
        }

        .model-value {
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
            background: #e9ecef;
            padding: 2px 6px;
            border-radius: 4px;
            color: #495057;
            font-size: 0.85em;
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

        .requests {
            margin-top: 15px;
        }

        .requests-toggle {
            background: #6c757d;
            color: white;
            border: none;
            padding: 6px 12px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 0.8em;
            margin-bottom: 10px;
            transition: background-color 0.3s;
        }

        .requests-toggle:hover {
            background: #5a6268;
        }

        .requests-content {
            display: none;
        }

        .requests-content.show {
            display: block;
        }

        .request-item {
            background: #f1f3f4;
            border: 1px solid #dadce0;
            border-radius: 4px;
            padding: 8px;
            margin: 6px 0;
        }

        .request-header {
            display: flex;
            align-items: center;
            gap: 10px;
            margin-bottom: 6px;
        }

        .method {
            background: #007bff;
            color: white;
            padding: 2px 6px;
            border-radius: 4px;
            font-size: 0.8em;
            font-weight: bold;
        }

        .url {
            color: #495057;
            font-family: monospace;
            font-size: 0.9em;
        }

        .status {
            padding: 2px 6px;
            border-radius: 4px;
            font-size: 0.8em;
            font-weight: bold;
        }

        .status.success {
            background: #d4edda;
            color: #155724;
        }

        .status.failure {
            background: #f8d7da;
            color: #721c24;
        }

        .request-body, .response-body {
            background: #ffffff;
            border: 1px solid #e9ecef;
            border-radius: 4px;
            padding: 6px;
            margin: 4px 0;
            font-family: monospace;
            font-size: 0.8em;
            max-height: 100px;
            overflow-y: auto;
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

        .screenshots-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 10px;
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
            min-height: 300px;
            padding: 20px 0;
            background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%);
            border-radius: 8px;
            margin: 10px 0;
        }

        .screenshot-image img {
            max-width: 100%;
            max-height: 400px;
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
            min-height: 250px;
            padding: 15px 0;
        }

        .screenshot-item.small .screenshot-image img {
            max-height: 200px;
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
            gap: 10px;
            margin-bottom: 8px;
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

        .validator-expect,         .validator-message {
            margin: 4px 0;
            font-size: 0.9em;
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

        .controls {
            text-align: center;
            margin-bottom: 20px;
        }

        .controls button {
            background: #007bff;
            color: white;
            border: none;
            padding: 8px 16px;
            border-radius: 4px;
            margin: 0 5px;
            cursor: pointer;
            transition: background-color 0.3s;
        }

        .controls button:hover {
            background: #0056b3;
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

            .screenshot-image {
                min-height: 250px;
                padding: 15px 0;
            }

            .screenshot-image img {
                max-height: 250px;
            }

            .screenshot-item.small .screenshot-image {
                min-height: 200px;
                padding: 10px 0;
            }

            .screenshot-item.small .screenshot-image img {
                max-height: 150px;
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

        .action-output {
            background: #f8f9fa;
            border: 2px solid #6f42c1;
            border-radius: 6px;
            padding: 10px;
            font-size: 0.85em;
            max-height: 120px;
            overflow-y: auto;
            white-space: pre-wrap;
            word-wrap: break-word;
            color: #495057;
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
            line-height: 1.4;
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
                        <div class="download-title">📥 Download</div>
                        <div class="download-buttons">
                            <button class="download-btn" onclick="downloadSummary()">
                                <span>📄</span>
                                <span>summary.json</span>
                            </button>
                            <button class="download-btn" onclick="downloadLog()">
                                <span>📋</span>
                                <span>hrp.log</span>
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <div class="summary">
            <h2>📊 Test Summary</h2>
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
                                <div class="planning-results">
                                    {{range $planningIndex, $planning := $action.Plannings}}
                                    <div class="planning-item">
                                        <div class="planning-header">
                                            <span class="planning-label">🧠 Planning & Execution {{add $planningIndex 1}}</span>
                                            <span class="duration">{{formatDuration $planning.Elapsed}}</span>
                                            {{if $planning.Error}}<span class="error">Error: {{$planning.Error}}</span>{{end}}
                                        </div>

                                        {{if $planning.Thought}}
                                        <div class="thought">{{$planning.Thought}}</div>
                                        {{end}}

                                        <!-- Three-column layout: screenshot left, model output and actions right -->
                                        <div class="planning-three-columns">
                                                <!-- Left column: Screenshot (larger) -->
                                                <div class="planning-column-screenshot">
                                                    <div class="planning-step-compact">
                                                        <div class="step-header-compact">
                                                            <span class="step-name">📸 Take Screenshot</span>
                                                            <span class="duration">{{formatDuration $planning.ScreenshotElapsed}}</span>
                                                        </div>
                                                        {{if $planning.ScreenResult}}
                                                        <div class="screenshot-display">
                                                            {{$screenshot := $planning.ScreenResult}}
                                                            {{$base64Image := encodeImageBase64 $screenshot.ImagePath}}
                                                            {{if $base64Image}}
                                                            <div class="screenshot-item-compact">
                                                                <div class="screenshot-image">
                                                                    <img src="data:image/jpeg;base64,{{$base64Image}}" alt="Planning Screenshot" onclick="openImageModal(this.src)" />
                                                                </div>
                                                            </div>
                                                            {{end}}
                                                        </div>
                                                        {{end}}
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
                                                                <div class="actions-info">🎯 Actions: {{safeHTML (toJSON $planning.ActionNames)}}</div>
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
                                                                    <div class="action-arguments">{{safeHTML (toJSON $subAction.Arguments)}}</div>
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

                                        {{/* SubActions are now displayed in the right panel, so we don't show them here */}}
                                    </div>
                                    {{end}}
                                </div>
                                {{end}}

                                {{if $action.SubActions}}
                                <div class="sub-actions">
                                    {{range $subAction := $action.SubActions}}
                                    <div class="sub-action-item">
                                        <div class="sub-action-header">
                                            <span class="action-name">{{$subAction.ActionName}}</span>
                                            <span class="duration">{{formatDuration $subAction.Elapsed}}</span>
                                        </div>

                                        <div class="sub-action-content">
                                            <div class="sub-action-left">
                                                {{if $subAction.Arguments}}
                                                <div class="arguments">Arguments: {{safeHTML (toJSON $subAction.Arguments)}}</div>
                                                {{end}}

                                                {{if $subAction.Requests}}
                                                <div class="requests">
                                                    <button class="requests-toggle" onclick="toggleRequests(this)">
                                                        📡 Show Requests ({{len $subAction.Requests}})
                                                    </button>
                                                    <div class="requests-content">
                                                        {{range $request := $subAction.Requests}}
                                                        <div class="request-item">
                                                            <div class="request-header">
                                                                <span class="method">{{$request.RequestMethod}}</span>
                                                                <span class="url">{{$request.RequestUrl}}</span>
                                                                <span class="status {{if $request.Success}}success{{else}}failure{{end}}">Status: {{$request.ResponseStatus}}</span>
                                                                <span class="duration">{{formatDuration $request.ResponseDuration}}</span>
                                                            </div>
                                                            {{if $request.RequestBody}}
                                                            <div class="request-body">Request: {{$request.RequestBody}}</div>
                                                            {{end}}
                                                            {{if $request.ResponseBody}}
                                                            <div class="response-body">Response: {{$request.ResponseBody}}</div>
                                                            {{end}}
                                                        </div>
                                                        {{end}}
                                                    </div>
                                                </div>
                                                {{end}}
                                            </div>

                                            {{if $subAction.ScreenResults}}
                                            <div class="sub-action-right">
                                                <div class="sub-action-screenshots">
                                                    <h5>📸 Screenshots</h5>
                                                    <div class="screenshots-grid">
                                                        {{range $screenshot := $subAction.ScreenResults}}
                                                        {{$base64Image := encodeImageBase64 $screenshot.ImagePath}}
                                                        {{if $base64Image}}
                                                        <div class="screenshot-item small">
                                                            <div class="screenshot-info">
                                                                <span class="filename">{{base $screenshot.ImagePath}}</span>
                                                                {{if $screenshot.Resolution}}
                                                                <span class="resolution">{{$screenshot.Resolution.Width}}x{{$screenshot.Resolution.Height}}</span>
                                                                {{end}}
                                                            </div>
                                                            <div class="screenshot-image">
                                                                <img src="data:image/jpeg;base64,{{$base64Image}}" alt="Screenshot" onclick="openImageModal(this.src)" />
                                                            </div>
                                                        </div>
                                                        {{end}}
                                                        {{end}}
                                                    </div>
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
                            {{end}}
                        </div>
                        {{end}}

                        <!-- Validators -->
                        {{if and $step.Data $step.Data.validators}}
                        <div class="validators-section">
                            <h4>Validators</h4>
                            {{range $validator := $step.Data.validators}}
                            <div class="validator-item {{if eq $validator.check_result "pass"}}success{{else}}failure{{end}}">
                                <div class="validator-header">
                                    <span class="check-type">{{$validator.check}}</span>
                                    <span class="assert-type">{{$validator.assert}}</span>
                                    <span class="result">{{$validator.check_result}}</span>
                                </div>
                                <div class="validator-expect">Expected: {{$validator.expect}}</div>
                                {{if and $validator.msg (ne $validator.check_result "pass")}}
                                    <div class="validator-message">{{$validator.msg}}</div>
                                {{end}}
                            </div>
                            {{end}}
                        </div>
                        {{end}}

                        <!-- Screenshots -->
                        {{if $step.Attachments}}
                        {{$attachments := $step.Attachments}}
                        {{if eq (printf "%T" $attachments) "map[string]interface {}"}}
                        {{if index $attachments "screen_results"}}
                        <div class="screenshots-section">
                            <h4>Screenshots</h4>
                            {{range $screenshot := index $attachments "screen_results"}}
                            {{$base64Image := encodeImageBase64 $screenshot.ImagePath}}
                            {{if $base64Image}}
                            <div class="screenshot-item">
                                <div class="screenshot-info">
                                    <span class="filename">{{base $screenshot.ImagePath}}</span>
                                    {{if $screenshot.Resolution}}
                                    <span class="resolution">{{$screenshot.Resolution.Width}}x{{$screenshot.Resolution.Height}}</span>
                                    {{end}}
                                </div>
                                <div class="screenshot-image">
                                    <img src="data:image/jpeg;base64,{{$base64Image}}" alt="Screenshot" onclick="openImageModal(this.src)" />
                                </div>
                            </div>
                            {{end}}
                            {{end}}
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
                                     <div class="log-fields collapsed">{{safeHTML (toJSON $logEntry.Fields)}}</div>
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

    <script>
        // Embedded file contents for download (Base64 encoded)
        const summaryContentBase64 = "{{getSummaryContentBase64}}";
        const logContentBase64 = "{{getLogContentBase64}}";

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

                function toggleRequests(buttonElement) {
            const requestsDiv = buttonElement.parentElement;
            const requestsContent = requestsDiv.querySelector('.requests-content');

            if (requestsContent.classList.contains('show')) {
                requestsContent.classList.remove('show');
                buttonElement.textContent = buttonElement.textContent.replace('Hide', 'Show');
            } else {
                requestsContent.classList.add('show');
                buttonElement.textContent = buttonElement.textContent.replace('Show', 'Hide');
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
            }
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

        // Close modal when clicking outside the image
        window.onclick = function(event) {
            const modal = document.getElementById('imageModal');
            if (event.target == modal) {
                modal.style.display = 'none';
            }
        }



        // Auto-expand all steps on load to show actions
        document.addEventListener('DOMContentLoaded', function() {
            // Expand all steps to show the actions list
            const contents = document.querySelectorAll('.step-content');
            const icons = document.querySelectorAll('.toggle-icon');

            contents.forEach(content => content.classList.add('show'));
            icons.forEach(icon => icon.classList.add('rotated'));
        });
    </script>
</body>
</html>`
