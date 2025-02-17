//go:build localtest

package uixt

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/pkg/uixt/ai"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
	"github.com/httprunner/httprunner/v5/pkg/uixt/types"
)

var (
	driver    IDriver
	driverExt *XTDriver
)

func setupAndroidAdbDriver(t *testing.T) {
	device, err := NewAndroidDevice()
	checkErr(t, err)
	device.Options.UIA2 = false
	device.Options.LogOn = false
	driver, err = device.NewDriver()
	checkErr(t, err)
	driverExt = NewXTDriver(driver,
		ai.WithCVService(ai.CVServiceTypeVEDEM))
}

func setupAndroidUIA2Driver(t *testing.T) {
	device, err := NewAndroidDevice()
	checkErr(t, err)
	device.Options.UIA2 = true
	device.Options.LogOn = false
	driver, err = device.NewDriver()
	checkErr(t, err)
	driverExt = NewXTDriver(driver,
		ai.WithCVService(ai.CVServiceTypeVEDEM))
}

func TestAndroidDevice_GetPackageInfo(t *testing.T) {
	device, err := NewAndroidDevice()
	checkErr(t, err)
	appInfo, err := device.GetPackageInfo("com.android.settings")
	checkErr(t, err)
	t.Log(appInfo)
}

func TestAndroidDevice_GetCurrentWindow(t *testing.T) {
	device, err := NewAndroidDevice()
	checkErr(t, err)
	windowInfo, err := device.GetCurrentWindow()
	checkErr(t, err)
	t.Logf("packageName: %s\tactivityName: %s", windowInfo.PackageName, windowInfo.Activity)
}

func TestDriver_Quit(t *testing.T) {
	if err := driver.DeleteSession(); err != nil {
		t.Fatal(err)
	}
}

func TestDriver_Status(t *testing.T) {
	_, err := driver.Status()
	if err != nil {
		t.Fatal(err)
	}
}

func TestDriver_Screenshot(t *testing.T) {
	screenshot, err := driver.ScreenShot()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(os.WriteFile("/Users/hero/Desktop/s1.png", screenshot.Bytes(), 0o600))
}

func TestDriver_Rotation(t *testing.T) {
	rotation, err := driver.Rotation()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("x = %d\ty = %d\tz = %d", rotation.X, rotation.Y, rotation.Z)
}

func TestDriver_DeviceSize(t *testing.T) {
	deviceSize, err := driver.WindowSize()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("width = %d\theight = %d", deviceSize.Width, deviceSize.Height)
}

func TestDriver_Source(t *testing.T) {
	setupAndroidUIA2Driver(t)

	source, err := driverExt.Source()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(source)
}

func TestDriver_BatteryInfo(t *testing.T) {
	batteryInfo, err := driver.BatteryInfo()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(batteryInfo)
}

func TestDriver_DeviceInfo(t *testing.T) {
	devInfo, err := driver.DeviceInfo()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("api version: %s", devInfo.APIVersion)
	t.Logf("platform version: %s", devInfo.PlatformVersion)
	t.Logf("bluetooth state: %s", devInfo.Bluetooth.State)
}

func TestDriver_Tap(t *testing.T) {
	setupAndroidUIA2Driver(t)
	driverExt.StartCaptureLog("")
	err := driverExt.TapXY(0.5, 0.5,
		option.WithIdentifier("test"),
		option.WithPressDuration(4))
	if err != nil {
		t.Fatal(err)
	}
	//time.Sleep(time.Second)
	//
	//err = driverExt.Tap(60.5, 125.5, WithIdentifier("test"))
	//if err != nil {
	//	t.Fatal(err)
	//}
	//time.Sleep(time.Second)
	//result, _ := driverExt.StopCaptureLog()
	//t.Log(result)
}

func TestDriver_Swipe(t *testing.T) {
	setupAndroidUIA2Driver(t)
	err := driverExt.Swipe(400, 1000, 400, 500,
		option.WithPressDuration(0.5))
	if err != nil {
		t.Fatal(err)
	}
}

func TestDriver_Swipe_Relative(t *testing.T) {
	setupAndroidUIA2Driver(t)
	err := driverExt.Swipe(0.5, 0.7, 0.5, 0.5)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDriver_Drag(t *testing.T) {
	err := driver.Drag(400, 260, 400, 500)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Millisecond * 200)

	err = driver.Drag(400, 501.5, 400, 261.5)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Millisecond * 200)
}

func TestDriver_SendKeys(t *testing.T) {
	setupAndroidUIA2Driver(t)

	err := driverExt.Input("辽宁省沈阳市新民市民族街36-4",
		option.WithIdentifier("test"))
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Second * 2)

	//err = driver.SendKeys("def")
	//if err != nil {
	//	t.Fatal(err)
	//}
	//time.Sleep(time.Second * 2)

	//err = driver.SendKeys("\\n")
	// err = driver.SendKeys(`\n`, false)
	//if err != nil {
	//	t.Fatal(err)
	//}
}

func TestDriver_PressBack(t *testing.T) {
	err := driver.Back()
	if err != nil {
		t.Fatal(err)
	}
}

func TestDriver_SetRotation(t *testing.T) {
	// err = driver.SetRotation(Rotation{Z: 0})
	err := driver.SetRotation(types.Rotation{Z: 270})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDriver_GetOrientation(t *testing.T) {
	setupAndroidUIA2Driver(t)
	_, _ = driverExt.AppTerminate("com.quark.browser")
	_ = driverExt.AppLaunch("com.quark.browser")
	time.Sleep(2 * time.Second)
	_ = driverExt.Home()
}

func Test_getFreePort(t *testing.T) {
	freePort, err := builtin.GetFreePort()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(freePort)
}

func TestDriver_AppLaunch(t *testing.T) {
	device, _ := NewAndroidDevice()
	driver, err := device.NewDriver()
	if err != nil {
		t.Fatal(err)
	}

	err = driver.AppLaunch("com.android.settings")
	if err != nil {
		t.Fatal(err)
	}

	raw, err := driver.ScreenShot()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(os.WriteFile("s1.png", raw.Bytes(), 0o600))
}

func TestDriver_IsAppInForeground(t *testing.T) {
	setupAndroidUIA2Driver(t)
	// setupAndroidAdbDriver(t)

	err := driverExt.AppLaunch("com.android.settings")
	checkErr(t, err)

	app, err := driverExt.ForegroundInfo()
	checkErr(t, err)
	if app.PackageName != "com.android.settings" {
		t.FailNow()
	}
	if app.Activity != ".Settings" {
		t.FailNow()
	}
}

func TestDriver_KeepAlive(t *testing.T) {
	device, _ := NewAndroidDevice()
	driver, err := device.NewDriver()
	if err != nil {
		t.Fatal(err)
	}

	err = driver.AppLaunch("com.android.settings")
	if err != nil {
		t.Fatal(err)
	}

	_, err = driver.ScreenShot()
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(60 * time.Second)

	_, err = driver.ScreenShot()
	if err != nil {
		t.Fatal(err)
	}
}

func TestDriver_AppTerminate(t *testing.T) {
	device, _ := NewAndroidDevice()
	driver, err := device.NewDriver()
	if err != nil {
		t.Fatal(err)
	}

	_, err = driver.AppTerminate("tv.danmaku.bili")
	if err != nil {
		t.Fatal(err)
	}
}

func TestConvertPoints(t *testing.T) {
	data := "10-09 20:16:48.216 I/iesqaMonitor(17845): {\"duration\":0,\"end\":1665317808206,\"ext\":\"输入\",\"from\":{\"x\":0.0,\"y\":0.0},\"operation\":\"Gtf-SendKeys\",\"run_time\":627,\"start\":1665317807579,\"start_first\":0,\"start_last\":0,\"to\":{\"x\":0.0,\"y\":0.0}}\n10-09 20:18:22.899 I/iesqaMonitor(17845): {\"duration\":0,\"end\":1665317902898,\"ext\":\"进入直播间\",\"from\":{\"x\":717.0,\"y\":2117.5},\"operation\":\"Gtf-Tap\",\"run_time\":121,\"start\":1665317902777,\"start_first\":0,\"start_last\":0,\"to\":{\"x\":717.0,\"y\":2117.5}}\n10-09 20:18:32.063 I/iesqaMonitor(17845): {\"duration\":0,\"end\":1665317912062,\"ext\":\"第一次上划\",\"from\":{\"x\":1437.0,\"y\":2409.9},\"operation\":\"Gtf-Swipe\",\"run_time\":32,\"start\":1665317912030,\"start_first\":0,\"start_last\":0,\"to\":{\"x\":1437.0,\"y\":2409.9}}"

	eps := ConvertPoints(strings.Split(data, "\n"))
	if len(eps) != 3 {
		t.Fatal()
	}
}

func TestDriver_ShellInputUnicode(t *testing.T) {
	device, _ := NewAndroidDevice()
	driver, err := NewADBDriver(device)
	if err != nil {
		t.Fatal(err)
	}

	err = driver.Input("test中文输入&")
	if err != nil {
		t.Fatal(err)
	}

	raw, err := driver.ScreenShot()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(os.WriteFile("s1.png", raw.Bytes(), 0o600))
}

func TestRecordVideo(t *testing.T) {
	setupAndroidAdbDriver(t)
	path, err := driverExt.ScreenRecord(5 * time.Second)
	if err != nil {
		t.Fatal(err)
	}
	println(path)
}

func Test_Android_Backspace(t *testing.T) {
	setupAndroidAdbDriver(t)

	err := driverExt.Backspace(1)
	if err != nil {
		t.Fatal(err)
	}
}
