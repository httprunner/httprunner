package cmd

import (
	"github.com/httprunner/httprunner/v5/server"
	"github.com/spf13/cobra"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server start",
	Short: "start hrp server",
	Long:  `start hrp server, call httprunner by HTTP`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return server.NewRouter().Run(port)
	},
}

var port int

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().IntVarP(&port, "port", "p", 8082, "Port to run the server on")
}
