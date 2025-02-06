package builtin

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/stretchr/testify/assert"
)

var Assertions = map[string]func(t assert.TestingT, actual interface{}, expected interface{}, msgAndArgs ...interface{}) bool{
	"eq":                EqualValues,
	"equals":            EqualValues,
	"equal":             EqualValues,
	"lt":                assert.Less,
	"less_than":         assert.Less,
	"le":                assert.LessOrEqual,
	"less_or_equals":    assert.LessOrEqual,
	"gt":                assert.Greater,
	"greater_than":      assert.Greater,
	"ge":                assert.GreaterOrEqual,
	"greater_or_equals": assert.GreaterOrEqual,
	"ne":                NotEqual,
	"not_equal":         NotEqual,
	"contains":          assert.Contains,
	"type_match":        assert.IsType,
	// custom assertions
	"startswith":               StartsWith,
	"endswith":                 EndsWith,
	"len_eq":                   EqualLength,
	"length_equals":            EqualLength,
	"length_equal":             EqualLength,
	"len_lt":                   LessThanLength,
	"count_lt":                 LessThanLength,
	"length_less_than":         LessThanLength,
	"len_le":                   LessOrEqualsLength,
	"count_le":                 LessOrEqualsLength,
	"length_less_or_equals":    LessOrEqualsLength,
	"len_gt":                   GreaterThanLength,
	"count_gt":                 GreaterThanLength,
	"length_greater_than":      GreaterThanLength,
	"len_ge":                   GreaterOrEqualsLength,
	"count_ge":                 GreaterOrEqualsLength,
	"length_greater_or_equals": GreaterOrEqualsLength,
	"contained_by":             ContainedBy,
	"str_eq":                   StringEqual,
	"string_equals":            StringEqual,
	"equal_fold":               EqualFold,
	"regex_match":              RegexMatch,
}

func EqualValues(t assert.TestingT, actual, expected interface{}, msgAndArgs ...interface{}) bool {
	return assert.EqualValues(t, expected, actual, msgAndArgs)
}

func NotEqual(t assert.TestingT, actual, expected interface{}, msgAndArgs ...interface{}) bool {
	return assert.NotEqual(t, expected, actual, msgAndArgs)
}

// StartsWith check if string starts with substring
func StartsWith(t assert.TestingT, actual, expected interface{}, msgAndArgs ...interface{}) bool {
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

// EndsWith check if string ends with substring
func EndsWith(t assert.TestingT, actual, expected interface{}, msgAndArgs ...interface{}) bool {
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

func EqualLength(t assert.TestingT, actual, expected interface{}, msgAndArgs ...interface{}) bool {
	length, err := convertInt(expected)
	if err != nil {
		return assert.Fail(t, fmt.Sprintf("expect int type, got %#v", expected), msgAndArgs...)
	}
	ok, l := getLen(actual)
	if !ok {
		return assert.Fail(t, fmt.Sprintf("actual value %v(%T) can't get length", actual, actual), msgAndArgs...)
	}
	if l != length {
		return assert.Fail(t, fmt.Sprintf("%v length == %d, expect == %d", actual, l, length), msgAndArgs...)
	}
	return true
}

func GreaterThanLength(t assert.TestingT, actual, expected interface{}, msgAndArgs ...interface{}) bool {
	length, err := convertInt(expected)
	if err != nil {
		return assert.Fail(t, fmt.Sprintf("expect int type, got %#v", expected), msgAndArgs...)
	}
	ok, l := getLen(actual)
	if !ok {
		return assert.Fail(t, fmt.Sprintf("actual value %v(%T) can't get length", actual, actual), msgAndArgs...)
	}
	if l <= length {
		return assert.Fail(t, fmt.Sprintf("%v length == %d, expect > %d", actual, l, length), msgAndArgs...)
	}
	return true
}

func GreaterOrEqualsLength(t assert.TestingT, actual, expected interface{}, msgAndArgs ...interface{}) bool {
	length, err := convertInt(expected)
	if err != nil {
		return assert.Fail(t, fmt.Sprintf("expect int type, got %#v", expected), msgAndArgs...)
	}
	ok, l := getLen(actual)
	if !ok {
		return assert.Fail(t, fmt.Sprintf("actual value %v(%T) can't get length", actual, actual), msgAndArgs...)
	}
	if l < length {
		return assert.Fail(t, fmt.Sprintf("%v length == %d, expect >= %d", actual, l, length), msgAndArgs...)
	}
	return true
}

func LessThanLength(t assert.TestingT, actual, expected interface{}, msgAndArgs ...interface{}) bool {
	length, err := convertInt(expected)
	if err != nil {
		return assert.Fail(t, fmt.Sprintf("expect int type, got %#v", expected), msgAndArgs...)
	}
	ok, l := getLen(actual)
	if !ok {
		return assert.Fail(t, fmt.Sprintf("actual value %v(%T) can't get length", actual, actual), msgAndArgs...)
	}
	if l >= length {
		return assert.Fail(t, fmt.Sprintf("%v length == %d, expect < %d", actual, l, length), msgAndArgs...)
	}
	return true
}

func LessOrEqualsLength(t assert.TestingT, actual, expected interface{}, msgAndArgs ...interface{}) bool {
	length, err := convertInt(expected)
	if err != nil {
		return assert.Fail(t, fmt.Sprintf("expect int type, got %#v", expected), msgAndArgs...)
	}
	ok, l := getLen(actual)
	if !ok {
		return assert.Fail(t, fmt.Sprintf("actual value %v(%T) can't get length", actual, actual), msgAndArgs...)
	}
	if l > length {
		return assert.Fail(t, fmt.Sprintf("%v length == %d, expect <= %d", actual, l, length), msgAndArgs...)
	}
	return true
}

// ContainedBy assert whether actual element contains expected element
func ContainedBy(t assert.TestingT, actual, expected interface{}, msgAndArgs ...interface{}) bool {
	return assert.Contains(t, expected, actual, msgAndArgs)
}

func StringEqual(t assert.TestingT, actual, expected interface{}, msgAndArgs ...interface{}) bool {
	a := fmt.Sprintf("%v", actual)
	e := fmt.Sprintf("%v", expected)
	return assert.True(t, a == e, msgAndArgs)
}

func EqualFold(t assert.TestingT, actual, expected interface{}, msgAndArgs ...interface{}) bool {
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

func RegexMatch(t assert.TestingT, actual, expected interface{}, msgAndArgs ...interface{}) bool {
	return assert.Regexp(t, expected, actual, msgAndArgs)
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
	case float32:
		return int(v), nil
	case float64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("unsupported int convertion for %v(%T)", v, v)
	}
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
