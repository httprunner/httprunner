package option

import (
	"context"
	"encoding/json"
	"math/rand/v2"
	"reflect"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/uixt/types"
)

type MobileAction struct {
	Method  ActionName     `json:"method,omitempty" yaml:"method,omitempty"`
	Params  interface{}    `json:"params,omitempty" yaml:"params,omitempty"`
	Fn      func()         `json:"-" yaml:"-"` // used for function action, not serialized
	Options *ActionOptions `json:"options,omitempty" yaml:"options,omitempty"`
	ActionOptions
}

func (ma MobileAction) GetOptions() []ActionOption {
	var actionOptionList []ActionOption
	// Notice: merge options from ma.Options and ma.ActionOptions
	if ma.Options != nil {
		actionOptionList = append(actionOptionList, ma.Options.Options()...)
	}
	actionOptionList = append(actionOptionList, ma.ActionOptions.Options()...)
	return actionOptionList
}

type ActionName string

const (
	ACTION_LOG              ActionName = "log"
	ACTION_ListPackages     ActionName = "list_packages"
	ACTION_AppInstall       ActionName = "app_install"
	ACTION_AppUninstall     ActionName = "app_uninstall"
	ACTION_WebLoginNoneUI   ActionName = "web_login_none_ui"
	ACTION_AppClear         ActionName = "app_clear"
	ACTION_AppStart         ActionName = "app_start"
	ACTION_AppLaunch        ActionName = "app_launch" // 启动 app 并堵塞等待 app 首屏加载完成
	ACTION_AppTerminate     ActionName = "app_terminate"
	ACTION_ColdLaunch       ActionName = "cold_launch"
	ACTION_AppStop          ActionName = "app_stop"
	ACTION_ScreenShot       ActionName = "screenshot"
	ACTION_ScreenRecord     ActionName = "screenrecord"
	ACTION_GetScreenSize    ActionName = "get_screen_size"
	ACTION_Sleep            ActionName = "sleep"
	ACTION_SleepMS          ActionName = "sleep_ms"
	ACTION_SleepRandom      ActionName = "sleep_random"
	ACTION_SetIme           ActionName = "set_ime"
	ACTION_GetSource        ActionName = "get_source"
	ACTION_GetForegroundApp ActionName = "get_foreground_app"
	ACTION_GetPasteboard    ActionName = "get_pasteboard" // get clipboard content
	ACTION_AppInfo          ActionName = "app_info"       // get app info action

	// UI handling
	ACTION_Home                     ActionName = "home"
	ACTION_TapXY                    ActionName = "tap_xy"
	ACTION_TapAbsXY                 ActionName = "tap_abs_xy"
	ACTION_TapByOCR                 ActionName = "tap_ocr"
	ACTION_TapByCV                  ActionName = "tap_cv"
	ACTION_DoubleTap                ActionName = "double_tap" // generic double tap action
	ACTION_DoubleTapXY              ActionName = "double_tap_xy"
	ACTION_Swipe                    ActionName = "swipe"                    // swipe by direction or coordinates
	ACTION_SwipeDirection           ActionName = "swipe_direction"          // swipe by direction (up, down, left, right)
	ACTION_SwipeCoordinate          ActionName = "swipe_coordinate"         // swipe by coordinates (fromX, fromY, toX, toY)
	ACTION_SIMSwipeDirection        ActionName = "sim_swipe_direction"      // simulated swipe by direction with random distance
	ACTION_SIMSwipeInArea           ActionName = "sim_swipe_in_area"        // simulated swipe in area with direction and distance
	ACTION_SIMSwipeFromPointToPoint ActionName = "sim_swipe_point_to_point" // simulated swipe from point to point
	ACTION_SIMClickAtPoint          ActionName = "sim_click_at_point"       // simulated click at point
	ACTION_SIMInput                 ActionName = "sim_input"                // simulated text input with segments
	ACTION_Drag                     ActionName = "drag"
	ACTION_Input                    ActionName = "input"
	ACTION_PressButton              ActionName = "press_button"
	ACTION_Back                     ActionName = "back"
	ACTION_KeyCode                  ActionName = "keycode"
	ACTION_Delete                   ActionName = "delete"    // delete action
	ACTION_Backspace                ActionName = "backspace" // backspace action
	ACTION_TapBySelector            ActionName = "tap_by_selector"
	ACTION_HoverBySelector          ActionName = "hover_by_selector"
	ACTION_Hover                    ActionName = "hover"       // generic hover action
	ACTION_RightClick               ActionName = "right_click" // right click action
	ACTION_WebCloseTab              ActionName = "web_close_tab"
	ACTION_SecondaryClick           ActionName = "secondary_click"
	ACTION_SecondaryClickBySelector ActionName = "secondary_click_by_selector"
	ACTION_GetElementTextBySelector ActionName = "get_element_text_by_selector"
	ACTION_Scroll                   ActionName = "scroll"         // scroll action
	ACTION_Upload                   ActionName = "upload"         // upload action
	ACTION_PushMedia                ActionName = "push_media"     // push media action
	ACTION_CreateBrowser            ActionName = "create_browser" // create browser action

	// device actions
	ACTION_ListAvailableDevices ActionName = "list_available_devices"
	ACTION_SelectDevice         ActionName = "select_device"

	// custom actions
	ACTION_SwipeToTapApp   ActionName = "swipe_to_tap_app"   // swipe left & right to find app and tap
	ACTION_SwipeToTapText  ActionName = "swipe_to_tap_text"  // swipe up & down to find text and tap
	ACTION_SwipeToTapTexts ActionName = "swipe_to_tap_texts" // swipe up & down to find text and tap
	ACTION_ClosePopups     ActionName = "close_popups"
	ACTION_EndToEndDelay   ActionName = "live_e2e"
	ACTION_InstallApp      ActionName = "install_app"
	ACTION_UninstallApp    ActionName = "uninstall_app"
	ACTION_DownloadApp     ActionName = "download_app"
	ACTION_CallFunction    ActionName = "call_function"

	// AI actions
	ACTION_StartToGoal ActionName = "start_to_goal" // start to goal action
	ACTION_AIAction    ActionName = "ai_action"     // action with ai
	ACTION_AIAssert    ActionName = "ai_assert"     // assert with ai
	ACTION_Query       ActionName = "ai_query"      // query with ai
	ACTION_Finished    ActionName = "finished"      // finished action

	// anti-risk actions
	ACTION_SetTouchInfo     ActionName = "set_touch_info"
	ACTION_SetTouchInfoList ActionName = "set_touch_info_list"
)

const (
	// UI validation
	// selectors
	SelectorName          string = "ui_name"
	SelectorLabel         string = "ui_label"
	SelectorOCR           string = "ui_ocr"
	SelectorImage         string = "ui_image"
	SelectorAI            string = "ui_ai" // ui query with ai
	SelectorForegroundApp string = "ui_foreground_app"
	SelectorSelector      string = "ui_selector"
	// assertions
	AssertionEqual     string = "equal"
	AssertionNotEqual  string = "not_equal"
	AssertionExists    string = "exists"
	AssertionNotExists string = "not_exists"
	AssertionAI        string = "ai_assert" // assert with ai
)

type ActionOptions struct {
	// Device targeting
	Platform string `json:"platform,omitempty" yaml:"platform,omitempty" binding:"omitempty" desc:"Device platform: android/ios/browser"`
	Serial   string `json:"serial,omitempty" yaml:"serial,omitempty" binding:"omitempty" desc:"Device serial/udid/browser id"`

	// Common action parameters
	X     float64 `json:"x,omitempty" yaml:"x,omitempty" binding:"omitempty,min=0" desc:"X coordinate (0.0~1.0 for percent, or absolute pixel value)"`
	Y     float64 `json:"y,omitempty" yaml:"y,omitempty" binding:"omitempty,min=0" desc:"Y coordinate (0.0~1.0 for percent, or absolute pixel value)"`
	FromX float64 `json:"from_x,omitempty" yaml:"from_x,omitempty" binding:"omitempty,min=0" desc:"Starting X coordinate"`
	FromY float64 `json:"from_y,omitempty" yaml:"from_y,omitempty" binding:"omitempty,min=0" desc:"Starting Y coordinate"`
	ToX   float64 `json:"to_x,omitempty" yaml:"to_x,omitempty" binding:"omitempty,min=0" desc:"Ending X coordinate"`
	ToY   float64 `json:"to_y,omitempty" yaml:"to_y,omitempty" binding:"omitempty,min=0" desc:"Ending Y coordinate"`
	Text  string  `json:"text,omitempty" yaml:"text,omitempty" desc:"Text content for input/search operations"`

	// App/Package related
	PackageName        string `json:"packageName,omitempty" yaml:"packageName,omitempty" desc:"Package name of the app"`
	AppName            string `json:"appName,omitempty" yaml:"appName,omitempty" desc:"App name to find"`
	AppUrl             string `json:"appUrl,omitempty" yaml:"appUrl,omitempty" desc:"App URL for installation"`
	MappingUrl         string `json:"mappingUrl,omitempty" yaml:"mappingUrl,omitempty" desc:"Mapping URL for app installation"`
	ResourceMappingUrl string `json:"resourceMappingUrl,omitempty" yaml:"resourceMappingUrl,omitempty" desc:"Resource mapping URL for app installation"`

	// Web/Browser related
	Selector    string `json:"selector,omitempty" yaml:"selector,omitempty" desc:"CSS or XPath selector"`
	TabIndex    int    `json:"tabIndex,omitempty" yaml:"tabIndex,omitempty" desc:"Browser tab index"`
	PhoneNumber string `json:"phoneNumber,omitempty" yaml:"phoneNumber,omitempty" desc:"Phone number for login"`
	Captcha     string `json:"captcha,omitempty" yaml:"captcha,omitempty" desc:"Captcha code"`
	Password    string `json:"password,omitempty" yaml:"password,omitempty" desc:"Password for login"`

	// Button/Key related
	Button  types.DeviceButton `json:"button,omitempty" yaml:"button,omitempty" desc:"Device button to press"`
	Ime     string             `json:"ime,omitempty" yaml:"ime,omitempty" desc:"IME package name"`
	Count   int                `json:"count,omitempty" yaml:"count,omitempty" desc:"Count for delete operations"`
	Keycode int                `json:"keycode,omitempty" yaml:"keycode,omitempty" desc:"Keycode for key press operations"`

	// Image/CV related
	ImagePath string `json:"imagePath,omitempty" yaml:"imagePath,omitempty" desc:"Path to reference image for CV recognition"`

	// HTTP API specific fields
	FileUrl    string `json:"file_url,omitempty" yaml:"file_url,omitempty" desc:"File URL for upload operations"`
	FileFormat string `json:"file_format,omitempty" yaml:"file_format,omitempty" desc:"File format for upload operations"`
	ImageUrl   string `json:"imageUrl,omitempty" yaml:"imageUrl,omitempty" desc:"Image URL for media operations"`
	VideoUrl   string `json:"videoUrl,omitempty" yaml:"videoUrl,omitempty" desc:"Video URL for media operations"`
	Delta      int    `json:"delta,omitempty" yaml:"delta,omitempty" desc:"Delta value for scroll operations"`
	Width      int    `json:"width,omitempty" yaml:"width,omitempty" desc:"Width for browser creation"`
	Height     int    `json:"height,omitempty" yaml:"height,omitempty" desc:"Height for browser creation"`

	// Array parameters
	Texts  []string  `json:"texts,omitempty" yaml:"texts,omitempty" desc:"List of texts to search"`
	Params []float64 `json:"params,omitempty" yaml:"params,omitempty" desc:"Generic parameter array"`

	// AI related
	Prompt       string      `json:"prompt,omitempty" yaml:"prompt,omitempty" desc:"AI action prompt"`
	Content      string      `json:"content,omitempty" yaml:"content,omitempty" desc:"Content for finished action"`
	LLMService   string      `json:"llm_service,omitempty" yaml:"llm_service,omitempty" desc:"LLM service type for AI actions"`
	CVService    string      `json:"cv_service,omitempty" yaml:"cv_service,omitempty" desc:"Computer vision service type for AI actions"`
	ResetHistory bool        `json:"reset_history,omitempty" yaml:"reset_history,omitempty" desc:"Whether to reset conversation history before AI planning"`
	OutputSchema interface{} `json:"output_schema,omitempty" yaml:"output_schema,omitempty" desc:"Custom output schema for structured AI query response"`

	// Control options
	Context       context.Context `json:"-" yaml:"-"`
	Identifier    string          `json:"identifier,omitempty" yaml:"identifier,omitempty" desc:"Action identifier for logging"`
	MaxRetryTimes int             `json:"max_retry_times,omitempty" yaml:"max_retry_times,omitempty" desc:"Maximum retry times"`
	Interval      float64         `json:"interval,omitempty" yaml:"interval,omitempty" desc:"Interval between retries in seconds"`
	Duration      float64         `json:"duration,omitempty" yaml:"duration,omitempty" desc:"Action duration in seconds"`
	PressDuration float64         `json:"press_duration,omitempty" yaml:"press_duration,omitempty" desc:"Press duration in seconds"`
	Steps         int             `json:"steps,omitempty" yaml:"steps,omitempty" desc:"Number of steps for action"`
	Direction     interface{}     `json:"direction,omitempty" yaml:"direction,omitempty" desc:"Direction for swipe operations or custom coordinates"`

	// SIM specific options with SIM prefix
	SIMMinDistance float64 `json:"sim_min_distance,omitempty" yaml:"sim_min_distance,omitempty" desc:"Minimum distance for SIM simulated actions"`
	SIMMaxDistance float64 `json:"sim_max_distance,omitempty" yaml:"sim_max_distance,omitempty" desc:"Maximum distance for SIM simulated actions"`
	SIMAreaStartX  float64 `json:"sim_area_start_x,omitempty" yaml:"sim_area_start_x,omitempty" desc:"Area starting X coordinate for SIM simulated swipe"`
	SIMAreaStartY  float64 `json:"sim_area_start_y,omitempty" yaml:"sim_area_start_y,omitempty" desc:"Area starting Y coordinate for SIM simulated swipe"`
	SIMAreaEndX    float64 `json:"sim_area_end_x,omitempty" yaml:"sim_area_end_x,omitempty" desc:"Area ending X coordinate for SIM simulated swipe"`
	SIMAreaEndY    float64 `json:"sim_area_end_y,omitempty" yaml:"sim_area_end_y,omitempty" desc:"Area ending Y coordinate for SIM simulated swipe"`

	Timeout   int `json:"timeout,omitempty" yaml:"timeout,omitempty" desc:"Timeout in seconds for action execution"`
	TimeLimit int `json:"time_limit,omitempty" yaml:"time_limit,omitempty" desc:"Time limit in seconds for action execution, stops gracefully when reached"`
	Frequency int `json:"frequency,omitempty" yaml:"frequency,omitempty" desc:"Action frequency"`

	ScreenOptions

	// Anti-risk options
	AntiRisk bool `json:"anti_risk,omitempty" yaml:"anti_risk,omitempty" desc:"Enable anti-risk MCP tool calls"`

	// Custom options
	Custom map[string]interface{} `json:"custom,omitempty" yaml:"custom,omitempty" desc:"Custom options"`
}

func (o *ActionOptions) Options() []ActionOption {
	options := make([]ActionOption, 0)

	if o == nil {
		return options
	}

	if o.Context != nil {
		options = append(options, WithContext(o.Context))
	}
	if o.Identifier != "" {
		options = append(options, WithIdentifier(o.Identifier))
	}

	if o.MaxRetryTimes != 0 {
		options = append(options, WithMaxRetryTimes(o.MaxRetryTimes))
	}
	if o.IgnoreNotFoundError {
		options = append(options, WithIgnoreNotFoundError(true))
	}
	if o.Interval != 0 {
		options = append(options, WithInterval(o.Interval))
	}
	if o.Duration != 0 {
		options = append(options, WithDuration(o.Duration))
	}
	if o.PressDuration != 0 {
		options = append(options, WithPressDuration(o.PressDuration))
	}
	if o.Steps != 0 {
		options = append(options, WithSteps(o.Steps))
	}

	switch v := o.Direction.(type) {
	case string:
		options = append(options, WithDirection(v))
	case []float64:
		options = append(options, WithCustomDirection(
			v[0], v[1],
			v[2], v[3],
		))
	case []interface{}:
		// loaded from json case
		// custom direction: [fromX, fromY, toX, toY]
		sx, err := builtin.Interface2Float64(v[0])
		if err != nil {
			log.Error().Err(err).Interface("fromX", v[0]).Msg("convert float64 failed")
		}
		sy, err := builtin.Interface2Float64(v[1])
		if err != nil {
			log.Error().Err(err).Interface("fromY", v[1]).Msg("convert float64 failed")
		}
		ex, err := builtin.Interface2Float64(v[2])
		if err != nil {
			log.Error().Err(err).Interface("toX", v[2]).Msg("convert float64 failed")
		}
		ey, err := builtin.Interface2Float64(v[3])
		if err != nil {
			log.Error().Err(err).Interface("toY", v[3]).Msg("convert float64 failed")
		}
		options = append(options, WithCustomDirection(
			sx, sy,
			ex, ey,
		))
	}

	if o.Timeout != 0 {
		options = append(options, WithTimeout(o.Timeout))
	}
	if o.TimeLimit != 0 {
		options = append(options, WithTimeLimit(o.TimeLimit))
	}
	if o.Frequency != 0 {
		options = append(options, WithFrequency(o.Frequency))
	}
	if len(o.AbsScope) == 4 {
		options = append(options, WithAbsScope(
			o.AbsScope[0], o.AbsScope[1], o.AbsScope[2], o.AbsScope[3]))
	} else if len(o.Scope) == 4 {
		options = append(options, WithScope(
			o.Scope[0], o.Scope[1], o.Scope[2], o.Scope[3]))
	}
	if len(o.TapOffset) == 2 {
		// for tap [x,y] offset
		options = append(options, WithTapOffset(o.TapOffset[0], o.TapOffset[1]))
	}
	if o.TapRandomRect {
		// tap random point in OCR/CV rectangle
		options = append(options, WithTapRandomRect(true))
	}
	if len(o.SwipeOffset) == 4 {
		// for swipe [fromX, fromY, toX, toY] offset
		options = append(options, WithSwipeOffset(
			o.SwipeOffset[0], o.SwipeOffset[1], o.SwipeOffset[2], o.SwipeOffset[3]))
	}
	if len(o.OffsetRandomRange) == 2 {
		options = append(options, WithOffsetRandomRange(
			o.OffsetRandomRange[0], o.OffsetRandomRange[1]))
	}

	if o.Regex {
		options = append(options, WithRegex(true))
	}
	if o.Index != 0 {
		options = append(options, WithIndex(o.Index))
	}
	if o.MatchOne {
		options = append(options, WithMatchOne(true))
	}

	if o.AntiRisk {
		options = append(options, WithAntiRisk(true))
	}

	if o.PreMarkOperation {
		options = append(options, WithPreMarkOperation(true))
	}

	// custom options
	if o.Custom != nil {
		for k, v := range o.Custom {
			options = append(options, WithCustomOption(k, v))
		}
	}

	options = append(options, o.GetScreenShotOptions()...)
	options = append(options, o.GetScreenRecordOptions()...)
	options = append(options, o.GetMarkOperationOptions()...)

	return options
}

func (o *ActionOptions) ToMap() map[string]interface{} {
	result := make(map[string]interface{})
	b, err := json.Marshal(o)
	if err != nil {
		return nil
	}
	if err := json.Unmarshal(b, &result); err != nil {
		return nil
	}
	return result
}

func (o *ActionOptions) ApplyTapOffset(absX, absY float64) (float64, float64) {
	xOffset := 0.0
	yOffset := 0.0
	if len(o.TapOffset) == 2 {
		xOffset += float64(o.TapOffset[0])
		yOffset += float64(o.TapOffset[1])
	}
	xOffset += o.generateRandomOffset()
	yOffset += o.generateRandomOffset()

	if xOffset != 0.0 || yOffset != 0.0 {
		log.Debug().
			Float64("x_offset", xOffset).
			Float64("y_offset", yOffset).
			Msg("ApplyTapOffset")
	}
	return absX + xOffset, absY + yOffset
}

func (o *ActionOptions) ApplySwipeOffset(absFromX, absFromY, absToX, absToY float64) (
	float64, float64, float64, float64,
) {
	fromXOffset := 0.0
	fromYOffset := 0.0
	toXOffset := 0.0
	toYOffset := 0.0
	if len(o.SwipeOffset) == 4 {
		fromXOffset += float64(o.SwipeOffset[0])
		fromYOffset += float64(o.SwipeOffset[1])
		toXOffset += float64(o.SwipeOffset[2])
		toYOffset += float64(o.SwipeOffset[3])
	}
	fromXOffset += o.generateRandomOffset()
	fromYOffset += o.generateRandomOffset()
	toXOffset += o.generateRandomOffset()
	toYOffset += o.generateRandomOffset()

	if fromXOffset != 0.0 || fromYOffset != 0.0 || toXOffset != 0.0 || toYOffset != 0.0 {
		log.Debug().
			Float64("from_x_offset", fromXOffset).
			Float64("from_y_offset", fromYOffset).
			Float64("to_x_offset", toXOffset).
			Float64("to_y_offset", toYOffset).
			Msg("ApplySwipeOffset")
	}
	return absFromX + fromXOffset, absFromY + fromYOffset,
		absToX + toXOffset, absToY + toYOffset
}

func (o *ActionOptions) generateRandomOffset() float64 {
	if len(o.OffsetRandomRange) != 2 {
		// invalid offset random range, should be [min, max]
		return 0
	}

	minOffset := o.OffsetRandomRange[0]
	maxOffset := o.OffsetRandomRange[1]
	return float64(builtin.GetRandomNumber(minOffset, maxOffset)) + rand.Float64()
}

func MergeOptions(data map[string]interface{}, opts ...ActionOption) {
	o := NewActionOptions(opts...)
	if o.Identifier != "" {
		data["log"] = map[string]interface{}{
			"enable": true,
			"data":   o.Identifier,
		}
	}

	if o.Steps > 0 {
		data["steps"] = o.Steps
	}
	if _, ok := data["steps"]; !ok {
		data["steps"] = 12 // default steps
	}

	if o.Duration > 0 {
		data["duration"] = o.Duration
	}
	if _, ok := data["duration"]; !ok {
		data["duration"] = 0 // default duration
	}

	if o.PressDuration > 0 {
		data["pressDuration"] = o.PressDuration
	}

	if o.Frequency > 0 {
		data["frequency"] = o.Frequency
	}
	if _, ok := data["frequency"]; !ok {
		data["frequency"] = 10 // default frequency
	}

	if _, ok := data["replace"]; !ok {
		data["replace"] = true // default true
	}

	// custom options
	if o.Custom != nil {
		for k, v := range o.Custom {
			data[k] = v
		}
	}
}

func NewActionOptions(opts ...ActionOption) *ActionOptions {
	actionOptions := &ActionOptions{}
	for _, option := range opts {
		option(actionOptions)
	}
	return actionOptions
}

type ActionOption func(o *ActionOptions)

func WithContext(ctx context.Context) ActionOption {
	return func(o *ActionOptions) {
		o.Context = ctx
	}
}

func WithCustomOption(key string, value interface{}) ActionOption {
	return func(o *ActionOptions) {
		if o.Custom == nil {
			o.Custom = make(map[string]interface{})
		}
		o.Custom[key] = value
	}
}

func WithIdentifier(identifier string) ActionOption {
	return func(o *ActionOptions) {
		o.Identifier = identifier
	}
}

// set alias for compatibility
var WithWaitTime = WithInterval

func WithInterval(sec float64) ActionOption {
	return func(o *ActionOptions) {
		o.Interval = sec
	}
}

func WithDuration(duration float64) ActionOption {
	return func(o *ActionOptions) {
		o.Duration = duration
	}
}

func WithPressDuration(pressDuration float64) ActionOption {
	return func(o *ActionOptions) {
		o.PressDuration = pressDuration
	}
}

func WithSteps(steps int) ActionOption {
	return func(o *ActionOptions) {
		o.Steps = steps
	}
}

// WithDirection inputs direction (up, down, left, right)
func WithDirection(direction string) ActionOption {
	return func(o *ActionOptions) {
		o.Direction = direction
	}
}

// WithCustomDirection inputs sx, sy, ex, ey
func WithCustomDirection(sx, sy, ex, ey float64) ActionOption {
	return func(o *ActionOptions) {
		o.Direction = []float64{sx, sy, ex, ey}
	}
}

// swipe [fromX, fromY, toX, toY] with offset [offsetFromX, offsetFromY, offsetToX, offsetToY]
func WithSwipeOffset(offsetFromX, offsetFromY, offsetToX, offsetToY int) ActionOption {
	return func(o *ActionOptions) {
		o.SwipeOffset = []int{offsetFromX, offsetFromY, offsetToX, offsetToY}
	}
}

func WithOffsetRandomRange(min, max int) ActionOption {
	return func(o *ActionOptions) {
		o.OffsetRandomRange = []int{min, max}
	}
}

func WithFrequency(frequency int) ActionOption {
	return func(o *ActionOptions) {
		o.Frequency = frequency
	}
}

func WithMaxRetryTimes(maxRetryTimes int) ActionOption {
	return func(o *ActionOptions) {
		o.MaxRetryTimes = maxRetryTimes
	}
}

func WithTimeout(seconds int) ActionOption {
	return func(o *ActionOptions) {
		o.Timeout = seconds
	}
}

func WithTimeLimit(seconds int) ActionOption {
	return func(o *ActionOptions) {
		o.TimeLimit = seconds
	}
}

func WithIgnoreNotFoundError(ignoreError bool) ActionOption {
	return func(o *ActionOptions) {
		o.IgnoreNotFoundError = ignoreError
	}
}

func WithAntiRisk(antiRisk bool) ActionOption {
	return func(o *ActionOptions) {
		o.AntiRisk = antiRisk
	}
}

func WithResetHistory(resetHistory bool) ActionOption {
	return func(o *ActionOptions) {
		o.ResetHistory = resetHistory
	}
}

// WithOutputSchema sets the custom output schema for structured AI query response
func WithOutputSchema(schema interface{}) ActionOption {
	return func(o *ActionOptions) {
		o.OutputSchema = schema
	}
}

// GetMCPOptions generates MCP tool options for specific action types
func (o *ActionOptions) GetMCPOptions(actionType ActionName) []mcp.ToolOption {
	// Define field mappings for different action types
	fieldMappings := map[ActionName][]string{
		ACTION_TapXY:                    {"platform", "serial", "x", "y", "duration"},
		ACTION_TapAbsXY:                 {"platform", "serial", "x", "y", "duration"},
		ACTION_TapByOCR:                 {"platform", "serial", "text", "ignoreNotFoundError", "maxRetryTimes", "index", "regex", "tapRandomRect"},
		ACTION_TapByCV:                  {"platform", "serial", "ignoreNotFoundError", "maxRetryTimes", "index", "tapRandomRect"},
		ACTION_DoubleTapXY:              {"platform", "serial", "x", "y"},
		ACTION_SwipeDirection:           {"platform", "serial", "direction", "duration", "pressDuration"},
		ACTION_SwipeCoordinate:          {"platform", "serial", "fromX", "fromY", "toX", "toY", "duration", "pressDuration"},
		ACTION_Swipe:                    {"platform", "serial", "direction", "fromX", "fromY", "toX", "toY", "duration", "pressDuration"},
		ACTION_Drag:                     {"platform", "serial", "fromX", "fromY", "toX", "toY", "duration", "pressDuration"},
		ACTION_Input:                    {"platform", "serial", "text", "frequency"},
		ACTION_Backspace:                {"platform", "serial", "count"},
		ACTION_AppLaunch:                {"platform", "serial", "packageName"},
		ACTION_AppTerminate:             {"platform", "serial", "packageName"},
		ACTION_ColdLaunch:               {"platform", "serial", "packageName"},
		ACTION_AppInstall:               {"platform", "serial", "appUrl", "packageName"},
		ACTION_AppUninstall:             {"platform", "serial", "packageName"},
		ACTION_AppClear:                 {"platform", "serial", "packageName"},
		ACTION_GetForegroundApp:         {"platform", "serial"},
		ACTION_PressButton:              {"platform", "serial", "button"},
		ACTION_SwipeToTapApp:            {"platform", "serial", "appName", "ignoreNotFoundError", "maxRetryTimes", "index"},
		ACTION_SwipeToTapText:           {"platform", "serial", "text", "ignoreNotFoundError", "maxRetryTimes", "index", "regex"},
		ACTION_SwipeToTapTexts:          {"platform", "serial", "texts", "ignoreNotFoundError", "maxRetryTimes", "index", "regex"},
		ACTION_SecondaryClick:           {"platform", "serial", "x", "y"},
		ACTION_HoverBySelector:          {"platform", "serial", "selector"},
		ACTION_TapBySelector:            {"platform", "serial", "selector"},
		ACTION_SecondaryClickBySelector: {"platform", "serial", "selector"},
		ACTION_WebCloseTab:              {"platform", "serial", "tabIndex"},
		ACTION_WebLoginNoneUI:           {"platform", "serial", "packageName", "phoneNumber", "captcha", "password"},
		ACTION_SetIme:                   {"platform", "serial", "ime"},
		ACTION_GetSource:                {"platform", "serial", "packageName"},
		ACTION_Sleep:                    {"seconds"},
		ACTION_SleepMS:                  {"platform", "serial", "milliseconds"},
		ACTION_SleepRandom:              {"platform", "serial", "params"},
		ACTION_AIAction:                 {"platform", "serial", "prompt", "llm_service", "cv_service"},
		ACTION_StartToGoal:              {"platform", "serial", "prompt", "llm_service", "cv_service"},
		ACTION_Query:                    {"platform", "serial", "prompt", "llm_service", "cv_service", "output_schema"},
		ACTION_AIAssert:                 {"platform", "serial", "prompt", "llm_service", "cv_service"},
		ACTION_Finished:                 {"content"},
		ACTION_ListAvailableDevices:     {},
		ACTION_SelectDevice:             {"platform", "serial"},
		ACTION_ScreenShot:               {"platform", "serial"},
		ACTION_ScreenRecord:             {"platform", "serial", "duration", "screenRecordPath", "screenRecordWithAudio", "screenRecordWithScrcpy"},
		ACTION_GetScreenSize:            {"platform", "serial"},
		ACTION_Home:                     {"platform", "serial"},
		ACTION_Back:                     {"platform", "serial"},
		ACTION_ListPackages:             {"platform", "serial"},
		ACTION_ClosePopups:              {"platform", "serial"},

		// SIM specific actions using fromX/fromY for startX/startY and SIM-prefixed fields
		ACTION_SIMSwipeDirection:        {"platform", "serial", "direction", "fromX", "fromY", "sim_min_distance", "sim_max_distance", "duration", "pressDuration"},
		ACTION_SIMSwipeInArea:           {"platform", "serial", "direction", "sim_area_start_x", "sim_area_start_y", "sim_area_end_x", "sim_area_end_y", "sim_min_distance", "sim_max_distance", "duration", "pressDuration"},
		ACTION_SIMSwipeFromPointToPoint: {"platform", "serial", "fromX", "fromY", "toX", "toY", "duration", "pressDuration"},
		ACTION_SIMClickAtPoint:          {"platform", "serial", "x", "y", "duration", "pressDuration"},
		ACTION_SIMInput:                 {"platform", "serial", "text", "frequency"},
	}

	fields := fieldMappings[actionType]
	// Generate options for specified fields, or all fields if not mapped
	return o.generateMCPOptionsForFields(fields)
}

// generateMCPOptionsForFields generates MCP options for specific fields
func (o *ActionOptions) generateMCPOptionsForFields(fields []string) []mcp.ToolOption {
	options := make([]mcp.ToolOption, 0)

	// If no fields are specified, return empty options (e.g., for ACTION_ListAvailableDevices)
	if len(fields) == 0 {
		return options
	}

	rType := reflect.TypeOf(*o)

	// Process specific fields
	fieldMap := make(map[string]reflect.StructField)
	for i := 0; i < rType.NumField(); i++ {
		field := rType.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" && jsonTag != "-" {
			name := strings.Split(jsonTag, ",")[0]
			fieldMap[name] = field
		}
	}

	for _, fieldName := range fields {
		field, exists := fieldMap[fieldName]
		if !exists {
			continue
		}

		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}
		name := strings.Split(jsonTag, ",")[0]
		binding := field.Tag.Get("binding")
		required := strings.Contains(binding, "required")
		desc := field.Tag.Get("desc")

		// Handle pointer types
		fieldType := field.Type
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		switch fieldType.Kind() {
		case reflect.Float64, reflect.Float32, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if required {
				options = append(options, mcp.WithNumber(name, mcp.Required(), mcp.Description(desc)))
			} else {
				options = append(options, mcp.WithNumber(name, mcp.Description(desc)))
			}
		case reflect.String:
			if required {
				options = append(options, mcp.WithString(name, mcp.Required(), mcp.Description(desc)))
			} else {
				options = append(options, mcp.WithString(name, mcp.Description(desc)))
			}
		case reflect.Bool:
			if required {
				options = append(options, mcp.WithBoolean(name, mcp.Required(), mcp.Description(desc)))
			} else {
				options = append(options, mcp.WithBoolean(name, mcp.Description(desc)))
			}
		case reflect.Slice:
			if fieldType.Elem().Kind() == reflect.String || fieldType.Elem().Kind() == reflect.Float64 {
				if required {
					options = append(options, mcp.WithArray(name, mcp.Required(), mcp.Description(desc)))
				} else {
					options = append(options, mcp.WithArray(name, mcp.Description(desc)))
				}
			}
		case reflect.Map, reflect.Interface:
			// Handle OutputSchema as object type
			if name == "output_schema" {
				if required {
					options = append(options, mcp.WithObject(name, mcp.Required(), mcp.Description(desc)))
				} else {
					options = append(options, mcp.WithObject(name, mcp.Description(desc)))
				}
			}
			// Skip other map and interface types for now
			continue
		default:
			log.Warn().Str("field_type", fieldType.String()).Msg("Unsupported field type")
		}
	}

	return options
}
