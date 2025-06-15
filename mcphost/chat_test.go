package mcphost

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunPromptWithNoToolCall(t *testing.T) {
	host, err := NewMCPHost("./testdata/test.mcp.json", true)
	require.NoError(t, err)

	chat, err := host.NewChat(context.Background())
	assert.NoError(t, err)

	err = chat.runPrompt(context.Background(), "hi")
	assert.NoError(t, err)
	assert.True(t, len(*chat.planner.History()) > 1)
}

func TestRunPromptWithToolCall(t *testing.T) {
	host, err := NewMCPHost("./testdata/test.mcp.json", true)
	require.NoError(t, err)

	chat, err := host.NewChat(context.Background())
	assert.NoError(t, err)

	err = chat.runPrompt(context.Background(), "what is the weather in CA")
	assert.NoError(t, err)
	assert.True(t, len(*chat.planner.History()) > 1)
}
