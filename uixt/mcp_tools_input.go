package uixt

import (
	"context"
	"fmt"

	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ToolInput implements the input tool call.
type ToolInput struct{}

func (t *ToolInput) Name() option.ActionName {
	return option.ACTION_Input
}

func (t *ToolInput) Description() string {
	return "Input text into the currently focused element or input field"
}

func (t *ToolInput) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_Input)
}

func (t *ToolInput) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		if unifiedReq.Text == "" {
			return nil, fmt.Errorf("text is required")
		}

		// Input action logic
		err = driverExt.Input(unifiedReq.Text)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Input failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully input text: %s", unifiedReq.Text)), nil
	}
}

func (t *ToolInput) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	text := fmt.Sprintf("%v", action.Params)
	arguments := map[string]any{
		"text": text,
	}
	return buildMCPCallToolRequest(t.Name(), arguments), nil
}

func (t *ToolInput) ReturnSchema() map[string]string {
	return map[string]string{
		"message": "string: Success message confirming text was input",
		"text":    "string: Text content that was input into the field",
	}
}

// ToolSetIme implements the set_ime tool call.
type ToolSetIme struct{}

func (t *ToolSetIme) Name() option.ActionName {
	return option.ACTION_SetIme
}

func (t *ToolSetIme) Description() string {
	return "Set the input method editor (IME) on the device"
}

func (t *ToolSetIme) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_SetIme)
}

func (t *ToolSetIme) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// Set IME action logic
		err = driverExt.SetIme(unifiedReq.Ime)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Set IME failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully set IME to: %s", unifiedReq.Ime)), nil
	}
}

func (t *ToolSetIme) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	if ime, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"ime": ime,
		}
		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid set ime params: %v", action.Params)
}

func (t *ToolSetIme) ReturnSchema() map[string]string {
	return map[string]string{
		"message": "string: Success message confirming IME was set",
		"ime":     "string: Input method editor that was set",
	}
}
