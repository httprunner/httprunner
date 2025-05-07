//go:build localtest

package uixt

import (
	"fmt"
	"testing"

	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupHDCDriverExt(t *testing.T) *XTDriver {
	device, err := NewHarmonyDevice()
	require.Nil(t, err)
	hdcDriver, err := NewHDCDriver(device)
	require.Nil(t, err)
	driverExt, err := NewXTDriver(hdcDriver, option.WithCVService(option.CVServiceTypeVEDEM))
	require.Nil(t, err)
	return driverExt
}

func TestWindowSize(t *testing.T) {
	driver := setupHDCDriverExt(t)
	size, err := driver.WindowSize()
	assert.Nil(t, err)
	t.Log(fmt.Sprintf("width: %d, height: %d", size.Width, size.Height))
}

func TestHarmonyTap(t *testing.T) {
	driver := setupHDCDriverExt(t)
	err := driver.TapAbsXY(200, 2000)
	assert.Nil(t, err)
}

func TestHarmonySwipe(t *testing.T) {
	driver := setupHDCDriverExt(t)
	err := driver.Swipe(0.5, 0.5, 0.1, 0.5)
	assert.Nil(t, err)
}

func TestHarmonyInput(t *testing.T) {
	driver := setupHDCDriverExt(t)
	err := driver.Input("test")
	assert.Nil(t, err)
}

func TestHomeScreen(t *testing.T) {
	driver := setupHDCDriverExt(t)
	err := driver.Home()
	assert.Nil(t, err)
}

func TestUnlock(t *testing.T) {
	driver := setupHDCDriverExt(t)
	err := driver.Unlock()
	assert.Nil(t, err)
}

func TestPressBack(t *testing.T) {
	driver := setupHDCDriverExt(t)
	err := driver.Back()
	assert.Nil(t, err)
}

func TestScreenshot(t *testing.T) {
	driver := setupHDCDriverExt(t)
	screenshot, err := driver.ScreenShot()
	assert.Nil(t, err)
	t.Log(screenshot)
}

func TestLaunch(t *testing.T) {
	driver := setupHDCDriverExt(t)
	err := driver.AppLaunch("")
	assert.Nil(t, err)
}

func TestForegroundApp(t *testing.T) {
	driver := setupHDCDriverExt(t)
	appInfo, err := driver.ForegroundInfo()
	assert.Nil(t, err)
	t.Log(appInfo)
}
