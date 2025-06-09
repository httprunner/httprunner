package uixt

import (
	"context"
	"encoding/base64"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/internal/json"
	"github.com/httprunner/httprunner/v5/uixt/ai"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/httprunner/httprunner/v5/uixt/types"
)

func (dExt *XTDriver) StartToGoal(ctx context.Context, prompt string, opts ...option.ActionOption) ([]*PlanningExecutionResult, error) {
	options := option.NewActionOptions(opts...)
	log.Info().Int("max_retry_times", options.MaxRetryTimes).Msg("StartToGoal")

	var allPlannings []*PlanningExecutionResult
	var attempt int
	for {
		attempt++
		log.Info().Int("attempt", attempt).Msg("planning attempt")

		// Check for context cancellation (interrupt signal)
		select {
		case <-ctx.Done():
			log.Warn().Msg("interrupted in StartToGoal")
			return allPlannings, errors.Wrap(code.InterruptError, "StartToGoal interrupted")
		default:
		}

		// Plan next action with history reset on first attempt
		planningStartTime := time.Now()
		planningOpts := opts
		if attempt == 1 {
			// Add ResetHistory option for the first attempt
			planningOpts = append(planningOpts, option.WithResetHistory(true))
		}

		planningResult, err := dExt.PlanNextAction(ctx, prompt, planningOpts...)
		if err != nil {
			// Check if this is a LLM service request error that should be retried
			if errors.Is(err, code.LLMRequestServiceError) {
				log.Warn().Err(err).Int("attempt", attempt).
					Msg("LLM service request failed, retrying...")
				continue
			}
			// Create planning result with error
			errorResult := &PlanningExecutionResult{
				PlanningResult: ai.PlanningResult{
					Thought:   "Planning failed",
					ModelName: "",
					Error:     err.Error(),
				},
				StartTime: planningStartTime.Unix(),
				Elapsed:   time.Since(planningStartTime).Milliseconds(),
			}
			allPlannings = append(allPlannings, errorResult)
			return allPlannings, err
		}

		// Set planning execution timing
		planningResult.StartTime = planningStartTime.Unix()
		planningResult.SubActions = []*SubActionResult{}

		// Check if task is finished BEFORE executing actions
		if dExt.isTaskFinished(planningResult) {
			log.Info().Msg("task finished, stopping StartToGoal")
			planningResult.Elapsed = time.Since(planningStartTime).Milliseconds()
			allPlannings = append(allPlannings, planningResult)
			return allPlannings, nil
		}

		// Invoke tool calls
		for _, toolCall := range planningResult.ToolCalls {
			// Check for context cancellation before each action
			select {
			case <-ctx.Done():
				log.Warn().Msg("interrupted in invokeToolCalls")
				planningResult.Elapsed = time.Since(planningStartTime).Milliseconds()
				allPlannings = append(allPlannings, planningResult)
				return allPlannings, errors.Wrap(code.InterruptError, "invokeToolCalls interrupted")
			default:
			}

			// Execute each tool call in a separate function to ensure proper defer execution
			err := func() error {
				subActionStartTime := time.Now()
				subActionResult := &SubActionResult{
					ActionName: toolCall.Function.Name,
					Arguments:  toolCall.Function.Arguments,
					StartTime:  subActionStartTime.Unix(),
				}

				// Use defer to ensure sub-action is always processed and added to results
				defer func() {
					subActionResult.Elapsed = time.Since(subActionStartTime).Milliseconds()
					subActionResult.SessionData = dExt.GetSession().GetData(true) // reset after getting data
					planningResult.SubActions = append(planningResult.SubActions, subActionResult)
				}()

				// Execute the tool call
				if err := dExt.invokeToolCall(ctx, toolCall); err != nil {
					subActionResult.Error = err
					return err
				}
				return nil
			}()
			if err != nil {
				planningResult.Elapsed = time.Since(planningStartTime).Milliseconds()
				planningResult.Error = err.Error()
				allPlannings = append(allPlannings, planningResult)
				return allPlannings, err
			}
		}

		// Complete this planning cycle
		planningResult.Elapsed = time.Since(planningStartTime).Milliseconds()
		allPlannings = append(allPlannings, planningResult)

		if options.MaxRetryTimes > 1 && attempt >= options.MaxRetryTimes {
			return allPlannings, errors.New("reached max retry times")
		}
	}
}

func (dExt *XTDriver) AIAction(ctx context.Context, prompt string, opts ...option.ActionOption) error {
	log.Info().Str("prompt", prompt).Msg("performing AI action")

	// plan next action
	planningResult, err := dExt.PlanNextAction(ctx, prompt, opts...)
	if err != nil {
		return err
	}

	// Invoke tool calls
	for _, toolCall := range planningResult.ToolCalls {
		err = dExt.invokeToolCall(ctx, toolCall)
		if err != nil {
			return err
		}
	}

	return nil
}

// PlanNextAction performs planning and returns unified planning information
func (dExt *XTDriver) PlanNextAction(ctx context.Context, prompt string, opts ...option.ActionOption) (*PlanningExecutionResult, error) {
	if dExt.LLMService == nil {
		return nil, errors.New("LLM service is not initialized")
	}

	// Parse action options to get ResetHistory setting
	options := option.NewActionOptions(opts...)
	resetHistory := options.ResetHistory

	// Step 1: Take screenshot
	screenshotStartTime := time.Now()
	// Use GetScreenResult to handle screenshot capture, save, and session tracking
	screenResult, err := dExt.GetScreenResult(
		option.WithScreenShotFileName(builtin.GenNameWithTimestamp("%d_screenshot")),
	)
	screenshotElapsed := time.Since(screenshotStartTime).Milliseconds()
	if err != nil {
		return nil, err
	}

	// Clear session data after planning screenshot to avoid including it in sub-actions
	// The planning screenshot is already stored in planningResult.ScreenResult
	dExt.GetSession().GetData(true) // reset session data to exclude planning screenshot from sub-actions

	// convert buffer to base64 string for LLM
	screenShotBase64 := "data:image/jpeg;base64," +
		base64.StdEncoding.EncodeToString(screenResult.bufSource.Bytes())

	// get window size
	size, err := dExt.IDriver.WindowSize()
	if err != nil {
		return nil, errors.Wrap(code.DeviceGetInfoError, err.Error())
	}

	// Step 2: Call model
	modelCallStartTime := time.Now()
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
	modelCallElapsed := time.Since(modelCallStartTime).Milliseconds()

	if err != nil {
		return nil, errors.Wrap(err, "failed to get next action from planner")
	}

	// Step 3: Parse result (this is already done in LLMService.Call, but we record it separately)
	actionNames := make([]string, len(result.ToolCalls))
	for i, toolCall := range result.ToolCalls {
		actionNames[i] = toolCall.Function.Name
	}

	// Create unified planning result that inherits from ai.PlanningResult
	planningResult := &PlanningExecutionResult{
		PlanningResult: *result, // Inherit all fields from ai.PlanningResult
		// Planning process timing and metadata
		ScreenshotElapsed: screenshotElapsed,
		ImagePath:         screenResult.ImagePath,
		Resolution:        &screenResult.Resolution,
		ScreenResult:      screenResult,
		ModelCallElapsed:  modelCallElapsed,
		ToolCallsCount:    len(result.ToolCalls),
		ActionNames:       actionNames,
		// Execution timing (will be set by StartToGoal)
		StartTime:  0,   // Will be set by caller
		Elapsed:    0,   // Will be set by caller
		SubActions: nil, // Will be populated during execution
	}

	return planningResult, nil
}

// isTaskFinished checks if the task is completed based on the planning result
func (dExt *XTDriver) isTaskFinished(planningResult *PlanningExecutionResult) bool {
	// Check if there are no tool calls (no actions to execute)
	if len(planningResult.ToolCalls) == 0 {
		log.Info().Msg("no tool calls returned, task may be finished")
		return true
	}

	// Check if any tool call is a "finished" action
	for _, toolCall := range planningResult.ToolCalls {
		if toolCall.Function.Name == "uixt__finished" {
			log.Info().Str("reason", toolCall.Function.Arguments).Msg("finished action detected")
			return true
		}
	}

	return false
}

// invokeToolCall invokes the tool call
func (dExt *XTDriver) invokeToolCall(ctx context.Context, toolCall schema.ToolCall) error {
	// Parse arguments
	arguments := make(map[string]interface{})
	err := json.Unmarshal([]byte(toolCall.Function.Arguments), &arguments)
	if err != nil {
		return err
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
			Name:      toolCall.Function.Name,
			Arguments: arguments,
		},
	}

	_, err = dExt.client.CallTool(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

// PlanningExecutionResult represents a unified planning result that contains both planning information and execution results
type PlanningExecutionResult struct {
	ai.PlanningResult // Inherit all fields from ai.PlanningResult (ToolCalls, Thought, Content, Error, ModelName)
	// Planning process information
	ScreenshotElapsed int64         `json:"screenshot_elapsed_ms"` // screenshot elapsed time(ms)
	ImagePath         string        `json:"image_path"`            // screenshot image path
	Resolution        *types.Size   `json:"resolution"`            // image resolution
	ScreenResult      *ScreenResult `json:"screen_result"`         // complete screen result data
	ModelCallElapsed  int64         `json:"model_call_elapsed_ms"` // model call elapsed time(ms)
	ToolCallsCount    int           `json:"tool_calls_count"`      // number of tool calls generated
	ActionNames       []string      `json:"action_names"`          // names of parsed actions
	// Execution information
	StartTime  int64              `json:"start_time"`            // planning start time
	Elapsed    int64              `json:"elapsed_ms"`            // planning elapsed time(ms)
	SubActions []*SubActionResult `json:"sub_actions,omitempty"` // sub-actions generated from this planning
}

// SubActionResult represents a sub-action within a start_to_goal action
type SubActionResult struct {
	ActionName string      `json:"action_name"`         // name of the sub-action (e.g., "tap", "input")
	Arguments  interface{} `json:"arguments,omitempty"` // arguments passed to the sub-action
	StartTime  int64       `json:"start_time"`          // sub-action start time
	Elapsed    int64       `json:"elapsed_ms"`          // sub-action elapsed time(ms)
	Error      error       `json:"error,omitempty"`     // sub-action execution result
	SessionData
}

type SessionData struct {
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
