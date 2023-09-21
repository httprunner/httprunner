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

func TestTapUIWithScreenshot(t *testing.T) {
	serialNumber := os.Getenv("SERIAL_NUMBER")
	device, _ := NewAndroidDevice(WithSerialNumber(serialNumber))
	driver, err := device.NewDriver()
	if err != nil {
		t.Fatal(err)
	}

	err = driver.TapByUIDetection(
		WithScreenShotUITypes("dyhouse", "shoppingbag"))
	if err != nil {
		t.Fatal(err)
	}
}

func TestDriverExtOCR(t *testing.T) {
	driverExt, err := iosDevice.NewDriver()
	checkErr(t, err)

	point, err := driverExt.FindScreenText("抖音")
	checkErr(t, err)

	t.Logf("point.X: %v, point.Y: %v", point.X, point.Y)
	driverExt.Driver.TapFloat(point.X, point.Y-20)
}

func TestClosePopup(t *testing.T) {
	setupAndroid(t)

	screenResult, err := driverExt.GetScreenResult(
		WithScreenShotClosePopups(true), WithScreenShotUpload(true))
	if err != nil {
		t.Logf("get screen result failed for popup handler: %v", err)
		return
	}
	t.Logf("screen result: %v", screenResult)

	if screenResult.Popup == nil {
		t.Log("there are no popups here")
		return
	}
	t.Logf("popup info: %v", screenResult.Popup)

	if err = driverExt.tapPopupHandler(screenResult.Popup); err != nil {
		t.Fatal(err)
	}
}
