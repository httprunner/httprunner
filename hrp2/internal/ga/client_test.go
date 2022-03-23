package ga

import (
	"testing"
)

func TestSendEvents(t *testing.T) {
	event := EventTracking{
		Category: "unittest",
		Action:   "SendEvents",
		Value:    123,
	}
	err := gaClient.SendEvent(event)
	if err != nil {
		t.Fatal(err)
	}
}

func TestStructToUrlValues(t *testing.T) {
	event := EventTracking{
		Category: "unittest",
		Action:   "convert",
		Label:    "v0.3.0",
		Value:    123,
	}
	val := structToUrlValues(event)
	if val.Encode() != "ea=convert&ec=unittest&el=v0.3.0&ev=123" {
		t.Fail()
	}
}
