package hrp

import (
	"fmt"
)

// StepRequestValidation implements IStep interface.
type StepRequestValidation struct {
	step *TStep
}

func (s *StepRequestValidation) Name() string {
	if s.step.Name != "" {
		return s.step.Name
	}
	return fmt.Sprintf("%s %s", s.step.Request.Method, s.step.Request.URL)
}

func (s *StepRequestValidation) Type() string {
	return fmt.Sprintf("request-%v", s.step.Request.Method)
}

func (s *StepRequestValidation) ToStruct() *TStep {
	return s.step
}

func (s *StepRequestValidation) AssertEqual(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "equals",
		Expect:  expected,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertGreater(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "greater_than",
		Expect:  expected,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertLess(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "less_than",
		Expect:  expected,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertGreaterOrEqual(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "greater_or_equals",
		Expect:  expected,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertLessOrEqual(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "less_or_equals",
		Expect:  expected,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertNotEqual(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "not_equal",
		Expect:  expected,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertContains(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "contains",
		Expect:  expected,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertTypeMatch(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "type_match",
		Expect:  expected,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertRegexp(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "regex_match",
		Expect:  expected,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertStartsWith(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "startswith",
		Expect:  expected,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertEndsWith(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "endswith",
		Expect:  expected,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertLengthEqual(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "length_equals",
		Expect:  expected,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertContainedBy(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "contained_by",
		Expect:  expected,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertLengthLessThan(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "length_less_than",
		Expect:  expected,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertStringEqual(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "string_equals",
		Expect:  expected,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertLengthLessOrEquals(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "length_less_or_equals",
		Expect:  expected,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertLengthGreaterThan(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "length_greater_than",
		Expect:  expected,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertLengthGreaterOrEquals(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "length_greater_or_equals",
		Expect:  expected,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}
