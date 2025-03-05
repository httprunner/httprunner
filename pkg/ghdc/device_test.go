package ghdc

import (
	"fmt"
	"testing"
)

func TestDevice_Product(t *testing.T) {
	hdClient, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	devices, err := hdClient.DeviceList()
	if err != nil {
		t.Fatal(err)
	}

	for i := range devices {
		dev := devices[i]
		product, err := dev.Product()
		if err != nil {
			t.Fatal(err)
		}
		t.Log(dev.Serial(), product)
	}
}

func TestDevice_Model(t *testing.T) {
	hdClient, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	devices, err := hdClient.DeviceList()
	if err != nil {
		t.Fatal(err)
	}

	for i := range devices {
		dev := devices[i]
		model, err := dev.Model()
		if err != nil {
			t.Fatal(err)
		}
		t.Log(dev.Serial(), model)
	}
}

func TestDevice_Usb(t *testing.T) {
	hdClient, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	devices, err := hdClient.DeviceList()
	if err != nil {
		t.Fatal(err)
	}

	for i := range devices {
		dev := devices[i]
		usb, err := dev.Usb()
		if err != nil {
			t.Fatal(err)
		}
		isUsb, err := dev.IsUsb()
		if err != nil {
			t.Fatal(err)
		}
		t.Log(dev.Serial(), usb, isUsb)
	}
}

func TestDevice_DeviceInfo(t *testing.T) {
	hdClient, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	devices, err := hdClient.DeviceList()
	if err != nil {
		t.Fatal(err)
	}

	for i := range devices {
		dev := devices[i]
		t.Log(dev.DeviceInfo())
	}
}

func TestDevice_Forward(t *testing.T) {
	hdClient, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	devices, err := hdClient.DeviceList()
	if err != nil {
		t.Fatal(err)
	}
	if len(devices) == 0 {
		t.Fatal("not found available device")
	}
	SetDebug(true)

	localPort := 61000
	localPort, err = devices[0].Forward(6790)
	t.Log(fmt.Sprintf("forward local port %d \n", localPort))
	if err != nil {
		t.Fatal(err)
	}

	err = devices[0].ForwardKill(localPort)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDevice_ForwardKill(t *testing.T) {
	hdClient, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	devices, err := hdClient.DeviceList()
	if err != nil {
		t.Fatal(err)
	}
	if len(devices) == 0 {
		t.Fatal("not found available device")
	}
	SetDebug(true)

	err = devices[0].ForwardKill(6790)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDevice_RunShellCommand(t *testing.T) {
	hdClient, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	devices, err := hdClient.DeviceList()
	if err != nil {
		t.Fatal(err)
	}
	if len(devices) == 0 {
		t.Fatal("not found available device")
	}
	dev := devices[0]

	cmdOutput, err := dev.RunShellCommand("pwd")
	if err != nil {
		t.Fatal(dev.serial, err)
	}
	t.Log("\n⬇️"+dev.serial+"⬇️\n", cmdOutput)
}

func TestDevice_Push(t *testing.T) {
	hdClient, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	devices, err := hdClient.DeviceList()
	if err != nil {
		t.Fatal(err)
	}
	if len(devices) == 0 {
		t.Fatal("not found available device")
	}
	dev := devices[0]

	SetDebug(true)

	err = dev.PushFile("/tmp/test.txt", "/data/local/tmp/push.txt")
	if err != nil {
		t.Fatal(err)
	}
}

func TestDevice_Pull(t *testing.T) {
	hdClient, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	devices, err := hdClient.DeviceList()
	if err != nil {
		t.Fatal(err)
	}
	if len(devices) == 0 {
		t.Fatal("not found available device")
	}
	dev := devices[0]

	SetDebug(true)

	err = dev.PullFile("/data/local/tmp/push.txt", "/tmp/test2.txt")
	if err != nil {
		t.Fatal(err)
	}
}

func TestDevice_Screenshot(t *testing.T) {
	hdClient, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	devices, err := hdClient.DeviceList()
	if err != nil {
		t.Fatal(err)
	}
	if len(devices) == 0 {
		t.Fatal("not found available device")
	}
	dev := devices[0]

	SetDebug(true)

	err = dev.Screenshot("/tmp/test.jpeg")
	if err != nil {
		t.Fatal(err)
	}
}

func TestDevice_GetSoVersion(t *testing.T) {
	hdClient, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	devices, err := hdClient.DeviceList()
	if err != nil {
		t.Fatal(err)
	}
	if len(devices) == 0 {
		t.Fatal("not found available device")
	}
	dev := devices[0]

	SetDebug(true)
	res, err := dev.RunShellCommand("cat data/local/tmp/agent.so |grep -a UITEST_AGENT_LIBRARY")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(res)
}
