package sdk

import (
	"testing"
)

func TestGA4(t *testing.T) {
	ga4Client := NewGA4Client(ga4MeasurementID, ga4APISecret, false)

	event := Event{
		Name:   "hrp_debug_event",
		Params: map[string]interface{}{},
	}
	ga4Client.SendEvent(event)
}
