package builtin

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStartsWith(t *testing.T) {
	testData := []struct {
		raw      string
		expected string
	}{
		{"", ""},
		{"a", "a"},
		{"abc", "a"},
		{"abc", "ab"},
	}

	for _, data := range testData {
		if !assert.True(t, StartsWith(t, data.raw, data.expected)) {
			t.Fail()
		}
	}
}

func TestEndsWith(t *testing.T) {
	testData := []struct {
		raw      string
		expected string
	}{
		{"", ""},
		{"a", "a"},
		{"abc", "c"},
		{"abc", "bc"},
	}

	for _, data := range testData {
		if !assert.True(t, EndsWith(t, data.raw, data.expected)) {
			t.Fail()
		}
	}
}

func TestEqualLength(t *testing.T) {
	testData := []struct {
		raw      interface{}
		expected int
	}{
		{"", 0},
		{[]string{}, 0},
		{map[string]interface{}{}, 0},
		{"a", 1},
		{[]string{"a"}, 1},
		{map[string]interface{}{"a": 123}, 1},
	}

	for _, data := range testData {
		if !assert.True(t, EqualLength(t, data.raw, data.expected)) {
			t.Fail()
		}
	}
}

func TestLessThanLength(t *testing.T) {
	testData := []struct {
		raw      interface{}
		expected int
	}{
		{"", 1},
		{[]string{}, 1},
		{map[string]interface{}{}, 1},
		{"a", 2},
		{[]string{"a"}, 2},
		{map[string]interface{}{"a": 123}, 2},
	}

	for _, data := range testData {
		if !assert.True(t, LessThanLength(t, data.raw, data.expected)) {
			t.Fail()
		}
	}
}

func TestLessOrEqualsLength(t *testing.T) {
	testData := []struct {
		raw      interface{}
		expected int
	}{
		{"", 1},
		{[]string{}, 1},
		{map[string]interface{}{"A": 111}, 1},
		{"a", 1},
		{[]string{"a"}, 2},
		{map[string]interface{}{"a": 123}, 2},
	}

	for _, data := range testData {
		if !assert.True(t, LessOrEqualsLength(t, data.raw, data.expected)) {
			t.Fail()
		}
	}
}

func TestGreaterThanLength(t *testing.T) {
	testData := []struct {
		raw      interface{}
		expected int
	}{
		{"abcd", 3},
		{[]string{"a", "b", "c"}, 2},
		{map[string]interface{}{"a": 123, "b": 223, "c": 323}, 2},
	}

	for _, data := range testData {
		if !assert.True(t, GreaterThanLength(t, data.raw, data.expected)) {
			t.Fail()
		}
	}
}

func TestGreaterOrEqualsLength(t *testing.T) {
	testData := []struct {
		raw      interface{}
		expected int
	}{
		{"abcd", 3},
		{[]string{"w"}, 1},
		{map[string]interface{}{"A": 111}, 1},
		{"a", 1},
		{[]string{"a", "b", "c"}, 2},
		{map[string]interface{}{"a": 123, "b": 223, "c": 323}, 2},
	}

	for _, data := range testData {
		if !assert.True(t, GreaterOrEqualsLength(t, data.raw, data.expected)) {
			t.Fail()
		}
	}
}

func TestContainedBy(t *testing.T) {
	testData := []struct {
		raw      interface{}
		expected interface{}
	}{
		{"abcd", "abcdefg"},
		{"a", []string{"a", "b", "c"}},
		{"A", map[string]interface{}{"A": 111, "B": 222}},
	}

	for _, data := range testData {
		if !assert.True(t, ContainedBy(t, data.raw, data.expected)) {
			t.Fail()
		}
	}
}

func TestStringEqual(t *testing.T) {
	testData := []struct {
		raw      interface{}
		expected interface{}
	}{
		{"abcd", "abcd"},
		{"abcd", "ABCD"},
		{"ABcd", "abCD"},
	}

	for _, data := range testData {
		if !assert.True(t, StringEqual(t, data.raw, data.expected)) {
			t.Fail()
		}
	}
}

func TestRegexMatch(t *testing.T) {
	testData := []struct {
		raw      interface{}
		expected interface{}
	}{
		{"it's starting...", regexp.MustCompile("start")},
		{"it's not starting", "starting$"},
	}

	for _, data := range testData {
		if !assert.True(t, RegexMatch(t, data.raw, data.expected)) {
			t.Fail()
		}
	}
}
