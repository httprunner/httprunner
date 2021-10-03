package builtin

import (
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
	"length_equals": EqualLength,
	"length_equal":  EqualLength, // alias for length_equals
}

func EqualLength(t assert.TestingT, expected, actual interface{}, msgAndArgs ...interface{}) bool {
	return assert.Len(t, actual, expected.(int), msgAndArgs...)
}
