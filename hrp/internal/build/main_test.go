package build

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	err := Run("examples/debugtalk_no_funppy.py", "")
	if !assert.Nil(t, err) {
		t.Fatal()
	}

	err = Run("examples/debugtalk_no_fungo.go", "")
	if !assert.Nil(t, err) {
		t.Fatal()
	}

	err = Run("examples/debugtalk_no_funppy.py", "./debugtalk_gen.py")
	if !assert.Nil(t, err) {
		t.Fatal()
	}

	err = Run("examples/debugtalk_no_fungo.go", "./debugtalk_gen.bin")
	if !assert.Nil(t, err) {
		t.Fatal()
	}
}
