package uixt

import (
	"context"
	"fmt"
	"slices"

	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rs/zerolog/log"
)

// ToolSwipe implements the generic swipe tool call.
// It automatically determines whether to use direction-based or coordinate-based swipe
// based on the params type.
type ToolSwipe struct{}

func (t *ToolSwipe) Name() option.ActionName {
	return option.ACTION_Swipe
}

func (t *ToolSwipe) Description() string {
	return "Swipe on the screen by direction (up/down/left/right) or coordinates [fromX, fromY, toX, toY]"
}

func (t *ToolSwipe) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_Swipe)
}

func (t *ToolSwipe) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Check if it's direction-based swipe (has "direction" parameter)
		if _, exists := request.Params.Arguments["direction"]; exists {
			// Delegate to ToolSwipeDirection
			directionTool := &ToolSwipeDirection{}
			return directionTool.Implement()(ctx, request)
		} else {
			// Delegate to ToolSwipeCoordinate
			coordinateTool := &ToolSwipeCoordinate{}
			return coordinateTool.Implement()(ctx, request)
		}
	}
}

func (t *ToolSwipe) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	// Check if params is a string (direction-based swipe)
	if _, ok := action.Params.(string); ok {
		// Delegate to ToolSwipeDirection but use our tool name
		directionTool := &ToolSwipeDirection{}
		request, err := directionTool.ConvertActionToCallToolRequest(action)
		if err != nil {
			return mcp.CallToolRequest{}, err
		}
		// Change the tool name to use generic swipe
		request.Params.Name = string(t.Name())
		return request, nil
	}

	// Check if params is a coordinate array (coordinate-based swipe)
	if paramSlice, err := builtin.ConvertToFloat64Slice(action.Params); err == nil && len(paramSlice) == 4 {
		// Delegate to ToolSwipeCoordinate but use our tool name
		coordinateTool := &ToolSwipeCoordinate{}
		request, err := coordinateTool.ConvertActionToCallToolRequest(action)
		if err != nil {
			return mcp.CallToolRequest{}, err
		}
		// Change the tool name to use generic swipe
		request.Params.Name = string(t.Name())
		return request, nil
	}

	return mcp.CallToolRequest{}, fmt.Errorf("invalid swipe params: %v, expected string direction or [fromX, fromY, toX, toY] coordinates", action.Params)
}

func (t *ToolSwipe) ReturnSchema() map[string]string {
	return map[string]string{
		"message":   "string: Success message confirming the swipe operation",
		"direction": "string: Direction of swipe (for directional swipes)",
		"fromX":     "float64: Starting X coordinate (for coordinate-based swipes)",
		"fromY":     "float64: Starting Y coordinate (for coordinate-based swipes)",
		"toX":       "float64: Ending X coordinate (for coordinate-based swipes)",
		"toY":       "float64: Ending Y coordinate (for coordinate-based swipes)",
	}
}

// ToolSwipeDirection implements the swipe_direction tool call.
type ToolSwipeDirection struct{}

func (t *ToolSwipeDirection) Name() option.ActionName {
	return option.ACTION_SwipeDirection
}

func (t *ToolSwipeDirection) Description() string {
	return "Swipe on the screen in a specific direction (up, down, left, right)"
}

func (t *ToolSwipeDirection) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_SwipeDirection)
}

func (t *ToolSwipeDirection) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}
		swipeDirection := unifiedReq.Direction.(string)

		// Swipe action logic
		log.Info().Str("direction", swipeDirection).Msg("performing swipe")

		// Validate direction
		validDirections := []string{"up", "down", "left", "right"}
		if !slices.Contains(validDirections, swipeDirection) {
			return nil, fmt.Errorf("invalid swipe direction: %s, expected one of: %v",
				swipeDirection, validDirections)
		}

		opts := []option.ActionOption{
			option.WithDuration(getFloat64ValueOrDefault(unifiedReq.Duration, 0.5)),
			option.WithPressDuration(getFloat64ValueOrDefault(unifiedReq.PressDuration, 0.1)),
		}
		if unifiedReq.AntiRisk {
			opts = append(opts, option.WithAntiRisk(true))
		}
		if unifiedReq.PreMarkOperation {
			opts = append(opts, option.WithPreMarkOperation(true))
		}

		// Convert direction to coordinates and perform swipe
		switch swipeDirection {
		case "up":
			err = driverExt.Swipe(0.5, 0.5, 0.5, 0.1, opts...)
		case "down":
			err = driverExt.Swipe(0.5, 0.5, 0.5, 0.9, opts...)
		case "left":
			err = driverExt.Swipe(0.5, 0.5, 0.1, 0.5, opts...)
		case "right":
			err = driverExt.Swipe(0.5, 0.5, 0.9, 0.5, opts...)
		default:
			return mcp.NewToolResultError(
				fmt.Sprintf("Unexpected swipe direction: %s", swipeDirection)), nil
		}

		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Swipe failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully swiped %s", swipeDirection)), nil
	}
}

func (t *ToolSwipeDirection) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	// Handle direction swipe like "up", "down", "left", "right"
	if direction, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"direction": direction,
		}
		// Add duration and press duration from options
		if duration := action.ActionOptions.Duration; duration > 0 {
			arguments["duration"] = duration
		}
		if pressDuration := action.ActionOptions.PressDuration; pressDuration > 0 {
			arguments["pressDuration"] = pressDuration
		}

		// Extract all action options
		extractActionOptionsToArguments(action.GetOptions(), arguments)

		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid swipe params: %v", action.Params)
}

func (t *ToolSwipeDirection) ReturnSchema() map[string]string {
	return map[string]string{
		"message":   "string: Success message confirming the directional swipe",
		"direction": "string: Direction that was swiped (up/down/left/right)",
	}
}

// ToolSwipeCoordinate implements the swipe_coordinate tool call.
type ToolSwipeCoordinate struct{}

func (t *ToolSwipeCoordinate) Name() option.ActionName {
	return option.ACTION_SwipeCoordinate
}

func (t *ToolSwipeCoordinate) Description() string {
	return "Perform swipe with specific start and end coordinates and custom timing"
}

func (t *ToolSwipeCoordinate) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_SwipeCoordinate)
}

func (t *ToolSwipeCoordinate) Implement() server.ToolHandlerFunc {
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
		if unifiedReq.FromX == 0 || unifiedReq.FromY == 0 || unifiedReq.ToX == 0 || unifiedReq.ToY == 0 {
			return nil, fmt.Errorf("fromX, fromY, toX, and toY coordinates are required")
		}

		// Advanced swipe action logic using prepareSwipeAction like the original DoAction
		log.Info().
			Float64("fromX", unifiedReq.FromX).Float64("fromY", unifiedReq.FromY).
			Float64("toX", unifiedReq.ToX).Float64("toY", unifiedReq.ToY).
			Msg("performing advanced swipe")

		params := []float64{unifiedReq.FromX, unifiedReq.FromY, unifiedReq.ToX, unifiedReq.ToY}

		// Build action options from the unified request
		opts := []option.ActionOption{}
		if unifiedReq.Duration > 0 {
			opts = append(opts, option.WithDuration(unifiedReq.Duration))
		}
		if unifiedReq.PressDuration > 0 {
			opts = append(opts, option.WithPressDuration(unifiedReq.PressDuration))
		}
		if unifiedReq.AntiRisk {
			opts = append(opts, option.WithAntiRisk(true))
		}

		swipeAction := prepareSwipeAction(driverExt, params, opts...)
		err = swipeAction(driverExt)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Advanced swipe failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully performed advanced swipe from (%.2f, %.2f) to (%.2f, %.2f)",
			unifiedReq.FromX, unifiedReq.FromY, unifiedReq.ToX, unifiedReq.ToY)), nil
	}
}

func (t *ToolSwipeCoordinate) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	if paramSlice, err := builtin.ConvertToFloat64Slice(action.Params); err == nil && len(paramSlice) == 4 {
		arguments := map[string]any{
			"from_x": paramSlice[0],
			"from_y": paramSlice[1],
			"to_x":   paramSlice[2],
			"to_y":   paramSlice[3],
		}
		// Add duration and press duration from options
		if duration := action.ActionOptions.Duration; duration > 0 {
			arguments["duration"] = duration
		}
		if pressDuration := action.ActionOptions.PressDuration; pressDuration > 0 {
			arguments["pressDuration"] = pressDuration
		}

		// Extract all action options
		extractActionOptionsToArguments(action.GetOptions(), arguments)

		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid swipe advanced params: %v", action.Params)
}

func (t *ToolSwipeCoordinate) ReturnSchema() map[string]string {
	return map[string]string{
		"message": "string: Success message confirming the coordinate-based swipe",
		"fromX":   "float64: Starting X coordinate of the swipe",
		"fromY":   "float64: Starting Y coordinate of the swipe",
		"toX":     "float64: Ending X coordinate of the swipe",
		"toY":     "float64: Ending Y coordinate of the swipe",
	}
}

// ToolSwipeToTapApp implements the swipe_to_tap_app tool call.
type ToolSwipeToTapApp struct{}

func (t *ToolSwipeToTapApp) Name() option.ActionName {
	return option.ACTION_SwipeToTapApp
}

func (t *ToolSwipeToTapApp) Description() string {
	return "Swipe to find and tap an app by name"
}

func (t *ToolSwipeToTapApp) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_SwipeToTapApp)
}

func (t *ToolSwipeToTapApp) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// Build action options from request structure
		var opts []option.ActionOption

		// Add boolean options
		if unifiedReq.IgnoreNotFoundError {
			opts = append(opts, option.WithIgnoreNotFoundError(true))
		}

		// Add numeric options
		if unifiedReq.MaxRetryTimes > 0 {
			opts = append(opts, option.WithMaxRetryTimes(unifiedReq.MaxRetryTimes))
		}
		if unifiedReq.Index > 0 {
			opts = append(opts, option.WithIndex(unifiedReq.Index))
		}

		// Swipe to tap app action logic
		err = driverExt.SwipeToTapApp(unifiedReq.AppName, opts...)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Swipe to tap app failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully found and tapped app: %s", unifiedReq.AppName)), nil
	}
}

func (t *ToolSwipeToTapApp) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	if appName, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"appName": appName,
		}

		// Extract options to arguments
		extractActionOptionsToArguments(action.GetOptions(), arguments)

		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid swipe to tap app params: %v", action.Params)
}

func (t *ToolSwipeToTapApp) ReturnSchema() map[string]string {
	return map[string]string{
		"message": "string: Success message confirming the app was found and tapped",
		"appName": "string: Name of the app that was found and tapped",
	}
}

// ToolSwipeToTapText implements the swipe_to_tap_text tool call.
type ToolSwipeToTapText struct{}

func (t *ToolSwipeToTapText) Name() option.ActionName {
	return option.ACTION_SwipeToTapText
}

func (t *ToolSwipeToTapText) Description() string {
	return "Swipe to find and tap text on screen"
}

func (t *ToolSwipeToTapText) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_SwipeToTapText)
}

func (t *ToolSwipeToTapText) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// Build action options from request structure
		var opts []option.ActionOption

		// Add boolean options
		if unifiedReq.IgnoreNotFoundError {
			opts = append(opts, option.WithIgnoreNotFoundError(true))
		}
		if unifiedReq.Regex {
			opts = append(opts, option.WithRegex(true))
		}

		// Add numeric options
		if unifiedReq.MaxRetryTimes > 0 {
			opts = append(opts, option.WithMaxRetryTimes(unifiedReq.MaxRetryTimes))
		}
		if unifiedReq.Index > 0 {
			opts = append(opts, option.WithIndex(unifiedReq.Index))
		}

		// Swipe to tap text action logic
		err = driverExt.SwipeToTapTexts([]string{unifiedReq.Text}, opts...)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Swipe to tap text failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully found and tapped text: %s", unifiedReq.Text)), nil
	}
}

func (t *ToolSwipeToTapText) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	if text, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"text": text,
		}

		// Extract options to arguments
		extractActionOptionsToArguments(action.GetOptions(), arguments)

		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid swipe to tap text params: %v", action.Params)
}

func (t *ToolSwipeToTapText) ReturnSchema() map[string]string {
	return map[string]string{
		"message": "string: Success message confirming the text was found and tapped",
		"text":    "string: Text content that was found and tapped",
	}
}

// ToolSwipeToTapTexts implements the swipe_to_tap_texts tool call.
type ToolSwipeToTapTexts struct{}

func (t *ToolSwipeToTapTexts) Name() option.ActionName {
	return option.ACTION_SwipeToTapTexts
}

func (t *ToolSwipeToTapTexts) Description() string {
	return "Swipe to find and tap one of multiple texts on screen"
}

func (t *ToolSwipeToTapTexts) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_SwipeToTapTexts)
}

func (t *ToolSwipeToTapTexts) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// Build action options from request structure
		var opts []option.ActionOption

		// Add boolean options
		if unifiedReq.IgnoreNotFoundError {
			opts = append(opts, option.WithIgnoreNotFoundError(true))
		}
		if unifiedReq.Regex {
			opts = append(opts, option.WithRegex(true))
		}

		// Add numeric options
		if unifiedReq.MaxRetryTimes > 0 {
			opts = append(opts, option.WithMaxRetryTimes(unifiedReq.MaxRetryTimes))
		}
		if unifiedReq.Index > 0 {
			opts = append(opts, option.WithIndex(unifiedReq.Index))
		}

		// Swipe to tap texts action logic
		log.Info().Strs("texts", unifiedReq.Texts).Msg("swipe to tap texts")
		err = driverExt.SwipeToTapTexts(unifiedReq.Texts, opts...)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Swipe to tap texts failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully found and tapped one of texts: %v", unifiedReq.Texts)), nil
	}
}

func (t *ToolSwipeToTapTexts) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	var texts []string
	if textsSlice, ok := action.Params.([]string); ok {
		texts = textsSlice
	} else if textsInterface, err := builtin.ConvertToStringSlice(action.Params); err == nil {
		texts = textsInterface
	} else {
		return mcp.CallToolRequest{}, fmt.Errorf("invalid swipe to tap texts params: %v", action.Params)
	}
	arguments := map[string]any{
		"texts": texts,
	}

	// Extract options to arguments
	extractActionOptionsToArguments(action.GetOptions(), arguments)

	return buildMCPCallToolRequest(t.Name(), arguments), nil
}

func (t *ToolSwipeToTapTexts) ReturnSchema() map[string]string {
	return map[string]string{
		"message":   "string: Success message confirming one of the texts was found and tapped",
		"texts":     "[]string: List of text options that were searched for",
		"foundText": "string: The specific text that was actually found and tapped",
	}
}

// ToolDrag implements the drag tool call.
type ToolDrag struct{}

func (t *ToolDrag) Name() option.ActionName {
	return option.ACTION_Drag
}

func (t *ToolDrag) Description() string {
	return "Drag from one point to another on the mobile device screen"
}

func (t *ToolDrag) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_Drag)
}

func (t *ToolDrag) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// Validate required parameters - check if coordinates are provided (not just non-zero)
		_, hasFromX := request.Params.Arguments["from_x"]
		_, hasFromY := request.Params.Arguments["from_y"]
		_, hasToX := request.Params.Arguments["to_x"]
		_, hasToY := request.Params.Arguments["to_y"]
		if !hasFromX || !hasFromY || !hasToX || !hasToY {
			return nil, fmt.Errorf("from_x, from_y, to_x, and to_y coordinates are required")
		}

		opts := []option.ActionOption{}
		if unifiedReq.Duration > 0 {
			opts = append(opts, option.WithDuration(unifiedReq.Duration/1000.0))
		}
		if unifiedReq.AntiRisk {
			opts = append(opts, option.WithAntiRisk(true))
		}

		// Drag action logic
		log.Info().
			Float64("fromX", unifiedReq.FromX).Float64("fromY", unifiedReq.FromY).
			Float64("toX", unifiedReq.ToX).Float64("toY", unifiedReq.ToY).
			Msg("performing drag")

		err = driverExt.Swipe(unifiedReq.FromX, unifiedReq.FromY, unifiedReq.ToX, unifiedReq.ToY, opts...)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Drag failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully dragged from (%.2f, %.2f) to (%.2f, %.2f)",
			unifiedReq.FromX, unifiedReq.FromY, unifiedReq.ToX, unifiedReq.ToY)), nil
	}
}

func (t *ToolDrag) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	if paramSlice, err := builtin.ConvertToFloat64Slice(action.Params); err == nil && len(paramSlice) == 4 {
		arguments := map[string]any{
			"from_x": paramSlice[0],
			"from_y": paramSlice[1],
			"to_x":   paramSlice[2],
			"to_y":   paramSlice[3],
		}
		// Add duration from options
		if duration := action.ActionOptions.Duration; duration > 0 {
			arguments["duration"] = duration * 1000 // convert to milliseconds
		}

		// Extract all action options
		extractActionOptionsToArguments(action.GetOptions(), arguments)

		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid drag parameters: %v", action.Params)
}

func (t *ToolDrag) ReturnSchema() map[string]string {
	return map[string]string{
		"message": "string: Success message confirming the drag operation",
		"fromX":   "float64: Starting X coordinate of the drag",
		"fromY":   "float64: Starting Y coordinate of the drag",
		"toX":     "float64: Ending X coordinate of the drag",
		"toY":     "float64: Ending Y coordinate of the drag",
	}
}
