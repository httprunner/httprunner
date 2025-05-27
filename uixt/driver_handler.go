package uixt

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/internal/config"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rs/zerolog/log"
)

// Call custom function, used for pre/post action hook
func (dExt *XTDriver) Call(desc string, fn func(), opts ...option.ActionOption) error {
	actionOptions := option.NewActionOptions(opts...)

	startTime := time.Now()
	defer func() {
		log.Info().Str("desc", desc).
			Int64("duration(ms)", time.Since(startTime).Milliseconds()).
			Msg("function called")
	}()

	if actionOptions.Timeout == 0 {
		// wait for function to finish
		fn()
		return nil
	}

	// set timeout for function execution
	done := make(chan struct{})
	go func() {
		defer close(done)
		fn()
	}()

	select {
	case <-done:
		// function completed within timeout
		return nil
	case <-time.After(time.Duration(actionOptions.Timeout) * time.Second):
		return fmt.Errorf("function execution exceeded timeout of %d seconds", actionOptions.Timeout)
	}
}

func preHandler_TapAbsXY(driver IDriver, options *option.ActionOptions, rawX, rawY float64) (
	x, y float64, err error) {

	// Call MCP action tool if anti-risk is enabled
	if options.AntiRisk {
		callMCPActionTool(driver, option.ACTION_TapAbsXY, map[string]any{
			"x": rawX,
			"y": rawY,
		})
	}

	x, y = options.ApplyTapOffset(rawX, rawY)

	// mark UI operation
	if options.PreMarkOperation {
		if markErr := MarkUIOperation(driver, option.ACTION_TapAbsXY, []float64{x, y}); markErr != nil {
			log.Warn().Err(markErr).Msg("Failed to mark tap operation")
		}
	}

	return x, y, nil
}

func preHandler_DoubleTap(driver IDriver, options *option.ActionOptions, rawX, rawY float64) (
	x, y float64, err error) {

	x, y, err = convertToAbsolutePoint(driver, rawX, rawY)
	if err != nil {
		return 0, 0, err
	}

	x, y = options.ApplyTapOffset(x, y)

	// mark UI operation
	if options.PreMarkOperation {
		if markErr := MarkUIOperation(driver, option.ACTION_DoubleTapXY, []float64{x, y}); markErr != nil {
			log.Warn().Err(markErr).Msg("Failed to mark double tap operation")
		}
	}

	return x, y, nil
}

func preHandler_Drag(driver IDriver, options *option.ActionOptions, rawFomX, rawFromY, rawToX, rawToY float64) (
	fromX, fromY, toX, toY float64, err error) {

	fromX, fromY, toX, toY, err = convertToAbsoluteCoordinates(driver, rawFomX, rawFromY, rawToX, rawToY)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	fromX, fromY, toX, toY = options.ApplySwipeOffset(fromX, fromY, toX, toY)

	// mark UI operation
	if options.PreMarkOperation {
		if markErr := MarkUIOperation(driver, option.ACTION_Drag, []float64{fromX, fromY, toX, toY}); markErr != nil {
			log.Warn().Err(markErr).Msg("Failed to mark drag operation")
		}
	}

	return fromX, fromY, toX, toY, nil
}

func preHandler_Swipe(driver IDriver, actionType option.ActionName,
	options *option.ActionOptions, rawFomX, rawFromY, rawToX, rawToY float64) (
	fromX, fromY, toX, toY float64, err error) {

	fromX, fromY, toX, toY, err = convertToAbsoluteCoordinates(driver, rawFomX, rawFromY, rawToX, rawToY)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	fromX, fromY, toX, toY = options.ApplySwipeOffset(fromX, fromY, toX, toY)

	// save screenshot before action and mark UI operation
	if options.PreMarkOperation {
		if markErr := MarkUIOperation(driver, actionType, []float64{fromX, fromY, toX, toY}); markErr != nil {
			log.Warn().Err(markErr).Msg("Failed to mark swipe operation")
		}
	}

	return fromX, fromY, toX, toY, nil
}

func postHandler(driver IDriver, actionType option.ActionName, options *option.ActionOptions) error {
	// save screenshot after action
	if options.PostMarkOperation {
		// get compressed screenshot buffer
		compressBufSource, err := getScreenShotBuffer(driver)
		if err != nil {
			return err
		}

		// save compressed screenshot to file
		timestamp := builtin.GenNameWithTimestamp("%d")
		imagePath := filepath.Join(
			config.GetConfig().ScreenShotsPath,
			fmt.Sprintf("action_%s_post_%s.png", timestamp, actionType),
		)

		go func() {
			err := saveScreenShot(compressBufSource, imagePath)
			if err != nil {
				log.Error().Err(err).Msg("save screenshot file failed")
			}
		}()
	}
	return nil
}

// callMCPActionTool calls MCP tool for the given action
func callMCPActionTool(driver IDriver, actionType option.ActionName, arguments map[string]any) {
	// Get XTDriver from cache
	dExt := getXTDriverFromCache(driver)
	if dExt == nil {
		return
	}

	// Define action to MCP server mapping for pre-hooks
	serverMapping := getPreHookServerMapping(actionType)
	if serverMapping == nil {
		return // No MCP hook configured for this action
	}

	callMCPTool(dExt, serverMapping.ServerName, serverMapping.ToolName, arguments, actionType)
}

// MCPServerMapping defines the mapping between action and MCP server/tool
type MCPServerMapping struct {
	ServerName string
	ToolName   string
}

// getPreHookServerMapping returns MCP server mapping for pre-hooks
// TODO: You can customize these mappings according to your needs
func getPreHookServerMapping(actionType option.ActionName) *MCPServerMapping {
	mappings := map[option.ActionName]*MCPServerMapping{
		option.ACTION_TapAbsXY: {
			ServerName: "evalpkgs",
			ToolName:   "log_pre_action",
		},
		// Add more mappings as needed
		// option.ACTION_Swipe: {
		//     ServerName: "monitor",
		//     ToolName:   "start_timer",
		// },
	}
	return mappings[actionType]
}

// getXTDriverFromCache gets XTDriver from cache using device UUID
func getXTDriverFromCache(driver IDriver) *XTDriver {
	// Get device info to find the corresponding XTDriver
	device := driver.GetDevice()
	if device == nil {
		log.Warn().Msg("Cannot get device from driver for MCP hook")
		return nil
	}

	// Get device UUID (serial/udid/connectKey/browserID)
	deviceUUID := device.UUID()
	if deviceUUID == "" {
		log.Warn().Msg("Cannot get device UUID for MCP hook")
		return nil
	}

	// Get XTDriver from cache using device UUID as serial
	cachedDrivers := ListCachedDrivers()
	for _, cached := range cachedDrivers {
		if cached.Serial == deviceUUID {
			return cached.Driver
		}
	}

	log.Warn().Str("uuid", deviceUUID).
		Msg("Cannot find cached XTDriver for MCP hook")
	return nil
}

// callMCPTool calls the specified MCP tool
func callMCPTool(dExt *XTDriver, serverName, toolName string, arguments map[string]any, actionType option.ActionName) {
	// Get MCP client
	mcpClient, exists := dExt.GetMCPClient(serverName)
	if !exists {
		log.Debug().Str("server", serverName).Msg("MCP server not found for hook")
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Prepare arguments
	if arguments == nil {
		arguments = make(map[string]any)
	}
	// Add action type and hook type to arguments
	arguments["action_type"] = string(actionType)

	// Call MCP tool
	req := mcp.CallToolRequest{
		Params: struct {
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments,omitempty"`
			Meta      *struct {
				ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
			} `json:"_meta,omitempty"`
		}{
			Name:      toolName,
			Arguments: arguments,
		},
	}

	result, err := mcpClient.CallTool(ctx, req)
	if err != nil {
		log.Debug().Err(err).
			Str("server", serverName).
			Str("tool", toolName).
			Msg("MCP hook call failed")
		return
	}

	if result.IsError {
		log.Debug().
			Str("server", serverName).
			Str("tool", toolName).
			Interface("content", result.Content).
			Msg("MCP hook returned error")
		return
	}

	log.Debug().
		Str("server", serverName).
		Str("tool", toolName).
		Str("action", string(actionType)).
		Msg("MCP hook called successfully")
}
