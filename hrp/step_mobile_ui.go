package hrp

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/code"
	"github.com/httprunner/httprunner/v4/hrp/internal/sdk"
	"github.com/httprunner/httprunner/v4/hrp/pkg/uixt"
)

type MobileUI struct {
	OSType            string `json:"os_type,omitempty" yaml:"os_type,omitempty"` // ios or harmony or android
	Serial            string `json:"serial,omitempty" yaml:"serial,omitempty"`   // android serial or ios udid
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

	cache *MobileUI // used for caching
}

// uniform interface for all types of mobile systems
func (s *StepMobile) obj() *MobileUI {
	if s.cache != nil {
		return s.cache
	}

	if s.IOS != nil {
		s.cache = s.IOS
		s.cache.OSType = string(stepTypeIOS)
		return s.cache
	} else if s.Harmony != nil {
		s.cache = s.Harmony
		s.cache.OSType = string(stepTypeHarmony)
		return s.cache
	} else if s.Android != nil {
		s.cache = s.Android
		s.cache.OSType = string(stepTypeAndroid)
		return s.cache
	} else if s.Mobile != nil {
		s.cache = s.Mobile
		return s.cache
	}

	panic("no mobile device config")
}

func (s *StepMobile) OSType(ostype string) *StepMobile {
	s.obj().OSType = ostype
	return s
}

func (s *StepMobile) Serial(serial string) *StepMobile {
	s.obj().Serial = serial
	return s
}

func (s *StepMobile) Log(actionName uixt.ActionMethod) *StepMobile {
	s.obj().Actions = append(s.obj().Actions, uixt.MobileAction{
		Method: uixt.ACTION_LOG,
		Params: actionName,
	})
	return s
}

func (s *StepMobile) InstallApp(path string) *StepMobile {
	s.obj().Actions = append(s.obj().Actions, uixt.MobileAction{
		Method: uixt.ACTION_AppInstall,
		Params: path,
	})
	return s
}

func (s *StepMobile) AppLaunch(bundleId string) *StepMobile {
	s.obj().Actions = append(s.obj().Actions, uixt.MobileAction{
		Method: uixt.ACTION_AppLaunch,
		Params: bundleId,
	})
	return s
}

func (s *StepMobile) AppTerminate(bundleId string) *StepMobile {
	s.obj().Actions = append(s.obj().Actions, uixt.MobileAction{
		Method: uixt.ACTION_AppTerminate,
		Params: bundleId,
	})
	return s
}

func (s *StepMobile) Home() *StepMobile {
	s.obj().Actions = append(s.obj().Actions, uixt.MobileAction{
		Method: uixt.ACTION_Home,
		Params: nil,
	})
	return s
}

// TapXY taps the point {X,Y}, X & Y is percentage of coordinates
func (s *StepMobile) TapXY(x, y float64, options ...uixt.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  uixt.ACTION_TapXY,
		Params:  []float64{x, y},
		Options: uixt.NewActionOptions(options...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

// TapAbsXY taps the point {X,Y}, X & Y is absolute coordinates
func (s *StepMobile) TapAbsXY(x, y float64, options ...uixt.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  uixt.ACTION_TapAbsXY,
		Params:  []float64{x, y},
		Options: uixt.NewActionOptions(options...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

// Tap taps on the target element
func (s *StepMobile) Tap(params string, options ...uixt.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  uixt.ACTION_Tap,
		Params:  params,
		Options: uixt.NewActionOptions(options...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

// TapByOCR taps on the target element by OCR recognition
func (s *StepMobile) TapByOCR(ocrText string, options ...uixt.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  uixt.ACTION_TapByOCR,
		Params:  ocrText,
		Options: uixt.NewActionOptions(options...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

// TapByCV taps on the target element by CV recognition
func (s *StepMobile) TapByCV(imagePath string, options ...uixt.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  uixt.ACTION_TapByCV,
		Params:  imagePath,
		Options: uixt.NewActionOptions(options...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

// TapByUITypes taps on the target element specified by uiTypes, the higher the uiTypes, the higher the priority
func (s *StepMobile) TapByUITypes(options ...uixt.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  uixt.ACTION_TapByCV,
		Options: uixt.NewActionOptions(options...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

// DoubleTapXY double taps the point {X,Y}, X & Y is percentage of coordinates
func (s *StepMobile) DoubleTapXY(x, y float64, options ...uixt.ActionOption) *StepMobile {
	s.obj().Actions = append(s.obj().Actions, uixt.MobileAction{
		Method:  uixt.ACTION_DoubleTapXY,
		Params:  []float64{x, y},
		Options: uixt.NewActionOptions(options...),
	})
	return s
}

func (s *StepMobile) DoubleTap(params string, options ...uixt.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  uixt.ACTION_DoubleTap,
		Params:  params,
		Options: uixt.NewActionOptions(options...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

func (s *StepMobile) Back(options ...uixt.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  uixt.ACTION_Back,
		Params:  nil,
		Options: uixt.NewActionOptions(options...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

// Swipe drags from [sx, sy] to [ex, ey]
func (s *StepMobile) Swipe(sx, sy, ex, ey float64, options ...uixt.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  uixt.ACTION_Swipe,
		Params:  []float64{sx, sy, ex, ey},
		Options: uixt.NewActionOptions(options...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

func (s *StepMobile) SwipeUp(options ...uixt.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  uixt.ACTION_Swipe,
		Params:  "up",
		Options: uixt.NewActionOptions(options...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

func (s *StepMobile) SwipeDown(options ...uixt.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  uixt.ACTION_Swipe,
		Params:  "down",
		Options: uixt.NewActionOptions(options...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

func (s *StepMobile) SwipeLeft(options ...uixt.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  uixt.ACTION_Swipe,
		Params:  "left",
		Options: uixt.NewActionOptions(options...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

func (s *StepMobile) SwipeRight(options ...uixt.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  uixt.ACTION_Swipe,
		Params:  "right",
		Options: uixt.NewActionOptions(options...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

func (s *StepMobile) SwipeToTapApp(appName string, options ...uixt.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  uixt.ACTION_SwipeToTapApp,
		Params:  appName,
		Options: uixt.NewActionOptions(options...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

func (s *StepMobile) SwipeToTapText(text string, options ...uixt.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  uixt.ACTION_SwipeToTapText,
		Params:  text,
		Options: uixt.NewActionOptions(options...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

func (s *StepMobile) SwipeToTapTexts(texts interface{}, options ...uixt.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  uixt.ACTION_SwipeToTapTexts,
		Params:  texts,
		Options: uixt.NewActionOptions(options...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

func (s *StepMobile) Input(text string, options ...uixt.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  uixt.ACTION_Input,
		Params:  text,
		Options: uixt.NewActionOptions(options...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

// Sleep specify sleep seconds after last action
func (s *StepMobile) Sleep(nSeconds float64, startTime ...time.Time) *StepMobile {
	action := uixt.MobileAction{
		Method:  uixt.ACTION_Sleep,
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
		Method:  uixt.ACTION_SleepMS,
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
		Method:  uixt.ACTION_SleepRandom,
		Params:  params,
		Options: nil,
	})
	return s
}

func (s *StepMobile) EndToEndDelay(options ...uixt.ActionOption) *StepMobile {
	s.obj().Actions = append(s.obj().Actions, uixt.MobileAction{
		Method:  uixt.ACTION_EndToEndDelay,
		Params:  nil,
		Options: uixt.NewActionOptions(options...),
	})
	return s
}

func (s *StepMobile) ScreenShot(options ...uixt.ActionOption) *StepMobile {
	s.obj().Actions = append(s.obj().Actions, uixt.MobileAction{
		Method:  uixt.ACTION_ScreenShot,
		Params:  nil,
		Options: uixt.NewActionOptions(options...),
	})
	return s
}

func (s *StepMobile) StartCamera() *StepMobile {
	s.obj().Actions = append(s.obj().Actions, uixt.MobileAction{
		Method:  uixt.ACTION_StartCamera,
		Params:  nil,
		Options: nil,
	})
	return s
}

func (s *StepMobile) StopCamera() *StepMobile {
	s.obj().Actions = append(s.obj().Actions, uixt.MobileAction{
		Method:  uixt.ACTION_StopCamera,
		Params:  nil,
		Options: nil,
	})
	return s
}

func (s *StepMobile) DisableAutoPopupHandler() *StepMobile {
	s.IgnorePopup = true
	return s
}

func (s *StepMobile) ClosePopups(options ...uixt.ActionOption) *StepMobile {
	s.obj().Actions = append(s.obj().Actions, uixt.MobileAction{
		Method:  uixt.ACTION_ClosePopups,
		Params:  nil,
		Options: uixt.NewActionOptions(options...),
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
	return StepType(s.obj().OSType)
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
		Check:  uixt.SelectorName,
		Assert: uixt.AssertionExists,
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
		Check:  uixt.SelectorName,
		Assert: uixt.AssertionNotExists,
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
		Check:  uixt.SelectorLabel,
		Assert: uixt.AssertionExists,
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
		Check:  uixt.SelectorLabel,
		Assert: uixt.AssertionNotExists,
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
		Check:  uixt.SelectorOCR,
		Assert: uixt.AssertionExists,
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
		Check:  uixt.SelectorOCR,
		Assert: uixt.AssertionNotExists,
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
		Check:  uixt.SelectorImage,
		Assert: uixt.AssertionExists,
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
		Check:  uixt.SelectorImage,
		Assert: uixt.AssertionNotExists,
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

func (s *StepMobileUIValidation) AssertAppInForeground(packageName string, msg ...string) *StepMobileUIValidation {
	v := Validator{
		Check:  uixt.SelectorForegroundApp,
		Assert: uixt.AssertionEqual,
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
		Check:  uixt.SelectorForegroundApp,
		Assert: uixt.AssertionNotEqual,
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
	uiDriver, err := s.caseRunner.GetUIXTDriver(mobileStep.Serial)
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
			if app, err1 := uiDriver.Driver.GetForegroundApp(); err1 == nil {
				attachments["foreground_app"] = app.AppBaseInfo
			} else {
				log.Warn().Err(err1).Msg("save foreground app failed, ignore")
			}
		}

		// automatic handling of pop-up windows on each step finished
		if !ignorePopup && !s.caseRunner.Config.Get().IgnorePopup {
			if err2 := uiDriver.ClosePopupsHandler(); err2 != nil {
				log.Error().Err(err2).Str("step", step.Name()).Msg("auto handle popup failed")
			}
		}

		// save attachments
		session := uiDriver.Driver.GetSession()
		for key, value := range session.Get(true) {
			attachments[key] = value
		}
		stepResult.Attachments = attachments
		stepResult.Elapsed = time.Since(start).Milliseconds()
	}()

	// prepare actions
	var actions []uixt.MobileAction
	if mobileStep.Actions == nil {
		actions = []uixt.MobileAction{
			{
				Method: mobileStep.Method,
				Params: mobileStep.Params,
			},
		}
	} else {
		actions = mobileStep.Actions
	}

	// run actions
	for _, action := range actions {
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
			if action.Method == uixt.ACTION_LOG {
				log.Info().Interface("action", action.Params).Msg("stat uixt action")
				actionMethod := uixt.ActionMethod(action.Params.(string))
				s.summary.Stat.Actions[actionMethod]++
				continue
			}

			err = uiDriver.DoAction(action)
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

func validateUI(ud *uixt.DriverExt, iValidators []interface{}) (validateResults []*ValidationResult, err error) {
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

		if !ud.DoValidation(validator.Check, validator.Assert, expected, validator.Message) {
			return validateResults, errors.New("step validation failed")
		}

		validataResult.CheckResult = "pass"
		validateResults = append(validateResults, validataResult)
	}
	return validateResults, nil
}
