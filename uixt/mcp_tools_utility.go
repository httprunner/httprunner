package uixt

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rs/zerolog/log"
)

// ToolSleep implements the sleep tool call.
type ToolSleep struct{}

func (t *ToolSleep) Name() option.ActionName {
	return option.ACTION_Sleep
}

func (t *ToolSleep) Description() string {
	return "Sleep for a specified number of seconds"
}

func (t *ToolSleep) Options() []mcp.ToolOption {
	return []mcp.ToolOption{
		mcp.WithNumber("seconds", mcp.Description("Number of seconds to sleep")),
	}
}

func (t *ToolSleep) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		seconds, ok := request.Params.Arguments["seconds"]
		if !ok {
			log.Warn().Msg("seconds parameter is required, using default value 5.0 seconds")
			seconds = 5.0
		}

		// Sleep action logic
		log.Info().Interface("seconds", seconds).Msg("sleeping")

		var duration time.Duration
		switch v := seconds.(type) {
		case float64:
			duration = time.Duration(v*1000) * time.Millisecond
		case int:
			duration = time.Duration(v) * time.Second
		case int64:
			duration = time.Duration(v) * time.Second
		case string:
			s, err := builtin.ConvertToFloat64(v)
			if err != nil {
				return nil, fmt.Errorf("invalid sleep duration: %v", v)
			}
			duration = time.Duration(s*1000) * time.Millisecond
		default:
			return nil, fmt.Errorf("unsupported sleep duration type: %T", v)
		}

		time.Sleep(duration)

		return mcp.NewToolResultText(fmt.Sprintf("Successfully slept for %v seconds", seconds)), nil
	}
}

func (t *ToolSleep) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	arguments := map[string]any{
		"seconds": action.Params,
	}
	return buildMCPCallToolRequest(t.Name(), arguments), nil
}

func (t *ToolSleep) ReturnSchema() map[string]string {
	return map[string]string{
		"message": "string: Success message confirming sleep operation completed",
		"seconds": "float64: Duration in seconds that was slept",
	}
}

// ToolSleepMS implements the sleep_ms tool call.
type ToolSleepMS struct{}

func (t *ToolSleepMS) Name() option.ActionName {
	return option.ACTION_SleepMS
}

func (t *ToolSleepMS) Description() string {
	return "Sleep for specified milliseconds"
}

func (t *ToolSleepMS) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_SleepMS)
}

func (t *ToolSleepMS) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// Validate required parameters
		if unifiedReq.Milliseconds == 0 {
			return nil, fmt.Errorf("milliseconds is required")
		}

		// Sleep MS action logic
		log.Info().Int64("milliseconds", unifiedReq.Milliseconds).Msg("sleeping in milliseconds")
		time.Sleep(time.Duration(unifiedReq.Milliseconds) * time.Millisecond)

		return mcp.NewToolResultText(fmt.Sprintf("Successfully slept for %d milliseconds", unifiedReq.Milliseconds)), nil
	}
}

func (t *ToolSleepMS) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	var milliseconds int64
	if param, ok := action.Params.(json.Number); ok {
		milliseconds, _ = param.Int64()
	} else if param, ok := action.Params.(int64); ok {
		milliseconds = param
	} else {
		return mcp.CallToolRequest{}, fmt.Errorf("invalid sleep ms params: %v", action.Params)
	}
	arguments := map[string]any{
		"milliseconds": milliseconds,
	}
	return buildMCPCallToolRequest(t.Name(), arguments), nil
}

func (t *ToolSleepMS) ReturnSchema() map[string]string {
	return map[string]string{
		"message":      "string: Success message confirming sleep operation completed",
		"milliseconds": "int64: Duration in milliseconds that was slept",
	}
}

// ToolSleepRandom implements the sleep_random tool call.
type ToolSleepRandom struct{}

func (t *ToolSleepRandom) Name() option.ActionName {
	return option.ACTION_SleepRandom
}

func (t *ToolSleepRandom) Description() string {
	return "Sleep for a random duration based on parameters"
}

func (t *ToolSleepRandom) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_SleepRandom)
}

func (t *ToolSleepRandom) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// Sleep random action logic
		sleepStrict(time.Now(), getSimulationDuration(unifiedReq.Params))

		return mcp.NewToolResultText(fmt.Sprintf("Successfully slept for random duration with params: %v", unifiedReq.Params)), nil
	}
}

func (t *ToolSleepRandom) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	if params, err := builtin.ConvertToFloat64Slice(action.Params); err == nil {
		arguments := map[string]any{
			"params": params,
		}
		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid sleep random params: %v", action.Params)
}

func (t *ToolSleepRandom) ReturnSchema() map[string]string {
	return map[string]string{
		"message":        "string: Success message confirming random sleep operation completed",
		"params":         "[]float64: Parameters used for random duration calculation",
		"actualDuration": "float64: Actual duration that was slept (in seconds)",
	}
}

// ToolClosePopups implements the close_popups tool call.
type ToolClosePopups struct{}

func (t *ToolClosePopups) Name() option.ActionName {
	return option.ACTION_ClosePopups
}

func (t *ToolClosePopups) Description() string {
	return "Close any popup windows or dialogs on screen"
}

func (t *ToolClosePopups) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_ClosePopups)
}

func (t *ToolClosePopups) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		// Close popups action logic
		err = driverExt.ClosePopupsHandler()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Close popups failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText("Successfully closed popups"), nil
	}
}

func (t *ToolClosePopups) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	return buildMCPCallToolRequest(t.Name(), map[string]any{}), nil
}

func (t *ToolClosePopups) ReturnSchema() map[string]string {
	return map[string]string{
		"message":      "string: Success message confirming popups were closed",
		"popupsClosed": "int: Number of popup windows or dialogs that were closed",
	}
}
