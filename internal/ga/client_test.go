package ga

import (
	"testing"
)

func TestSendEvents(t *testing.T) {
	event := EventTracking{
		Category: "unittest",
		Action:   "SendEvents",
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
		Label:    "StructToUrlValues",
	}
	val := structToUrlValues(event)
	if val.Encode() != "ea=convert&ec=unittest&el=StructToUrlValues" {
		t.Fail()
	}
}
