//go:build localtest

package uixt

import (
	"fmt"
	"os"
	"testing"
)

func checkOCR(buff []byte) error {
	service, err := newVEDEMOCRService()
	if err != nil {
		return err
	}
	ocrResults, err := service.getOCRResult(buff)
	if err != nil {
		return err
	}
	fmt.Println(ocrResults)
	return nil
}

func TestOCRWithScreenshot(t *testing.T) {
	device, _ := NewAndroidDevice()
	driver, err := device.NewUSBDriver(nil)
	if err != nil {
		t.Fatal(err)
	}

	raw, err := driver.Screenshot()
	if err != nil {
		t.Fatal(err)
	}

	if err := checkOCR(raw.Bytes()); err != nil {
		t.Fatal(err)
	}
}

func TestOCRWithLocalFile(t *testing.T) {
	imagePath := "/Users/debugtalk/Downloads/s1.png"
	file, err := os.ReadFile(imagePath)
	if err != nil {
		t.Fatal(err)
	}

	if err := checkOCR(file); err != nil {
		t.Fatal(err)
	}
}
