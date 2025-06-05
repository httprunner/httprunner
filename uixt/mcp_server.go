package uixt

import (
	"encoding/json"
	"fmt"

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
	return s.actionToolMap[actionMethod]
}

// registerTools registers all MCP tools.
func (s *MCPServer4XTDriver) registerTools() {
	// Device Tool
	s.registerTool(&ToolListAvailableDevices{}) // ListAvailableDevices
	s.registerTool(&ToolSelectDevice{})         // SelectDevice

	// Touch Tools
	s.registerTool(&ToolTapXY{})       // tap xy
	s.registerTool(&ToolTapAbsXY{})    // tap abs xy
	s.registerTool(&ToolTapByOCR{})    // tap by OCR
	s.registerTool(&ToolTapByCV{})     // tap by CV
	s.registerTool(&ToolDoubleTapXY{}) // double tap xy

	// Swipe Tools
	s.registerTool(&ToolSwipe{})           // generic swipe, auto-detect direction or coordinate
	s.registerTool(&ToolSwipeDirection{})  // swipe direction, up/down/left/right
	s.registerTool(&ToolSwipeCoordinate{}) // swipe coordinate, [fromX, fromY, toX, toY]
	s.registerTool(&ToolSwipeToTapApp{})
	s.registerTool(&ToolSwipeToTapText{})
	s.registerTool(&ToolSwipeToTapTexts{})
	s.registerTool(&ToolDrag{})

	// Input Tools
	s.registerTool(&ToolInput{})
	s.registerTool(&ToolSetIme{})

	// Button Tools
	s.registerTool(&ToolPressButton{})
	s.registerTool(&ToolHome{}) // Home
	s.registerTool(&ToolBack{}) // Back

	// App Tools
	s.registerTool(&ToolListPackages{}) // ListPackages
	s.registerTool(&ToolLaunchApp{})    // LaunchApp
	s.registerTool(&ToolTerminateApp{}) // TerminateApp
	s.registerTool(&ToolAppInstall{})   // AppInstall
	s.registerTool(&ToolAppUninstall{}) // AppUninstall
	s.registerTool(&ToolAppClear{})     // AppClear

	// Screen Tools
	s.registerTool(&ToolScreenShot{})
	s.registerTool(&ToolGetScreenSize{})
	s.registerTool(&ToolGetSource{})

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

	log.Debug().Str("name", toolName).Str("type", toolName).Msg("register tool")
}

// ActionTool interface defines the contract for MCP tools
type ActionTool interface {
	Name() option.ActionName
	Description() string
	Options() []mcp.ToolOption
	Implement() server.ToolHandlerFunc
	// ConvertActionToCallToolRequest converts MobileAction to mcp.CallToolRequest
	ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error)
	// ReturnSchema returns the expected return value schema based on mcp.CallToolResult conventions
	ReturnSchema() map[string]string
}

// buildMCPCallToolRequest is a helper function to build mcp.CallToolRequest
func buildMCPCallToolRequest(toolName option.ActionName, arguments map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: struct {
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments,omitempty"`
			Meta      *struct {
				ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
			} `json:"_meta,omitempty"`
		}{
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
}

func getFloat64ValueOrDefault(value float64, defaultValue float64) float64 {
	if value == 0 {
		return defaultValue
	}
	return value
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
