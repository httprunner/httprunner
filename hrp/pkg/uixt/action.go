package uixt

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
)

type ActionMethod string

const (
	ACTION_AppInstall   ActionMethod = "install"
	ACTION_AppUninstall ActionMethod = "uninstall"
	ACTION_AppStart     ActionMethod = "app_start"
	ACTION_AppLaunch    ActionMethod = "app_launch" // 启动 app 并堵塞等待 app 首屏加载完成
	ACTION_AppTerminate ActionMethod = "app_terminate"
	ACTION_AppStop      ActionMethod = "app_stop"
	ACTION_ScreenShot   ActionMethod = "screenshot"
	ACTION_Sleep        ActionMethod = "sleep"
	ACTION_SleepRandom  ActionMethod = "sleep_random"
	ACTION_StartCamera  ActionMethod = "camera_start" // alias for app_launch camera
	ACTION_StopCamera   ActionMethod = "camera_stop"  // alias for app_terminate camera

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

	// custom actions
	ACTION_SwipeToTapApp   ActionMethod = "swipe_to_tap_app"   // swipe left & right to find app and tap
	ACTION_SwipeToTapText  ActionMethod = "swipe_to_tap_text"  // swipe up & down to find text and tap
	ACTION_SwipeToTapTexts ActionMethod = "swipe_to_tap_texts" // swipe up & down to find text and tap
	ACTION_VideoCrawler    ActionMethod = "video_crawler"
)

type MobileAction struct {
	Method  ActionMethod   `json:"method,omitempty" yaml:"method,omitempty"`
	Params  interface{}    `json:"params,omitempty" yaml:"params,omitempty"`
	Options *ActionOptions `json:"options,omitempty" yaml:"options,omitempty"`
	ActionOptions
}

func (ma MobileAction) GetOptions() []ActionOption {
	var actionOptionList []ActionOption
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
	PressDuration       float64     `json:"duration,omitempty" yaml:"duration,omitempty"`                         // used to set duration of ios swipe action
	Steps               int         `json:"steps,omitempty" yaml:"steps,omitempty"`                               // used to set steps of android swipe action
	Direction           interface{} `json:"direction,omitempty" yaml:"direction,omitempty"`                       // used by swipe to tap text or app
	Timeout             int         `json:"timeout,omitempty" yaml:"timeout,omitempty"`                           // TODO: wait timeout in seconds for mobile action
	Frequency           int         `json:"frequency,omitempty" yaml:"frequency,omitempty"`

	// scope related
	Scope    Scope    `json:"scope,omitempty" yaml:"scope,omitempty"`
	AbsScope AbsScope `json:"abs_scope,omitempty" yaml:"abs_scope,omitempty"`

	Regex  bool  `json:"regex,omitempty" yaml:"regex,omitempty"`   // use regex to match text
	Offset []int `json:"offset,omitempty" yaml:"offset,omitempty"` // used to tap offset of point
	Index  int   `json:"index,omitempty" yaml:"index,omitempty"`   // index of the target element

	// set custiom options such as textview, id, description
	Custom map[string]interface{} `json:"custom,omitempty" yaml:"custom,omitempty"`
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
	}
	if len(o.Offset) == 2 {
		options = append(options, WithOffset(o.Offset[0], o.Offset[1]))
	}
	if o.Regex {
		options = append(options, WithRegex(true))
	}

	// custom options
	if o.Custom != nil {
		for k, v := range o.Custom {
			options = append(options, WithCustomOption(k, v))
		}
	}

	return options
}

func NewActionOptions(options ...ActionOption) *ActionOptions {
	actionOptions := &ActionOptions{}
	for _, option := range options {
		option(actionOptions)
	}
	return actionOptions
}

func mergeDataWithOptions(data map[string]interface{}, options ...ActionOption) map[string]interface{} {
	actionOptions := NewActionOptions(options...)

	if actionOptions.Identifier != "" {
		data["log"] = map[string]interface{}{
			"enable": true,
			"data":   actionOptions.Identifier,
		}
	}

	// handle point offset
	if len(actionOptions.Offset) == 2 {
		if x, ok := data["x"]; ok {
			xf, _ := builtin.Interface2Float64(x)
			data["x"] = xf + float64(actionOptions.Offset[0])
		}
		if y, ok := data["y"]; ok {
			yf, _ := builtin.Interface2Float64(y)
			data["y"] = yf + float64(actionOptions.Offset[1])
		}
	}

	if actionOptions.Steps > 0 {
		data["steps"] = actionOptions.Steps
	}
	if _, ok := data["steps"]; !ok {
		data["steps"] = 12 // default steps
	}

	if actionOptions.PressDuration > 0 {
		data["duration"] = actionOptions.PressDuration
	}
	if _, ok := data["duration"]; !ok {
		data["duration"] = 0 // default duration
	}

	if actionOptions.Frequency > 0 {
		data["frequency"] = actionOptions.Frequency
	}
	if _, ok := data["frequency"]; !ok {
		data["frequency"] = 60 // default frequency
	}

	if _, ok := data["isReplace"]; !ok {
		data["isReplace"] = true // default true
	}

	// custom options
	if actionOptions.Custom != nil {
		for k, v := range actionOptions.Custom {
			data[k] = v
		}
	}

	return data
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

func WithPressDuration(duration float64) ActionOption {
	return func(o *ActionOptions) {
		o.PressDuration = duration
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

func WithOffset(offsetX, offsetY int) ActionOption {
	return func(o *ActionOptions) {
		o.Offset = []int{offsetX, offsetY}
	}
}

func WithRegex(regex bool) ActionOption {
	return func(o *ActionOptions) {
		o.Regex = regex
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
	absX1 := int(x1 * float64(dExt.windowSize.Width))
	absY1 := int(y1 * float64(dExt.windowSize.Height))
	absX2 := int(x2 * float64(dExt.windowSize.Width))
	absY2 := int(y2 * float64(dExt.windowSize.Height))
	return AbsScope{absX1, absY1, absX2, absY2}
}

func (dExt *DriverExt) DoAction(action MobileAction) error {
	log.Debug().
		Str("method", string(action.Method)).
		Interface("params", action.Params).
		Msg("uixt action start")
	actionStartTime := time.Now()

	defer func() {
		log.Debug().
			Str("method", string(action.Method)).
			Interface("params", action.Params).
			Float64("elapsed(s)", time.Since(actionStartTime).Seconds()).
			Msg("uixt action end")
	}()

	switch action.Method {
	case ACTION_AppInstall:
		// TODO
		return errActionNotImplemented
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
		return fmt.Errorf("invalid %s params, should be app text([]string), got %v",
			ACTION_SwipeToTapText, action.Params)
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
	case ACTION_Home:
		return dExt.Driver.Homescreen()
	case ACTION_TapXY:
		if location, ok := action.Params.([]interface{}); ok {
			// relative x,y of window size: [0.5, 0.5]
			if len(location) != 2 {
				return fmt.Errorf("invalid tap location params: %v", location)
			}
			x, _ := location[0].(float64)
			y, _ := location[1].(float64)
			return dExt.TapXY(x, y, action.GetOptions()...)
		}
		return fmt.Errorf("invalid %s params: %v", ACTION_TapXY, action.Params)
	case ACTION_TapAbsXY:
		if location, ok := action.Params.([]interface{}); ok {
			// absolute coordinates x,y of window size: [100, 300]
			if len(location) != 2 {
				return fmt.Errorf("invalid tap location params: %v", location)
			}
			x, _ := location[0].(float64)
			y, _ := location[1].(float64)
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
		if imagePath, ok := action.Params.(string); ok {
			return dExt.TapByCV(imagePath, action.GetOptions()...)
		}
		return fmt.Errorf("invalid %s params: %v", ACTION_TapByCV, action.Params)
	case ACTION_DoubleTapXY:
		if location, ok := action.Params.([]interface{}); ok {
			// relative x,y of window size: [0.5, 0.5]
			if len(location) != 2 {
				return fmt.Errorf("invalid tap location params: %v", location)
			}
			x, _ := location[0].(float64)
			y, _ := location[1].(float64)
			return dExt.DoubleTapXY(x, y)
		}
		return fmt.Errorf("invalid %s params: %v", ACTION_DoubleTapXY, action.Params)
	case ACTION_DoubleTap:
		if param, ok := action.Params.(string); ok {
			return dExt.DoubleTap(param)
		}
		return fmt.Errorf("invalid %s params: %v", ACTION_DoubleTap, action.Params)
	case ACTION_Swipe:
		swipeAction := dExt.prepareSwipeAction(action.GetOptions()...)
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
		}
		return fmt.Errorf("invalid sleep params: %v(%T)", action.Params, action.Params)
	case ACTION_SleepRandom:
		if params, ok := action.Params.([]interface{}); ok {
			return sleepRandom(time.Now(), params)
		}
		return fmt.Errorf("invalid sleep random params: %v(%T)", action.Params, action.Params)
	case ACTION_ScreenShot:
		// take screenshot
		log.Info().Msg("take screenshot for current screen")
		_, _, err := dExt.takeScreenShot(builtin.GenNameWithTimestamp("%d_screenshot"))
		return err
	case ACTION_StartCamera:
		return dExt.Driver.StartCamera()
	case ACTION_StopCamera:
		return dExt.Driver.StopCamera()
	case ACTION_VideoCrawler:
		params, ok := action.Params.(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid video crawler params: %v(%T)", action.Params, action.Params)
		}
		data, _ := json.Marshal(params)
		configs := &VideoCrawlerConfigs{}
		if err := json.Unmarshal(data, configs); err != nil {
			return errors.Wrapf(err, "invalid video crawler params: %v(%T)", action.Params, action.Params)
		}
		return dExt.VideoCrawler(configs)
	}
	return nil
}

var errActionNotImplemented = errors.New("UI action not implemented")

func convertToFloat64(val interface{}) (float64, error) {
	switch v := val.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	default:
		return 0, fmt.Errorf("invalid type for conversion to float64: %T, value: %+v", val, val)
	}
}

// sleepRandom sleeps random time with given params
// startTime is used to correct sleep duration caused by process time
func sleepRandom(startTime time.Time, params []interface{}) error {
	if len(params) == 1 {
		// constant sleep time
		params = append(params, params[0], 1.0)
	} else if len(params) == 2 {
		// append default weight 1
		params = append(params, 1.0)
	}

	var sections []struct {
		min, max, weight float64
	}
	totalProb := 0.0
	for i := 0; i+3 <= len(params); i += 3 {
		min, err := convertToFloat64(params[i])
		if err != nil {
			return errors.Wrapf(err, "invalid minimum time: %v", params[i])
		}
		max, err := convertToFloat64(params[i+1])
		if err != nil {
			return errors.Wrapf(err, "invalid maximum time: %v", params[i+1])
		}
		weight, err := convertToFloat64(params[i+2])
		if err != nil {
			return errors.Wrapf(err, "invalid weight value: %v", params[i+2])
		}
		totalProb += weight
		sections = append(sections,
			struct{ min, max, weight float64 }{min, max, weight},
		)
	}

	if totalProb == 0 {
		log.Warn().Msg("total weight is 0, skip sleep")
		return nil
	}

	r := rand.Float64()
	accProb := 0.0
	for _, s := range sections {
		accProb += s.weight / totalProb
		if r < accProb {
			elapsed := time.Since(startTime).Seconds()
			randomSeconds := s.min + rand.Float64()*(s.max-s.min)
			dur := randomSeconds - elapsed

			// if elapsed time is greater than random seconds, skip sleep to reduce deviation caused by process time
			if dur <= 0 {
				log.Info().Float64("elapsed", elapsed).Float64("randomSeconds", randomSeconds).
					Interface("strategy_params", params).Msg("elapsed duration >= random seconds, skip sleep")
			} else {
				log.Info().Float64("sleepDuration", dur).Float64("elapsed", elapsed).Float64("randomSeconds", randomSeconds).
					Interface("strategy_params", params).Msg("sleep remaining random seconds")
				time.Sleep(time.Duration(math.Ceil(dur*1000)) * time.Millisecond)
			}

			return nil
		}
	}
	return nil
}
