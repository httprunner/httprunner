package httpboomer

// implements IStep interface
type stepRequestExtraction struct {
	runner *Runner
	step   *TStep
}

func (s *stepRequestExtraction) WithJmesPath(jmesPath string, varName string) *stepRequestExtraction {
	s.step.Extract[varName] = jmesPath
	return s
}

func (s *stepRequestExtraction) Validate() *stepRequestValidation {
	return &stepRequestValidation{
		TStep: s.step,
	}
}

func (s *stepRequestExtraction) ToStruct() *TStep {
	return s.step
}

func (s *stepRequestExtraction) Run() error {
	return s.runner.runStep(s.step)
}
