package httpboomer

// implements IStep interface
type StepRequestExtraction struct {
	*TStep
}

func (req *StepRequestExtraction) WithJmesPath(jmesPath string, varName string) *StepRequestExtraction {
	req.TStep.Extract[varName] = jmesPath
	return req
}

func (req *StepRequestExtraction) Validate() *StepRequestValidation {
	return &StepRequestValidation{
		TStep: req.TStep,
	}
}

func (req *StepRequestExtraction) ToStruct() *TStep {
	return req.TStep
}

func (req *StepRequestExtraction) Run() error {
	return req.TStep.Run()
}
