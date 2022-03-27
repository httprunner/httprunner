package scaffold

import (
	"os"
	"testing"
)

func TestGenDemoExamples(t *testing.T) {
	dir := "../../../examples/demo-with-go-plugin"
	os.RemoveAll(dir)
	err := CreateScaffold(dir, Go)
	if err != nil {
		t.Fail()
	}

	dir = "../../../examples/demo-with-py-plugin"
	os.RemoveAll(dir)
	err = CreateScaffold(dir, Py)
	if err != nil {
		t.Fail()
	}

	dir = "../../../examples/demo-without-plugin"
	os.RemoveAll(dir)
	err = CreateScaffold(dir, Ignore)
	if err != nil {
		t.Fail()
	}
}
