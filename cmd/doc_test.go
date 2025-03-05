package cmd

import (
	"testing"

	"github.com/spf13/cobra/doc"
	"github.com/stretchr/testify/assert"
)

// run this test to generate markdown docs for hrp command
func TestGenMarkdownTree(t *testing.T) {
	err := doc.GenMarkdownTree(rootCmd, "../../docs/cmd")
	assert.Nil(t, err)
}
