package uixt

import (
	"encoding/json"
	"fmt"
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
	Method ActionMethod `json:"method,omitempty" yaml:"method,omitempty"`
	Params interface{}  `json:"params,omitempty" yaml:"params,omitempty"`

	Identifier          string      `json:"identifier,omitempty" yaml:"identifier,omitempty"`                     // used to identify the action in log
	MaxRetryTimes       int         `json:"max_retry_times,omitempty" yaml:"max_retry_times,omitempty"`           // max retry times
	WaitTime            float64     `json:"wait_time,omitempty" yaml:"wait_time,omitempty"`                       // wait time between swipe and ocr, unit: second
	Duration            float64     `json:"duration,omitempty" yaml:"duration,omitempty"`                         // used to set duration of ios swipe action
	Steps               int         `json:"steps,omitempty" yaml:"steps,omitempty"`                               // used to set steps of android swipe action
	Direction           interface{} `json:"direction,omitempty" yaml:"direction,omitempty"`                       // used by swipe to tap text or app
	Scope               []float64   `json:"scope,omitempty" yaml:"scope,omitempty"`                               // used by ocr to get text position in the scope
	Offset              []int       `json:"offset,omitempty" yaml:"offset,omitempty"`                             // used to tap offset of point
	Index               int         `json:"index,omitempty" yaml:"index,omitempty"`                               // index of the target element, should start from 1
	Timeout             int         `json:"timeout,omitempty" yaml:"timeout,omitempty"`                           // TODO: wait timeout in seconds for mobile action
	IgnoreNotFoundError bool        `json:"ignore_NotFoundError,omitempty" yaml:"ignore_NotFoundError,omitempty"` // ignore error if target element not found
	Text                string      `json:"text,omitempty" yaml:"text,omitempty"`
	ID                  string      `json:"id,omitempty" yaml:"id,omitempty"`
	Description         string      `json:"description,omitempty" yaml:"description,omitempty"`
}

type ActionOption func(o *MobileAction)

func WithIdentifier(identifier string) ActionOption {
	return func(o *MobileAction) {
		o.Identifier = identifier
	}
}

func WithIndex(index int) ActionOption {
	return func(o *MobileAction) {
		o.Index = index
	}
}

func WithWaitTime(sec float64) ActionOption {
	return func(o *MobileAction) {
		o.WaitTime = sec
	}
}

func WithDuration(duration float64) ActionOption {
	return func(o *MobileAction) {
		o.Duration = duration
	}
}

func WithSteps(steps int) ActionOption {
	return func(o *MobileAction) {
		o.Steps = steps
	}
}

// WithDirection inputs direction (up, down, left, right)
func WithDirection(direction string) ActionOption {
	return func(o *MobileAction) {
		o.Direction = direction
	}
}

// WithCustomDirection inputs sx, sy, ex, ey
func WithCustomDirection(sx, sy, ex, ey float64) ActionOption {
	return func(o *MobileAction) {
		o.Direction = []float64{sx, sy, ex, ey}
	}
}

// WithScope inputs area of [(x1,y1), (x2,y2)]
func WithScope(x1, y1, x2, y2 float64) ActionOption {
	return func(o *MobileAction) {
		o.Scope = []float64{x1, y1, x2, y2}
	}
}

func WithOffset(offsetX, offsetY int) ActionOption {
	return func(o *MobileAction) {
		o.Offset = []int{offsetX, offsetY}
	}
}

func WithText(text string) ActionOption {
	return func(o *MobileAction) {
		o.Text = text
	}
}

func WithID(id string) ActionOption {
	return func(o *MobileAction) {
		o.ID = id
	}
}

func WithDescription(description string) ActionOption {
	return func(o *MobileAction) {
		o.Description = description
	}
}

func WithMaxRetryTimes(maxRetryTimes int) ActionOption {
	return func(o *MobileAction) {
		o.MaxRetryTimes = maxRetryTimes
	}
}

func WithTimeout(timeout int) ActionOption {
	return func(o *MobileAction) {
		o.Timeout = timeout
	}
}

func WithIgnoreNotFoundError(ignoreError bool) ActionOption {
	return func(o *MobileAction) {
		o.IgnoreNotFoundError = ignoreError
	}
}

func (dExt *DriverExt) DoAction(action MobileAction) error {
	log.Info().Str("method", string(action.Method)).Interface("params", action.Params).Msg("start UI action")

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
			return dExt.swipeToTapApp(appName, action)
		}
		return fmt.Errorf("invalid %s params, should be app name(string), got %v",
			ACTION_SwipeToTapApp, action.Params)
	case ACTION_SwipeToTapText:
		if text, ok := action.Params.(string); ok {
			return dExt.swipeToTapTexts([]string{text}, action)
		}
		return fmt.Errorf("invalid %s params, should be app text(string), got %v",
			ACTION_SwipeToTapText, action.Params)
	case ACTION_SwipeToTapTexts:
		if texts, ok := action.Params.([]interface{}); ok {
			var textList []string
			for _, t := range texts {
				textList = append(textList, t.(string))
			}
			action.Params = textList
		}
		if texts, ok := action.Params.([]string); ok {
			return dExt.swipeToTapTexts(texts, action)
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
			return dExt.TapXY(x, y, WithDataIdentifier(action.Identifier))
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
			if len(action.Offset) != 2 {
				action.Offset = []int{0, 0}
			}
			return dExt.TapAbsXY(x, y, WithDataIdentifier(action.Identifier), WithDataOffset(action.Offset[0], action.Offset[1]))
		}
		return fmt.Errorf("invalid %s params: %v", ACTION_TapAbsXY, action.Params)
	case ACTION_Tap:
		if param, ok := action.Params.(string); ok {
			return dExt.Tap(param, WithDataIdentifier(action.Identifier), WithDataIgnoreNotFoundError(true), WithDataIndex(action.Index))
		}
		return fmt.Errorf("invalid %s params: %v", ACTION_Tap, action.Params)
	case ACTION_TapByOCR:
		if ocrText, ok := action.Params.(string); ok {
			if len(action.Scope) != 4 {
				action.Scope = []float64{0, 0, 1, 1}
			}
			if len(action.Offset) != 2 {
				action.Offset = []int{0, 0}
			}

			indexOption := WithDataIndex(action.Index)
			offsetOption := WithDataOffset(action.Offset[0], action.Offset[1])
			scopeOption := WithDataScope(dExt.getAbsScope(action.Scope[0], action.Scope[1], action.Scope[2], action.Scope[3]))
			identifierOption := WithDataIdentifier(action.Identifier)
			IgnoreNotFoundErrorOption := WithDataIgnoreNotFoundError(action.IgnoreNotFoundError)
			return dExt.TapByOCR(ocrText, identifierOption, IgnoreNotFoundErrorOption, indexOption, scopeOption, offsetOption)
		}
		return fmt.Errorf("invalid %s params: %v", ACTION_TapByOCR, action.Params)
	case ACTION_TapByCV:
		if imagePath, ok := action.Params.(string); ok {
			return dExt.TapByCV(imagePath, WithDataIdentifier(action.Identifier), WithDataIgnoreNotFoundError(true), WithDataIndex(action.Index))
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
		swipeAction := dExt.prepareSwipeAction(action)
		return swipeAction(dExt)
	case ACTION_Input:
		// input text on current active element
		// append \n to send text with enter
		// send \b\b\b to delete 3 chars
		param := fmt.Sprintf("%v", action.Params)
		options := []DataOption{}
		if action.Text != "" {
			options = append(options, WithCustomOption("textview", action.Text))
		}
		if action.ID != "" {
			options = append(options, WithCustomOption("id", action.ID))
		}
		if action.Description != "" {
			options = append(options, WithCustomOption("description", action.Description))
		}
		if action.Identifier != "" {
			options = append(options, WithDataIdentifier(action.Identifier))
		}
		return dExt.Driver.Input(param, options...)
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
		params, ok := action.Params.([]interface{})
		if !ok {
			return fmt.Errorf("invalid sleep random params: %v(%T)", action.Params, action.Params)
		}
		return sleepRandom(params)
	case ACTION_ScreenShot:
		// take screenshot
		log.Info().Msg("take screenshot for current screen")
		_, _, err := dExt.TakeScreenShot(builtin.GenNameWithTimestamp("%d_screenshot"))
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

func sleepRandom(params []interface{}) error {
	// append default weight 1
	if len(params) == 2 {
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
			n := s.min + rand.Float64()*(s.max-s.min)
			log.Info().Float64("duration", n).
				Interface("strategy_params", params).Msg("sleep random seconds")
			time.Sleep(time.Duration(n*1000) * time.Millisecond)
			return nil
		}
	}
	return nil
}
