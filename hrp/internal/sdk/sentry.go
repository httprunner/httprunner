package sdk

import (
	"fmt"

	"github.com/getsentry/sentry-go"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/env"
	"github.com/httprunner/httprunner/v4/hrp/internal/version"
)

const (
	sentryDSN = "https://cff5efc69b1a4325a4cf873f1e70c13a@o334324.ingest.sentry.io/6070292"
)

func init() {
	// init sentry sdk
	if env.DISABLE_SENTRY == "true" {
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
