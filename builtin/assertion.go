package builtin

import (
	"fmt"
	"strings"

	"github.com/stretchr/testify/assert"
)

var Assertions = map[string]func(t assert.TestingT, expected interface{}, actual interface{}, msgAndArgs ...interface{}) bool{
	"equals":            assert.EqualValues,
	"equal":             assert.EqualValues, // alias for equals
	"greater_than":      assert.Greater,
	"less_than":         assert.Less,
	"greater_or_equals": assert.GreaterOrEqual,
	"less_or_equals":    assert.LessOrEqual,
	"not_equal":         assert.NotEqual,
	"contains":          assert.Contains,
	"regex_match":       assert.Regexp,
	// custom assertions
	"startswith":    StartsWith, // check if string starts with substring
	"endswith":      EndsWith,   // check if string ends with substring
	"length_equals": EqualLength,
	"length_equal":  EqualLength, // alias for length_equals
}

func StartsWith(t assert.TestingT, expected, actual interface{}, msgAndArgs ...interface{}) bool {
	if !assert.IsType(t, "string", actual, fmt.Sprintf("actual is %v", actual)) {
		return false
	}
	if !assert.IsType(t, "string", expected, fmt.Sprintf("expected is %v", expected)) {
		return false
	}
	actualString := actual.(string)
	expectedString := expected.(string)
	return assert.True(t, strings.HasPrefix(actualString, expectedString), msgAndArgs...)
}

func EndsWith(t assert.TestingT, expected, actual interface{}, msgAndArgs ...interface{}) bool {
	if !assert.IsType(t, "string", actual, fmt.Sprintf("actual is %v", actual)) {
		return false
	}
	if !assert.IsType(t, "string", expected, fmt.Sprintf("expected is %v", expected)) {
		return false
	}
	actualString := actual.(string)
	expectedString := expected.(string)
	return assert.True(t, strings.HasSuffix(actualString, expectedString), msgAndArgs...)
}

func EqualLength(t assert.TestingT, expected, actual interface{}, msgAndArgs ...interface{}) bool {
	if !assert.IsType(t, 129, expected, fmt.Sprintf("expected type is not int, got %#v", expected)) {
		return false
	}
	return assert.Len(t, actual, expected.(int), msgAndArgs...)
}
