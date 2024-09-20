//go:build localtest

package uixt

import (
	"fmt"
	"testing"
)

var harmonyDriverExt *DriverExt

func setupHarmonyDevice(t *testing.T) {
	device, err := NewHarmonyDevice()
	if err != nil {
		t.Fatal(err)
	}
	driver, err = device.NewUSBDriver()
	if err != nil {
		t.Fatal(err)
	}
	harmonyDriverExt, err = newDriverExt(device, driver)
	if err != nil {
		t.Fatal(err)
	}
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
	err := harmonyDriverExt.TapAbsXY(200, 2000)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSwipe(t *testing.T) {
	setupHarmonyDevice(t)
	err := harmonyDriverExt.SwipeLeft()
	if err != nil {
		t.Fatal(err)
	}
}

func TestInput(t *testing.T) {
	setupHarmonyDevice(t)
	err := harmonyDriverExt.Input("test")
	if err != nil {
		t.Fatal(err)
	}
}

func TestHomeScreen(t *testing.T) {
	setupHarmonyDevice(t)
	err := driver.Homescreen()
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
	err := driver.PressBack()
	if err != nil {
		t.Fatal(err)
	}
}

func TestScreenshot(t *testing.T) {
	setupHarmonyDevice(t)
	screenshot, err := driver.Screenshot()
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
	appInfo, err := driver.GetForegroundApp()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(appInfo)
}
