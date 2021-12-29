package builtin

import (
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
		if !assert.True(t, StartsWith(t, data.expected, data.raw)) {
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
		if !assert.True(t, EndsWith(t, data.expected, data.raw)) {
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
		if !assert.True(t, EqualLength(t, data.expected, data.raw)) {
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
		if !assert.True(t, LessThanLength(t, data.expected, data.raw)) {
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
		if !assert.True(t, LessOrEqualsLength(t, data.expected, data.raw)) {
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
		if !assert.True(t, GreaterThanLength(t, data.expected, data.raw)) {
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
		if !assert.True(t, GreaterOrEqualsLength(t, data.expected, data.raw)) {
			t.Fail()
		}
	}
}
