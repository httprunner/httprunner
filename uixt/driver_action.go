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

type MobileAction struct {
	Method  option.ActionMethod   `json:"method,omitempty" yaml:"method,omitempty"`
	Params  interface{}           `json:"params,omitempty" yaml:"params,omitempty"`
	Fn      func()                `json:"-" yaml:"-"` // used for function action, not serialized
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

// TODO: merge to uixt MCP Server
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
	case option.ACTION_WebLoginNoneUI:
		if len(action.Params.([]interface{})) == 4 {
			driver, ok := dExt.IDriver.(*BrowserDriver)
			if !ok {
				return errors.New("invalid browser driver")
			}
			params := action.Params.([]interface{})
			_, err = driver.LoginNoneUI(params[0].(string), params[1].(string), params[2].(string), params[3].(string))
			return err
		}
		return fmt.Errorf("invalid %s params: %v", option.ACTION_WebLoginNoneUI, action.Params)
	case option.ACTION_AppInstall:
		if app, ok := action.Params.(string); ok {
			if err = dExt.GetDevice().Install(app,
				option.WithRetryTimes(action.MaxRetryTimes)); err != nil {
				return errors.Wrap(err, "failed to install app")
			}
		}
	case option.ACTION_AppUninstall:
		if packageName, ok := action.Params.(string); ok {
			if err = dExt.GetDevice().Uninstall(packageName); err != nil {
				return errors.Wrap(err, "failed to uninstall app")
			}
		}
	case option.ACTION_AppClear:
		if packageName, ok := action.Params.(string); ok {
			if err = dExt.AppClear(packageName); err != nil {
				return errors.Wrap(err, "failed to clear app")
			}
		}
	case option.ACTION_AppLaunch:
		if bundleId, ok := action.Params.(string); ok {
			return dExt.AppLaunch(bundleId)
		}
		return fmt.Errorf("invalid %s params, should be bundleId(string), got %v",
			option.ACTION_AppLaunch, action.Params)
	case option.ACTION_SwipeToTapApp:
		if appName, ok := action.Params.(string); ok {
			return dExt.SwipeToTapApp(appName, action.GetOptions()...)
		}
		return fmt.Errorf("invalid %s params, should be app name(string), got %v",
			option.ACTION_SwipeToTapApp, action.Params)
	case option.ACTION_SwipeToTapText:
		if text, ok := action.Params.(string); ok {
			return dExt.SwipeToTapTexts([]string{text}, action.GetOptions()...)
		}
		return fmt.Errorf("invalid %s params, should be app text(string), got %v",
			option.ACTION_SwipeToTapText, action.Params)
	case option.ACTION_SwipeToTapTexts:
		if texts, ok := action.Params.([]string); ok {
			return dExt.SwipeToTapTexts(texts, action.GetOptions()...)
		}
		if texts, err := builtin.ConvertToStringSlice(action.Params); err == nil {
			return dExt.SwipeToTapTexts(texts, action.GetOptions()...)
		}
		return fmt.Errorf("invalid %s params: %v", option.ACTION_SwipeToTapTexts, action.Params)
	case option.ACTION_AppTerminate:
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
	case option.ACTION_Home:
		return dExt.Home()
	case option.ACTION_SecondaryClick:
		if params, err := builtin.ConvertToFloat64Slice(action.Params); err == nil {
			if len(params) != 2 {
				return fmt.Errorf("invalid tap location params: %v", params)
			}
			x, y := params[0], params[1]
			return dExt.SecondaryClick(x, y)
		}
		return fmt.Errorf("invalid %s params: %v", option.ACTION_SecondaryClick, action.Params)
	case option.ACTION_HoverBySelector:
		if selector, ok := action.Params.(string); ok {
			return dExt.HoverBySelector(selector, action.GetOptions()...)
		}
		return fmt.Errorf("invalid %s params: %v", option.ACTION_HoverBySelector, action.Params)
	case option.ACTION_TapBySelector:
		if selector, ok := action.Params.(string); ok {
			return dExt.TapBySelector(selector, action.GetOptions()...)
		}
		return fmt.Errorf("invalid %s params: %v", option.ACTION_TapBySelector, action.Params)
	case option.ACTION_SecondaryClickBySelector:
		if selector, ok := action.Params.(string); ok {
			return dExt.SecondaryClickBySelector(selector, action.GetOptions()...)
		}
		return fmt.Errorf("invalid %s params: %v", option.ACTION_SecondaryClickBySelector, action.Params)
	case option.ACTION_WebCloseTab:
		if param, ok := action.Params.(json.Number); ok {
			paramInt64, _ := param.Int64()
			return dExt.IDriver.(*BrowserDriver).CloseTab(int(paramInt64))
		} else if param, ok := action.Params.(int64); ok {
			return dExt.IDriver.(*BrowserDriver).CloseTab(int(param))
		} else {
			return dExt.IDriver.(*BrowserDriver).CloseTab(action.Params.(int))
		}
		// return fmt.Errorf("invalid %s params: %v", ACTION_WebCloseTab, action.Params)
	case option.ACTION_SetIme:
		if ime, ok := action.Params.(string); ok {
			err = dExt.SetIme(ime)
			if err != nil {
				return errors.Wrap(err, "failed to set ime")
			}
			return nil
		}
	case option.ACTION_GetSource:
		if packageName, ok := action.Params.(string); ok {
			_, err = dExt.Source(option.WithProcessName(packageName))
			if err != nil {
				return errors.Wrap(err, "failed to set ime")
			}
			return nil
		}
	case option.ACTION_TapXY:
		if params, err := builtin.ConvertToFloat64Slice(action.Params); err == nil {
			// relative x,y of window size: [0.5, 0.5]
			if len(params) != 2 {
				return fmt.Errorf("invalid tap location params: %v", params)
			}
			x, y := params[0], params[1]
			return dExt.TapXY(x, y, action.GetOptions()...)
		}
		return fmt.Errorf("invalid %s params: %v", option.ACTION_TapXY, action.Params)
	case option.ACTION_TapAbsXY:
		if params, err := builtin.ConvertToFloat64Slice(action.Params); err == nil {
			// absolute coordinates x,y of window size: [100, 300]
			if len(params) != 2 {
				return fmt.Errorf("invalid tap location params: %v", params)
			}
			x, y := params[0], params[1]
			return dExt.TapAbsXY(x, y, action.GetOptions()...)
		}
		return fmt.Errorf("invalid %s params: %v", option.ACTION_TapAbsXY, action.Params)
	case option.ACTION_TapByOCR:
		if ocrText, ok := action.Params.(string); ok {
			return dExt.TapByOCR(ocrText, action.GetOptions()...)
		}
		return fmt.Errorf("invalid %s params: %v", option.ACTION_TapByOCR, action.Params)
	case option.ACTION_TapByCV:
		actionOptions := option.NewActionOptions(action.GetOptions()...)
		if len(actionOptions.ScreenShotWithUITypes) > 0 {
			return dExt.TapByCV(action.GetOptions()...)
		}
		return fmt.Errorf("invalid %s params: %v", option.ACTION_TapByCV, action.Params)
	case option.ACTION_DoubleTapXY:
		if params, err := builtin.ConvertToFloat64Slice(action.Params); err == nil {
			// relative x,y of window size: [0.5, 0.5]
			if len(params) != 2 {
				return fmt.Errorf("invalid tap location params: %v", params)
			}
			x, y := params[0], params[1]
			return dExt.DoubleTap(x, y)
		}
		return fmt.Errorf("invalid %s params: %v", option.ACTION_DoubleTapXY, action.Params)
	case option.ACTION_Swipe:
		params := action.Params
		swipeAction := prepareSwipeAction(dExt, params, action.GetOptions()...)
		return swipeAction(dExt)
	case option.ACTION_Input:
		// input text on current active element
		// append \n to send text with enter
		// send \b\b\b to delete 3 chars
		param := fmt.Sprintf("%v", action.Params)
		return dExt.Input(param)
	case option.ACTION_Back:
		return dExt.Back()
	case option.ACTION_Sleep:
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
		} else if param, ok := action.Params.(string); ok {
			seconds, err := builtin.ConvertToFloat64(param)
			if err != nil {
				return errors.Wrapf(err, "invalid sleep params: %v(%T)", action.Params, action.Params)
			}
			time.Sleep(time.Duration(seconds*1000) * time.Millisecond)
			return nil
		}
		return fmt.Errorf("invalid sleep params: %v(%T)", action.Params, action.Params)
	case option.ACTION_SleepMS:
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
	case option.ACTION_SleepRandom:
		if params, err := builtin.ConvertToFloat64Slice(action.Params); err == nil {
			sleepStrict(time.Now(), getSimulationDuration(params))
			return nil
		}
		return fmt.Errorf("invalid sleep random params: %v(%T)", action.Params, action.Params)
	case option.ACTION_ScreenShot:
		// take screenshot
		log.Info().Msg("take screenshot for current screen")
		_, err := dExt.GetScreenResult(action.GetScreenShotOptions()...)
		return err
	case option.ACTION_ClosePopups:
		return dExt.ClosePopupsHandler()
	case option.ACTION_CallFunction:
		if funcDesc, ok := action.Params.(string); ok {
			return dExt.Call(funcDesc, action.Fn, action.GetOptions()...)
		}
		return fmt.Errorf("invalid function description: %v", action.Params)
	case option.ACTION_AIAction:
		if prompt, ok := action.Params.(string); ok {
			return dExt.AIAction(prompt, action.GetOptions()...)
		}
		return fmt.Errorf("invalid %s params: %v", option.ACTION_AIAction, action.Params)
	default:
		log.Warn().Str("action", string(action.Method)).Msg("action not implemented")
		return errors.Wrapf(code.InvalidCaseError,
			"UI action %v not implemented", action.Method)
	}
	return nil
}
