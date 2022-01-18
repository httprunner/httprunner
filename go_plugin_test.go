// +build linux freebsd darwin
// go plugin doesn't support windows

package hrp

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func buildGoPlugin() {
	fmt.Println("[setup] build go plugin")
	// flag -race is necessary in order to be consistent with go test
	cmd := exec.Command("go", "build", "-buildmode=plugin", "-race",
		"-o=examples/debugtalk.so", "examples/plugin/debugtalk.go")
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func removeGoPlugin() {
	fmt.Println("[teardown] remove go plugin")
	os.Remove("examples/debugtalk.so")
}

func TestLocatePlugin(t *testing.T) {
	buildGoPlugin()
	defer removeGoPlugin()

	cwd, _ := os.Getwd()
	_, err := locatePlugin(cwd, goPluginFile)
	if !assert.Error(t, err) {
		t.Fail()
	}

	_, err = locatePlugin("", goPluginFile)
	if !assert.Error(t, err) {
		t.Fail()
	}

	startPath := "examples/debugtalk.so"
	_, err = locatePlugin(startPath, goPluginFile)
	if !assert.Nil(t, err) {
		t.Fail()
	}

	startPath = "examples/demo.json"
	_, err = locatePlugin(startPath, goPluginFile)
	if !assert.Nil(t, err) {
		t.Fail()
	}

	startPath = "examples/"
	_, err = locatePlugin(startPath, goPluginFile)
	if !assert.Nil(t, err) {
		t.Fail()
	}

	startPath = "examples/plugin/debugtalk.go"
	_, err = locatePlugin(startPath, goPluginFile)
	if !assert.Nil(t, err) {
		t.Fail()
	}

	startPath = "/abc"
	_, err = locatePlugin(startPath, goPluginFile)
	if !assert.Error(t, err) {
		t.Fail()
	}
}

func TestCallPluginFunction(t *testing.T) {
	buildGoPlugin()
	removeHashicorpPlugin()
	defer removeGoPlugin()

	parser := newParser()
	err := parser.initPlugin("examples/debugtalk.so")
	if err != nil {
		t.Fatal(err)
	}

	// call function without arguments
	result, err := parser.callFunc("Concatenate", "1", 2, "3.14")
	if !assert.NoError(t, err) {
		t.Fail()
	}
	if !assert.Equal(t, "123.14", result) {
		t.Fail()
	}
}
