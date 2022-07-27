package cmd

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v4/hrp/internal/dial"
)

var pingOptions dial.PingOptions

var pingCmd = &cobra.Command{
	Use:   "ping $url",
	Short: "run integrated ping command",
	Args:  cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		setLogLevel(logLevel)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return dial.DoPing(&pingOptions, args)
	},
}

func init() {
	rootCmd.AddCommand(pingCmd)
	pingCmd.Flags().IntVarP(&pingOptions.Count, "count", "c", 10, "Stop after sending (and receiving) N packets")
	pingCmd.Flags().DurationVarP(&pingOptions.Timeout, "timeout", "t", 20*time.Second, "Ping exits after N seconds")
	pingCmd.Flags().DurationVarP(&pingOptions.Interval, "interval", "i", 1*time.Second, "Wait N seconds between sending each packet")
	pingCmd.Flags().BoolVar(&pingOptions.SaveTests, "save-tests", false, "Save ping results json")
}
