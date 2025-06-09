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

	// parse console log level
	consoleLevel := parseLogLevel(logLevel)

	// file writer - write to results/taskID/hrp.log
	cfg := config.GetConfig()
	logFilePath := filepath.Join(cfg.ResultsPath, "hrp.log")

	// create or open log file
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		// if file creation failed, use console logger only
		log.Logger = zerolog.New(consoleWriter).With().Timestamp().Logger().Level(consoleLevel)
		log.Error().Err(err).Str("logFilePath", logFilePath).Msg("create log file failed")
	} else {
		// create a custom writer that applies different log levels
		multiWriter := &leveledMultiWriter{
			consoleWriter: consoleWriter,
			consoleLevel:  consoleLevel,
			fileWriter:    logFile,
			fileLevel:     zerolog.DebugLevel,
		}
		log.Logger = zerolog.New(multiWriter).With().Timestamp().Logger()
		log.Info().Str("logFilePath", logFilePath).Msg("log file created successfully")
	}

	log.Info().Msg(msg)
	log.Info().Str("console_log_level", strings.ToUpper(logLevel)).Str("file_log_level", "DEBUG").Msg("logger initialized with different levels")
}

// parseLogLevel converts string log level to zerolog.Level
func parseLogLevel(logLevel string) zerolog.Level {
	level := strings.ToUpper(logLevel)
	switch level {
	case "DEBUG":
		return zerolog.DebugLevel
	case "INFO":
		return zerolog.InfoLevel
	case "WARN":
		return zerolog.WarnLevel
	case "ERROR":
		return zerolog.ErrorLevel
	case "FATAL":
		return zerolog.FatalLevel
	case "PANIC":
		return zerolog.PanicLevel
	default:
		return zerolog.InfoLevel
	}
}

// leveledMultiWriter is a custom writer that applies different log levels to different outputs
type leveledMultiWriter struct {
	consoleWriter io.Writer
	consoleLevel  zerolog.Level
	fileWriter    io.Writer
	fileLevel     zerolog.Level
}

func (w *leveledMultiWriter) Write(p []byte) (n int, err error) {
	// Parse the log level from the JSON log entry
	logLevel := extractLogLevel(p)

	var writeErrors []error

	// Write to console if log level meets console threshold
	if logLevel >= w.consoleLevel {
		if _, err := w.consoleWriter.Write(p); err != nil {
			writeErrors = append(writeErrors, err)
		}
	}

	// Write to file if log level meets file threshold (always debug, so always write)
	if logLevel >= w.fileLevel {
		if _, err := w.fileWriter.Write(p); err != nil {
			writeErrors = append(writeErrors, err)
		}
	}

	// Return the length of the original message and any write errors
	if len(writeErrors) > 0 {
		return len(p), writeErrors[0]
	}
	return len(p), nil
}

// extractLogLevel extracts the log level from a JSON log entry
func extractLogLevel(p []byte) zerolog.Level {
	// Simple parsing to extract level from JSON
	logStr := string(p)
	if strings.Contains(logStr, `"level":"debug"`) {
		return zerolog.DebugLevel
	} else if strings.Contains(logStr, `"level":"info"`) {
		return zerolog.InfoLevel
	} else if strings.Contains(logStr, `"level":"warn"`) {
		return zerolog.WarnLevel
	} else if strings.Contains(logStr, `"level":"error"`) {
		return zerolog.ErrorLevel
	} else if strings.Contains(logStr, `"level":"fatal"`) {
		return zerolog.FatalLevel
	} else if strings.Contains(logStr, `"level":"panic"`) {
		return zerolog.PanicLevel
	}
	return zerolog.InfoLevel // default
}
