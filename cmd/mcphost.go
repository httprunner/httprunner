package cmd

import (
	"context"
	"fmt"

	"github.com/httprunner/httprunner/v5/pkg/mcphost"
	"github.com/spf13/cobra"
)

// CmdMCPHost represents the mcphost command
var CmdMCPHost = &cobra.Command{
	Use:   "mcphost",
	Short: "Export MCP server tools to JSON description",
	Long: `Export all tools from MCP servers to JSON description.
The tools will be exported with their descriptions, parameters, and return values.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create MCP host
		host, err := mcphost.NewMCPHost(mcpConfigPath)
		if err != nil {
			return fmt.Errorf("failed to create MCP host: %w", err)
		}

		// Initialize servers
		ctx := context.Background()
		if err := host.InitServers(ctx); err != nil {
			return fmt.Errorf("failed to initialize MCP servers: %w", err)
		}
		defer host.CloseServers()

		// Export tools to JSON
		if dumpPath != "" {
			if err := host.ExportToolsToJSON(ctx, dumpPath); err != nil {
				return err
			}
		}
		return nil
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
