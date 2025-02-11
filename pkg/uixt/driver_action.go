package uixt

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
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
	ACTION_SetIme           ActionMethod = "set_ime"
	ACTION_GetSource        ActionMethod = "get_source"
	ACTION_GetForegroundApp ActionMethod = "get_foreground_app"
	ACTION_CallFunction     ActionMethod = "call_function"

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
	Method  ActionMethod          `json:"method,omitempty" yaml:"method,omitempty"`
	Params  interface{}           `json:"params,omitempty" yaml:"params,omitempty"`
	Fn      func()                `json:"-" yaml:"-"` // only used for function action, not serialized
	Options *option.ActionOptions `json:"options,omitempty" yaml:"options,omitempty"`
	option.ActionOptions
}

func (ma MobileAction) GetOptions() []option.ActionOption {
	var actionOptionList []option.ActionOption
	// Notice: merge options from ma.Options and ma.ActionOptions
	if ma.Options != nil {
		actionOptionList = append(actionOptionList, ma.Options.Options()...)
	}
	actionOptionList = append(actionOptionList, ma.ActionOptions.Options()...)
	return actionOptionList
}

type TapTextAction struct {
	Text    string
	Options []option.ActionOption
}

func (dExt *XTDriver) ParseActionOptions(opts ...option.ActionOption) []option.ActionOption {
	actionOptions := option.NewActionOptions(opts...)

	// convert relative scope to absolute scope
	if len(actionOptions.AbsScope) != 4 && len(actionOptions.Scope) == 4 {
		scope := actionOptions.Scope
		actionOptions.AbsScope = dExt.GenAbsScope(
			scope[0], scope[1], scope[2], scope[3])
	}

	return actionOptions.Options()
}

func (dExt *XTDriver) GenAbsScope(x1, y1, x2, y2 float64) option.AbsScope {
	// convert relative scope to absolute scope
	windowSize, _ := dExt.WindowSize()
	absX1 := int(x1 * float64(windowSize.Width))
	absY1 := int(y1 * float64(windowSize.Height))
	absX2 := int(x2 * float64(windowSize.Width))
	absY2 := int(y2 * float64(windowSize.Height))
	return option.AbsScope{absX1, absY1, absX2, absY2}
}

func (dExt *XTDriver) DoAction(action MobileAction) (err error) {
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
			if err = dExt.InstallByUrl(appUrl,
				option.WithRetryTimes(action.MaxRetryTimes)); err != nil {
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
			if err = dExt.AppClear(packageName); err != nil {
				return errors.Wrap(err, "failed to clear app")
			}
		}
	case ACTION_AppLaunch:
		if bundleId, ok := action.Params.(string); ok {
			return dExt.AppLaunch(bundleId)
		}
		return fmt.Errorf("invalid %s params, should be bundleId(string), got %v",
			ACTION_AppLaunch, action.Params)
	case ACTION_SwipeToTapApp:
		if appName, ok := action.Params.(string); ok {
			return dExt.SwipeToTapApp(appName, action.GetOptions()...)
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
			success, err := dExt.AppTerminate(bundleId)
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
		return dExt.Homescreen()
	case ACTION_SetIme:
		if ime, ok := action.Params.(string); ok {
			err = dExt.SetIme(ime)
			if err != nil {
				return errors.Wrap(err, "failed to set ime")
			}
			return nil
		}
	case ACTION_GetSource:
		if packageName, ok := action.Params.(string); ok {
			source := option.NewSourceOption().WithProcessName(packageName)
			_, err = dExt.Source(source)
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
		actionOptions := option.NewActionOptions(action.GetOptions()...)
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
		swipeAction := prepareSwipeAction(dExt, params, action.GetOptions()...)
		return swipeAction(dExt)
	case ACTION_Input:
		// input text on current active element
		// append \n to send text with enter
		// send \b\b\b to delete 3 chars
		param := fmt.Sprintf("%v", action.Params)
		return dExt.Input(param)
	case ACTION_Back:
		return dExt.PressBack()
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
		_, err := dExt.GetScreenResult(action.GetScreenOptions()...)
		return err
	case ACTION_ClosePopups:
		return dExt.ClosePopupsHandler()
	case ACTION_CallFunction:
		fn := action.Fn
		fn()
		return nil
	default:
		log.Warn().Str("action", string(action.Method)).Msg("action not implemented")
		return errors.Wrapf(code.InvalidCaseError,
			"UI action %v not implemented", action.Method)
	}
	return nil
}

type SleepConfig struct {
	StartTime    time.Time `json:"start_time"`
	Seconds      float64   `json:"seconds,omitempty"`
	Milliseconds int64     `json:"milliseconds,omitempty"`
}

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
