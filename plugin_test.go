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
	// go tool compile -help
	// all=: apply to all packages
	// -N: disable optimizations
	// -l: disable inlining
	cmd := exec.Command("go", "build", "-buildmode=plugin", `-gcflags="all=-N -l"`, "-o=examples/debugtalk.so", "examples/plugin/debugtalk.go")
	if err := cmd.Run(); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestCallPluginFunction(t *testing.T) {
	parser := newParser()
	parser.loadPlugin("examples/debugtalk.so")

	// call function without arguments
	result, err := parser.callFunc("Concatenate", 1, "2", 3.14)
	if !assert.NoError(t, err) {
		t.Fail()
	}
	if !assert.Equal(t, result, "1_2_3.14") {
		t.Fail()
	}
}
