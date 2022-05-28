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

	genDebugTalkPyPath := "../scaffold/templates/plugin/" + genDebugTalkPy
	err = Run("../scaffold/templates/plugin/debugtalk.py", genDebugTalkPyPath)
	if !assert.Nil(t, err) {
		t.Fatal()
	}

	contentBytes, err := builtin.ReadFile(genDebugTalkPyPath)
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
