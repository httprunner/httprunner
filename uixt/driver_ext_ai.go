package uixt

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/uixt/ai"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/httprunner/httprunner/v5/uixt/types"
)

// StartToGoal (original implementation - preserved)
func (dExt *XTDriver) StartToGoal(ctx context.Context, prompt string, opts ...option.ActionOption) ([]*PlanningExecutionResult, error) {
	options := option.NewActionOptions(opts...)
	logger := log.Info().Str("prompt", prompt)
	if options.MaxRetryTimes > 0 {
		logger = logger.Int("max_retry_times", options.MaxRetryTimes)
	}

	// Handle TimeLimit and Timeout with unified context mechanism
	var isTimeLimitMode bool
	if options.TimeLimit > 0 {
		// TimeLimit takes precedence over Timeout
		logger = logger.Int("time_limit_seconds", options.TimeLimit)
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(options.TimeLimit)*time.Second)
		defer cancel()
		isTimeLimitMode = true
	} else if options.Timeout > 0 {
		// Use Timeout only if TimeLimit is not set
		logger = logger.Int("timeout_seconds", options.Timeout)
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(options.Timeout)*time.Second)
		defer cancel()
	}
	logger.Msg("StartToGoal")

	var allPlannings []*PlanningExecutionResult
	var attempt int
	for {
		attempt++
		log.Info().Int("attempt", attempt).Msg("planning attempt")

		// Check for context cancellation (timeout, time limit, or interrupt)
		select {
		case <-ctx.Done():
			cause := context.Cause(ctx)
			// Handle TimeLimit timeout - return success
			if isTimeLimitMode && errors.Is(cause, context.DeadlineExceeded) {
				log.Info().
					Int("attempt", attempt).
					Int("completed_plannings", len(allPlannings)).
					Int("time_limit_seconds", options.TimeLimit).
					Msg("StartToGoal time limit reached, stopping gracefully")
				return allPlannings, nil
			}

			// Handle other cancellations (Timeout, interrupt, external cancellation) - return error
			log.Warn().
				Int("attempt", attempt).
				Int("completed_plannings", len(allPlannings)).
				Err(cause).
				Msg("StartToGoal cancelled")
			return allPlannings, errors.Wrap(cause, "StartToGoal cancelled")
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
				time.Sleep(5 * time.Second)
				continue
			}
			// Create planning result with error
			errorResult := &PlanningExecutionResult{
				PlanningResult: ai.PlanningResult{
					Thought:   "Planning failed",
					ModelName: "",
					Error:     err.Error(),
				},
				StartTime: planningStartTime.UnixMilli(),
				Elapsed:   time.Since(planningStartTime).Milliseconds(),
			}
			allPlannings = append(allPlannings, errorResult)
			return allPlannings, err
		}

		// Set planning execution timing
		planningResult.StartTime = planningStartTime.UnixMilli()
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
			// Check for context cancellation (timeout, time limit, or interrupt) before each action
			select {
			case <-ctx.Done():
				cause := context.Cause(ctx)
				// Handle TimeLimit timeout - return success
				if isTimeLimitMode && errors.Is(cause, context.DeadlineExceeded) {
					log.Info().
						Int("attempt", attempt).
						Int("completed_plannings", len(allPlannings)).
						Int("completed_tool_calls", len(planningResult.SubActions)).
						Int("total_tool_calls", len(planningResult.ToolCalls)).
						Int("time_limit_seconds", options.TimeLimit).
						Msg("StartToGoal time limit reached during tool call execution, stopping gracefully")
					planningResult.Elapsed = time.Since(planningStartTime).Milliseconds()
					allPlannings = append(allPlannings, planningResult)
					return allPlannings, nil
				}

				// Handle other cancellations (Timeout, external cancellation) - return error
				log.Warn().
					Int("attempt", attempt).
					Int("completed_plannings", len(allPlannings)).
					Int("completed_tool_calls", len(planningResult.SubActions)).
					Int("total_tool_calls", len(planningResult.ToolCalls)).
					Err(cause).
					Msg("invokeToolCalls cancelled")
				planningResult.Elapsed = time.Since(planningStartTime).Milliseconds()
				allPlannings = append(allPlannings, planningResult)
				return allPlannings, errors.Wrap(cause, "invokeToolCalls cancelled")
			default:
			}

			// Execute each tool call in a separate function to ensure proper defer execution
			err := func() error {
				subActionStartTime := time.Now()
				subActionResult := &SubActionResult{
					ActionName: toolCall.Function.Name,
					Arguments:  toolCall.Function.Arguments,
					StartTime:  subActionStartTime.UnixMilli(),
				}

				// Use defer to ensure sub-action is always processed and added to results
				defer func() {
					subActionResult.Elapsed = time.Since(subActionStartTime).Milliseconds()
					subActionResult.SessionData = dExt.GetSession().GetData(true) // reset after getting data
					planningResult.SubActions = append(planningResult.SubActions, subActionResult)
				}()

				if err := dExt.invokeToolCall(ctx, toolCall, opts...); err != nil {
					log.Error().Err(err).
						Str("action", toolCall.Function.Name).
						Msg("invoke tool call failed")
					subActionResult.Error = err.Error()
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

		if options.MaxRetryTimes > 0 && attempt > options.MaxRetryTimes {
			return allPlannings, errors.New("reached max retry times")
		}

		// wait 3 seconds for tool calls to complete
		time.Sleep(3 * time.Second)
	}
}

// AIAction with WingsService priority support
func (dExt *XTDriver) AIAction(ctx context.Context, prompt string, opts ...option.ActionOption) (*AIExecutionResult, error) {
	log.Info().Str("prompt", prompt).Msg("performing AI action")

	// Step 1: Take screenshot and convert to base64
	screenResult, err := dExt.GetScreenResult(
		option.WithScreenShotFileName("ai_action"),
		option.WithScreenShotBase64(true),
	)
	if err != nil {
		return nil, err
	}

	// Step 2: Check if WingsService is available and prioritize it
	if dExt.LLMService != nil {
		log.Info().Msg("using Wings service for AI action")
		return dExt.executeAIAction(ctx, prompt, screenResult, dExt.LLMService, "wings", opts...)
	} else {
		return nil, errors.New("no LLM service is initialized")
	}
}

// executeAIAction executes AIAction using any AI service (generic implementation)
func (dExt *XTDriver) executeAIAction(ctx context.Context, prompt string, screenResult *ScreenResult, service ai.ILLMService, serviceType string, opts ...option.ActionOption) (*AIExecutionResult, error) {
	// Step 1: Plan next action and measure time
	modelCallStartTime := time.Now()

	var planningResult *ai.PlanningResult
	var err error

	// For Wings service, call Plan directly
	planningOpts := &ai.PlanningOptions{
		UserInstruction: prompt,
		Message: &schema.Message{
			Role: schema.User,
			MultiContent: []schema.ChatMessagePart{
				{
					Type: schema.ChatMessagePartTypeImageURL,
					ImageURL: &schema.ChatMessageImageURL{
						URL: screenResult.Base64,
					},
				},
			},
		},
		Size: screenResult.Resolution,
	}

	planningResult, err = service.Plan(ctx, planningOpts)
	if err != nil {
		modelCallElapsed := time.Since(modelCallStartTime).Milliseconds()
		return &AIExecutionResult{
			Type:              "action",
			ModelCallElapsed:  modelCallElapsed,
			ScreenshotElapsed: screenResult.Elapsed,
			ImagePath:         screenResult.ImagePath,
			Resolution:        &screenResult.Resolution,
			Error:             err.Error(),
		}, errors.Wrap(err, fmt.Sprintf("%s service planning failed", serviceType))
	}
	modelCallElapsed := time.Since(modelCallStartTime).Milliseconds()

	aiExecutionResult := &AIExecutionResult{
		Type:              "action",
		ModelCallElapsed:  modelCallElapsed,
		ScreenshotElapsed: screenResult.Elapsed,
		ImagePath:         screenResult.ImagePath,
		Resolution:        &screenResult.Resolution,
		PlanningResult:    planningResult,
	}

	// Step 2: Execute tool calls
	for _, toolCall := range planningResult.ToolCalls {
		err = dExt.invokeToolCall(ctx, toolCall, opts...)
		if err != nil {
			log.Error().Err(err).
				Str("action", toolCall.Function.Name).
				Msg("invoke tool call failed")
			aiExecutionResult.Error = err.Error()
			return aiExecutionResult, errors.Wrap(err, "invoke tool call failed")
		}
	}

	return aiExecutionResult, nil
}

// AIAssert with WingsService priority support
func (dExt *XTDriver) AIAssert(assertion string, opts ...option.ActionOption) (*AIExecutionResult, error) {
	log.Info().Str("assertion", assertion).Msg("performing AI assertion")

	// Step 1: Take screenshot and convert to base64
	screenResult, err := dExt.GetScreenResult(
		option.WithScreenShotFileName("ai_assert"),
		option.WithScreenShotBase64(true),
	)
	if err != nil {
		return nil, err
	}

	if dExt.LLMService != nil {
		log.Info().Msg("using Wings service for AI assertion")
		return dExt.executeAIAssert(assertion, screenResult, dExt.LLMService, "wings", opts...)
	} else {
		return nil, errors.New("no LLM service is initialized")
	}
}

// executeAIAssert executes AIAssert using any AI service (generic implementation)
func (dExt *XTDriver) executeAIAssert(assertion string, screenResult *ScreenResult, service ai.ILLMService, serviceType string, opts ...option.ActionOption) (*AIExecutionResult, error) {
	// Step 1: Prepare context and options
	ctx := context.Background()

	assertResult := &AIExecutionResult{
		Type:              "assert",
		ScreenshotElapsed: screenResult.Elapsed,
		ImagePath:         screenResult.ImagePath,
		Resolution:        &screenResult.Resolution,
	}

	// Step 2: Call service and measure time
	modelCallStartTime := time.Now()
	assertOpts := &ai.AssertOptions{
		Assertion:  assertion,
		Screenshot: screenResult.Base64,
		Size:       screenResult.Resolution,
	}

	result, err := service.Assert(ctx, assertOpts)
	assertResult.ModelCallElapsed = time.Since(modelCallStartTime).Milliseconds()
	assertResult.AssertionResult = result

	if err != nil {
		assertResult.Error = err.Error()
		return assertResult, errors.Wrap(err, fmt.Sprintf("%s assertion failed", serviceType))
	}

	if !result.Pass {
		assertResult.Error = result.Thought
	}

	return assertResult, nil
}

// PlanNextAction (original implementation - preserved)
func (dExt *XTDriver) PlanNextAction(ctx context.Context, prompt string, opts ...option.ActionOption) (*PlanningExecutionResult, error) {
	if dExt.LLMService == nil {
		return nil, errors.New("LLM service is not initialized")
	}

	// Parse action options to get ResetHistory setting
	options := option.NewActionOptions(opts...)
	resetHistory := options.ResetHistory

	// Step 1: Take screenshot and convert to base64
	screenResult, err := dExt.GetScreenResult(
		option.WithScreenShotFileName("ai_planning"),
		option.WithScreenShotBase64(true),
	)
	if err != nil {
		return nil, err
	}

	// Clear session data after planning screenshot to avoid including it in sub-actions
	// The planning screenshot is already stored in planningResult.ScreenResult
	dExt.GetSession().GetData(true) // reset session data to exclude planning screenshot from sub-actions

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
						URL: screenResult.Base64,
					},
				},
			},
		},
		Size:         screenResult.Resolution,
		ResetHistory: resetHistory,
	}

	result, err := dExt.LLMService.Plan(ctx, planningOpts)
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
		ScreenshotElapsed: screenResult.Elapsed,
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

// isTaskFinished (original implementation - preserved)
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

// invokeToolCall (original implementation - preserved)
func (dExt *XTDriver) invokeToolCall(ctx context.Context, toolCall schema.ToolCall, opts ...option.ActionOption) error {
	// Parse arguments
	arguments := make(map[string]interface{})
	err := json.Unmarshal([]byte(toolCall.Function.Arguments), &arguments)
	if err != nil {
		return err
	}

	// Create a MobileAction with options to reuse BuildMCPCallToolRequest
	action := option.MobileAction{
		Options: option.NewActionOptions(opts...),
	}

	req := BuildMCPCallToolRequest(
		option.ActionName(toolCall.Function.Name),
		arguments,
		action,
	)
	_, err = dExt.client.CallTool(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

// PlanningExecutionResult (original implementation - preserved)
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

// AIExecutionResult (original implementation - preserved)
type AIExecutionResult struct {
	Type              string      `json:"type"`               // operation type: "query", "action", "assert"
	ModelCallElapsed  int64       `json:"model_call_elapsed"` // model call elapsed time in milliseconds
	ScreenshotElapsed int64       `json:"screenshot_elapsed"` // screenshot elapsed time in milliseconds
	ImagePath         string      `json:"image_path"`         // path to screenshot used for operation
	Resolution        *types.Size `json:"resolution"`         // screen resolution

	// Operation-specific results (only one will be populated based on Type)
	QueryResult     *ai.QueryResult     `json:"query_result,omitempty"`     // for ai_query operations
	PlanningResult  *ai.PlanningResult  `json:"planning_result,omitempty"`  // for ai_action operations
	AssertionResult *ai.AssertionResult `json:"assertion_result,omitempty"` // for ai_assert operations

	// Common fields
	Error string `json:"error,omitempty"` // error message if operation failed
}

// SubActionResult (original implementation - preserved)
type SubActionResult struct {
	ActionName string      `json:"action_name"`         // name of the sub-action (e.g., "tap", "input")
	Arguments  interface{} `json:"arguments,omitempty"` // arguments passed to the sub-action
	StartTime  int64       `json:"start_time"`          // sub-action start time
	Elapsed    int64       `json:"elapsed_ms"`          // sub-action elapsed time(ms)
	Error      string      `json:"error,omitempty"`     // sub-action execution result
	SessionData
}

type SessionData struct {
	Requests      []*DriverRequests `json:"requests,omitempty"`       // store sub-action specific requests
	ScreenResults []*ScreenResult   `json:"screen_results,omitempty"` // store sub-action specific screen_results
}

// AIQuery (original implementation - preserved)
func (dExt *XTDriver) AIQuery(text string, opts ...option.ActionOption) (*AIExecutionResult, error) {
	if dExt.LLMService == nil {
		return nil, errors.New("LLM service is not initialized")
	}

	// Step 1: Take screenshot and convert to base64
	screenResult, err := dExt.GetScreenResult(
		option.WithScreenShotFileName("ai_query"),
		option.WithScreenShotBase64(true),
	)
	if err != nil {
		return nil, err
	}

	// parse action options to extract OutputSchema
	actionOptions := option.NewActionOptions(opts...)

	// Step 2: Call model and measure time
	modelCallStartTime := time.Now()

	// execute query
	queryOpts := &ai.QueryOptions{
		Query:        text,
		Screenshot:   screenResult.Base64,
		Size:         screenResult.Resolution,
		OutputSchema: actionOptions.OutputSchema,
	}
	result, err := dExt.LLMService.Query(context.Background(), queryOpts)
	modelCallElapsed := time.Since(modelCallStartTime).Milliseconds()
	if err != nil {
		return nil, errors.Wrap(err, "AI query failed")
	}

	// Create AIExecutionResult with all timing and metadata
	aiResult := &AIExecutionResult{
		Type:              "query",
		ModelCallElapsed:  modelCallElapsed,         // model call timing
		ScreenshotElapsed: screenResult.Elapsed,     // screenshot timing
		ImagePath:         screenResult.ImagePath,   // screenshot path
		Resolution:        &screenResult.Resolution, // screen resolution
		QueryResult:       result,                   // query-specific result
	}
	return aiResult, nil
}
