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
	assert.Nil(t, err)

	genDebugTalkPyPath := filepath.Join(tmpl("plugin/"), hrp.PluginPySourceGenFile)
	err = hrp.BuildPlugin(tmpl("plugin/debugtalk.py"), genDebugTalkPyPath)
	assert.Nil(t, err)

	contentBytes, err := builtin.LoadFile(genDebugTalkPyPath)
	assert.Nil(t, err)

	content := string(contentBytes)
	assert.Contains(t, content, "import funppy")
	assert.Contains(t, content, "funppy.register")

	reg, _ := regexp.Compile(`funppy\.register`)
	matchedSlice := reg.FindAllStringSubmatch(content, -1)
	assert.Len(t, matchedSlice, 10)
}
