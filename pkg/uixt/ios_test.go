//go:build localtest

package uixt

import (
	"fmt"
	"testing"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/pkg/uixt/ai"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
	"github.com/httprunner/httprunner/v5/pkg/uixt/types"
)

func setupWDADriverExt(t *testing.T) *XTDriver {
	device, err := NewIOSDevice(
		option.WithWDAPort(8700),
		option.WithWDAMjpegPort(8800),
		option.WithWDALogOn(true))
	if err != nil {
		t.Fatal(err)
	}
	driver, err := device.NewDriver()
	if err != nil {
		t.Fatal(err)
	}
	return NewXTDriver(driver, ai.WithCVService(ai.CVServiceTypeVEDEM))
}

func TestInstall(t *testing.T) {
	driver := setupWDADriverExt(t)
	err := driver.GetDevice().Install("xxx.ipa",
		option.WithRetryTimes(5))
	log.Error().Err(err)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNewIOSDevice(t *testing.T) {
	device, _ := NewIOSDevice(
		option.WithWDAPort(8700),
		option.WithWDAMjpegPort(8800))
	if device != nil {
		t.Log(device)
	}

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

func TestIOSDevice_GetPackageInfo(t *testing.T) {
	device, err := NewIOSDevice(option.WithWDAPort(8700))
	checkErr(t, err)
	appInfo, err := device.GetPackageInfo("com.ss.iphone.ugc.Aweme")
	checkErr(t, err)
	t.Logf("%+v", appInfo)
}

func TestDriver_WDA_DeviceScaleRatio(t *testing.T) {
	driver := setupWDADriverExt(t)

	scaleRatio, err := driver.IDriver.(*WDADriver).Scale()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%+v", scaleRatio)
}

func TestDriver_WDA_DeleteSession(t *testing.T) {
	driver := setupWDADriverExt(t)

	err := driver.DeleteSession()
	if err != nil {
		t.Fatal(err)
	}
}

func TestDriver_WDA_HealthCheck(t *testing.T) {
	driver := setupWDADriverExt(t)

	err := driver.IDriver.(*WDADriver).HealthCheck()
	if err != nil {
		t.Fatal(err)
	}
}

func TestDriver_WDA_GetAppiumSettings(t *testing.T) {
	driver := setupWDADriverExt(t)

	settings, err := driver.IDriver.(*WDADriver).GetAppiumSettings()
	if err != nil {
		t.Fatal(err)
	}
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
	if err != nil {
		t.Fatal(err)
	}
	if settings[key] != value {
		t.Fatal(settings[key])
	}
}

func TestDriver_WDA_IsWdaHealthy(t *testing.T) {
	driver := setupWDADriverExt(t)

	healthy, err := driver.IDriver.(*WDADriver).IsHealthy()
	if err != nil {
		t.Fatal(err)
	}
	if !healthy {
		t.Fatal("assert healthy failed")
	}
}

func TestDriver_WDA_Status(t *testing.T) {
	driver := setupWDADriverExt(t)

	status, err := driver.Status()
	if err != nil {
		t.Fatal(err)
	}
	if !status.Ready {
		t.Fatal("assert device status failed")
	}
}

func TestDriver_WDA_DeviceInfo(t *testing.T) {
	driver := setupWDADriverExt(t)

	info, err := driver.DeviceInfo()
	if err != nil {
		t.Fatal(err)
	}
	if len(info.Model) == 0 {
		t.Fatal(info)
	}
}

func TestDriver_WDA_BatteryInfo(t *testing.T) {
	driver := setupWDADriverExt(t)

	batteryInfo, err := driver.BatteryInfo()
	if err != nil {
		t.Fatal()
	}
	t.Log(batteryInfo)
}

func TestDriver_WDA_WindowSize(t *testing.T) {
	driver := setupWDADriverExt(t)

	size, err := driver.WindowSize()
	if err != nil {
		t.Fatal()
	}
	t.Log(size)
}

func TestDriver_WDA_Screen(t *testing.T) {
	driver := setupWDADriverExt(t)

	screen, err := driver.IDriver.(*WDADriver).Screen()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(screen)
}

func TestDriver_WDA_Home(t *testing.T) {
	driver := setupWDADriverExt(t)

	err := driver.Home()
	if err != nil {
		t.Fatal(err)
	}
}

func TestDriver_WDA_AppLaunchTerminate(t *testing.T) {
	driver := setupWDADriverExt(t)

	bundleId := "com.apple.Preferences"
	err := driver.AppLaunch(bundleId)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(2 * time.Second)

	_, err = driver.AppTerminate(bundleId)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDriver_WDA_TapXY(t *testing.T) {
	driver := setupWDADriverExt(t)

	err := driver.TapXY(0.2, 0.2)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDriver_WDA_DoubleTapXY(t *testing.T) {
	driver := setupWDADriverExt(t)

	err := driver.DoubleTapXY(0.2, 0.2)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDriver_WDA_TouchAndHold(t *testing.T) {
	driver := setupWDADriverExt(t)

	err := driver.TouchAndHold(0.2, 0.2)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDriver_WDA_Drag(t *testing.T) {
	driver := setupWDADriverExt(t)

	err := driver.Drag(0.8, 0.5, 0.2, 0.5,
		option.WithDuration(0.5))
	if err != nil {
		t.Fatal(err)
	}
}

func TestDriver_WDA_Swipe(t *testing.T) {
	driver := setupWDADriverExt(t)

	err := driver.Swipe(0.8, 0.5, 0.2, 0.5)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDriver_WDA_Input(t *testing.T) {
	driver := setupWDADriverExt(t)
	driver.StartCaptureLog("hrp_wda_log")
	err := driver.Input("test中文", option.WithIdentifier("test"))
	result, _ := driver.StopCaptureLog()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(result)
}

func TestDriver_WDA_PressButton(t *testing.T) {
	driver := setupWDADriverExt(t)

	err := driver.IDriver.(*WDADriver).PressButton(types.DeviceButtonVolumeUp)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second * 1)
	err = driver.IDriver.(*WDADriver).PressButton(types.DeviceButtonVolumeDown)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second * 1)
	err = driver.IDriver.(*WDADriver).PressButton(types.DeviceButtonHome)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDriver_WDA_ScreenShot(t *testing.T) {
	driver := setupWDADriverExt(t)

	// without save file
	screenshot, err := driver.ScreenShot()
	if err != nil {
		t.Fatal(err)
	}
	_ = screenshot

	// save file
	screenshot, err = driver.ScreenShot(option.WithScreenShotFileName("123"))
	if err != nil {
		t.Fatal(err)
	}
	_ = screenshot

	path, err := saveScreenShot(screenshot, "1234")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("save screenshot to %s", path)
}

func TestDriver_WDA_Source(t *testing.T) {
	driver := setupWDADriverExt(t)

	var source string
	var err error

	source, err = driver.Source()
	if err != nil {
		t.Fatal(err)
	}

	source, err = driver.Source(option.WithFormat(option.SourceFormatJSON))
	if err != nil {
		t.Fatal(err)
	}

	source, err = driver.Source(option.WithFormat(option.SourceFormatDescription))
	if err != nil {
		t.Fatal(err)
	}

	source, err = driver.Source(
		option.WithFormat(option.SourceFormatXML),
		option.WithExcludedAttributes([]string{"label", "type", "index"}))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("source: %s", source)
}

func TestDriver_WDA_GetForegroundApp(t *testing.T) {
	driver := setupWDADriverExt(t)
	app, err := driver.ForegroundInfo()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(app)
}

func TestDriver_WDA_AccessibleSource(t *testing.T) {
	driver := setupWDADriverExt(t)

	source, err := driver.IDriver.(*WDADriver).AccessibleSource()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(source)
}

func TestDriver_WDA_ScreenRecord(t *testing.T) {
	driver := setupWDADriverExt(t)
	path, err := driver.ScreenRecord(5 * time.Second)
	if err != nil {
		t.Fatal(err)
	}
	println(path)
}

func TestDriver_WDA_Backspace(t *testing.T) {
	driver := setupWDADriverExt(t)

	err := driver.Backspace(3)
	if err != nil {
		t.Fatal(err)
	}
}
