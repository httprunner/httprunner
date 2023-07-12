package uixt

import (
	"bytes"
	"fmt"
	"os"
	"testing"
)

func checkSD(buff *bytes.Buffer) error {
	service, err := newVEDEMSDService()
	if err != nil {
		return err
	}
	imageResult, err := service.GetImage(buff)
	if err != nil {
		return err
	}
	fmt.Printf("%+v\n", imageResult.LiveType)
	return nil
}

func TestSDWithScreenshot(t *testing.T) {
	serialNumber := os.Getenv("SERIAL_NUMBER")
	device, _ := NewAndroidDevice(WithSerialNumber(serialNumber))
	driver, err := device.NewUSBDriver(nil)
	if err != nil {
		t.Fatal(err)
	}

	raw, err := driver.Screenshot()
	if err != nil {
		t.Fatal(err)
	}

	if err := checkSD(raw); err != nil {
		t.Fatal(err)
	}
}
