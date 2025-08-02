package ghdc

import (
	"testing"
)

func TestClient_ServerVersion(t *testing.T) {
	t.Skip("Skip HDC test - requires HarmonyOS Device Client server and connected device")
	SetDebug(true)

	hdcClient, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	hdServerVersion, err := hdcClient.ServerVersion()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(hdServerVersion)
}

func TestClient_DeviceSerialList(t *testing.T) {
	t.Skip("Skip HDC test - requires HarmonyOS Device Client server and connected device")
	SetDebug(true)

	hdcClient, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	serials, err := hdcClient.DeviceSerialList()
	if err != nil {
		t.Fatal(err)
	}

	for i := range serials {
		t.Log(serials[i])
	}
}

func TestClient_DeviceList(t *testing.T) {
	t.Skip("Skip HDC test - requires HarmonyOS Device Client server and connected device")
	SetDebug(true)

	hdcClient, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	devices, err := hdcClient.DeviceList()
	if err != nil {
		t.Fatal(err)
	}

	for i := range devices {
		t.Log(devices[i].serial, devices[i].DeviceInfo())
	}
}

func TestClient_ForwardList(t *testing.T) {
	t.Skip("Skip HDC test - requires HarmonyOS Device Client server and connected device")
	SetDebug(true)

	hdcClient, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	deviceForwardList, err := hdcClient.ForwardList()
	if err != nil {
		t.Fatal(err)
	}

	for i := range deviceForwardList {
		t.Log(deviceForwardList[i])
	}
}

func TestClient_Connect(t *testing.T) {
	t.Skip("Skip HDC test - requires HarmonyOS Device Client server and connected device")
	hdcClient, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	SetDebug(true)

	err = hdcClient.Connect("192.168.1.28")
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_KillServer(t *testing.T) {
	t.Skip("Skip HDC test - requires HarmonyOS Device Client server and connected device")
	SetDebug(true)

	hdcClient, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	err = hdcClient.KillServer()
	if err != nil {
		t.Fatal(err)
	}
}
