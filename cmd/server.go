package cmd

import (
	"github.com/spf13/cobra"

	server_ext "github.com/httprunner/httprunner/v5/server/ext"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server start",
	Short: "start hrp server",
	Long:  `start hrp server, call httprunner by HTTP`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return server_ext.NewExtRouter().Run(port)
	},
}

var port int

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().IntVarP(&port, "port", "p", 8082, "Port to run the server on")
}
