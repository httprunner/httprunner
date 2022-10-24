//go:build localtest

package gidevice

import (
	"testing"
	"time"

	"github.com/httprunner/httprunner/v4/hrp/pkg/gidevice/pkg/libimobiledevice"
)

var um Usbmux

func setupUsbmux(t *testing.T) {
	var err error
	um, err = NewUsbmux()
	if err != nil {
		t.Fatal(err)
	}
}

func Test_usbmux_Devices(t *testing.T) {
	setupUsbmux(t)

	devices, err := um.Devices()
	if err != nil {
		t.Fatal(err)
	}

	for _, dev := range devices {
		t.Log(dev.Properties().SerialNumber, dev.Properties().ProductID, dev.Properties().DeviceID)
	}
}

func Test_usbmux_ReadBUID(t *testing.T) {
	setupUsbmux(t)

	buid, err := um.ReadBUID()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(buid)
}

func Test_usbmux_Listen(t *testing.T) {
	setupUsbmux(t)

	devNotifier := make(chan Device)
	cancelFunc, err := um.Listen(devNotifier)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		time.Sleep(20 * time.Second)
		cancelFunc()
	}()

	for dev := range devNotifier {
		if dev.Properties().ConnectionType != "" {
			t.Log(dev.Properties().SerialNumber, dev.Properties().ProductID, dev.Properties().DeviceID)
		} else {
			t.Log(libimobiledevice.MessageTypeDeviceRemove, dev.Properties().DeviceID)
		}
	}

	time.Sleep(5 * time.Second)
	t.Log("Done")
}
