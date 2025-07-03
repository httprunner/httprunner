package uixt

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/uixt/option"
)

// ToolScreenShot implements the screenshot tool call.
type ToolScreenShot struct { // Return data fields - these define the structure of data returned by this tool
	// Note: This tool returns image data, not JSON, so no additional fields needed
}

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
		driverExt, err := setupXTDriver(ctx, request.GetArguments())
		if err != nil {
			return nil, err
		}
		screenResult, err := driverExt.GetScreenResult(
			option.WithScreenShotFileName("tool_screenshot"),
			option.WithScreenShotBase64(true),
		)
		if err != nil {
			log.Error().Err(err).Msg("ScreenShot failed")
			return mcp.NewToolResultError(fmt.Sprintf("Failed to take screenshot: %v", err)), nil
		}
		log.Debug().Int("imageBytes", len(screenResult.Base64)).Msg("take screenshot success")

		return mcp.NewToolResultImage("screenshot", screenResult.Base64, "image/jpeg"), nil
	}
}

func (t *ToolScreenShot) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	return BuildMCPCallToolRequest(t.Name(), map[string]any{}, action), nil
}

// ToolGetScreenSize implements the get_screen_size tool call.
type ToolGetScreenSize struct {
	// Return data fields - these define the structure of data returned by this tool
	Width  int `json:"width" desc:"Screen width in pixels"`
	Height int `json:"height" desc:"Screen height in pixels"`
}

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
		driverExt, err := setupXTDriver(ctx, request.GetArguments())
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		screenSize, err := driverExt.IDriver.WindowSize()
		if err != nil {
			return NewMCPErrorResponse("Get screen size failed: " + err.Error()), nil
		}

		message := fmt.Sprintf("Screen size: %d x %d pixels", screenSize.Width, screenSize.Height)
		returnData := ToolGetScreenSize{
			Width:  screenSize.Width,
			Height: screenSize.Height,
		}

		return NewMCPSuccessResponse(message, &returnData), nil
	}
}

func (t *ToolGetScreenSize) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	return BuildMCPCallToolRequest(t.Name(), map[string]any{}, action), nil
}

// ToolGetSource implements the get_source tool call.
type ToolGetSource struct {
	// Return data fields - these define the structure of data returned by this tool
	PackageName string `json:"packageName" desc:"Package name of the app whose source was retrieved"`
	Source      string `json:"source" desc:"UI hierarchy/source tree data in XML or JSON format"`
}

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
		arguments := request.GetArguments()
		driverExt, err := setupXTDriver(ctx, arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(arguments)
		if err != nil {
			return nil, err
		}

		// Get source action logic
		sourceData, err := driverExt.Source(option.WithProcessName(unifiedReq.PackageName))
		if err != nil {
			return NewMCPErrorResponse(fmt.Sprintf("Get source failed: %s", err.Error())), err
		}

		message := fmt.Sprintf("Successfully retrieved source for package: %s", unifiedReq.PackageName)
		returnData := ToolGetSource{
			PackageName: unifiedReq.PackageName,
			Source:      sourceData,
		}

		return NewMCPSuccessResponse(message, &returnData), nil
	}
}

func (t *ToolGetSource) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	if packageName, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"packageName": packageName,
		}
		return BuildMCPCallToolRequest(t.Name(), arguments, action), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid get source params: %v", action.Params)
}
