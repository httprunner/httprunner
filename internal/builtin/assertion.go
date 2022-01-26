package builtin

import (
	"fmt"
	"reflect"
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
	"contained_by":      assert.Contains,
	"regex_match":       assert.Regexp,
	"type_match":        assert.IsType,
	// custom assertions
	"startswith":               StartsWith, // check if string starts with substring
	"endswith":                 EndsWith,   // check if string ends with substring
	"length_equals":            EqualLength,
	"length_equal":             EqualLength, // alias for length_equals
	"length_less_than":         LessThanLength,
	"length_less_or_equals":    LessOrEqualsLength,
	"length_greater_than":      GreaterThanLength,
	"length_greater_or_equals": GreaterOrEqualsLength,
	"contains":                 Contains,
	"string_equals":            EqualString,
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

func GreaterThanLength(t assert.TestingT, expected, actual interface{}, msgAndArgs ...interface{}) bool {
	length, err := convertInt(expected)
	if err != nil {
		return assert.Fail(t, fmt.Sprintf("expected type is not int, got %#v", expected), msgAndArgs...)
	}
	ok, l := getLen(actual)
	if !ok {
		return assert.Fail(t, fmt.Sprintf("\"%s\" could not be applied builtin len()", actual), msgAndArgs...)
	}
	if l <= length {
		return assert.Fail(t, fmt.Sprintf("\"%s\" should be more than %d item(s), but has %d", actual, length, l), msgAndArgs...)
	}
	return true
}

func GreaterOrEqualsLength(t assert.TestingT, expected, actual interface{}, msgAndArgs ...interface{}) bool {
	length, err := convertInt(expected)
	if err != nil {
		return assert.Fail(t, fmt.Sprintf("expected type is not int, got %#v", expected), msgAndArgs...)
	}
	ok, l := getLen(actual)
	if !ok {
		return assert.Fail(t, fmt.Sprintf("\"%s\" could not be applied builtin len()", actual), msgAndArgs...)
	}
	if l < length {
		return assert.Fail(t, fmt.Sprintf("\"%s\" should be no less than %d item(s), but has %d", actual, length, l), msgAndArgs...)
	}
	return true
}

func LessThanLength(t assert.TestingT, expected, actual interface{}, msgAndArgs ...interface{}) bool {
	length, err := convertInt(expected)
	if err != nil {
		return assert.Fail(t, fmt.Sprintf("expected type is not int, got %#v", expected), msgAndArgs...)
	}
	ok, l := getLen(actual)
	if !ok {
		return assert.Fail(t, fmt.Sprintf("\"%s\" could not be applied builtin len()", actual), msgAndArgs...)
	}
	if l >= length {
		return assert.Fail(t, fmt.Sprintf("\"%s\" should be less than %d item(s), but has %d", actual, length, l), msgAndArgs...)
	}
	return true
}

func LessOrEqualsLength(t assert.TestingT, expected, actual interface{}, msgAndArgs ...interface{}) bool {
	length, err := convertInt(expected)
	if err != nil {
		return assert.Fail(t, fmt.Sprintf("expected type is not int, got %#v", expected), msgAndArgs...)
	}
	ok, l := getLen(actual)
	if !ok {
		return assert.Fail(t, fmt.Sprintf("\"%s\" could not be applied builtin len()", actual), msgAndArgs...)
	}
	if l > length {
		return assert.Fail(t, fmt.Sprintf("\"%s\" should be no more than %d item(s), but has %d", actual, length, l), msgAndArgs...)
	}
	return true
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

// Contains assert whether actual element contains expected element
func Contains(t assert.TestingT, expected, actual interface{}, msgAndArgs ...interface{}) bool {
	return assert.Contains(t, actual, expected, msgAndArgs)
}

func EqualString(t assert.TestingT, expected, actual interface{}, msgAndArgs ...interface{}) bool {
	if !assert.IsType(t, "string", actual, msgAndArgs) {
		return false
	}
	if !assert.IsType(t, "string", expected, msgAndArgs) {
		return false
	}
	actualString := actual.(string)
	expectedString := expected.(string)
	return assert.True(t, strings.EqualFold(actualString, expectedString), msgAndArgs)
}

// getLen try to get length of object.
// return (false, 0) if impossible.
func getLen(x interface{}) (ok bool, length int) {
	v := reflect.ValueOf(x)
	defer func() {
		if e := recover(); e != nil {
			ok = false
		}
	}()
	return true, v.Len()
}
