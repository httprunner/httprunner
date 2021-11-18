package cmd

import (
	"testing"

	"github.com/spf13/cobra/doc"

	"github.com/httprunner/hrp/hrp/cmd"
)

// run this test to generate markdown docs
func TestGenMarkdownTree(t *testing.T) {
	err := doc.GenMarkdownTree(cmd.RootCmd, "./")
	if err != nil {
		t.Fatal(err)
	}
}
