package hrp

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/uixt"
)

type (
	WDAOptions = uixt.WDAOptions
	WDAOption  = uixt.WDAOption
)

var (
	WithUDID      = uixt.WithUDID
	WithPort      = uixt.WithPort
	WithMjpegPort = uixt.WithMjpegPort
	WithLogOn     = uixt.WithLogOn
)

type IOSStep struct {
	WDAOptions   `yaml:",inline"` // inline refers to https://pkg.go.dev/gopkg.in/yaml.v3#Marshal
	MobileAction `yaml:",inline"`
	Actions      []MobileAction `json:"actions,omitempty" yaml:"actions,omitempty"`
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
func (s *StepIOS) TapXY(x, y float64, options ...ActionOption) *StepIOS {
	action := MobileAction{
		Method: uiTapXY,
		Params: []float64{x, y},
	}
	for _, option := range options {
		option(&action)
	}
	s.step.IOS.Actions = append(s.step.IOS.Actions, action)
	return &StepIOS{step: s.step}
}

// Tap taps on the target element
func (s *StepIOS) Tap(params string, options ...ActionOption) *StepIOS {
	action := MobileAction{
		Method: uiTap,
		Params: params,
	}
	for _, option := range options {
		option(&action)
	}
	s.step.IOS.Actions = append(s.step.IOS.Actions, action)
	return &StepIOS{step: s.step}
}

// Tap taps on the target element by OCR recognition
func (s *StepIOS) TapByOCR(ocrText string, options ...ActionOption) *StepIOS {
	action := MobileAction{
		Method: uiTapByOCR,
		Params: ocrText,
	}
	for _, option := range options {
		option(&action)
	}
	s.step.IOS.Actions = append(s.step.IOS.Actions, action)
	return &StepIOS{step: s.step}
}

// Tap taps on the target element by CV recognition
func (s *StepIOS) TapByCV(imagePath string, options ...ActionOption) *StepIOS {
	action := MobileAction{
		Method: uiTapByCV,
		Params: imagePath,
	}
	for _, option := range options {
		option(&action)
	}
	s.step.IOS.Actions = append(s.step.IOS.Actions, action)
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

func (s *StepIOS) DoubleTap(params string, options ...ActionOption) *StepIOS {
	action := MobileAction{
		Method: uiDoubleTap,
		Params: params,
	}
	for _, option := range options {
		option(&action)
	}
	s.step.IOS.Actions = append(s.step.IOS.Actions, action)
	return &StepIOS{step: s.step}
}

func (s *StepIOS) Swipe(sx, sy, ex, ey int, options ...ActionOption) *StepIOS {
	action := MobileAction{
		Method: uiSwipe,
		Params: []int{sx, sy, ex, ey},
	}
	for _, option := range options {
		option(&action)
	}
	s.step.IOS.Actions = append(s.step.IOS.Actions, action)
	return &StepIOS{step: s.step}
}

func (s *StepIOS) SwipeUp(options ...ActionOption) *StepIOS {
	action := MobileAction{
		Method: uiSwipe,
		Params: "up",
	}
	for _, option := range options {
		option(&action)
	}
	s.step.IOS.Actions = append(s.step.IOS.Actions, action)
	return &StepIOS{step: s.step}
}

func (s *StepIOS) SwipeDown(options ...ActionOption) *StepIOS {
	action := MobileAction{
		Method: uiSwipe,
		Params: "down",
	}
	for _, option := range options {
		option(&action)
	}
	s.step.IOS.Actions = append(s.step.IOS.Actions, action)
	return &StepIOS{step: s.step}
}

func (s *StepIOS) SwipeLeft(options ...ActionOption) *StepIOS {
	action := MobileAction{
		Method: uiSwipe,
		Params: "left",
	}
	for _, option := range options {
		option(&action)
	}
	s.step.IOS.Actions = append(s.step.IOS.Actions, action)
	return &StepIOS{step: s.step}
}

func (s *StepIOS) SwipeRight(options ...ActionOption) *StepIOS {
	action := MobileAction{
		Method: uiSwipe,
		Params: "right",
	}
	for _, option := range options {
		option(&action)
	}
	s.step.IOS.Actions = append(s.step.IOS.Actions, action)
	return &StepIOS{step: s.step}
}

func (s *StepIOS) SwipeToTapApp(appName string, options ...ActionOption) *StepIOS {
	action := MobileAction{
		Method: swipeToTapApp,
		Params: appName,
	}
	for _, option := range options {
		option(&action)
	}
	s.step.IOS.Actions = append(s.step.IOS.Actions, action)
	return &StepIOS{step: s.step}
}

func (s *StepIOS) SwipeToTapText(text string, options ...ActionOption) *StepIOS {
	action := MobileAction{
		Method: swipeToTapText,
		Params: text,
	}
	for _, option := range options {
		option(&action)
	}
	s.step.IOS.Actions = append(s.step.IOS.Actions, action)
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
func (s *StepIOS) Sleep(n float64) *StepIOS {
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
		v.Message = fmt.Sprintf("attribute name [%s] not found", expectedName)
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
		v.Message = fmt.Sprintf("attribute name [%s] should not exist", expectedName)
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
		v.Message = fmt.Sprintf("attribute label [%s] not found", expectedLabel)
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
		v.Message = fmt.Sprintf("attribute label [%s] should not exist", expectedLabel)
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
		v.Message = fmt.Sprintf("ocr text [%s] not found", expectedText)
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
		v.Message = fmt.Sprintf("ocr text [%s] should not exist", expectedText)
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepIOSValidation) AssertImageExists(expectedImagePath string, msg ...string) *StepIOSValidation {
	v := Validator{
		Check:  uiSelectorImage,
		Assert: assertionExists,
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

func (s *StepIOSValidation) AssertImageNotExists(expectedImagePath string, msg ...string) *StepIOSValidation {
	v := Validator{
		Check:  uiSelectorImage,
		Assert: assertionNotExists,
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

func (r *HRPRunner) InitWDAClient(options *WDAOptions) (client *uiDriver, err error) {
	// avoid duplicate init
	if options.UDID == "" && len(r.wdaClients) == 1 {
		for _, v := range r.wdaClients {
			return v, nil
		}
	}

	// avoid duplicate init
	if options.UDID != "" {
		if client, ok := r.wdaClients[options.UDID]; ok {
			return client, nil
		}
	}

	driverExt, err := uixt.InitWDAClient(options)
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
	r.wdaClients[options.UDID] = client

	return client, nil
}

func runStepIOS(s *SessionRunner, step *TStep) (stepResult *StepResult, err error) {
	stepResult = &StepResult{
		Name:        step.Name,
		StepType:    stepTypeIOS,
		Success:     false,
		ContentSize: 0,
	}
	screenshots := make([]string, 0)

	// init wdaClient driver
	wdaClient, err := s.hrpRunner.InitWDAClient(&step.IOS.WDAOptions)
	if err != nil {
		return
	}
	wdaClient.startTime = s.startTime

	defer func() {
		attachments := make(map[string]interface{})
		if err != nil {
			attachments["error"] = err.Error()
		}

		// save attachments
		screenshots = append(screenshots, wdaClient.screenShots...)
		attachments["screenshots"] = screenshots
		stepResult.Attachments = attachments

		// update summary
		s.summary.Records = append(s.summary.Records, stepResult)
		s.summary.Stat.Total += 1
		if stepResult.Success {
			s.summary.Stat.Successes += 1
		} else {
			s.summary.Stat.Failures += 1
			// update summary result to failed
			s.summary.Success = false
		}
	}()

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
	screenshotPath, err := wdaClient.DriverExt.ScreenShot(
		fmt.Sprintf("%d_validate_%d", wdaClient.startTime.Unix(), time.Now().Unix()))
	if err != nil {
		log.Warn().Err(err).Str("step", step.Name).Msg("take screenshot failed")
	} else {
		log.Info().Str("path", screenshotPath).Msg("take screenshot before validation")
		screenshots = append(screenshots, screenshotPath)
	}

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

	startTime   time.Time // used to associate screenshots name
	screenShots []string  // save screenshots path
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
		return fmt.Errorf("invalid %s params, should be bundleId(string), got %v",
			appLaunch, action.Params)
	case appLaunchUnattached:
		if bundleId, ok := action.Params.(string); ok {
			return ud.AppLaunchUnattached(bundleId)
		}
		return fmt.Errorf("invalid %s params, should be bundleId(string), got %v",
			appLaunchUnattached, action.Params)
	case swipeToTapApp:
		if appName, ok := action.Params.(string); ok {
			var x, y, width, height float64
			findApp := func(d *uixt.DriverExt) error {
				var err error
				x, y, width, height, err = d.FindTextByOCR(appName)
				return err
			}
			foundAppAction := func(d *uixt.DriverExt) error {
				// click app to launch
				return d.TapFloat(x+width*0.5, y+height*0.5-20)
			}

			// go to home screen
			if err := ud.WebDriver.Homescreen(); err != nil {
				return errors.Wrap(err, "go to home screen failed")
			}

			// swipe to first screen
			for i := 0; i < 5; i++ {
				ud.SwipeRight()
			}

			// default to retry 5 times
			if action.MaxRetryTimes == 0 {
				action.MaxRetryTimes = 5
			}
			// swipe next screen until app found
			return ud.SwipeUntil("left", findApp, foundAppAction, action.MaxRetryTimes)
		}
		return fmt.Errorf("invalid %s params, should be app name(string), got %v",
			swipeToTapApp, action.Params)
	case swipeToTapText:
		if text, ok := action.Params.(string); ok {
			var x, y, width, height float64
			findText := func(d *uixt.DriverExt) error {
				var err error
				x, y, width, height, err = d.FindTextByOCR(text)
				return err
			}
			foundTextAction := func(d *uixt.DriverExt) error {
				// tap text
				return d.TapFloat(x+width*0.5, y+height*0.5)
			}

			// default to retry 10 times
			if action.MaxRetryTimes == 0 {
				action.MaxRetryTimes = 10
			}
			// swipe until live room found
			return ud.SwipeUntil("up", findText, foundTextAction, action.MaxRetryTimes)
		}
		return fmt.Errorf("invalid %s params, should be app text(string), got %v",
			swipeToTapText, action.Params)
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
			return ud.TapXY(location[0], location[1], action.Identifier)
		}
		return fmt.Errorf("invalid %s params: %v", uiTapXY, action.Params)
	case uiTap:
		if param, ok := action.Params.(string); ok {
			return ud.Tap(param, action.Identifier, action.IgnoreNotFoundError)
		}
		return fmt.Errorf("invalid %s params: %v", uiTap, action.Params)
	case uiTapByOCR:
		if ocrText, ok := action.Params.(string); ok {
			return ud.TapByOCR(ocrText, action.Identifier, action.IgnoreNotFoundError)
		}
		return fmt.Errorf("invalid %s params: %v", uiTapByOCR, action.Params)
	case uiTapByCV:
		if imagePath, ok := action.Params.(string); ok {
			return ud.TapByCV(imagePath, action.Identifier, action.IgnoreNotFoundError)
		}
		return fmt.Errorf("invalid %s params: %v", uiTapByCV, action.Params)
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
		if positions, ok := action.Params.([]float64); ok {
			// relative fromX, fromY, toX, toY of window size: [0.5, 0.9, 0.5, 0.1]
			if len(positions) != 4 {
				return fmt.Errorf("invalid swipe params [fromX, fromY, toX, toY]: %v", positions)
			}
			return ud.SwipeRelative(
				positions[0], positions[1], positions[2], positions[3], action.Identifier)
		}
		if direction, ok := action.Params.(string); ok {
			return ud.SwipeTo(direction, action.Identifier)
		}
		return fmt.Errorf("invalid %s params: %v", uiSwipe, action.Params)
	case uiInput:
		// input text on current active element
		// append \n to send text with enter
		// send \b\b\b to delete 3 chars
		param := fmt.Sprintf("%v", action.Params)
		return ud.SendKeys(param)
	case ctlSleep:
		if param, ok := action.Params.(json.Number); ok {
			seconds, _ := param.Float64()
			time.Sleep(time.Duration(seconds*1000) * time.Millisecond)
			return nil
		} else if param, ok := action.Params.(float64); ok {
			time.Sleep(time.Duration(param*1000) * time.Millisecond)
			return nil
		}
		return fmt.Errorf("invalid sleep params: %v(%T)", action.Params, action.Params)
	case ctlScreenShot:
		// take snapshot
		log.Info().Msg("take snapshot for current screen")
		screenshotPath, err := ud.ScreenShot(fmt.Sprintf("%d_screenshot_%d",
			ud.startTime.Unix(), time.Now().Unix()))
		if err != nil {
			return errors.Wrap(err, "take screenshot failed")
		}
		log.Info().Str("path", screenshotPath).Msg("take screenshot")
		ud.screenShots = append(ud.screenShots, screenshotPath)
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
		case uiSelectorImage:
			result = (ud.IsImageExist(expected) == exists)
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
