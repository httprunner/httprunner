package uixt

import (
	"context"
	"fmt"

	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rs/zerolog/log"
)

// ToolScreenShot implements the screenshot tool call.
type ToolScreenShot struct{}

func (t *ToolScreenShot) Name() option.ActionName {
	return option.ACTION_ScreenShot
}

func (t *ToolScreenShot) Description() string {
	return "Take a screenshot of the mobile device screen. Use this to understand what's currently displayed on screen."
}

func (t *ToolScreenShot) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_ScreenShot)
}

func (t *ToolScreenShot) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, err
		}
		bufferBase64, err := GetScreenShotBufferBase64(driverExt.IDriver)
		if err != nil {
			log.Error().Err(err).Msg("ScreenShot failed")
			return mcp.NewToolResultError(fmt.Sprintf("Failed to take screenshot: %v", err)), nil
		}
		log.Debug().Int("imageBytes", len(bufferBase64)).Msg("take screenshot success")

		return mcp.NewToolResultImage("screenshot", bufferBase64, "image/jpeg"), nil
	}
}

func (t *ToolScreenShot) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	return buildMCPCallToolRequest(t.Name(), map[string]any{}), nil
}

func (t *ToolScreenShot) ReturnSchema() map[string]string {
	return map[string]string{
		"image": "string: Base64 encoded screenshot image in JPEG format",
		"name":  "string: Image name identifier (typically 'screenshot')",
		"type":  "string: MIME type of the image (image/jpeg)",
	}
}

// ToolGetScreenSize implements the get_screen_size tool call.
type ToolGetScreenSize struct{}

func (t *ToolGetScreenSize) Name() option.ActionName {
	return option.ACTION_GetScreenSize
}

func (t *ToolGetScreenSize) Description() string {
	return "Get the screen size of the mobile device in pixels"
}

func (t *ToolGetScreenSize) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_GetScreenSize)
}

func (t *ToolGetScreenSize) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		screenSize, err := driverExt.IDriver.WindowSize()
		if err != nil {
			return mcp.NewToolResultError("Get screen size failed: " + err.Error()), nil
		}
		return mcp.NewToolResultText(
			fmt.Sprintf("Screen size: %d x %d pixels", screenSize.Width, screenSize.Height),
		), nil
	}
}

func (t *ToolGetScreenSize) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	return buildMCPCallToolRequest(t.Name(), map[string]any{}), nil
}

func (t *ToolGetScreenSize) ReturnSchema() map[string]string {
	return map[string]string{
		"width":   "int: Screen width in pixels",
		"height":  "int: Screen height in pixels",
		"message": "string: Formatted message with screen dimensions",
	}
}

// ToolGetSource implements the get_source tool call.
type ToolGetSource struct{}

func (t *ToolGetSource) Name() option.ActionName {
	return option.ACTION_GetSource
}

func (t *ToolGetSource) Description() string {
	return "Get the UI hierarchy/source tree of the current screen for a specific app"
}

func (t *ToolGetSource) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_GetSource)
}

func (t *ToolGetSource) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// Get source action logic
		_, err = driverExt.Source(option.WithProcessName(unifiedReq.PackageName))
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Get source failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully retrieved source for package: %s", unifiedReq.PackageName)), nil
	}
}

func (t *ToolGetSource) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	if packageName, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"packageName": packageName,
		}
		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid get source params: %v", action.Params)
}

func (t *ToolGetSource) ReturnSchema() map[string]string {
	return map[string]string{
		"message":     "string: Success message confirming UI source was retrieved",
		"packageName": "string: Package name of the app whose source was retrieved",
		"source":      "string: UI hierarchy/source tree data in XML or JSON format",
	}
}
