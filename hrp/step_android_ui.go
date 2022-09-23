package hrp

import (
	"fmt"
	"time"

	"github.com/httprunner/httprunner/v4/hrp/internal/uixt"
	"github.com/rs/zerolog/log"
)

type AndroidStep struct {
	uixt.UIAOptions `yaml:",inline"` // inline refers to https://pkg.go.dev/gopkg.in/yaml.v3#Marshal
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

func (s *StepAndroid) Tap(params interface{}) *StepAndroid {
	s.step.Android.Actions = append(s.step.Android.Actions, uixt.MobileAction{
		Method: uixt.ACTION_Tap,
		Params: params,
	})
	return &StepAndroid{step: s.step}
}

func (s *StepAndroid) DoubleTap(params interface{}) *StepAndroid {
	s.step.Android.Actions = append(s.step.Android.Actions, uixt.MobileAction{
		Method: uixt.ACTION_DoubleTap,
		Params: params,
	})
	return &StepAndroid{step: s.step}
}

func (s *StepAndroid) Swipe(sx, sy, ex, ey int) *StepAndroid {
	s.step.Android.Actions = append(s.step.Android.Actions, uixt.MobileAction{
		Method: uixt.ACTION_Swipe,
		Params: []int{sx, sy, ex, ey},
	})
	return &StepAndroid{step: s.step}
}

func (s *StepAndroid) SwipeUp() *StepAndroid {
	s.step.Android.Actions = append(s.step.Android.Actions, uixt.MobileAction{
		Method: uixt.ACTION_Swipe,
		Params: "up",
	})
	return &StepAndroid{step: s.step}
}

func (s *StepAndroid) SwipeDown() *StepAndroid {
	s.step.Android.Actions = append(s.step.Android.Actions, uixt.MobileAction{
		Method: uixt.ACTION_Swipe,
		Params: "down",
	})
	return &StepAndroid{step: s.step}
}

func (s *StepAndroid) SwipeLeft() *StepAndroid {
	s.step.Android.Actions = append(s.step.Android.Actions, uixt.MobileAction{
		Method: uixt.ACTION_Swipe,
		Params: "left",
	})
	return &StepAndroid{step: s.step}
}

func (s *StepAndroid) SwipeRight() *StepAndroid {
	s.step.Android.Actions = append(s.step.Android.Actions, uixt.MobileAction{
		Method: uixt.ACTION_Swipe,
		Params: "right",
	})
	return &StepAndroid{step: s.step}
}

func (s *StepAndroid) Input(text string) *StepAndroid {
	s.step.Android.Actions = append(s.step.Android.Actions, uixt.MobileAction{
		Method: uixt.ACTION_Input,
		Params: text,
	})
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
	uiaClient, err := s.hrpRunner.initUIClient(&step.Android.UIAOptions)
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
