package hrp

import "fmt"

type AndroidAction struct {
	MobileAction
	Serial  string         `json:"serial,omitempty" yaml:"serial,omitempty"`
	Actions []MobileAction `json:"actions,omitempty" yaml:"actions,omitempty"`
}

// StepAndroid implements IStep interface.
type StepAndroid struct {
	step *TStep
}

func (s *StepAndroid) Serial(serial string) *StepAndroid {
	s.step.Android.Serial = serial
	return &StepAndroid{step: s.step}
}

func (s *StepAndroid) InstallApp(path string) *StepAndroid {
	s.step.Android.Actions = append(s.step.Android.Actions, MobileAction{
		Method: appInstall,
		Params: path,
	})
	return &StepAndroid{step: s.step}
}

func (s *StepAndroid) StartAppByIntent(activity string) *StepAndroid {
	s.step.Android.Actions = append(s.step.Android.Actions, MobileAction{
		Method: appStart,
		Params: activity,
	})
	return &StepAndroid{step: s.step}
}

func (s *StepAndroid) StartCamera() *StepAndroid {
	s.step.Android.Actions = append(s.step.Android.Actions, MobileAction{
		Method: ctlStartCamera,
		Params: nil,
	})
	return &StepAndroid{step: s.step}
}

func (s *StepAndroid) StopCamera() *StepAndroid {
	s.step.Android.Actions = append(s.step.Android.Actions, MobileAction{
		Method: ctlStopCamera,
		Params: nil,
	})
	return &StepAndroid{step: s.step}
}

func (s *StepAndroid) StartRecording() *StepAndroid {
	s.step.Android.Actions = append(s.step.Android.Actions, MobileAction{
		Method: recordStart,
		Params: nil,
	})
	return &StepAndroid{step: s.step}
}

func (s *StepAndroid) StopRecording() *StepAndroid {
	s.step.Android.Actions = append(s.step.Android.Actions, MobileAction{
		Method: recordStop,
		Params: nil,
	})
	return &StepAndroid{step: s.step}
}

func (s *StepAndroid) Click(params interface{}) *StepAndroid {
	s.step.Android.Actions = append(s.step.Android.Actions, MobileAction{
		Method: uiClick,
		Params: params,
	})
	return &StepAndroid{step: s.step}
}

func (s *StepAndroid) DoubleClick(params interface{}) *StepAndroid {
	s.step.Android.Actions = append(s.step.Android.Actions, MobileAction{
		Method: uiDoubleClick,
		Params: params,
	})
	return &StepAndroid{step: s.step}
}

func (s *StepAndroid) LongClick(params interface{}) *StepAndroid {
	s.step.Android.Actions = append(s.step.Android.Actions, MobileAction{
		Method: uiLongClick,
		Params: params,
	})
	return &StepAndroid{step: s.step}
}

func (s *StepAndroid) Swipe(sx, sy, ex, ey int) *StepAndroid {
	s.step.Android.Actions = append(s.step.Android.Actions, MobileAction{
		Method: uiSwipe,
		Params: []int{sx, sy, ex, ey},
	})
	return &StepAndroid{step: s.step}
}

func (s *StepAndroid) SwipeUp() *StepAndroid {
	s.step.Android.Actions = append(s.step.Android.Actions, MobileAction{
		Method: uiSwipe,
		Params: "up",
	})
	return &StepAndroid{step: s.step}
}

func (s *StepAndroid) SwipeDown() *StepAndroid {
	s.step.Android.Actions = append(s.step.Android.Actions, MobileAction{
		Method: uiSwipe,
		Params: "down",
	})
	return &StepAndroid{step: s.step}
}

func (s *StepAndroid) SwipeLeft() *StepAndroid {
	s.step.Android.Actions = append(s.step.Android.Actions, MobileAction{
		Method: uiSwipe,
		Params: "left",
	})
	return &StepAndroid{step: s.step}
}

func (s *StepAndroid) SwipeRight() *StepAndroid {
	s.step.Android.Actions = append(s.step.Android.Actions, MobileAction{
		Method: uiSwipe,
		Params: "right",
	})
	return &StepAndroid{step: s.step}
}

func (s *StepAndroid) Input(text string) *StepAndroid {
	s.step.Android.Actions = append(s.step.Android.Actions, MobileAction{
		Method: uiInput,
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

func (s *StepAndroidValidation) AssertXpathExists(expectedXpath string, msg ...string) *StepAndroidValidation {
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

func (s *StepAndroidValidation) AssertXpathNotExists(expectedXpath string, msg ...string) *StepAndroidValidation {
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

func runStepAndroid(r *SessionRunner, step *TStep) (stepResult *StepResult, err error) {
	stepResult = &StepResult{
		Name:        step.Name,
		StepType:    stepTypeAndroid,
		Success:     false,
		ContentSize: 0,
	}
	return stepResult, nil
}
