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

func TestMain(m *testing.M) {
	fmt.Println("[TestMain] build go plugin")
	// flag -race is necessary in order to be consistent with go test
	cmd := exec.Command("go", "build", "-buildmode=plugin", "-race", "-o=examples/debugtalk.so", "examples/plugin/debugtalk.go")
	if err := cmd.Run(); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestLocatePlugin(t *testing.T) {
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
	parser := newParser()
	parser.initPlugin("examples/debugtalk.so")

	// call function without arguments
	result, err := parser.callFunc("Concatenate", 1, "2", 3.14)
	if !assert.NoError(t, err) {
		t.Fail()
	}
	if !assert.Equal(t, result, "1_2_3.14") {
		t.Fail()
	}
}
