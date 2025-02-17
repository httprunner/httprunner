//go:build localtest

package uixt

import (
	"fmt"
	"testing"

	"github.com/httprunner/httprunner/v5/pkg/uixt/ai"
)

func setupHDCDriverExt(t *testing.T) *XTDriver {
	device, err := NewHarmonyDevice()
	if err != nil {
		t.Fatal(err)
	}
	hdcDriver, err := NewHDCDriver(device)
	if err != nil {
		t.Fatal(err)
	}
	return NewXTDriver(hdcDriver, ai.WithCVService(ai.CVServiceTypeVEDEM))
}

func TestWindowSize(t *testing.T) {
	driver := setupHDCDriverExt(t)
	size, err := driver.WindowSize()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(fmt.Sprintf("width: %d, height: %d", size.Width, size.Height))
}

func TestHarmonyTap(t *testing.T) {
	driver := setupHDCDriverExt(t)
	err := driver.TapAbsXY(200, 2000)
	if err != nil {
		t.Fatal(err)
	}
}

func TestHarmonySwipe(t *testing.T) {
	driver := setupHDCDriverExt(t)
	err := driver.Swipe(0.5, 0.5, 0.1, 0.5)
	if err != nil {
		t.Fatal(err)
	}
}

func TestHarmonyInput(t *testing.T) {
	driver := setupHDCDriverExt(t)
	err := driver.Input("test")
	if err != nil {
		t.Fatal(err)
	}
}

func TestHomeScreen(t *testing.T) {
	driver := setupHDCDriverExt(t)
	err := driver.Home()
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnlock(t *testing.T) {
	driver := setupHDCDriverExt(t)
	err := driver.Unlock()
	if err != nil {
		t.Fatal(err)
	}
}

func TestPressBack(t *testing.T) {
	driver := setupHDCDriverExt(t)
	err := driver.Back()
	if err != nil {
		t.Fatal(err)
	}
}

func TestScreenshot(t *testing.T) {
	driver := setupHDCDriverExt(t)
	screenshot, err := driver.ScreenShot()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(screenshot)
}

func TestLaunch(t *testing.T) {
	driver := setupHDCDriverExt(t)
	err := driver.AppLaunch("")
	if err != nil {
		t.Fatal(err)
	}
}

func TestForegroundApp(t *testing.T) {
	driver := setupHDCDriverExt(t)
	appInfo, err := driver.ForegroundInfo()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(appInfo)
}
