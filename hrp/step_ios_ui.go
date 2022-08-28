package hrp

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/uixt"
)

type IOSConfig struct {
	WDADevice
}

type WDADevice struct {
	UDID      string `json:"udid,omitempty" yaml:"udid,omitempty"`
	Port      int    `json:"port,omitempty" yaml:"port,omitempty"`
	MjpegPort int    `json:"mjpeg_port,omitempty" yaml:"mjpeg_port,omitempty"`
}

type IOSStep struct {
	WDADevice
	MobileAction
	Actions []MobileAction `json:"actions,omitempty" yaml:"actions,omitempty"`
}

// StepIOS implements IStep interface.
type StepIOS struct {
	step *TStep
}

func (s *StepIOS) UDID(udid string) *StepIOS {
	s.step.IOS.UDID = udid
	return &StepIOS{step: s.step}
}

func (s *StepIOS) InstallApp(path string) *StepIOS {
	s.step.IOS.Actions = append(s.step.IOS.Actions, MobileAction{
		Method: appInstall,
		Params: path,
	})
	return s
}

func (s *StepIOS) AppLaunch(bundleId string) *StepIOS {
	s.step.IOS.Actions = append(s.step.IOS.Actions, MobileAction{
		Method: appLaunch,
		Params: bundleId,
	})
	return s
}

func (s *StepIOS) AppLaunchUnattached(bundleId string) *StepIOS {
	s.step.IOS.Actions = append(s.step.IOS.Actions, MobileAction{
		Method: appLaunchUnattached,
		Params: bundleId,
	})
	return s
}

func (s *StepIOS) AppTerminate(bundleId string) *StepIOS {
	s.step.IOS.Actions = append(s.step.IOS.Actions, MobileAction{
		Method: appTerminate,
		Params: bundleId,
	})
	return s
}

func (s *StepIOS) Home() *StepIOS {
	s.step.IOS.Actions = append(s.step.IOS.Actions, MobileAction{
		Method: uiHome,
		Params: nil,
	})
	return &StepIOS{step: s.step}
}

// TapXY taps the point {X,Y}, X & Y is percentage of coordinates
func (s *StepIOS) TapXY(x, y float64) *StepIOS {
	s.step.IOS.Actions = append(s.step.IOS.Actions, MobileAction{
		Method: uiTapXY,
		Params: []float64{x, y},
	})
	return &StepIOS{step: s.step}
}

// Tap taps on the target element
func (s *StepIOS) Tap(params string) *StepIOS {
	s.step.IOS.Actions = append(s.step.IOS.Actions, MobileAction{
		Method: uiTap,
		Params: params,
	})
	return &StepIOS{step: s.step}
}

// DoubleTapXY double taps the point {X,Y}, X & Y is percentage of coordinates
func (s *StepIOS) DoubleTapXY(x, y float64) *StepIOS {
	s.step.IOS.Actions = append(s.step.IOS.Actions, MobileAction{
		Method: uiDoubleTapXY,
		Params: []float64{x, y},
	})
	return &StepIOS{step: s.step}
}

func (s *StepIOS) DoubleTap(params string) *StepIOS {
	s.step.IOS.Actions = append(s.step.IOS.Actions, MobileAction{
		Method: uiDoubleTap,
		Params: params,
	})
	return &StepIOS{step: s.step}
}

func (s *StepIOS) Swipe(sx, sy, ex, ey int) *StepIOS {
	s.step.IOS.Actions = append(s.step.IOS.Actions, MobileAction{
		Method: uiSwipe,
		Params: []int{sx, sy, ex, ey},
	})
	return &StepIOS{step: s.step}
}

func (s *StepIOS) SwipeUp() *StepIOS {
	s.step.IOS.Actions = append(s.step.IOS.Actions, MobileAction{
		Method: uiSwipe,
		Params: "up",
	})
	return &StepIOS{step: s.step}
}

func (s *StepIOS) SwipeDown() *StepIOS {
	s.step.IOS.Actions = append(s.step.IOS.Actions, MobileAction{
		Method: uiSwipe,
		Params: "down",
	})
	return &StepIOS{step: s.step}
}

func (s *StepIOS) SwipeLeft() *StepIOS {
	s.step.IOS.Actions = append(s.step.IOS.Actions, MobileAction{
		Method: uiSwipe,
		Params: "left",
	})
	return &StepIOS{step: s.step}
}

func (s *StepIOS) SwipeRight() *StepIOS {
	s.step.IOS.Actions = append(s.step.IOS.Actions, MobileAction{
		Method: uiSwipe,
		Params: "right",
	})
	return &StepIOS{step: s.step}
}

func (s *StepIOS) Input(text string) *StepIOS {
	s.step.IOS.Actions = append(s.step.IOS.Actions, MobileAction{
		Method: uiInput,
		Params: text,
	})
	return &StepIOS{step: s.step}
}

// Times specify running times for run last action
func (s *StepIOS) Times(n int) *StepIOS {
	if n <= 0 {
		log.Warn().Int("n", n).Msg("times should be positive, set to 1")
		n = 1
	}

	actionsTotal := len(s.step.IOS.Actions)
	if actionsTotal == 0 {
		return s
	}

	// actionsTotal >=1 && n >= 1
	lastAction := s.step.IOS.Actions[actionsTotal-1 : actionsTotal][0]
	for i := 0; i < n-1; i++ {
		// duplicate last action n-1 times
		s.step.IOS.Actions = append(s.step.IOS.Actions, lastAction)
	}
	return &StepIOS{step: s.step}
}

// Sleep specify sleep seconds after last action
func (s *StepIOS) Sleep(n int) *StepIOS {
	s.step.IOS.Actions = append(s.step.IOS.Actions, MobileAction{
		Method: ctlSleep,
		Params: n,
	})
	return &StepIOS{step: s.step}
}

func (s *StepIOS) ScreenShot() *StepIOS {
	s.step.IOS.Actions = append(s.step.IOS.Actions, MobileAction{
		Method: ctlScreenShot,
		Params: nil,
	})
	return &StepIOS{step: s.step}
}

func (s *StepIOS) StartCamera() *StepIOS {
	s.step.IOS.Actions = append(s.step.IOS.Actions, MobileAction{
		Method: ctlStartCamera,
		Params: nil,
	})
	return &StepIOS{step: s.step}
}

func (s *StepIOS) StopCamera() *StepIOS {
	s.step.IOS.Actions = append(s.step.IOS.Actions, MobileAction{
		Method: ctlStopCamera,
		Params: nil,
	})
	return &StepIOS{step: s.step}
}

// Validate switches to step validation.
func (s *StepIOS) Validate() *StepIOSValidation {
	return &StepIOSValidation{
		step: s.step,
	}
}

func (s *StepIOS) Name() string {
	return s.step.Name
}

func (s *StepIOS) Type() StepType {
	return stepTypeAndroid
}

func (s *StepIOS) Struct() *TStep {
	return s.step
}

func (s *StepIOS) Run(r *SessionRunner) (*StepResult, error) {
	return runStepIOS(r, s.step)
}

// StepIOSValidation implements IStep interface.
type StepIOSValidation struct {
	step *TStep
}

func (s *StepIOSValidation) AssertNameExists(expectedName string, msg ...string) *StepIOSValidation {
	v := Validator{
		Check:  uiSelectorName,
		Assert: assertionExists,
		Expect: expectedName,
	}
	if len(msg) > 0 {
		v.Message = msg[0]
	} else {
		v.Message = fmt.Sprintf("[%s] not found", expectedName)
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepIOSValidation) AssertNameNotExists(expectedName string, msg ...string) *StepIOSValidation {
	v := Validator{
		Check:  uiSelectorName,
		Assert: assertionNotExists,
		Expect: expectedName,
	}
	if len(msg) > 0 {
		v.Message = msg[0]
	} else {
		v.Message = fmt.Sprintf("[%s] should not exist", expectedName)
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepIOSValidation) AssertLabelExists(expectedLabel string, msg ...string) *StepIOSValidation {
	v := Validator{
		Check:  uiSelectorLabel,
		Assert: assertionExists,
		Expect: expectedLabel,
	}
	if len(msg) > 0 {
		v.Message = msg[0]
	} else {
		v.Message = fmt.Sprintf("[%s] not found", expectedLabel)
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepIOSValidation) AssertLabelNotExists(expectedLabel string, msg ...string) *StepIOSValidation {
	v := Validator{
		Check:  uiSelectorLabel,
		Assert: assertionNotExists,
		Expect: expectedLabel,
	}
	if len(msg) > 0 {
		v.Message = msg[0]
	} else {
		v.Message = fmt.Sprintf("[%s] should not exist", expectedLabel)
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepIOSValidation) AssertOCRExists(expectedText string, msg ...string) *StepIOSValidation {
	v := Validator{
		Check:  uiSelectorOCR,
		Assert: assertionExists,
		Expect: expectedText,
	}
	if len(msg) > 0 {
		v.Message = msg[0]
	} else {
		v.Message = fmt.Sprintf("[%s] not found", expectedText)
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepIOSValidation) AssertOCRNotExists(expectedText string, msg ...string) *StepIOSValidation {
	v := Validator{
		Check:  uiSelectorOCR,
		Assert: assertionNotExists,
		Expect: expectedText,
	}
	if len(msg) > 0 {
		v.Message = msg[0]
	} else {
		v.Message = fmt.Sprintf("[%s] should not exist", expectedText)
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepIOSValidation) Name() string {
	return s.step.Name
}

func (s *StepIOSValidation) Type() StepType {
	return stepTypeAndroid
}

func (s *StepIOSValidation) Struct() *TStep {
	return s.step
}

func (s *StepIOSValidation) Run(r *SessionRunner) (*StepResult, error) {
	return runStepIOS(r, s.step)
}

func (r *HRPRunner) InitWDAClient(device WDADevice) (client *uiDriver, err error) {
	// avoid duplicate init
	if device.UDID == "" && len(r.wdaClients) == 1 {
		for _, v := range r.wdaClients {
			return v, nil
		}
	}

	// avoid duplicate init
	if device.UDID != "" {
		if client, ok := r.wdaClients[device.UDID]; ok {
			return client, nil
		}
	}

	driverExt, err := uixt.InitWDAClient(device.UDID, device.Port, device.MjpegPort)
	if err != nil {
		return nil, err
	}
	client = &uiDriver{
		DriverExt: *driverExt,
	}

	// cache wda client
	if r.wdaClients == nil {
		r.wdaClients = make(map[string]*uiDriver)
	}
	r.wdaClients[device.UDID] = client

	return client, nil
}

func runStepIOS(s *SessionRunner, step *TStep) (stepResult *StepResult, err error) {
	stepResult = &StepResult{
		Name:        step.Name,
		StepType:    stepTypeIOS,
		Success:     false,
		ContentSize: 0,
	}

	// init wdaClient driver
	wdaClient, err := s.hrpRunner.InitWDAClient(step.IOS.WDADevice)
	if err != nil {
		return
	}

	// prepare actions
	var actions []MobileAction
	if step.IOS.Actions == nil {
		actions = []MobileAction{
			{
				Method: step.IOS.Method,
				Params: step.IOS.Params,
			},
		}
	} else {
		actions = step.IOS.Actions
	}

	// run actions
	for _, action := range actions {
		if err := wdaClient.doAction(action); err != nil {
			return stepResult, err
		}
	}

	// take snapshot
	screenshotPath, err := wdaClient.DriverExt.ScreenShot(fmt.Sprintf("validate_%s", step.Name))
	if err != nil {
		log.Warn().Err(err).Str("step", step.Name).Msg("take screenshot failed")
	}
	log.Info().Str("path", screenshotPath).Msg("take screenshot before validation")

	// validate
	validateResults, err := wdaClient.doValidation(step.Validators)
	if err != nil {
		return
	}
	sessionData := newSessionData()
	sessionData.Validators = validateResults
	stepResult.Data = sessionData
	stepResult.Success = true
	return stepResult, nil
}

var errActionNotImplemented = errors.New("UI action not implemented")

type uiDriver struct {
	uixt.DriverExt
}

func (ud *uiDriver) doAction(action MobileAction) error {
	log.Info().Str("method", string(action.Method)).Interface("params", action.Params).Msg("start iOS UI action")

	switch action.Method {
	case appInstall:
		// TODO
		return errActionNotImplemented
	case appLaunch:
		if bundleId, ok := action.Params.(string); ok {
			return ud.AppLaunch(bundleId)
		}
		return fmt.Errorf("app_launch params should be bundleId(string), got %v", action.Params)
	case appLaunchUnattached:
		if bundleId, ok := action.Params.(string); ok {
			return ud.AppLaunchUnattached(bundleId)
		}
		return fmt.Errorf("app_launch_unattached params should be bundleId(string), got %v", action.Params)
	case appTerminate:
		if bundleId, ok := action.Params.(string); ok {
			success, err := ud.AppTerminate(bundleId)
			if err != nil {
				return errors.Wrap(err, "failed to terminate app")
			}
			if !success {
				log.Warn().Str("bundleId", bundleId).Msg("app was not running")
			}
			return nil
		}
		return fmt.Errorf("app_terminate params should be bundleId(string), got %v", action.Params)
	case uiHome:
		return ud.Homescreen()
	case uiTapXY:
		if location, ok := action.Params.([]float64); ok {
			// relative x,y of window size: [0.5, 0.5]
			if len(location) != 2 {
				return fmt.Errorf("invalid tap location params: %v", location)
			}
			return ud.TapXY(location[0], location[1])
		}
		return fmt.Errorf("invalid %s params: %v", uiTapXY, action.Params)
	case uiTap:
		if param, ok := action.Params.(string); ok {
			return ud.Tap(param)
		}
		return fmt.Errorf("invalid %s params: %v", uiTap, action.Params)
	case uiDoubleTapXY:
		if location, ok := action.Params.([]float64); ok {
			// relative x,y of window size: [0.5, 0.5]
			if len(location) != 2 {
				return fmt.Errorf("invalid tap location params: %v", location)
			}
			return ud.DoubleTapXY(location[0], location[1])
		}
		return fmt.Errorf("invalid %s params: %v", uiDoubleTapXY, action.Params)
	case uiDoubleTap:
		if param, ok := action.Params.(string); ok {
			return ud.DoubleTap(param)
		}
		return fmt.Errorf("invalid %s params: %v", uiDoubleTap, action.Params)
	case uiSwipe:
		if param, ok := action.Params.(string); ok {
			return ud.SwipeTo(param)
		}
		return fmt.Errorf("invalid %s params: %v", uiSwipe, action.Params)
	case uiInput:
		// input text on current active element
		// append \n to send text with enter
		// send \b\b\b to delete 3 chars
		param := fmt.Sprintf("%v", action.Params)
		return ud.SendKeys(param)
	case ctlSleep:
		if param, ok := action.Params.(int); ok {
			time.Sleep(time.Duration(param) * time.Second)
			return nil
		}
		return fmt.Errorf("invalid sleep params: %v", action.Params)
	case ctlScreenShot:
		// take snapshot
		log.Info().Msg("take snapshot for current screen")
		var screenshotPath string
		var err error
		if param, ok := action.Params.(string); ok {
			screenshotPath, err = ud.ScreenShot(fmt.Sprintf("screenshot_%s", param))
		} else {
			screenshotPath, err = ud.ScreenShot(fmt.Sprintf("screenshot_%d", time.Now().Unix()))
		}
		log.Info().Str("path", screenshotPath).Msg("take screenshot")
		return err
	case ctlStartCamera:
		// start camera, alias for app_launch com.apple.camera
		return ud.AppLaunch("com.apple.camera")
	case ctlStopCamera:
		// stop camera, alias for app_terminate com.apple.camera
		success, err := ud.AppTerminate("com.apple.camera")
		if err != nil {
			return errors.Wrap(err, "failed to terminate camera")
		}
		if !success {
			log.Warn().Msg("camera was not running")
		}
		return nil
	}
	return nil
}

func (ud *uiDriver) doValidation(iValidators []interface{}) (validateResults []*ValidationResult, err error) {
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

		var exists bool
		if validator.Assert == assertionExists {
			exists = true
		} else {
			exists = false
		}
		var result bool
		switch validator.Check {
		case uiSelectorName:
			result = (ud.IsNameExist(expected) == exists)
		case uiSelectorLabel:
			result = (ud.IsLabelExist(expected) == exists)
		case uiSelectorOCR:
			result = (ud.IsOCRExist(expected) == exists)
		}

		if result {
			log.Info().
				Str("assert", validator.Assert).
				Str("expect", expected).
				Msg("validate UI success")
			validataResult.CheckResult = "pass"
			validateResults = append(validateResults, validataResult)
		} else {
			log.Error().
				Str("assert", validator.Assert).
				Str("expect", expected).
				Str("msg", validator.Message).
				Msg("validate UI failed")
			validateResults = append(validateResults, validataResult)
			return validateResults, errors.New("step validation failed")
		}
	}
	return validateResults, nil
}
