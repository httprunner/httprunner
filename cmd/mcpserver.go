package cmd

import (
	"github.com/httprunner/httprunner/v5/mcphost"
	"github.com/spf13/cobra"
)

var CmdMCPServer = &cobra.Command{
	Use:   "mcp-server",
	Short: "Start MCP server for UI automation",
	Long:  `Start MCP server for UI automation, expose device driver via MCP protocol`,
	RunE: func(cmd *cobra.Command, args []string) error {
		mcpServer := mcphost.NewMCPServer()
		return mcpServer.Start()
	},
}
