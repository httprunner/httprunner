package uixt

import (
	"context"
	"fmt"
	"slices"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/uixt/option"
)

// ToolSwipe implements the generic swipe tool call.
// It automatically determines whether to use direction-based or coordinate-based swipe
// based on the params type.
type ToolSwipe struct {
	// Return data fields - these define the structure of data returned by this tool
	SwipeType string `json:"swipeType" desc:"Type of swipe performed (direction or coordinate)"`
}

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
		if _, exists := request.GetArguments()["direction"]; exists {
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

// ToolSwipeDirection implements the swipe_direction tool call.
type ToolSwipeDirection struct {
	// Return data fields - these define the structure of data returned by this tool
	Direction string  `json:"direction" desc:"Direction that was swiped (up/down/left/right)"`
	FromX     float64 `json:"fromX" desc:"Starting X coordinate of the swipe"`
	FromY     float64 `json:"fromY" desc:"Starting Y coordinate of the swipe"`
	ToX       float64 `json:"toX" desc:"Ending X coordinate of the swipe"`
	ToY       float64 `json:"toY" desc:"Ending Y coordinate of the swipe"`
}

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
		arguments := request.GetArguments()
		driverExt, err := setupXTDriver(ctx, arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(arguments)
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

		// Build all options from request arguments
		opts := unifiedReq.Options()
		if unifiedReq.Duration == 0 {
			opts = append(opts, option.WithDuration(0.5))
		}
		if unifiedReq.PressDuration == 0 {
			opts = append(opts, option.WithPressDuration(0.1))
		}

		// Convert direction to coordinates and perform swipe
		var fromX, fromY, toX, toY float64
		switch swipeDirection {
		case "up":
			fromX, fromY, toX, toY = 0.5, 0.5, 0.5, 0.1
			err = driverExt.Swipe(fromX, fromY, toX, toY, opts...)
		case "down":
			fromX, fromY, toX, toY = 0.5, 0.5, 0.5, 0.9
			err = driverExt.Swipe(fromX, fromY, toX, toY, opts...)
		case "left":
			fromX, fromY, toX, toY = 0.5, 0.5, 0.1, 0.5
			err = driverExt.Swipe(fromX, fromY, toX, toY, opts...)
		case "right":
			fromX, fromY, toX, toY = 0.5, 0.5, 0.9, 0.5
			err = driverExt.Swipe(fromX, fromY, toX, toY, opts...)
		default:
			return NewMCPErrorResponse(fmt.Sprintf("Unexpected swipe direction: %s", swipeDirection)), nil
		}

		if err != nil {
			return NewMCPErrorResponse(fmt.Sprintf("Swipe failed: %s", err.Error())), err
		}

		message := fmt.Sprintf("Successfully swiped %s", swipeDirection)
		returnData := ToolSwipeDirection{
			Direction: swipeDirection,
			FromX:     fromX,
			FromY:     fromY,
			ToX:       toX,
			ToY:       toY,
		}

		return NewMCPSuccessResponse(message, &returnData), nil
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
		return BuildMCPCallToolRequest(t.Name(), arguments, action), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid swipe params: %v", action.Params)
}

// ToolSwipeCoordinate implements the swipe_coordinate tool call.
type ToolSwipeCoordinate struct {
	// Return data fields - these define the structure of data returned by this tool
	FromX float64 `json:"fromX" desc:"Starting X coordinate of the swipe"`
	FromY float64 `json:"fromY" desc:"Starting Y coordinate of the swipe"`
	ToX   float64 `json:"toX" desc:"Ending X coordinate of the swipe"`
	ToY   float64 `json:"toY" desc:"Ending Y coordinate of the swipe"`
}

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
		if unifiedReq.FromX == 0 || unifiedReq.FromY == 0 || unifiedReq.ToX == 0 || unifiedReq.ToY == 0 {
			return nil, fmt.Errorf("fromX, fromY, toX, and toY coordinates are required")
		}

		// Advanced swipe action logic using prepareSwipeAction like the original DoAction
		log.Info().
			Float64("fromX", unifiedReq.FromX).Float64("fromY", unifiedReq.FromY).
			Float64("toX", unifiedReq.ToX).Float64("toY", unifiedReq.ToY).
			Msg("performing advanced swipe")

		params := []float64{unifiedReq.FromX, unifiedReq.FromY, unifiedReq.ToX, unifiedReq.ToY}

		// Build all options from request arguments
		opts := unifiedReq.Options()

		swipeAction := prepareSwipeAction(driverExt, params, opts...)
		err = swipeAction(driverExt)
		if err != nil {
			return NewMCPErrorResponse(fmt.Sprintf("Advanced swipe failed: %s", err.Error())), err
		}

		message := fmt.Sprintf("Successfully performed advanced swipe from (%.2f, %.2f) to (%.2f, %.2f)",
			unifiedReq.FromX, unifiedReq.FromY, unifiedReq.ToX, unifiedReq.ToY)
		returnData := ToolSwipeCoordinate{
			FromX: unifiedReq.FromX,
			FromY: unifiedReq.FromY,
			ToX:   unifiedReq.ToX,
			ToY:   unifiedReq.ToY,
		}

		return NewMCPSuccessResponse(message, &returnData), nil
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
		return BuildMCPCallToolRequest(t.Name(), arguments, action), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid swipe advanced params: %v", action.Params)
}

// ToolSwipeToTapApp implements the swipe_to_tap_app tool call.
type ToolSwipeToTapApp struct {
	// Return data fields - these define the structure of data returned by this tool
	AppName string `json:"appName" desc:"Name of the app that was found and tapped"`
}

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
		arguments := request.GetArguments()
		driverExt, err := setupXTDriver(ctx, arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(arguments)
		if err != nil {
			return nil, err
		}

		// Build action options from request structure
		opts := unifiedReq.Options()

		// Swipe to tap app action logic
		err = driverExt.SwipeToTapApp(unifiedReq.AppName, opts...)
		if err != nil {
			return NewMCPErrorResponse(fmt.Sprintf("Swipe to tap app failed: %s", err.Error())), err
		}

		message := fmt.Sprintf("Successfully found and tapped app: %s", unifiedReq.AppName)
		returnData := ToolSwipeToTapApp{AppName: unifiedReq.AppName}

		return NewMCPSuccessResponse(message, &returnData), nil
	}
}

func (t *ToolSwipeToTapApp) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	if appName, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"appName": appName,
		}
		return BuildMCPCallToolRequest(t.Name(), arguments, action), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid swipe to tap app params: %v", action.Params)
}

// ToolSwipeToTapText implements the swipe_to_tap_text tool call.
type ToolSwipeToTapText struct {
	// Return data fields - these define the structure of data returned by this tool
	Text string `json:"text" desc:"Text that was found and tapped"`
}

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

		// Swipe to tap text action logic
		err = driverExt.SwipeToTapTexts([]string{unifiedReq.Text}, opts...)
		if err != nil {
			return NewMCPErrorResponse(fmt.Sprintf("Swipe to tap text failed: %s", err.Error())), err
		}

		message := fmt.Sprintf("Successfully found and tapped text: %s", unifiedReq.Text)
		returnData := ToolSwipeToTapText{Text: unifiedReq.Text}

		return NewMCPSuccessResponse(message, &returnData), nil
	}
}

func (t *ToolSwipeToTapText) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	if text, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"text": text,
		}
		return BuildMCPCallToolRequest(t.Name(), arguments, action), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid swipe to tap text params: %v", action.Params)
}

// ToolSwipeToTapTexts implements the swipe_to_tap_texts tool call.
type ToolSwipeToTapTexts struct {
	// Return data fields - these define the structure of data returned by this tool
	Texts      []string `json:"texts" desc:"List of texts that were searched for"`
	TappedText string   `json:"tappedText" desc:"The specific text that was found and tapped"`
}

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

		// Swipe to tap texts action logic
		err = driverExt.SwipeToTapTexts(unifiedReq.Texts, opts...)
		if err != nil {
			return NewMCPErrorResponse(fmt.Sprintf("Swipe to tap texts failed: %s", err.Error())), err
		}

		message := fmt.Sprintf("Successfully found and tapped one of texts: %v", unifiedReq.Texts)
		returnData := ToolSwipeToTapTexts{
			Texts:      unifiedReq.Texts,
			TappedText: "unknown", // We don't know which specific text was tapped
		}

		return NewMCPSuccessResponse(message, &returnData), nil
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
	return BuildMCPCallToolRequest(t.Name(), arguments, action), nil
}

// ToolDrag implements the drag tool call.
type ToolDrag struct {
	// Return data fields - these define the structure of data returned by this tool
	FromX float64 `json:"fromX" desc:"Starting X coordinate of the drag"`
	FromY float64 `json:"fromY" desc:"Starting Y coordinate of the drag"`
	ToX   float64 `json:"toX" desc:"Ending X coordinate of the drag"`
	ToY   float64 `json:"toY" desc:"Ending Y coordinate of the drag"`
}

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
		arguments := request.GetArguments()
		driverExt, err := setupXTDriver(ctx, arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(arguments)
		if err != nil {
			return nil, err
		}

		// Validate required parameters - check if coordinates are provided (not just non-zero)
		_, hasFromX := arguments["from_x"]
		_, hasFromY := arguments["from_y"]
		_, hasToX := arguments["to_x"]
		_, hasToY := arguments["to_y"]
		if !hasFromX || !hasFromY || !hasToX || !hasToY {
			return nil, fmt.Errorf("from_x, from_y, to_x, and to_y coordinates are required")
		}

		// Build all options from request arguments
		opts := unifiedReq.Options()
		if unifiedReq.Duration == 0 {
			opts = append(opts, option.WithDuration(0.5))
		}

		// Drag action logic
		err = driverExt.Swipe(unifiedReq.FromX, unifiedReq.FromY, unifiedReq.ToX, unifiedReq.ToY, opts...)
		if err != nil {
			return NewMCPErrorResponse(fmt.Sprintf("Drag failed: %s", err.Error())), err
		}

		message := fmt.Sprintf("Successfully dragged from (%.2f, %.2f) to (%.2f, %.2f)",
			unifiedReq.FromX, unifiedReq.FromY, unifiedReq.ToX, unifiedReq.ToY)
		returnData := ToolDrag{
			FromX: unifiedReq.FromX,
			FromY: unifiedReq.FromY,
			ToX:   unifiedReq.ToX,
			ToY:   unifiedReq.ToY,
		}

		return NewMCPSuccessResponse(message, &returnData), nil
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
		return BuildMCPCallToolRequest(t.Name(), arguments, action), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid drag parameters: %v", action.Params)
}

// ToolSIMSwipeDirection implements the sim_swipe_direction tool call.
type ToolSIMSwipeDirection struct {
	// Return data fields - these define the structure of data returned by this tool
	Direction      string  `json:"direction" desc:"Direction that was swiped (up/down/left/right)"`
	StartX         float64 `json:"startX" desc:"Starting X coordinate of the simulated swipe"`
	StartY         float64 `json:"startY" desc:"Starting Y coordinate of the simulated swipe"`
	MinDistance    float64 `json:"minDistance" desc:"Minimum distance of the simulated swipe"`
	MaxDistance    float64 `json:"maxDistance" desc:"Maximum distance of the simulated swipe"`
	ActualDistance float64 `json:"actualDistance" desc:"Actual distance of the simulated swipe"`
}

func (t *ToolSIMSwipeDirection) Name() option.ActionName {
	return option.ACTION_SIMSwipeDirection
}

func (t *ToolSIMSwipeDirection) Description() string {
	return "Perform simulated swipe in specified direction with random distance and human-like touch patterns"
}

func (t *ToolSIMSwipeDirection) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_SIMSwipeDirection)
}

func (t *ToolSIMSwipeDirection) Implement() server.ToolHandlerFunc {
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
		if unifiedReq.Direction == nil {
			return nil, fmt.Errorf("direction parameter is required")
		}
		direction, ok := unifiedReq.Direction.(string)
		if !ok {
			return nil, fmt.Errorf("direction must be a string")
		}

		// Validate direction
		validDirections := []string{"up", "down", "left", "right"}
		if !slices.Contains(validDirections, direction) {
			return nil, fmt.Errorf("invalid swipe direction: %s, expected one of: %v",
				direction, validDirections)
		}

		// Default values if not provided - use fromX/fromY instead of startX/startY
		fromX := unifiedReq.FromX
		fromY := unifiedReq.FromY
		simMinDistance := unifiedReq.SIMMinDistance
		simMaxDistance := unifiedReq.SIMMaxDistance

		if fromX == 0 {
			fromX = 0.5 // default to center
		}
		if fromY == 0 {
			fromY = 0.5 // default to center
		}
		if simMinDistance == 0 {
			simMinDistance = 100 // default minimum distance
		}
		if simMaxDistance == 0 {
			simMaxDistance = 300 // default maximum distance
		}

		log.Info().
			Str("direction", direction).
			Float64("startX", fromX).
			Float64("startY", fromY).
			Float64("minDistance", simMinDistance).
			Float64("maxDistance", simMaxDistance).
			Msg("performing simulated swipe with direction")

		// Build all options from request arguments
		opts := unifiedReq.Options()

		// Call the underlying SIMSwipeWithDirection method (check if driver supports SIM)
		if simDriver, ok := driverExt.IDriver.(SIMSupport); ok {
			err = simDriver.SIMSwipeWithDirection(direction, fromX, fromY, simMinDistance, simMaxDistance, opts...)
			if err != nil {
				return NewMCPErrorResponse(fmt.Sprintf("Simulated swipe failed: %s", err.Error())), err
			}
		} else {
			return NewMCPErrorResponse("SIMSwipeWithDirection is not supported by the current driver"), fmt.Errorf("driver does not implement SIMSupport interface")
		}

		// Calculate actual distance for response (approximate)
		actualDistance := simMinDistance
		if simMaxDistance > simMinDistance {
			actualDistance = simMinDistance + (simMaxDistance-simMinDistance)*0.5 // approximate middle value
		}

		message := fmt.Sprintf("Successfully performed simulated swipe %s from (%.2f, %.2f) with distance %.2f",
			direction, fromX, fromY, actualDistance)
		returnData := ToolSIMSwipeDirection{
			Direction:      direction,
			StartX:         fromX,
			StartY:         fromY,
			MinDistance:    simMinDistance,
			MaxDistance:    simMaxDistance,
			ActualDistance: actualDistance,
		}

		return NewMCPSuccessResponse(message, &returnData), nil
	}
}

func (t *ToolSIMSwipeDirection) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	// Handle params as map[string]interface{}
	if paramsMap, ok := action.Params.(map[string]interface{}); ok {
		arguments := map[string]any{}

		// Extract direction
		if direction, exists := paramsMap["direction"]; exists {
			arguments["direction"] = direction
		}

		// Extract coordinates and distances - use new field names directly
		if fromX, exists := paramsMap["from_x"]; exists {
			arguments["from_x"] = fromX
		}
		if fromY, exists := paramsMap["from_y"]; exists {
			arguments["from_y"] = fromY
		}
		if minDistance, exists := paramsMap["sim_min_distance"]; exists {
			arguments["sim_min_distance"] = minDistance
		}
		if maxDistance, exists := paramsMap["sim_max_distance"]; exists {
			arguments["sim_max_distance"] = maxDistance
		}

		// Add duration and press duration from options
		if duration := action.ActionOptions.Duration; duration > 0 {
			arguments["duration"] = duration
		}
		if pressDuration := action.ActionOptions.PressDuration; pressDuration > 0 {
			arguments["pressDuration"] = pressDuration
		}

		return BuildMCPCallToolRequest(t.Name(), arguments, action), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid SIM swipe direction params: %v", action.Params)
}

// ToolSIMSwipeInArea implements the sim_swipe_in_area tool call.
type ToolSIMSwipeInArea struct {
	// Return data fields - these define the structure of data returned by this tool
	Direction   string  `json:"direction" desc:"Direction that was swiped (up/down/left/right)"`
	AreaStartX  float64 `json:"areaStartX" desc:"Area starting X coordinate"`
	AreaStartY  float64 `json:"areaStartY" desc:"Area starting Y coordinate"`
	AreaEndX    float64 `json:"areaEndX" desc:"Area ending X coordinate"`
	AreaEndY    float64 `json:"areaEndY" desc:"Area ending Y coordinate"`
	MinDistance float64 `json:"minDistance" desc:"Minimum distance of the simulated swipe"`
	MaxDistance float64 `json:"maxDistance" desc:"Maximum distance of the simulated swipe"`
}

func (t *ToolSIMSwipeInArea) Name() option.ActionName {
	return option.ACTION_SIMSwipeInArea
}

func (t *ToolSIMSwipeInArea) Description() string {
	return "Perform simulated swipe in specified area with direction and random distance"
}

func (t *ToolSIMSwipeInArea) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_SIMSwipeInArea)
}

func (t *ToolSIMSwipeInArea) Implement() server.ToolHandlerFunc {
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
		if unifiedReq.Direction == nil {
			return nil, fmt.Errorf("direction parameter is required")
		}
		direction, ok := unifiedReq.Direction.(string)
		if !ok {
			return nil, fmt.Errorf("direction must be a string")
		}

		// Validate direction
		validDirections := []string{"up", "down", "left", "right"}
		if !slices.Contains(validDirections, direction) {
			return nil, fmt.Errorf("invalid swipe direction: %s, expected one of: %v",
				direction, validDirections)
		}

		// Get area coordinates - use SIM-prefixed fields
		simAreaStartX := unifiedReq.SIMAreaStartX
		simAreaStartY := unifiedReq.SIMAreaStartY
		simAreaEndX := unifiedReq.SIMAreaEndX
		simAreaEndY := unifiedReq.SIMAreaEndY
		simMinDistance := unifiedReq.SIMMinDistance
		simMaxDistance := unifiedReq.SIMMaxDistance

		// Default values
		if simMinDistance == 0 {
			simMinDistance = 100
		}
		if simMaxDistance == 0 {
			simMaxDistance = 300
		}

		log.Info().
			Str("direction", direction).
			Float64("areaStartX", simAreaStartX).
			Float64("areaStartY", simAreaStartY).
			Float64("areaEndX", simAreaEndX).
			Float64("areaEndY", simAreaEndY).
			Float64("minDistance", simMinDistance).
			Float64("maxDistance", simMaxDistance).
			Msg("performing simulated swipe in area")

		// Build all options from request arguments
		opts := unifiedReq.Options()

		// Call the underlying SIMSwipeInArea method (check if driver supports SIM)
		if simDriver, ok := driverExt.IDriver.(SIMSupport); ok {
			err = simDriver.SIMSwipeInArea(direction, simAreaStartX, simAreaStartY, simAreaEndX, simAreaEndY, simMinDistance, simMaxDistance, opts...)
			if err != nil {
				return NewMCPErrorResponse(fmt.Sprintf("Simulated swipe in area failed: %s", err.Error())), err
			}
		} else {
			return NewMCPErrorResponse("SIMSwipeInArea is not supported by the current driver"), fmt.Errorf("driver does not implement SIMSupport interface")
		}

		message := fmt.Sprintf("Successfully performed simulated swipe %s in area (%.2f,%.2f)-(%.2f,%.2f)",
			direction, simAreaStartX, simAreaStartY, simAreaEndX, simAreaEndY)
		returnData := ToolSIMSwipeInArea{
			Direction:   direction,
			AreaStartX:  simAreaStartX,
			AreaStartY:  simAreaStartY,
			AreaEndX:    simAreaEndX,
			AreaEndY:    simAreaEndY,
			MinDistance: simMinDistance,
			MaxDistance: simMaxDistance,
		}

		return NewMCPSuccessResponse(message, &returnData), nil
	}
}

func (t *ToolSIMSwipeInArea) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	// Handle params as map[string]interface{}
	if paramsMap, ok := action.Params.(map[string]interface{}); ok {
		arguments := map[string]any{}

		// Extract direction
		if direction, exists := paramsMap["direction"]; exists {
			arguments["direction"] = direction
		}

		// Extract area coordinates and distances - use SIM-prefixed field names
		if areaStartX, exists := paramsMap["sim_area_start_x"]; exists {
			arguments["sim_area_start_x"] = areaStartX
		}
		if areaStartY, exists := paramsMap["sim_area_start_y"]; exists {
			arguments["sim_area_start_y"] = areaStartY
		}
		if areaEndX, exists := paramsMap["sim_area_end_x"]; exists {
			arguments["sim_area_end_x"] = areaEndX
		}
		if areaEndY, exists := paramsMap["sim_area_end_y"]; exists {
			arguments["sim_area_end_y"] = areaEndY
		}
		if minDistance, exists := paramsMap["sim_min_distance"]; exists {
			arguments["sim_min_distance"] = minDistance
		}
		if maxDistance, exists := paramsMap["sim_max_distance"]; exists {
			arguments["sim_max_distance"] = maxDistance
		}

		// Add duration and press duration from options
		if duration := action.ActionOptions.Duration; duration > 0 {
			arguments["duration"] = duration
		}
		if pressDuration := action.ActionOptions.PressDuration; pressDuration > 0 {
			arguments["pressDuration"] = pressDuration
		}

		return BuildMCPCallToolRequest(t.Name(), arguments, action), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid SIM swipe in area params: %v", action.Params)
}

// ToolSIMSwipeFromPointToPoint implements the sim_swipe_point_to_point tool call.
type ToolSIMSwipeFromPointToPoint struct {
	// Return data fields - these define the structure of data returned by this tool
	StartX float64 `json:"startX" desc:"Starting X coordinate"`
	StartY float64 `json:"startY" desc:"Starting Y coordinate"`
	EndX   float64 `json:"endX" desc:"Ending X coordinate"`
	EndY   float64 `json:"endY" desc:"Ending Y coordinate"`
}

func (t *ToolSIMSwipeFromPointToPoint) Name() option.ActionName {
	return option.ACTION_SIMSwipeFromPointToPoint
}

func (t *ToolSIMSwipeFromPointToPoint) Description() string {
	return "Perform simulated swipe from point to point with human-like touch patterns"
}

func (t *ToolSIMSwipeFromPointToPoint) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_SIMSwipeFromPointToPoint)
}

func (t *ToolSIMSwipeFromPointToPoint) Implement() server.ToolHandlerFunc {
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

		// Get coordinates from arguments - use fromX/fromY instead of startX/startY
		fromX := unifiedReq.FromX
		fromY := unifiedReq.FromY
		toX := unifiedReq.ToX
		toY := unifiedReq.ToY

		log.Info().
			Float64("startX", fromX).
			Float64("startY", fromY).
			Float64("endX", toX).
			Float64("endY", toY).
			Msg("performing simulated point to point swipe")

		// Build all options from request arguments
		opts := unifiedReq.Options()

		// Call the underlying SIMSwipeFromPointToPoint method (check if driver supports SIM)
		if simDriver, ok := driverExt.IDriver.(SIMSupport); ok {
			err = simDriver.SIMSwipeFromPointToPoint(fromX, fromY, toX, toY, opts...)
			if err != nil {
				return NewMCPErrorResponse(fmt.Sprintf("Simulated point to point swipe failed: %s", err.Error())), err
			}
		} else {
			return NewMCPErrorResponse("SIMSwipeFromPointToPoint is not supported by the current driver"), fmt.Errorf("driver does not implement SIMSupport interface")
		}

		message := fmt.Sprintf("Successfully performed simulated swipe from (%.2f,%.2f) to (%.2f,%.2f)",
			fromX, fromY, toX, toY)
		returnData := ToolSIMSwipeFromPointToPoint{
			StartX: fromX,
			StartY: fromY,
			EndX:   toX,
			EndY:   toY,
		}

		return NewMCPSuccessResponse(message, &returnData), nil
	}
}

func (t *ToolSIMSwipeFromPointToPoint) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	// Handle params as map[string]interface{}
	if paramsMap, ok := action.Params.(map[string]interface{}); ok {
		arguments := map[string]any{}

		// Extract coordinates - use new field names directly
		if fromX, exists := paramsMap["from_x"]; exists {
			arguments["from_x"] = fromX
		}
		if fromY, exists := paramsMap["from_y"]; exists {
			arguments["from_y"] = fromY
		}
		if toX, exists := paramsMap["to_x"]; exists {
			arguments["to_x"] = toX
		}
		if toY, exists := paramsMap["to_y"]; exists {
			arguments["to_y"] = toY
		}

		// Add duration and press duration from options
		if duration := action.ActionOptions.Duration; duration > 0 {
			arguments["duration"] = duration
		}
		if pressDuration := action.ActionOptions.PressDuration; pressDuration > 0 {
			arguments["pressDuration"] = pressDuration
		}

		return BuildMCPCallToolRequest(t.Name(), arguments, action), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid SIM swipe point to point params: %v", action.Params)
}
