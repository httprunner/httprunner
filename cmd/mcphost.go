package cmd

import (
	"context"
	"fmt"

	"github.com/httprunner/httprunner/v5/mcphost"
	"github.com/spf13/cobra"
)

// CmdMCPHost represents the mcphost command
var CmdMCPHost = &cobra.Command{
	Use:   "mcphost",
	Short: "Start a chat session to interact with MCP tools",
	Long:  `mcphost is a command-line tool that allows you to interact with MCP tools.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create MCP host
		host, err := mcphost.NewMCPHost(mcpConfigPath)
		if err != nil {
			return fmt.Errorf("failed to create MCP host: %w", err)
		}
		defer host.CloseServers()

		// If dump flag is set, dump MCP server tools to JSON file
		if dumpPath != "" {
			return host.ExportToolsToJSON(context.Background(), dumpPath)
		}

		// Create chat session
		chat, err := host.NewChat(context.Background())
		if err != nil {
			return fmt.Errorf("failed to create chat session: %w", err)
		}

		// Start chat
		return chat.Start(context.Background())
	},
}

var (
	mcpConfigPath string
	dumpPath      string
)

func init() {
	CmdMCPHost.Flags().StringVarP(&mcpConfigPath, "mcp-config", "c", "$HOME/.hrp/mcp.json", "path to the MCP config file")
	CmdMCPHost.Flags().StringVar(&dumpPath, "dump", "", "path to save the exported tools JSON file")
}
