package hrp

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/uixt"
)

var (
	WithSerialNumber = uixt.WithSerialNumber
	WithAdbIP        = uixt.WithAdbIP
	WithAdbPort      = uixt.WithAdbPort
	WithAdbLogOn     = uixt.WithAdbLogOn
)

type AndroidStep struct {
	uixt.AndroidDevice `yaml:",inline"` // inline refers to https://pkg.go.dev/gopkg.in/yaml.v3#Marshal
	uixt.MobileAction
	Actions []uixt.MobileAction `json:"actions,omitempty" yaml:"actions,omitempty"`
}

// StepAndroid implements IStep interface.
type StepAndroid struct {
	step *TStep
}

func (s *StepAndroid) Serial(serial string) *StepAndroid {
	s.step.Android.SerialNumber = serial
	return &StepAndroid{step: s.step}
}

func (s *StepAndroid) InstallApp(path string) *StepAndroid {
	s.step.Android.Actions = append(s.step.Android.Actions, uixt.MobileAction{
		Method: uixt.AppInstall,
		Params: path,
	})
	return &StepAndroid{step: s.step}
}

func (s *StepAndroid) AppLaunch(bundleId string) *StepAndroid {
	s.step.Android.Actions = append(s.step.Android.Actions, uixt.MobileAction{
		Method: uixt.AppLaunch,
		Params: bundleId,
	})
	return s
}

func (s *StepAndroid) AppLaunchUnattached(bundleId string) *StepAndroid {
	s.step.Android.Actions = append(s.step.Android.Actions, uixt.MobileAction{
		Method: uixt.AppLaunchUnattached,
		Params: bundleId,
	})
	return s
}

func (s *StepAndroid) AppTerminate(bundleId string) *StepAndroid {
	s.step.Android.Actions = append(s.step.Android.Actions, uixt.MobileAction{
		Method: uixt.AppTerminate,
		Params: bundleId,
	})
	return s
}

func (s *StepAndroid) Home() *StepAndroid {
	s.step.Android.Actions = append(s.step.Android.Actions, uixt.MobileAction{
		Method: uixt.ACTION_Home,
		Params: nil,
	})
	return &StepAndroid{step: s.step}
}

func (s *StepAndroid) StartAppByIntent(activity string) *StepAndroid {
	s.step.Android.Actions = append(s.step.Android.Actions, uixt.MobileAction{
		Method: uixt.AppStart,
		Params: activity,
	})
	return &StepAndroid{step: s.step}
}

func (s *StepAndroid) StartCamera() *StepAndroid {
	s.step.Android.Actions = append(s.step.Android.Actions, uixt.MobileAction{
		Method: uixt.CtlStartCamera,
		Params: nil,
	})
	return &StepAndroid{step: s.step}
}

func (s *StepAndroid) StopCamera() *StepAndroid {
	s.step.Android.Actions = append(s.step.Android.Actions, uixt.MobileAction{
		Method: uixt.CtlStopCamera,
		Params: nil,
	})
	return &StepAndroid{step: s.step}
}

func (s *StepAndroid) StartRecording() *StepAndroid {
	s.step.Android.Actions = append(s.step.Android.Actions, uixt.MobileAction{
		Method: uixt.RecordStart,
		Params: nil,
	})
	return &StepAndroid{step: s.step}
}

func (s *StepAndroid) StopRecording() *StepAndroid {
	s.step.Android.Actions = append(s.step.Android.Actions, uixt.MobileAction{
		Method: uixt.RecordStop,
		Params: nil,
	})
	return &StepAndroid{step: s.step}
}

// TapXY taps the point {X,Y}, X & Y is percentage of coordinates
func (s *StepAndroid) TapXY(x, y float64, options ...uixt.ActionOption) *StepAndroid {
	action := uixt.MobileAction{
		Method: uixt.ACTION_TapXY,
		Params: []float64{x, y},
	}
	for _, option := range options {
		option(&action)
	}
	s.step.Android.Actions = append(s.step.Android.Actions, action)
	return &StepAndroid{step: s.step}
}

// TapAbsXY taps the point {X,Y}, X & Y is absolute coordinates
func (s *StepAndroid) TapAbsXY(x, y float64, options ...uixt.ActionOption) *StepAndroid {
	action := uixt.MobileAction{
		Method: uixt.ACTION_TapAbsXY,
		Params: []float64{x, y},
	}
	for _, option := range options {
		option(&action)
	}
	s.step.Android.Actions = append(s.step.Android.Actions, action)
	return &StepAndroid{step: s.step}
}

func (s *StepAndroid) Tap(params interface{}) *StepAndroid {
	s.step.Android.Actions = append(s.step.Android.Actions, uixt.MobileAction{
		Method: uixt.ACTION_Tap,
		Params: params,
	})
	return &StepAndroid{step: s.step}
}

// Tap taps on the target element by OCR recognition
func (s *StepAndroid) TapByOCR(ocrText string, options ...uixt.ActionOption) *StepAndroid {
	action := uixt.MobileAction{
		Method: uixt.ACTION_TapByOCR,
		Params: ocrText,
	}
	for _, option := range options {
		option(&action)
	}
	s.step.Android.Actions = append(s.step.Android.Actions, action)
	return &StepAndroid{step: s.step}
}

// Tap taps on the target element by CV recognition
func (s *StepAndroid) TapByCV(imagePath string, options ...uixt.ActionOption) *StepAndroid {
	action := uixt.MobileAction{
		Method: uixt.ACTION_TapByCV,
		Params: imagePath,
	}
	for _, option := range options {
		option(&action)
	}
	s.step.Android.Actions = append(s.step.Android.Actions, action)
	return &StepAndroid{step: s.step}
}

func (s *StepAndroid) DoubleTap(params string, options ...uixt.ActionOption) *StepAndroid {
	action := uixt.MobileAction{
		Method: uixt.ACTION_DoubleTap,
		Params: params,
	}
	for _, option := range options {
		option(&action)
	}
	s.step.Android.Actions = append(s.step.Android.Actions, action)
	return &StepAndroid{step: s.step}
}

func (s *StepAndroid) Swipe(sx, sy, ex, ey int, options ...uixt.ActionOption) *StepAndroid {
	action := uixt.MobileAction{
		Method: uixt.ACTION_Swipe,
		Params: []int{sx, sy, ex, ey},
	}
	for _, option := range options {
		option(&action)
	}
	s.step.Android.Actions = append(s.step.Android.Actions, action)
	return &StepAndroid{step: s.step}
}

func (s *StepAndroid) SwipeUp(options ...uixt.ActionOption) *StepAndroid {
	action := uixt.MobileAction{
		Method: uixt.ACTION_Swipe,
		Params: "up",
	}
	for _, option := range options {
		option(&action)
	}
	s.step.Android.Actions = append(s.step.Android.Actions, action)
	return &StepAndroid{step: s.step}
}

func (s *StepAndroid) SwipeDown(options ...uixt.ActionOption) *StepAndroid {
	action := uixt.MobileAction{
		Method: uixt.ACTION_Swipe,
		Params: "down",
	}
	for _, option := range options {
		option(&action)
	}
	s.step.Android.Actions = append(s.step.Android.Actions, action)
	return &StepAndroid{step: s.step}
}

func (s *StepAndroid) SwipeLeft(options ...uixt.ActionOption) *StepAndroid {
	action := uixt.MobileAction{
		Method: uixt.ACTION_Swipe,
		Params: "left",
	}
	for _, option := range options {
		option(&action)
	}
	s.step.Android.Actions = append(s.step.Android.Actions, action)
	return &StepAndroid{step: s.step}
}

func (s *StepAndroid) SwipeRight(options ...uixt.ActionOption) *StepAndroid {
	action := uixt.MobileAction{
		Method: uixt.ACTION_Swipe,
		Params: "right",
	}
	for _, option := range options {
		option(&action)
	}
	s.step.Android.Actions = append(s.step.Android.Actions, action)
	return &StepAndroid{step: s.step}
}

func (s *StepAndroid) Input(text string) *StepAndroid {
	s.step.Android.Actions = append(s.step.Android.Actions, uixt.MobileAction{
		Method: uixt.ACTION_Input,
		Params: text,
	})
	return &StepAndroid{step: s.step}
}

// Sleep specify sleep seconds after last action
func (s *StepAndroid) Sleep(n float64) *StepAndroid {
	s.step.Android.Actions = append(s.step.Android.Actions, uixt.MobileAction{
		Method: uixt.CtlSleep,
		Params: n,
	})
	return &StepAndroid{step: s.step}
}

func (s *StepAndroid) ScreenShot() *StepAndroid {
	s.step.Android.Actions = append(s.step.Android.Actions, uixt.MobileAction{
		Method: uixt.CtlScreenShot,
		Params: nil,
	})
	return &StepAndroid{step: s.step}
}

func (s *StepAndroid) SwipeToTapApp(appName string, options ...uixt.ActionOption) *StepAndroid {
	action := uixt.MobileAction{
		Method: uixt.ACTION_SwipeToTapApp,
		Params: appName,
	}
	for _, option := range options {
		option(&action)
	}
	s.step.Android.Actions = append(s.step.Android.Actions, action)
	return &StepAndroid{step: s.step}
}

func (s *StepAndroid) SwipeToTapText(text string, options ...uixt.ActionOption) *StepAndroid {
	action := uixt.MobileAction{
		Method: uixt.ACTION_SwipeToTapText,
		Params: text,
	}
	for _, option := range options {
		option(&action)
	}
	s.step.Android.Actions = append(s.step.Android.Actions, action)
	return &StepAndroid{step: s.step}
}

// Validate switches to step validation.
func (s *StepAndroid) Validate() *StepAndroidValidation {
	return &StepAndroidValidation{
		step: s.step,
	}
}

func (s *StepAndroid) Name() string {
	return s.step.Name
}

func (s *StepAndroid) Type() StepType {
	return stepTypeAndroid
}

func (s *StepAndroid) Struct() *TStep {
	return s.step
}

func (s *StepAndroid) Run(r *SessionRunner) (*StepResult, error) {
	return runStepAndroid(r, s.step)
}

// StepAndroidValidation implements IStep interface.
type StepAndroidValidation struct {
	step *TStep
}

func (s *StepAndroidValidation) AssertNameExists(expectedName string, msg ...string) *StepAndroidValidation {
	v := Validator{
		Check:  uixt.SelectorName,
		Assert: uixt.AssertionExists,
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

func (s *StepAndroidValidation) AssertNameNotExists(expectedName string, msg ...string) *StepAndroidValidation {
	v := Validator{
		Check:  uixt.SelectorName,
		Assert: uixt.AssertionNotExists,
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

func (s *StepAndroidValidation) AssertLabelExists(expectedLabel string, msg ...string) *StepAndroidValidation {
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

func (s *StepAndroidValidation) AssertLabelNotExists(expectedLabel string, msg ...string) *StepAndroidValidation {
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

func (s *StepAndroidValidation) AssertOCRExists(expectedText string, msg ...string) *StepAndroidValidation {
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

func (s *StepAndroidValidation) AssertOCRNotExists(expectedText string, msg ...string) *StepAndroidValidation {
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

func (s *StepAndroidValidation) AssertImageExists(expectedImagePath string, msg ...string) *StepAndroidValidation {
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

func (s *StepAndroidValidation) AssertImageNotExists(expectedImagePath string, msg ...string) *StepAndroidValidation {
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

func (s *StepAndroidValidation) Name() string {
	return s.step.Name
}

func (s *StepAndroidValidation) Type() StepType {
	return stepTypeAndroid
}

func (s *StepAndroidValidation) Struct() *TStep {
	return s.step
}

func (s *StepAndroidValidation) Run(r *SessionRunner) (*StepResult, error) {
	return runStepAndroid(r, s.step)
}

func runStepAndroid(s *SessionRunner, step *TStep) (stepResult *StepResult, err error) {
	stepResult = &StepResult{
		Name:        step.Name,
		StepType:    stepTypeAndroid,
		Success:     false,
		ContentSize: 0,
	}
	screenshots := make([]string, 0)

	// init uiaClient driver
	uiaClient, err := s.hrpRunner.initUIClient(&step.Android.AndroidDevice)
	if err != nil {
		return
	}
	uiaClient.StartTime = s.startTime

	defer func() {
		attachments := make(map[string]interface{})
		if err != nil {
			attachments["error"] = err.Error()
		}

		// save attachments
		screenshots = append(screenshots, uiaClient.ScreenShots...)
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
	if step.Android.Actions == nil {
		actions = []uixt.MobileAction{
			{
				Method: step.Android.Method,
				Params: step.Android.Params,
			},
		}
	} else {
		actions = step.Android.Actions
	}

	// run actions
	for _, action := range actions {
		if err := uiaClient.DoAction(action); err != nil {
			return stepResult, err
		}
	}

	// take snapshot
	screenshotPath, err := uiaClient.ScreenShot(
		fmt.Sprintf("%d_validate_%d", uiaClient.StartTime.Unix(), time.Now().Unix()))
	if err != nil {
		log.Warn().Err(err).Str("step", step.Name).Msg("take screenshot failed")
	} else {
		log.Info().Str("path", screenshotPath).Msg("take screenshot before validation")
		screenshots = append(screenshots, screenshotPath)
	}

	// validate
	validateResults, err := validateUI(uiaClient, step.Validators)
	if err != nil {
		return
	}
	sessionData := newSessionData()
	sessionData.Validators = validateResults
	stepResult.Data = sessionData
	stepResult.Success = true
	return stepResult, nil
}
