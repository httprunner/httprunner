package main

import (
	"errors"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v5/pkg/uixt"
	"github.com/httprunner/httprunner/v5/pkg/uixt/ai"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
)

var rootCmd = &cobra.Command{
	Use:     "wcl",
	Short:   "Monitor FIFA World Cup Live",
	Version: "2022.12.03.0018",
	RunE: func(cmd *cobra.Command, args []string) error {
		var driver uixt.IDriver
		var bundleID string
		if iosApp != "" {
			log.Info().Str("bundleID", iosApp).Msg("init ios device")
			device, err := uixt.NewIOSDevice(option.WithUDID(uuid))
			if err != nil {
				log.Fatal().Err(err).Msg("failed to init ios device")
			}
			driver, _ = uixt.NewWDADriver(device)
			bundleID = iosApp
		} else if androidApp != "" {
			log.Info().Str("bundleID", androidApp).Msg("init android device")
			device, err := uixt.NewAndroidDevice(option.WithSerialNumber(uuid))
			if err != nil {
				log.Fatal().Err(err).Msg("failed to init android device")
			}
			driver, _ = uixt.NewADBDriver(device)
			bundleID = androidApp
		} else {
			return errors.New("android or ios app bundldID is required")
		}

		driverExt := uixt.NewXTDriver(driver,
			ai.WithCVService(ai.CVServiceTypeVEDEM),
		)

		wc := NewWorldCupLive(driverExt, matchName, bundleID, duration, interval)

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
