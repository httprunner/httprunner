//go:build localtest

package uixt

import (
	"io/ioutil"
	"testing"
	"time"
)

var uiaServerURL = "http://localhost:6790/wd/hub"

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

	if err = driver.Close(); err != nil {
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

func TestDriver_SessionIDs(t *testing.T) {
	driver, err := NewUIADriver(nil, uiaServerURL)
	if err != nil {
		t.Fatal(err)
	}

	sessions, err := driver.SessionIDs()
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) == 0 {
		t.Fatal("should have at least one")
	}
	t.Log(len(sessions), sessions)
}

func TestDriver_SessionDetails(t *testing.T) {
	// firstMatchEntry := make(map[string]interface{})
	// firstMatchEntry["package"] = "com.android.settings"
	// firstMatchEntry["activity"] = "com.android.settings/.Settings"
	// caps = Capabilities{
	// 	"firstMatch":  []interface{}{firstMatchEntry},
	// 	"alwaysMatch": struct{}{},
	// }
	driver, err := NewUIADriver(nil, uiaServerURL)
	if err != nil {
		t.Fatal(err)
	}

	scrollData, err := driver.SessionDetails()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(scrollData)
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

	t.Log(ioutil.WriteFile("/Users/hero/Desktop/s1.png", screenshot.Bytes(), 0o600))
}

func TestDriver_Orientation(t *testing.T) {
	driver, err := NewUIADriver(nil, uiaServerURL)
	if err != nil {
		t.Fatal(err)
	}

	orientation, err := driver.Orientation()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(orientation)
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

func TestDriver_DeviceScaleRatio(t *testing.T) {
	driver, err := NewUIADriver(nil, uiaServerURL)
	if err != nil {
		t.Fatal(err)
	}

	scaleRatio, err := driver.Scale()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(scaleRatio)
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

func TestDriver_AlertText(t *testing.T) {
	driver, err := NewUIADriver(nil, uiaServerURL)
	if err != nil {
		t.Fatal(err)
	}

	alertText, err := driver.AlertText()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(alertText)
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

//func TestDriver_PressBack(t *testing.T) {
//	driver, err := NewUIADriver(nil, uiaServerURL)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	err = driver.PressBack()
//	if err != nil {
//		t.Fatal(err)
//	}
//}

//func TestDriver_PressKeyCode(t *testing.T) {
//	driver, err := NewUIADriver(nil, uiaServerURL)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	err = driver.PressKeyCodeAsync(KCx)
//	if err != nil {
//		t.Fatal(err)
//	}
//	err = driver.PressKeyCodeAsync(KCx, KMCapLocked)
//	if err != nil {
//		t.Fatal(err)
//	}
//	// err = driver.PressKeyCodeAsync(KCExplorer)
//	// if err != nil {
//	// 	t.Fatal(err)
//	// }
//
//	err = driver.PressKeyCode(KCExplorer, KMEmpty)
//	if err != nil {
//		t.Fatal(err)
//	}
//}

//func TestDriver_LongPressKeyCode(t *testing.T) {
//	driver, err := NewUIADriver(nil, uiaServerURL)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	err = driver.LongPressKeyCode(KCAt, KMEmpty)
//	if err != nil {
//		t.Fatal(err)
//	}
//}
//
//func TestDriver_TouchDown(t *testing.T) {
//	driver, err := NewUIADriver(nil, uiaServerURL)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	doTouchUp := func() {
//		err = driver.TouchUp(400, 260)
//		if err != nil {
//			t.Fatal(err)
//		}
//	}
//
//	err = driver.TouchDown(400, 260)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	// _ = driver.TapPoint(Point{400, 500})
//	doTouchUp()
//
//	err = driver.TouchDownPoint(Point{400, 260})
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	doTouchUp()
//}
//
//func TestDriver_TouchUp(t *testing.T) {
//	driver, err := NewUIADriver(nil, uiaServerURL)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	err = driver.TouchDown(400, 260)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	// err = driver.TouchUp(400, 260)
//	err = driver.TouchUpPoint(Point{400, 260})
//	if err != nil {
//		t.Fatal(err)
//	}
//}
//
//func TestDriver_TouchMove(t *testing.T) {
//	driver, err := NewUIADriver(nil, uiaServerURL)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	doTouchDown := func(x, y int) {
//		err = driver.TouchDown(x, y)
//		if err != nil {
//			t.Fatal(err)
//		}
//	}
//
//	doTouchUp := func(x, y int) {
//		err = driver.TouchUp(x, y)
//		if err != nil {
//			t.Fatal(err)
//		}
//	}
//
//	doTouchDown(400, 260)
//
//	err = driver.TouchMove(400, 500)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	doTouchUp(400, 500)
//
//	doTouchDown(400, 500)
//
//	err = driver.TouchMove(400, 260)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	doTouchUp(400, 260)
//}
//
//func TestDriver_OpenNotification(t *testing.T) {
//	driver, err := NewUIADriver(nil, uiaServerURL)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	err = driver.OpenNotification()
//	if err != nil {
//		t.Fatal(err)
//	}
//}
//
//func TestDriver_Flick(t *testing.T) {
//	driver, err := NewUIADriver(nil, uiaServerURL)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	err = driver.Flick(50, -100)
//	if err != nil {
//		t.Fatal(err)
//	}
//}
//
//func TestDriver_ScrollTo(t *testing.T) {
//	driver, err := NewUIADriver(nil, uiaServerURL)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	err = driver.ScrollTo(BySelector{ClassName: "android.widget.SeekBar"})
//	if err != nil {
//		t.Fatal(err)
//	}
//}

//func TestDriver_MultiPointerGesture(t *testing.T) {
//	driver, err := NewUIADriver(nil, uiaServerURL)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	gesture1 := NewTouchAction().Add(150, 340, 0.35).AddFloat(50, 300)
//	gesture2 := NewTouchAction().Add(200, 340).AddFloat(300, 300)
//	gesture3 := NewTouchAction().Add(300, 500).AddFloat(350, 500).AddPoint(Point{300, 550}).AddPointF(PointF{350, 550})
//	_ = gesture3
//
//	// err = driver.MultiPointerGesture(gesture1, gesture2)
//	err = driver.MultiPointerGesture(gesture1, gesture2, gesture3)
//	if err != nil {
//		t.Fatal(err)
//	}
//}
//
//func TestDriver_PerformW3CActions(t *testing.T) {
//	driver, err := NewUIADriver(nil, uiaServerURL)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	// actionKey := NewW3CAction(ATKey, NewW3CGestures().KeyDown("g").KeyUp("g").Pause().KeyDown("o").KeyUp("o"))
//	// actionKey := NewW3CAction(ATKey, NewW3CGestures().SendKeys("golang"))
//	// err = driver.PerformW3CActions(actionKey)
//	// if err != nil {
//	// 	t.Fatal(err)
//	// }
//
//	// var queryField map[string]string
//	// queryField = make(map[string]string)
//	// {
//	// 	queryField = map[string]string{
//	// 		"a": "",
//	// 	}
//	// }
//
//	elem, err := driver.FindElement(BySelector{ResourceIdID: "com.android.settings:id/search"})
//	if err != nil {
//		t.Fatal(err)
//	}
//	// actionPointer := NewW3CAction(ATPointer, NewW3CGestures().PointerMove(0, 0, elem.id).PointerDown().Pause(3).PointerUp())
//	// actionPointer := NewW3CAction(ATPointer,
//	// 	NewW3CGestures().PointerMove(400, 500, "viewport").PointerDown().Pause(2).
//	// 		PointerMove(0, 0, elem.id).Pause(2).
//	// 		PointerMove(20, 0, "pointer").Pause(2).
//	// 		PointerUp(),
//	// )
//	actionPointer := NewW3CAction(ATPointer,
//		NewW3CGestures().PointerMoveTo(400, 500).PointerDown().
//			PointerMouseOver(0, 0, elem).
//			PointerMoveRelative(20, 0).PointerUp())
//	err = driver.PerformW3CActions(actionPointer)
//	if err != nil {
//		t.Fatal(err)
//	}
//}
//
//func TestDriver_GetClipboard(t *testing.T) {
//	driver, err := NewUIADriver(nil, uiaServerURL)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	text, err := driver.GetClipboard()
//	if err != nil {
//		t.Fatal(err)
//	}
//	t.Log(text)
//}
//
//func TestDriver_SetClipboard(t *testing.T) {
//	driver, err := NewUIADriver(nil, uiaServerURL)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	content := "test123"
//	err = driver.SetClipboard(ClipDataTypePlaintext, content)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	text, err := driver.GetClipboard()
//	if err != nil {
//		t.Fatal(err)
//	}
//	if text != content {
//		t.Fatal("should be the same")
//	}
//}

func TestDriver_AlertAccept(t *testing.T) {
	driver, err := NewUIADriver(nil, uiaServerURL)
	if err != nil {
		t.Fatal(err)
	}

	err = driver.AlertAccept()
	// err = driver.AlertAccept("是")
	if err != nil {
		t.Fatal(err)
	}
}

func TestDriver_AlertDismiss(t *testing.T) {
	driver, err := NewUIADriver(nil, uiaServerURL)
	if err != nil {
		t.Fatal(err)
	}

	// err = driver.AlertDismiss()
	err = driver.AlertDismiss("否")
	if err != nil {
		t.Fatal(err)
	}
}

//func TestDriver_SetAppiumSettings(t *testing.T) {
//	driver, err := NewUIADriver(nil, uiaServerURL)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	appiumSettings, err := driver.GetAppiumSettings()
//	if err != nil {
//		t.Fatal(err)
//	}
//	sdopd := appiumSettings["shutdownOnPowerDisconnect"]
//	t.Log("shutdownOnPowerDisconnect:", sdopd)
//
//	err = driver.SetAppiumSettings(map[string]interface{}{"shutdownOnPowerDisconnect": !sdopd.(bool)})
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	appiumSettings, err = driver.GetAppiumSettings()
//	if err != nil {
//		t.Fatal(err)
//	}
//	if appiumSettings["shutdownOnPowerDisconnect"] == sdopd.(bool) {
//		t.Fatal("should not be equal")
//	}
//	t.Log("shutdownOnPowerDisconnect:", appiumSettings["shutdownOnPowerDisconnect"])
//}

func TestDriver_SetOrientation(t *testing.T) {
	driver, err := NewUIADriver(nil, uiaServerURL)
	if err != nil {
		t.Fatal(err)
	}

	err = driver.SetOrientation(OrientationLandscapeLeft)
	// err = driver.SetOrientation(OrientationPortrait)
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

//func TestDriver_NetworkConnection(t *testing.T) {
//	driver, err := NewUIADriver(nil, uiaServerURL)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	err = driver.NetworkConnection(NetworkTypeWifi)
//	if err != nil {
//		t.Fatal(err)
//	}
//}

func TestDriver_FindElement(t *testing.T) {
	driver, err := NewUIADriver(nil, uiaServerURL)
	if err != nil {
		t.Fatal(err)
	}

	elem, err := driver.FindElement(BySelector{ResourceIdID: "android:id/content"})
	if err != nil {
		t.Fatal(err)
	}
	e := ElementAttribute{}.WithLabel("class")
	t.Log(elem.GetAttribute(e))
}

func TestDriver_FindElements(t *testing.T) {
	driver, err := NewUIADriver(nil, uiaServerURL)
	if err != nil {
		t.Fatal(err)
	}

	// elements, err := driver.FindElements(BySelector{ResourceIdID: "com.android.settings:id/title"})
	elements, err := driver.FindElements(BySelector{UiAutomator: "new UiSelector().textStartsWith(\"应\");"})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(len(elements))
}

func TestDriver_WaitWithTimeoutAndInterval(t *testing.T) {
	driver, err := NewUIADriver(nil, uiaServerURL)
	if err != nil {
		t.Fatal(err)
	}
	element, err := driver.FindElement(BySelector{UiAutomator: "new UiSelector().className(\"android.view.ViewGroup\");"})
	if err != nil {
		t.Fatal(err)
	}

	elem, err := element.FindElement(BySelector{UiAutomator: "new UiSelector().className(\"android.widget.LinearLayout\").index(6);"})
	if err != nil {
		t.Fatal(err)
	}

	rect, err := elem.Rect()
	if err != nil {
		t.Fatal(err)
	}

	x := rect.X + int(float64(rect.Width)*2)
	y := rect.Y + rect.Height/2
	err = driver.Tap(x, y)
	if err != nil {
		t.Fatal(err)
	}

	by := BySelector{UiAutomator: "new UiSelector().text(\"科技\");"}
	exists := func(d WebDriver) (bool, error) {
		element, err = d.FindElement(by)
		if err == nil {
			return true, nil
		}
		return false, nil
	}

	err = driver.WaitWithTimeoutAndInterval(exists, 1, 1)
	if err != nil {
		t.Fatal(err)
	}

	// element, err = driver.FindElement(by)
	// if err != nil {
	// 	t.Fatal(err)
	// }

	err = element.Click()
	if err != nil {
		t.Fatal(err)
	}
}

//func TestDriver_ActiveElement(t *testing.T) {
//	device, _ := NewAndroidDevice()
//	driver, err := device.NewUSBDriver(nil)
//	if err != nil {
//		t.Fatal(err)
//	}
//	defer func() {
//		_ = driver.Dispose()
//	}()
//
//	element, err := driver.ActiveElement()
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	if err = element.SendKeys("test"); err != nil {
//		t.Fatal(err)
//	}
//}

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
	devices, err := DeviceList()
	if err != nil {
		t.Fatal(err)
	}
	for i := range devices {
		t.Log(devices[i].Serial())
	}
}

//func TestAndroidNewUSBDriver(t *testing.T) {
//	device, _ := NewAndroidDevice()
//	driver, err := device.NewUSBDriver(nil)
//	if err != nil {
//		t.Fatal(err)
//	}
//	defer driver.Dispose()
//
//	ready, err := driver.Status()
//	if err != nil {
//		t.Fatal(err)
//	}
//	if !ready {
//		t.Fatal("should be 'true'")
//	}
//}

//func TestDriver_ActiveAppPackageName(t *testing.T) {
//	device, _ := NewAndroidDevice()
//	driver, err := device.NewUSBDriver(nil)
//	if err != nil {
//		t.Fatal(err)
//	}
//	defer driver.Dispose()
//
//	appPackageName, err := driver.ActiveAppPackageName()
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	t.Log(appPackageName)
//}

func TestDriver_AppLaunch(t *testing.T) {
	device, _ := NewAndroidDevice()
	driver, err := device.NewUSBDriver(nil)
	if err != nil {
		t.Fatal(err)
	}

	err = driver.AppLaunch("com.android.settings")
	if err != nil {
		t.Fatal(err)
	}

	raw, err := driver.Screenshot()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(ioutil.WriteFile("s1.png", raw.Bytes(), 0o600))
}

func TestDriver_KeepAlive(t *testing.T) {
	device, _ := NewAndroidDevice()
	driver, err := device.NewUSBDriver(nil)
	if err != nil {
		t.Fatal(err)
	}

	err = driver.AppLaunch("com.android.settings")
	if err != nil {
		t.Fatal(err)
	}

	_, err = driver.Screenshot()
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(60 * time.Second)

	_, err = driver.Screenshot()
	if err != nil {
		t.Fatal(err)
	}
}

func TestDriver_AppTerminate(t *testing.T) {
	device, _ := NewAndroidDevice()
	driver, err := device.NewUSBDriver(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer driver.Dispose()

	_, err = driver.AppTerminate("tv.danmaku.bili")
	if err != nil {
		t.Fatal(err)
	}
}

//func TestNewWiFiDriver(t *testing.T) {
//	device, _ := NewAndroidDevice(WithAdbIP("192.168.1.28"))
//	driver, err := device.NewHTTPDriver(nil)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	// SetDebug(false, true)
//	_, err = driver.ActiveAppActivity()
//	if err != nil {
//		t.Fatal(err)
//	}
//}

//func TestDriver_AppInstall(t *testing.T) {
//	device, _ := NewAndroidDevice()
//	driver, err := device.NewUSBDriver(nil)
//	if err != nil {
//		t.Fatal(err)
//	}
//	defer driver.Dispose()
//
//	err = driver.AppInstall("/Users/hero/Desktop/xuexi_android_10002068.apk")
//	if err != nil {
//		t.Fatal(err)
//	}
//}

//func TestDriver_AppUninstall(t *testing.T) {
//	device, _ := NewAndroidDevice()
//	driver, err := device.NewUSBDriver(nil)
//	if err != nil {
//		t.Fatal(err)
//	}
//	defer driver.Dispose()
//
//	err = driver.AppUninstall("cn.xuexi.android")
//	if err != nil {
//		t.Fatal(err)
//	}
//}

func TestBySelector_getMethodAndSelector(t *testing.T) {
	testVal := "test id"
	bySelector := BySelector{ResourceIdID: testVal}
	method, selector := bySelector.getMethodAndSelector()
	if method != "id" || selector != testVal {
		t.Fatal(method, "=", selector)
	}

	bySelector = BySelector{ContentDescription: testVal}
	method, selector = bySelector.getMethodAndSelector()
	if method != "accessibility id" || selector != testVal {
		t.Fatal(method, "=", selector)
	}
}

func TestElement_Text(t *testing.T) {
	driver, err := NewUIADriver(nil, uiaServerURL)
	if err != nil {
		t.Fatal(err)
	}

	elem, err := driver.FindElement(BySelector{ResourceIdID: "com.android.settings:id/category_title"})
	if err != nil {
		t.Fatal(err)
	}

	text, err := elem.Text()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(text)
}

func TestElement_GetAttribute(t *testing.T) {
	driver, err := NewUIADriver(nil, uiaServerURL)
	if err != nil {
		t.Fatal(err)
	}

	elem, err := driver.FindElement(BySelector{ResourceIdID: "com.android.settings:id/category_title"})
	if err != nil {
		t.Fatal(err)
	}

	e := ElementAttribute{}.WithName("class")
	attribute, err := elem.GetAttribute(e)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(attribute)
}

//func TestElement_ContentDescription(t *testing.T) {
//	driver, err := NewUIADriver(nil, uiaServerURL)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	elem, err := driver.FindElement(BySelector{ResourceIdID: "com.android.settings:id/search"})
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	name, err := elem.ContentDescription()
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	t.Log(name)
//}

func TestElement_Size(t *testing.T) {
	driver, err := NewUIADriver(nil, uiaServerURL)
	if err != nil {
		t.Fatal(err)
	}

	elem, err := driver.FindElement(BySelector{ResourceIdID: "com.android.settings:id/search"})
	if err != nil {
		t.Fatal(err)
	}

	size, err := elem.Size()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(size)
}

func TestElement_Rect(t *testing.T) {
	driver, err := NewUIADriver(nil, uiaServerURL)
	if err != nil {
		t.Fatal(err)
	}

	elem, err := driver.FindElement(BySelector{ResourceIdID: "com.android.settings:id/category_title"})
	if err != nil {
		t.Fatal(err)
	}

	rect, err := elem.Rect()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(rect)
}

func TestElement_Screenshot(t *testing.T) {
	driver, err := NewUIADriver(nil, uiaServerURL)
	if err != nil {
		t.Fatal(err)
	}

	elem, err := driver.FindElement(BySelector{ResourceIdID: "com.android.settings:id/category_title"})
	if err != nil {
		t.Fatal(err)
	}

	screenshot, err := elem.Screenshot()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(ioutil.WriteFile("/Users/hero/Desktop/e1.png", screenshot.Bytes(), 0o600))
}

func TestElement_Location(t *testing.T) {
	driver, err := NewUIADriver(nil, uiaServerURL)
	if err != nil {
		t.Fatal(err)
	}

	elem, err := driver.FindElement(BySelector{ResourceIdID: "com.android.settings:id/category_title"})
	if err != nil {
		t.Fatal(err)
	}

	location, err := elem.Location()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(location)
}

func TestElement_Click(t *testing.T) {
	driver, err := NewUIADriver(nil, uiaServerURL)
	if err != nil {
		t.Fatal(err)
	}

	elem, err := driver.FindElement(BySelector{ResourceIdID: "com.android.settings:id/title"})
	if err != nil {
		t.Fatal(err)
	}

	err = elem.Click()
	if err != nil {
		t.Fatal(err)
	}
}

func TestElement_Clear(t *testing.T) {
	driver, err := NewUIADriver(nil, uiaServerURL)
	if err != nil {
		t.Fatal(err)
	}

	elem, err := driver.FindElement(BySelector{ResourceIdID: "android:id/search_src_text"})
	if err != nil {
		t.Fatal(err)
	}

	err = elem.Clear()
	if err != nil {
		t.Fatal(err)
	}
}

func TestElement_SendKeys(t *testing.T) {
	driver, err := NewUIADriver(nil, uiaServerURL)
	if err != nil {
		t.Fatal(err)
	}

	elem, err := driver.FindElement(BySelector{ResourceIdID: "android:id/search_src_text"})
	if err != nil {
		t.Fatal(err)
	}

	// return

	// err = elem.SendKeys("abc")
	err = elem.SendKeys("456")
	if err != nil {
		t.Fatal(err)
	}
}

func TestElement_FindElements(t *testing.T) {
	driver, err := NewUIADriver(nil, uiaServerURL)
	if err != nil {
		t.Fatal(err)
	}

	parentElem, err := driver.FindElement(BySelector{ResourceIdID: "com.android.settings:id/main_content"})
	if err != nil {
		t.Fatal(err)
	}

	elements, err := parentElem.FindElements(BySelector{ResourceIdID: "com.android.settings:id/category"})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(len(elements))
}

func TestElement_FindElement(t *testing.T) {
	driver, err := NewUIADriver(nil, uiaServerURL)
	if err != nil {
		t.Fatal(err)
	}

	parentElem, err := driver.FindElement(BySelector{ResourceIdID: "com.android.settings:id/main_content"})
	if err != nil {
		t.Fatal(err)
	}

	elem, err := parentElem.FindElement(BySelector{ResourceIdID: "com.android.settings:id/category_title"})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(elem.Text())
}

func TestElement_Swipe(t *testing.T) {
	driver, err := NewUIADriver(nil, uiaServerURL)
	if err != nil {
		t.Fatal(err)
	}

	elem, err := driver.FindElement(BySelector{ResourceIdID: "com.android.settings:id/category_title"})
	if err != nil {
		t.Fatal(err)
	}

	rect, err := elem.Rect()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(rect)

	var startX, startY, endX, endY int
	startX = rect.X + rect.Width/20
	startY = rect.Y + rect.Height/2
	endX = startX
	endY = startY - startY/2
	err = elem.Swipe(startX, startY, endX, endY)
	if err != nil {
		t.Fatal(err)
	}
}

//func TestElement_Drag(t *testing.T) {
//	driver, err := NewUIADriver(nil, uiaServerURL)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	elements, err := driver.FindElements(BySelector{ClassName: "android.widget.TextView"})
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	for i, elem := range elements {
//		text, _ := elem.Text()
//		t.Log(i, text)
//	}
//
//	rect, err := elements[0].Rect()
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	// err = elements[0].Drag(300, 450, 256)
//	err = elements[0].Drag(300, 450, 256)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	err = elements[0].DragTo(elements[1], 256)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	endPoint := PointF{X: float64(rect.X + rect.Width/3*2), Y: float64(rect.Y + rect.Height/2)}
//	err = elements[0].DragPointF(endPoint, 256)
//	if err != nil {
//		t.Fatal()
//	}
//}

//func TestElement_Flick(t *testing.T) {
//	driver, err := NewUIADriver(nil, uiaServerURL)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	elem, err := driver.FindElement(BySelector{UiAutomator: "new UiSelector().text(\"提示音和通知\");"})
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	err = elem.Flick(36, 20, 100)
//	if err != nil {
//		t.Fatal(err)
//	}
//}

//func TestElement_ScrollTo(t *testing.T) {
//	driver, err := NewUIADriver(nil, uiaServerURL)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	// how to make it work?
//	// parentElem, err := driver.FindElement(BySelector{ClassName: "android.widget.ScrollView"})
//	// parentElem, err := driver.FindElement(BySelector{ResourceIdID: "com.cyanogenmod.filemanager:id/navigation_view_layout"})
//	parentElem, err := driver.FindElement(BySelector{ResourceIdID: "com.android.settings:id/dashboard"})
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	err = parentElem.ScrollTo(BySelector{ContentDescription: "电池"})
//	if err != nil {
//		t.Fatal(err)
//	}
//}

//func TestElement_ScrollToElement(t *testing.T) {
//	// android.widget.HorizontalScrollView
//	driver, err := NewUIADriver(nil, uiaServerURL)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	// how to make it work?
//	parentElem, err := driver.FindElement(BySelector{UiAutomator: "new UiSelector().resourceId(\"com.android.settings:id/dashboard\");"})
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	element, err := driver.FindElement(BySelector{UiAutomator: "new UiSelector().text(\"电池\");"})
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	err = parentElem.ScrollToElement(element)
//	if err != nil {
//		t.Fatal(err)
//	}
//}
