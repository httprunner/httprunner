//go:build localtest

package uixt

import (
	"fmt"
	"testing"

	"github.com/httprunner/httprunner/v5/pkg/uixt/ai"
)

var (
	hdcDriver    *HDCDriver
	hdcDriverExt *XTDriver
)

func setupHarmonyDevice(t *testing.T) {
	device, err := NewHarmonyDevice()
	if err != nil {
		t.Fatal(err)
	}
	hdcDriver, err = NewHDCDriver(device)
	if err != nil {
		t.Fatal(err)
	}
	hdcDriverExt = NewXTDriver(hdcDriver,
		ai.WithCVService(ai.CVServiceTypeVEDEM))
}

func TestWindowSize(t *testing.T) {
	setupHarmonyDevice(t)
	size, err := driver.WindowSize()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(fmt.Sprintf("width: %d, height: %d", size.Width, size.Height))
}

func TestHarmonyTap(t *testing.T) {
	setupHarmonyDevice(t)
	err := hdcDriverExt.TapAbsXY(200, 2000)
	if err != nil {
		t.Fatal(err)
	}
}

func TestHarmonySwipe(t *testing.T) {
	setupHarmonyDevice(t)
	err := hdcDriverExt.SwipeLeft()
	if err != nil {
		t.Fatal(err)
	}
}

func TestHarmonyInput(t *testing.T) {
	setupHarmonyDevice(t)
	err := hdcDriver.Input("test")
	if err != nil {
		t.Fatal(err)
	}
}

func TestHomeScreen(t *testing.T) {
	setupHarmonyDevice(t)
	err := driver.Home()
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnlock(t *testing.T) {
	setupHarmonyDevice(t)
	err := driver.Unlock()
	if err != nil {
		t.Fatal(err)
	}
}

func TestPressBack(t *testing.T) {
	setupHarmonyDevice(t)
	err := driver.Back()
	if err != nil {
		t.Fatal(err)
	}
}

func TestScreenshot(t *testing.T) {
	setupHarmonyDevice(t)
	screenshot, err := driver.ScreenShot()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(screenshot)
}

func TestLaunch(t *testing.T) {
	setupHarmonyDevice(t)
	err := driver.AppLaunch("")
	if err != nil {
		t.Fatal(err)
	}
}

func TestForegroundApp(t *testing.T) {
	setupHarmonyDevice(t)
	appInfo, err := driver.ForegroundInfo()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(appInfo)
}
