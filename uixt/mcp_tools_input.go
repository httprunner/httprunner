package uixt

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/uixt/option"
)

// ToolInput implements the input tool call.
type ToolInput struct {
	// Return data fields - these define the structure of data returned by this tool
	Text string `json:"text" desc:"Text that was input"`
}

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
		arguments := request.GetArguments()
		driverExt, err := setupXTDriver(ctx, arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(arguments)
		if err != nil {
			return nil, err
		}

		if unifiedReq.Text == "" {
			return nil, fmt.Errorf("text is required")
		}

		opts := unifiedReq.Options()

		// Input action logic
		err = driverExt.Input(unifiedReq.Text, opts...)
		if err != nil {
			return NewMCPErrorResponse(fmt.Sprintf("Input failed: %s", err.Error())), err
		}

		message := fmt.Sprintf("Successfully input text: %s", unifiedReq.Text)
		returnData := ToolInput{Text: unifiedReq.Text}

		return NewMCPSuccessResponse(message, &returnData), nil
	}
}

func (t *ToolInput) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	text := fmt.Sprintf("%v", action.Params)
	arguments := map[string]any{
		"text": text,
	}
	return BuildMCPCallToolRequest(t.Name(), arguments, action), nil
}

// ToolSetIme implements the set_ime tool call.
type ToolSetIme struct {
	// Return data fields - these define the structure of data returned by this tool
	Ime string `json:"ime" desc:"IME that was set"`
}

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
		arguments := request.GetArguments()
		driverExt, err := setupXTDriver(ctx, arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(arguments)
		if err != nil {
			return nil, err
		}

		// Set IME action logic
		err = driverExt.SetIme(unifiedReq.Ime)
		if err != nil {
			return NewMCPErrorResponse(fmt.Sprintf("Set IME failed: %s", err.Error())), err
		}

		message := fmt.Sprintf("Successfully set IME to: %s", unifiedReq.Ime)
		returnData := ToolSetIme{Ime: unifiedReq.Ime}

		return NewMCPSuccessResponse(message, &returnData), nil
	}
}

func (t *ToolSetIme) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	if ime, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"ime": ime,
		}
		return BuildMCPCallToolRequest(t.Name(), arguments, action), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid set ime params: %v", action.Params)
}

// ToolBackspace implements the backspace tool call.
type ToolBackspace struct {
	// Return data fields - these define the structure of data returned by this tool
	Count int `json:"count" desc:"Number of backspace operations performed"`
}

func (t *ToolBackspace) Name() option.ActionName {
	return option.ACTION_Backspace
}

func (t *ToolBackspace) Description() string {
	return "Perform backspace operations to delete characters"
}

func (t *ToolBackspace) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_Backspace)
}

func (t *ToolBackspace) Implement() server.ToolHandlerFunc {
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

		count := unifiedReq.Count
		if count <= 0 {
			count = 1 // Default to 1 backspace if not specified or invalid
		}

		opts := unifiedReq.Options()

		// Backspace action logic
		err = driverExt.Backspace(count, opts...)
		if err != nil {
			return NewMCPErrorResponse(fmt.Sprintf("Backspace failed: %s", err.Error())), err
		}

		message := fmt.Sprintf("Successfully performed %d backspace operations", count)
		returnData := ToolBackspace{Count: count}

		return NewMCPSuccessResponse(message, &returnData), nil
	}
}

func (t *ToolBackspace) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	var count int
	switch v := action.Params.(type) {
	case int:
		count = v
	case float64:
		count = int(v)
	default:
		count = 1 // Default count
	}

	arguments := map[string]any{
		"count": count,
	}
	return BuildMCPCallToolRequest(t.Name(), arguments, action), nil
}

// ToolSIMInput implements the sim_input tool call.
type ToolSIMInput struct {
	// Return data fields - these define the structure of data returned by this tool
	Text     string `json:"text" desc:"Text that was input with simulation"`
	Segments int    `json:"segments" desc:"Number of segments the text was split into"`
}

func (t *ToolSIMInput) Name() option.ActionName {
	return option.ACTION_SIMInput
}

func (t *ToolSIMInput) Description() string {
	return "Input text with intelligent segmentation and human-like typing patterns"
}

func (t *ToolSIMInput) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_SIMInput)
}

func (t *ToolSIMInput) Implement() server.ToolHandlerFunc {
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

		if unifiedReq.Text == "" {
			return nil, fmt.Errorf("text is required")
		}

		text := unifiedReq.Text

		log.Info().
			Str("text", text).
			Int("textLength", len(text)).
			Msg("performing simulated input")

		opts := unifiedReq.Options()

		// Call the underlying SIMInput method (check if driver supports SIM)
		if simDriver, ok := driverExt.IDriver.(SIMSupport); ok {
			err = simDriver.SIMInput(text, opts...)
			if err != nil {
				return NewMCPErrorResponse(fmt.Sprintf("Simulated input failed: %s", err.Error())), err
			}
		} else {
			return NewMCPErrorResponse("SIMInput is not supported by the current driver"), fmt.Errorf("driver does not implement SIMSupport interface")
		}

		// Estimate segments count (this is approximate since the actual segmentation happens in the driver)
		estimatedSegments := len([]rune(text))/2 + 1
		if estimatedSegments < 1 {
			estimatedSegments = 1
		}

		message := fmt.Sprintf("Successfully performed simulated input: %s", text)
		returnData := ToolSIMInput{
			Text:     text,
			Segments: estimatedSegments,
		}

		return NewMCPSuccessResponse(message, &returnData), nil
	}
}

func (t *ToolSIMInput) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	text := fmt.Sprintf("%v", action.Params)
	arguments := map[string]any{
		"text": text,
	}
	return BuildMCPCallToolRequest(t.Name(), arguments, action), nil
}
