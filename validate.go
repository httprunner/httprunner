package httpboomer

import "fmt"

// implements IStep interface
type stepRequestValidation struct {
	runner *Runner
	step   *TStep
}

func (s *stepRequestValidation) AssertEqual(jmesPath string, expected interface{}, msg string) *stepRequestValidation {
	validator := TValidator{
		Check:      jmesPath,
		Comparator: "equals",
		Expect:     expected,
		Message:    msg,
	}
	s.step.Validators = append(s.step.Validators, validator)
	return s
}

func (s *stepRequestValidation) Name() string {
	return s.step.Name
}

func (s *stepRequestValidation) Type() string {
	return fmt.Sprintf("request-%v", s.step.Request.Method)
}

func (s *stepRequestValidation) Run() error {
	return s.runner.runStep(s.step)
}
