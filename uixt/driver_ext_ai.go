package uixt

import (
	"context"
	"encoding/base64"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/internal/json"
	"github.com/httprunner/httprunner/v5/uixt/ai"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func (dExt *XTDriver) StartToGoal(ctx context.Context, prompt string, opts ...option.ActionOption) ([]*SubActionResult, error) {
	options := option.NewActionOptions(opts...)
	log.Info().Int("max_retry_times", options.MaxRetryTimes).Msg("StartToGoal")

	var allSubActions []*SubActionResult
	var attempt int
	for {
		attempt++
		log.Info().Int("attempt", attempt).Msg("planning attempt")

		// Check for context cancellation (interrupt signal)
		select {
		case <-ctx.Done():
			log.Warn().Msg("interrupted in StartToGoal")
			return allSubActions, errors.Wrap(code.InterruptError, "StartToGoal interrupted")
		default:
		}

		// Plan next action with history reset on first attempt
		planningOpts := opts
		if attempt == 1 {
			// Add ResetHistory option for the first attempt
			planningOpts = append(planningOpts, option.WithResetHistory(true))
		}
		result, err := dExt.PlanNextAction(ctx, prompt, planningOpts...)
		if err != nil {
			// Check if this is a LLM service request error that should be retried
			if errors.Is(err, code.LLMRequestServiceError) {
				log.Warn().Err(err).Int("attempt", attempt).
					Msg("LLM service request failed, retrying...")
				continue
			}
			return allSubActions, err
		}

		// Check if task is finished BEFORE executing actions
		if dExt.isTaskFinished(result) {
			log.Info().Msg("task finished, stopping StartToGoal")
			return allSubActions, nil
		}

		// Invoke tool calls
		subActions, err := dExt.invokeToolCalls(ctx, result.Thought, result.ToolCalls)
		allSubActions = append(allSubActions, subActions...)
		if err != nil {
			return allSubActions, err
		}

		if options.MaxRetryTimes > 1 && attempt >= options.MaxRetryTimes {
			return allSubActions, errors.New("reached max retry times")
		}
	}
}

func (dExt *XTDriver) AIAction(ctx context.Context, prompt string, opts ...option.ActionOption) ([]*SubActionResult, error) {
	log.Info().Str("prompt", prompt).Msg("performing AI action")

	// plan next action
	result, err := dExt.PlanNextAction(ctx, prompt, opts...)
	if err != nil {
		return nil, err
	}

	// Invoke tool calls
	subActionResults, err := dExt.invokeToolCalls(ctx, result.Thought, result.ToolCalls)
	if err != nil {
		return subActionResults, err
	}

	return subActionResults, nil
}

func (dExt *XTDriver) PlanNextAction(ctx context.Context, prompt string, opts ...option.ActionOption) (*ai.PlanningResult, error) {
	if dExt.LLMService == nil {
		return nil, errors.New("LLM service is not initialized")
	}

	// Parse action options to get ResetHistory setting
	options := option.NewActionOptions(opts...)
	resetHistory := options.ResetHistory

	// Use GetScreenResult to handle screenshot capture, save, and session tracking
	screenResult, err := dExt.GetScreenResult(
		option.WithScreenShotFileName(builtin.GenNameWithTimestamp("%d_screenshot")),
	)
	if err != nil {
		return nil, err
	}

	// convert buffer to base64 string for LLM
	screenShotBase64 := "data:image/jpeg;base64," +
		base64.StdEncoding.EncodeToString(screenResult.bufSource.Bytes())

	// get window size
	size, err := dExt.IDriver.WindowSize()
	if err != nil {
		return nil, errors.Wrap(code.DeviceGetInfoError, err.Error())
	}

	planningOpts := &ai.PlanningOptions{
		UserInstruction: prompt,
		Message: &schema.Message{
			Role: schema.User,
			MultiContent: []schema.ChatMessagePart{
				{
					Type: schema.ChatMessagePartTypeImageURL,
					ImageURL: &schema.ChatMessageImageURL{
						URL: screenShotBase64,
					},
				},
			},
		},
		Size:         size,
		ResetHistory: resetHistory,
	}

	result, err := dExt.LLMService.Call(ctx, planningOpts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get next action from planner")
	}
	return result, nil
}

// isTaskFinished checks if the task is completed based on the planning result
func (dExt *XTDriver) isTaskFinished(result *ai.PlanningResult) bool {
	// Check if there are no tool calls (no actions to execute)
	if len(result.ToolCalls) == 0 {
		log.Info().Msg("no tool calls returned, task may be finished")
		return true
	}

	// Check if any tool call is a "finished" action
	for _, toolCall := range result.ToolCalls {
		if toolCall.Function.Name == "uixt__finished" {
			log.Info().Str("reason", toolCall.Function.Arguments).Msg("finished action detected")
			return true
		}
	}

	return false
}

// invokeToolCalls invokes the tool calls and returns sub-action results
func (dExt *XTDriver) invokeToolCalls(ctx context.Context, thought string, toolCalls []schema.ToolCall) ([]*SubActionResult, error) {
	var subActionResults []*SubActionResult

	for _, action := range toolCalls {
		// Check for context cancellation before each action
		select {
		case <-ctx.Done():
			log.Warn().Msg("interrupted in invokeToolCalls")
			return subActionResults, errors.Wrap(code.InterruptError, "invokeToolCalls interrupted")
		default:
		}

		subActionStartTime := time.Now()

		// Extract action name (remove "uixt__" prefix)
		actionName := strings.TrimPrefix(action.Function.Name, "uixt__")

		// Parse arguments
		arguments := make(map[string]interface{})
		err := json.Unmarshal([]byte(action.Function.Arguments), &arguments)
		if err != nil {
			return subActionResults, err
		}

		// Create sub-action result
		subActionResult := &SubActionResult{
			ActionName: actionName,
			Arguments:  arguments,
			StartTime:  subActionStartTime.Unix(),
			Thought:    thought,
		}

		// Execute the action
		req := mcp.CallToolRequest{
			Params: struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments,omitempty"`
				Meta      *struct {
					ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
				} `json:"_meta,omitempty"`
			}{
				Name:      action.Function.Name,
				Arguments: arguments,
			},
		}

		_, err = dExt.client.CallTool(ctx, req)
		subActionResult.Elapsed = time.Since(subActionStartTime).Milliseconds()
		if err != nil {
			subActionResult.Error = err
			subActionResults = append(subActionResults, subActionResult)
			return subActionResults, err
		}

		// Collect sub-action specific attachments and reset session data
		subActionData := dExt.GetData(true) // reset after getting data

		// Add requests if any
		if requests, ok := subActionData["requests"].([]*DriverRequests); ok && len(requests) > 0 {
			subActionResult.Requests = requests
		}

		// Add screen_results if any
		if screenResults, ok := subActionData["screen_results"].([]*ScreenResult); ok && len(screenResults) > 0 {
			subActionResult.ScreenResults = screenResults
		}

		subActionResults = append(subActionResults, subActionResult)
	}

	return subActionResults, nil
}

// SubActionResult represents a sub-action within a start_to_goal action
type SubActionResult struct {
	ActionName    string            `json:"action_name"`              // name of the sub-action (e.g., "tap", "input")
	Arguments     interface{}       `json:"arguments,omitempty"`      // arguments passed to the sub-action
	StartTime     int64             `json:"start_time"`               // sub-action start time
	Elapsed       int64             `json:"elapsed_ms"`               // sub-action elapsed time(ms)
	Error         error             `json:"error,omitempty"`          // sub-action execution result
	Thought       string            `json:"thought,omitempty"`        // sub-action thought
	Requests      []*DriverRequests `json:"requests,omitempty"`       // store sub-action specific requests
	ScreenResults []*ScreenResult   `json:"screen_results,omitempty"` // store sub-action specific screen_results
}

func (dExt *XTDriver) AIQuery(text string, opts ...option.ActionOption) (string, error) {
	return "", nil
}

func (dExt *XTDriver) AIAssert(assertion string, opts ...option.ActionOption) error {
	if dExt.LLMService == nil {
		return errors.New("LLM service is not initialized")
	}

	screenShotBase64, err := GetScreenShotBufferBase64(dExt.IDriver)
	if err != nil {
		return err
	}

	// get window size
	size, err := dExt.IDriver.WindowSize()
	if err != nil {
		return errors.Wrap(err, "get window size for AI assertion failed")
	}

	// execute assertion
	assertOpts := &ai.AssertOptions{
		Assertion:  assertion,
		Screenshot: screenShotBase64,
		Size:       size,
	}
	result, err := dExt.LLMService.Assert(context.Background(), assertOpts)
	if err != nil {
		return errors.Wrap(err, "AI assertion failed")
	}

	if !result.Pass {
		return errors.New(result.Thought)
	}

	return nil
}
