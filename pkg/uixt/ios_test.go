//go:build localtest

package uixt

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
)

var (
	bundleId     = "com.apple.Preferences"
	driver       IDriver
	iOSDriverExt *DriverExt
)

func setup(t *testing.T) {
	device, err := NewIOSDevice(
		option.WithWDAPort(8700),
		option.WithWDAMjpegPort(8800),
		option.WithWDALogOn(true))
	if err != nil {
		t.Fatal(err)
	}
	capabilities := option.NewCapabilities()
	capabilities.WithDefaultAlertAction(option.AlertActionAccept)
	driver, err = device.NewHTTPDriver(capabilities)
	if err != nil {
		t.Fatal(err)
	}
	iOSDriverExt, err = newDriverExt(device, driver)
	if err != nil {
		t.Fatal(err)
	}
}

func TestViaUSB(t *testing.T) {
	setup(t)
	t.Log(driver.Status())
}

func TestInstall(t *testing.T) {
	setup(t)
	err := iOSDriverExt.Install("xxx.ipa",
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
	t.Log(appInfo)
}

func TestNewWDAHTTPDriver(t *testing.T) {
	device, _ := NewIOSDevice()
	var err error
	_, err = device.NewHTTPDriver(nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNewUSBDriver(t *testing.T) {
	setup(t)

	// t.Log(driver.IsWdaHealthy())
}

func TestDriver_DeviceScaleRatio(t *testing.T) {
	setup(t)

	scaleRatio, err := driver.Scale()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(scaleRatio)
}

func Test_remoteWD_DeleteSession(t *testing.T) {
	setup(t)

	err := driver.DeleteSession()
	if err != nil {
		t.Fatal(err)
	}
}

func Test_remoteWD_HealthCheck(t *testing.T) {
	setup(t)

	err := driver.HealthCheck()
	if err != nil {
		t.Fatal(err)
	}
}

func Test_remoteWD_GetAppiumSettings(t *testing.T) {
	setup(t)

	settings, err := driver.GetAppiumSettings()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(settings)
}

func Test_remoteWD_SetAppiumSettings(t *testing.T) {
	setup(t)

	const _acceptAlertButtonSelector = "**/XCUIElementTypeButton[`label IN {'允许','好','仅在使用应用期间','暂不'}`]"
	const _dismissAlertButtonSelector = "**/XCUIElementTypeButton[`label IN {'不允许','暂不'}`]"

	key := "acceptAlertButtonSelector"
	value := _acceptAlertButtonSelector

	// settings, err := driver.SetAppiumSettings(map[string]interface{}{"dismissAlertButtonSelector": "暂不"})
	settings, err := driver.SetAppiumSettings(map[string]interface{}{key: value})
	if err != nil {
		t.Fatal(err)
	}
	if settings[key] != value {
		t.Fatal(settings[key])
	}
}

func Test_remoteWD_IsWdaHealthy(t *testing.T) {
	setup(t)

	healthy, err := driver.IsHealthy()
	if err != nil {
		t.Fatal(err)
	}
	if healthy == false {
		t.Fatal("healthy =", healthy)
	}
}

// func Test_remoteWD_WdaShutdown(t *testing.T) {
// 	setup(t)
//
// 	if err := driver.WdaShutdown(); err != nil {
// 		t.Fatal(err)
// 	}
// }

func Test_remoteWD_Status(t *testing.T) {
	setup(t)

	status, err := driver.Status()
	if err != nil {
		t.Fatal(err)
	}
	if status.Ready == false {
		t.Fatal("deviceStatus =", status)
	}
}

func Test_remoteWD_DeviceInfo(t *testing.T) {
	setup(t)

	info, err := driver.DeviceInfo()
	if err != nil {
		t.Fatal(err)
	}
	if len(info.Model) == 0 {
		t.Fatal(info)
	}
}

func Test_remoteWD_BatteryInfo(t *testing.T) {
	setup(t)

	batteryInfo, err := driver.BatteryInfo()
	if err != nil {
		t.Fatal()
	}
	t.Log(batteryInfo)
}

func Test_remoteWD_WindowSize(t *testing.T) {
	setup(t)

	size, err := driver.WindowSize()
	if err != nil {
		t.Fatal()
	}
	t.Log(size)
}

func Test_remoteWD_Screen(t *testing.T) {
	setup(t)

	screen, err := driver.Screen()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(screen)
}

func Test_remoteWD_Homescreen(t *testing.T) {
	setup(t)

	err := driver.Homescreen()
	if err != nil {
		t.Fatal(err)
	}
}

func Test_remoteWD_AppLaunch(t *testing.T) {
	setup(t)

	err := driver.AppLaunch(bundleId)
	// err := driver.AppLaunch(bundleId, NewAppLaunchOption().WithShouldWaitForQuiescence(true))
	// err := driver.AppLaunch(bundleId, NewAppLaunchOption().WithArguments([]string{"-AppleLanguages", "(Russian)"}))
	if err != nil {
		t.Fatal(err)
	}
}

func Test_remoteWD_AppTerminate(t *testing.T) {
	setup(t)

	_, err := driver.AppTerminate(bundleId)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_remoteWD_Tap(t *testing.T) {
	setup(t)

	err := driver.Tap(200, 300)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_remoteWD_DoubleTap(t *testing.T) {
	setup(t)

	err := driver.DoubleTap(200, 300)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_remoteWD_TouchAndHold(t *testing.T) {
	setup(t)

	// err := driver.TouchAndHold(200, 300)
	err := driver.TouchAndHold(200, 300)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_remoteWD_Drag(t *testing.T) {
	setup(t)

	// err := driver.Drag(200, 300, 200, 500, WithDataPressDuration(0.5))
	err := driver.Drag(200, 300, 200, 500,
		option.WithPressDuration(2), option.WithDuration(3))
	if err != nil {
		t.Fatal(err)
	}
}

func Test_Relative_Drag(t *testing.T) {
	setup(t)

	// err := driver.Drag(200, 300, 200, 500, WithDataPressDuration(0.5))
	err := iOSDriverExt.SwipeRelative(0.5, 0.7, 0.5, 0.5)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_remoteWD_SetPasteboard(t *testing.T) {
	setup(t)

	// err := driver.SetPasteboard(PasteboardTypePlaintext, "gwda")
	err := driver.SetPasteboard(PasteboardTypeUrl, "Clock-stopwatch://")
	// userHomeDir, _ := os.UserHomeDir()
	// bytesImg, _ := ioutil.ReadFile(userHomeDir + "/Pictures/IMG_0806.jpg")
	// err := driver.SetPasteboard(PasteboardTypeImage, string(bytesImg))
	if err != nil {
		t.Fatal(err)
	}
}

func Test_remoteWD_GetPasteboard(t *testing.T) {
	setup(t)

	var buffer *bytes.Buffer
	var err error

	buffer, err = driver.GetPasteboard(PasteboardTypePlaintext)
	// buffer, err = driver.GetPasteboard(PasteboardTypeUrl)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(buffer.String())

	// buffer, err = driver.GetPasteboard(PasteboardTypeImage)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// userHomeDir, _ := os.UserHomeDir()
	// if err = ioutil.WriteFile(userHomeDir+"/Desktop/p1.png", buffer.Bytes(), 0600); err != nil {
	// 	t.Error(err)
	// }
}

func Test_remoteWD_SendKeys(t *testing.T) {
	setup(t)
	// driver.StartCaptureLog("hrp_wda_log")
	err := driver.SendKeys("test", option.WithIdentifier("test"))
	// result, _ := driver.StopCaptureLog()
	// err := driver.SendKeys("App Store", WithFrequency(3))
	if err != nil {
		t.Fatal(err)
	}
	// t.Log(result)
}

func Test_remoteWD_PressButton(t *testing.T) {
	setup(t)

	err := driver.PressButton(DeviceButtonVolumeUp)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second * 1)
	err = driver.PressButton(DeviceButtonVolumeDown)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second * 1)
	err = driver.PressButton(DeviceButtonHome)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_remoteWD_Screenshot(t *testing.T) {
	setup(t)

	screenshot, err := driver.Screenshot()
	if err != nil {
		t.Fatal(err)
	}
	_ = screenshot

	// img, format, err := image.Decode(screenshot)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// userHomeDir, _ := os.UserHomeDir()
	// file, err := os.Create(userHomeDir + "/Desktop/s1." + format)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// defer func() { _ = file.Close() }()
	// switch format {
	// case "png":
	// 	err = png.Encode(file, img)
	// case "jpeg":
	// 	err = jpeg.Encode(file, img, nil)
	// }
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// t.Log(file.Name())
}

func Test_remoteWD_Source(t *testing.T) {
	setup(t)

	var source string
	var err error

	// source, err = driver.Source()
	// if err != nil {
	// 	t.Fatal(err)
	// }

	source, err = driver.Source()
	if err != nil {
		t.Fatal(err)
	}

	// source, err = driver.Source(NewSourceOption().WithFormatAsJson())
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// source, err = driver.Source(NewSourceOption().WithFormatAsDescription())
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// source, err = driver.Source(NewSourceOption().WithFormatAsXml().WithExcludedAttributes([]string{"label", "type", "index"}))
	// if err != nil {
	// 	t.Fatal(err)
	// }

	_ = source
	fmt.Println(source)
}

func TestGetForegroundApp(t *testing.T) {
	setup(t)
	app, err := driver.GetForegroundApp()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(app)
}

func Test_remoteWD_AccessibleSource(t *testing.T) {
	setup(t)

	source, err := driver.AccessibleSource()
	if err != nil {
		t.Fatal(err)
	}
	_ = source
	fmt.Println(source)
}

func TestRecord(t *testing.T) {
	setup(t)
	path, err := driver.(*wdaDriver).RecordScreen("", 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	println(path)
}

// func Test_Backspace(t *testing.T) {
// 	setup(t)

// 	err := driver.Backspace(3)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// }
