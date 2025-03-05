package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"

	hrp "github.com/httprunner/httprunner/v5"
)

func TestLocateFile(t *testing.T) {
	// specify target file path
	_, err := hrp.LocateFile(tmpl("plugin/debugtalk.go"), hrp.PluginGoSourceFile)
	if !assert.Nil(t, err) {
		t.Fatal()
	}

	// specify path with the same dir
	_, err = hrp.LocateFile(tmpl("plugin/debugtalk.py"), hrp.PluginGoSourceFile)
	if !assert.Nil(t, err) {
		t.Fatal()
	}

	// specify target file path dir
	_, err = hrp.LocateFile(tmpl("plugin/"), hrp.PluginGoSourceFile)
	if !assert.Nil(t, err) {
		t.Fatal()
	}

	// specify wrong path
	_, err = hrp.LocateFile(".", hrp.PluginGoSourceFile)
	if !assert.Error(t, err) {
		t.Fatal()
	}
	_, err = hrp.LocateFile("/abc", hrp.PluginGoSourceFile)
	if !assert.Error(t, err) {
		t.Fatal()
	}
}

func TestLocatePythonPlugin(t *testing.T) {
	_, err := hrp.LocatePlugin(tmpl("plugin/debugtalk.py"))
	if !assert.Nil(t, err) {
		t.Fatal()
	}
}

func TestLocateGoPlugin(t *testing.T) {
	buildHashicorpGoPlugin()
	defer removeHashicorpGoPlugin()

	_, err := hrp.LocatePlugin(tmpl("debugtalk.bin"))
	if !assert.Nil(t, err) {
		t.Fatal()
	}
}
