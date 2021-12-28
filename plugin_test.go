// +build linux freebsd darwin
// go plugin doesn't support windows
package hrp

import (
	"fmt"
	"os"
	"os/exec"
	"plugin"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	fmt.Println("[TestMain] build go plugin")
	cmd := exec.Command("go", "build", "-buildmode=plugin", "-o=examples/debugtalk.so", "examples/plugin/debugtalk.go")
	if err := cmd.Run(); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestCallPluginFunction(t *testing.T) {
	plugins, err := plugin.Open("examples/debugtalk.so")
	if err != nil {
		t.Fatalf(err.Error())
	}
	pluginLoader := &pluginLoader{plugins}

	// call function without arguments
	f1, _ := getMappingFunction("Concatenate", pluginLoader)
	result, err := callFunc(f1, 1, "2", 3.14)
	if !assert.NoError(t, err) {
		t.Fail()
	}
	if !assert.Equal(t, result, "1_2_3.14") {
		t.Fail()
	}
}
