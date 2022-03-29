package boomer

const (
	typeClientReady      = "client_ready"
	typeClientStopped    = "client_stopped"
	typeHeartbeat        = "heartbeat"
	typeSpawning         = "spawning"
	typeSpawningComplete = "spawning_complete"
	typeQuit             = "quit"
	typeException        = "exception"
)

type genericMessage struct {
	Type   string           `codec:"type"`
	Data   map[string]int64 `codec:"data"`
	NodeID string           `codec:"node_id"`
}

func newGenericMessage(t string, data map[string]int64, nodeID string) (msg *genericMessage) {
	return &genericMessage{
		Type:   t,
		Data:   data,
		NodeID: nodeID,
	}
}

func newClientReadyMessage(nodeID string) (msg *genericMessage) {
	return &genericMessage{
		Type:   "client_ready",
		Data:   map[string]int64{},
		NodeID: nodeID,
	}
}
