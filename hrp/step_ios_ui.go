package hrp

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/pkg/uixt"
)

var (
	WithUDID         = uixt.WithUDID
	WithWDAPort      = uixt.WithWDAPort
	WithWDAMjpegPort = uixt.WithWDAMjpegPort
	WithLogOn        = uixt.WithLogOn
	WithPerfOptions  = uixt.WithPerfOptions
)

type IOSStep struct {
	uixt.IOSDevice    `yaml:",inline"` // inline refers to https://pkg.go.dev/gopkg.in/yaml.v3#Marshal
	uixt.MobileAction `yaml:",inline"`
	Actions           []uixt.MobileAction `json:"actions,omitempty" yaml:"actions,omitempty"`
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
	s.step.IOS.Actions = append(s.step.IOS.Actions, uixt.MobileAction{
		Method: uixt.AppInstall,
		Params: path,
	})
	return s
}

func (s *StepIOS) AppLaunch(bundleId string) *StepIOS {
	s.step.IOS.Actions = append(s.step.IOS.Actions, uixt.MobileAction{
		Method: uixt.AppLaunch,
		Params: bundleId,
	})
	return s
}

func (s *StepIOS) AppLaunchUnattached(bundleId string) *StepIOS {
	s.step.IOS.Actions = append(s.step.IOS.Actions, uixt.MobileAction{
		Method: uixt.AppLaunchUnattached,
		Params: bundleId,
	})
	return s
}

func (s *StepIOS) AppTerminate(bundleId string) *StepIOS {
	s.step.IOS.Actions = append(s.step.IOS.Actions, uixt.MobileAction{
		Method: uixt.AppTerminate,
		Params: bundleId,
	})
	return s
}

func (s *StepIOS) Home() *StepIOS {
	s.step.IOS.Actions = append(s.step.IOS.Actions, uixt.MobileAction{
		Method: uixt.ACTION_Home,
		Params: nil,
	})
	return &StepIOS{step: s.step}
}

// TapXY taps the point {X,Y}, X & Y is percentage of coordinates
func (s *StepIOS) TapXY(x, y float64, options ...uixt.ActionOption) *StepIOS {
	action := uixt.MobileAction{
		Method: uixt.ACTION_TapXY,
		Params: []float64{x, y},
	}
	for _, option := range options {
		option(&action)
	}
	s.step.IOS.Actions = append(s.step.IOS.Actions, action)
	return &StepIOS{step: s.step}
}

// TapAbsXY taps the point {X,Y}, X & Y is absolute coordinates
func (s *StepIOS) TapAbsXY(x, y float64, options ...uixt.ActionOption) *StepIOS {
	action := uixt.MobileAction{
		Method: uixt.ACTION_TapAbsXY,
		Params: []float64{x, y},
	}
	for _, option := range options {
		option(&action)
	}
	s.step.IOS.Actions = append(s.step.IOS.Actions, action)
	return &StepIOS{step: s.step}
}

// Tap taps on the target element
func (s *StepIOS) Tap(params string, options ...uixt.ActionOption) *StepIOS {
	action := uixt.MobileAction{
		Method: uixt.ACTION_Tap,
		Params: params,
	}
	for _, option := range options {
		option(&action)
	}
	s.step.IOS.Actions = append(s.step.IOS.Actions, action)
	return &StepIOS{step: s.step}
}

// Tap taps on the target element by OCR recognition
func (s *StepIOS) TapByOCR(ocrText string, options ...uixt.ActionOption) *StepIOS {
	action := uixt.MobileAction{
		Method: uixt.ACTION_TapByOCR,
		Params: ocrText,
	}
	for _, option := range options {
		option(&action)
	}
	s.step.IOS.Actions = append(s.step.IOS.Actions, action)
	return &StepIOS{step: s.step}
}

// Tap taps on the target element by CV recognition
func (s *StepIOS) TapByCV(imagePath string, options ...uixt.ActionOption) *StepIOS {
	action := uixt.MobileAction{
		Method: uixt.ACTION_TapByCV,
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
	s.step.IOS.Actions = append(s.step.IOS.Actions, uixt.MobileAction{
		Method: uixt.ACTION_DoubleTapXY,
		Params: []float64{x, y},
	})
	return &StepIOS{step: s.step}
}

func (s *StepIOS) DoubleTap(params string, options ...uixt.ActionOption) *StepIOS {
	action := uixt.MobileAction{
		Method: uixt.ACTION_DoubleTap,
		Params: params,
	}
	for _, option := range options {
		option(&action)
	}
	s.step.IOS.Actions = append(s.step.IOS.Actions, action)
	return &StepIOS{step: s.step}
}

func (s *StepIOS) Swipe(sx, sy, ex, ey float64, options ...uixt.ActionOption) *StepIOS {
	action := uixt.MobileAction{
		Method: uixt.ACTION_Swipe,
		Params: []float64{sx, sy, ex, ey},
	}
	for _, option := range options {
		option(&action)
	}
	s.step.IOS.Actions = append(s.step.IOS.Actions, action)
	return &StepIOS{step: s.step}
}

func (s *StepIOS) SwipeUp(options ...uixt.ActionOption) *StepIOS {
	action := uixt.MobileAction{
		Method: uixt.ACTION_Swipe,
		Params: "up",
	}
	for _, option := range options {
		option(&action)
	}
	s.step.IOS.Actions = append(s.step.IOS.Actions, action)
	return &StepIOS{step: s.step}
}

func (s *StepIOS) SwipeDown(options ...uixt.ActionOption) *StepIOS {
	action := uixt.MobileAction{
		Method: uixt.ACTION_Swipe,
		Params: "down",
	}
	for _, option := range options {
		option(&action)
	}
	s.step.IOS.Actions = append(s.step.IOS.Actions, action)
	return &StepIOS{step: s.step}
}

func (s *StepIOS) SwipeLeft(options ...uixt.ActionOption) *StepIOS {
	action := uixt.MobileAction{
		Method: uixt.ACTION_Swipe,
		Params: "left",
	}
	for _, option := range options {
		option(&action)
	}
	s.step.IOS.Actions = append(s.step.IOS.Actions, action)
	return &StepIOS{step: s.step}
}

func (s *StepIOS) SwipeRight(options ...uixt.ActionOption) *StepIOS {
	action := uixt.MobileAction{
		Method: uixt.ACTION_Swipe,
		Params: "right",
	}
	for _, option := range options {
		option(&action)
	}
	s.step.IOS.Actions = append(s.step.IOS.Actions, action)
	return &StepIOS{step: s.step}
}

func (s *StepIOS) SwipeToTapApp(appName string, options ...uixt.ActionOption) *StepIOS {
	action := uixt.MobileAction{
		Method: uixt.ACTION_SwipeToTapApp,
		Params: appName,
	}
	for _, option := range options {
		option(&action)
	}
	s.step.IOS.Actions = append(s.step.IOS.Actions, action)
	return &StepIOS{step: s.step}
}

func (s *StepIOS) SwipeToTapText(text string, options ...uixt.ActionOption) *StepIOS {
	action := uixt.MobileAction{
		Method: uixt.ACTION_SwipeToTapText,
		Params: text,
	}
	for _, option := range options {
		option(&action)
	}
	s.step.IOS.Actions = append(s.step.IOS.Actions, action)
	return &StepIOS{step: s.step}
}

func (s *StepIOS) SwipeToTapTexts(texts []string, options ...uixt.ActionOption) *StepIOS {
	action := uixt.MobileAction{
		Method: uixt.ACTION_SwipeToTapTexts,
		Params: texts,
	}
	for _, option := range options {
		option(&action)
	}
	s.step.IOS.Actions = append(s.step.IOS.Actions, action)
	return &StepIOS{step: s.step}
}

func (s *StepIOS) Input(text string, options ...uixt.ActionOption) *StepIOS {
	action := uixt.MobileAction{
		Method: uixt.ACTION_Input,
		Params: text,
	}
	for _, option := range options {
		option(&action)
	}
	s.step.IOS.Actions = append(s.step.IOS.Actions, action)
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
	s.step.IOS.Actions = append(s.step.IOS.Actions, uixt.MobileAction{
		Method: uixt.CtlSleep,
		Params: n,
	})
	return &StepIOS{step: s.step}
}

func (s *StepIOS) ScreenShot() *StepIOS {
	s.step.IOS.Actions = append(s.step.IOS.Actions, uixt.MobileAction{
		Method: uixt.CtlScreenShot,
		Params: nil,
	})
	return &StepIOS{step: s.step}
}

func (s *StepIOS) StartCamera() *StepIOS {
	s.step.IOS.Actions = append(s.step.IOS.Actions, uixt.MobileAction{
		Method: uixt.CtlStartCamera,
		Params: nil,
	})
	return &StepIOS{step: s.step}
}

func (s *StepIOS) StopCamera() *StepIOS {
	s.step.IOS.Actions = append(s.step.IOS.Actions, uixt.MobileAction{
		Method: uixt.CtlStopCamera,
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
	return stepTypeIOS
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

func (s *StepIOSValidation) AssertNameNotExists(expectedName string, msg ...string) *StepIOSValidation {
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

func (s *StepIOSValidation) AssertLabelExists(expectedLabel string, msg ...string) *StepIOSValidation {
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

func (s *StepIOSValidation) AssertLabelNotExists(expectedLabel string, msg ...string) *StepIOSValidation {
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

func (s *StepIOSValidation) AssertOCRExists(expectedText string, msg ...string) *StepIOSValidation {
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

func (s *StepIOSValidation) AssertOCRNotExists(expectedText string, msg ...string) *StepIOSValidation {
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

func (s *StepIOSValidation) AssertImageExists(expectedImagePath string, msg ...string) *StepIOSValidation {
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

func (s *StepIOSValidation) AssertImageNotExists(expectedImagePath string, msg ...string) *StepIOSValidation {
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

func (s *StepIOSValidation) Name() string {
	return s.step.Name
}

func (s *StepIOSValidation) Type() StepType {
	return stepTypeIOS
}

func (s *StepIOSValidation) Struct() *TStep {
	return s.step
}

func (s *StepIOSValidation) Run(r *SessionRunner) (*StepResult, error) {
	return runStepIOS(r, s.step)
}

func (r *HRPRunner) initUIClient(device uixt.Device) (client *uixt.DriverExt, err error) {
	uuid := device.UUID()

	// avoid duplicate init
	if uuid == "" && len(r.uiClients) > 0 {
		for _, v := range r.uiClients {
			return v, nil
		}
	}

	// avoid duplicate init
	if uuid != "" {
		if client, ok := r.uiClients[uuid]; ok {
			return client, nil
		}
	}

	client, err = device.NewDriver(nil)
	if err != nil {
		return nil, err
	}

	// cache wda client
	if r.uiClients == nil {
		r.uiClients = make(map[string]*uixt.DriverExt)
	}
	r.uiClients[client.UUID] = client

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

	// override step variables
	stepVariables, err := s.MergeStepVariables(step.Variables)
	if err != nil {
		return
	}
	parser := s.GetParser()

	// parse device udid
	if step.IOS.IOSDevice.UDID != "" {
		udid, err := parser.ParseString(step.IOS.IOSDevice.UDID, stepVariables)
		if err != nil {
			return stepResult, err
		}
		step.IOS.IOSDevice.UDID = udid.(string)
	}

	// init wdaClient driver
	wdaClient, err := s.hrpRunner.initUIClient(&step.IOS.IOSDevice)
	if err != nil {
		return
	}
	wdaClient.StartTime = s.startTime

	defer func() {
		attachments := make(map[string]interface{})
		if err != nil {
			attachments["error"] = err.Error()
		}

		// save attachments
		screenshots = append(screenshots, wdaClient.ScreenShots...)
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
	var actions []uixt.MobileAction
	if step.IOS.Actions == nil {
		actions = []uixt.MobileAction{
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
		if action.Params, err = parser.Parse(action.Params, stepVariables); err != nil {
			return stepResult, errors.Wrap(err, "parse action params failed")
		}
		if err := wdaClient.DoAction(action); err != nil {
			return stepResult, err
		}
	}

	// take snapshot
	screenshotPath, err := wdaClient.ScreenShot(
		fmt.Sprintf("%d_validate_%d", wdaClient.StartTime.Unix(), time.Now().Unix()))
	if err != nil {
		log.Warn().Err(err).Str("step", step.Name).Msg("take screenshot failed")
	} else {
		log.Info().Str("path", screenshotPath).Msg("take screenshot before validation")
		screenshots = append(screenshots, screenshotPath)
	}

	// validate
	validateResults, err := validateUI(wdaClient, step.Validators)
	if err != nil {
		return
	}
	sessionData := newSessionData()
	sessionData.Validators = validateResults
	stepResult.Data = sessionData
	stepResult.Success = true
	return stepResult, nil
}
