package ga

const (
	trackingID = "UA-114587036-1" // Tracking ID for Google Analytics
)

var gaClient *GAClient

func init() {
	gaClient = NewGAClient(trackingID)
}

func SendEvent(e IEvent) error {
	return gaClient.SendEvent(e)
}
