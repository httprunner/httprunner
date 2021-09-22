package httpboomer

// implements IStep interface
type stepRequestValidation struct {
	runner *Runner
	TStep  *TStep
}

func (s *stepRequestValidation) AssertEqual(jmesPath string, expected interface{}, msg string) *stepRequestValidation {
	validator := TValidator{
		Check:      jmesPath,
		Comparator: "equals",
		Expect:     expected,
		Message:    msg,
	}
	s.TStep.Validators = append(s.TStep.Validators, validator)
	return s
}

func (s *stepRequestValidation) ToStruct() *TStep {
	return s.TStep
}

func (s *stepRequestValidation) Run() error {
	return s.runner.runStep(s.TStep)
}
