package uixt

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/internal/version"
	"github.com/httprunner/httprunner/v5/uixt/option"
)

func NewMCPServer() *MCPServer4XTDriver {
	mcpServer := server.NewMCPServer(
		"uixt",
		version.GetVersionInfo(),
		server.WithToolCapabilities(false),
	)
	s := &MCPServer4XTDriver{
		mcpServer:     mcpServer,
		actionToolMap: make(map[option.ActionName]ActionTool),
	}
	s.registerTools()
	return s
}

// MCPServer4XTDriver wraps a MCPServer to expose XTDriver functionality via MCP protocol.
type MCPServer4XTDriver struct {
	mcpServer     *server.MCPServer
	mcpTools      []mcp.Tool                       // tools list for uixt
	actionToolMap map[option.ActionName]ActionTool // action method to tool mapping
}

// Start runs the MCP server (blocking).
func (s *MCPServer4XTDriver) Start() error {
	log.Info().Msg("Starting HttpRunner UIXT MCP Server...")
	return server.ServeStdio(s.mcpServer)
}

// ListTools returns all registered tools
func (s *MCPServer4XTDriver) ListTools() []mcp.Tool {
	return s.mcpTools
}

// GetTool returns a pointer to the mcp.Tool with the given name
func (s *MCPServer4XTDriver) GetTool(name string) *mcp.Tool {
	for i := range s.mcpTools {
		if s.mcpTools[i].Name == name {
			return &s.mcpTools[i]
		}
	}
	return nil
}

// GetToolByAction returns the tool that handles the given action method
func (s *MCPServer4XTDriver) GetToolByAction(actionMethod option.ActionName) ActionTool {
	if s.actionToolMap == nil {
		return nil
	}
	actionMethod = getActionNameByAlias(actionMethod)
	return s.actionToolMap[actionMethod]
}

func getActionNameByAlias(actionMethod option.ActionName) option.ActionName {
	switch strings.ToLower(string(actionMethod)) {
	case "terminal_app":
		return option.ACTION_AppTerminate
	case "open_app":
		return option.ACTION_AppLaunch
	case "text":
		return option.ACTION_Input
	case "tap":
		return option.ACTION_TapXY
	default:
		return actionMethod
	}
}

// registerTools registers all MCP tools.
func (s *MCPServer4XTDriver) registerTools() {
	// Device Tool
	s.registerTool(&ToolListAvailableDevices{}) // ListAvailableDevices
	s.registerTool(&ToolSelectDevice{})         // SelectDevice

	// Touch Tools
	s.registerTool(&ToolTapXY{})           // tap xy
	s.registerTool(&ToolTapAbsXY{})        // tap abs xy
	s.registerTool(&ToolTapByOCR{})        // tap by OCR
	s.registerTool(&ToolTapByCV{})         // tap by CV
	s.registerTool(&ToolDoubleTapXY{})     // double tap xy
	s.registerTool(&ToolSIMClickAtPoint{}) // simulated click at point

	// Swipe Tools
	s.registerTool(&ToolSwipe{})                    // generic swipe, auto-detect direction or coordinate
	s.registerTool(&ToolSwipeDirection{})           // swipe direction, up/down/left/right
	s.registerTool(&ToolSwipeCoordinate{})          // swipe coordinate, [fromX, fromY, toX, toY]
	s.registerTool(&ToolSIMSwipeDirection{})        // simulated swipe direction with random distance
	s.registerTool(&ToolSIMSwipeInArea{})           // simulated swipe in area with direction and distance
	s.registerTool(&ToolSIMSwipeFromPointToPoint{}) // simulated swipe from point to point
	s.registerTool(&ToolSwipeToTapApp{})
	s.registerTool(&ToolSwipeToTapText{})
	s.registerTool(&ToolSwipeToTapTexts{})
	s.registerTool(&ToolDrag{})

	// Input Tools
	s.registerTool(&ToolInput{})    // regular input
	s.registerTool(&ToolSIMInput{}) // simulated input with intelligent segmentation
	s.registerTool(&ToolBackspace{})
	s.registerTool(&ToolSetIme{})

	// Button Tools
	s.registerTool(&ToolPressButton{})
	s.registerTool(&ToolHome{}) // Home
	s.registerTool(&ToolBack{}) // Back

	// App Tools
	s.registerTool(&ToolListPackages{})     // ListPackages
	s.registerTool(&ToolLaunchApp{})        // LaunchApp
	s.registerTool(&ToolTerminateApp{})     // TerminateApp
	s.registerTool(&ToolColdLaunch{})       // ColdLaunch
	s.registerTool(&ToolAppInstall{})       // AppInstall
	s.registerTool(&ToolAppUninstall{})     // AppUninstall
	s.registerTool(&ToolAppClear{})         // AppClear
	s.registerTool(&ToolGetForegroundApp{}) // GetForegroundApp

	// Screen Tools
	s.registerTool(&ToolScreenShot{})
	s.registerTool(&ToolScreenRecord{})
	s.registerTool(&ToolGetScreenSize{})
	s.registerTool(&ToolGetSource{})

	// Media Album Tools
	s.registerTool(&ToolPushAlbums{})
	s.registerTool(&ToolClearAlbums{})

	// Utility Tools
	s.registerTool(&ToolSleep{})
	s.registerTool(&ToolSleepMS{})
	s.registerTool(&ToolSleepRandom{})
	s.registerTool(&ToolClosePopups{})

	// PC/Web Tools
	s.registerTool(&ToolWebLoginNoneUI{})
	s.registerTool(&ToolSecondaryClick{})
	s.registerTool(&ToolHoverBySelector{})
	s.registerTool(&ToolTapBySelector{})
	s.registerTool(&ToolSecondaryClickBySelector{})
	s.registerTool(&ToolWebCloseTab{})

	// AI Tools
	s.registerTool(&ToolStartToGoal{})
	s.registerTool(&ToolAIAction{})
	s.registerTool(&ToolAIQuery{})
	s.registerTool(&ToolFinished{})
}

func (s *MCPServer4XTDriver) registerTool(tool ActionTool) {
	options := []mcp.ToolOption{
		mcp.WithDescription(tool.Description()),
	}
	options = append(options, tool.Options()...)

	toolName := string(tool.Name())
	mcpTool := mcp.NewTool(toolName, options...)
	s.mcpServer.AddTool(mcpTool, tool.Implement())

	s.mcpTools = append(s.mcpTools, mcpTool)
	s.actionToolMap[tool.Name()] = tool

	log.Debug().Str("name", toolName).Msg("register tool")
}

// ActionTool interface defines the contract for MCP tools
type ActionTool interface {
	Name() option.ActionName
	Description() string
	Options() []mcp.ToolOption
	Implement() server.ToolHandlerFunc
	// ConvertActionToCallToolRequest converts MobileAction to mcp.CallToolRequest
	ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error)
}

// BuildMCPCallToolRequest is a helper function to build mcp.CallToolRequest
func BuildMCPCallToolRequest(toolName option.ActionName, arguments map[string]any, action option.MobileAction) mcp.CallToolRequest {
	// Automatically extract action options and add them to arguments
	extractActionOptionsToArguments(action.GetOptions(), arguments)

	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      string(toolName),
			Arguments: arguments,
		},
	}
}

// extractActionOptionsToArguments extracts action options and adds them to arguments map
// This is a generic helper that can be used by multiple tools
func extractActionOptionsToArguments(actionOptions []option.ActionOption, arguments map[string]any) {
	if len(actionOptions) == 0 {
		return
	}

	// Apply all options to a temporary ActionOptions to extract values
	tempOptions := &option.ActionOptions{}
	for _, opt := range actionOptions {
		opt(tempOptions)
	}

	// Define option mappings for common boolean options
	booleanOptions := map[string]bool{
		"ignore_NotFoundError": tempOptions.IgnoreNotFoundError,
		"regex":                tempOptions.Regex,
		"tap_random_rect":      tempOptions.TapRandomRect,
		"anti_risk":            tempOptions.AntiRisk,
		"pre_mark_operation":   tempOptions.PreMarkOperation,
		"reset_history":        tempOptions.ResetHistory,
		"match_one":            tempOptions.MatchOne,
	}

	// Add boolean options only if they are true
	for key, value := range booleanOptions {
		if value {
			arguments[key] = true
		}
	}

	// Add numeric options only if they have meaningful values and don't already exist
	if tempOptions.MaxRetryTimes > 0 {
		arguments["max_retry_times"] = tempOptions.MaxRetryTimes
	}
	if tempOptions.Index != 0 {
		arguments["index"] = tempOptions.Index
	}
	if tempOptions.Interval > 0 {
		arguments["interval"] = tempOptions.Interval
	}
	if tempOptions.Steps > 0 {
		arguments["steps"] = tempOptions.Steps
	}
	if tempOptions.Timeout > 0 {
		arguments["timeout"] = tempOptions.Timeout
	}
	if tempOptions.Frequency > 0 {
		arguments["frequency"] = tempOptions.Frequency
	}
	// Only set duration if it's not already set (to avoid overriding tool-specific conversions)
	if tempOptions.Duration > 0 {
		if _, exists := arguments["duration"]; !exists {
			arguments["duration"] = tempOptions.Duration
		}
	}
	if tempOptions.PressDuration > 0 {
		arguments["press_duration"] = tempOptions.PressDuration
	}

	// Add AI service options
	if tempOptions.LLMService != "" {
		arguments["llm_service"] = tempOptions.LLMService
	}
	if tempOptions.CVService != "" {
		arguments["cv_service"] = tempOptions.CVService
	}

	// Add UI/CV related options
	if len(tempOptions.ScreenShotWithUITypes) > 0 {
		arguments["screenshot_with_ui_types"] = tempOptions.ScreenShotWithUITypes
	}
	if len(tempOptions.Scope) == 4 {
		arguments["scope"] = tempOptions.Scope
	}
	if len(tempOptions.AbsScope) == 4 {
		arguments["abs_scope"] = tempOptions.AbsScope
	}

	// Add other screenshot options
	if tempOptions.ScreenShotWithOCR {
		arguments["screenshot_with_ocr"] = true
	}
	if tempOptions.ScreenShotWithUpload {
		arguments["screenshot_with_upload"] = true
	}
	if tempOptions.ScreenShotWithLiveType {
		arguments["screenshot_with_live_type"] = true
	}
	if tempOptions.ScreenShotWithLivePopularity {
		arguments["screenshot_with_live_popularity"] = true
	}
	if tempOptions.ScreenShotWithClosePopups {
		arguments["screenshot_with_close_popups"] = true
	}
	if tempOptions.ScreenShotWithOCRCluster != "" {
		arguments["screenshot_with_ocr_cluster"] = tempOptions.ScreenShotWithOCRCluster
	}
	if tempOptions.ScreenShotFileName != "" {
		arguments["screenshot_file_name"] = tempOptions.ScreenShotFileName
	}

	// Add tap/swipe offset options
	if len(tempOptions.TapOffset) == 2 {
		arguments["offset"] = tempOptions.TapOffset
	}
	if len(tempOptions.SwipeOffset) == 4 {
		arguments["swipe_offset"] = tempOptions.SwipeOffset
	}
	if len(tempOptions.OffsetRandomRange) == 2 {
		arguments["offset_random_range"] = tempOptions.OffsetRandomRange
	}

	// Add string options
	if tempOptions.Text != "" {
		arguments["text"] = tempOptions.Text
	}
	if tempOptions.ImagePath != "" {
		arguments["image_path"] = tempOptions.ImagePath
	}
	if tempOptions.AppName != "" {
		arguments["app_name"] = tempOptions.AppName
	}
	if tempOptions.PackageName != "" {
		arguments["package_name"] = tempOptions.PackageName
	}
	if tempOptions.Selector != "" {
		arguments["selector"] = tempOptions.Selector
	}
	if tempOptions.Identifier != "" {
		arguments["identifier"] = tempOptions.Identifier
	}

	// Add direction option (can be string or []float64)
	if tempOptions.Direction != nil {
		arguments["direction"] = tempOptions.Direction
	}

	// Add custom options
	if len(tempOptions.Custom) > 0 {
		arguments["custom"] = tempOptions.Custom
	}
}

// parseActionOptions converts MCP request arguments to ActionOptions struct
func parseActionOptions(arguments map[string]any) (*option.ActionOptions, error) {
	b, err := json.Marshal(arguments)
	if err != nil {
		return nil, fmt.Errorf("marshal arguments failed: %w", err)
	}

	var actionOptions option.ActionOptions
	if err := json.Unmarshal(b, &actionOptions); err != nil {
		return nil, fmt.Errorf("unmarshal to ActionOptions failed: %w", err)
	}

	return &actionOptions, nil
}

// MCPResponse represents the standard response structure for all MCP tools
type MCPResponse struct {
	Action  string `json:"action" desc:"Action performed"`
	Success bool   `json:"success" desc:"Whether the operation was successful"`
	Message string `json:"message" desc:"Human-readable message describing the result"`
}

// NewMCPSuccessResponse creates a successful response with structured data
func NewMCPSuccessResponse(message string, actionTool ActionTool) *mcp.CallToolResult {
	// Create base response with standard fields
	response := map[string]any{
		"action":  string(actionTool.Name()),
		"success": true,
		"message": message,
	}

	// Add tool-specific fields if provided
	toolData := convertToolToData(actionTool)
	for key, value := range toolData {
		response[key] = value
	}

	return marshalToMCPResult(response)
}

// convertToolToData converts tool struct to map for response
func convertToolToData(tool interface{}) map[string]any {
	data := make(map[string]any)

	// Use reflection to extract fields from the tool struct
	structValue := reflect.ValueOf(tool)
	structType := reflect.TypeOf(tool)

	// Handle pointer types
	if structType.Kind() == reflect.Ptr {
		structValue = structValue.Elem()
		structType = structType.Elem()
	}

	// Extract all fields except MCPResponse
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		fieldValue := structValue.Field(i)

		// Skip MCPResponse embedded fields
		if field.Type.Name() == "MCPResponse" {
			continue
		}

		// Get JSON tag name
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// Parse JSON tag (remove omitempty, etc.)
		jsonName := strings.Split(jsonTag, ",")[0]
		if jsonName == "" {
			jsonName = strings.ToLower(field.Name)
		}

		// Add field value to data
		if fieldValue.IsValid() && fieldValue.CanInterface() {
			data[jsonName] = fieldValue.Interface()
		}
	}

	return data
}

// NewMCPErrorResponse creates an error MCP response
func NewMCPErrorResponse(message string) *mcp.CallToolResult {
	response := map[string]any{
		"success": false,
		"message": message,
	}
	return marshalToMCPResult(response)
}

// marshalToMCPResult converts any data to mcp.CallToolResult
func marshalToMCPResult(data interface{}) *mcp.CallToolResult {
	jsonData, err := json.Marshal(data)
	if err != nil {
		// Fallback to error response if marshaling fails
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal response: %s", err.Error()))
	}
	return mcp.NewToolResultText(string(jsonData))
}

// GenerateReturnSchema generates return schema from a struct using reflection
func GenerateReturnSchema(toolStruct interface{}) map[string]string {
	schema := make(map[string]string)

	// Add standard MCPResponse fields
	schema["action"] = "string: Action performed"
	schema["success"] = "boolean: Whether the operation was successful"
	schema["message"] = "string: Human-readable message describing the result"

	// Get the type of the struct
	structType := reflect.TypeOf(toolStruct)
	if structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}

	// Iterate through all fields and add them at the same level
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		// Skip embedded MCPResponse fields
		if field.Type.Name() == "MCPResponse" {
			continue
		}

		// Get JSON tag
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// Parse JSON tag (remove omitempty, etc.)
		jsonName := strings.Split(jsonTag, ",")[0]
		if jsonName == "" {
			jsonName = strings.ToLower(field.Name)
		}

		// Get description from tag
		description := field.Tag.Get("desc")
		if description == "" {
			description = fmt.Sprintf("%s field", field.Name)
		}

		// Get field type
		fieldType := getFieldTypeString(field.Type)

		// Add to schema at the same level as standard fields
		schema[jsonName] = fmt.Sprintf("%s: %s", fieldType, description)
	}

	return schema
}

// getFieldTypeString converts Go type to string representation
func getFieldTypeString(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "int"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "uint"
	case reflect.Float32, reflect.Float64:
		return "float64"
	case reflect.Bool:
		return "boolean"
	case reflect.Slice:
		elemType := getFieldTypeString(t.Elem())
		return fmt.Sprintf("[]%s", elemType)
	case reflect.Map:
		keyType := getFieldTypeString(t.Key())
		valueType := getFieldTypeString(t.Elem())
		return fmt.Sprintf("map[%s]%s", keyType, valueType)
	case reflect.Struct:
		return "object"
	case reflect.Ptr:
		return getFieldTypeString(t.Elem())
	case reflect.Interface:
		return "interface{}"
	default:
		return t.String()
	}
}
