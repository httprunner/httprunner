package cmd

import (
	"io"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v4/hrp/cmd/adb"
	"github.com/httprunner/httprunner/v4/hrp/cmd/ios"
	"github.com/httprunner/httprunner/v4/hrp/internal/code"
	"github.com/httprunner/httprunner/v4/hrp/internal/version"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "hrp",
	Short: "Next-Generation API Testing Solution.",
	Long: `
██╗  ██╗████████╗████████╗██████╗ ██████╗ ██╗   ██╗███╗   ██╗███╗   ██╗███████╗██████╗
██║  ██║╚══██╔══╝╚══██╔══╝██╔══██╗██╔══██╗██║   ██║████╗  ██║████╗  ██║██╔════╝██╔══██╗
███████║   ██║      ██║   ██████╔╝██████╔╝██║   ██║██╔██╗ ██║██╔██╗ ██║█████╗  ██████╔╝
██╔══██║   ██║      ██║   ██╔═══╝ ██╔══██╗██║   ██║██║╚██╗██║██║╚██╗██║██╔══╝  ██╔══██╗
██║  ██║   ██║      ██║   ██║     ██║  ██║╚██████╔╝██║ ╚████║██║ ╚████║███████╗██║  ██║
╚═╝  ╚═╝   ╚═╝      ╚═╝   ╚═╝     ╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═══╝╚═╝  ╚═══╝╚══════╝╚═╝  ╚═╝

HttpRunner is an open source API testing tool that supports HTTP(S)/HTTP2/WebSocket/RPC
network protocols, covering API testing, performance testing and digital experience
monitoring (DEM) test types. Enjoy! ✨ 🚀 ✨

License: Apache-2.0
Website: https://httprunner.com
Github: https://github.com/httprunner/httprunner
Copyright 2017 debugtalk`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		initLogger(logLevel, logJSON)
	},
	Version:          version.VERSION,
	TraverseChildren: true, // parses flags on all parents before executing child command
	SilenceUsage:     true, // silence usage when an error occurs
}

var (
	logLevel string
	logJSON  bool
	venv     string
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() int {
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "INFO", "set log level")
	rootCmd.PersistentFlags().BoolVar(&logJSON, "log-json", false, "set log to json format (default colorized console)")
	rootCmd.PersistentFlags().StringVar(&venv, "venv", "", "specify python3 venv path")

	ios.Init(rootCmd)
	adb.Init(rootCmd)

	err := rootCmd.Execute()
	return code.GetErrorCode(err)
}

func initLogger(logLevel string, logJSON bool) {
	// Error Logging with Stacktrace
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	// set log timestamp precise to milliseconds
	zerolog.TimeFieldFormat = "2006-01-02T15:04:05.999Z0700"

	// init log writer
	var writer io.Writer
	if !logJSON {
		// log a human-friendly, colorized output
		noColor := false
		if runtime.GOOS == "windows" {
			noColor = true
		}

		writer = zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.RFC3339Nano,
			NoColor:    noColor,
		}
		log.Info().Msg("log with colorized console")
	} else {
		// default logger
		log.Info().Msg("log with json output")
		writer = os.Stderr
	}
	log.Logger = zerolog.New(writer).With().Timestamp().Logger()

	// Setting Global Log Level
	level := strings.ToUpper(logLevel)
	log.Info().Str("log_level", level).Msg("set global log level")
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
