package hrp

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/httprunner/httprunner/v5/internal/builtin"
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
	SummaryFile string
	LogFile     string
	SummaryData *Summary
	LogData     []LogEntry
	ReportDir   string
}

// LogEntry represents a single log entry
type LogEntry struct {
	Time    string         `json:"time"`
	Level   string         `json:"level"`
	Message string         `json:"message"`
	Fields  map[string]any `json:"-"` // Store all other fields
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

	g.SummaryData = &Summary{}
	return json.Unmarshal(data, g.SummaryData)
}

// loadLogData loads test log data from log file
func (g *HTMLReportGenerator) loadLogData() error {
	if g.LogFile == "" || !builtin.FileExists(g.LogFile) {
		return nil
	}

	file, err := os.Open(g.LogFile)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
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
			Fields: make(map[string]any),
		}

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

	// startTime is in seconds, elapsed is in milliseconds
	// Calculate end time (startTime in seconds + elapsed in milliseconds converted to seconds)
	endTime := startTime + elapsed/1000

	// Convert Unix timestamps to time.Time for comparison
	startTimeObj := time.Unix(startTime, 0)
	endTimeObj := time.Unix(endTime, 0)

	for _, logEntry := range g.LogData {
		// Parse log entry time
		logTime, err := g.parseLogTime(logEntry.Time)
		if err != nil {
			continue
		}

		// Check if log entry falls within step time range
		if (logTime.Equal(startTimeObj) || logTime.After(startTimeObj)) &&
			(logTime.Equal(endTimeObj) || logTime.Before(endTimeObj)) {
			stepLogs = append(stepLogs, logEntry)
		}
	}

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

// encodeImageToBase64 encodes an image file to base64 string
func (g *HTMLReportGenerator) encodeImageToBase64(imagePath string) string {
	// Convert relative path to absolute path
	if !filepath.IsAbs(imagePath) {
		imagePath = filepath.Join(g.ReportDir, imagePath)
	}

	if !builtin.FileExists(imagePath) {
		log.Warn().Str("path", imagePath).Msg("image file not found")
		return ""
	}

	data, err := os.ReadFile(imagePath)
	if err != nil {
		log.Warn().Err(err).Str("path", imagePath).Msg("failed to read image file")
		return ""
	}

	return base64.StdEncoding.EncodeToString(data)
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
					if action.SubActions != nil {
						total += len(action.SubActions)
					}
				}
			}
		}
	}
	return total
}

// calculateTotalRequests calculates the total number of requests across all test cases
func (g *HTMLReportGenerator) calculateTotalRequests() int {
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
					if action.SubActions != nil {
						for _, subAction := range action.SubActions {
							if subAction.Requests != nil {
								total += len(subAction.Requests)
							}
						}
					}
				}
			}
		}
	}
	return total
}

// calculateTotalScreenshots calculates the total number of screenshots across all test cases
func (g *HTMLReportGenerator) calculateTotalScreenshots() int {
	total := 0
	if g.SummaryData == nil || g.SummaryData.Details == nil {
		return total
	}

	for _, testCase := range g.SummaryData.Details {
		if testCase.Records == nil {
			continue
		}
		for _, step := range testCase.Records {
			// Count screenshots in actions
			if step.Actions != nil {
				for _, action := range step.Actions {
					if action.SubActions != nil {
						for _, subAction := range action.SubActions {
							if subAction.ScreenResults != nil {
								total += len(subAction.ScreenResults)
							}
						}
					}
				}
			}
			// Count screenshots in attachments
			if step.Attachments != nil {
				if attachments, ok := step.Attachments.(map[string]any); ok {
					if screenResults, exists := attachments["screen_results"]; exists {
						if screenResultsSlice, ok := screenResults.([]any); ok {
							total += len(screenResultsSlice)
						}
					}
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
		"formatDuration":            g.formatDuration,
		"encodeImageBase64":         g.encodeImageToBase64,
		"getStepLogs":               g.getStepLogsForTemplate,
		"calculateTotalActions":     g.calculateTotalActions,
		"calculateTotalSubActions":  g.calculateTotalSubActions,
		"calculateTotalRequests":    g.calculateTotalRequests,
		"calculateTotalScreenshots": g.calculateTotalScreenshots,
		"safeHTML":                  func(s string) template.HTML { return template.HTML(s) },
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
		"base":  filepath.Base,
		"index": func(m map[string]any, key string) any { return m[key] },
	}

	// Parse template
	tmpl, err := template.New("report").Funcs(funcMap).Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Create output file
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Execute template
	if err := tmpl.Execute(file, g.SummaryData); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
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
        }

        .start-time {
            background: rgba(255, 255, 255, 0.2);
            padding: 12px 20px;
            border-radius: 8px;
            backdrop-filter: blur(10px);
        }

        .time-label {
            display: block;
            font-size: 0.9em;
            opacity: 0.8;
            margin-bottom: 4px;
        }

        .time-value {
            display: block;
            font-size: 1.1em;
            font-weight: bold;
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

        .controls {
            background: white;
            padding: 20px;
            border-radius: 10px;
            margin-bottom: 30px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            text-align: center;
        }

        .controls button {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border: none;
            padding: 12px 24px;
            margin: 0 10px;
            border-radius: 6px;
            cursor: pointer;
            font-size: 14px;
            font-weight: 500;
            transition: all 0.3s ease;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }

        .controls button:hover {
            transform: translateY(-2px);
            box-shadow: 0 4px 8px rgba(0,0,0,0.2);
        }

        .controls button:active {
            transform: translateY(0);
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }

        .step-container {
            background: white;
            margin-bottom: 20px;
            border-radius: 10px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            overflow: hidden;
        }

        .step-header {
            background: #f8f9fa;
            padding: 20px;
            cursor: pointer;
            border-bottom: 1px solid #dee2e6;
            transition: background-color 0.3s;
        }

        .step-header:hover {
            background: #e9ecef;
        }

        .step-header h3 {
            display: flex;
            align-items: center;
            gap: 15px;
            margin: 0;
            font-size: 1.3em;
        }

        .step-number {
            background: #007bff;
            color: white;
            width: 30px;
            height: 30px;
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 0.9em;
            font-weight: bold;
        }

        .status-badge {
            padding: 5px 12px;
            border-radius: 20px;
            font-size: 0.8em;
            font-weight: bold;
        }

        .status-badge.success {
            background: #28a745;
            color: white;
        }

        .status-badge.failure {
            background: #dc3545;
            color: white;
        }

        .duration {
            background: #6c757d;
            color: white;
            padding: 3px 8px;
            border-radius: 12px;
            font-size: 0.8em;
        }

        .toggle-icon {
            margin-left: auto;
            font-size: 0.8em;
            transition: transform 0.3s;
        }

        .toggle-icon.rotated {
            transform: rotate(-90deg);
        }

        .step-meta {
            margin-top: 10px;
            color: #6c757d;
        }

        .step-type {
            background: #17a2b8;
            color: white;
            padding: 2px 8px;
            border-radius: 10px;
            font-size: 0.8em;
        }

        .step-content {
            padding: 25px;
            display: none;
        }

        .step-content.show {
            display: block;
        }

        .actions-section, .validators-section, .screenshots-section, .logs-section {
            margin-bottom: 25px;
        }

        .actions-section h4, .validators-section h4, .screenshots-section h4, .logs-section h4 {
            color: #495057;
            margin-bottom: 15px;
            padding-bottom: 8px;
            border-bottom: 1px solid #dee2e6;
        }

        .action-item {
            background: #f8f9fa;
            border: 1px solid #dee2e6;
            border-radius: 8px;
            padding: 15px;
            margin-bottom: 15px;
        }

        .action-header {
            display: flex;
            align-items: center;
            gap: 15px;
            margin-bottom: 10px;
            cursor: pointer;
            transition: background-color 0.3s;
            padding: 8px;
            border-radius: 6px;
        }

        .action-header:hover {
            background-color: rgba(0, 123, 255, 0.1);
        }

        .action-header strong {
            color: #007bff;
        }

        .action-toggle {
            margin-left: auto;
            font-size: 0.8em;
            color: #6c757d;
            transition: transform 0.3s;
        }

        .action-toggle.rotated {
            transform: rotate(-90deg);
        }

        .action-toggle.collapsed {
            transform: rotate(-90deg);
        }

        .action-content {
            display: none;
        }

        .action-content.expanded {
            display: block;
        }

        .action-params {
            color: #6c757d;
            font-style: italic;
            margin-bottom: 10px;
            white-space: pre-wrap;
            word-wrap: break-word;
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
            background: #f8f9fa;
            border: 1px solid #dee2e6;
            border-radius: 4px;
            padding: 10px;
            font-size: 0.9em;
            line-height: 1.4;
        }

        .error {
            color: #dc3545;
            font-weight: bold;
        }

        .sub-actions {
            margin-top: 15px;
            padding-left: 20px;
            border-left: 3px solid #dee2e6;
        }

        .sub-action-item {
            background: white;
            border: 1px solid #e9ecef;
            border-radius: 6px;
            padding: 12px;
            margin-bottom: 10px;
        }

        .sub-action-content {
            display: flex;
            gap: 20px;
            align-items: flex-start;
        }

        .sub-action-left {
            flex: 1;
            min-width: 0;
        }

        .sub-action-right {
            flex: 1;
            min-width: 0;
        }

        .sub-action-header {
            display: flex;
            align-items: center;
            gap: 10px;
            margin-bottom: 8px;
        }

        .action-name {
            background: #6f42c1;
            color: white;
            padding: 2px 8px;
            border-radius: 12px;
            font-size: 0.8em;
            font-weight: bold;
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

        .sub-action-screenshots, .screenshots-section {
            background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%);
            border: 2px solid #28a745;
            border-radius: 12px;
            padding: 12px;
            box-shadow: 0 4px 12px rgba(40, 167, 69, 0.15);
        }

        .sub-action-screenshots h5, .screenshots-section h4 {
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
        }

        .screenshot-image img {
            max-width: 100%;
            max-height: 400px;
            border-radius: 6px;
            cursor: pointer;
            transition: transform 0.2s;
        }

        .screenshot-image img:hover {
            transform: scale(1.02);
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
                text-align: left;
                width: 100%;
            }

            .start-time {
                width: 100%;
                text-align: center;
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
                font-size: 1.1em;
                gap: 10px;
            }

            .step-number {
                width: 25px;
                height: 25px;
                font-size: 0.8em;
            }

            .action-header {
                flex-direction: column;
                align-items: flex-start;
                gap: 8px;
            }

            .controls button {
                padding: 6px 10px;
                font-size: 0.8em;
                margin: 2px;
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

            .sub-action-content {
                flex-direction: column;
                gap: 15px;
            }

            .screenshots-grid {
                grid-template-columns: 1fr;
                gap: 10px;
            }

            .screenshot-image img {
                max-height: 250px;
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
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="header-content">
                <div class="header-left">
                    <h1>🚀 HttpRunner Test Report</h1>
                    <div class="subtitle">Automated Testing Results</div>
                </div>
                <div class="header-right">
                    <div class="start-time">
                        <span class="time-label">Start Time:</span>
                        <span class="time-value">{{.Time.StartAt.Format "2006-01-02 15:04:05"}}</span>
                    </div>
                </div>
            </div>
        </div>

        <div class="summary">
            <h2>📊 Test Summary</h2>
            <div class="summary-grid">
                <div class="summary-item">
                    <div class="value">{{.Stat.TestCases.Total}}</div>
                    <div class="label">Total Test Cases</div>
                </div>
                <div class="summary-item success">
                    <div class="value">{{.Stat.TestCases.Success}}</div>
                    <div class="label">Passed</div>
                </div>
                <div class="summary-item failure">
                    <div class="value">{{.Stat.TestCases.Fail}}</div>
                    <div class="label">Failed</div>
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
                    <div class="value">{{calculateTotalRequests}}</div>
                    <div class="label">Total Requests</div>
                </div>
                <div class="summary-item">
                    <div class="value">{{calculateTotalScreenshots}}</div>
                    <div class="label">Total Screenshots</div>
                </div>
                <div class="summary-item">
                    <div class="value">{{printf "%.1f" .Time.Duration}}s</div>
                    <div class="label">Duration</div>
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

        <div class="controls">
            <button id="toggleStepsBtn" onclick="toggleAllSteps()">Collapse All Steps</button>
            <button id="toggleActionsBtn" onclick="toggleAllActions()">Expand All Actions</button>
        </div>

        <div class="test-cases">
            {{range $caseIndex, $testCase := .Details}}
            <div class="test-case">
                <h2>📋 {{$testCase.Name}}</h2>
                <div class="case-info">
                    <span class="status-badge {{if $testCase.Success}}success{{else}}failure{{end}}">
                        {{if $testCase.Success}}✓ PASS{{else}}✗ FAIL{{end}}
                    </span>
                    <span class="duration">{{printf "%.1f" $testCase.Time.Duration}}s</span>
                </div>

                {{range $stepIndex, $step := $testCase.Records}}
                <div class="step-container">
                    <div class="step-header" onclick="toggleStep({{$stepIndex}})">
                        <h3>
                            <span class="step-number">{{add $stepIndex 1}}</span>
                            {{$step.Name}}
                            <span class="status-badge {{if $step.Success}}success{{else}}failure{{end}}">
                                {{if $step.Success}}✓ PASS{{else}}✗ FAIL{{end}}
                            </span>
                            <span class="duration">{{formatDuration $step.Elapsed}}</span>
                            <span class="toggle-icon" id="toggle-{{$stepIndex}}">▼</span>
                        </h3>
                        <div class="step-meta">
                            <span class="step-type">{{$step.StepType}}</span>
                        </div>
                    </div>

                    <div class="step-content" id="step-{{$stepIndex}}">
                        <!-- Actions -->
                        {{if $step.Actions}}
                        <div class="actions-section">
                            <h4>Actions</h4>
                            {{range $actionIndex, $action := $step.Actions}}
                            <div class="action-item">
                                <div class="action-header" onclick="toggleAction({{$stepIndex}}, {{$actionIndex}})">
                                    <strong>{{$action.Method}}</strong>
                                    <span class="duration">{{formatDuration $action.Elapsed}}</span>
                                    {{if $action.Error}}<span class="error">Error: {{$action.Error}}</span>{{end}}
                                    <span class="action-toggle collapsed" id="action-toggle-{{$stepIndex}}-{{$actionIndex}}">▶</span>
                                </div>
                                <div class="action-content" id="action-content-{{$stepIndex}}-{{$actionIndex}}">
                                    <div class="action-params">{{$action.Params}}</div>

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

                                                {{if $subAction.Thought}}
                                                <div class="thought">{{$subAction.Thought}}</div>
                                                {{end}}

                                                {{if $subAction.ModelName}}
                                                <div class="model-name-container">
                                                    <span class="model-label">🤖 Model:</span>
                                                    <span class="model-value">{{$subAction.ModelName}}</span>
                                                </div>
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
                                {{if $validator.msg}}
                                <div class="validator-message">{{$validator.msg}}</div>
                                {{end}}
                            </div>
                            {{end}}
                        </div>
                        {{end}}

                        <!-- Screenshots -->
                        {{if $step.Attachments}}{{if $step.Attachments.ScreenResults}}
                        <div class="screenshots-section">
                            <h4>Screenshots</h4>
                            {{range $screenshot := $step.Attachments.ScreenResults}}
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
                        {{end}}{{end}}

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

        function toggleAction(stepIndex, actionIndex) {
            const content = document.getElementById('action-content-' + stepIndex + '-' + actionIndex);
            const toggle = document.getElementById('action-toggle-' + stepIndex + '-' + actionIndex);

            if (content.classList.contains('expanded')) {
                content.classList.remove('expanded');
                toggle.classList.add('collapsed');
                toggle.textContent = '▶';
            } else {
                content.classList.add('expanded');
                toggle.classList.remove('collapsed');
                toggle.textContent = '▼';
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

        // Toggle all steps
        function toggleAllSteps() {
            const contents = document.querySelectorAll('.step-content');
            const icons = document.querySelectorAll('.toggle-icon');
            const button = document.getElementById('toggleStepsBtn');

            // Check if any step is currently expanded
            const anyExpanded = Array.from(contents).some(content => content.classList.contains('show'));

            if (anyExpanded) {
                // Collapse all
                contents.forEach(content => content.classList.remove('show'));
                icons.forEach(icon => icon.classList.remove('rotated'));
                button.textContent = 'Expand All Steps';
            } else {
                // Expand all
                contents.forEach(content => content.classList.add('show'));
                icons.forEach(icon => icon.classList.add('rotated'));
                button.textContent = 'Collapse All Steps';
            }
        }

        // Toggle all actions
        function toggleAllActions() {
            const contents = document.querySelectorAll('.action-content');
            const toggles = document.querySelectorAll('.action-toggle');
            const button = document.getElementById('toggleActionsBtn');

            // Check if any action is currently expanded
            const anyExpanded = Array.from(contents).some(content => content.classList.contains('expanded'));

            if (anyExpanded) {
                // Collapse all
                contents.forEach(content => content.classList.remove('expanded'));
                toggles.forEach(toggle => {
                    toggle.classList.add('collapsed');
                    toggle.textContent = '▶';
                });
                button.textContent = 'Expand All Actions';
            } else {
                // Expand all
                contents.forEach(content => content.classList.add('expanded'));
                toggles.forEach(toggle => {
                    toggle.classList.remove('collapsed');
                    toggle.textContent = '▼';
                });
                button.textContent = 'Collapse All Actions';
            }
        }

        // Auto-expand all steps on load to show actions
        document.addEventListener('DOMContentLoaded', function() {
            // Expand all steps to show the actions list
            const contents = document.querySelectorAll('.step-content');
            const icons = document.querySelectorAll('.toggle-icon');
            const stepsButton = document.getElementById('toggleStepsBtn');

            contents.forEach(content => content.classList.add('show'));
            icons.forEach(icon => icon.classList.add('rotated'));

            // Update button text to reflect current state (steps are expanded)
            if (stepsButton) {
                stepsButton.textContent = 'Collapse All Steps';
            }
        });
    </script>
</body>
</html>`
