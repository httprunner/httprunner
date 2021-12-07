package hrp

import "fmt"

// implements IStep interface
type stepRequestExtraction struct {
	step *TStep
}

// WithJmesPath sets the JMESPath expression to extract from the response.
func (s *stepRequestExtraction) WithJmesPath(jmesPath string, varName string) *stepRequestExtraction {
	s.step.Extract[varName] = jmesPath
	return s
}

// Validate switches to step validation.
func (s *stepRequestExtraction) Validate() *stepRequestValidation {
	return &stepRequestValidation{
		step: s.step,
	}
}

func (s *stepRequestExtraction) Name() string {
	return s.step.Name
}

func (s *stepRequestExtraction) Type() string {
	return fmt.Sprintf("request-%v", s.step.Request.Method)
}

func (s *stepRequestExtraction) ToStruct() *TStep {
	return s.step
}
