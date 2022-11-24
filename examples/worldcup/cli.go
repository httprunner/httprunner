package main

import (
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "wcl",
	Short:   "Monitor FIFA World Cup Live",
	Version: "0.1",
	PreRun: func(cmd *cobra.Command, args []string) {
		log.Logger = zerolog.New(
			zerolog.ConsoleWriter{NoColor: false, Out: os.Stderr},
		).With().Timestamp().Logger()
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		setLogLevel(logLevel)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		wc := NewWorldCupLive(matchName, osType, duration, interval)
		wc.Start()
		wc.DumpResult()
		return nil
	},
}

var (
	uuid      string
	osType    string
	duration  int
	interval  int
	logLevel  string
	matchName string
)

func main() {
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "INFO", "set log level")
	rootCmd.PersistentFlags().StringVarP(&uuid, "uuid", "u", "", "specify device serial or udid")
	rootCmd.PersistentFlags().StringVarP(&osType, "os-type", "t", "ios", "specify mobile os type")
	rootCmd.PersistentFlags().IntVarP(&duration, "duration", "d", 30, "set duration in seconds")
	rootCmd.PersistentFlags().IntVarP(&interval, "interval", "i", 15, "set interval in seconds")
	rootCmd.PersistentFlags().StringVarP(&matchName, "match-name", "n", "", "specify match name")

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func setLogLevel(level string) {
	level = strings.ToUpper(level)
	log.Info().Str("level", level).Msg("Set log level")
	switch level {
	case "DEBUG":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "INFO":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "WARN":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "ERROR":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "FATAL":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "PANIC":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	}
}
