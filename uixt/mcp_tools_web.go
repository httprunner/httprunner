package uixt

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rs/zerolog/log"
)

// ToolWebLoginNoneUI implements the web_login_none_ui tool call.
type ToolWebLoginNoneUI struct{}

func (t *ToolWebLoginNoneUI) Name() option.ActionName {
	return option.ACTION_WebLoginNoneUI
}

func (t *ToolWebLoginNoneUI) Description() string {
	return "Perform login without UI interaction for web applications"
}

func (t *ToolWebLoginNoneUI) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_WebLoginNoneUI)
}

func (t *ToolWebLoginNoneUI) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// Web login none UI action logic
		log.Info().Str("packageName", unifiedReq.PackageName).Msg("performing web login without UI")
		driver, ok := driverExt.IDriver.(*BrowserDriver)
		if !ok {
			return nil, fmt.Errorf("invalid browser driver for web login")
		}

		_, err = driver.LoginNoneUI(unifiedReq.PackageName, unifiedReq.PhoneNumber, unifiedReq.Captcha, unifiedReq.Password)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Web login failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText("Successfully performed web login without UI"), nil
	}
}

func (t *ToolWebLoginNoneUI) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	return buildMCPCallToolRequest(t.Name(), map[string]any{}), nil
}

func (t *ToolWebLoginNoneUI) ReturnSchema() map[string]string {
	return map[string]string{
		"message":     "string: Success message confirming web login was completed",
		"loginResult": "object: Result of the login operation (success/failure details)",
	}
}

// ToolSecondaryClick implements the secondary_click tool call.
type ToolSecondaryClick struct{}

func (t *ToolSecondaryClick) Name() option.ActionName {
	return option.ACTION_SecondaryClick
}

func (t *ToolSecondaryClick) Description() string {
	return "Perform secondary click (right click) at specified coordinates"
}

func (t *ToolSecondaryClick) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_SecondaryClick)
}

func (t *ToolSecondaryClick) Implement() server.ToolHandlerFunc {
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

		// Secondary click action logic
		err = driverExt.SecondaryClick(unifiedReq.X, unifiedReq.Y)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Secondary click failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully performed secondary click at (%.2f, %.2f)", unifiedReq.X, unifiedReq.Y)), nil
	}
}

func (t *ToolSecondaryClick) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	if params, err := builtin.ConvertToFloat64Slice(action.Params); err == nil && len(params) == 2 {
		arguments := map[string]any{
			"x": params[0],
			"y": params[1],
		}
		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid secondary click params: %v", action.Params)
}

func (t *ToolSecondaryClick) ReturnSchema() map[string]string {
	return map[string]string{
		"message": "string: Success message confirming secondary click (right-click) operation",
		"x":       "float64: X coordinate where secondary click was performed",
		"y":       "float64: Y coordinate where secondary click was performed",
	}
}

// ToolHoverBySelector implements the hover_by_selector tool call.
type ToolHoverBySelector struct{}

func (t *ToolHoverBySelector) Name() option.ActionName {
	return option.ACTION_HoverBySelector
}

func (t *ToolHoverBySelector) Description() string {
	return "Hover over an element selected by CSS selector or XPath"
}

func (t *ToolHoverBySelector) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_HoverBySelector)
}

func (t *ToolHoverBySelector) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// Hover by selector action logic
		err = driverExt.HoverBySelector(unifiedReq.Selector)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Hover by selector failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully hovered over element with selector: %s", unifiedReq.Selector)), nil
	}
}

func (t *ToolHoverBySelector) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	if selector, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"selector": selector,
		}
		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid hover by selector params: %v", action.Params)
}

func (t *ToolHoverBySelector) ReturnSchema() map[string]string {
	return map[string]string{
		"message":  "string: Success message confirming hover operation",
		"selector": "string: CSS selector or XPath of the element that was hovered over",
	}
}

// ToolTapBySelector implements the tap_by_selector tool call.
type ToolTapBySelector struct{}

func (t *ToolTapBySelector) Name() option.ActionName {
	return option.ACTION_TapBySelector
}

func (t *ToolTapBySelector) Description() string {
	return "Tap an element selected by CSS selector or XPath"
}

func (t *ToolTapBySelector) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_TapBySelector)
}

func (t *ToolTapBySelector) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// Tap by selector action logic
		err = driverExt.TapBySelector(unifiedReq.Selector)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Tap by selector failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully tapped element with selector: %s", unifiedReq.Selector)), nil
	}
}

func (t *ToolTapBySelector) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	if selector, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"selector": selector,
		}
		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid tap by selector params: %v", action.Params)
}

func (t *ToolTapBySelector) ReturnSchema() map[string]string {
	return map[string]string{
		"message":  "string: Success message confirming tap operation",
		"selector": "string: CSS selector or XPath of the element that was tapped",
	}
}

// ToolSecondaryClickBySelector implements the secondary_click_by_selector tool call.
type ToolSecondaryClickBySelector struct{}

func (t *ToolSecondaryClickBySelector) Name() option.ActionName {
	return option.ACTION_SecondaryClickBySelector
}

func (t *ToolSecondaryClickBySelector) Description() string {
	return "Perform secondary click on an element selected by CSS selector or XPath"
}

func (t *ToolSecondaryClickBySelector) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_SecondaryClickBySelector)
}

func (t *ToolSecondaryClickBySelector) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// Secondary click by selector action logic
		err = driverExt.SecondaryClickBySelector(unifiedReq.Selector)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Secondary click by selector failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully performed secondary click on element with selector: %s", unifiedReq.Selector)), nil
	}
}

func (t *ToolSecondaryClickBySelector) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	if selector, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"selector": selector,
		}
		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid secondary click by selector params: %v", action.Params)
}

func (t *ToolSecondaryClickBySelector) ReturnSchema() map[string]string {
	return map[string]string{
		"message":  "string: Success message confirming secondary click operation",
		"selector": "string: CSS selector or XPath of the element that was right-clicked",
	}
}

// ToolWebCloseTab implements the web_close_tab tool call.
type ToolWebCloseTab struct{}

func (t *ToolWebCloseTab) Name() option.ActionName {
	return option.ACTION_WebCloseTab
}

func (t *ToolWebCloseTab) Description() string {
	return "Close a browser tab by index"
}

func (t *ToolWebCloseTab) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_WebCloseTab)
}

func (t *ToolWebCloseTab) Implement() server.ToolHandlerFunc {
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
		if unifiedReq.TabIndex == 0 {
			return nil, fmt.Errorf("tabIndex is required")
		}

		// Web close tab action logic
		browserDriver, ok := driverExt.IDriver.(*BrowserDriver)
		if !ok {
			return nil, fmt.Errorf("web close tab is only supported for browser drivers")
		}

		err = browserDriver.CloseTab(unifiedReq.TabIndex)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Close tab failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully closed tab at index: %d", unifiedReq.TabIndex)), nil
	}
}

func (t *ToolWebCloseTab) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	var tabIndex int
	if param, ok := action.Params.(json.Number); ok {
		paramInt64, _ := param.Int64()
		tabIndex = int(paramInt64)
	} else if param, ok := action.Params.(int64); ok {
		tabIndex = int(param)
	} else if param, ok := action.Params.(int); ok {
		tabIndex = param
	} else {
		return mcp.CallToolRequest{}, fmt.Errorf("invalid web close tab params: %v", action.Params)
	}
	arguments := map[string]any{
		"tabIndex": tabIndex,
	}
	return buildMCPCallToolRequest(t.Name(), arguments), nil
}

func (t *ToolWebCloseTab) ReturnSchema() map[string]string {
	return map[string]string{
		"message":  "string: Success message confirming browser tab was closed",
		"tabIndex": "int: Index of the tab that was closed",
	}
}
