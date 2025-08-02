//go:build localtest

package uixt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/httprunner/httprunner/v5/uixt/option"
)

func setupHDCDriverExt(t *testing.T) *XTDriver {
	// Use cache mechanism for Harmony HDC driver
	config := DriverCacheConfig{
		Platform: "harmony",
		Serial:   "", // Let it auto-detect the device serial
		AIOptions: []option.AIServiceOption{
			option.WithCVService(option.CVServiceTypeVEDEM),
			option.WithLLMConfig(
				option.NewLLMServiceConfig(option.DOUBAO_1_5_UI_TARS_250328).
					WithPlannerModel(option.WINGS_SERVICE).
					WithAsserterModel(option.WINGS_SERVICE),
			),
		},
	}

	driverExt, err := GetOrCreateXTDriver(config)
	require.Nil(t, err)
	return driverExt
}

func TestWindowSize(t *testing.T) {
	t.Skip("Skip HarmonyOS test - requires physical HarmonyOS device with HDC")
	driver := setupHDCDriverExt(t)
	size, err := driver.WindowSize()
	assert.Nil(t, err)
	assert.NotNil(t, size)
}

func TestHarmonyTap(t *testing.T) {
	t.Skip("Skip HarmonyOS test - requires physical HarmonyOS device with HDC")
	driver := setupHDCDriverExt(t)
	err := driver.TapAbsXY(200, 2000)
	assert.Nil(t, err)
}

func TestHarmonySwipe(t *testing.T) {
	t.Skip("Skip HarmonyOS test - requires physical HarmonyOS device with HDC")
	driver := setupHDCDriverExt(t)
	err := driver.Swipe(0.5, 0.5, 0.1, 0.5)
	assert.Nil(t, err)
}

func TestHarmonyInput(t *testing.T) {
	t.Skip("Skip HarmonyOS test - requires physical HarmonyOS device with HDC")
	driver := setupHDCDriverExt(t)
	err := driver.Input("test")
	assert.Nil(t, err)
}

func TestHomeScreen(t *testing.T) {
	t.Skip("Skip HarmonyOS test - requires physical HarmonyOS device with HDC")
	driver := setupHDCDriverExt(t)
	err := driver.Home()
	assert.Nil(t, err)
}

func TestUnlock(t *testing.T) {
	t.Skip("Skip HarmonyOS test - requires physical HarmonyOS device with HDC")
	driver := setupHDCDriverExt(t)
	err := driver.Unlock()
	assert.Nil(t, err)
}

func TestPressBack(t *testing.T) {
	t.Skip("Skip HarmonyOS test - requires physical HarmonyOS device with HDC")
	driver := setupHDCDriverExt(t)
	err := driver.Back()
	assert.Nil(t, err)
}

func TestScreenshot(t *testing.T) {
	t.Skip("Skip HarmonyOS test - requires physical HarmonyOS device with HDC")
	driver := setupHDCDriverExt(t)
	screenshot, err := driver.ScreenShot()
	assert.Nil(t, err)
	t.Log(screenshot)
}

func TestLaunch(t *testing.T) {
	t.Skip("Skip HarmonyOS test - requires physical HarmonyOS device with HDC")
	driver := setupHDCDriverExt(t)
	err := driver.AppLaunch("")
	assert.Nil(t, err)
}

func TestForegroundApp(t *testing.T) {
	t.Skip("Skip HarmonyOS test - requires physical HarmonyOS device with HDC")
	driver := setupHDCDriverExt(t)
	appInfo, err := driver.ForegroundInfo()
	assert.Nil(t, err)
	t.Log(appInfo)
}
