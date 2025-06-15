package uixt

import (
	"context"
	"fmt"

	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/httprunner/httprunner/v5/uixt/types"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ToolPressButton implements the press_button tool call.
type ToolPressButton struct {
	// Return data fields - these define the structure of data returned by this tool
	Button string `json:"button" desc:"Name of the button that was pressed"`
}

func (t *ToolPressButton) Name() option.ActionName {
	return option.ACTION_PressButton
}

func (t *ToolPressButton) Description() string {
	return "Press a button on the device"
}

func (t *ToolPressButton) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_PressButton)
}

func (t *ToolPressButton) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// Press button action logic
		err = driverExt.PressButton(types.DeviceButton(unifiedReq.Button))
		if err != nil {
			return NewMCPErrorResponse(fmt.Sprintf("Press button failed: %s", err.Error())), nil
		}

		message := fmt.Sprintf("Successfully pressed button: %s", unifiedReq.Button)
		returnData := ToolPressButton{Button: string(unifiedReq.Button)}

		return NewMCPSuccessResponse(message, &returnData), nil
	}
}

func (t *ToolPressButton) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	if button, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"button": button,
		}
		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid press button params: %v", action.Params)
}

// ToolHome implements the home tool call.
type ToolHome struct { // Return data fields - these define the structure of data returned by this tool
}

func (t *ToolHome) Name() option.ActionName {
	return option.ACTION_Home
}

func (t *ToolHome) Description() string {
	return "Press the home button on the device"
}

func (t *ToolHome) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_Home)
}

func (t *ToolHome) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		// Home action logic
		err = driverExt.Home()
		if err != nil {
			return NewMCPErrorResponse(fmt.Sprintf("Home button press failed: %s", err.Error())), nil
		}

		message := "Successfully pressed home button"
		returnData := ToolHome{}

		return NewMCPSuccessResponse(message, &returnData), nil
	}
}

func (t *ToolHome) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	return buildMCPCallToolRequest(t.Name(), map[string]any{}), nil
}

// ToolBack implements the back tool call.
type ToolBack struct { // Return data fields - these define the structure of data returned by this tool
}

func (t *ToolBack) Name() option.ActionName {
	return option.ACTION_Back
}

func (t *ToolBack) Description() string {
	return "Press the back button on the device"
}

func (t *ToolBack) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_Back)
}

func (t *ToolBack) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		// Back action logic
		err = driverExt.Back()
		if err != nil {
			return NewMCPErrorResponse(fmt.Sprintf("Back button press failed: %s", err.Error())), nil
		}

		message := "Successfully pressed back button"
		returnData := ToolBack{}

		return NewMCPSuccessResponse(message, &returnData), nil
	}
}

func (t *ToolBack) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	return buildMCPCallToolRequest(t.Name(), map[string]any{}), nil
}
