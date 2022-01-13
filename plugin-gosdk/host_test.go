package plugin

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	fmt.Println("[TestMain] build go plugin")
	cmd := exec.Command("go", "build", "-o=debugtalk", "plugin/debugtalk.go")
	if err := cmd.Run(); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestInitHashicorpPlugin(t *testing.T) {
	f, err := Init("./debugtalk")
	if err != nil {
		t.Fatal(err)
	}
	defer Quit()

	v, err := f.Call("sum_int", 1, 2, 3, 4)
	if err != nil {
		t.Fatal(err)
	}
	if !assert.Equal(t, 10, v) {
		t.Fail()
	}
	v, err = f.Call("concatenate_string", "a", "b", "c")
	if err != nil {
		t.Fatal(err)
	}
	if !assert.Equal(t, "abc", v) {
		t.Fail()
	}
}
