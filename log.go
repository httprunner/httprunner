package hrp

import (
	"os"
	"strings"

	"github.com/getsentry/sentry-go"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
)

func init() {
	// init sentry sdk
	err := sentry.Init(sentry.ClientOptions{
		Dsn:     "https://cff5efc69b1a4325a4cf873f1e70c13a@o334324.ingest.sentry.io/6070292",
		Release: VERSION,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("init sentry sdk failed!")
	}
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetLevel(sentry.LevelError)
	})
}

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
