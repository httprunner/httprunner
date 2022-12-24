package cmd

import (
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v4/hrp/internal/dial"
)

var (
	pingOptions dial.PingOptions
	dnsOptions  dial.DnsOptions
)

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

var dnsCmd = &cobra.Command{
	Use:   "dns $url",
	Short: "DNS resolution for different source and record types",
	Args:  cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		setLogLevel(logLevel)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if dnsOptions.DnsSourceType != dial.DnsSourceTypeLocal && dnsOptions.DnsServer != "" {
			log.Warn().Msg("DNS server not supported for non-local DNS source, ignored")
		}
		if dnsOptions.DnsSourceType == dial.DnsSourceTypeHttp && dnsOptions.DnsRecordType == dial.DnsRecordTypeCNAME {
			log.Warn().Msg("CNAME record not supported for http DNS source, using default record type(A)")
		}
		return dial.DoDns(&dnsOptions, args)
	},
}

func init() {
	rootCmd.AddCommand(pingCmd)
	pingCmd.Flags().IntVarP(&pingOptions.Count, "count", "c", 10, "Stop after sending (and receiving) N packets")
	pingCmd.Flags().DurationVarP(&pingOptions.Timeout, "timeout", "t", 20*time.Second, "Ping exits after N seconds")
	pingCmd.Flags().DurationVarP(&pingOptions.Interval, "interval", "i", 1*time.Second, "Wait N seconds between sending each packet")
	pingCmd.Flags().BoolVar(&pingOptions.SaveTests, "save-tests", false, "Save ping result as json")

	rootCmd.AddCommand(dnsCmd)
	dnsCmd.Flags().IntVar(&dnsOptions.DnsSourceType, "dns-source", 0, "DNS source type\n0: local DNS\n1: http DNS\n2: google DNS")
	dnsCmd.Flags().IntVar(&dnsOptions.DnsRecordType, "dns-record", 1, "DNS record type\n1: A\n28: AAAA\n5: CNAME")
	dnsCmd.Flags().StringVar(&dnsOptions.DnsServer, "dns-server", "", "DNS server, only available for local DNS source")
	dnsCmd.Flags().BoolVar(&dnsOptions.SaveTests, "save-tests", false, "Save DNS resolution result as json")
}
