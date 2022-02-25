package pluginInternal

import (
	"os"
	"os/exec"
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func buildHashicorpPlugin() {
	log.Info().Msg("[init] build hashicorp go plugin")
	cmd := exec.Command("go", "build",
		"-o", "../../examples/debugtalk.bin",
		"../../examples/plugin/hashicorp.go", "../../examples/plugin/debugtalk.go")
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func removeHashicorpPlugin() {
	log.Info().Msg("[teardown] remove hashicorp plugin")
	os.Remove("../../examples/debugtalk.bin")
}

func TestInitHashicorpPlugin(t *testing.T) {
	buildHashicorpPlugin()
	defer removeHashicorpPlugin()

	plugin, err := Init("../../examples/debugtalk.bin", false)
	if err != nil {
		t.Fatal(err)
	}
	defer plugin.Quit()

	if !assert.True(t, plugin.Has("sum_ints")) {
		t.Fatal(err)
	}
	if !assert.True(t, plugin.Has("concatenate")) {
		t.Fatal(err)
	}

	var v2 interface{}
	v2, err = plugin.Call("sum_ints", 1, 2, 3, 4)
	if err != nil {
		t.Fatal(err)
	}
	if !assert.Equal(t, 10, v2) {
		t.Fail()
	}
	v2, err = plugin.Call("sum_two_int", 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	if !assert.Equal(t, 3, v2) {
		t.Fail()
	}
	v2, err = plugin.Call("sum", 1, 2, 3.4, 5)
	if err != nil {
		t.Fatal(err)
	}
	if !assert.Equal(t, 11.4, v2) {
		t.Fail()
	}

	var v3 interface{}
	v3, err = plugin.Call("sum_two_string", "a", "b")
	if err != nil {
		t.Fatal(err)
	}
	if !assert.Equal(t, "ab", v3) {
		t.Fail()
	}
	v3, err = plugin.Call("sum_strings", "a", "b", "c")
	if err != nil {
		t.Fatal(err)
	}
	if !assert.Equal(t, "abc", v3) {
		t.Fail()
	}

	v3, err = plugin.Call("concatenate", "a", 2, "c", 3.4)
	if err != nil {
		t.Fatal(err)
	}
	if !assert.Equal(t, "a2c3.4", v3) {
		t.Fail()
	}
}
