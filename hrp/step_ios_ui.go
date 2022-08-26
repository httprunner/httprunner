package hrp

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	gwdaExt "github.com/debugtalk/gwda-ext"
	"github.com/electricbubble/gwda"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

const (
	// Changes the value of maximum depth for traversing elements source tree.
	// It may help to prevent out of memory or timeout errors while getting the elements source tree,
	// but it might restrict the depth of source tree.
	// A part of elements source tree might be lost if the value was too small. Defaults to 50
	snapshotMaxDepth = 10
	// Allows to customize accept/dismiss alert button selector.
	// It helps you to handle an arbitrary element as accept button in accept alert command.
	// The selector should be a valid class chain expression, where the search root is the alert element itself.
	// The default button location algorithm is used if the provided selector is wrong or does not match any element.
	// e.g. **/XCUIElementTypeButton[`label CONTAINS[c] ‘accept’`]
	acceptAlertButtonSelector  = "**/XCUIElementTypeButton[`label IN {'允许','好','仅在使用应用期间','稍后再说'}`]"
	dismissAlertButtonSelector = "**/XCUIElementTypeButton[`label IN {'不允许','暂不'}`]"
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

func (s *StepIOSValidation) AssertXpathExists(expectedXpath string, msg ...string) *StepIOSValidation {
	v := Validator{
		Check:  uiSelectorXpath,
		Assert: assertionExists,
		Expect: expectedXpath,
	}
	if len(msg) > 0 {
		v.Message = msg[0]
	} else {
		v.Message = fmt.Sprintf("xpath [%s] not found", expectedXpath)
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepIOSValidation) AssertXpathNotExists(expectedXpath string, msg ...string) *StepIOSValidation {
	v := Validator{
		Check:  uiSelectorXpath,
		Assert: assertionNotExists,
		Expect: expectedXpath,
	}
	if len(msg) > 0 {
		v.Message = msg[0]
	} else {
		v.Message = fmt.Sprintf("xpath [%s] should not exist", expectedXpath)
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

func (r *HRPRunner) InitWDAClient(device WDADevice) (client *wdaClient, err error) {
	defer func() {
		if err != nil {
			return
		}
		// check if WDA is healthy
		ok, e := client.DriverExt.IsWdaHealthy()
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
	if device.UDID == "" && len(r.wdaClients) == 1 {
		for _, v := range r.wdaClients {
			return v, nil
		}
	}

	// init wda device
	var options []gwda.DeviceOption
	if device.UDID != "" {
		options = append(options, gwda.WithSerialNumber(device.UDID))
	}
	if device.Port != 0 {
		options = append(options, gwda.WithPort(device.Port))
	}
	if device.MjpegPort != 0 {
		options = append(options, gwda.WithMjpegPort(device.MjpegPort))
	}
	targetDevice, err := gwda.NewDevice(options...)
	if err != nil {
		return nil, err
	}

	// avoid duplicate init
	if client, ok := r.wdaClients[targetDevice.SerialNumber()]; ok {
		return client, nil
	}

	// switch to iOS springboard before init WDA session
	// aviod getting stuck when some super app is activate such as douyin or wexin
	log.Info().Msg("switch to iOS springboard")
	bundleID := "com.apple.springboard"
	_, err = targetDevice.GIDevice().AppLaunch(bundleID)
	if err != nil {
		return nil, errors.Wrap(err, "launch springboard failed")
	}

	// init WDA driver
	gwda.SetDebug(true)
	capabilities := gwda.NewCapabilities()
	capabilities.WithDefaultAlertAction(gwda.AlertActionAccept)
	driver, err := gwda.NewUSBDriver(capabilities, *targetDevice)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init WDA driver")
	}
	driverExt, err := gwdaExt.Extend(driver, 0.95)
	if err != nil {
		return nil, errors.Wrap(err, "failed to extend gwda.WebDriver")
	}
	settings, err := driverExt.SetAppiumSettings(map[string]interface{}{
		"snapshotMaxDepth":          snapshotMaxDepth,
		"acceptAlertButtonSelector": acceptAlertButtonSelector,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to set appium WDA settings")
	}
	log.Info().Interface("appiumWDASettings", settings).Msg("set appium WDA settings")

	// get device window size
	windowSize, err := driverExt.WindowSize()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get windows size")
	}

	// cache wda client
	r.wdaClients = make(map[string]*wdaClient)
	client = &wdaClient{
		ID:         time.Now().Unix(),
		Device:     targetDevice,
		DriverExt:  driverExt,
		WindowSize: windowSize,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
	r.wdaClients[targetDevice.SerialNumber()] = client

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

type wdaClient struct {
	ID         int64
	Device     *gwda.Device
	DriverExt  *gwdaExt.DriverExt
	WindowSize gwda.Size
	httpClient *http.Client
}

func (w *wdaClient) doAction(action MobileAction) error {
	log.Info().Str("method", string(action.Method)).Interface("params", action.Params).Msg("start iOS UI action")

	switch action.Method {
	case appInstall:
		// TODO
		return errActionNotImplemented
	case appLaunch:
		if bundleId, ok := action.Params.(string); ok {
			return w.DriverExt.AppLaunch(bundleId)
		}
		return fmt.Errorf("app_launch params should be bundleId(string), got %v", action.Params)
	case appLaunchUnattached:
		if bundleId, ok := action.Params.(string); ok {
			return w.DriverExt.AppLaunchUnattached(bundleId)
		}
		return fmt.Errorf("app_launch_unattached params should be bundleId(string), got %v", action.Params)
	case appTerminate:
		if bundleId, ok := action.Params.(string); ok {
			success, err := w.DriverExt.AppTerminate(bundleId)
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
		return w.DriverExt.Homescreen()
	case uiClick:
		// click on coordinate
		if location, ok := action.Params.([]int); ok {
			// absolute x,y
			if len(location) != 2 {
				return fmt.Errorf("invalid click location params: %v", location)
			}
			return w.DriverExt.WebDriver.Tap(location[0], location[1])
		}
		if location, ok := action.Params.([]float64); ok {
			// relative x,y of window size
			if len(location) != 2 {
				return fmt.Errorf("invalid click location params: %v", location)
			}
			x := location[0] * float64(w.WindowSize.Width)
			y := location[1] * float64(w.WindowSize.Height)
			return w.DriverExt.TapFloat(x, y)
		}
		// click on name or xpath
		if param, ok := action.Params.(string); ok {
			ele, err := w.findElement(param)
			if err != nil {
				return errors.Wrap(err, "failed to find element")
			}
			return ele.Click()
		}
		return fmt.Errorf("invalid click params: %v", action.Params)
	case uiDoubleClick:
		// double click on name or xpath
		if param, ok := action.Params.(string); ok {
			ele, err := w.findElement(param)
			if err != nil {
				return errors.Wrap(err, "failed to find element")
			}
			return ele.DoubleTap()
		}
		return fmt.Errorf("invalid click params: %v", action.Params)
	case uiLongClick:
		// long click 2s on name or xpath
		if param, ok := action.Params.(string); ok {
			ele, err := w.findElement(param)
			if err != nil {
				return errors.Wrap(err, "failed to find element")
			}
			return ele.TouchAndHold(2)
		}
		return fmt.Errorf("invalid click params: %v", action.Params)
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
		return w.DriverExt.WebDriver.Swipe(fromX, fromY, toX, toY)
	case uiInput:
		// input text on current active element
		// append \n to send text with enter
		// send \b\b\b to delete 3 chars
		param := fmt.Sprintf("%v", action.Params)
		return w.DriverExt.SendKeys(param)
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
			screenshotPath, err = w.DriverExt.ScreenShot(fmt.Sprintf("screenshot_%s", param))
		} else {
			screenshotPath, err = w.DriverExt.ScreenShot(fmt.Sprintf("screenshot_%d", time.Now().Unix()))
		}
		log.Info().Str("path", screenshotPath).Msg("take screenshot")
		return err
	case ctlStartCamera:
		// start camera, alias for app_launch com.apple.camera
		return w.DriverExt.AppLaunch("com.apple.camera")
	case ctlStopCamera:
		// stop camera, alias for app_terminate com.apple.camera
		success, err := w.DriverExt.AppTerminate("com.apple.camera")
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
			result = w.assertName(expected, exists)
		case uiSelectorXpath:
			result = w.assertXpath(expected, exists)
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

func (w *wdaClient) findElement(param string) (ele gwda.WebElement, err error) {
	var selector gwda.BySelector
	if strings.HasPrefix(param, "/") {
		// xpath
		selector = gwda.BySelector{
			XPath: param,
		}
	} else {
		// name
		selector = gwda.BySelector{
			LinkText: gwda.NewElementAttribute().WithName(param),
		}
	}

	return w.DriverExt.FindElement(selector)
}

func (w *wdaClient) assertName(name string, exists bool) bool {
	selector := gwda.BySelector{
		LinkText: gwda.NewElementAttribute().WithName(name),
	}
	_, err := w.DriverExt.FindElement(selector)
	return exists == (err == nil)
}

func (w *wdaClient) assertXpath(xpath string, exists bool) bool {
	selector := gwda.BySelector{
		XPath: xpath,
	}
	_, err := w.DriverExt.FindElement(selector)
	return exists == (err == nil)
}
