package uixt

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
)

type ActionMethod string

const (
	ACTION_LOG              ActionMethod = "log"
	ACTION_AppInstall       ActionMethod = "install"
	ACTION_AppUninstall     ActionMethod = "uninstall"
	ACTION_AppClear         ActionMethod = "app_clear"
	ACTION_AppStart         ActionMethod = "app_start"
	ACTION_AppLaunch        ActionMethod = "app_launch" // 启动 app 并堵塞等待 app 首屏加载完成
	ACTION_AppTerminate     ActionMethod = "app_terminate"
	ACTION_AppStop          ActionMethod = "app_stop"
	ACTION_ScreenShot       ActionMethod = "screenshot"
	ACTION_Sleep            ActionMethod = "sleep"
	ACTION_SleepMS          ActionMethod = "sleep_ms"
	ACTION_SleepRandom      ActionMethod = "sleep_random"
	ACTION_StartCamera      ActionMethod = "camera_start" // alias for app_launch camera
	ACTION_StopCamera       ActionMethod = "camera_stop"  // alias for app_terminate camera
	ACTION_SetClipboard     ActionMethod = "set_clipboard"
	ACTION_GetClipboard     ActionMethod = "get_clipboard"
	ACTION_SetIme           ActionMethod = "set_ime"
	ACTION_GetSource        ActionMethod = "get_source"
	ACTION_GetForegroundApp ActionMethod = "get_foreground_app"

	// UI handling
	ACTION_Home        ActionMethod = "home"
	ACTION_TapXY       ActionMethod = "tap_xy"
	ACTION_TapAbsXY    ActionMethod = "tap_abs_xy"
	ACTION_TapByOCR    ActionMethod = "tap_ocr"
	ACTION_TapByCV     ActionMethod = "tap_cv"
	ACTION_Tap         ActionMethod = "tap"
	ACTION_DoubleTapXY ActionMethod = "double_tap_xy"
	ACTION_DoubleTap   ActionMethod = "double_tap"
	ACTION_Swipe       ActionMethod = "swipe"
	ACTION_Input       ActionMethod = "input"
	ACTION_Back        ActionMethod = "back"
	ACTION_KeyCode     ActionMethod = "keycode"

	// custom actions
	ACTION_SwipeToTapApp   ActionMethod = "swipe_to_tap_app"   // swipe left & right to find app and tap
	ACTION_SwipeToTapText  ActionMethod = "swipe_to_tap_text"  // swipe up & down to find text and tap
	ACTION_SwipeToTapTexts ActionMethod = "swipe_to_tap_texts" // swipe up & down to find text and tap
	ACTION_ClosePopups     ActionMethod = "close_popups"
	ACTION_EndToEndDelay   ActionMethod = "live_e2e"
	ACTION_InstallApp      ActionMethod = "install_app"
	ACTION_UninstallApp    ActionMethod = "uninstall_app"
	ACTION_DownloadApp     ActionMethod = "download_app"
)

const (
	// UI validation
	// selectors
	SelectorName          string = "ui_name"
	SelectorLabel         string = "ui_label"
	SelectorOCR           string = "ui_ocr"
	SelectorImage         string = "ui_image"
	SelectorForegroundApp string = "ui_foreground_app"
	// assertions
	AssertionEqual     string = "equal"
	AssertionNotEqual  string = "not_equal"
	AssertionExists    string = "exists"
	AssertionNotExists string = "not_exists"
)

type MobileAction struct {
	Method  ActionMethod   `json:"method,omitempty" yaml:"method,omitempty"`
	Params  interface{}    `json:"params,omitempty" yaml:"params,omitempty"`
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

// (x1, y1) is the top left corner, (x2, y2) is the bottom right corner
// [x1, y1, x2, y2] in percentage of the screen
type Scope []float64

// [x1, y1, x2, y2] in absolute pixels
type AbsScope []int

func (s AbsScope) Option() ActionOption {
	return WithAbsScope(s[0], s[1], s[2], s[3])
}

type ActionOptions struct {
	// log
	Identifier string `json:"identifier,omitempty" yaml:"identifier,omitempty"` // used to identify the action in log

	// control related
	MaxRetryTimes       int         `json:"max_retry_times,omitempty" yaml:"max_retry_times,omitempty"`           // max retry times
	IgnoreNotFoundError bool        `json:"ignore_NotFoundError,omitempty" yaml:"ignore_NotFoundError,omitempty"` // ignore error if target element not found
	Interval            float64     `json:"interval,omitempty" yaml:"interval,omitempty"`                         // interval between retries in seconds
	Duration            float64     `json:"duration,omitempty" yaml:"duration,omitempty"`                         // used to set duration of ios swipe action
	PressDuration       float64     `json:"press_duration,omitempty" yaml:"press_duration,omitempty"`             // used to set duration of ios swipe action
	Steps               int         `json:"steps,omitempty" yaml:"steps,omitempty"`                               // used to set steps of android swipe action
	Direction           interface{} `json:"direction,omitempty" yaml:"direction,omitempty"`                       // used by swipe to tap text or app
	Timeout             int         `json:"timeout,omitempty" yaml:"timeout,omitempty"`                           // TODO: wait timeout in seconds for mobile action
	Frequency           int         `json:"frequency,omitempty" yaml:"frequency,omitempty"`

	// scope related
	Scope    Scope    `json:"scope,omitempty" yaml:"scope,omitempty"`
	AbsScope AbsScope `json:"abs_scope,omitempty" yaml:"abs_scope,omitempty"`

	Regex             bool  `json:"regex,omitempty" yaml:"regex,omitempty"`                             // use regex to match text
	Offset            []int `json:"offset,omitempty" yaml:"offset,omitempty"`                           // used to tap offset of point
	OffsetRandomRange []int `json:"offset_random_range,omitempty" yaml:"offset_random_range,omitempty"` // set random range [min, max] for tap/swipe points
	Index             int   `json:"index,omitempty" yaml:"index,omitempty"`                             // index of the target element
	MatchOne          bool  `json:"match_one,omitempty" yaml:"match_one,omitempty"`                     // match one of the targets if existed

	// set custiom options such as textview, id, description
	Custom map[string]interface{} `json:"custom,omitempty" yaml:"custom,omitempty"`

	// screenshot related
	ScreenShotWithOCR            bool     `json:"screenshot_with_ocr,omitempty" yaml:"screenshot_with_ocr,omitempty"`
	ScreenShotWithUpload         bool     `json:"screenshot_with_upload,omitempty" yaml:"screenshot_with_upload,omitempty"`
	ScreenShotWithLiveType       bool     `json:"screenshot_with_live_type,omitempty" yaml:"screenshot_with_live_type,omitempty"`
	ScreenShotWithLivePopularity bool     `json:"screenshot_with_live_popularity,omitempty" yaml:"screenshot_with_live_popularity,omitempty"`
	ScreenShotWithUITypes        []string `json:"screenshot_with_ui_types,omitempty" yaml:"screenshot_with_ui_types,omitempty"`
	ScreenShotWithClosePopups    bool     `json:"screenshot_with_close_popups,omitempty" yaml:"screenshot_with_close_popups,omitempty"`
	ScreenShotWithOCRCluster     string   `json:"screenshot_with_ocr_cluster,omitempty" yaml:"screenshot_with_ocr_cluster,omitempty"`
	ScreenShotFileName           string   `json:"screenshot_file_name,omitempty" yaml:"screenshot_file_name,omitempty"`
}

func (o *ActionOptions) Options() []ActionOption {
	options := make([]ActionOption, 0)

	if o == nil {
		return options
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
		sx, _ := builtin.Interface2Float64(v[0])
		sy, _ := builtin.Interface2Float64(v[1])
		ex, _ := builtin.Interface2Float64(v[2])
		ey, _ := builtin.Interface2Float64(v[3])
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
	if len(o.Offset) == 2 {
		// for tap [x,y] offset
		options = append(options, WithTapOffset(o.Offset[0], o.Offset[1]))
	} else if len(o.Offset) == 4 {
		// for swipe [fromX, fromY, toX, toY] offset
		options = append(options, WithSwipeOffset(
			o.Offset[0], o.Offset[1], o.Offset[2], o.Offset[3]))
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

	// screenshot options
	if o.ScreenShotWithOCR {
		options = append(options, WithScreenShotOCR(true))
	}
	if o.ScreenShotWithUpload {
		options = append(options, WithScreenShotUpload(true))
	}
	if o.ScreenShotWithLiveType {
		options = append(options, WithScreenShotLiveType(true))
	}
	if o.ScreenShotWithLivePopularity {
		options = append(options, WithScreenShotLivePopularity(true))
	}
	if len(o.ScreenShotWithUITypes) > 0 {
		options = append(options, WithScreenShotUITypes(o.ScreenShotWithUITypes...))
	}
	if o.ScreenShotWithClosePopups {
		options = append(options, WithScreenShotClosePopups(true))
	}
	if o.ScreenShotWithOCRCluster != "" {
		options = append(options, WithScreenOCRCluster(o.ScreenShotWithOCRCluster))
	}
	if o.ScreenShotFileName != "" {
		options = append(options, WithScreenShotFileName(o.ScreenShotFileName))
	}

	return options
}

func (o *ActionOptions) screenshotActions() []string {
	actions := []string{}
	if o.ScreenShotWithUpload {
		actions = append(actions, "upload")
	}
	if o.ScreenShotWithOCR {
		actions = append(actions, "ocr")
	}
	if o.ScreenShotWithLiveType {
		actions = append(actions, "liveType")
	}
	if o.ScreenShotWithLivePopularity {
		actions = append(actions, "livePopularity")
	}
	// UI detection
	if len(o.ScreenShotWithUITypes) > 0 {
		actions = append(actions, "ui")
	}
	if o.ScreenShotWithClosePopups {
		actions = append(actions, "close")
	}
	return actions
}

func (o *ActionOptions) getRandomOffset() float64 {
	if len(o.OffsetRandomRange) != 2 {
		// invalid offset random range, should be [min, max]
		return 0
	}

	minOffset := o.OffsetRandomRange[0]
	maxOffset := o.OffsetRandomRange[1]
	return float64(builtin.GetRandomNumber(minOffset, maxOffset)) + rand.Float64()
}

func (o *ActionOptions) updateData(data map[string]interface{}) {
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

func NewActionOptions(options ...ActionOption) *ActionOptions {
	actionOptions := &ActionOptions{}
	for _, option := range options {
		option(actionOptions)
	}
	return actionOptions
}

type TapTextAction struct {
	Text    string
	Options []ActionOption
}

type ActionOption func(o *ActionOptions)

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

func WithIndex(index int) ActionOption {
	return func(o *ActionOptions) {
		o.Index = index
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

// WithScope inputs area of [(x1,y1), (x2,y2)]
// x1, y1, x2, y2 are all in [0, 1], which means the relative position of the screen
func WithScope(x1, y1, x2, y2 float64) ActionOption {
	return func(o *ActionOptions) {
		o.Scope = Scope{x1, y1, x2, y2}
	}
}

// WithAbsScope inputs area of [(x1,y1), (x2,y2)]
// x1, y1, x2, y2 are all absolute position of the screen
func WithAbsScope(x1, y1, x2, y2 int) ActionOption {
	return func(o *ActionOptions) {
		o.AbsScope = AbsScope{x1, y1, x2, y2}
	}
}

// Deprecated: use WithTapOffset instead
func WithOffset(offsetX, offsetY int) ActionOption {
	return func(o *ActionOptions) {
		o.Offset = []int{offsetX, offsetY}
	}
}

// tap [x, y] with offset [offsetX, offsetY]
var WithTapOffset = WithOffset

// swipe [fromX, fromY, toX, toY] with offset [offsetFromX, offsetFromY, offsetToX, offsetToY]
func WithSwipeOffset(offsetFromX, offsetFromY, offsetToX, offsetToY int) ActionOption {
	return func(o *ActionOptions) {
		o.Offset = []int{offsetFromX, offsetFromY, offsetToX, offsetToY}
	}
}

func WithOffsetRandomRange(min, max int) ActionOption {
	return func(o *ActionOptions) {
		o.OffsetRandomRange = []int{min, max}
	}
}

func WithRegex(regex bool) ActionOption {
	return func(o *ActionOptions) {
		o.Regex = regex
	}
}

func WithMatchOne(matchOne bool) ActionOption {
	return func(o *ActionOptions) {
		o.MatchOne = matchOne
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

func WithScreenShotOCR(ocrOn bool) ActionOption {
	return func(o *ActionOptions) {
		o.ScreenShotWithOCR = ocrOn
	}
}

func WithScreenShotUpload(uploadOn bool) ActionOption {
	return func(o *ActionOptions) {
		o.ScreenShotWithUpload = uploadOn
	}
}

func WithScreenShotLiveType(liveTypeOn bool) ActionOption {
	return func(o *ActionOptions) {
		o.ScreenShotWithLiveType = liveTypeOn
	}
}

func WithScreenShotLivePopularity(livePopularityOn bool) ActionOption {
	return func(o *ActionOptions) {
		o.ScreenShotWithLivePopularity = livePopularityOn
	}
}

func WithScreenShotUITypes(uiTypes ...string) ActionOption {
	return func(o *ActionOptions) {
		o.ScreenShotWithUITypes = uiTypes
	}
}

func WithScreenShotClosePopups(closeOn bool) ActionOption {
	return func(o *ActionOptions) {
		o.ScreenShotWithClosePopups = closeOn
	}
}

func WithScreenOCRCluster(ocrCluster string) ActionOption {
	return func(o *ActionOptions) {
		o.ScreenShotWithOCRCluster = ocrCluster
	}
}

func WithScreenShotFileName(fileName string) ActionOption {
	return func(o *ActionOptions) {
		o.ScreenShotFileName = fileName
	}
}

func (dExt *DriverExt) ParseActionOptions(options ...ActionOption) []ActionOption {
	actionOptions := NewActionOptions(options...)

	// convert relative scope to absolute scope
	if len(actionOptions.AbsScope) != 4 && len(actionOptions.Scope) == 4 {
		scope := actionOptions.Scope
		actionOptions.AbsScope = dExt.GenAbsScope(
			scope[0], scope[1], scope[2], scope[3])
	}

	return actionOptions.Options()
}

func (dExt *DriverExt) GenAbsScope(x1, y1, x2, y2 float64) AbsScope {
	// convert relative scope to absolute scope
	windowSize, _ := dExt.Driver.WindowSize()
	absX1 := int(x1 * float64(windowSize.Width))
	absY1 := int(y1 * float64(windowSize.Height))
	absX2 := int(x2 * float64(windowSize.Width))
	absY2 := int(y2 * float64(windowSize.Height))
	return AbsScope{absX1, absY1, absX2, absY2}
}

func (dExt *DriverExt) DoAction(action MobileAction) (err error) {
	actionStartTime := time.Now()
	defer func() {
		var logger *zerolog.Event
		if err != nil {
			logger = log.Error().Bool("success", false).Err(err)
		} else {
			logger = log.Debug().Bool("success", true)
		}
		logger = logger.
			Str("method", string(action.Method)).
			Interface("params", action.Params).
			Int64("elapsed(ms)", time.Since(actionStartTime).Milliseconds())
		logger.Msg("exec uixt action")
	}()

	switch action.Method {
	case ACTION_AppInstall:
		if appUrl, ok := action.Params.(string); ok {
			if err = dExt.InstallByUrl(appUrl, WithRetryTimes(action.MaxRetryTimes)); err != nil {
				return errors.Wrap(err, "failed to install app")
			}
		}
	case ACTION_AppUninstall:
		if packageName, ok := action.Params.(string); ok {
			if err = dExt.Uninstall(packageName, action.GetOptions()...); err != nil {
				return errors.Wrap(err, "failed to uninstall app")
			}
		}
	case ACTION_AppClear:
		if packageName, ok := action.Params.(string); ok {
			if err = dExt.Driver.Clear(packageName); err != nil {
				return errors.Wrap(err, "failed to clear app")
			}
		}
	case ACTION_AppLaunch:
		if bundleId, ok := action.Params.(string); ok {
			return dExt.Driver.AppLaunch(bundleId)
		}
		return fmt.Errorf("invalid %s params, should be bundleId(string), got %v",
			ACTION_AppLaunch, action.Params)
	case ACTION_SwipeToTapApp:
		if appName, ok := action.Params.(string); ok {
			return dExt.swipeToTapApp(appName, action.GetOptions()...)
		}
		return fmt.Errorf("invalid %s params, should be app name(string), got %v",
			ACTION_SwipeToTapApp, action.Params)
	case ACTION_SwipeToTapText:
		if text, ok := action.Params.(string); ok {
			return dExt.swipeToTapTexts([]string{text}, action.GetOptions()...)
		}
		return fmt.Errorf("invalid %s params, should be app text(string), got %v",
			ACTION_SwipeToTapText, action.Params)
	case ACTION_SwipeToTapTexts:
		if texts, ok := action.Params.([]string); ok {
			return dExt.swipeToTapTexts(texts, action.GetOptions()...)
		}
		if texts, err := builtin.ConvertToStringSlice(action.Params); err == nil {
			return dExt.swipeToTapTexts(texts, action.GetOptions()...)
		}
		return fmt.Errorf("invalid %s params: %v", ACTION_SwipeToTapTexts, action.Params)
	case ACTION_AppTerminate:
		if bundleId, ok := action.Params.(string); ok {
			success, err := dExt.Driver.AppTerminate(bundleId)
			if err != nil {
				return errors.Wrap(err, "failed to terminate app")
			}
			if !success {
				log.Warn().Str("bundleId", bundleId).Msg("app was not running")
			}
			return nil
		}
		return fmt.Errorf("app_terminate params should be bundleId(string), got %v", action.Params)
	case ACTION_SetClipboard:
		if text, ok := action.Params.(string); ok {
			err := dExt.Driver.SetPasteboard(PasteboardTypePlaintext, text)
			if err != nil {
				return errors.Wrap(err, "failed to set clipboard")
			}
			return nil
		}
		return fmt.Errorf("set_clioboard params should be text(string), got %v", action.Params)
	case ACTION_Home:
		return dExt.Driver.Homescreen()
	case ACTION_SetIme:
		if ime, ok := action.Params.(string); ok {
			err = dExt.Driver.SetIme(ime)
			if err != nil {
				return errors.Wrap(err, "failed to set ime")
			}
			return nil
		}
	case ACTION_GetSource:
		if packageName, ok := action.Params.(string); ok {
			source := NewSourceOption().WithProcessName(packageName)
			_, err = dExt.Driver.Source(source)
			if err != nil {
				return errors.Wrap(err, "failed to set ime")
			}
			return nil
		}
	case ACTION_TapXY:
		if params, err := builtin.ConvertToFloat64Slice(action.Params); err == nil {
			// relative x,y of window size: [0.5, 0.5]
			if len(params) != 2 {
				return fmt.Errorf("invalid tap location params: %v", params)
			}
			x, y := params[0], params[1]
			return dExt.TapXY(x, y, action.GetOptions()...)
		}
		return fmt.Errorf("invalid %s params: %v", ACTION_TapXY, action.Params)
	case ACTION_TapAbsXY:
		if params, err := builtin.ConvertToFloat64Slice(action.Params); err == nil {
			// absolute coordinates x,y of window size: [100, 300]
			if len(params) != 2 {
				return fmt.Errorf("invalid tap location params: %v", params)
			}
			x, y := params[0], params[1]
			return dExt.TapAbsXY(x, y, action.GetOptions()...)
		}
		return fmt.Errorf("invalid %s params: %v", ACTION_TapAbsXY, action.Params)
	case ACTION_Tap:
		if param, ok := action.Params.(string); ok {
			return dExt.Tap(param, action.GetOptions()...)
		}
		return fmt.Errorf("invalid %s params: %v", ACTION_Tap, action.Params)
	case ACTION_TapByOCR:
		if ocrText, ok := action.Params.(string); ok {
			return dExt.TapByOCR(ocrText, action.GetOptions()...)
		}
		return fmt.Errorf("invalid %s params: %v", ACTION_TapByOCR, action.Params)
	case ACTION_TapByCV:
		actionOptions := NewActionOptions(action.GetOptions()...)
		if len(actionOptions.ScreenShotWithUITypes) > 0 {
			return dExt.TapByUIDetection(action.GetOptions()...)
		}
		return fmt.Errorf("invalid %s params: %v", ACTION_TapByCV, action.Params)
	case ACTION_DoubleTapXY:
		if params, err := builtin.ConvertToFloat64Slice(action.Params); err == nil {
			// relative x,y of window size: [0.5, 0.5]
			if len(params) != 2 {
				return fmt.Errorf("invalid tap location params: %v", params)
			}
			x, y := params[0], params[1]
			return dExt.DoubleTapXY(x, y)
		}
		return fmt.Errorf("invalid %s params: %v", ACTION_DoubleTapXY, action.Params)
	case ACTION_DoubleTap:
		if param, ok := action.Params.(string); ok {
			return dExt.DoubleTap(param)
		}
		return fmt.Errorf("invalid %s params: %v", ACTION_DoubleTap, action.Params)
	case ACTION_Swipe:
		params := action.Params
		swipeAction := dExt.prepareSwipeAction(params, action.GetOptions()...)
		return swipeAction(dExt)
	case ACTION_Input:
		// input text on current active element
		// append \n to send text with enter
		// send \b\b\b to delete 3 chars
		param := fmt.Sprintf("%v", action.Params)
		return dExt.Driver.Input(param, action.GetOptions()...)
	case ACTION_Back:
		return dExt.Driver.PressBack()
	case ACTION_Sleep:
		if param, ok := action.Params.(json.Number); ok {
			seconds, _ := param.Float64()
			time.Sleep(time.Duration(seconds*1000) * time.Millisecond)
			return nil
		} else if param, ok := action.Params.(float64); ok {
			time.Sleep(time.Duration(param*1000) * time.Millisecond)
			return nil
		} else if param, ok := action.Params.(int64); ok {
			time.Sleep(time.Duration(param) * time.Second)
			return nil
		} else if sd, ok := action.Params.(SleepConfig); ok {
			sleepStrict(sd.StartTime, int64(sd.Seconds*1000))
			return nil
		}
		return fmt.Errorf("invalid sleep params: %v(%T)", action.Params, action.Params)
	case ACTION_SleepMS:
		if param, ok := action.Params.(json.Number); ok {
			milliseconds, _ := param.Int64()
			time.Sleep(time.Duration(milliseconds) * time.Millisecond)
			return nil
		} else if param, ok := action.Params.(int64); ok {
			time.Sleep(time.Duration(param) * time.Millisecond)
			return nil
		} else if sd, ok := action.Params.(SleepConfig); ok {
			sleepStrict(sd.StartTime, sd.Milliseconds)
			return nil
		}
		return fmt.Errorf("invalid sleep ms params: %v(%T)", action.Params, action.Params)
	case ACTION_SleepRandom:
		if params, err := builtin.ConvertToFloat64Slice(action.Params); err == nil {
			sleepStrict(time.Now(), getSimulationDuration(params))
			return nil
		}
		return fmt.Errorf("invalid sleep random params: %v(%T)", action.Params, action.Params)
	case ACTION_ScreenShot:
		// take screenshot
		log.Info().Msg("take screenshot for current screen")
		_, err := dExt.GetScreenResult(action.GetOptions()...)
		return err
	case ACTION_StartCamera:
		return dExt.Driver.StartCamera()
	case ACTION_StopCamera:
		return dExt.Driver.StopCamera()
	case ACTION_ClosePopups:
		return dExt.ClosePopupsHandler()
	case ACTION_EndToEndDelay:
		CollectEndToEndDelay(dExt, action.GetOptions()...)
		return nil
	}
	return nil
}

type SleepConfig struct {
	StartTime    time.Time `json:"start_time"`
	Seconds      float64   `json:"seconds,omitempty"`
	Milliseconds int64     `json:"milliseconds,omitempty"`
}

var errActionNotImplemented = errors.New("UI action not implemented")

// getSimulationDuration returns simulation duration by given params (in seconds)
func getSimulationDuration(params []float64) (milliseconds int64) {
	if len(params) == 1 {
		// given constant duration time
		return int64(params[0] * 1000)
	}

	if len(params) == 2 {
		// given [min, max], missing weight
		// append default weight 1
		params = append(params, 1.0)
	}

	var sections []struct {
		min, max, weight float64
	}
	totalProb := 0.0
	for i := 0; i+3 <= len(params); i += 3 {
		min := params[i]
		max := params[i+1]
		weight := params[i+2]
		totalProb += weight
		sections = append(sections,
			struct{ min, max, weight float64 }{min, max, weight},
		)
	}

	if totalProb == 0 {
		log.Warn().Msg("total weight is 0, skip simulation")
		return 0
	}

	r := rand.Float64()
	accProb := 0.0
	for _, s := range sections {
		accProb += s.weight / totalProb
		if r < accProb {
			milliseconds := int64((s.min + rand.Float64()*(s.max-s.min)) * 1000)
			log.Info().Int64("random(ms)", milliseconds).
				Interface("strategy_params", params).Msg("get simulation duration")
			return milliseconds
		}
	}

	log.Warn().Interface("strategy_params", params).
		Msg("get simulation duration failed, skip simulation")
	return 0
}

// sleepStrict sleeps strict duration with given params
// startTime is used to correct sleep duration caused by process time
func sleepStrict(startTime time.Time, strictMilliseconds int64) {
	var elapsed int64
	if !startTime.IsZero() {
		elapsed = time.Since(startTime).Milliseconds()
	}
	dur := strictMilliseconds - elapsed

	// if elapsed time is greater than given duration, skip sleep to reduce deviation caused by process time
	if dur <= 0 {
		log.Warn().
			Int64("elapsed(ms)", elapsed).
			Int64("strictSleep(ms)", strictMilliseconds).
			Msg("elapsed >= simulation duration, skip sleep")
		return
	}

	log.Info().Int64("sleepDuration(ms)", dur).
		Int64("elapsed(ms)", elapsed).
		Int64("strictSleep(ms)", strictMilliseconds).
		Msg("sleep remaining duration time")
	time.Sleep(time.Duration(dur) * time.Millisecond)
}
