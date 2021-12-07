package hrp

import (
	"fmt"
)

// implements IStep interface
type stepRequestValidation struct {
	step *TStep
}

func (s *stepRequestValidation) name() string {
	if s.step.Name != "" {
		return s.step.Name
	}
	return fmt.Sprintf("%s %s", s.step.Request.Method, s.step.Request.URL)
}

func (s *stepRequestValidation) getType() string {
	return fmt.Sprintf("request-%v", s.step.Request.Method)
}

func (s *stepRequestValidation) toStruct() *TStep {
	return s.step
}

func (s *stepRequestValidation) AssertEqual(jmesPath string, expected interface{}, msg string) *stepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "equals",
		Expect:  expected,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *stepRequestValidation) AssertStartsWith(jmesPath string, expected interface{}, msg string) *stepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "startswith",
		Expect:  expected,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *stepRequestValidation) AssertEndsWith(jmesPath string, expected interface{}, msg string) *stepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "endswith",
		Expect:  expected,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *stepRequestValidation) AssertLengthEqual(jmesPath string, expected interface{}, msg string) *stepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "length_equals",
		Expect:  expected,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}
