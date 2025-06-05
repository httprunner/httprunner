package uixt

import (
	"context"
	"fmt"

	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rs/zerolog/log"
)

// ToolTapXY implements the tap_xy tool call.
type ToolTapXY struct{}

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
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// Get options directly since ActionOptions is now ActionOptions
		opts := unifiedReq.Options()

		// Add configurable options based on request
		if unifiedReq.PreMarkOperation {
			opts = append(opts, option.WithPreMarkOperation(true))
		}

		// Validate required parameters
		if unifiedReq.X == 0 || unifiedReq.Y == 0 {
			return nil, fmt.Errorf("x and y coordinates are required")
		}

		// Tap action logic
		err = driverExt.TapXY(unifiedReq.X, unifiedReq.Y, opts...)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Tap failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully tapped at coordinates (%.2f, %.2f)", unifiedReq.X, unifiedReq.Y)), nil
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

		// Extract options to arguments
		extractActionOptionsToArguments(action.GetOptions(), arguments)

		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid tap params: %v", action.Params)
}

func (t *ToolTapXY) ReturnSchema() map[string]string {
	return map[string]string{
		"message": "string: Success message confirming tap operation at specified coordinates",
	}
}

// ToolTapAbsXY implements the tap_abs_xy tool call.
type ToolTapAbsXY struct{}

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
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// Get options directly since ActionOptions is now ActionOptions
		opts := unifiedReq.Options()

		// Add configurable options based on request
		if unifiedReq.PreMarkOperation {
			opts = append(opts, option.WithPreMarkOperation(true))
		}

		// Add AntiRisk support
		if unifiedReq.AntiRisk {
			opts = append(opts, option.WithAntiRisk(true))
		}

		// Validate required parameters
		if unifiedReq.X == 0 || unifiedReq.Y == 0 {
			return nil, fmt.Errorf("x and y coordinates are required")
		}

		// Tap absolute XY action logic
		log.Info().Float64("x", unifiedReq.X).Float64("y", unifiedReq.Y).Msg("tapping at absolute coordinates")

		err = driverExt.TapAbsXY(unifiedReq.X, unifiedReq.Y, opts...)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Tap absolute XY failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully tapped at absolute coordinates (%.0f, %.0f)", unifiedReq.X, unifiedReq.Y)), nil
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

		// Extract options to arguments
		extractActionOptionsToArguments(action.GetOptions(), arguments)

		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid tap abs params: %v", action.Params)
}

func (t *ToolTapAbsXY) ReturnSchema() map[string]string {
	return map[string]string{
		"message": "string: Success message confirming tap operation at absolute coordinates",
	}
}

// defaultReturnSchema provides a standard return schema for most tools
func defaultReturnSchema() map[string]string {
	return map[string]string{
		"message": "string: Success message confirming the operation was completed",
	}
}

// ToolTapByOCR implements the tap_ocr tool call.
type ToolTapByOCR struct{}

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
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// Get options directly since ActionOptions is now ActionOptions
		opts := unifiedReq.Options()

		// Add configurable options based on request
		if unifiedReq.PreMarkOperation {
			opts = append(opts, option.WithPreMarkOperation(true))
		}

		// Validate required parameters
		if unifiedReq.Text == "" {
			return nil, fmt.Errorf("text parameter is required")
		}

		// Tap by OCR action logic
		log.Info().Str("text", unifiedReq.Text).Msg("tapping by OCR")
		err = driverExt.TapByOCR(unifiedReq.Text, opts...)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Tap by OCR failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully tapped on OCR text: %s", unifiedReq.Text)), nil
	}
}

func (t *ToolTapByOCR) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	if text, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"text": text,
		}

		// Extract options to arguments
		extractActionOptionsToArguments(action.GetOptions(), arguments)

		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid tap by OCR params: %v", action.Params)
}

func (t *ToolTapByOCR) ReturnSchema() map[string]string {
	return map[string]string{
		"message": "string: Success message confirming the operation was completed",
	}
}

// ToolTapByCV implements the tap_cv tool call.
type ToolTapByCV struct{}

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
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// Get options directly since ActionOptions is now ActionOptions
		opts := unifiedReq.Options()

		// Add configurable options based on request
		if unifiedReq.PreMarkOperation {
			opts = append(opts, option.WithPreMarkOperation(true))
		}

		// Tap by CV action logic
		log.Info().Msg("tapping by CV")

		// For TapByCV, we need to check if there are UI types in the options
		// In the original DoAction, it requires ScreenShotWithUITypes to be set
		// We'll add a basic implementation that triggers CV recognition
		err = driverExt.TapByCV(opts...)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Tap by CV failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText("Successfully tapped by computer vision"), nil
	}
}

func (t *ToolTapByCV) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	// For TapByCV, the original action might not have params but relies on options
	arguments := map[string]any{
		"imagePath": "", // Will be handled by the tool based on UI types
	}

	// Extract options to arguments
	extractActionOptionsToArguments(action.GetOptions(), arguments)

	return buildMCPCallToolRequest(t.Name(), arguments), nil
}

func (t *ToolTapByCV) ReturnSchema() map[string]string {
	return defaultReturnSchema()
}

// ToolDoubleTapXY implements the double_tap_xy tool call.
type ToolDoubleTapXY struct{}

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
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
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
			return mcp.NewToolResultError(fmt.Sprintf("Double tap failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully double tapped at (%.2f, %.2f)", unifiedReq.X, unifiedReq.Y)), nil
	}
}

func (t *ToolDoubleTapXY) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	if params, err := builtin.ConvertToFloat64Slice(action.Params); err == nil && len(params) == 2 {
		x, y := params[0], params[1]
		arguments := map[string]any{
			"x": x,
			"y": y,
		}
		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid double tap params: %v", action.Params)
}

func (t *ToolDoubleTapXY) ReturnSchema() map[string]string {
	return defaultReturnSchema()
}
