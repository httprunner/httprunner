package hrp

import "fmt"

// StepRequestExtraction implements IStep interface.
type StepRequestExtraction struct {
	step *TStep
}

// WithJmesPath sets the JMESPath expression to extract from the response.
func (s *StepRequestExtraction) WithJmesPath(jmesPath string, varName string) *StepRequestExtraction {
	s.step.Extract[varName] = jmesPath
	return s
}

// Validate switches to step validation.
func (s *StepRequestExtraction) Validate() *StepRequestValidation {
	return &StepRequestValidation{
		step: s.step,
	}
}

func (s *StepRequestExtraction) Name() string {
	return s.step.Name
}

func (s *StepRequestExtraction) Type() string {
	return fmt.Sprintf("request-%v", s.step.Request.Method)
}

func (s *StepRequestExtraction) ToStruct() *TStep {
	return s.step
}
