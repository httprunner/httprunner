package uixt

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/httprunner/httprunner/v5/uixt/option"
)

// ToolGetPasteboard implements the get_pasteboard tool call.
type ToolGetPasteboard struct {
	// Return data fields - these define the structure of data returned by this tool
	Content string `json:"content" desc:"Clipboard content that was retrieved"`
}

func (t *ToolGetPasteboard) Name() option.ActionName {
	return option.ACTION_GetPasteboard
}

func (t *ToolGetPasteboard) Description() string {
	return "Get the clipboard content from the device"
}

func (t *ToolGetPasteboard) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_GetPasteboard)
}

func (t *ToolGetPasteboard) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments := request.GetArguments()
		driverExt, err := setupXTDriver(ctx, arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		// Directly call the GetPasteboard method on the driver
		content, err := driverExt.IDriver.GetPasteboard()
		if err != nil {
			return NewMCPErrorResponse(fmt.Sprintf("Get pasteboard failed: %s", err.Error())), err
		}

		message := "Successfully retrieved clipboard content"
		returnData := ToolGetPasteboard{Content: content}

		return NewMCPSuccessResponse(message, &returnData), nil
	}
}

func (t *ToolGetPasteboard) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	arguments := map[string]any{}
	return BuildMCPCallToolRequest(t.Name(), arguments, action), nil
}
