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
	}, nil
}

// Plan implements the ILLMService.Plan method using Wings API
func (w *WingsService) Plan(ctx context.Context, opts *PlanningOptions) (*PlanningResult, error) {
	// Validate input parameters
	if err := validatePlanningInput(opts); err != nil {
		return nil, errors.Wrap(err, "validate planning parameters failed")
	}

	// Extract screenshot from message
	screenshot, err := w.extractScreenshotFromMessage(opts.Message)
	if err != nil {
		return nil, errors.Wrap(err, "extract screenshot failed")
	}

	// Get device info from context (if available)
	deviceInfo := w.getDeviceInfoFromContext(ctx, screenshot)

	// Prepare Wings API request
	apiRequest := WingsActionRequest{
		Historys: []interface{}{}, // empty as specified
		DeviceInfos: []WingsDeviceInfo{
			deviceInfo,
		},
		StepText: opts.UserInstruction,
		BizId:    w.bizId,
		TextCase: "整体描述：\\n前置条件：\\n获取 1 台设备 A。\\n获取 1 个[万粉创作者]账号a。\\n获取 2 个[普通]账号 b、c。\\n账号 a 和账号 b 互相关注。\\n账号 a 和账号 c 互相关注。\\n账号 a 给账号 b 设置备注为 “11131b”。\\n账号 a 给账号 c 设置备注为 “11131c”。\\n账号 a 创建一个粉丝群 m。\\n 账号 a 修改粉丝群 m 名称为“11131群”。\\n 账号 a 邀请账号 b 加入粉丝群 m。\\n账号 a 邀请账号 c 加入粉丝群 m。\\n账号 a 给群聊 m 发送一条文字消息。\\n设备 A 打开抖音 app。\\n设备 A 登录账号 a。\\n设备 A 退出抖音 app。\\n操作步骤：\\n账号a打开抖音app。\\n点击“消息”。\\n点击“11131群”cell。\\n点击“聊天信息页入口”按钮。\\n点击“分享公开群”按钮。\\n点击文字“群口令”。\\n断言：屏幕中存在文字“口令复制成功”。\\n停止操作。\\n注意事项：\\n",
		StepType: "automation",
		DeviceID: deviceInfo.DeviceID,
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
	if response.BaseResp.StatusCode != 0 {
		err = fmt.Errorf("API returned error: %s", response.BaseResp.StatusMessage)
		return &PlanningResult{
			Thought:   response.ThoughtChain.Thought,
			Error:     err.Error(),
			ModelName: "wings-api",
		}, err
	}

	// Convert Wings API response to tool calls
	toolCalls, err := w.convertWingsResponseToToolCalls(response.ActionParams)
	if err != nil {
		return &PlanningResult{
			Thought:   response.ThoughtChain.Thought,
			Error:     err.Error(),
			ModelName: "wings-api",
		}, errors.Wrap(err, "convert Wings response to tool calls failed")
	}

	log.Info().
		Str("thought", response.ThoughtChain.Thought).
		Int("tool_calls_count", len(toolCalls)).
		Int64("elapsed_ms", elapsed).
		Msg("Wings API planning completed")

	return &PlanningResult{
		ToolCalls: toolCalls,
		Thought:   response.ThoughtChain.Thought,
		Content:   response.ThoughtChain.Summary,
		ModelName: "wings-api",
	}, nil
}

// Assert implements the ILLMService.Assert method using Wings API
func (w *WingsService) Assert(ctx context.Context, opts *AssertOptions) (*AssertionResult, error) {
	// Validate input parameters
	if err := validateAssertionInput(opts); err != nil {
		return nil, errors.Wrap(err, "validate assertion parameters failed")
	}

	// Clean screenshot data URL prefix
	cleanScreenshot := w.cleanScreenshotDataURL(opts.Screenshot)

	// Get device info from context (if available)
	deviceInfo := w.getDeviceInfoFromScreenshot(ctx, cleanScreenshot)

	// Prepare Wings API request for assertion
	apiRequest := WingsActionRequest{
		Historys: []interface{}{}, // empty as specified
		DeviceInfos: []WingsDeviceInfo{
			deviceInfo,
		},
		StepText: opts.Assertion,
		BizId:    w.bizId,
		TextCase: "整体描述：\\n前置条件：\\n获取 1 台设备 A。\\n获取 1 个[万粉创作者]账号a。\\n获取 2 个[普通]账号 b、c。\\n账号 a 和账号 b 互相关注。\\n账号 a 和账号 c 互相关注。\\n账号 a 给账号 b 设置备注为 “11131b”。\\n账号 a 给账号 c 设置备注为 “11131c”。\\n账号 a 创建一个粉丝群 m。\\n 账号 a 修改粉丝群 m 名称为“11131群”。\\n 账号 a 邀请账号 b 加入粉丝群 m。\\n账号 a 邀请账号 c 加入粉丝群 m。\\n账号 a 给群聊 m 发送一条文字消息。\\n设备 A 打开抖音 app。\\n设备 A 登录账号 a。\\n设备 A 退出抖音 app。\\n操作步骤：\\n账号a打开抖音app。\\n点击“消息”。\\n点击“11131群”cell。\\n点击“聊天信息页入口”按钮。\\n点击“分享公开群”按钮。\\n点击文字“群口令”。\\n断言：屏幕中存在文字“口令复制成功”。\\n停止操作。\\n注意事项：\\n",
		StepType: "assert", // Different from automation
		DeviceID: deviceInfo.DeviceID,
		Base: WingsBase{
			LogID: generateWingsUUID(),
		},
	}
	log.Info().Interface("apiRequest", apiRequest).Msg("Wings API request")

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
	if response.BaseResp.StatusCode != 0 {
		err = fmt.Errorf("API returned error: %s", response.BaseResp.StatusMessage)
		return &AssertionResult{
			Pass:      false,
			Thought:   response.ThoughtChain.Thought,
			ModelName: "wings-api",
		}, err
	}

	// Parse assertion result from action_params
	passed, assertionThought, err := w.parseAssertionResult(response.ActionParams, response.ThoughtChain)
	if err != nil {
		return &AssertionResult{
			Pass:      false,
			Thought:   response.ThoughtChain.Thought,
			ModelName: "wings-api",
		}, errors.Wrap(err, "parse assertion result failed")
	}

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
	Historys    []interface{}     `json:"historys"`
	DeviceInfos []WingsDeviceInfo `json:"device_infos"`
	StepText    string            `json:"step_text"`
	BizId       string            `json:"biz_id"`
	TextCase    string            `json:"text_case"`
	StepType    string            `json:"step_type"`
	DeviceID    string            `json:"device_id"`
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
	StepType     string            `json:"step_type"`
	ActionParams string            `json:"action_params"`
	ThoughtChain WingsThoughtChain `json:"thought_chain"`
	BaseResp     WingsBaseResp     `json:"BaseResp"`
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
			// Extract base64 data from data URL
			screenshot := content.ImageURL.URL
			if strings.HasPrefix(screenshot, "data:image/") {
				// Remove data URL prefix
				parts := strings.Split(screenshot, ",")
				if len(parts) == 2 {
					return parts[1], nil
				}
			}
			return screenshot, nil
		}
	}

	return "", errors.New("no image found in message")
}

// getDeviceInfoFromContext gets device info from context with fallback
func (w *WingsService) getDeviceInfoFromContext(_ context.Context, screenshot string) WingsDeviceInfo {
	// use default device info
	return WingsDeviceInfo{
		DeviceID:        "default-device",
		NowImage:        screenshot,
		PreImage:        screenshot,
		NowLayoutJSON:   "",
		OperationSystem: "android",
	}
}

// getDeviceInfoFromScreenshot gets device info from screenshot (for Assert)
func (w *WingsService) getDeviceInfoFromScreenshot(ctx context.Context, screenshot string) WingsDeviceInfo {
	return w.getDeviceInfoFromContext(ctx, screenshot)
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
		Timeout: 60 * time.Second,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, errors.Wrap(err, "HTTP request failed")
	}
	defer resp.Body.Close()

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
	if actionParamsStr == "" {
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
