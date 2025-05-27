package option

import (
	"context"
	"math/rand/v2"

	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/rs/zerolog/log"
)

type ActionMethod string

const (
	ACTION_LOG              ActionMethod = "log"
	ACTION_ListPackages     ActionMethod = "list_packages"
	ACTION_AppInstall       ActionMethod = "app_install"
	ACTION_AppUninstall     ActionMethod = "app_uninstall"
	ACTION_WebLoginNoneUI   ActionMethod = "web_login_none_ui"
	ACTION_AppClear         ActionMethod = "app_clear"
	ACTION_AppStart         ActionMethod = "app_start"
	ACTION_AppLaunch        ActionMethod = "app_launch" // 启动 app 并堵塞等待 app 首屏加载完成
	ACTION_AppTerminate     ActionMethod = "app_terminate"
	ACTION_AppStop          ActionMethod = "app_stop"
	ACTION_ScreenShot       ActionMethod = "screenshot"
	ACTION_GetScreenSize    ActionMethod = "get_screen_size"
	ACTION_Sleep            ActionMethod = "sleep"
	ACTION_SleepMS          ActionMethod = "sleep_ms"
	ACTION_SleepRandom      ActionMethod = "sleep_random"
	ACTION_SetIme           ActionMethod = "set_ime"
	ACTION_GetSource        ActionMethod = "get_source"
	ACTION_GetForegroundApp ActionMethod = "get_foreground_app"

	// UI handling
	ACTION_Home                     ActionMethod = "home"
	ACTION_Tap                      ActionMethod = "tap" // generic tap action
	ACTION_TapXY                    ActionMethod = "tap_xy"
	ACTION_TapAbsXY                 ActionMethod = "tap_abs_xy"
	ACTION_TapByOCR                 ActionMethod = "tap_ocr"
	ACTION_TapByCV                  ActionMethod = "tap_cv"
	ACTION_DoubleTap                ActionMethod = "double_tap" // generic double tap action
	ACTION_DoubleTapXY              ActionMethod = "double_tap_xy"
	ACTION_Swipe                    ActionMethod = "swipe"            // swipe by direction or coordinates
	ACTION_SwipeDirection           ActionMethod = "swipe_direction"  // swipe by direction (up, down, left, right)
	ACTION_SwipeCoordinate          ActionMethod = "swipe_coordinate" // swipe by coordinates (fromX, fromY, toX, toY)
	ACTION_Drag                     ActionMethod = "drag"
	ACTION_Input                    ActionMethod = "input"
	ACTION_PressButton              ActionMethod = "press_button"
	ACTION_Back                     ActionMethod = "back"
	ACTION_KeyCode                  ActionMethod = "keycode"
	ACTION_Delete                   ActionMethod = "delete"    // delete action
	ACTION_Backspace                ActionMethod = "backspace" // backspace action
	ACTION_AIAction                 ActionMethod = "ai_action" // action with ai
	ACTION_TapBySelector            ActionMethod = "tap_by_selector"
	ACTION_HoverBySelector          ActionMethod = "hover_by_selector"
	ACTION_Hover                    ActionMethod = "hover"       // generic hover action
	ACTION_RightClick               ActionMethod = "right_click" // right click action
	ACTION_WebCloseTab              ActionMethod = "web_close_tab"
	ACTION_SecondaryClick           ActionMethod = "secondary_click"
	ACTION_SecondaryClickBySelector ActionMethod = "secondary_click_by_selector"
	ACTION_GetElementTextBySelector ActionMethod = "get_element_text_by_selector"
	ACTION_Scroll                   ActionMethod = "scroll"         // scroll action
	ACTION_Upload                   ActionMethod = "upload"         // upload action
	ACTION_PushMedia                ActionMethod = "push_media"     // push media action
	ACTION_CreateBrowser            ActionMethod = "create_browser" // create browser action
	ACTION_AppInfo                  ActionMethod = "app_info"       // get app info action

	// device actions
	ACTION_ListAvailableDevices ActionMethod = "list_available_devices"
	ACTION_SelectDevice         ActionMethod = "select_device"

	// custom actions
	ACTION_SwipeToTapApp   ActionMethod = "swipe_to_tap_app"   // swipe left & right to find app and tap
	ACTION_SwipeToTapText  ActionMethod = "swipe_to_tap_text"  // swipe up & down to find text and tap
	ACTION_SwipeToTapTexts ActionMethod = "swipe_to_tap_texts" // swipe up & down to find text and tap
	ACTION_ClosePopups     ActionMethod = "close_popups"
	ACTION_EndToEndDelay   ActionMethod = "live_e2e"
	ACTION_InstallApp      ActionMethod = "install_app"
	ACTION_UninstallApp    ActionMethod = "uninstall_app"
	ACTION_DownloadApp     ActionMethod = "download_app"
	ACTION_Finished        ActionMethod = "finished"
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
	Context context.Context `json:"-" yaml:"-"`
	// log
	Identifier string `json:"identifier,omitempty" yaml:"identifier,omitempty"` // used to identify the action in log

	// control related
	MaxRetryTimes int         `json:"max_retry_times,omitempty" yaml:"max_retry_times,omitempty"` // max retry times
	Interval      float64     `json:"interval,omitempty" yaml:"interval,omitempty"`               // interval between retries in seconds
	Duration      float64     `json:"duration,omitempty" yaml:"duration,omitempty"`               // used to set duration in seconds
	PressDuration float64     `json:"press_duration,omitempty" yaml:"press_duration,omitempty"`   // used to set press duration in seconds
	Steps         int         `json:"steps,omitempty" yaml:"steps,omitempty"`                     // used to set steps of action
	Direction     interface{} `json:"direction,omitempty" yaml:"direction,omitempty"`             // used by swipe to tap text or app
	Timeout       int         `json:"timeout,omitempty" yaml:"timeout,omitempty"`                 // TODO: wait timeout in seconds for mobile action
	Frequency     int         `json:"frequency,omitempty" yaml:"frequency,omitempty"`

	ScreenOptions

	// set custiom options such as textview, id, description
	Custom map[string]interface{} `json:"custom,omitempty" yaml:"custom,omitempty"`
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

func (o *ActionOptions) ApplyTapOffset(absX, absY float64) (float64, float64) {
	if len(o.TapOffset) == 2 {
		absX += float64(o.TapOffset[0])
		absY += float64(o.TapOffset[1])
	}
	absX += o.generateRandomOffset()
	absY += o.generateRandomOffset()
	return absX, absY
}

func (o *ActionOptions) ApplySwipeOffset(absFromX, absFromY, absToX, absToY float64) (
	float64, float64, float64, float64) {
	if len(o.SwipeOffset) == 4 {
		absFromX += float64(o.SwipeOffset[0])
		absFromY += float64(o.SwipeOffset[1])
		absToX += float64(o.SwipeOffset[2])
		absToY += float64(o.SwipeOffset[3])
	}
	absFromX += o.generateRandomOffset()
	absFromY += o.generateRandomOffset()
	absToX += o.generateRandomOffset()
	absToY += o.generateRandomOffset()
	return absFromX, absFromY, absToX, absToY
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
	if actionOptions.MaxRetryTimes == 0 {
		actionOptions.MaxRetryTimes = 1
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

func WithTimeout(timeout int) ActionOption {
	return func(o *ActionOptions) {
		o.Timeout = timeout
	}
}

func WithIgnoreNotFoundError(ignoreError bool) ActionOption {
	return func(o *ActionOptions) {
		o.IgnoreNotFoundError = ignoreError
	}
}
