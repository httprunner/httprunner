package cmd

import (
	"testing"

	"github.com/spf13/cobra/doc"
)

// run this test to generate markdown docs for hrp command
func TestGenMarkdownTree(t *testing.T) {
	err := doc.GenMarkdownTree(rootCmd, "../../../docs/cmd")
	if err != nil {
		t.Fatal(err)
	}
}
