package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"

	hrp "github.com/httprunner/httprunner/v5"
)

func TestLocateFile(t *testing.T) {
	// specify target file path
	_, err := hrp.LocateFile(tmpl("plugin/debugtalk.go"), hrp.PluginGoSourceFile)
	assert.Nil(t, err)

	// specify path with the same dir
	_, err = hrp.LocateFile(tmpl("plugin/debugtalk.py"), hrp.PluginGoSourceFile)
	assert.Nil(t, err)

	// specify target file path dir
	_, err = hrp.LocateFile(tmpl("plugin/"), hrp.PluginGoSourceFile)
	assert.Nil(t, err)

	// specify wrong path
	_, err = hrp.LocateFile(".", hrp.PluginGoSourceFile)
	assert.Error(t, err)
	_, err = hrp.LocateFile("/abc", hrp.PluginGoSourceFile)
	assert.Error(t, err)
}

func TestLocatePythonPlugin(t *testing.T) {
	_, err := hrp.LocatePlugin(tmpl("plugin/debugtalk.py"))
	assert.Nil(t, err)
}

func TestLocateGoPlugin(t *testing.T) {
	buildHashicorpGoPlugin()
	defer removeHashicorpGoPlugin()

	_, err := hrp.LocatePlugin(tmpl("debugtalk.bin"))
	assert.Nil(t, err)
}
