package sdk

import (
	"fmt"
	"os"

	"github.com/getsentry/sentry-go"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/internal/version"
)

const (
	sentryDSN = "https://cff5efc69b1a4325a4cf873f1e70c13a@o334324.ingest.sentry.io/6070292"
)

func init() {
	// init sentry sdk
	if os.Getenv("DISABLE_SENTRY") == "true" {
		return
	}
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              sentryDSN,
		Release:          fmt.Sprintf("httprunner@%s", version.VERSION),
		AttachStacktrace: true,
	})
	if err != nil {
		log.Error().Err(err).Msg("init sentry sdk failed!")
		return
	}
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetLevel(sentry.LevelError)
		scope.SetUser(sentry.User{
			ID: userID,
		})
	})
}
