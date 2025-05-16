package mcphost

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewChat(t *testing.T) {
	systemPromptFile := "test_system_prompt.txt"
	_ = os.WriteFile(systemPromptFile, []byte("You are a helpful assistant."), 0o644)
	defer os.Remove(systemPromptFile)

	host, err := NewMCPHost("./testdata/test.mcp.json")
	require.NoError(t, err)

	chat, err := host.NewChat(context.Background(), systemPromptFile)
	assert.NoError(t, err)
	assert.NotNil(t, chat)
	assert.NotEmpty(t, chat.systemPrompt)
	assert.NotNil(t, chat.tools)
}

// func TestRunPromptWithNoToolCall(t *testing.T) {
// 	host, err := NewMCPHost("./testdata/test.mcp.json")
// 	require.NoError(t, err)

// 	chat, err := host.NewChat(context.Background(), "")
// 	assert.NoError(t, err)

// 	err = chat.runPrompt("hi")
// 	assert.NoError(t, err)
// 	assert.True(t, len(chat.history) > 1)
// }

// func TestRunPromptWithToolCall(t *testing.T) {
// 	host, err := NewMCPHost("./testdata/test.mcp.json")
// 	require.NoError(t, err)

// 	chat, err := host.NewChat(context.Background(), "")
// 	assert.NoError(t, err)
// 	assert.True(t, len(chat.tools) > 0)

// 	err = chat.runPrompt("what is the weather in CA")
// 	assert.NoError(t, err)
// 	assert.True(t, len(chat.history) > 1)
// }
