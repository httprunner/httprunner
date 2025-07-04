package uixt

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/uixt/option"
)

// ToolSleep implements the sleep tool call.
type ToolSleep struct {
	// Return data fields - these define the structure of data returned by this tool
	Seconds  float64 `json:"seconds" desc:"Duration in seconds that was slept"`
	Duration string  `json:"duration" desc:"Human-readable duration string"`
}

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
		seconds, ok := request.GetArguments()["seconds"]
		if !ok {
			log.Warn().Msg("seconds parameter is required, using default value 5.0 seconds")
			seconds = 5.0
		}

		// Sleep action logic
		log.Info().Interface("seconds", seconds).Msg("sleeping")

		var duration time.Duration
		var actualSeconds float64
		switch v := seconds.(type) {
		case float64:
			actualSeconds = v
			duration = time.Duration(v*1000) * time.Millisecond
		case int:
			actualSeconds = float64(v)
			duration = time.Duration(v) * time.Second
		case int64:
			actualSeconds = float64(v)
			duration = time.Duration(v) * time.Second
		case string:
			s, err := builtin.ConvertToFloat64(v)
			if err != nil {
				return nil, fmt.Errorf("invalid sleep duration: %v", v)
			}
			actualSeconds = s
			duration = time.Duration(s*1000) * time.Millisecond
		default:
			return nil, fmt.Errorf("unsupported sleep duration type: %T", v)
		}

		// Use context-aware sleep instead of blocking time.Sleep
		select {
		case <-time.After(duration):
			// Normal completion
		case <-ctx.Done():
			// Interrupted by context cancellation (interrupt signal, timeout, time limit)
			log.Info().Msg("sleep interrupted by context cancellation")
			// Don't return error - let the upper layer handle timeout/time limit logic
		}

		message := fmt.Sprintf("Successfully slept for %v seconds", actualSeconds)
		returnData := ToolSleep{
			Seconds:  actualSeconds,
			Duration: duration.String(),
		}

		return NewMCPSuccessResponse(message, &returnData), nil
	}
}

func (t *ToolSleep) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	arguments := map[string]any{
		"seconds": action.Params,
	}
	return BuildMCPCallToolRequest(t.Name(), arguments, action), nil
}

// ToolSleepMS implements the sleep_ms tool call.
type ToolSleepMS struct {
	// Return data fields - these define the structure of data returned by this tool
	Milliseconds int64  `json:"milliseconds" desc:"Duration in milliseconds that was slept"`
	Duration     string `json:"duration" desc:"Human-readable duration string"`
}

func (t *ToolSleepMS) Name() option.ActionName {
	return option.ACTION_SleepMS
}

func (t *ToolSleepMS) Description() string {
	return "Sleep for specified milliseconds"
}

func (t *ToolSleepMS) Options() []mcp.ToolOption {
	return []mcp.ToolOption{
		mcp.WithNumber("milliseconds", mcp.Description("Number of milliseconds to sleep")),
	}
}

func (t *ToolSleepMS) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		milliseconds, ok := request.GetArguments()["milliseconds"]
		if !ok {
			log.Warn().Msg("milliseconds parameter is required, using default value 1000 milliseconds")
			milliseconds = 1000
		}

		// Sleep MS action logic
		log.Info().Interface("milliseconds", milliseconds).Msg("sleeping in milliseconds")

		var duration time.Duration
		var actualMilliseconds int64
		switch v := milliseconds.(type) {
		case float64:
			actualMilliseconds = int64(v)
			duration = time.Duration(v) * time.Millisecond
		case int:
			actualMilliseconds = int64(v)
			duration = time.Duration(v) * time.Millisecond
		case int64:
			actualMilliseconds = v
			duration = time.Duration(v) * time.Millisecond
		case string:
			ms, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid sleep duration: %v", v)
			}
			actualMilliseconds = ms
			duration = time.Duration(ms) * time.Millisecond
		default:
			return nil, fmt.Errorf("unsupported sleep duration type: %T", v)
		}

		// Use context-aware sleep instead of blocking time.Sleep
		select {
		case <-time.After(duration):
			// Normal completion
		case <-ctx.Done():
			// Interrupted by context cancellation (interrupt signal, timeout, time limit)
			log.Info().Msg("sleep interrupted by context cancellation")
			// Don't return error - let the upper layer handle timeout/time limit logic
		}

		message := fmt.Sprintf("Successfully slept for %d milliseconds", actualMilliseconds)
		returnData := ToolSleepMS{
			Milliseconds: actualMilliseconds,
			Duration:     duration.String(),
		}

		return NewMCPSuccessResponse(message, &returnData), nil
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
	return BuildMCPCallToolRequest(t.Name(), arguments, action), nil
}

// ToolSleepRandom implements the sleep_random tool call.
type ToolSleepRandom struct {
	// Return data fields - these define the structure of data returned by this tool
	Params []float64 `json:"params" desc:"Random sleep parameters used"`
}

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
		unifiedReq, err := parseActionOptions(request.GetArguments())
		if err != nil {
			return nil, err
		}

		// Sleep random action logic with context support
		sleepStrict(ctx, time.Now(), getSimulationDuration(unifiedReq.Params))

		message := fmt.Sprintf("Successfully slept for random duration with params: %v", unifiedReq.Params)
		returnData := ToolSleepRandom{Params: unifiedReq.Params}

		return NewMCPSuccessResponse(message, &returnData), nil
	}
}

func (t *ToolSleepRandom) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	if params, err := builtin.ConvertToFloat64Slice(action.Params); err == nil {
		arguments := map[string]any{
			"params": params,
		}
		return BuildMCPCallToolRequest(t.Name(), arguments, action), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid sleep random params: %v", action.Params)
}

// ToolClosePopups implements the close_popups tool call.
type ToolClosePopups struct { // Return data fields - these define the structure of data returned by this tool
}

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
		driverExt, err := setupXTDriver(ctx, request.GetArguments())
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		// Close popups action logic
		err = driverExt.ClosePopupsHandler()
		if err != nil {
			return NewMCPErrorResponse(fmt.Sprintf("Close popups failed: %s", err.Error())), err
		}

		message := "Successfully closed popups"
		returnData := ToolClosePopups{}

		return NewMCPSuccessResponse(message, &returnData), nil
	}
}

func (t *ToolClosePopups) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	return BuildMCPCallToolRequest(t.Name(), map[string]any{}, action), nil
}
