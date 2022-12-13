//go:build localtest

package uixt

import (
	"bytes"
	"fmt"
	"os"
	"testing"
)

func checkOCR(buff *bytes.Buffer) error {
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

	if err := checkOCR(raw); err != nil {
		t.Fatal(err)
	}
}

func TestOCRWithLocalFile(t *testing.T) {
	imagePath := "/Users/debugtalk/Downloads/s1.png"

	file, err := os.ReadFile(imagePath)
	if err != nil {
		t.Fatal(err)
	}

	buf := new(bytes.Buffer)
	buf.Read(file)

	if err := checkOCR(buf); err != nil {
		t.Fatal(err)
	}
}
