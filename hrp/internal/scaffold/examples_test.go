package scaffold

import (
	"path/filepath"
	"testing"
)

func TestGenDemoExamples(t *testing.T) {
	dir := "../../../examples/demo-with-go-plugin"
	err := CreateScaffold(dir, Go, "", true)
	if err != nil {
		t.Fatal(err)
	}

	dir = "../../../examples/demo-with-py-plugin"
	venv := filepath.Join(dir, ".venv")
	_ = CreateScaffold(dir, Py, venv, true)
	// FIXME
	// if err != nil {
	// 	t.Fatal(err)
	// }

	dir = "../../../examples/demo-without-plugin"
	err = CreateScaffold(dir, Ignore, "", true)
	if err != nil {
		t.Fatal(err)
	}

	dir = "../../../examples/demo-empty-project"
	err = CreateScaffold(dir, Empty, "", true)
	if err != nil {
		t.Fatal(err)
	}
}
