package sdk

import (
	"github.com/httprunner/httprunner/v4/hrp/internal/env"
)

const (
	trackingID = "UA-114587036-1" // Tracking ID for Google Analytics
)

var gaClient *GAClient

func init() {
	// init GA client
	gaClient = NewGAClient(trackingID, userID)
}

func SendEvent(e IEvent) error {
	if env.DISABLE_GA == "true" {
		// do not send GA events in CI environment
		return nil
	}
	return gaClient.SendEvent(e)
}
