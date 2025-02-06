package hrp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLocateFile(t *testing.T) {
	// specify target file path
	_, err := locateFile(tmpl("plugin/debugtalk.go"), PluginGoSourceFile)
	if !assert.Nil(t, err) {
		t.Fatal()
	}

	// specify path with the same dir
	_, err = locateFile(tmpl("plugin/debugtalk.py"), PluginGoSourceFile)
	if !assert.Nil(t, err) {
		t.Fatal()
	}

	// specify target file path dir
	_, err = locateFile(tmpl("plugin/"), PluginGoSourceFile)
	if !assert.Nil(t, err) {
		t.Fatal()
	}

	// specify wrong path
	_, err = locateFile(".", PluginGoSourceFile)
	if !assert.Error(t, err) {
		t.Fatal()
	}
	_, err = locateFile("/abc", PluginGoSourceFile)
	if !assert.Error(t, err) {
		t.Fatal()
	}
}

func TestLocatePythonPlugin(t *testing.T) {
	_, err := locatePlugin(tmpl("plugin/debugtalk.py"))
	if !assert.Nil(t, err) {
		t.Fatal()
	}
}

func TestLocateGoPlugin(t *testing.T) {
	buildHashicorpGoPlugin()
	defer removeHashicorpGoPlugin()

	_, err := locatePlugin(tmpl("debugtalk.bin"))
	if !assert.Nil(t, err) {
		t.Fatal()
	}
}
