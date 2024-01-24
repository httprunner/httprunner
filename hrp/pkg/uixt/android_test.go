//go:build localtest

package uixt

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"
)

var (
	uiaServerURL = "http://localhost:6790/wd/hub"
	driverExt    *DriverExt
)

func setupAndroid(t *testing.T) {
	device, err := NewAndroidDevice()
	checkErr(t, err)
	driverExt, err = device.NewDriver()
	checkErr(t, err)
}

func TestDriver_NewSession(t *testing.T) {
	driver, err := NewUIADriver(nil, uiaServerURL)
	if err != nil {
		t.Fatal(err)
	}

	firstMatchEntry := make(map[string]interface{})
	firstMatchEntry["package"] = "com.android.settings"
	firstMatchEntry["activity"] = "com.android.settings/.Settings"
	caps := Capabilities{
		"firstMatch":  []interface{}{firstMatchEntry},
		"alwaysMatch": struct{}{},
	}
	session, err := driver.NewSession(caps)
	if err != nil {
		t.Fatal(err)
	}
	if len(session.SessionId) == 0 {
		t.Fatal("should not be empty")
	}
}

func TestNewDriver(t *testing.T) {
	driver, err := NewUIADriver(nil, uiaServerURL)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(driver.sessionId)
}

func TestDriver_Quit(t *testing.T) {
	driver, err := NewUIADriver(nil, uiaServerURL)
	if err != nil {
		t.Fatal(err)
	}

	if err = driver.DeleteSession(); err != nil {
		t.Fatal(err)
	}
}

func TestDriver_Status(t *testing.T) {
	driver, err := NewUIADriver(nil, uiaServerURL)
	if err != nil {
		t.Fatal(err)
	}

	_, err = driver.Status()
	if err != nil {
		t.Fatal(err)
	}
}

func TestDriver_Screenshot(t *testing.T) {
	driver, err := NewUIADriver(nil, uiaServerURL)
	if err != nil {
		t.Fatal(err)
	}

	screenshot, err := driver.Screenshot()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(os.WriteFile("/Users/hero/Desktop/s1.png", screenshot.Bytes(), 0o600))
}

func TestDriver_Rotation(t *testing.T) {
	driver, err := NewUIADriver(nil, uiaServerURL)
	if err != nil {
		t.Fatal(err)
	}

	rotation, err := driver.Rotation()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("x = %d\ty = %d\tz = %d", rotation.X, rotation.Y, rotation.Z)
}

func TestDriver_DeviceSize(t *testing.T) {
	driver, err := NewUIADriver(nil, uiaServerURL)
	if err != nil {
		t.Fatal(err)
	}

	deviceSize, err := driver.WindowSize()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("width = %d\theight = %d", deviceSize.Width, deviceSize.Height)
}

func TestDriver_Source(t *testing.T) {
	driver, err := NewUIADriver(nil, uiaServerURL)
	if err != nil {
		t.Fatal(err)
	}

	source, err := driver.Source()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(source)
}

func TestDriver_BatteryInfo(t *testing.T) {
	driver, err := NewUIADriver(nil, uiaServerURL)
	if err != nil {
		t.Fatal(err)
	}

	batteryInfo, err := driver.BatteryInfo()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(batteryInfo)
}

func TestDriver_GetAppiumSettings(t *testing.T) {
	driver, err := NewUIADriver(nil, uiaServerURL)
	if err != nil {
		t.Fatal(err)
	}

	appiumSettings, err := driver.GetAppiumSettings()
	if err != nil {
		t.Fatal(err)
	}

	for k := range appiumSettings {
		t.Logf("key: %s\tvalue: %v", k, appiumSettings[k])
	}
	// t.Log(appiumSettings)
}

func TestDriver_DeviceInfo(t *testing.T) {
	driver, err := NewUIADriver(nil, uiaServerURL)
	if err != nil {
		t.Fatal(err)
	}

	devInfo, err := driver.DeviceInfo()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("api version: %s", devInfo.APIVersion)
	t.Logf("platform version: %s", devInfo.PlatformVersion)
	t.Logf("bluetooth state: %s", devInfo.Bluetooth.State)
}

func TestDriver_Tap(t *testing.T) {
	driver, err := NewUIADriver(nil, uiaServerURL)
	if err != nil {
		t.Fatal(err)
	}

	err = driver.Tap(150, 340)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second)

	err = driver.TapFloat(60.5, 125.5)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second)
}

func TestDriver_Swipe(t *testing.T) {
	driver, err := NewUIADriver(nil, uiaServerURL)
	if err != nil {
		t.Fatal(err)
	}

	err = driver.Swipe(400, 1000, 400, 500)
	if err != nil {
		t.Fatal(err)
	}

	err = driver.SwipeFloat(400, 555.5, 400, 1255.5)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDriver_Drag(t *testing.T) {
	driver, err := NewUIADriver(nil, uiaServerURL)
	if err != nil {
		t.Fatal(err)
	}

	err = driver.Drag(400, 260, 400, 500)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Millisecond * 200)

	err = driver.DragFloat(400, 501.5, 400, 261.5)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Millisecond * 200)
}

func TestDriver_SendKeys(t *testing.T) {
	driver, err := NewUIADriver(nil, uiaServerURL)
	if err != nil {
		t.Fatal(err)
	}

	err = driver.SendKeys("abc")
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second * 2)

	err = driver.SendKeys("def")
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second * 2)

	err = driver.SendKeys("\\n")
	// err = driver.SendKeys(`\n`, false)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDriver_PressBack(t *testing.T) {
	driver, err := NewUIADriver(nil, uiaServerURL)
	if err != nil {
		t.Fatal(err)
	}

	err = driver.PressBack()
	if err != nil {
		t.Fatal(err)
	}
}

func TestDriver_SetRotation(t *testing.T) {
	driver, err := NewUIADriver(nil, uiaServerURL)
	if err != nil {
		t.Fatal(err)
	}

	// err = driver.SetRotation(Rotation{Z: 0})
	err = driver.SetRotation(Rotation{Z: 270})
	if err != nil {
		t.Fatal(err)
	}
}

func TestUiSelectorHelper_NewUiSelectorHelper(t *testing.T) {
	uiSelector := NewUiSelectorHelper().Text("a").String()
	if uiSelector != `new UiSelector().text("a");` {
		t.Fatal("[ERROR]", uiSelector)
	}

	uiSelector = NewUiSelectorHelper().Text("a").TextStartsWith("b").String()
	if uiSelector != `new UiSelector().text("a").textStartsWith("b");` {
		t.Fatal("[ERROR]", uiSelector)
	}

	uiSelector = NewUiSelectorHelper().ClassName("android.widget.LinearLayout").Index(6).String()
	if uiSelector != `new UiSelector().className("android.widget.LinearLayout").index(6);` {
		t.Fatal("[ERROR]", uiSelector)
	}

	uiSelector = NewUiSelectorHelper().Focused(false).Instance(6).String()
	if uiSelector != `new UiSelector().focused(false).instance(6);` {
		t.Fatal("[ERROR]", uiSelector)
	}

	uiSelector = NewUiSelectorHelper().ChildSelector(NewUiSelectorHelper().Enabled(true)).String()
	if uiSelector != `new UiSelector().childSelector(new UiSelector().enabled(true));` {
		t.Fatal("[ERROR]", uiSelector)
	}
}

func Test_getFreePort(t *testing.T) {
	freePort, err := getFreePort()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(freePort)
}

func TestDeviceList(t *testing.T) {
	devices, err := GetAndroidDevices()
	if err != nil {
		t.Fatal(err)
	}
	for i := range devices {
		t.Log(devices[i].Serial())
	}
}

func TestDriver_AppLaunch(t *testing.T) {
	device, _ := NewAndroidDevice()
	driver, err := device.NewDriver()
	if err != nil {
		t.Fatal(err)
	}

	err = driver.Driver.AppLaunch("com.android.settings")
	if err != nil {
		t.Fatal(err)
	}

	raw, err := driver.Driver.Screenshot()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(os.WriteFile("s1.png", raw.Bytes(), 0o600))
}

func TestDriver_IsAppInForeground(t *testing.T) {
	setupAndroid(t)

	err := driverExt.Driver.AppLaunch("com.android.settings")
	checkErr(t, err)

	app, err := driverExt.Driver.GetForegroundApp()
	checkErr(t, err)
	if app.PackageName != "com.android.settings" {
		t.FailNow()
	}
	if app.Activity != ".Settings" {
		t.FailNow()
	}

	err = driverExt.Driver.AssertForegroundApp("com.android.settings")
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(2 * time.Second)
	_, err = driverExt.Driver.AppTerminate("com.android.settings")
	if err != nil {
		t.Fatal(err)
	}

	err = driverExt.Driver.AssertForegroundApp("com.android.settings")
	if err == nil {
		t.Fatal(err)
	}
}

func TestDriver_KeepAlive(t *testing.T) {
	device, _ := NewAndroidDevice()
	driver, err := device.NewDriver()
	if err != nil {
		t.Fatal(err)
	}

	err = driver.Driver.AppLaunch("com.android.settings")
	if err != nil {
		t.Fatal(err)
	}

	_, err = driver.Driver.Screenshot()
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(60 * time.Second)

	_, err = driver.Driver.Screenshot()
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

	_, err = driver.Driver.AppTerminate("tv.danmaku.bili")
	if err != nil {
		t.Fatal(err)
	}
}

func TestConvertPoints(t *testing.T) {
	data := "10-09 20:16:48.216 I/iesqaMonitor(17845): {\"duration\":0,\"end\":1665317808206,\"ext\":\"输入\",\"from\":{\"x\":0.0,\"y\":0.0},\"operation\":\"Gtf-SendKeys\",\"run_time\":627,\"start\":1665317807579,\"start_first\":0,\"start_last\":0,\"to\":{\"x\":0.0,\"y\":0.0}}\n10-09 20:18:22.899 I/iesqaMonitor(17845): {\"duration\":0,\"end\":1665317902898,\"ext\":\"进入直播间\",\"from\":{\"x\":717.0,\"y\":2117.5},\"operation\":\"Gtf-Tap\",\"run_time\":121,\"start\":1665317902777,\"start_first\":0,\"start_last\":0,\"to\":{\"x\":717.0,\"y\":2117.5}}\n10-09 20:18:32.063 I/iesqaMonitor(17845): {\"duration\":0,\"end\":1665317912062,\"ext\":\"第一次上划\",\"from\":{\"x\":1437.0,\"y\":2409.9},\"operation\":\"Gtf-Swipe\",\"run_time\":32,\"start\":1665317912030,\"start_first\":0,\"start_last\":0,\"to\":{\"x\":1437.0,\"y\":2409.9}}"
	eps := ConvertPoints(data)
	if len(eps) != 3 {
		t.Fatal()
	}
	jsons, _ := json.Marshal(eps)
	println(fmt.Sprintf("%v", string(jsons)))
}
