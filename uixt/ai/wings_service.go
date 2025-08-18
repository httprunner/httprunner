package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/builtin"
)

// WingsService implements ILLMService interface using external Wings API
type WingsService struct {
	apiURL    string
	bizId     string
	accessKey string
	secretKey string
	history   []History // Conversation history for Wings API
}

// NewWingsService creates a new Wings service instance
func NewWingsService() (ILLMService, error) {
	// Check for environment variables for external API access
	apiURL := os.Getenv("VEDEM_WINGS_API_URL")
	accessKey := os.Getenv("VEDEM_WINGS_AK")
	secretKey := os.Getenv("VEDEM_WINGS_SK")
	bizID := os.Getenv("VEDEM_WINGS_BIZ_ID")

	// check required env
	if apiURL == "" {
		return nil, errors.Wrap(code.LLMEnvMissedError, "missed env VEDEM_WINGS_API_URL")
	}
	if bizID == "" {
		return nil, errors.Wrap(code.LLMEnvMissedError, "missed env VEDEM_WINGS_BIZ_ID")
	}

	return &WingsService{
		apiURL:    apiURL,
		bizId:     bizID,
		accessKey: accessKey,
		secretKey: secretKey,
		history:   []History{},
	}, nil
}

// Plan implements the ILLMService.Plan method using Wings API
func (w *WingsService) Plan(ctx context.Context, opts *PlanningOptions) (*PlanningResult, error) {
	// Validate input parameters
	if err := validatePlanningInput(opts); err != nil {
		return nil, errors.Wrap(err, "validate planning parameters failed")
	}

	// Reset history if requested
	if opts.ResetHistory {
		w.resetHistory()
	}

	// Extract screenshot from message
	screenshot, err := w.extractScreenshotFromMessage(opts.Message)
	if err != nil {
		return nil, errors.Wrap(err, "extract screenshot failed")
	}

	// Get device info from context (if available)
	deviceInfo := w.getDeviceInfoFromScreenshot(ctx, screenshot)

	// Prepare Wings API request
	apiRequest := WingsActionRequest{
		Historys:    w.history,
		DeviceInfos: deviceInfo,
		StepText:    fmt.Sprintf("%s", opts.UserInstruction),
		BizId:       w.bizId,
		TextCase:    fmt.Sprintf("整体描述：\n前置条件：\n操作步骤：\n%s\n停止操作。\n注意事项：\n", opts.UserInstruction),
		Base: WingsBase{
			LogID: generateWingsUUID(),
		},
	}

	// Call Wings API
	startTime := time.Now()
	response, err := w.callWingsAPI(ctx, apiRequest)
	elapsed := time.Since(startTime).Milliseconds()

	if err != nil {
		return &PlanningResult{
			Thought:   "Wings API call failed",
			Error:     err.Error(),
			ModelName: "wings-api",
		}, errors.Wrap(err, "Wings API call failed")
	}

	// Check API response status
	if response.BaseResp.StatusCode != 0 && response.BaseResp.StatusCode != 200 {
		err = fmt.Errorf("API returned error: %s", response.BaseResp.StatusMessage)
		return &PlanningResult{
			Thought:   response.ThoughtChain.Thought,
			Error:     err.Error(),
			ModelName: "wings-api",
		}, err
	}

	// Update history with response data
	newHistoryEntry := History{
		ThoughtChain:  response.ThoughtChain,
		StepText:      response.StepText,
		StepTextTrans: response.StepTextTrans,
		OriStepIndex:  response.OriStepIndex,
		DeviceID:      deviceInfo[0].DeviceID,
		AgentType:     response.AgentType,
		ActionResult:  "", // Always empty as requested
		DeviceInfos:   &deviceInfo,
		ActionParams:  response.ActionParams,
	}
	w.history = append(w.history, newHistoryEntry)
	var toolCalls []schema.ToolCall
	if response.StepType != "FINISH" {
		// Convert Wings API response to tool calls
		toolCalls, err = w.convertWingsResponseToToolCalls(response.ActionParams)
		if err != nil {
			return &PlanningResult{
				Thought:   response.ThoughtChain.Thought,
				Error:     err.Error(),
				ModelName: "wings-api",
			}, errors.Wrap(err, "convert Wings response to tool calls failed")
		}
	}

	// No need to update ActionResult as per user request
	// ActionResult should always be empty

	log.Info().
		Str("thought", response.ThoughtChain.Thought).
		Str("action", response.AgentType).
		Str("action_params", response.ActionParams).
		Str("log_id", fmt.Sprintf("%v", response.BaseResp.Extra)).
		Int("tool_calls_count", len(toolCalls)).
		Int64("elapsed_ms", elapsed).
		Msg("Wings API planning completed")

	return &PlanningResult{
		ToolCalls: toolCalls,
		Thought:   response.StepTextTrans,
		Content:   response.StepTextTrans,
		ModelName: "wings-api",
	}, nil
}

// Assert implements the ILLMService.Assert method using Wings API
func (w *WingsService) Assert(ctx context.Context, opts *AssertOptions) (*AssertionResult, error) {
	// Validate input parameters
	if err := validateAssertionInput(opts); err != nil {
		return nil, errors.Wrap(err, "validate assertion parameters failed")
	}

	// Get device info from context (if available)
	deviceInfos := w.getDeviceInfoFromScreenshot(ctx, opts.Screenshot)

	// Prepare Wings API request for assertion
	apiRequest := WingsActionRequest{
		Historys:    []History{},
		DeviceInfos: deviceInfos,
		StepText:    fmt.Sprintf("断言:%s", opts.Assertion),
		BizId:       w.bizId,
		TextCase:    fmt.Sprintf("整体描述：\n前置条件：\n操作步骤：\n断言: %s\n停止操作。\n注意事项：\n", opts.Assertion),
		Base: WingsBase{
			LogID: generateWingsUUID(),
		},
	}

	// Call Wings API
	startTime := time.Now()
	response, err := w.callWingsAPI(ctx, apiRequest)
	elapsed := time.Since(startTime).Milliseconds()

	if err != nil {
		return &AssertionResult{
			Pass:      false,
			Thought:   "Wings API call failed",
			ModelName: "wings-api",
		}, errors.Wrap(err, "Wings API call failed")
	}

	// Check API response status
	if response.BaseResp.StatusCode != 0 && response.BaseResp.StatusCode != 200 {
		err = fmt.Errorf("API returned error: %s", response.BaseResp.StatusMessage)
		return &AssertionResult{
			Pass:      false,
			Thought:   response.ThoughtChain.Thought,
			ModelName: "wings-api",
		}, err
	}

	// Update history with response data
	newHistoryEntry := History{
		ThoughtChain:  response.ThoughtChain,
		StepText:      response.StepText,
		StepTextTrans: response.StepTextTrans,
		OriStepIndex:  response.OriStepIndex,
		DeviceID:      response.DeviceId,
		AgentType:     response.AgentType,
		DeviceInfos:   &apiRequest.DeviceInfos,
		ActionParams:  response.ActionParams,
	}
	w.history = append(w.history, newHistoryEntry)

	// Parse assertion result from action_params
	passed, assertionThought, err := w.parseAssertionResult(response.ActionParams, response.ThoughtChain)
	if err != nil {
		return &AssertionResult{
			Pass:      false,
			Thought:   response.ThoughtChain.Thought,
			ModelName: "wings-api",
		}, errors.Wrap(err, "parse assertion result failed")
	}

	// No need to update ActionResult as per user request
	// ActionResult should always be empty

	log.Info().
		Bool("passed", passed).
		Str("thought", assertionThought).
		Int64("elapsed_ms", elapsed).
		Msg("Wings API assertion completed")

	result := &AssertionResult{
		Pass:      passed,
		Thought:   assertionThought,
		ModelName: "wings-api",
	}

	// Return error if assertion failed (consistent with original behavior)
	if !passed {
		return result, errors.New(assertionThought)
	}

	return result, nil
}

// Query implements the ILLMService.Query method (not supported)
func (w *WingsService) Query(ctx context.Context, opts *QueryOptions) (*QueryResult, error) {
	return nil, errors.New("Query operation is not supported by Wings service")
}

// RegisterTools implements the ILLMService.RegisterTools method (no-op for Wings)
func (w *WingsService) RegisterTools(tools []*schema.ToolInfo) error {
	// Wings service doesn't need tool registration as it determines actions via API
	log.Debug().Int("tools_count", len(tools)).Msg("Wings service ignoring tool registration")
	return nil
}

// Wings API data structures
type WingsActionRequest struct {
	Historys    []History         `json:"historys"`
	DeviceInfos []WingsDeviceInfo `json:"device_infos"`
	StepText    string            `json:"step_text"`
	TextCase    string            `json:"text_case"`
	BizId       string            `json:"biz_id"`
	TaskType    string            `json:"task_type"`
	Base        WingsBase         `json:"Base"`
}

type WingsDeviceInfo struct {
	DeviceID        string `json:"device_id"`
	NowImage        string `json:"now_image"`
	PreImage        string `json:"pre_image"`
	NowImageUrl     string `json:"now_image_url"`
	PreImageUrl     string `json:"pre_image_url"`
	NowLayoutJSON   string `json:"now_layout_json"`
	OperationSystem string `json:"operation_system"`
}

type WingsBase struct {
	LogID string `json:"LogID"`
}

type WingsActionResponse struct {
	AgentType     string            `json:"agent_type"`
	StepText      string            `json:"step_text"`
	StepTextTrans string            `json:"step_text_trans"`
	OriStepIndex  int               `json:"ori_step_index"`
	StepType      string            `json:"step_type"`
	ActionParams  string            `json:"action_params"`
	DeviceId      string            `json:"device_id"`
	NextIsFinish  bool              `json:"next_is_finish"`
	ThoughtChain  WingsThoughtChain `json:"thought_chain"`
	BaseResp      WingsBaseResp     `json:"BaseResp"`
}

type WingsThoughtChain struct {
	Observation string `json:"observation"`
	Thought     string `json:"thought"`
	Summary     string `json:"summary"`
}

type WingsBaseResp struct {
	StatusCode    int        `json:"StatusCode"`
	StatusMessage string     `json:"StatusMessage"`
	Extra         WingsExtra `json:"Extra"`
}

type WingsExtra struct {
	CostTime string `json:"cost_time"`
	LogID    string `json:"_log_id"`
}

// History structure for request and response
type History struct {
	ThoughtChain  WingsThoughtChain  `json:"thought_chain"`   // 思考结果
	StepText      string             `json:"step_text"`       // 操作的指令
	DeviceID      string             `json:"device_id"`       // 操作的设备id
	AgentType     string             `json:"agent_type"`      // 最终决策的agent类型
	ActionResult  string             `json:"action_result"`   // 操作结果, 断言=断言结果, 自动化=自动化操作是否成功, 物料构造=物料构造结果
	DeviceInfos   *[]WingsDeviceInfo `json:"device_infos"`    // 所有设备的信息
	ActionParams  string             `json:"action_params"`   // 历史操作解析结果(断言，自动化，物料构造)
	StepTextTrans string             `json:"step_text_trans"` // 归一化的步骤文本(为后续的实际执行解析文本)
	OriStepIndex  int                `json:"ori_step_index"`  // 原本的执行序列（扩展前、目标导向原始文本步骤）
}

// Action parameter structures
type WingsActionParams struct {
	Type    string      `json:"Type"`
	Params  interface{} `json:"Params"`
	Bounds  [][]float64 `json:"Bounds"`
	UiDict  interface{} `json:"UiDict"`
	UiIndex string      `json:"UiIndex"`
}

type WingsTapParams struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type WingsDoubleTapParams struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type WingsLongPressParams struct {
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Duration float64 `json:"duration"`
}

type WingsSwipeParams struct {
	FromX    float64 `json:"from_x"`
	FromY    float64 `json:"from_y"`
	ToX      float64 `json:"to_x"`
	ToY      float64 `json:"to_y"`
	Duration float64 `json:"duration"`
}

type WingsTextParams struct {
	Text string `json:"text"`
}

// Helper methods

// resetHistory resets the conversation history
func (w *WingsService) resetHistory() {
	w.history = []History{}
}

// generateWingsUUID generates a random UUID for LogID
func generateWingsUUID() string {
	return uuid.New().String()
}

// extractScreenshotFromMessage extracts base64 screenshot from message
func (w *WingsService) extractScreenshotFromMessage(message *schema.Message) (string, error) {
	if message == nil || len(message.MultiContent) == 0 {
		return "", errors.New("no message content found")
	}

	for _, content := range message.MultiContent {
		if content.Type == schema.ChatMessagePartTypeImageURL && content.ImageURL != nil {
			return content.ImageURL.URL, nil
		}
	}

	return "", errors.New("no image found in message")
}

// getDeviceInfoFromBase gets device info from base64 screenshot
func (w *WingsService) getDeviceInfoFromBase64(screenshotBase64 string) []WingsDeviceInfo {
	// TODO: Extract device info from context if available

	// Use last history's NowImage as PreImage if history exists
	preImage := screenshotBase64
	if len(w.history) > 0 && w.history[len(w.history)-1].DeviceInfos != nil && len(*w.history[len(w.history)-1].DeviceInfos) > 0 {
		preImage = (*w.history[len(w.history)-1].DeviceInfos)[0].NowImage
	}

	preImageUrl := ""
	if len(w.history) > 0 && w.history[len(w.history)-1].DeviceInfos != nil && len(*w.history[len(w.history)-1].DeviceInfos) > 0 {
		preImageUrl = (*w.history[len(w.history)-1].DeviceInfos)[0].NowImageUrl
	}

	// use default device info with optimized PreImage
	return []WingsDeviceInfo{
		{
			DeviceID:        "default-device",
			NowImage:        screenshotBase64,
			NowImageUrl:     "",
			PreImage:        preImage,
			PreImageUrl:     preImageUrl,
			NowLayoutJSON:   "",
			OperationSystem: "android",
		},
	}
}

// getDeviceInfoFromUrl gets device info from url screenshot
func (w *WingsService) getDeviceInfoFromUrl(screenshotUrl string) []WingsDeviceInfo {
	// TODO: Extract device info from context if available

	// Use last history's NowImage as PreImage if history exists
	preImage := ""
	if len(w.history) > 0 && w.history[len(w.history)-1].DeviceInfos != nil && len(*w.history[len(w.history)-1].DeviceInfos) > 0 {
		preImage = (*w.history[len(w.history)-1].DeviceInfos)[0].NowImage
	}

	preImageUrl := screenshotUrl
	if len(w.history) > 0 && w.history[len(w.history)-1].DeviceInfos != nil && len(*w.history[len(w.history)-1].DeviceInfos) > 0 {
		preImageUrl = (*w.history[len(w.history)-1].DeviceInfos)[0].NowImageUrl
	}

	// use default device info with optimized PreImage
	return []WingsDeviceInfo{
		{
			DeviceID:        "default-device",
			NowImage:        "",
			NowImageUrl:     screenshotUrl,
			PreImage:        preImage,
			PreImageUrl:     preImageUrl,
			NowLayoutJSON:   "",
			OperationSystem: "android",
		},
	}
}

// getDeviceInfoFromScreenshot gets device info from screenshot (for Assert)
func (w *WingsService) getDeviceInfoFromScreenshot(ctx context.Context, screenshot string) []WingsDeviceInfo {
	if strings.HasPrefix(screenshot, "data:image/") {
		// Remove data URL prefix like "data:image/jpeg;base64,"
		parts := strings.Split(screenshot, ",")
		if len(parts) == 2 {
			return w.getDeviceInfoFromBase64(parts[1])
		}
	}
	return w.getDeviceInfoFromUrl(screenshot)
}

// cleanScreenshotDataURL removes data URL prefix from screenshot string
func (w *WingsService) cleanScreenshotDataURL(screenshot string) string {
	if strings.HasPrefix(screenshot, "data:image/") {
		// Remove data URL prefix like "data:image/jpeg;base64,"
		parts := strings.Split(screenshot, ",")
		if len(parts) == 2 {
			return parts[1]
		}
	}
	return screenshot
}

// callWingsAPI calls the external Wings API
func (w *WingsService) callWingsAPI(ctx context.Context, request WingsActionRequest) (*WingsActionResponse, error) {
	// Marshal request to JSON
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, errors.Wrap(err, "marshal request failed")
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", w.apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, errors.Wrap(err, "create HTTP request failed")
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Add("x-use-ppe", "1")
	httpReq.Header.Add("x-tt-env", "ppe_refactor_merge")

	// Add authentication headers if using external API
	if w.accessKey != "" && w.secretKey != "" {
		signToken := "UNSIGNED-PAYLOAD"
		token := builtin.Sign("auth-v2", w.accessKey, w.secretKey, []byte(signToken))

		httpReq.Header.Add("Agw-Auth", token)
		httpReq.Header.Add("Agw-Auth-Content", signToken)
		httpReq.Header.Add("Content-Type", "application/json")
	}

	// Execute HTTP request
	client := &http.Client{
		Timeout: 120 * time.Second,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, errors.Wrap(err, "HTTP request failed")
	}
	defer resp.Body.Close()

	logID := resp.Header.Get("X-Tt-Logid")
	log.Info().Str("step_text", request.StepText).
		Str("image_url", request.DeviceInfos[0].NowImageUrl).
		Str("log_id", logID).Str("biz_id", request.BizId).
		Str("url", w.apiURL).Msg("call wings api")

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "read response body failed")
	}

	// Check HTTP status
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	// Parse response
	var apiResponse WingsActionResponse
	if err := json.Unmarshal(responseBody, &apiResponse); err != nil {
		return nil, errors.Wrap(err, "unmarshal response failed")
	}

	return &apiResponse, nil
}

// convertWingsResponseToToolCalls converts Wings API response to tool calls using generic approach
func (w *WingsService) convertWingsResponseToToolCalls(actionParamsStr string) ([]schema.ToolCall, error) {
	if actionParamsStr == "" || actionParamsStr == "FINISH" {
		return []schema.ToolCall{}, nil
	}

	var actionParams WingsActionParams
	if err := json.Unmarshal([]byte(actionParamsStr), &actionParams); err != nil {
		return nil, fmt.Errorf("parse action params failed: %w", err)
	}

	// Use Wings API Type as tool name directly
	toolName := actionParams.Type
	params := actionParams.Params

	// Create tool call using generic method
	toolCall, err := w.createToolCall(toolName, params)
	if err != nil {
		return nil, fmt.Errorf("create tool call for %s failed: %w", toolName, err)
	}

	return []schema.ToolCall{toolCall}, nil
}

// createToolCall creates a generic tool call with given name and arguments
func (w *WingsService) createToolCall(toolName string, params interface{}) (schema.ToolCall, error) {
	// Convert params to arguments map
	arguments := make(map[string]interface{})

	if params != nil {
		// Try to convert params to map[string]interface{}
		switch p := params.(type) {
		case map[string]interface{}:
			arguments = p
		case string:
			// If params is a string, try to unmarshal it as JSON
			if err := json.Unmarshal([]byte(p), &arguments); err != nil {
				// If not JSON, treat as simple text parameter
				arguments["text"] = p
			}
		default:
			// For other types, try to marshal and unmarshal
			paramsBytes, err := json.Marshal(params)
			if err != nil {
				return schema.ToolCall{}, fmt.Errorf("marshal params failed: %w", err)
			}
			if err := json.Unmarshal(paramsBytes, &arguments); err != nil {
				// If unmarshal fails, create a generic params field
				arguments["params"] = params
			}
		}
	}

	// Convert arguments to JSON string
	argumentsJSON, err := json.Marshal(arguments)
	if err != nil {
		return schema.ToolCall{}, fmt.Errorf("marshal arguments failed: %w", err)
	}

	// Generate unique tool call ID
	toolCallID := fmt.Sprintf("call_%s", uuid.New().String()[:8])

	return schema.ToolCall{
		ID: toolCallID,
		Function: schema.FunctionCall{
			Name:      toolName,
			Arguments: string(argumentsJSON),
		},
	}, nil
}

// parseAssertionResult parses the assertion result from action_params
func (w *WingsService) parseAssertionResult(actionParamsStr string, thoughtChain WingsThoughtChain) (bool, string, error) {
	// Parse action parameters JSON
	var actionParams map[string]interface{}
	if err := json.Unmarshal([]byte(actionParamsStr), &actionParams); err != nil {
		return false, "", errors.Wrap(err, "parse action params failed")
	}

	// Extract action_type from the parsed JSON
	actionType, exists := actionParams["action_type"]
	if !exists {
		// If no action_type field, try to parse nested structure
		if totalRes, ok := actionParams["total_res"].([]interface{}); ok && len(totalRes) > 0 {
			if firstRes, ok := totalRes[0].(map[string]interface{}); ok {
				if actionParamsNested, ok := firstRes["action_params"].(map[string]interface{}); ok {
					if nestedActionType, ok := actionParamsNested["action_type"]; ok {
						actionType = nestedActionType
					}
				}
			}
		}
	}

	// Default to failed if no action_type found
	if actionType == nil {
		return false, thoughtChain.Summary, nil
	}

	// Convert action_type to string and check result
	actionTypeStr, ok := actionType.(string)
	if !ok {
		return false, thoughtChain.Summary, nil
	}

	// Determine assertion result based on action_type
	passed := strings.ToLower(actionTypeStr) == "passed"

	// Use thoughtChain.Summary as the assertion thought
	assertionThought := thoughtChain.Summary
	if assertionThought == "" {
		assertionThought = thoughtChain.Thought
	}
	if assertionThought == "" {
		assertionThought = thoughtChain.Observation
	}

	log.Info().
		Str("action_type", actionTypeStr).
		Bool("passed", passed).
		Str("thought", assertionThought).
		Msg("parsed Wings assertion result")

	return passed, assertionThought, nil
}
