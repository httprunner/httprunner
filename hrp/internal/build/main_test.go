package build

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	err := Run([]string{"examples/debugtalk_no_funppy.py", "examples/debugtalk_no_fungo.go"})
	if !assert.Nil(t, err) {
		t.Fatal()
	}
}

func TestConvertSnakeName(t *testing.T) {
	testData := []struct {
		expectedValue string
		originalValue string
	}{
		{
			expectedValue: "test_name",
			originalValue: "testName",
		},
		{
			expectedValue: "test",
			originalValue: "test",
		},
		{
			expectedValue: "test_name",
			originalValue: "TestName",
		},
		{
			expectedValue: "test_name",
			originalValue: "test_name",
		},
	}
	for _, data := range testData {
		name := convertSnakeName(data.originalValue)
		if !assert.Equal(t, data.expectedValue, name) {
			t.Fatal()
		}
	}
}
