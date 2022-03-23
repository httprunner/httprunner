package ga

import (
	"github.com/denisbrodbeck/machineid"
	"github.com/google/uuid"
)

const (
	trackingID = "UA-114587036-1" // Tracking ID for Google Analytics
)

var gaClient *GAClient

func init() {
	clientID, err := machineid.ProtectedID("hrp")
	if err != nil {
		nodeUUID, _ := uuid.NewUUID()
		clientID = nodeUUID.String()
	}
	gaClient = NewGAClient(trackingID, clientID)
}

func SendEvent(e IEvent) error {
	return gaClient.SendEvent(e)
}
