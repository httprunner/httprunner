package uixt

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/uixt/option"
)

// ToolTapXY implements the tap_xy tool call.
type ToolTapXY struct {
	// Return data fields - these define the structure of data returned by this tool
	X float64 `json:"x" desc:"X coordinate where tap was performed"`
	Y float64 `json:"y" desc:"Y coordinate where tap was performed"`
}

func (t *ToolTapXY) Name() option.ActionName {
	return option.ACTION_TapXY
}

func (t *ToolTapXY) Description() string {
	return "Tap on the screen at given relative coordinates (0.0-1.0 range)"
}

func (t *ToolTapXY) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_TapXY)
}

func (t *ToolTapXY) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments := request.GetArguments()
		driverExt, err := setupXTDriver(ctx, arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(arguments)
		if err != nil {
			return nil, err
		}

		// Build all options from request arguments
		opts := unifiedReq.Options()

		// Validate required parameters
		if unifiedReq.X == 0 || unifiedReq.Y == 0 {
			return nil, fmt.Errorf("x and y coordinates are required")
		}

		// Tap action logic
		err = driverExt.TapXY(unifiedReq.X, unifiedReq.Y, opts...)
		if err != nil {
			return NewMCPErrorResponse(fmt.Sprintf("Tap failed: %s", err.Error())), err
		}

		message := fmt.Sprintf("Successfully tapped at coordinates (%.2f, %.2f)", unifiedReq.X, unifiedReq.Y)
		returnData := ToolTapXY{
			X: unifiedReq.X,
			Y: unifiedReq.Y,
		}

		return NewMCPSuccessResponse(message, &returnData), nil
	}
}

func (t *ToolTapXY) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	if params, err := builtin.ConvertToFloat64Slice(action.Params); err == nil && len(params) == 2 {
		x, y := params[0], params[1]
		arguments := map[string]any{
			"x": x,
			"y": y,
		}
		// Add duration if available from action options
		if duration := action.ActionOptions.Duration; duration > 0 {
			arguments["duration"] = duration
		}
		return BuildMCPCallToolRequest(t.Name(), arguments, action), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid tap params: %v", action.Params)
}

// ToolTapAbsXY implements the tap_abs_xy tool call.
type ToolTapAbsXY struct {
	// Return data fields - these define the structure of data returned by this tool
	X float64 `json:"x" desc:"X coordinate where tap was performed (absolute pixels)"`
	Y float64 `json:"y" desc:"Y coordinate where tap was performed (absolute pixels)"`
}

func (t *ToolTapAbsXY) Name() option.ActionName {
	return option.ACTION_TapAbsXY
}

func (t *ToolTapAbsXY) Description() string {
	return "Tap at absolute pixel coordinates on the screen"
}

func (t *ToolTapAbsXY) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_TapAbsXY)
}

func (t *ToolTapAbsXY) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments := request.GetArguments()
		driverExt, err := setupXTDriver(ctx, arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(arguments)
		if err != nil {
			return nil, err
		}

		// Build all options from request arguments
		opts := unifiedReq.Options()

		// Validate required parameters
		if unifiedReq.X == 0 || unifiedReq.Y == 0 {
			return nil, fmt.Errorf("x and y coordinates are required")
		}

		// Tap absolute XY action logic
		err = driverExt.TapAbsXY(unifiedReq.X, unifiedReq.Y, opts...)
		if err != nil {
			return NewMCPErrorResponse(fmt.Sprintf("Tap absolute XY failed: %s", err.Error())), err
		}

		message := fmt.Sprintf("Successfully tapped at absolute coordinates (%.0f, %.0f)", unifiedReq.X, unifiedReq.Y)
		returnData := ToolTapAbsXY{
			X: unifiedReq.X,
			Y: unifiedReq.Y,
		}

		return NewMCPSuccessResponse(message, &returnData), nil
	}
}

func (t *ToolTapAbsXY) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	if params, err := builtin.ConvertToFloat64Slice(action.Params); err == nil && len(params) == 2 {
		x, y := params[0], params[1]
		arguments := map[string]any{
			"x": x,
			"y": y,
		}
		// Add duration if available
		if duration := action.ActionOptions.Duration; duration > 0 {
			arguments["duration"] = duration
		}
		return BuildMCPCallToolRequest(t.Name(), arguments, action), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid tap abs params: %v", action.Params)
}

// ToolTapByOCR implements the tap_ocr tool call.
type ToolTapByOCR struct {
	// Return data fields - these define the structure of data returned by this tool
	Text string `json:"text" desc:"Text that was tapped by OCR"`
}

func (t *ToolTapByOCR) Name() option.ActionName {
	return option.ACTION_TapByOCR
}

func (t *ToolTapByOCR) Description() string {
	return "Tap on text found by OCR recognition"
}

func (t *ToolTapByOCR) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_TapByOCR)
}

func (t *ToolTapByOCR) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments := request.GetArguments()
		driverExt, err := setupXTDriver(ctx, arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(arguments)
		if err != nil {
			return nil, err
		}

		// Build all options from request arguments
		opts := unifiedReq.Options()

		// Validate required parameters
		if unifiedReq.Text == "" {
			return nil, fmt.Errorf("text parameter is required")
		}

		// Tap by OCR action logic
		err = driverExt.TapByOCR(unifiedReq.Text, opts...)
		if err != nil {
			return NewMCPErrorResponse(fmt.Sprintf("Tap by OCR failed: %s", err.Error())), err
		}

		message := fmt.Sprintf("Successfully tapped on OCR text: %s", unifiedReq.Text)
		returnData := ToolTapByOCR{Text: unifiedReq.Text}

		return NewMCPSuccessResponse(message, &returnData), nil
	}
}

func (t *ToolTapByOCR) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	if text, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"text": text,
		}
		return BuildMCPCallToolRequest(t.Name(), arguments, action), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid tap by OCR params: %v", action.Params)
}

// ToolTapByCV implements the tap_cv tool call.
type ToolTapByCV struct { // Return data fields - these define the structure of data returned by this tool
}

func (t *ToolTapByCV) Name() option.ActionName {
	return option.ACTION_TapByCV
}

func (t *ToolTapByCV) Description() string {
	return "Tap on element found by computer vision"
}

func (t *ToolTapByCV) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_TapByCV)
}

func (t *ToolTapByCV) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments := request.GetArguments()
		driverExt, err := setupXTDriver(ctx, arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(arguments)
		if err != nil {
			return nil, err
		}

		// Build all options from request arguments
		opts := unifiedReq.Options()

		// For TapByCV, we need to check if there are UI types in the options
		// In the original DoAction, it requires ScreenShotWithUITypes to be set
		// We'll add a basic implementation that triggers CV recognition
		err = driverExt.TapByCV(opts...)
		if err != nil {
			return NewMCPErrorResponse(fmt.Sprintf("Tap by CV failed: %s", err.Error())), err
		}

		message := "Successfully tapped by computer vision"
		returnData := ToolTapByCV{}

		return NewMCPSuccessResponse(message, &returnData), nil
	}
}

func (t *ToolTapByCV) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	// For TapByCV, the original action might not have params but relies on options
	arguments := map[string]any{
		"imagePath": "", // Will be handled by the tool based on UI types
	}
	return BuildMCPCallToolRequest(t.Name(), arguments, action), nil
}

// ToolDoubleTapXY implements the double_tap_xy tool call.
type ToolDoubleTapXY struct {
	// Return data fields - these define the structure of data returned by this tool
	X float64 `json:"x" desc:"X coordinate where double tap was performed"`
	Y float64 `json:"y" desc:"Y coordinate where double tap was performed"`
}

func (t *ToolDoubleTapXY) Name() option.ActionName {
	return option.ACTION_DoubleTapXY
}

func (t *ToolDoubleTapXY) Description() string {
	return "Double tap at given relative coordinates (0.0-1.0 range)"
}

func (t *ToolDoubleTapXY) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_DoubleTapXY)
}

func (t *ToolDoubleTapXY) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments := request.GetArguments()
		driverExt, err := setupXTDriver(ctx, arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(arguments)
		if err != nil {
			return nil, err
		}

		// Validate required parameters
		if unifiedReq.X == 0 || unifiedReq.Y == 0 {
			return nil, fmt.Errorf("x and y coordinates are required")
		}

		// Double tap XY action logic
		err = driverExt.DoubleTap(unifiedReq.X, unifiedReq.Y)
		if err != nil {
			return NewMCPErrorResponse(fmt.Sprintf("Double tap failed: %s", err.Error())), err
		}

		message := fmt.Sprintf("Successfully double tapped at (%.2f, %.2f)", unifiedReq.X, unifiedReq.Y)
		returnData := ToolDoubleTapXY{
			X: unifiedReq.X,
			Y: unifiedReq.Y,
		}

		return NewMCPSuccessResponse(message, &returnData), nil
	}
}

func (t *ToolDoubleTapXY) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	if params, err := builtin.ConvertToFloat64Slice(action.Params); err == nil && len(params) == 2 {
		x, y := params[0], params[1]
		arguments := map[string]any{
			"x": x,
			"y": y,
		}
		return BuildMCPCallToolRequest(t.Name(), arguments, action), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid double tap params: %v", action.Params)
}

// ToolSIMClickAtPoint implements the sim_click_at_point tool call.
type ToolSIMClickAtPoint struct {
	// Return data fields - these define the structure of data returned by this tool
	X float64 `json:"x" desc:"X coordinate where simulated click was performed"`
	Y float64 `json:"y" desc:"Y coordinate where simulated click was performed"`
}

func (t *ToolSIMClickAtPoint) Name() option.ActionName {
	return option.ACTION_SIMClickAtPoint
}

func (t *ToolSIMClickAtPoint) Description() string {
	return "Perform simulated click at specified point with human-like touch patterns"
}

func (t *ToolSIMClickAtPoint) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_SIMClickAtPoint)
}

func (t *ToolSIMClickAtPoint) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments := request.GetArguments()
		driverExt, err := setupXTDriver(ctx, arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(arguments)
		if err != nil {
			return nil, err
		}

		// Validate required parameters
		if unifiedReq.X == 0 || unifiedReq.Y == 0 {
			return nil, fmt.Errorf("x and y coordinates are required")
		}

		x := unifiedReq.X
		y := unifiedReq.Y

		log.Info().
			Float64("x", x).
			Float64("y", y).
			Msg("performing simulated click at point")

		// Build all options from request arguments
		opts := unifiedReq.Options()

		// Call the underlying SIMClickAtPoint method (check if driver supports SIM)
		if simDriver, ok := driverExt.IDriver.(SIMSupport); ok {
			err = simDriver.SIMClickAtPoint(x, y, opts...)
			if err != nil {
				return NewMCPErrorResponse(fmt.Sprintf("Simulated click failed: %s", err.Error())), err
			}
		} else {
			return NewMCPErrorResponse("SIMClickAtPoint is not supported by the current driver"), fmt.Errorf("driver does not implement SIMSupport interface")
		}

		message := fmt.Sprintf("Successfully performed simulated click at (%.2f, %.2f)", x, y)
		returnData := ToolSIMClickAtPoint{
			X: x,
			Y: y,
		}

		return NewMCPSuccessResponse(message, &returnData), nil
	}
}

func (t *ToolSIMClickAtPoint) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	// Handle params as map[string]interface{}
	if paramsMap, ok := action.Params.(map[string]interface{}); ok {
		arguments := map[string]any{}

		// Extract coordinates
		if x, exists := paramsMap["x"]; exists {
			arguments["x"] = x
		}
		if y, exists := paramsMap["y"]; exists {
			arguments["y"] = y
		}

		// Add duration from options
		if duration := action.ActionOptions.Duration; duration > 0 {
			arguments["duration"] = duration
		}

		return BuildMCPCallToolRequest(t.Name(), arguments, action), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid SIM click at point params: %v", action.Params)
}
