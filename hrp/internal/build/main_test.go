package build

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
)

func TestRun(t *testing.T) {
	err := Run("../scaffold/templates/plugin/debugtalk.go", "./debugtalk.bin")
	if !assert.Nil(t, err) {
		t.Fatal()
	}

	genDebugTalkPy := "../scaffold/templates/plugin/debugtalk_gen.py"
	err = Run("../scaffold/templates/plugin/debugtalk.py", genDebugTalkPy)
	if !assert.Nil(t, err) {
		t.Fatal()
	}

	contentBytes, err := builtin.ReadFile(genDebugTalkPy)
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
