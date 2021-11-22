package sentry

import (
	"time"

	"github.com/getsentry/sentry-go"

	"github.com/httprunner/hrp/internal/version"
)

func init() {
	// init sentry sdk
	err := sentry.Init(sentry.ClientOptions{
		Dsn:     "https://cff5efc69b1a4325a4cf873f1e70c13a@o334324.ingest.sentry.io/6070292",
		Release: version.VERSION,
	})
	if err != nil {
		panic("init sentry sdk failed!")
	}
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetLevel(sentry.LevelError)
	})
}

func Flush() {
	sentry.Flush(3 * time.Second)
}
