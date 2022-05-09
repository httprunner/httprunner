package sdk

import (
	"fmt"
	"os"

	"github.com/denisbrodbeck/machineid"
	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/version"
)

const (
	trackingID = "UA-114587036-1" // Tracking ID for Google Analytics
	sentryDSN  = "https://cff5efc69b1a4325a4cf873f1e70c13a@o334324.ingest.sentry.io/6070292"
)

var gaClient *GAClient

func init() {
	// init GA client
	clientID, err := machineid.ProtectedID("hrp")
	if err != nil {
		nodeUUID, _ := uuid.NewUUID()
		clientID = nodeUUID.String()
	}
	gaClient = NewGAClient(trackingID, clientID)

	// init sentry sdk
	if os.Getenv("DISABLE_SENTRY") == "true" {
		return
	}
	err = sentry.Init(sentry.ClientOptions{
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
			ID: clientID,
		})
	})
}

func SendEvent(e IEvent) error {
	if os.Getenv("DISABLE_GA") == "true" {
		// do not send GA events in CI environment
		return nil
	}
	return gaClient.SendEvent(e)
}
