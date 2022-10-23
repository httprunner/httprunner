package gadb

import (
	"testing"
)

func TestClient_ServerVersion(t *testing.T) {
	SetDebug(true)

	adbClient, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	adbServerVersion, err := adbClient.ServerVersion()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(adbServerVersion)
}

func TestClient_DeviceSerialList(t *testing.T) {
	SetDebug(true)

	adbClient, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	serials, err := adbClient.DeviceSerialList()
	if err != nil {
		t.Fatal(err)
	}

	for i := range serials {
		t.Log(serials[i])
	}
}

func TestClient_DeviceList(t *testing.T) {
	SetDebug(true)

	adbClient, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	devices, err := adbClient.DeviceList()
	if err != nil {
		t.Fatal(err)
	}

	for i := range devices {
		t.Log(devices[i].serial, devices[i].DeviceInfo())
	}
}

func TestClient_ForwardList(t *testing.T) {
	SetDebug(true)

	adbClient, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	deviceForwardList, err := adbClient.ForwardList()
	if err != nil {
		t.Fatal(err)
	}

	for i := range deviceForwardList {
		t.Log(deviceForwardList[i])
	}
}

func TestClient_ForwardKillAll(t *testing.T) {
	SetDebug(true)

	adbClient, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	err = adbClient.ForwardKillAll()
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_Connect(t *testing.T) {
	adbClient, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	SetDebug(true)

	err = adbClient.Connect("192.168.1.28")
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_Disconnect(t *testing.T) {
	adbClient, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	SetDebug(true)

	err = adbClient.Disconnect("192.168.1.28")
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_DisconnectAll(t *testing.T) {
	adbClient, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	SetDebug(true)

	err = adbClient.DisconnectAll()
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_KillServer(t *testing.T) {
	SetDebug(true)

	adbClient, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	err = adbClient.KillServer()
	if err != nil {
		t.Fatal(err)
	}
}
