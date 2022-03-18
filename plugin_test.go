package hrp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLocateFile(t *testing.T) {
	// specify target file path
	_, err := locateFile("examples/plugin/debugtalk.go", "debugtalk.go")
	if !assert.Nil(t, err) {
		t.Fail()
	}

	// specify path with the same dir
	_, err = locateFile("examples/plugin/hashicorp.go", "debugtalk.go")
	if !assert.Nil(t, err) {
		t.Fail()
	}

	// specify target file path dir
	_, err = locateFile("examples/plugin/", "debugtalk.go")
	if !assert.Nil(t, err) {
		t.Fail()
	}

	// specify wrong path
	_, err = locateFile("examples", "debugtalk.go")
	if !assert.Error(t, err) {
		t.Fail()
	}
	_, err = locateFile("examples/demo.json", "debugtalk.go")
	if !assert.Error(t, err) {
		t.Fail()
	}
	_, err = locateFile(".", "debugtalk.go")
	if !assert.Error(t, err) {
		t.Fail()
	}
	_, err = locateFile("/abc", "debugtalk.go")
	if !assert.Error(t, err) {
		t.Fail()
	}
}

func TestLocatePythonPlugin(t *testing.T) {
	_, err := locatePlugin("examples/debugtalk.py")
	if !assert.Nil(t, err) {
		t.Fail()
	}
}

func TestLocateGoPlugin(t *testing.T) {
	buildHashicorpPlugin()
	defer removeHashicorpPlugin()

	_, err := locatePlugin("examples/debugtalk.bin")
	if !assert.Nil(t, err) {
		t.Fail()
	}
}
