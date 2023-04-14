//go:build localtest

package gadb

import (
	"io/ioutil"
	"testing"
)

func TestClient_ServerVersion(t *testing.T) {
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

	err = adbClient.DisconnectAll()
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_KillServer(t *testing.T) {
	adbClient, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	err = adbClient.KillServer()
	if err != nil {
		t.Fatal(err)
	}
}

func TestScreenCap(t *testing.T) {
	adbClient, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	dl, err := adbClient.DeviceList()
	if err != nil {
		t.Error(err)
	}
	d := dl[0]
	res, err := d.ScreenCap()
	if err != nil {
		t.Error(err)
	}
	t.Log(len(res))
	ioutil.WriteFile("/tmp/1.png", res, 0o644)
}
