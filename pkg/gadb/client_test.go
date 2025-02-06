//go:build localtest

package gadb

import (
	"os"
	"testing"
)

var adbClient Client

func setupClient(t *testing.T) {
	var err error
	adbClient, err = NewClient()
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_ServerVersion(t *testing.T) {
	setupClient(t)

	adbServerVersion, err := adbClient.ServerVersion()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(adbServerVersion)
}

func TestClient_DeviceSerialList(t *testing.T) {
	setupClient(t)

	serials, err := adbClient.DeviceSerialList()
	if err != nil {
		t.Fatal(err)
	}

	for i := range serials {
		t.Log(serials[i])
	}
}

func TestClient_DeviceList(t *testing.T) {
	setupDevices(t)

	for i := range devices {
		t.Log(devices[i].serial, devices[i].DeviceInfo())
	}
}

func TestClient_ForwardList(t *testing.T) {
	setupClient(t)

	deviceForwardList, err := adbClient.ForwardList()
	if err != nil {
		t.Fatal(err)
	}

	for i := range deviceForwardList {
		t.Log(deviceForwardList[i])
	}
}

func TestClient_ForwardKillAll(t *testing.T) {
	setupClient(t)

	err := adbClient.ForwardKillAll()
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_Connect(t *testing.T) {
	setupClient(t)

	err := adbClient.Connect("192.168.1.28")
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_Disconnect(t *testing.T) {
	setupClient(t)

	err := adbClient.Disconnect("192.168.1.28")
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_DisconnectAll(t *testing.T) {
	setupClient(t)

	err := adbClient.DisconnectAll()
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_KillServer(t *testing.T) {
	setupClient(t)

	err := adbClient.KillServer()
	if err != nil {
		t.Fatal(err)
	}
}

func TestScreenCap(t *testing.T) {
	setupDevices(t)

	for _, d := range devices {
		res, err := d.ScreenCap()
		if err != nil {
			t.Error(err)
		}
		t.Log(len(res))
		os.WriteFile("/tmp/1.png", res, 0o644)
	}
}
