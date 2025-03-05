package tests

import (
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	hrp "github.com/httprunner/httprunner/v5"
	"github.com/httprunner/httprunner/v5/internal/builtin"
)

func TestRun(t *testing.T) {
	err := hrp.BuildPlugin(tmpl("plugin/debugtalk.go"), "./debugtalk.bin")
	if !assert.Nil(t, err) {
		t.Fatal()
	}

	genDebugTalkPyPath := filepath.Join(tmpl("plugin/"), hrp.PluginPySourceGenFile)
	err = hrp.BuildPlugin(tmpl("plugin/debugtalk.py"), genDebugTalkPyPath)
	if !assert.Nil(t, err) {
		t.Fatal()
	}

	contentBytes, err := builtin.LoadFile(genDebugTalkPyPath)
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
