package uixt

import (
	"bytes"
	"fmt"
	"os"
	"testing"
)

func checkCP(buff *bytes.Buffer) error {
	service, err := newVEDEMCPService()
	if err != nil {
		return err
	}
	imageResult, err := service.GetImage(buff)
	if err != nil {
		return err
	}
	fmt.Printf("%+v\n", imageResult)
	return nil
}

func TestCPWithScreenshot(t *testing.T) {
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

	if err := checkCP(raw); err != nil {
		t.Fatal(err)
	}
}

func TestCPWithLocalFile(t *testing.T) {
	imagePath := os.Getenv("IMAGE_PATH")
	file, err := os.ReadFile(imagePath)
	b := bytes.NewBuffer(file)
	if err != nil {
		t.Fatal(err)
	}

	if err := checkCP(b); err != nil {
		t.Fatal(err)
	}
}
