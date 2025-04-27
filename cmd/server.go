package cmd

import (
	"os"

	"github.com/httprunner/httprunner/v5/server"
	"github.com/spf13/cobra"
)

// serverCmd represents the server command
var CmdServer = &cobra.Command{
	Use:   "server start",
	Short: "start hrp server",
	Long:  `start hrp server, call httprunner by HTTP`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		router := server.NewRouter()
		mcpConfigPath = os.ExpandEnv(mcpConfigPath)
		if mcpConfigPath != "" {
			router.InitMCPHub(mcpConfigPath)
		}
		return router.Run(port)
	},
}

var (
	port          int
	mcpConfigPath string
)

func init() {
	CmdServer.Flags().IntVarP(&port, "port", "p", 8082, "port to run the server on")
	CmdServer.Flags().StringVarP(&mcpConfigPath, "mcp-config", "c", "$HOME/.hrp/mcp.json", "path to the MCP config file")
}
