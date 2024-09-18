package hrp

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/code"
	"github.com/httprunner/httprunner/v4/hrp/internal/sdk"
	"github.com/httprunner/httprunner/v4/hrp/pkg/uixt"
)

var uiClients map[string]*uixt.DriverExt // UI automation clients for iOS and Android, key is udid/serial

func initUIClient(serial, osType string) (client *uixt.DriverExt, err error) {
	if uiClients == nil {
		uiClients = make(map[string]*uixt.DriverExt)
	}

	// avoid duplicate init
	if serial == "" && len(uiClients) > 0 {
		for _, v := range uiClients {
			return v, nil
		}
	}

	// avoid duplicate init
	if serial != "" {
		if client, ok := uiClients[serial]; ok {
			return client, nil
		}
	}

	var device uixt.IDevice
	if osType == "ios" {
		device, err = uixt.NewIOSDevice(uixt.WithUDID(serial))
	} else if osType == "harmony" {
		device, err = uixt.NewHarmonyDevice(uixt.WithConnectKey(serial))
	} else {
		device, err = uixt.NewAndroidDevice(uixt.WithSerialNumber(serial))
	}
	if err != nil {
		return nil, errors.Wrapf(err, "init %s device failed", osType)
	}

	client, err = device.NewDriver()
	if err != nil {
		return nil, err
	}

	// cache wda client
	if uiClients == nil {
		uiClients = make(map[string]*uixt.DriverExt)
	}
	uiClients[client.Device.UUID()] = client

	return client, nil
}

type MobileUI struct {
	Serial            string `json:"serial,omitempty" yaml:"serial,omitempty"` // android serial or ios udid
	uixt.MobileAction `yaml:",inline"`
	Actions           []uixt.MobileAction `json:"actions,omitempty" yaml:"actions,omitempty"`
}

// StepMobile implements IStep interface.
type StepMobile struct {
	step *TStep
}

// uniform interface for all types of mobile systems
func (s *StepMobile) obj() *MobileUI {
	if s.step.IOS != nil {
		return s.step.IOS
	}
	return s.step.Android
}

func (s *StepMobile) Serial(serial string) *StepMobile {
	s.obj().Serial = serial
	return &StepMobile{step: s.step}
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
	return &StepMobile{step: s.step}
}

// TapXY taps the point {X,Y}, X & Y is percentage of coordinates
func (s *StepMobile) TapXY(x, y float64, options ...uixt.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  uixt.ACTION_TapXY,
		Params:  []float64{x, y},
		Options: uixt.NewActionOptions(options...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return &StepMobile{step: s.step}
}

// TapAbsXY taps the point {X,Y}, X & Y is absolute coordinates
func (s *StepMobile) TapAbsXY(x, y float64, options ...uixt.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  uixt.ACTION_TapAbsXY,
		Params:  []float64{x, y},
		Options: uixt.NewActionOptions(options...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return &StepMobile{step: s.step}
}

// Tap taps on the target element
func (s *StepMobile) Tap(params string, options ...uixt.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  uixt.ACTION_Tap,
		Params:  params,
		Options: uixt.NewActionOptions(options...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return &StepMobile{step: s.step}
}

// TapByOCR taps on the target element by OCR recognition
func (s *StepMobile) TapByOCR(ocrText string, options ...uixt.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  uixt.ACTION_TapByOCR,
		Params:  ocrText,
		Options: uixt.NewActionOptions(options...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return &StepMobile{step: s.step}
}

// TapByCV taps on the target element by CV recognition
func (s *StepMobile) TapByCV(imagePath string, options ...uixt.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  uixt.ACTION_TapByCV,
		Params:  imagePath,
		Options: uixt.NewActionOptions(options...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return &StepMobile{step: s.step}
}

// TapByUITypes taps on the target element specified by uiTypes, the higher the uiTypes, the higher the priority
func (s *StepMobile) TapByUITypes(options ...uixt.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  uixt.ACTION_TapByCV,
		Options: uixt.NewActionOptions(options...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return &StepMobile{step: s.step}
}

// DoubleTapXY double taps the point {X,Y}, X & Y is percentage of coordinates
func (s *StepMobile) DoubleTapXY(x, y float64, options ...uixt.ActionOption) *StepMobile {
	s.obj().Actions = append(s.obj().Actions, uixt.MobileAction{
		Method:  uixt.ACTION_DoubleTapXY,
		Params:  []float64{x, y},
		Options: uixt.NewActionOptions(options...),
	})
	return &StepMobile{step: s.step}
}

func (s *StepMobile) DoubleTap(params string, options ...uixt.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  uixt.ACTION_DoubleTap,
		Params:  params,
		Options: uixt.NewActionOptions(options...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return &StepMobile{step: s.step}
}

func (s *StepMobile) Back(options ...uixt.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  uixt.ACTION_Back,
		Params:  nil,
		Options: uixt.NewActionOptions(options...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return &StepMobile{step: s.step}
}

func (s *StepMobile) Swipe(sx, sy, ex, ey float64, options ...uixt.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  uixt.ACTION_Swipe,
		Params:  []float64{sx, sy, ex, ey},
		Options: uixt.NewActionOptions(options...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return &StepMobile{step: s.step}
}

func (s *StepMobile) SwipeUp(options ...uixt.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  uixt.ACTION_Swipe,
		Params:  "up",
		Options: uixt.NewActionOptions(options...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return &StepMobile{step: s.step}
}

func (s *StepMobile) SwipeDown(options ...uixt.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  uixt.ACTION_Swipe,
		Params:  "down",
		Options: uixt.NewActionOptions(options...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return &StepMobile{step: s.step}
}

func (s *StepMobile) SwipeLeft(options ...uixt.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  uixt.ACTION_Swipe,
		Params:  "left",
		Options: uixt.NewActionOptions(options...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return &StepMobile{step: s.step}
}

func (s *StepMobile) SwipeRight(options ...uixt.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  uixt.ACTION_Swipe,
		Params:  "right",
		Options: uixt.NewActionOptions(options...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return &StepMobile{step: s.step}
}

func (s *StepMobile) SwipeToTapApp(appName string, options ...uixt.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  uixt.ACTION_SwipeToTapApp,
		Params:  appName,
		Options: uixt.NewActionOptions(options...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return &StepMobile{step: s.step}
}

func (s *StepMobile) SwipeToTapText(text string, options ...uixt.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  uixt.ACTION_SwipeToTapText,
		Params:  text,
		Options: uixt.NewActionOptions(options...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return &StepMobile{step: s.step}
}

func (s *StepMobile) SwipeToTapTexts(texts interface{}, options ...uixt.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  uixt.ACTION_SwipeToTapTexts,
		Params:  texts,
		Options: uixt.NewActionOptions(options...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return &StepMobile{step: s.step}
}

func (s *StepMobile) Input(text string, options ...uixt.ActionOption) *StepMobile {
	action := uixt.MobileAction{
		Method:  uixt.ACTION_Input,
		Params:  text,
		Options: uixt.NewActionOptions(options...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return &StepMobile{step: s.step}
}

// Sleep specify sleep seconds after last action
func (s *StepMobile) Sleep(n float64) *StepMobile {
	s.obj().Actions = append(s.obj().Actions, uixt.MobileAction{
		Method:  uixt.ACTION_Sleep,
		Params:  n,
		Options: nil,
	})
	return &StepMobile{step: s.step}
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
	return &StepMobile{step: s.step}
}

func (s *StepMobile) EndToEndDelay(options ...uixt.ActionOption) *StepMobile {
	s.obj().Actions = append(s.obj().Actions, uixt.MobileAction{
		Method:  uixt.ACTION_EndToEndDelay,
		Params:  nil,
		Options: uixt.NewActionOptions(options...),
	})
	return &StepMobile{step: s.step}
}

func (s *StepMobile) ScreenShot(options ...uixt.ActionOption) *StepMobile {
	s.obj().Actions = append(s.obj().Actions, uixt.MobileAction{
		Method:  uixt.ACTION_ScreenShot,
		Params:  nil,
		Options: uixt.NewActionOptions(options...),
	})
	return &StepMobile{step: s.step}
}

func (s *StepMobile) StartCamera() *StepMobile {
	s.obj().Actions = append(s.obj().Actions, uixt.MobileAction{
		Method:  uixt.ACTION_StartCamera,
		Params:  nil,
		Options: nil,
	})
	return &StepMobile{step: s.step}
}

func (s *StepMobile) StopCamera() *StepMobile {
	s.obj().Actions = append(s.obj().Actions, uixt.MobileAction{
		Method:  uixt.ACTION_StopCamera,
		Params:  nil,
		Options: nil,
	})
	return &StepMobile{step: s.step}
}

func (s *StepMobile) ClosePopups(options ...uixt.ActionOption) *StepMobile {
	s.obj().Actions = append(s.obj().Actions, uixt.MobileAction{
		Method:  uixt.ACTION_ClosePopups,
		Params:  nil,
		Options: uixt.NewActionOptions(options...),
	})
	return &StepMobile{step: s.step}
}

// Validate switches to step validation.
func (s *StepMobile) Validate() *StepMobileUIValidation {
	return &StepMobileUIValidation{
		step: s.step,
	}
}

func (s *StepMobile) Name() string {
	return s.step.Name
}

func (s *StepMobile) Type() StepType {
	if s.step.Android != nil {
		return stepTypeAndroid
	}
	return stepTypeIOS
}

func (s *StepMobile) Struct() *TStep {
	return s.step
}

func (s *StepMobile) Run(r *SessionRunner) (*StepResult, error) {
	return runStepMobileUI(r, s.step)
}

// StepMobileUIValidation implements IStep interface.
type StepMobileUIValidation struct {
	step *TStep
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
	s.step.Validators = append(s.step.Validators, v)
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
	s.step.Validators = append(s.step.Validators, v)
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
	s.step.Validators = append(s.step.Validators, v)
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
	s.step.Validators = append(s.step.Validators, v)
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
	s.step.Validators = append(s.step.Validators, v)
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
	s.step.Validators = append(s.step.Validators, v)
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
	s.step.Validators = append(s.step.Validators, v)
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
	s.step.Validators = append(s.step.Validators, v)
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
	s.step.Validators = append(s.step.Validators, v)
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
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepMobileUIValidation) Name() string {
	return s.step.Name
}

func (s *StepMobileUIValidation) Type() StepType {
	if s.step.Android != nil {
		return stepTypeAndroid + stepTypeSuffixValidation
	}
	return stepTypeIOS + stepTypeSuffixValidation
}

func (s *StepMobileUIValidation) Struct() *TStep {
	return s.step
}

func (s *StepMobileUIValidation) Run(r *SessionRunner) (*StepResult, error) {
	return runStepMobileUI(r, s.step)
}

func runStepMobileUI(s *SessionRunner, step *TStep) (stepResult *StepResult, err error) {
	var osType string
	var mobileStep *MobileUI
	if step.IOS != nil {
		// ios step
		osType = "ios"
		mobileStep = step.IOS
	} else {
		// android step
		osType = "android"
		mobileStep = step.Android
	}

	// report GA event
	go sdk.SendGA4Event("hrp_run_ui", map[string]interface{}{
		"osType": osType,
	})

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

	stepResult = &StepResult{
		Name:        step.Name,
		Identifier:  identifier,
		StepType:    StepType(osType),
		Success:     false,
		ContentSize: 0,
	}

	// init wda/uia driver
	uiDriver, err := initUIClient(mobileStep.Serial, osType)
	if err != nil {
		return
	}

	defer func() {
		attachments := make(map[string]interface{})
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
		if !step.IgnorePopup && !s.IgnorePopup() {
			if err2 := uiDriver.ClosePopupsHandler(); err2 != nil {
				log.Error().Err(err2).Str("step", step.Name).Msg("auto handle popup failed")
			}
		}

		// save attachments
		session := uiDriver.Driver.GetSession()
		for key, value := range session.GetAll() {
			attachments[key] = value
		}
		stepResult.Attachments = attachments
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
			if action.Params, err = s.caseRunner.parser.Parse(action.Params, step.Variables); err != nil {
				if !code.IsErrorPredefined(err) {
					err = errors.Wrap(code.ParseError,
						fmt.Sprintf("parse action params failed: %v", err))
				}
				return stepResult, err
			}
			if err := uiDriver.DoAction(action); err != nil {
				if !code.IsErrorPredefined(err) {
					err = errors.Wrap(code.MobileUIDriverError, err.Error())
				}
				return stepResult, err
			}
		}
	}

	// validate
	validateResults, err := validateUI(uiDriver, step.Validators)
	if err != nil {
		if !code.IsErrorPredefined(err) {
			err = errors.Wrap(code.MobileUIValidationError, err.Error())
		}
		return
	}
	sessionData := newSessionData()
	sessionData.Validators = validateResults
	stepResult.Data = sessionData
	stepResult.Success = true
	return stepResult, nil
}
