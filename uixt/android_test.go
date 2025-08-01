//go:build localtest

package uixt

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/httprunner/httprunner/v5/uixt/types"
)

func setupADBDriverExt(t *testing.T) *XTDriver {
	config := DriverCacheConfig{
		Platform: "android",
		Serial:   "", // Let it auto-detect the device serial
		AIOptions: []option.AIServiceOption{
			option.WithCVService(option.CVServiceTypeVEDEM),
		},
	}

	driverExt, err := GetOrCreateXTDriver(config)
	require.Nil(t, err)
	return driverExt
}

func setupUIA2DriverExt(t *testing.T) *XTDriver {
	// Use cache mechanism with UIA2 enabled
	deviceOpts := option.NewDeviceOptions(
		option.WithPlatform("android"),
		option.WithDeviceUIA2(true),
		option.WithDeviceLogOn(false),
	)

	config := DriverCacheConfig{
		Platform:   "android",
		Serial:     "", // Let it auto-detect the device serial
		DeviceOpts: deviceOpts,
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

func TestDevice_Android_GetPackageInfo(t *testing.T) {
	driver := setupADBDriverExt(t)
	appInfo, err := driver.GetDevice().GetPackageInfo("com.android.settings")
	require.Nil(t, err)
	t.Log(appInfo)
	assert.Equal(t, "com.android.settings", appInfo.Name)
	assert.NotEmpty(t, appInfo.AppPath)
	assert.NotEmpty(t, appInfo.AppMD5)
}

func TestDevice_Android_GetCurrentWindow(t *testing.T) {
	driver := setupADBDriverExt(t)
	driver.AppLaunch("com.android.settings")
	windowInfo, err := driver.GetDevice().(*AndroidDevice).GetCurrentWindow()
	require.Nil(t, err)
	assert.Equal(t, "com.android.settings", windowInfo.PackageName)
}

func TestDriver_ADB_Session_TODO(t *testing.T) {
	driver := setupADBDriverExt(t)
	err := driver.InitSession(nil)
	require.Nil(t, err)
	err = driver.DeleteSession()
	assert.Nil(t, err)
}

func TestDriver_ADB_Status_TODO(t *testing.T) {
	driver := setupADBDriverExt(t)
	status, err := driver.Status()
	require.Nil(t, err)
	t.Log(status)
}

func TestDriver_ADB_ScreenShot(t *testing.T) {
	driver := setupADBDriverExt(t)
	screenshot, err := driver.ScreenShot()
	assert.Nil(t, err)
	path := "1234.jpeg"
	err = saveScreenShot(screenshot, path)
	require.Nil(t, err)
	defer os.Remove(path)
	t.Logf("save screenshot to %s", path)
}

func TestDriver_ADB_Rotation_TODO(t *testing.T) {
	driver := setupADBDriverExt(t)
	rotation, err := driver.Rotation()
	require.Nil(t, err)
	t.Logf("x = %d\ty = %d\tz = %d", rotation.X, rotation.Y, rotation.Z)
}

func TestDriver_ADB_DeviceSize(t *testing.T) {
	driver := setupADBDriverExt(t)
	deviceSize, err := driver.WindowSize()
	require.Nil(t, err)
	assert.Greater(t, deviceSize.Width, 200)
	assert.Greater(t, deviceSize.Height, 200)
}

func TestDriver_ADB_Source(t *testing.T) {
	driver := setupADBDriverExt(t)
	source, err := driver.Source()
	require.Nil(t, err)
	assert.Contains(t, source, "<?xml version")
	assert.Contains(t, source, "android.widget.TextView")
	t.Log(source)
}

func TestDriver_ADB_BatteryInfo_TODO(t *testing.T) {
	driver := setupADBDriverExt(t)
	batteryInfo, err := driver.BatteryInfo()
	require.Nil(t, err)
	t.Log(batteryInfo)
}

func TestDriver_ADB_DeviceInfo_TODO(t *testing.T) {
	driver := setupADBDriverExt(t)
	devInfo, err := driver.DeviceInfo()
	require.Nil(t, err)
	t.Logf("api version: %s", devInfo.APIVersion)
	t.Logf("platform version: %s", devInfo.PlatformVersion)
	t.Logf("bluetooth state: %s", devInfo.Bluetooth.State)
}

func TestDriver_ADB_TapXY(t *testing.T) {
	driver := setupADBDriverExt(t)
	err := driver.TapXY(0.4, 0.5)
	assert.Nil(t, err)
}

func TestDriver_ADB_TapAbsXY(t *testing.T) {
	driver := setupADBDriverExt(t)
	err := driver.TapAbsXY(100, 300)
	assert.Nil(t, err)
}

func TestDriver_ADB_Swipe(t *testing.T) {
	driver := setupADBDriverExt(t)
	err := driver.Swipe(0.5, 0.7, 0.5, 0.5,
		option.WithPressDuration(0.5))
	assert.Nil(t, err)
}

func TestDriver_ADB_Drag(t *testing.T) {
	driver := setupADBDriverExt(t)
	err := driver.Drag(0.5, 0.7, 0.5, 0.5)
	assert.Nil(t, err)
}

func TestDriver_ADB_Input(t *testing.T) {
	driver := setupADBDriverExt(t)
	err := driver.Input("Hi 你好\n",
		option.WithIdentifier("test"))
	assert.Nil(t, err)
	time.Sleep(time.Second * 1)
	err = driver.Input("123\n")
	assert.Nil(t, err)
}

func TestDriver_ADB_PressBack(t *testing.T) {
	driver := setupADBDriverExt(t)
	err := driver.Back()
	assert.Nil(t, err)
}

func TestDriver_ADB_SetRotation_TODO(t *testing.T) {
	driver := setupADBDriverExt(t)
	err := driver.SetRotation(types.Rotation{Z: 270})
	assert.Nil(t, err)
}

func TestDriver_ADB_Orientation(t *testing.T) {
	driver := setupADBDriverExt(t)
	orientation, err := driver.Orientation()
	assert.Nil(t, err)
	assert.Equal(t, types.OrientationPortrait, orientation)
}

func TestDriver_ADB_AppLaunchTerminate(t *testing.T) {
	driver := setupADBDriverExt(t)
	err := driver.AppLaunch("com.android.settings")
	assert.Nil(t, err)
	time.Sleep(1 * time.Second)
	ok, err := driver.AppTerminate("com.android.settings")
	assert.Nil(t, err)
	assert.True(t, ok)
}

func TestDriver_ADB_ForegroundInfo(t *testing.T) {
	driver := setupADBDriverExt(t)
	err := driver.AppLaunch("com.android.settings")
	assert.Nil(t, err)
	app, err := driver.ForegroundInfo()
	assert.Nil(t, err)
	assert.Equal(t, "com.android.settings", app.PackageName)
}

func TestDriver_ADB_ScreenRecord(t *testing.T) {
	driver := setupADBDriverExt(t)

	// adb screenrecord --time-limit 5
	path1, err := driver.ScreenRecord(
		option.WithScreenRecordDuation(5))
	assert.Nil(t, err)
	defer os.Remove(path1)
	t.Log(path1)

	// scrcpy with time limit
	path2, err := driver.ScreenRecord(
		option.WithScreenRecordDuation(5),
		option.WithScreenRecordAudio(true),
	)
	assert.Nil(t, err)
	defer os.Remove(path2)
	t.Log(path2)

	// scrcpy with time limit
	path3, err := driver.ScreenRecord(
		option.WithScreenRecordDuation(5),
		option.WithScreenRecordScrcpy(true),
	)
	assert.Nil(t, err)
	defer os.Remove(path3)
	t.Log(path3)

	// scrcpy with cancel signal
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error)
	go func() {
		path4, err := driver.ScreenRecord(
			option.WithContext(ctx),
			option.WithScreenRecordScrcpy(true),
		)
		assert.Nil(t, err)
		defer os.Remove(path4)
		t.Log(path4)
		done <- err
	}()

	// record for 3 seconds
	time.Sleep(time.Second * 3)
	cancel()

	err = <-done
	assert.Nil(t, err)
}

func TestDriver_ADB_PushImage(t *testing.T) {
	driver := setupADBDriverExt(t)

	screenshot, err := driver.ScreenShot()
	assert.Nil(t, err)
	path := "1234.jpeg"
	err = saveScreenShot(screenshot, path)
	require.Nil(t, err)
	defer os.Remove(path)

	err = driver.PushImage(path)
	assert.Nil(t, err)

	err = driver.PullImages("./test")
	assert.Nil(t, err)

	err = driver.ClearImages()
	assert.Nil(t, err)
}

func TestDriver_ADB_Backspace(t *testing.T) {
	driver := setupADBDriverExt(t)
	err := driver.Backspace(1)
	assert.Nil(t, err)
}

func TestDriver_UIA2_TapXY(t *testing.T) {
	driver := setupUIA2DriverExt(t)
	driver.StartCaptureLog("tap_xy")
	err := driver.TapXY(0.5, 0.5,
		option.WithIdentifier("test"),
		option.WithPressDuration(4))
	assert.Nil(t, err)
	result, _ := driver.StopCaptureLog()
	t.Log(result)
}

func TestDriver_UIA2_Swipe(t *testing.T) {
	driver := setupUIA2DriverExt(t)
	err := driver.Swipe(0.5, 0.7, 0.5, 0.5,
		option.WithPressDuration(0.5))
	assert.Nil(t, err)
}

func TestDriver_UIA2_Input(t *testing.T) {
	driver := setupUIA2DriverExt(t)
	err := driver.Input("Hi 你好\n",
		option.WithIdentifier("test"))
	assert.Nil(t, err)
	time.Sleep(time.Second * 1)
	err = driver.Input("123\n")
	assert.Nil(t, err)
}

func TestDriver_ADB_GetPasteboard(t *testing.T) {
	driver := setupADBDriverExt(t)
	pasteboard, err := driver.IDriver.(*ADBDriver).GetPasteboard()
	assert.Nil(t, err)
	t.Log(pasteboard)
}
