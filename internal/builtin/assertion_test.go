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
