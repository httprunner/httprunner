package httpboomer

// implements IStep interface
type StepRequestValidation struct {
	*TStep
}

func (req *StepRequestValidation) AssertEqual(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	validator := TValidator{
		Check:      jmesPath,
		Comparator: "equals",
		Expect:     expected,
		Message:    msg,
	}
	req.TStep.Validators = append(req.TStep.Validators, validator)
	return req
}

func (req *StepRequestValidation) ToStruct() *TStep {
	return req.TStep
}

func (req *StepRequestValidation) Run() error {
	return req.TStep.Run()
}
