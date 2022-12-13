package main

import (
	"errors"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v4/hrp/pkg/uixt"
)

var rootCmd = &cobra.Command{
	Use:     "wcl",
	Short:   "Monitor FIFA World Cup Live",
	Version: "2022.12.03.0018",
	PreRun: func(cmd *cobra.Command, args []string) {
		log.Logger = zerolog.New(
			zerolog.ConsoleWriter{NoColor: false, Out: os.Stderr},
		).With().Timestamp().Logger()
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		setLogLevel(logLevel)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var device uixt.Device
		var bundleID string
		if iosApp != "" {
			log.Info().Str("bundleID", iosApp).Msg("init ios device")
			device = initIOSDevice(uuid)
			bundleID = iosApp
		} else if androidApp != "" {
			log.Info().Str("bundleID", androidApp).Msg("init android device")
			device = initAndroidDevice(uuid)
			bundleID = androidApp
		} else {
			return errors.New("android or ios app bundldID is required")
		}

		wc := NewWorldCupLive(device, matchName, bundleID, duration, interval)

		if auto {
			wc.EnterLive(bundleID)
		}

		wc.Start()
		return nil
	},
}

var (
	uuid       string
	iosApp     string
	androidApp string
	auto       bool
	duration   int
	interval   int
	logLevel   string
	matchName  string
	perf       []string
)

func main() {
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "INFO", "set log level")
	rootCmd.PersistentFlags().StringVarP(&uuid, "uuid", "u", "", "specify device serial or udid")
	rootCmd.PersistentFlags().StringVar(&iosApp, "ios", "", "run ios app")
	rootCmd.PersistentFlags().StringVar(&androidApp, "android", "", "run android app")
	rootCmd.PersistentFlags().BoolVar(&auto, "auto", false, "auto enter live")
	rootCmd.PersistentFlags().IntVarP(&duration, "duration", "d", 30, "set duration in seconds")
	rootCmd.PersistentFlags().IntVarP(&interval, "interval", "i", 15, "set interval in seconds")
	rootCmd.PersistentFlags().StringVarP(&matchName, "match-name", "n", "", "specify match name")
	rootCmd.PersistentFlags().StringSliceVarP(&perf, "perf", "p", nil,
		"specify performance monitor, e.g. sys_cpu,sys_mem,sys_net,sys_disk,fps,network,gpu")

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
