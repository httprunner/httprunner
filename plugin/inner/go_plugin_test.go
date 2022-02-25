// +build linux freebsd darwin
// go plugin doesn't support windows

package pluginInternal

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
		"-o=debugtalk.so", "../../examples/plugin/debugtalk.go")
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func removeGoPlugin() {
	fmt.Println("[teardown] remove go plugin")
	os.Remove("debugtalk.so")
}

func TestLocatePlugin(t *testing.T) {
	buildGoPlugin()
	defer removeGoPlugin()

	_, err := locateFile("../", goPluginFile)
	if !assert.Error(t, err) {
		t.Fail()
	}

	_, err = locateFile("", goPluginFile)
	if !assert.Error(t, err) {
		t.Fail()
	}

	startPath := "debugtalk.so"
	_, err = locateFile(startPath, goPluginFile)
	if !assert.Nil(t, err) {
		t.Fail()
	}

	startPath = "call.go"
	_, err = locateFile(startPath, goPluginFile)
	if !assert.Nil(t, err) {
		t.Fail()
	}

	startPath = "."
	_, err = locateFile(startPath, goPluginFile)
	if !assert.Nil(t, err) {
		t.Fail()
	}

	startPath = "/abc"
	_, err = locateFile(startPath, goPluginFile)
	if !assert.Error(t, err) {
		t.Fail()
	}
}

func TestCallPluginFunction(t *testing.T) {
	buildGoPlugin()
	defer removeGoPlugin()

	plugin, err := Init("debugtalk.so", false)
	if err != nil {
		t.Fatal(err)
	}

	if !assert.True(t, plugin.Has("Concatenate")) {
		t.Fail()
	}

	// call function without arguments
	result, err := plugin.Call("Concatenate", "1", 2, "3.14")
	if !assert.NoError(t, err) {
		t.Fail()
	}
	if !assert.Equal(t, "123.14", result) {
		t.Fail()
	}
}
