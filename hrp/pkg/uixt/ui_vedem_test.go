package uixt

import (
	"bytes"
	"fmt"
	"os"
	"testing"
)

func checkUI(buff *bytes.Buffer) error {
	service, err := newVEDEMUIService([]string{"dyhouse", "shoppingbag"})
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

func TestUIWithScreenshot(t *testing.T) {
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

	if err := checkUI(raw); err != nil {
		t.Fatal(err)
	}
}

func TestTapUIWithScreenshot(t *testing.T) {
	serialNumber := os.Getenv("SERIAL_NUMBER")
	device, _ := NewAndroidDevice(WithSerialNumber(serialNumber))
	driver, err := device.NewDriver(nil)
	if err != nil {
		t.Fatal(err)
	}

	err = driver.TapByUIDetection([]string{"dyhouse", "shoppingbag"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestUIWithLocalFile(t *testing.T) {
	imagePath := os.Getenv("IMAGE_PATH")
	file, err := os.ReadFile(imagePath)
	b := bytes.NewBuffer(file)
	if err != nil {
		t.Fatal(err)
	}

	if err := checkUI(b); err != nil {
		t.Fatal(err)
	}
}
