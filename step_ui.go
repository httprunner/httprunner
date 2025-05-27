package hrp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/sdk"
	"github.com/httprunner/httprunner/v5/uixt"
	"github.com/httprunner/httprunner/v5/uixt/option"
)

type MobileUI struct {
	OSType            string `json:"os_type,omitempty" yaml:"os_type,omitempty"` // mobile device os type
	Serial            string `json:"serial,omitempty" yaml:"serial,omitempty"`   // mobile device serial number
	uixt.MobileAction `yaml:",inline"`
	Actions           []uixt.MobileAction `json:"actions,omitempty" yaml:"actions,omitempty"`
}

// StepMobile implements IStep interface.
type StepMobile struct {
	StepConfig
	Mobile  *MobileUI `json:"mobile,omitempty" yaml:"mobile,omitempty"`
	Android *MobileUI `json:"android,omitempty" yaml:"android,omitempty"`
	Harmony *MobileUI `json:"harmony,omitempty" yaml:"harmony,omitempty"`
	IOS     *MobileUI `json:"ios,omitempty" yaml:"ios,omitempty"`
	Browser *MobileUI `json:"browser,omitempty" yaml:"browser,omitempty"`
	cache   *MobileUI // used for caching
}

// uniform interface for all types of mobile systems
func (s *StepMobile) obj() *MobileUI {
	if s.cache != nil {
		return s.cache
	}

	if s.IOS != nil {
		s.cache = s.IOS
		s.cache.OSType = string(StepTypeIOS)
		return s.cache
	} else if s.Harmony != nil {
		s.cache = s.Harmony
		s.cache.OSType = string(StepTypeHarmony)
		return s.cache
	} else if s.Android != nil {
		s.cache = s.Android
		s.cache.OSType = string(StepTypeAndroid)
		return s.cache
	} else if s.Browser != nil {
		s.cache = s.Browser
		s.cache.OSType = string(stepTypeBrowser)
		return s.cache
	} else if s.Mobile != nil {
		s.cache = s.Mobile
		return s.cache
	}

	panic("no mobile device config")
}

func (s *StepMobile) Serial(serial string) *StepMobile {
	s.obj().Serial = serial
	return s
}

func (s *StepMobile) Log(actionName option.ActionName) *StepMobile {
	s.obj().Actions = append(s.obj().Actions, uixt.MobileAction{
		Method: option.ACTION_LOG,
		Params: actionName,
	})
	return s
}

func (s *StepMobile) InstallApp(path string) *StepMobile {
	s.obj().Actions = append(s.obj().Actions, uixt.MobileAction{
		Method: option.ACTION_AppInstall,
		Params: path,
	})
	return s
}

func (s *StepMobile) WebLoginNoneUI(packageName, phoneNumber string, captcha, password string) *StepMobile {
	s.obj().Actions = append(s.obj().Actions, uixt.MobileAction{
		Method: option.ACTION_WebLoginNoneUI,
		Params: []string{packageName, phoneNumber, captcha, password},
	})
	return s
}

func (s *StepMobile) AppLaunch(bundleId string) *StepMobile {
	s.obj().Actions = append(s.obj().Actions, uixt.MobileAction{
		Method: option.ACTION_AppLaunch,
		Params: bundleId,
	})
	return s
}

func (s *StepMobile) AppTerminate(bundleId string) *StepMobile {
	s.obj().Actions = append(s.obj().Actions, uixt.MobileAction{
		Method: option.ACTION_AppTerminate,
		Params: bundleId,
	})
	return s
}

func (s *StepMobile) Home() *StepMobile {
	s.obj().Actions = append(s.obj().Actions, uixt.MobileAction{
		Method: option.ACTION_Home,
		Params: nil,
	})
	return s
}

// TapXY taps the point {X,Y}
// if X<1 & Y<1, {X,Y} will be considered as percentage
// else, X & Y will be considered as absolute coordinates
func (s *StepMobile) TapXY(x, y float64, opts ...option.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  option.ACTION_TapXY,
		Params:  []float64{x, y},
		Options: option.NewActionOptions(opts...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

// TapAbsXY taps the point {X,Y}, X & Y is absolute coordinates
func (s *StepMobile) TapAbsXY(x, y float64, opts ...option.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  option.ACTION_TapAbsXY,
		Params:  []float64{x, y},
		Options: option.NewActionOptions(opts...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

// TapByOCR taps on the target element by OCR recognition
func (s *StepMobile) TapByOCR(ocrText string, opts ...option.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  option.ACTION_TapByOCR,
		Params:  ocrText,
		Options: option.NewActionOptions(opts...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

// TapByCV taps on the target element by CV recognition
func (s *StepMobile) TapByCV(imagePath string, opts ...option.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  option.ACTION_TapByCV,
		Params:  imagePath,
		Options: option.NewActionOptions(opts...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

// TapByUITypes taps on the target element specified by uiTypes, the higher the uiTypes, the higher the priority
func (s *StepMobile) TapByUITypes(opts ...option.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  option.ACTION_TapByCV,
		Options: option.NewActionOptions(opts...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

// AIAction do actions with VLM
func (s *StepMobile) AIAction(prompt string, opts ...option.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  option.ACTION_AIAction,
		Params:  prompt,
		Options: option.NewActionOptions(opts...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

// DoubleTapXY double taps the point {X,Y}, X & Y is percentage of coordinates
func (s *StepMobile) DoubleTapXY(x, y float64, opts ...option.ActionOption) *StepMobile {
	s.obj().Actions = append(s.obj().Actions, uixt.MobileAction{
		Method:  option.ACTION_DoubleTapXY,
		Params:  []float64{x, y},
		Options: option.NewActionOptions(opts...),
	})
	return s
}

func (s *StepMobile) Back() *StepMobile {
	action := uixt.MobileAction{
		Method:  option.ACTION_Back,
		Params:  nil,
		Options: nil,
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

// Swipe drags from [sx, sy] to [ex, ey]
func (s *StepMobile) Swipe(sx, sy, ex, ey float64, opts ...option.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  option.ACTION_SwipeCoordinate,
		Params:  []float64{sx, sy, ex, ey},
		Options: option.NewActionOptions(opts...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

func (s *StepMobile) SwipeUp(opts ...option.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  option.ACTION_SwipeDirection,
		Params:  "up",
		Options: option.NewActionOptions(opts...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

func (s *StepMobile) SwipeDown(opts ...option.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  option.ACTION_SwipeDirection,
		Params:  "down",
		Options: option.NewActionOptions(opts...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

func (s *StepMobile) SwipeLeft(opts ...option.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  option.ACTION_SwipeDirection,
		Params:  "left",
		Options: option.NewActionOptions(opts...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

func (s *StepMobile) SwipeRight(opts ...option.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  option.ACTION_SwipeDirection,
		Params:  "right",
		Options: option.NewActionOptions(opts...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

func (s *StepMobile) SwipeToTapApp(appName string, opts ...option.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  option.ACTION_SwipeToTapApp,
		Params:  appName,
		Options: option.NewActionOptions(opts...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

func (s *StepMobile) SwipeToTapText(text string, opts ...option.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  option.ACTION_SwipeToTapText,
		Params:  text,
		Options: option.NewActionOptions(opts...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

func (s *StepMobile) SwipeToTapTexts(texts interface{}, opts ...option.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  option.ACTION_SwipeToTapTexts,
		Params:  texts,
		Options: option.NewActionOptions(opts...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

func (s *StepMobile) SecondaryClick(x, y float64, options ...option.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  option.ACTION_SecondaryClick,
		Params:  []float64{x, y},
		Options: option.NewActionOptions(options...),
	}
	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

func (s *StepMobile) SecondaryClickBySelector(selector string, options ...option.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  option.ACTION_SecondaryClickBySelector,
		Params:  selector,
		Options: option.NewActionOptions(options...),
	}
	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

func (s *StepMobile) HoverBySelector(selector string, options ...option.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  option.ACTION_HoverBySelector,
		Params:  selector,
		Options: option.NewActionOptions(options...),
	}
	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

func (s *StepMobile) TapBySelector(selector string, options ...option.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  option.ACTION_TapBySelector,
		Params:  selector,
		Options: option.NewActionOptions(options...),
	}
	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

func (s *StepMobile) WebCloseTab(idx int, options ...option.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  option.ACTION_WebCloseTab,
		Params:  idx,
		Options: option.NewActionOptions(options...),
	}
	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

func (s *StepMobile) GetElementTextBySelector(selector string, options ...option.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  option.ACTION_GetElementTextBySelector,
		Params:  selector,
		Options: option.NewActionOptions(options...),
	}
	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

func (s *StepMobile) Input(text string, opts ...option.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  option.ACTION_Input,
		Params:  text,
		Options: option.NewActionOptions(opts...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

// Sleep specify sleep seconds after last action
func (s *StepMobile) Sleep(nSeconds float64, startTime ...time.Time) *StepMobile {
	action := uixt.MobileAction{
		Method:  option.ACTION_Sleep,
		Params:  nSeconds,
		Options: nil,
	}
	if len(startTime) > 0 {
		action.Params = uixt.SleepConfig{
			StartTime: startTime[0],
			Seconds:   nSeconds,
		}
	}
	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

func (s *StepMobile) SleepMS(nMilliseconds int64, startTime ...time.Time) *StepMobile {
	action := uixt.MobileAction{
		Method:  option.ACTION_SleepMS,
		Params:  nMilliseconds,
		Options: nil,
	}
	if len(startTime) > 0 {
		action.Params = uixt.SleepConfig{
			StartTime:    startTime[0],
			Milliseconds: nMilliseconds,
		}
	}
	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

// SleepRandom specify random sleeping seconds after last action
// params have two different kinds:
// 1. [min, max] : min and max are float64 time range boundaries
// 2. [min1, max1, weight1, min2, max2, weight2, ...] : weight is the probability of the time range
func (s *StepMobile) SleepRandom(params ...float64) *StepMobile {
	s.obj().Actions = append(s.obj().Actions, uixt.MobileAction{
		Method:  option.ACTION_SleepRandom,
		Params:  params,
		Options: nil,
	})
	return s
}

func (s *StepMobile) EndToEndDelay(opts ...option.ActionOption) *StepMobile {
	s.obj().Actions = append(s.obj().Actions, uixt.MobileAction{
		Method:  option.ACTION_EndToEndDelay,
		Params:  nil,
		Options: option.NewActionOptions(opts...),
	})
	return s
}

func (s *StepMobile) ScreenShot(opts ...option.ActionOption) *StepMobile {
	s.obj().Actions = append(s.obj().Actions, uixt.MobileAction{
		Method:  option.ACTION_ScreenShot,
		Params:  nil,
		Options: option.NewActionOptions(opts...),
	})
	return s
}

func (s *StepMobile) DisableAutoPopupHandler() *StepMobile {
	s.IgnorePopup = true
	return s
}

func (s *StepMobile) ClosePopups(opts ...option.ActionOption) *StepMobile {
	s.obj().Actions = append(s.obj().Actions, uixt.MobileAction{
		Method:  option.ACTION_ClosePopups,
		Params:  nil,
		Options: option.NewActionOptions(opts...),
	})
	return s
}

// Validate switches to step validation.
func (s *StepMobile) Validate() *StepMobileUIValidation {
	return &StepMobileUIValidation{
		StepMobile: s,
		Validators: make([]interface{}, 0),
	}
}

func (s *StepMobile) Name() string {
	return s.StepName
}

func (s *StepMobile) Type() StepType {
	osType := s.obj().OSType
	if osType != "" {
		return StepType(osType)
	}
	return StepType("mobile")
}

func (s *StepMobile) Config() *StepConfig {
	return &s.StepConfig
}

func (s *StepMobile) Run(r *SessionRunner) (*StepResult, error) {
	return runStepMobileUI(r, s)
}

// StepMobileUIValidation implements IStep interface.
type StepMobileUIValidation struct {
	*StepMobile
	Validators []interface{} `json:"validate,omitempty" yaml:"validate,omitempty"`
}

func (s *StepMobileUIValidation) AssertNameExists(expectedName string, msg ...string) *StepMobileUIValidation {
	v := Validator{
		Check:  option.SelectorName,
		Assert: option.AssertionExists,
		Expect: expectedName,
	}
	if len(msg) > 0 {
		v.Message = msg[0]
	} else {
		v.Message = fmt.Sprintf("attribute name [%s] not found", expectedName)
	}
	s.Validators = append(s.Validators, v)
	return s
}

func (s *StepMobileUIValidation) AssertNameNotExists(expectedName string, msg ...string) *StepMobileUIValidation {
	v := Validator{
		Check:  option.SelectorName,
		Assert: option.AssertionNotExists,
		Expect: expectedName,
	}
	if len(msg) > 0 {
		v.Message = msg[0]
	} else {
		v.Message = fmt.Sprintf("attribute name [%s] should not exist", expectedName)
	}
	s.Validators = append(s.Validators, v)
	return s
}

func (s *StepMobileUIValidation) AssertLabelExists(expectedLabel string, msg ...string) *StepMobileUIValidation {
	v := Validator{
		Check:  option.SelectorLabel,
		Assert: option.AssertionExists,
		Expect: expectedLabel,
	}
	if len(msg) > 0 {
		v.Message = msg[0]
	} else {
		v.Message = fmt.Sprintf("attribute label [%s] not found", expectedLabel)
	}
	s.Validators = append(s.Validators, v)
	return s
}

func (s *StepMobileUIValidation) AssertLabelNotExists(expectedLabel string, msg ...string) *StepMobileUIValidation {
	v := Validator{
		Check:  option.SelectorLabel,
		Assert: option.AssertionNotExists,
		Expect: expectedLabel,
	}
	if len(msg) > 0 {
		v.Message = msg[0]
	} else {
		v.Message = fmt.Sprintf("attribute label [%s] should not exist", expectedLabel)
	}
	s.Validators = append(s.Validators, v)
	return s
}

func (s *StepMobileUIValidation) AssertOCRExists(expectedText string, msg ...string) *StepMobileUIValidation {
	v := Validator{
		Check:  option.SelectorOCR,
		Assert: option.AssertionExists,
		Expect: expectedText,
	}
	if len(msg) > 0 {
		v.Message = msg[0]
	} else {
		v.Message = fmt.Sprintf("ocr text [%s] not found", expectedText)
	}
	s.Validators = append(s.Validators, v)
	return s
}

func (s *StepMobileUIValidation) AssertOCRNotExists(expectedText string, msg ...string) *StepMobileUIValidation {
	v := Validator{
		Check:  option.SelectorOCR,
		Assert: option.AssertionNotExists,
		Expect: expectedText,
	}
	if len(msg) > 0 {
		v.Message = msg[0]
	} else {
		v.Message = fmt.Sprintf("ocr text [%s] should not exist", expectedText)
	}
	s.Validators = append(s.Validators, v)
	return s
}

func (s *StepMobileUIValidation) AssertImageExists(expectedImagePath string, msg ...string) *StepMobileUIValidation {
	v := Validator{
		Check:  option.SelectorImage,
		Assert: option.AssertionExists,
		Expect: expectedImagePath,
	}
	if len(msg) > 0 {
		v.Message = msg[0]
	} else {
		v.Message = fmt.Sprintf("cv image [%s] not found", expectedImagePath)
	}
	s.Validators = append(s.Validators, v)
	return s
}

func (s *StepMobileUIValidation) AssertImageNotExists(expectedImagePath string, msg ...string) *StepMobileUIValidation {
	v := Validator{
		Check:  option.SelectorImage,
		Assert: option.AssertionNotExists,
		Expect: expectedImagePath,
	}
	if len(msg) > 0 {
		v.Message = msg[0]
	} else {
		v.Message = fmt.Sprintf("cv image [%s] should not exist", expectedImagePath)
	}
	s.Validators = append(s.Validators, v)
	return s
}

func (s *StepMobileUIValidation) AssertAI(prompt string, msg ...string) *StepMobileUIValidation {
	v := Validator{
		Check:  option.SelectorAI,
		Assert: option.AssertionAI,
		Expect: prompt,
	}
	if len(msg) > 0 {
		v.Message = msg[0]
	} else {
		v.Message = fmt.Sprintf("assert ai prompt [%s] failed", prompt)
	}
	s.Validators = append(s.Validators, v)
	return s
}

func (s *StepMobileUIValidation) AssertAppInForeground(packageName string, msg ...string) *StepMobileUIValidation {
	v := Validator{
		Check:  option.SelectorForegroundApp,
		Assert: option.AssertionEqual,
		Expect: packageName,
	}
	if len(msg) > 0 {
		v.Message = msg[0]
	} else {
		v.Message = fmt.Sprintf("app [%s] should be in foreground", packageName)
	}
	s.Validators = append(s.Validators, v)
	return s
}

func (s *StepMobileUIValidation) AssertAppNotInForeground(packageName string, msg ...string) *StepMobileUIValidation {
	v := Validator{
		Check:  option.SelectorForegroundApp,
		Assert: option.AssertionNotEqual,
		Expect: packageName,
	}
	if len(msg) > 0 {
		v.Message = msg[0]
	} else {
		v.Message = fmt.Sprintf("app [%s] should not be in foreground", packageName)
	}
	s.Validators = append(s.Validators, v)
	return s
}

func (s *StepMobileUIValidation) Name() string {
	return s.StepName
}

func (s *StepMobileUIValidation) Type() StepType {
	return s.StepMobile.Type() + stepTypeSuffixValidation
}

func (s *StepMobileUIValidation) Config() *StepConfig {
	return &StepConfig{
		StepName:   s.StepName,
		Variables:  s.Variables,
		Validators: s.Validators,
	}
}

func (s *StepMobileUIValidation) Run(r *SessionRunner) (*StepResult, error) {
	return runStepMobileUI(r, s)
}

func runStepMobileUI(s *SessionRunner, step IStep) (stepResult *StepResult, err error) {
	var stepVariables map[string]interface{}
	var stepValidators []interface{}
	var ignorePopup bool

	var mobileStep *MobileUI
	switch stepMobile := step.(type) {
	case *StepMobile:
		mobileStep = stepMobile.obj()
		stepVariables = stepMobile.Variables
		ignorePopup = stepMobile.IgnorePopup
	case *StepMobileUIValidation:
		mobileStep = stepMobile.obj()
		stepVariables = stepMobile.Variables
		stepValidators = stepMobile.Validators
		ignorePopup = stepMobile.StepMobile.IgnorePopup
	default:
		return nil, errors.New("invalid mobile UI step type")
	}

	// report GA event
	go sdk.SendGA4Event("hrp_run_ui", map[string]interface{}{
		"osType": mobileStep.OSType,
	})

	// init wda/uia/hdc driver
	config := uixt.DriverCacheConfig{
		Platform: mobileStep.OSType,
		Serial:   mobileStep.Serial,
	}
	uiDriver, err := uixt.GetOrCreateXTDriver(config)
	if err != nil {
		return
	}

	identifier := mobileStep.Identifier
	if mobileStep.Options != nil && identifier == "" {
		identifier = mobileStep.Options.Identifier
	}
	if len(mobileStep.Actions) != 0 && identifier == "" {
		for _, action := range mobileStep.Actions {
			if action.Identifier != "" {
				identifier = action.Identifier
				break
			}
		}
	}

	start := time.Now()
	stepResult = &StepResult{
		Name:        step.Name(),
		Identifier:  identifier,
		StepType:    step.Type(),
		Success:     false,
		ContentSize: 0,
		StartTime:   start.Unix(),
	}

	defer func() {
		attachments := uixt.Attachments{}
		if err != nil {
			attachments["error"] = err.Error()

			// save foreground app
			startTime := time.Now()
			actionResult := &ActionResult{
				MobileAction: uixt.MobileAction{
					Method: option.ACTION_GetForegroundApp,
					Params: "[ForDebug] check foreground app",
				},
				StartTime: startTime.Unix(),
			}
			if app, err1 := uiDriver.ForegroundInfo(); err1 == nil {
				attachments["foreground_app"] = app.AppBaseInfo
			} else {
				log.Warn().Err(err1).Msg("save foreground app failed, ignore")
			}
			actionResult.Elapsed = time.Since(startTime).Milliseconds()
			stepResult.Actions = append(stepResult.Actions, actionResult)
		}

		// automatic handling of pop-up windows on each step finished
		if !ignorePopup && !s.caseRunner.Config.Get().IgnorePopup {
			startTime := time.Now()
			actionResult := &ActionResult{
				MobileAction: uixt.MobileAction{
					Method: option.ACTION_ClosePopups,
					Params: "[ForDebug] close popups handler",
				},
				StartTime: startTime.Unix(),
			}
			if err2 := uiDriver.ClosePopupsHandler(); err2 != nil {
				log.Error().Err(err2).Str("step", step.Name()).Msg("auto handle popup failed")
			}
			actionResult.Elapsed = time.Since(startTime).Milliseconds()
			stepResult.Actions = append(stepResult.Actions, actionResult)
		}

		// save attachments
		for key, value := range uiDriver.GetData(true) {
			attachments[key] = value
		}
		stepResult.Attachments = attachments
		stepResult.Elapsed = time.Since(start).Milliseconds()
	}()

	// run actions
	for _, action := range mobileStep.Actions {
		select {
		case <-s.caseRunner.hrpRunner.caseTimeoutTimer.C:
			log.Warn().Msg("timeout in mobile UI runner")
			return stepResult, errors.Wrap(code.TimeoutError, "mobile UI runner timeout")
		case <-s.caseRunner.hrpRunner.interruptSignal:
			log.Warn().Msg("interrupted in mobile UI runner")
			return stepResult, errors.Wrap(code.InterruptError, "mobile UI runner interrupted")
		default:
			actionStartTime := time.Now()
			actionResult := &ActionResult{
				MobileAction: action,
				StartTime:    actionStartTime.Unix(), // action 开始时间
			}
			if action.Params, err = s.caseRunner.parser.Parse(action.Params, stepVariables); err != nil {
				if !code.IsErrorPredefined(err) {
					err = errors.Wrap(code.ParseError,
						fmt.Sprintf("parse action params failed: %v", err))
				}
				return stepResult, err
			}

			// stat uixt action
			if action.Method == option.ACTION_LOG {
				log.Info().Interface("action", action.Params).Msg("stat uixt action")
				actionMethod := option.ActionName(action.Params.(string))
				s.summary.Stat.Actions[actionMethod]++
				continue
			}

			err = uiDriver.ExecuteAction(context.Background(), action)
			actionResult.Elapsed = time.Since(actionStartTime).Milliseconds()
			stepResult.Actions = append(stepResult.Actions, actionResult)
			if err != nil {
				if !code.IsErrorPredefined(err) {
					err = errors.Wrap(code.MobileUIDriverError, err.Error())
				}
				return stepResult, err
			}
		}
	}

	// validate
	validateResults, err := validateUI(uiDriver, stepValidators)
	if err != nil {
		if !code.IsErrorPredefined(err) {
			err = errors.Wrap(code.MobileUIValidationError, err.Error())
		}
		return
	}
	if len(validateResults) > 0 {
		sessionData := &SessionData{
			Validators: validateResults,
		}
		stepResult.Data = sessionData
	}
	stepResult.Success = true
	return stepResult, nil
}

func validateUI(ud *uixt.XTDriver, iValidators []interface{}) (validateResults []*ValidationResult, err error) {
	for _, iValidator := range iValidators {
		validator, ok := iValidator.(Validator)
		if !ok {
			return nil, errors.New("validator type error")
		}

		validataResult := &ValidationResult{
			Validator:   validator,
			CheckResult: "fail",
		}

		// parse check value
		if !strings.HasPrefix(validator.Check, "ui_") {
			validataResult.CheckResult = "skip"
			log.Warn().Interface("validator", validator).Msg("skip validator")
			validateResults = append(validateResults, validataResult)
			continue
		}

		expected, ok := validator.Expect.(string)
		if !ok {
			return nil, errors.New("validator expect should be string")
		}

		err := ud.DoValidation(validator.Check, validator.Assert, expected, validator.Message)
		if err != nil {
			return validateResults, errors.Wrap(err, "step validation failed")
		}

		validataResult.CheckResult = "pass"
		validateResults = append(validateResults, validataResult)
	}
	return validateResults, nil
}
