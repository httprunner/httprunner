package scaffold

import (
	"testing"
)

func TestGenDemoExamples(t *testing.T) {
	dir := "../../../examples/demo-with-go-plugin"
	err := CreateScaffold(dir, Go, false, true)
	if err != nil {
		t.Fatal()
	}

	dir = "../../../examples/demo-with-py-plugin"
	err = CreateScaffold(dir, Py, false, true)
	if err != nil {
		t.Fatal()
	}

	dir = "../../../examples/demo-without-plugin"
	err = CreateScaffold(dir, Ignore, false, true)
	if err != nil {
		t.Fatal()
	}

	dir = "../../../examples/empty-demo-without-plugin"
	err = CreateScaffold(dir, Ignore, true, true)
	if err != nil {
		t.Fatal()
	}

	dir = "../../../examples/empty-demo-with-py-plugin"
	err = CreateScaffold(dir, Py, true, true)
	if err != nil {
		t.Fatal()
	}

	dir = "../../../examples/empty-demo-with-go-plugin"
	err = CreateScaffold(dir, Go, true, true)
	if err != nil {
		t.Fatal()
	}
}
