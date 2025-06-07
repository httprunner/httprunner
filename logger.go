package hrp

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"

	"github.com/httprunner/httprunner/v5/internal/config"
)

func InitLogger(logLevel string, logJSON bool) {
	// Error Logging with Stacktrace
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	// set log timestamp precise to milliseconds
	zerolog.TimeFieldFormat = "2006-01-02T15:04:05.999Z0700"

	// init log writers
	var msg string
	var writers []io.Writer

	// console writer
	var consoleWriter io.Writer
	if !logJSON {
		// log a human-friendly, colorized output
		noColor := false
		if runtime.GOOS == "windows" {
			noColor = true
		}

		consoleWriter = zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.RFC3339Nano,
			NoColor:    noColor,
		}
		msg = "log with colorized console and file output"
	} else {
		// default logger
		consoleWriter = os.Stderr
		msg = "log with json console and file output"
	}
	writers = append(writers, consoleWriter)

	// file writer - write to results/taskID/hrp.log
	cfg := config.GetConfig()
	logFilePath := filepath.Join(cfg.ResultsPath, "hrp.log")

	// create or open log file
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		log.Error().Err(err).Str("logFilePath", logFilePath).Msg("create log file failed")
	} else {
		// add file writer to writers list
		writers = append(writers, logFile)
		log.Info().Str("logFilePath", logFilePath).Msg("log file created successfully")
	}

	// create multi writer to write to both console and file
	multiWriter := io.MultiWriter(writers...)
	log.Logger = zerolog.New(multiWriter).With().Timestamp().Logger()
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
