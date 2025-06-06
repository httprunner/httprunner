package uixt

import (
	"context"
	"encoding/base64"
	"fmt"
	"path/filepath"

	"github.com/cloudwego/eino/schema"
	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/internal/config"
	"github.com/httprunner/httprunner/v5/internal/json"
	"github.com/httprunner/httprunner/v5/uixt/ai"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func (dExt *XTDriver) StartToGoal(ctx context.Context, prompt string, opts ...option.ActionOption) error {
	options := option.NewActionOptions(opts...)
	log.Info().Int("max_retry_times", options.MaxRetryTimes).Msg("StartToGoal")

	var attempt int
	for {
		attempt++
		log.Info().Int("attempt", attempt).Msg("planning attempt")

		// Check for context cancellation (interrupt signal)
		select {
		case <-ctx.Done():
			log.Warn().Msg("interrupted in StartToGoal")
			return errors.Wrap(code.InterruptError, "StartToGoal interrupted")
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
			return err
		}

		// Check if task is finished BEFORE executing actions
		if dExt.isTaskFinished(result) {
			log.Info().Msg("task finished, stopping StartToGoal")
			return nil
		}

		// Execute actions only if task is not finished
		if err := dExt.executeActions(ctx, result.ToolCalls); err != nil {
			return err
		}

		if options.MaxRetryTimes > 1 && attempt >= options.MaxRetryTimes {
			return errors.New("reached max retry times")
		}
	}
}

func (dExt *XTDriver) AIAction(ctx context.Context, prompt string, opts ...option.ActionOption) error {
	log.Info().Str("prompt", prompt).Msg("performing AI action")

	// plan next action
	result, err := dExt.PlanNextAction(ctx, prompt, opts...)
	if err != nil {
		return err
	}

	// execute actions
	return dExt.executeActions(ctx, result.ToolCalls)
}

func (dExt *XTDriver) PlanNextAction(ctx context.Context, prompt string, opts ...option.ActionOption) (*ai.PlanningResult, error) {
	if dExt.LLMService == nil {
		return nil, errors.New("LLM service is not initialized")
	}

	compressedBufSource, err := getScreenShotBuffer(dExt.IDriver)
	if err != nil {
		return nil, err
	}

	// convert buffer to base64 string
	screenShotBase64 := "data:image/jpeg;base64," +
		base64.StdEncoding.EncodeToString(compressedBufSource.Bytes())

	// save screenshot to file
	imagePath := filepath.Join(
		config.GetConfig().ScreenShotsPath,
		fmt.Sprintf("%s.jpeg", builtin.GenNameWithTimestamp("%d_screenshot")),
	)
	go func() {
		err := saveScreenShot(compressedBufSource, imagePath)
		if err != nil {
			log.Error().Err(err).Msg("save screenshot file failed")
		}
	}()

	size, err := dExt.IDriver.WindowSize()
	if err != nil {
		return nil, errors.Wrap(code.DeviceGetInfoError, err.Error())
	}

	// Parse action options to get ResetHistory setting
	options := option.NewActionOptions(opts...)
	resetHistory := options.ResetHistory

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

// executeActions executes the planned actions
func (dExt *XTDriver) executeActions(ctx context.Context, toolCalls []schema.ToolCall) error {
	for _, action := range toolCalls {
		// Check for context cancellation before each action
		select {
		case <-ctx.Done():
			log.Warn().Msg("interrupted in executeActions")
			return errors.Wrap(code.InterruptError, "executeActions interrupted")
		default:
		}

		// call eino tool
		arguments := make(map[string]interface{})
		err := json.Unmarshal([]byte(action.Function.Arguments), &arguments)
		if err != nil {
			return err
		}
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
		if err != nil {
			return err
		}
	}

	return nil
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
