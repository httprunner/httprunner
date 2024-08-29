package hrp

import (
	"io"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

func InitLogger(logLevel string, logJSON bool) {
	// Error Logging with Stacktrace
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	// set log timestamp precise to milliseconds
	zerolog.TimeFieldFormat = "2006-01-02T15:04:05.999Z0700"

	// init log writer
	var msg string
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
		msg = "log with colorized console"
	} else {
		// default logger
		writer = os.Stderr
		msg = "log with json output"
	}
	log.Logger = zerolog.New(writer).With().Timestamp().Logger()
	log.Info().Msg(msg)

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
