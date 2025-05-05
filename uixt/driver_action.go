package uixt

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/uixt/option"
)

type ActionMethod string

const (
	ACTION_LOG              ActionMethod = "log"
	ACTION_AppInstall       ActionMethod = "install"
	ACTION_AppUninstall     ActionMethod = "uninstall"
	ACTION_LoginNoneUI      ActionMethod = "login_none_ui"
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
	ACTION_Home                     ActionMethod = "home"
	ACTION_TapXY                    ActionMethod = "tap_xy"
	ACTION_TapAbsXY                 ActionMethod = "tap_abs_xy"
	ACTION_TapByOCR                 ActionMethod = "tap_ocr"
	ACTION_TapByCV                  ActionMethod = "tap_cv"
	ACTION_DoubleTapXY              ActionMethod = "double_tap_xy"
	ACTION_Swipe                    ActionMethod = "swipe"
	ACTION_Drag                     ActionMethod = "drag"
	ACTION_Input                    ActionMethod = "input"
	ACTION_Back                     ActionMethod = "back"
	ACTION_KeyCode                  ActionMethod = "keycode"
	ACTION_AIAction                 ActionMethod = "ai_action" // action with ai
	ACTION_TapBySelector            ActionMethod = "tap_by_selector"
	ACTION_HoverBySelector          ActionMethod = "hover_by_selector"
	ACTION_ClosePage                ActionMethod = "close_page"
	ACTION_RightClick               ActionMethod = "right_click"
	ACTION_RightClickBySelector     ActionMethod = "right_click_by_selector"
	ACTION_GetElementTextBySelector ActionMethod = "get_element_text_by_selector"

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
	SelectorAI            string = "ui_ai" // ui query with ai
	SelectorForegroundApp string = "ui_foreground_app"
	// assertions
	AssertionEqual     string = "equal"
	AssertionNotEqual  string = "not_equal"
	AssertionExists    string = "exists"
	AssertionNotExists string = "not_exists"
	AssertionAI        string = "ai_assert" // assert with ai
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
	case ACTION_LoginNoneUI:
		if len(action.Params.([]interface{})) == 4 {
			params := action.Params.([]interface{})
			_, err = dExt.IDriver.(*BrowserDriver).LoginNoneUI(params[0].(string), params[1].(string), params[2].(string), params[3].(string))
			return err
		}
		return fmt.Errorf("invalid %s params: %v", ACTION_LoginNoneUI, action.Params)
	case ACTION_AppInstall:
		if app, ok := action.Params.(string); ok {
			if err = dExt.GetDevice().Install(app,
				option.WithRetryTimes(action.MaxRetryTimes)); err != nil {
				return errors.Wrap(err, "failed to install app")
			}
		}
	case ACTION_AppUninstall:
		if packageName, ok := action.Params.(string); ok {
			if err = dExt.GetDevice().Uninstall(packageName); err != nil {
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
			return dExt.SwipeToTapTexts([]string{text}, action.GetOptions()...)
		}
		return fmt.Errorf("invalid %s params, should be app text(string), got %v",
			ACTION_SwipeToTapText, action.Params)
	case ACTION_SwipeToTapTexts:
		if texts, ok := action.Params.([]string); ok {
			return dExt.SwipeToTapTexts(texts, action.GetOptions()...)
		}
		if texts, err := builtin.ConvertToStringSlice(action.Params); err == nil {
			return dExt.SwipeToTapTexts(texts, action.GetOptions()...)
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
		return dExt.Home()
	case ACTION_RightClick:
		if params, err := builtin.ConvertToFloat64Slice(action.Params); err == nil {
			if len(params) != 2 {
				return fmt.Errorf("invalid tap location params: %v", params)
			}
			x, y := params[0], params[1]
			return dExt.IDriver.(*BrowserDriver).RightClick(x, y)
		}
		return fmt.Errorf("invalid %s params: %v", ACTION_RightClick, action.Params)
	case ACTION_HoverBySelector:
		if selector, ok := action.Params.(string); ok {
			return dExt.IDriver.(*BrowserDriver).HoverBySelector(selector, action.GetOptions()...)
		}
		return fmt.Errorf("invalid %s params: %v", ACTION_HoverBySelector, action.Params)
	case ACTION_TapBySelector:
		if selector, ok := action.Params.(string); ok {
			return dExt.IDriver.(*BrowserDriver).TapBySelector(selector, action.GetOptions()...)
		}
		return fmt.Errorf("invalid %s params: %v", ACTION_TapBySelector, action.Params)
	case ACTION_RightClickBySelector:
		if selector, ok := action.Params.(string); ok {
			return dExt.IDriver.(*BrowserDriver).RightClickBySelector(selector, action.GetOptions()...)
		}
		return fmt.Errorf("invalid %s params: %v", ACTION_RightClickBySelector, action.Params)
	case ACTION_ClosePage:
		if param, ok := action.Params.(json.Number); ok {
			paramInt64, _ := param.Int64()
			return dExt.IDriver.(*BrowserDriver).ClosePage(int(paramInt64))
		} else if param, ok := action.Params.(int64); ok {
			return dExt.IDriver.(*BrowserDriver).ClosePage(int(param))
		} else {
			return dExt.IDriver.(*BrowserDriver).ClosePage(action.Params.(int))
		}
		// return fmt.Errorf("invalid %s params: %v", ACTION_ClosePage, action.Params)
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
			_, err = dExt.Source(option.WithProcessName(packageName))
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
	case ACTION_TapByOCR:
		if ocrText, ok := action.Params.(string); ok {
			return dExt.TapByOCR(ocrText, action.GetOptions()...)
		}
		return fmt.Errorf("invalid %s params: %v", ACTION_TapByOCR, action.Params)
	case ACTION_TapByCV:
		actionOptions := option.NewActionOptions(action.GetOptions()...)
		if len(actionOptions.ScreenShotWithUITypes) > 0 {
			return dExt.TapByCV(action.GetOptions()...)
		}
		return fmt.Errorf("invalid %s params: %v", ACTION_TapByCV, action.Params)
	case ACTION_DoubleTapXY:
		if params, err := builtin.ConvertToFloat64Slice(action.Params); err == nil {
			// relative x,y of window size: [0.5, 0.5]
			if len(params) != 2 {
				return fmt.Errorf("invalid tap location params: %v", params)
			}
			x, y := params[0], params[1]
			return dExt.DoubleTap(x, y)
		}
		return fmt.Errorf("invalid %s params: %v", ACTION_DoubleTapXY, action.Params)
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
		return dExt.Back()
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
		_, err := dExt.GetScreenResult(action.GetScreenShotOptions()...)
		return err
	case ACTION_ClosePopups:
		return dExt.ClosePopupsHandler()
	case ACTION_CallFunction:
		fn := action.Fn
		fn()
		return nil
	case ACTION_AIAction:
		if prompt, ok := action.Params.(string); ok {
			return dExt.AIAction(prompt, action.GetOptions()...)
		}
		return fmt.Errorf("invalid %s params: %v", ACTION_AIAction, action.Params)
	default:
		log.Warn().Str("action", string(action.Method)).Msg("action not implemented")
		return errors.Wrapf(code.InvalidCaseError,
			"UI action %v not implemented", action.Method)
	}
	return nil
}
