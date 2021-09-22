package httpboomer

// implements IStep interface
type StepRequestValidation struct {
	*TStep
}

func (step *StepRequestValidation) AssertEqual(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	validator := TValidator{
		Check:      jmesPath,
		Comparator: "equals",
		Expect:     expected,
		Message:    msg,
	}
	step.TStep.Validators = append(step.TStep.Validators, validator)
	return step
}

func (step *StepRequestValidation) ToStruct() *TStep {
	return step.TStep
}

func (step *StepRequestValidation) Run() error {
	return step.TStep.Run()
}
