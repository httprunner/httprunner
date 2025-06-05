package uixt

import (
	"context"
	"fmt"

	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rs/zerolog/log"
)

// ToolStartToGoal implements the start_to_goal tool call.
type ToolStartToGoal struct{}

func (t *ToolStartToGoal) Name() option.ActionName {
	return option.ACTION_StartToGoal
}

func (t *ToolStartToGoal) Description() string {
	return "Start AI-driven automation to achieve a specific goal using natural language description"
}

func (t *ToolStartToGoal) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_StartToGoal)
}

func (t *ToolStartToGoal) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// Start to goal logic
		err = driverExt.StartToGoal(ctx, unifiedReq.Prompt)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to achieve goal: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully achieved goal: %s", unifiedReq.Prompt)), nil
	}
}

func (t *ToolStartToGoal) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	if prompt, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"prompt": prompt,
		}

		// Extract options to arguments
		extractActionOptionsToArguments(action.GetOptions(), arguments)

		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid start to goal params: %v", action.Params)
}

func (t *ToolStartToGoal) ReturnSchema() map[string]string {
	return map[string]string{
		"message": "string: Success message confirming goal was achieved, or error message if failed",
	}
}

// ToolAIAction implements the ai_action tool call.
type ToolAIAction struct{}

func (t *ToolAIAction) Name() option.ActionName {
	return option.ACTION_AIAction
}

func (t *ToolAIAction) Description() string {
	return "Perform AI-driven automation actions using natural language prompts to describe the desired operation"
}

func (t *ToolAIAction) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_AIAction)
}

func (t *ToolAIAction) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// AI action logic
		err = driverExt.AIAction(ctx, unifiedReq.Prompt)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("AI action failed: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully performed AI action with prompt: %s", unifiedReq.Prompt)), nil
	}
}

func (t *ToolAIAction) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	if prompt, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"prompt": prompt,
		}

		// Extract options to arguments
		extractActionOptionsToArguments(action.GetOptions(), arguments)

		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid AI action params: %v", action.Params)
}

func (t *ToolAIAction) ReturnSchema() map[string]string {
	return map[string]string{
		"message": "string: Success message confirming AI action was performed, or error message if failed",
	}
}

// ToolFinished implements the finished tool call.
type ToolFinished struct{}

func (t *ToolFinished) Name() option.ActionName {
	return option.ACTION_Finished
}

func (t *ToolFinished) Description() string {
	return "Mark the current automation task as completed with a result message"
}

func (t *ToolFinished) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_Finished)
}

func (t *ToolFinished) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}
		log.Info().Str("reason", unifiedReq.Content).Msg("task finished")

		return mcp.NewToolResultText(fmt.Sprintf("Task completed: %s", unifiedReq.Content)), nil
	}
}

func (t *ToolFinished) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	if reason, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"content": reason,
		}
		return buildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid finished params: %v", action.Params)
}

func (t *ToolFinished) ReturnSchema() map[string]string {
	return map[string]string{
		"message": "string: Success message confirming task completion, or error message if failed",
	}
}
