//go:build localtest

package gadb

import (
	"testing"
)

func Test_transport_VerifyResponse(t *testing.T) {
	transport, err := newTransport("localhost:5037")
	if err != nil {
		t.Fatal(err)
	}
	defer transport.Close()

	err = transport.Send("host:version")
	if err != nil {
		t.Fatal(err)
	}

	err = transport.VerifyResponse()
	if err != nil {
		t.Fatal(err)
	}
}
