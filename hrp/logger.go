package hrp

import (
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

	if !logJSON {
		// log a human-friendly, colorized output
		noColor := false
		if runtime.GOOS == "windows" {
			noColor = true
		}

		log.Logger = log.Output(
			zerolog.ConsoleWriter{
				Out:        os.Stderr,
				TimeFormat: time.RFC3339,
				NoColor:    noColor,
			},
		)
		log.Info().Msg("log with colorized console")
	} else {
		log.Info().Msg("log with json output")
		// default logger:
		// zerolog.New(os.Stderr).With().Timestamp().Logger()
	}

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
