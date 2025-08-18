package uixt

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/uixt/option"
)

// extractStartTimeMs extracts start_time_ms from MCP request arguments
// Returns time.Time (zero if not provided) and any conversion error
func extractStartTimeMs(request mcp.CallToolRequest) (time.Time, error) {
	startTimeMs, ok := request.GetArguments()["start_time_ms"]
	if !ok || startTimeMs == nil {
		return time.Time{}, nil // Return zero time for normal sleep
	}

	var ms int64
	switch v := startTimeMs.(type) {
	case float64:
		ms = int64(v)
	case int64:
		ms = v
	case int:
		ms = int64(v)
	default:
		return time.Time{}, fmt.Errorf("invalid start_time_ms type: %T", v)
	}

	return time.UnixMilli(ms), nil
}

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
		mcp.WithNumber("start_time_ms", mcp.Description("Start time as Unix milliseconds for strict sleep calculation")),
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

		// Use Interface2Float64 for unified type conversion
		actualSeconds, err := builtin.Interface2Float64(seconds)
		if err != nil {
			return nil, fmt.Errorf("invalid sleep duration: %v", seconds)
		}
		duration := time.Duration(actualSeconds) * time.Second

		// Extract start_time_ms and use sleepStrict for unified sleep logic
		startTime, err := extractStartTimeMs(request)
		if err != nil {
			return nil, err
		}

		milliseconds := int64(actualSeconds * 1000)
		sleepStrict(ctx, startTime, milliseconds)

		message := fmt.Sprintf("Successfully slept for %v seconds", actualSeconds)
		returnData := ToolSleep{
			Seconds:  actualSeconds,
			Duration: duration.String(),
		}

		return NewMCPSuccessResponse(message, &returnData), nil
	}
}

func (t *ToolSleep) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	arguments := map[string]any{}

	var seconds float64
	if sleepConfig, ok := action.Params.(SleepConfig); ok {
		// When startTime is provided, pass both seconds and startTime
		seconds = sleepConfig.Seconds
		arguments["seconds"] = seconds
		arguments["start_time_ms"] = sleepConfig.StartTime.UnixMilli()
	} else {
		// Use builtin.Interface2Float64 for unified parameter handling
		var err error
		seconds, err = builtin.Interface2Float64(action.Params)
		if err != nil {
			return mcp.CallToolRequest{}, fmt.Errorf("invalid sleep params: %v", action.Params)
		}
		arguments["seconds"] = seconds
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
		mcp.WithNumber("start_time_ms", mcp.Description("Start time as Unix milliseconds for strict sleep calculation")),
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

		// Use Interface2Float64 for unified type conversion, then convert to int64
		floatVal, err := builtin.Interface2Float64(milliseconds)
		if err != nil {
			return nil, fmt.Errorf("invalid sleep duration: %v", milliseconds)
		}
		actualMilliseconds := int64(floatVal)
		duration := time.Duration(actualMilliseconds) * time.Millisecond

		// Extract start_time_ms and use sleepStrict for unified sleep logic
		startTime, err := extractStartTimeMs(request)
		if err != nil {
			return nil, err
		}

		sleepStrict(ctx, startTime, actualMilliseconds)

		message := fmt.Sprintf("Successfully slept for %d milliseconds", actualMilliseconds)
		returnData := ToolSleepMS{
			Milliseconds: actualMilliseconds,
			Duration:     duration.String(),
		}

		return NewMCPSuccessResponse(message, &returnData), nil
	}
}

func (t *ToolSleepMS) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	arguments := map[string]any{}

	var milliseconds int64
	if sleepConfig, ok := action.Params.(SleepConfig); ok {
		// When startTime is provided, pass both milliseconds and startTime
		milliseconds = sleepConfig.Milliseconds
		arguments["milliseconds"] = milliseconds
		arguments["start_time_ms"] = sleepConfig.StartTime.UnixMilli()
	} else {
		// Use builtin.Interface2Float64 for unified parameter handling, then convert to int64
		floatVal, err := builtin.Interface2Float64(action.Params)
		if err != nil {
			return mcp.CallToolRequest{}, fmt.Errorf("invalid sleep ms params: %v", action.Params)
		}
		milliseconds = int64(floatVal)
		arguments["milliseconds"] = milliseconds
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
