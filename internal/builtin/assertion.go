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
	length, err := convertInt(expected)
	if err != nil {
		return assert.Fail(t, fmt.Sprintf("expected type is not int, got %#v", expected), msgAndArgs...)
	}

	return assert.Len(t, actual, length, msgAndArgs...)
}

func convertInt(value interface{}) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case int8:
		return int(v), nil
	case int16:
		return int(v), nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case uint:
		return int(v), nil
	case uint8:
		return int(v), nil
	case uint16:
		return int(v), nil
	case uint32:
		return int(v), nil
	case uint64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("unsupported int convertion for %v(%T)", v, v)
	}
}
