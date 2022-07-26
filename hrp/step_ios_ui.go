package hrp

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

func (s *StepIOSValidation) AssertTextExists(expectedText string, msg string) *StepIOSValidation {
	v := Validator{
		Check:   "ios_ui",
		Assert:  "text_exists",
		Expect:  expectedText,
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

func runStepIOS(r *SessionRunner, step *TStep) (stepResult *StepResult, err error) {
	stepResult = &StepResult{
		Name:        step.Name,
		StepType:    stepTypeAndroid,
		Success:     false,
		ContentSize: 0,
	}
	return stepResult, nil
}
