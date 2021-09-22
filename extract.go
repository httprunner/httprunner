package httpboomer

// implements IStep interface
type StepRequestExtraction struct {
	*TStep
}

func (step *StepRequestExtraction) WithJmesPath(jmesPath string, varName string) *StepRequestExtraction {
	step.TStep.Extract[varName] = jmesPath
	return step
}

func (step *StepRequestExtraction) Validate() *StepRequestValidation {
	return &StepRequestValidation{
		TStep: step.TStep,
	}
}

func (step *StepRequestExtraction) ToStruct() *TStep {
	return step.TStep
}

func (step *StepRequestExtraction) Run() error {
	return step.TStep.Run()
}
