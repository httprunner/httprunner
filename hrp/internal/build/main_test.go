package build

import (
	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	err := Run("../../../examples/demo-with-no-fungo/plugin/debugtalk.go", "")
	if !assert.Nil(t, err) {
		t.Fatal()
	}

	err = Run("../../../examples/demo-with-no-funppy/debugtalk.py", "")
	if !assert.Nil(t, err) {
		t.Fatal()
	}

	err = Run("../../../examples/demo-with-no-fungo/plugin/debugtalk.go", "./debugtalk_gen.bin")
	if !assert.Nil(t, err) {
		t.Fatal()
	}

	err = Run("../../../examples/demo-with-no-funppy/debugtalk.py", "./debugtalk_gen.py")
	if !assert.Nil(t, err) {
		t.Fatal()
	}

	contentBytes, err := builtin.ReadFile("./debugtalk_gen.py")
	if !assert.Nil(t, err) {
		t.Fatal()
	}

	content := string(contentBytes)
	if !assert.Contains(t, content, "import funppy") {
		t.Fatal()
	}

	if !assert.Contains(t, content, "funppy.register") {
		t.Fatal()
	}

	reg, _ := regexp.Compile(`funppy\.register`)
	matchedSlice := reg.FindAllStringSubmatch(content, -1)
	if !assert.Len(t, matchedSlice, 10) {
		t.Fatal()
	}
}
