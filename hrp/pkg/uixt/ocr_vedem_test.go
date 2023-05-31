//go:build localtest

package uixt

import (
	"bytes"
	"fmt"
	"os"
	"testing"
)

func checkOCR(buff *bytes.Buffer) error {
	service, err := newVEDEMImageService()
	if err != nil {
		return err
	}
	imageResult, err := service.GetImage(buff)
	if err != nil {
		return err
	}
	fmt.Println(fmt.Sprintf("imageResult: %v", imageResult))
	return nil
}

func TestOCRWithScreenshot(t *testing.T) {
	setupAndroid(t)

	raw, err := driverExt.Driver.Screenshot()
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
