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
type ToolStartToGoal struct {
	// Return data fields - these define the structure of data returned by this tool
	Prompt string `json:"prompt" desc:"Goal prompt that was executed"`
}

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
		_, err = driverExt.StartToGoal(ctx, unifiedReq.Prompt)
		if err != nil {
			return NewMCPErrorResponse(fmt.Sprintf("Failed to achieve goal: %s", err.Error())), nil
		}

		message := fmt.Sprintf("Successfully achieved goal: %s", unifiedReq.Prompt)
		returnData := ToolStartToGoal{
			Prompt: unifiedReq.Prompt,
		}

		return NewMCPSuccessResponse(message, &returnData), nil
	}
}

func (t *ToolStartToGoal) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	if prompt, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"prompt": prompt,
		}

		// Extract options to arguments
		extractActionOptionsToArguments(action.GetOptions(), arguments)

		return BuildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid start to goal params: %v", action.Params)
}

// ToolAIAction implements the ai_action tool call.
type ToolAIAction struct {
	// Return data fields - these define the structure of data returned by this tool
	Prompt string `json:"prompt" desc:"AI action prompt that was executed"`
}

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
		_, err = driverExt.AIAction(ctx, unifiedReq.Prompt)
		if err != nil {
			return NewMCPErrorResponse(fmt.Sprintf("AI action failed: %s", err.Error())), nil
		}

		message := fmt.Sprintf("Successfully performed AI action with prompt: %s", unifiedReq.Prompt)
		returnData := ToolAIAction{
			Prompt: unifiedReq.Prompt,
		}

		return NewMCPSuccessResponse(message, &returnData), nil
	}
}

func (t *ToolAIAction) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	if prompt, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"prompt": prompt,
		}

		// Extract options to arguments
		extractActionOptionsToArguments(action.GetOptions(), arguments)

		return BuildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid AI action params: %v", action.Params)
}

// ToolAIQuery implements the ai_query tool call.
type ToolAIQuery struct {
	// Return data fields - these define the structure of data returned by this tool
	Prompt string `json:"prompt" desc:"AI query prompt that was executed"`
	Result string `json:"result" desc:"Query result content"`
}

func (t *ToolAIQuery) Name() option.ActionName {
	return option.ACTION_Query
}

func (t *ToolAIQuery) Description() string {
	return "Query information from screen using AI vision model with natural language prompts"
}

func (t *ToolAIQuery) Options() []mcp.ToolOption {
	unifiedReq := &option.ActionOptions{}
	return unifiedReq.GetMCPOptions(option.ACTION_Query)
}

func (t *ToolAIQuery) Implement() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		driverExt, err := setupXTDriver(ctx, request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("setup driver failed: %w", err)
		}

		unifiedReq, err := parseActionOptions(request.Params.Arguments)
		if err != nil {
			return nil, err
		}

		// Build action options from unified request
		opts := unifiedReq.Options()

		// AI query logic with options
		queryResult, err := driverExt.AIQuery(unifiedReq.Prompt, opts...)
		if err != nil {
			return NewMCPErrorResponse(fmt.Sprintf("AI query failed: %s", err.Error())), nil
		}

		message := fmt.Sprintf("Successfully queried information with prompt: %s", unifiedReq.Prompt)
		returnData := ToolAIQuery{
			Prompt: unifiedReq.Prompt,
			Result: queryResult.Content,
		}

		return NewMCPSuccessResponse(message, &returnData), nil
	}
}

func (t *ToolAIQuery) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	if prompt, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"prompt": prompt,
		}

		// Extract options to arguments
		extractActionOptionsToArguments(action.GetOptions(), arguments)

		return BuildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid AI query params: %v", action.Params)
}

// ToolFinished implements the finished tool call.
type ToolFinished struct {
	// Return data fields - these define the structure of data returned by this tool
	Content string `json:"content" desc:"Task completion reason or result message"`
}

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

		message := fmt.Sprintf("Task completed: %s", unifiedReq.Content)
		returnData := ToolFinished{
			Content: unifiedReq.Content,
		}

		return NewMCPSuccessResponse(message, &returnData), nil
	}
}

func (t *ToolFinished) ConvertActionToCallToolRequest(action option.MobileAction) (mcp.CallToolRequest, error) {
	if reason, ok := action.Params.(string); ok {
		arguments := map[string]any{
			"content": reason,
		}
		return BuildMCPCallToolRequest(t.Name(), arguments), nil
	}
	return mcp.CallToolRequest{}, fmt.Errorf("invalid finished params: %v", action.Params)
}
