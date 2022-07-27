package hrp

import (
	"fmt"
	"strings"

	"github.com/electricbubble/gwda"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type IOSAction struct {
	MobileAction
	UDID    string         `json:"udid,omitempty" yaml:"udid,omitempty"`
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

func (s *StepIOS) Click(params interface{}) *StepIOS {
	s.step.IOS.Actions = append(s.step.IOS.Actions, MobileAction{
		Method: uiClick,
		Params: params,
	})
	return &StepIOS{step: s.step}
}

func (s *StepIOS) DoubleClick(params interface{}) *StepIOS {
	s.step.IOS.Actions = append(s.step.IOS.Actions, MobileAction{
		Method: uiDoubleClick,
		Params: params,
	})
	return &StepIOS{step: s.step}
}

func (s *StepIOS) LongClick(params interface{}) *StepIOS {
	s.step.IOS.Actions = append(s.step.IOS.Actions, MobileAction{
		Method: uiLongClick,
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

func (s *StepIOS) StartAppByClick(name string) *StepIOS {
	s.step.IOS.Actions = append(s.step.IOS.Actions, MobileAction{
		Method: appClick,
		Params: name,
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

func (s *StepIOSValidation) AssertNameExists(expectedName string, msg string) *StepIOSValidation {
	v := Validator{
		Check:   "UI",
		Assert:  "name_exists",
		Expect:  expectedName,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepIOSValidation) AssertNameNotExists(expectedName string, msg string) *StepIOSValidation {
	v := Validator{
		Check:   "UI",
		Assert:  "name_not_exists",
		Expect:  expectedName,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepIOSValidation) AssertXpathExists(expectedXpath string, msg string) *StepIOSValidation {
	v := Validator{
		Check:   "UI",
		Assert:  "xpath_exists",
		Expect:  expectedXpath,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepIOSValidation) AssertXpathNotExists(expectedXpath string, msg string) *StepIOSValidation {
	v := Validator{
		Check:   "UI",
		Assert:  "xpath_not_exists",
		Expect:  expectedXpath,
		Message: msg,
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

func (r *HRPRunner) InitWDAClient(udid string) (client *wdaClient, err error) {
	defer func() {
		if err != nil {
			return
		}
		// check if WDA is healthy
		ok, e := client.Driver.IsWdaHealthy()
		if err != nil {
			err = errors.Wrap(e, "check WDA health failed")
			return
		}
		if !ok {
			err = errors.New("WDA is not healthy")
			return
		}
	}()

	// avoid duplicate init
	if udid == "" && len(r.wdaClients) == 1 {
		for _, v := range r.wdaClients {
			return v, nil
		}
	}

	targetDevice, err := getAttachedIOSDevice(udid)
	if err != nil {
		return nil, err
	}

	// avoid duplicate init
	if client, ok := r.wdaClients[targetDevice.SerialNumber()]; ok {
		return client, nil
	}

	// init WDA driver
	driver, err := gwda.NewUSBDriver(nil, *targetDevice)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init WDA driver")
	}

	// get device window size
	windowSize, err := driver.WindowSize()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get windows size")
	}

	// cache wda client
	r.wdaClients = make(map[string]*wdaClient)
	client = &wdaClient{
		Driver:     driver,
		WindowSize: windowSize,
	}
	r.wdaClients[targetDevice.SerialNumber()] = client

	return client, nil
}

func getAttachedIOSDevice(udid string) (*gwda.Device, error) {
	// get all attached deivces
	devices, err := gwda.DeviceList()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get attached ios devices list")
	}
	if len(devices) == 0 {
		return nil, errors.New("no ios devices attached")
	}

	if udid == "" {
		return &devices[0], nil
	}

	// find device by udid
	for _, device := range devices {
		if device.SerialNumber() == udid {
			return &device, nil
		}
	}

	return nil, fmt.Errorf("device %s is not attached", udid)
}

func runStepIOS(r *SessionRunner, step *TStep) (stepResult *StepResult, err error) {
	stepResult = &StepResult{
		Name:        step.Name,
		StepType:    stepTypeIOS,
		Success:     false,
		ContentSize: 0,
	}

	// init wdaClient driver
	wdaClient, err := r.hrpRunner.InitWDAClient(step.IOS.UDID)
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

type wdaClient struct {
	Driver     gwda.WebDriver
	WindowSize gwda.Size
}

func (w *wdaClient) doAction(action MobileAction) error {
	log.Info().Str("method", string(action.Method)).Interface("params", action.Params).Msg("start iOS UI action")

	switch action.Method {
	case appInstall:
		// TODO
		return errActionNotImplemented
	case appStart:
		// TODO
		return errActionNotImplemented
	case uiClick:
		// click on coordinate
		if location, ok := action.Params.([]int); ok {
			if len(location) != 2 {
				return fmt.Errorf("invalid click location params: %v", location)
			}
			return w.Driver.Tap(location[0], location[1])
		}
		// click on name or xpath
		if param, ok := action.Params.(string); ok {
			var selector gwda.BySelector
			if strings.HasPrefix(param, "/") {
				// xpath
				selector = gwda.BySelector{
					XPath: param,
				}
			} else {
				// name
				selector = gwda.BySelector{
					Name: param,
				}
			}

			ele, err := w.Driver.FindElement(selector)
			if err != nil {
				return errors.Wrap(err, "failed to find element")
			}
			return ele.Click()
		}
		return fmt.Errorf("invalid click params: %v", action.Params)
	case uiDoubleClick:
		// TODO
		return errActionNotImplemented
	case uiLongClick:
		// TODO
		return errActionNotImplemented
	case uiSwipe:
		width := w.WindowSize.Width
		height := w.WindowSize.Height

		var fromX, fromY, toX, toY int
		if direction, ok := action.Params.(string); ok {
			switch direction {
			case "up":
				fromX, fromY, toX, toY = width/2, height*3/4, width/2, height*1/4
			case "down":
				fromX, fromY, toX, toY = width/2, height*1/4, width/2, height*3/4
			case "left":
				fromX, fromY, toX, toY = width*3/4, height/2, width*1/4, height/2
			case "right":
				fromX, fromY, toX, toY = width*1/4, height/2, width*3/4, height/2
			}
		} else if params, ok := action.Params.([]int); ok {
			if len(params) != 4 {
				return fmt.Errorf("invalid swipe params: %v", params)
			}
			fromX, fromY, toX, toY = params[0], params[1], params[2], params[3]
		} else {
			return fmt.Errorf("invalid swipe params: %v", action.Params)
		}
		return w.Driver.Swipe(fromX, fromY, toX, toY)
	case uiInput:
		// TODO
		return errActionNotImplemented
	case appClick:
		// TODO
		return errActionNotImplemented
	}
	return nil
}

func (w *wdaClient) doValidation(iValidators []interface{}) (validateResults []*ValidationResult, err error) {
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
		if validator.Check != "UI" {
			validataResult.CheckResult = "skip"
			log.Warn().Interface("validator", validator).Msg("skip validator")
			validateResults = append(validateResults, validataResult)
			continue
		}

		expected, ok := validator.Expect.(string)
		if !ok {
			return nil, errors.New("validator expect should be string")
		}

		var result bool
		switch validator.Assert {
		case "xpath_exists":
			result = w.assertXpath(expected, true)
		case "xpath_not_exists":
			result = w.assertXpath(expected, false)
		case "name_exists":
			result = w.assertName(expected, true)
		case "name_not_exists":
			result = w.assertName(expected, false)
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
			err = errors.New("step validation failed")
		}
	}
	return
}

func (w *wdaClient) assertName(name string, exists bool) bool {
	selector := gwda.BySelector{
		Name: name,
	}
	_, err := w.Driver.FindElement(selector)
	return exists == (err == nil)
}

func (w *wdaClient) assertXpath(xpath string, exists bool) bool {
	selector := gwda.BySelector{
		XPath: xpath,
	}
	_, err := w.Driver.FindElement(selector)
	return exists == (err == nil)
}