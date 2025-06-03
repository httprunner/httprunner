package uixt

import (
	"context"
	"fmt"

	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/httprunner/httprunner/v5/uixt/types"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rs/zerolog/log"
)

// ToolPressButton implements the press_button tool call.
type ToolPressButton struct{}

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
		log.Info().Str("button", string(unifiedReq.Button)).Msg("pressing button")
		err = driverExt.PressButton(types.DeviceButton(unifiedReq.Button))
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Press button failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully pressed button: %s", unifiedReq.Button)), nil
	}
}

func (t *ToolPressButton) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	if button, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"button": button,
		}
		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid press button params: %v", action.Params)
}

func (t *ToolPressButton) ReturnSchema() map[string]string {
	return map[string]string{
		"message": "string: Success message confirming the button press operation",
		"button":  "string: Name of the button that was pressed",
	}
}

// ToolHome implements the home tool call.
type ToolHome struct{}

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
		log.Info().Msg("pressing home button")
		err = driverExt.Home()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Home button press failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText("Successfully pressed home button"), nil
	}
}

func (t *ToolHome) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	return buildMCPCallToolRequest(t.Name(), map[string]any{}), nil
}

func (t *ToolHome) ReturnSchema() map[string]string {
	return map[string]string{
		"message": "string: Success message confirming home button was pressed",
	}
}

// ToolBack implements the back tool call.
type ToolBack struct{}

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
		log.Info().Msg("pressing back button")
		err = driverExt.Back()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Back button press failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText("Successfully pressed back button"), nil
	}
}

func (t *ToolBack) ConvertActionToCallToolRequest(action MobileAction) (mcp.CallToolRequest, error) {
	return buildMCPCallToolRequest(t.Name(), map[string]any{}), nil
}

func (t *ToolBack) ReturnSchema() map[string]string {
	return map[string]string{
		"message": "string: Success message confirming back button was pressed",
	}
}
