package httpboomer

import (
	"fmt"
)

// implements IStep interface
type stepRequestValidation struct {
	step *TStep
}

func (s *stepRequestValidation) Name() string {
	return s.step.Name
}

func (s *stepRequestValidation) Type() string {
	return fmt.Sprintf("request-%v", s.step.Request.Method)
}

func (s *stepRequestValidation) ToStruct() *TStep {
	return s.step
}

func (s *stepRequestValidation) AssertEqual(jmesPath string, expected interface{}, msg string) *stepRequestValidation {
	validator := TValidator{
		Check:   jmesPath,
		Assert:  "equals",
		Expect:  expected,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, validator)
	return s
}

func (s *stepRequestValidation) AssertStartsWith(jmesPath string, expected interface{}, msg string) *stepRequestValidation {
	validator := TValidator{
		Check:   jmesPath,
		Assert:  "startswith",
		Expect:  expected,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, validator)
	return s
}

func (s *stepRequestValidation) AssertEndsWith(jmesPath string, expected interface{}, msg string) *stepRequestValidation {
	validator := TValidator{
		Check:   jmesPath,
		Assert:  "endswith",
		Expect:  expected,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, validator)
	return s
}

func (s *stepRequestValidation) AssertLengthEqual(jmesPath string, expected interface{}, msg string) *stepRequestValidation {
	validator := TValidator{
		Check:   jmesPath,
		Assert:  "length_equals",
		Expect:  expected,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, validator)
	return s
}
