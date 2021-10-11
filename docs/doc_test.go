package cmd

import (
	"log"
	"testing"

	"github.com/spf13/cobra/doc"

	"github.com/httprunner/httpboomer/httpboomer/cmd"
)

// run this test to generate markdown docs
func TestGenMarkdownTree(t *testing.T) {
	err := doc.GenMarkdownTree(cmd.RootCmd, "./cmd/")
	if err != nil {
		log.Fatal(err)
	}
}
