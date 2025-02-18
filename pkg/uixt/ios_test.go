//go:build localtest

package uixt

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/httprunner/httprunner/v5/pkg/uixt/ai"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
	"github.com/httprunner/httprunner/v5/pkg/uixt/types"
)

func setupWDADriverExt(t *testing.T) *XTDriver {
	device, err := NewIOSDevice(
		option.WithWDAPort(8700),
		option.WithWDAMjpegPort(8800),
		option.WithWDALogOn(true))
	require.Nil(t, err)
	driver, err := device.NewDriver()
	require.Nil(t, err)
	return NewXTDriver(driver, ai.WithCVService(ai.CVServiceTypeVEDEM))
}

func TestDevice_IOS_Install(t *testing.T) {
	driver := setupWDADriverExt(t)
	err := driver.GetDevice().Install("xxx.ipa",
		option.WithRetryTimes(5))
	assert.Nil(t, err)
}

func TestDevice_IOS_New(t *testing.T) {
	device, err := NewIOSDevice(
		option.WithWDAPort(8700),
		option.WithWDAMjpegPort(8800))
	require.Nil(t, err)

	device, _ = NewIOSDevice(option.WithUDID("xxxx"))
	if device != nil {
		t.Log(device)
	}

	device, _ = NewIOSDevice(
		option.WithWDAPort(8700),
		option.WithWDAMjpegPort(8800))
	if device != nil {
		t.Log(device)
	}

	device, _ = NewIOSDevice(
		option.WithUDID("xxxx"),
		option.WithWDAPort(8700),
		option.WithWDAMjpegPort(8800))
	if device != nil {
		t.Log(device)
	}
}

func TestDevice_IOS_GetPackageInfo(t *testing.T) {
	device, err := NewIOSDevice(option.WithWDAPort(8700))
	require.Nil(t, err)
	appInfo, err := device.GetPackageInfo("com.ss.iphone.ugc.Aweme")
	assert.Nil(t, err)
	t.Logf("%+v", appInfo)
}

func TestDriver_WDA_DeviceScaleRatio(t *testing.T) {
	driver := setupWDADriverExt(t)
	scaleRatio, err := driver.IDriver.(*WDADriver).Scale()
	require.Nil(t, err)
	t.Logf("%+v", scaleRatio)
}

func TestDriver_WDA_DeleteSession(t *testing.T) {
	driver := setupWDADriverExt(t)
	err := driver.DeleteSession()
	assert.Nil(t, err)
}

func TestDriver_WDA_HealthCheck(t *testing.T) {
	driver := setupWDADriverExt(t)
	err := driver.IDriver.(*WDADriver).HealthCheck()
	assert.Nil(t, err)
}

func TestDriver_WDA_GetAppiumSettings(t *testing.T) {
	driver := setupWDADriverExt(t)
	settings, err := driver.IDriver.(*WDADriver).GetAppiumSettings()
	assert.Nil(t, err)
	t.Logf("%+v", settings)
}

func TestDriver_WDA_SetAppiumSettings(t *testing.T) {
	driver := setupWDADriverExt(t)

	const _acceptAlertButtonSelector = "**/XCUIElementTypeButton[`label IN {'允许','好','仅在使用应用期间','暂不'}`]"
	const _dismissAlertButtonSelector = "**/XCUIElementTypeButton[`label IN {'不允许','暂不'}`]"

	key := "acceptAlertButtonSelector"
	value := _acceptAlertButtonSelector

	// settings, err := driver.SetAppiumSettings(map[string]interface{}{"dismissAlertButtonSelector": "暂不"})
	settings, err := driver.IDriver.(*WDADriver).SetAppiumSettings(map[string]interface{}{key: value})
	assert.Nil(t, err)
	assert.Equal(t, settings[key], value)
}

func TestDriver_WDA_IsWdaHealthy(t *testing.T) {
	driver := setupWDADriverExt(t)
	healthy, err := driver.IDriver.(*WDADriver).IsHealthy()
	assert.Nil(t, err)
	assert.True(t, healthy)
}

func TestDriver_WDA_Status(t *testing.T) {
	driver := setupWDADriverExt(t)
	status, err := driver.Status()
	assert.Nil(t, err)
	assert.True(t, status.Ready)
}

func TestDriver_WDA_DeviceInfo(t *testing.T) {
	driver := setupWDADriverExt(t)
	info, err := driver.DeviceInfo()
	assert.Nil(t, err)
	assert.NotEmpty(t, info.Model)
}

func TestDriver_WDA_BatteryInfo(t *testing.T) {
	driver := setupWDADriverExt(t)
	batteryInfo, err := driver.BatteryInfo()
	assert.Nil(t, err)
	t.Log(batteryInfo)
}

func TestDriver_WDA_WindowSize(t *testing.T) {
	driver := setupWDADriverExt(t)
	size, err := driver.WindowSize()
	assert.Nil(t, err)
	t.Log(size)
}

func TestDriver_WDA_Screen(t *testing.T) {
	driver := setupWDADriverExt(t)
	screen, err := driver.IDriver.(*WDADriver).Screen()
	assert.Nil(t, err)
	t.Log(screen)
}

func TestDriver_WDA_Home(t *testing.T) {
	driver := setupWDADriverExt(t)
	err := driver.Home()
	assert.Nil(t, err)
}

func TestDriver_WDA_AppLaunchTerminate(t *testing.T) {
	driver := setupWDADriverExt(t)

	bundleId := "com.apple.Preferences"
	err := driver.AppLaunch(bundleId)
	assert.Nil(t, err)
	time.Sleep(2 * time.Second)

	_, err = driver.AppTerminate(bundleId)
	assert.Nil(t, err)
}

func TestDriver_WDA_TapXY(t *testing.T) {
	driver := setupWDADriverExt(t)

	err := driver.TapXY(0.2, 0.2)
	assert.Nil(t, err)
}

func TestDriver_WDA_DoubleTapXY(t *testing.T) {
	driver := setupWDADriverExt(t)

	err := driver.DoubleTapXY(0.2, 0.2)
	assert.Nil(t, err)
}

func TestDriver_WDA_TouchAndHold(t *testing.T) {
	driver := setupWDADriverExt(t)

	err := driver.TouchAndHold(0.2, 0.2)
	assert.Nil(t, err)
}

func TestDriver_WDA_Drag(t *testing.T) {
	driver := setupWDADriverExt(t)

	err := driver.Drag(0.8, 0.5, 0.2, 0.5,
		option.WithDuration(0.5))
	assert.Nil(t, err)
}

func TestDriver_WDA_Swipe(t *testing.T) {
	driver := setupWDADriverExt(t)

	err := driver.Swipe(0.8, 0.5, 0.2, 0.5)
	assert.Nil(t, err)
}

func TestDriver_WDA_Input(t *testing.T) {
	driver := setupWDADriverExt(t)
	driver.StartCaptureLog("hrp_wda_log")
	err := driver.Input("test中文", option.WithIdentifier("test"))
	assert.Nil(t, err)
	result, err := driver.StopCaptureLog()
	assert.Nil(t, err)
	t.Log(result)
}

func TestDriver_WDA_PressButton(t *testing.T) {
	driver := setupWDADriverExt(t)

	err := driver.IDriver.(*WDADriver).PressButton(types.DeviceButtonVolumeUp)
	assert.Nil(t, err)
	time.Sleep(time.Second * 1)
	err = driver.IDriver.(*WDADriver).PressButton(types.DeviceButtonVolumeDown)
	assert.Nil(t, err)
	time.Sleep(time.Second * 1)
	err = driver.IDriver.(*WDADriver).PressButton(types.DeviceButtonHome)
	assert.Nil(t, err)
}

func TestDriver_WDA_ScreenShot(t *testing.T) {
	driver := setupWDADriverExt(t)

	// without save file
	screenshot, err := driver.ScreenShot()
	assert.Nil(t, err)
	_ = screenshot

	// save file
	screenshot, err = driver.ScreenShot(option.WithScreenShotFileName("123"))
	assert.Nil(t, err)
	_ = screenshot

	path, err := saveScreenShot(screenshot, "1234")
	assert.Nil(t, err)
	defer os.Remove(path)
	t.Logf("save screenshot to %s", path)
}

func TestDriver_WDA_Source(t *testing.T) {
	driver := setupWDADriverExt(t)

	var source string
	var err error

	source, err = driver.Source()
	assert.Nil(t, err)

	source, err = driver.Source(option.WithFormat(option.SourceFormatJSON))
	assert.Nil(t, err)

	source, err = driver.Source(option.WithFormat(option.SourceFormatDescription))
	assert.Nil(t, err)

	source, err = driver.Source(
		option.WithFormat(option.SourceFormatXML),
		option.WithExcludedAttributes([]string{"label", "type", "index"}))
	assert.Nil(t, err)
	t.Logf("source: %s", source)
}

func TestDriver_WDA_GetForegroundApp(t *testing.T) {
	driver := setupWDADriverExt(t)
	app, err := driver.ForegroundInfo()
	assert.Nil(t, err)
	t.Log(app)
}

func TestDriver_WDA_AccessibleSource(t *testing.T) {
	driver := setupWDADriverExt(t)
	source, err := driver.IDriver.(*WDADriver).AccessibleSource()
	assert.Nil(t, err)
	t.Log(source)
}

func TestDriver_WDA_ScreenRecord(t *testing.T) {
	driver := setupWDADriverExt(t)
	path, err := driver.ScreenRecord(5 * time.Second)
	assert.Nil(t, err)
	t.Log(path)
}

func TestDriver_WDA_Backspace(t *testing.T) {
	driver := setupWDADriverExt(t)
	err := driver.Backspace(3)
	assert.Nil(t, err)
}
