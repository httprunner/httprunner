package hrp

import (
	"os"
	"strings"

	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
)

var log = zlog.Logger

func SetLogLevel(level string) {
	level = strings.ToUpper(level)
	log.Info().Msgf("Set log level to %s", level)
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

func SetLogPretty() {
	log = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Info().Msg("Set log to pretty console")
}

func GetLogger() zerolog.Logger {
	return log
}
